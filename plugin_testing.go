package iopipe

import "context"

// TestPluginConfig is a test plugin config
type TestPluginConfig struct {
	lastHook         string
	preSetupCalled   int
	postSetupCalled  int
	preInvokeCalled  int
	postInvokeCalled int
	preReportCalled  int
	postReportCalled int
}

type testPlugin struct {
	TestPluginConfig
}

func (p *testPlugin) Meta() *PluginMeta {
	return &PluginMeta{
		Name:     "@iopipe/test-plugin",
		Version:  "0.1.0",
		Homepage: "https://github.com/iopipe/iopipe-go",
		Enabled:  p.Enabled(),
	}
}

func (p *testPlugin) Enabled() bool {
	return true
}

func (p *testPlugin) PreSetup(agent *Agent) {
	p.preSetupCalled++
}

func (p *testPlugin) PostSetup(agent *Agent) {
	p.postSetupCalled++
}

func (p *testPlugin) PreInvoke(ctx context.Context, payload interface{}) {
	p.preInvokeCalled++
}

func (p *testPlugin) PostInvoke(ctx context.Context, payload interface{}) {
	p.postInvokeCalled++
}

func (p *testPlugin) PreReport(report *Report) {
	p.preReportCalled++
}

func (p *testPlugin) PostReport(report *Report) {
	p.postReportCalled++
}

// TestPlugin returns a test plugin
func TestPlugin(config TestPluginConfig) PluginInstantiator {
	return func() Plugin {
		return &testPlugin{config}
	}
}
