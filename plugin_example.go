package iopipe

import "fmt"

type LoggerPluginConfig struct {
	Key string
	wrapperInstance wrapper
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

func LoggerPlugin(config LoggerPluginConfig) PluginInstantiator {
	return func(w *wrapper) Plugin {
		return &loggerPlugin{config}
	}
}
