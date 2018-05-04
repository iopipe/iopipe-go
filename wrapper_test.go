package iopipe

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"context"
	"time"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"math/rand"
)

// TODO: this is where the fun begins ._.

func TestWrapper_PreInvoke(t *testing.T) {
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

func TestWrapper_Invoke(t *testing.T) {
	Convey("Given a wrapped function", t, func() {
		var (
			receivedContext context.Context
			receivedPayload interface{}
		)

		expectedContext := context.Background()
		// important because lambda defers cancelling the context
		expectedContext, cancelContext := context.WithCancel(expectedContext)

		rand.Seed(time.Now().UTC().UnixNano())
		expectedPayload := rand.Intn(20)
		expectedResponse := rand.Intn(20)
		expectedError := fmt.Errorf("some-error-%d", expectedResponse)

		wrapperHandlerThatPanics := func(ctx context.Context, payload interface{}) (interface{}, error) {
			defer cancelContext()

			panic(fmt.Sprintf("meow-%d", expectedResponse))
			return nil, nil
		}
		wrapperHandlerThatSleeps := func(ctx context.Context, payload interface{}) (interface{}, error) {
			defer cancelContext()

			time.Sleep(50 * time.Millisecond)
			return nil, nil
		}
		wrappedHandler := func(ctx context.Context, payload interface{}) (interface{}, error) {
			defer cancelContext()
			receivedContext = ctx
			receivedPayload = payload

			return expectedResponse, expectedError
		}
		w := wrapper{
			wrappedHandler: wrappedHandler,
		}

		Convey("Captures startTime of the invocation", func() {
			So(w.startTime, ShouldBeZeroValue)
			beforeInvokeTime := time.Now()
			w.Invoke(expectedContext, expectedPayload)
			So(w.startTime, ShouldHappenBetween, beforeInvokeTime, time.Now())
		})

		Convey("Passes arguments unchanged to wrappedHandler", func() {
			w.Invoke(expectedContext, expectedPayload)
			So(receivedContext, ShouldEqual, expectedContext)
			So(receivedPayload, ShouldEqual, expectedPayload)
		})

		Convey("Returns identical values to wrappedHandler", func() {
			actualResponse, actualError := w.Invoke(expectedContext, expectedPayload)
			So(actualResponse, ShouldEqual, expectedResponse)
			So(actualError, ShouldEqual, expectedError)
		})

		Convey("Panics when wrappedHandler panics", func() {
			w.wrappedHandler = wrapperHandlerThatPanics
			So(func() {
				w.Invoke(expectedContext, expectedPayload)
			}, ShouldPanicWith, fmt.Sprintf("meow-%d", expectedResponse))
		})

		Convey("WrappedHandler panic sends report with matching error message", func() {
			var actualMessage string

			w.wrappedHandler = wrapperHandlerThatPanics
			w.reporter = func(report Report) {
				actualMessage = report.Errors.Message
			}

			So(func() {
				w.Invoke(expectedContext, expectedPayload)
			}, ShouldPanic)
			So(actualMessage, ShouldEqual, fmt.Sprintf("meow-%d", expectedResponse))
		})

		Convey("Report deadline subtracts nothing when timeWindow not set", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// deadline is 100 ms from now
			// no timeWindow
			// wrapperHandlerThatSleeps succeeds (TODO: should it success or should it be terminated ?)
			// and report not called
			var reporterCalled bool

			w.deadline = time.Now().Add(100 * time.Millisecond)
			w.wrappedHandler = wrapperHandlerThatSleeps
			w.reporter = func(report Report) {
				reporterCalled = true
			}

			w.Invoke(expectedContext, expectedPayload)
			So(reporterCalled, ShouldBeFalse)
		})

		Convey("No timeout after Invoke succeeds", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// deadline is 100 ms from now
			// no timeWindow
			// after wrappedHandler succeeds 50 ms left for deadline
			// sleep for 60ms more
			// report notcalled
			var reporterCalled bool

			w.deadline = time.Now().Add(100 * time.Millisecond)
			w.wrappedHandler = wrapperHandlerThatSleeps
			w.reporter = func(report Report) {
				reporterCalled = true
			}

			w.Invoke(expectedContext, expectedPayload)
			time.Sleep(60 * time.Millisecond)
			So(reporterCalled, ShouldBeFalse)
		})

		Convey("Report deadline subtracts timeWindow", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// deadline is 100ms from now
			// timeWindow is 60ms
			// report sent with timeout exceeded error

			var actualMessage string

			timeOutWindow := 60 * time.Millisecond
			w.deadline = time.Now().Add(100 * time.Millisecond)
			w.wrappedHandler = wrapperHandlerThatSleeps
			w.agent = &agent{AgentConfig: &AgentConfig{
				TimeoutWindow: &timeOutWindow,
			}}
			w.reporter = func(report Report) {
				actualMessage = report.Errors.Message
			}

			w.Invoke(expectedContext, expectedPayload)
			So(actualMessage, ShouldEqual, "timeout exceeded")
		})

		Convey("Unset deadline results no timeout regardless of timeOutWindow", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// no deadline
			// timeWindow is 2000 ms
			// no report sent
			var reporterCalled bool

			timeOutWindow := 2000 * time.Millisecond
			w.wrappedHandler = wrapperHandlerThatSleeps
			w.deadline = time.Time{}
			w.agent = &agent{AgentConfig: &AgentConfig{
				TimeoutWindow: &timeOutWindow,
			}}
			w.reporter = func(report Report) {
				reporterCalled = true
			}

			w.Invoke(expectedContext, expectedPayload)
			time.Sleep(50 * time.Millisecond)
			So(reporterCalled, ShouldBeFalse)
		})

	})
}

func TestWrapper_RunHook(t *testing.T) {
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
