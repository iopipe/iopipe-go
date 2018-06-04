package iopipe

import (
	"context"
	"os"
	"sync"
	"time"
)

// Config is the config object passed to agent initialization
type Config struct {
	Reporter            Reporter
	Token               *string
	TimeoutWindow       *time.Duration
	Enabled             *bool
	PluginInstantiators []PluginInstantiator
}

// Agent is the IOpipe instance
type Agent struct {
	*Config
	plugins []Plugin
}

var (
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

	enabled := &defaultConfigEnabled
	if config.Enabled != nil {
		enabled = config.Enabled
	}

	reporter := defaultReporter
	if config.Reporter != nil {
		reporter = config.Reporter
	}

	timeoutWindow := &defaultConfigTimeoutWindow
	if config.TimeoutWindow != nil {
		timeoutWindow = config.TimeoutWindow
	}

	envtoken := os.Getenv("IOPIPE_TOKEN")
	token := &envtoken
	if config.Token != nil {
		token = config.Token
	}

	a.Config = &Config{
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
	// Only wrap the handler if the agent is enabled and the token is not nil or
	// an empty string
	if a.Enabled != nil && *a.Enabled && a.Token != nil && *a.Token != "" {
		return wrapHandler(handler, a)
	}
	return handler
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
