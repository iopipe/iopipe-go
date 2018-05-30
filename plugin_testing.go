package iopipe

// TestPluginConfig is a test plugin config
type TestPluginConfig struct {
	LastHook        string
	wrapperInstance HandlerWrapper
}

type testPlugin struct {
	TestPluginConfig
}

func (p *testPlugin) RunHook(hook string) {
	p.LastHook = hook
}

func (p *testPlugin) Meta() interface{} {
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

// TestPlugin return a test plugin
func TestPlugin(config TestPluginConfig) PluginInstantiator {
	return func(w *HandlerWrapper) Plugin {
		return &testPlugin{config}
	}
}
