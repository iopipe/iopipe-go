package iopipe

import (
	"os"
	"time"
)

// Config is the config object passed to NewAgent
type Config struct {
	Token               *string
	TimeoutWindow       *time.Duration
	Enabled             *bool
	PluginInstantiators []PluginInstantiator
}

type Agent struct {
	*Config
	Plugins []Plugin
}

// defaults
var defaultConfigTimeoutWindow = time.Duration(150 * time.Millisecond)
var defaultConfigEnabled = true

// NewAgent returns a new IOpipe with config
func NewAgent(config Config) *Agent {
	timeoutWindow := &defaultConfigTimeoutWindow
	enabled := &defaultConfigEnabled
	envtoken := os.Getenv("IOPIPE_TOKEN")
	token := &envtoken

	if config.Token != nil {
		token = config.Token
	}

	if config.Enabled != nil {
		enabled = config.Enabled
	}
	if config.TimeoutWindow != nil {
		timeoutWindow = config.TimeoutWindow
	}

	// Handle plugin nil case, TODO populate with default plugins, and include
	// overrides from users
	if config.PluginInstantiators == nil {
		config.PluginInstantiators = []PluginInstantiator{}
	}

	return &Agent{
		Config: &Config{
			Token:               token,
			TimeoutWindow:       timeoutWindow,
			Enabled:             enabled,
			PluginInstantiators: config.PluginInstantiators,
		},
	}
}

func (a *Agent) WrapHandler(handler interface{}) interface{} {
	// Only wrap the handler if the agent is enabled and the token is not nil or
	// an empty string
	if a.Enabled != nil && *a.Enabled && a.Token != nil && *a.Token != "" {
		return wrapHandler(handler, a)
	}
	return handler
}
