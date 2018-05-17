package iopipe

import (
	"time"
)

type AgentConfig struct {
	Token         string
	TimeoutWindow *time.Duration
	Enabled       *bool
	// We should be passing in plugins here, and the agent should instantiate
	// them
	PluginInstantiators []PluginInstantiator
}

type agent struct {
	*AgentConfig
	Plugins []Plugin
}

// defaults
var defaultAgentConfigTimeoutWindow = time.Duration(150 * time.Millisecond)
var defaultAgentConfigEnabled = true

func NewAgent(config AgentConfig) *agent {
	timeoutWindow := &defaultAgentConfigTimeoutWindow
	enabled := &defaultAgentConfigEnabled

	if config.Enabled != nil {
		enabled = config.Enabled
	}
	if config.TimeoutWindow != nil {
		timeoutWindow = config.TimeoutWindow
	}

	// Handle plugin nil case

	return &agent{
		AgentConfig: &AgentConfig{
			Token:               config.Token,
			TimeoutWindow:       timeoutWindow,
			Enabled:             enabled,
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
