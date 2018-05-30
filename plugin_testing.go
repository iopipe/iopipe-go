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
		LastHook string `json:"lastHook"`
	}{
		Name:     p.Name(),
		LastHook: p.LastHook,
	}
}

func (p *testPlugin) Name() string {
	return "test-plugin"
}

// TestPlugin return a test plugin
func TestPlugin(config TestPluginConfig) PluginInstantiator {
	return func(w *HandlerWrapper) Plugin {
		return &testPlugin{config}
	}
}
