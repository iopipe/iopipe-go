package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoggerPlugin_LoggerPlugin(t *testing.T) {
	Convey("Logger plugin should be initialized by agent", t, func() {
		a := NewAgent(Config{
			Plugins: []PluginInstantiator{
				LoggerPlugin(LoggerPluginConfig{}),
			},
		})

		So(len(a.plugins), ShouldEqual, 1)
	})
}
