package iopipe

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"context"
	"time"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// TODO: this is where the fun begins ._.

func TestPreInvoke(t *testing.T) {
	Convey("Captures context and runs pre-invoke hook", t, func() {
		w := wrapper{
			plugins: []Plugin{
				&testPlugin{},
				&testPlugin{},
			},
		}

		expectedDeadline := time.Now().Add(2 * time.Hour)
		expectedLC := lambdacontext.LambdaContext{
			AwsRequestID:       "a",
			InvokedFunctionArn: "b",
			Identity: lambdacontext.CognitoIdentity{
				CognitoIdentityID:     "c",
				CognitoIdentityPoolID: "d",
			},
			ClientContext: lambdacontext.ClientContext{
				Client: lambdacontext.ClientApplication{
					InstallationID: "e",
					AppTitle:       "f",
					AppVersionCode: "g",
					AppPackageName: "h",
				},
			},
		}

		ctx := context.Background()
		ctx = lambdacontext.NewContext(ctx, &expectedLC)
		ctx, cancel := context.WithDeadline(ctx, expectedDeadline)
		defer cancel()
		w.PreInvoke(ctx)

		Convey("Wrapper lambdaContext matches input context", func() {
			So(*w.lambdaContext, ShouldResemble, expectedLC)
		})

		Convey("Wrapper deadline matches input context", func() {
			So(w.deadline, ShouldEqual, expectedDeadline)
		})

		Convey("Pre-invoke hook is called on all plugins", func() {
			for _, plugin := range w.plugins  {
				plugin, _ := plugin.(*testPlugin)
				So(plugin.LastHook, ShouldEqual, HOOK_PRE_INVOKE)
			}
		})
	})
}

func TestRunHook(t *testing.T) {
	Convey("Runs hooks on all the plugins", t, func() {
		plugin1 := &testPlugin{}
		plugin2 := &testPlugin{}

		w := wrapper{
			plugins: []Plugin{
				plugin1,
				plugin2,
			},
		}

		Convey("All plugins should receive RunHook", func() {
			w.RunHook("test-hook")

			So(plugin1.LastHook, ShouldEqual, "test-hook")
			So(plugin2.LastHook, ShouldEqual, "test-hook")
		})

		Convey("Nil plugin does not result in panic", func() {
			w.plugins = append(w.plugins, nil)

			So(func() {
				w.RunHook("not-test-hook")
			}, ShouldNotPanic)
			So(plugin1.LastHook, ShouldEqual, "not-test-hook")
			So(plugin2.LastHook, ShouldEqual, "not-test-hook")
		})


	})
}
