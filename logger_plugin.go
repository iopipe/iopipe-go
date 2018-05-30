package iopipe

import "fmt"

// LoggerPluginConfig is the logger plugin configuration
type LoggerPluginConfig struct {
	Key             string
	wrapperInstance HandlerWrapper
}

type loggerPlugin struct {
	LoggerPluginConfig
}

func (p *loggerPlugin) RunHook(hook string) {
	fmt.Println(fmt.Sprintf("[LOGGER - %s] Running hook: %s", p.Key, hook))
}

func (p *loggerPlugin) Meta() interface{} {
	return struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Homepage string `json:"homepage"`
		Enabled  bool   `json:"enabled"`
	}{
		Name:     p.Name(),
		Version:  p.Version(),
		Homepage: p.Homepage(),
		Enabled:  p.Enabled(),
	}
}

func (p *loggerPlugin) Name() string {
	return "@iopipe/logger"
}

func (p *loggerPlugin) Version() string {
	return "0.1.0"
}

func (p *loggerPlugin) Homepage() string {
	return "https://github.com/iopipe/iopipe-go#logger-plugin"
}

func (p *loggerPlugin) Enabled() bool {
	return true
}

// LoggerPlugin loads the logger plugin
func LoggerPlugin(config LoggerPluginConfig) PluginInstantiator {
	return func(w *HandlerWrapper) Plugin {
		return &loggerPlugin{config}
	}
}
