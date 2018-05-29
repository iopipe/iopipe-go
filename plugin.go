package iopipe

// PluginInstantiator is the function passed the wrapped handler
type PluginInstantiator func(*Wrapper) Plugin

// Plugin is the interface a plugin should implement
type Plugin interface {
	RunHook(string)
	Meta() interface{}
	Name() string
}

// HookPreSetup is ahook run before agent setup
const HookPreSetup = "pre:setup"

// HookPostSetup is a hook run after agent setup
const HookPostSetup = "post:setup"

// HookPreInvoke is a hook run before an invocation
const HookPreInvoke = "pre:invoke"

// HookPostInvoke is a hook run after an invocation
const HookPostInvoke = "post:invoke"

// HookPreReport is a hook run before reporting
const HookPreReport = "pre:report"

// HookPostReport is a hook run after reporting
const HookPostReport = "post:report"
