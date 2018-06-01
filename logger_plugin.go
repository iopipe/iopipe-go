package iopipe

import "fmt"

// LoggerPluginConfig is the logger plugin configuration
type LoggerPluginConfig struct {
	Key string
}

type loggerPlugin struct {
	LoggerPluginConfig
}

func (p *loggerPlugin) RunHook(hook string) {
	fmt.Println(fmt.Sprintf("[LOGGER - %s] Running hook: %s", p.Key, hook))
}

func (p *loggerPlugin) Meta() *PluginMeta {
	return &PluginMeta{
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

func (p *loggerPlugin) PreSetup(agent *Agent)              {}
func (p *loggerPlugin) PostSetup(agent *Agent)             {}
func (p *loggerPlugin) PreInvoke(context *ContextWrapper)  {}
func (p *loggerPlugin) PostInvoke(context *ContextWrapper) {}
func (p *loggerPlugin) PreReport(report *Report)           {}
func (p *loggerPlugin) PostReport(report *Report)          {}

// LoggerPlugin loads the logger plugin
func LoggerPlugin(config LoggerPluginConfig) PluginInstantiator {
	return func() Plugin {
		return &loggerPlugin{config}
	}
}
