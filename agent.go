package iopipe

import (
	"os"
	"time"
)

// AgentConfig is the config object passed to NewAgent
type AgentConfig struct {
	Token         *string
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

// NewAgent returns a new IOpipe with config
func NewAgent(config AgentConfig) *agent {
	timeoutWindow := &defaultAgentConfigTimeoutWindow
	enabled := &defaultAgentConfigEnabled
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

	return &agent{
		AgentConfig: &AgentConfig{
			Token:               token,
			TimeoutWindow:       timeoutWindow,
			Enabled:             enabled,
			PluginInstantiators: config.PluginInstantiators,
		},
	}
}

func (a *agent) WrapHandler(handler interface{}) interface{} {
	// Only wrap the handler if the agent is enabled and the token is not nil or
	// an empty string
	if a.Enabled != nil && *a.Enabled && a.Token != nil && *a.Token != "" {
		return wrapHandler(handler, a)
	}
	return handler
}
