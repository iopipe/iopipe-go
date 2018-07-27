package iopipe

import "context"

// PluginInstantiator is the function that initializes the plugin
type PluginInstantiator func() Plugin

// PluginMeta is meta data about the plugin
type PluginMeta struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Homepage string   `json:"homepage"`
	Enabled  bool     `json:"enabled"`
	Uploads  []string `json:"uploads"`
}

// Plugin is the interface a plugin should implement
type Plugin interface {
	Meta() *PluginMeta
	Enabled() bool

	// Hook methods
	PreSetup(*Agent)
	PostSetup(*Agent)
	PreInvoke(context.Context, interface{})
	PostInvoke(context.Context, interface{})
	PreReport(*Report)
	PostReport(*Report)
}
