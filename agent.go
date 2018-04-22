package iopipe

import "time"

type AgentConfig struct {
	Token         *string
	TimeoutWindow *time.Duration
	Enabled       *bool
	PluginInstantiators *[]PluginInstantiator
}

type agent struct {
	*AgentConfig
	Plugins       []Plugin
}

func NewAgent(config AgentConfig) *agent {
	// defaults
	defaultTimeoutWindow := time.Duration(150 * time.Millisecond)
	defaultEnabled := true

	timeoutWindow := &defaultTimeoutWindow
	enabled := &defaultEnabled

	if config.Enabled != nil {
		enabled = config.Enabled
	}
	if config.TimeoutWindow != nil {
		timeoutWindow = config.TimeoutWindow
	}

	return &agent{
		AgentConfig: &AgentConfig{
			Token:         config.Token,
			TimeoutWindow: timeoutWindow,
			Enabled:       enabled,
			PluginInstantiators: config.PluginInstantiators,
		},

	}
}

func (a *agent) WrapHandler(handler interface{}) interface{} {
	if a.Enabled != nil && *a.Enabled {
		return wrapHandler(handler, a)
	} else {
		return handler
	}
}
