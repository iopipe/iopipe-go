package iopipe

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Config is the config object passed to agent initialization
type Config struct {
	Debug               *bool
	Enabled             *bool
	PluginInstantiators []PluginInstantiator
	Reporter            Reporter
	TimeoutWindow       *time.Duration
	Token               *string
}

// Agent is the IOpipe instance
type Agent struct {
	*Config
	plugins []Plugin
}

var (
	defaultConfigDebug         = false
	defaultConfigEnabled       = true
	defaultConfigTimeoutWindow = time.Duration(150 * time.Millisecond)
	defaultReporter            = sendReport
)

// NewAgent returns a new IOpipe instance with config
func NewAgent(config Config) *Agent {
	if config.PluginInstantiators == nil {
		config.PluginInstantiators = []PluginInstantiator{}
	}

	var plugins []Plugin
	pluginInstantiators := config.PluginInstantiators

	if pluginInstantiators != nil {
		plugins = make([]Plugin, len(pluginInstantiators))
		for index, pluginInstantiator := range pluginInstantiators {
			plugins[index] = pluginInstantiator()
		}
	}

	a := &Agent{
		plugins: plugins,
	}

	a.preSetup()

	// Debug
	debug := &defaultConfigDebug
	envDebug := os.Getenv("IOPIPE_DEBUG")
	if envDebug != "" {
		debug = strToBool(envDebug)
	}
	if config.Debug != nil {
		debug = config.Debug
	}
	if *debug {
		enableDebugMode()
	}

	// Enabled
	enabled := &defaultConfigEnabled
	envEnabled := os.Getenv("IOPIPE_ENABLED")
	if envEnabled != "" {
		enabled = strToBool(envEnabled)
	}
	if config.Enabled != nil {
		enabled = config.Enabled
	}

	// Reporter
	reporter := defaultReporter
	if config.Reporter != nil {
		reporter = config.Reporter
	}

	// TimeoutWindow
	timeoutWindow := &defaultConfigTimeoutWindow
	envTimeoutWindow := os.Getenv("IOPIPE_TIMEOUT_WINDOW")
	envTimeoutWindowInt, err := strconv.Atoi(envTimeoutWindow)
	if err == nil {
		envTimeoutWindowDuration := time.Duration(time.Duration(envTimeoutWindowInt) * time.Millisecond)
		timeoutWindow = &envTimeoutWindowDuration
	}
	if config.TimeoutWindow != nil {
		timeoutWindow = config.TimeoutWindow
	}

	// Token
	envToken := os.Getenv("IOPIPE_TOKEN")
	token := &envToken
	if config.Token != nil {
		token = config.Token
	}

	a.Config = &Config{
		Debug:               debug,
		Enabled:             enabled,
		PluginInstantiators: pluginInstantiators,
		Reporter:            reporter,
		TimeoutWindow:       timeoutWindow,
		Token:               token,
	}

	a.postSetup()

	return a
}

// WrapHandler wraps the handler with the IOpipe agent
func (a *Agent) WrapHandler(handler interface{}) interface{} {
	logger.Debug(fmt.Sprintf("%s wrapped with IOpipe decorator", getFuncName(handler)))

	if a.Enabled != nil && !*a.Enabled {
		logger.Debug("IOpipe agent disabled, skipping reporting")
		return handler
	}

	if a.Token != nil && *a.Token == "" {
		logger.Debug("Your function is decorated with iopipe, but a valid token was not found. Set the IOPIPE_TOKEN environment variable with your IOpipe project token.")
		return handler
	}

	return wrapHandler(handler, a)
}

func (a *Agent) preSetup() {
	var wg sync.WaitGroup
	wg.Add(len(a.plugins))

	for _, plugin := range a.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.PreSetup(a)
			}
		}(plugin)
	}

	wg.Wait()
}

func (a *Agent) postSetup() {
	var wg sync.WaitGroup
	wg.Add(len(a.plugins))

	for _, plugin := range a.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.PostSetup(a)
			}
		}(plugin)
	}

	wg.Wait()
}

func wrapHandler(handler interface{}, agentInstance *Agent) lambdaHandler {
	// decorate the handler
	return func(context context.Context, payload interface{}) (interface{}, error) {
		handlerWrapper := NewHandlerWrapper(handler, agentInstance)
		return handlerWrapper.Invoke(context, payload)
	}
}
