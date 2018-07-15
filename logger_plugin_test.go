package iopipe

import (
	"context"
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

		a.Reporter = func(report *Report) error {
			return nil
		}

		handler := func(ctx context.Context, payload interface{}) (interface{}, error) {
			context, _ := FromContext(ctx)
			context.IOpipe.Log.Debug("Some debug message")

			return nil, nil
		}

		hw := NewHandlerWrapper(handler, a)

		Convey("Logger plugin invoke hooks fired", func() {
			ctx := context.Background()
			hw.Invoke(ctx, nil)
		})
	})
}
