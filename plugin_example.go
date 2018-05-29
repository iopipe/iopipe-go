package iopipe

import "fmt"

// LoggerPluginConfig is the logger plugin config
type LoggerPluginConfig struct {
	Key             string
	wrapperInstance Wrapper
}

type loggerPlugin struct {
	LoggerPluginConfig
}

func (p *loggerPlugin) RunHook(hook string) {
	fmt.Println(fmt.Sprintf("[LOGGER - %s] Running hook: %s", p.Key, hook))
}

func (p *loggerPlugin) Meta() interface{} {
	return struct {
		Name string `json:"name"`
	}{
		Name: p.Name(),
	}
}

func (p *loggerPlugin) Name() string {
	return "logger-plugin"
}

// LoggerPlugin loads the logger plugin
func LoggerPlugin(config LoggerPluginConfig) PluginInstantiator {
	return func(w *Wrapper) Plugin {
		return &loggerPlugin{config}
	}
}
