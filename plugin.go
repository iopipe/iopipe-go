package iopipe

// PluginInstantiator is the function that initializes the plugin
type PluginInstantiator func() Plugin

// PluginMeta is meta data about the plugin
type PluginMeta struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Homepage string `json:"homepage"`
	Enabled  bool   `json:"enabled"`
}

// Plugin is the interface a plugin should implement
type Plugin interface {
	Meta() *PluginMeta
	Name() string
	Version() string
	Homepage() string
	Enabled() bool

	// Hook methods
	PreSetup(*Agent)
	PostSetup(*Agent)
	PreInvoke(*ContextWrapper, interface{})
	PostInvoke(*ContextWrapper, interface{})
	PreReport(*Report)
	PostReport(*Report)
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
