package iopipe

// TestPluginConfig is a test plugin config
type TestPluginConfig struct {
	LastHook string
}

type testPlugin struct {
	TestPluginConfig
}

func (p *testPlugin) RunHook(hook string) {
	p.LastHook = hook
}

func (p *testPlugin) Meta() *PluginMeta {
	return &PluginMeta{
		Name:     p.Name(),
		Version:  p.Version(),
		Homepage: p.Homepage(),
		Enabled:  p.Enabled(),
	}
}

func (p *testPlugin) Name() string {
	return "test-plugin"
}

func (p *testPlugin) Version() string {
	return "0.1.0"
}

func (p *testPlugin) Homepage() string {
	return "https://github.com/iopipe/iopipe-go"
}

func (p *testPlugin) Enabled() bool {
	return true
}

func (p *testPlugin) PreSetup(agent *Agent)              {}
func (p *testPlugin) PostSetup(agent *Agent)             {}
func (p *testPlugin) PreInvoke(context *ContextWrapper)  {}
func (p *testPlugin) PostInvoke(context *ContextWrapper) {}
func (p *testPlugin) PreReport(report *Report)           {}
func (p *testPlugin) PostReport(report *Report)          {}

// TestPlugin return a test plugin
func TestPlugin(config TestPluginConfig) PluginInstantiator {
	return func() Plugin {
		return &testPlugin{config}
	}
}
