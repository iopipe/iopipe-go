package iopipe

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	. "github.com/smartystreets/goconvey/convey"
)

// TODO: this is where the fun begins ._.

func TestHandlerWrapper_PreInvoke(t *testing.T) {
	Convey("Captures context and runs pre-invoke hook", t, func() {
		hw := HandlerWrapper{
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
		hw.PreInvoke(ctx)

		Convey("Wrapper lambdaContext matches input context", func() {
			So(*hw.lambdaContext, ShouldResemble, expectedLC)
		})

		Convey("Wrapper deadline matches input context", func() {
			So(hw.deadline, ShouldEqual, expectedDeadline)
		})

		Convey("Pre-invoke hook is called on all plugins", func() {
			for _, plugin := range hw.plugins {
				plugin, _ := plugin.(*testPlugin)
				So(plugin.LastHook, ShouldEqual, HookPreInvoke)
			}
		})
	})
}

func TestHandlerWrapper_Invoke(t *testing.T) {
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

		hw := HandlerWrapper{
			wrappedHandler: wrappedHandler,
		}

		Convey("Captures startTime of the invocation", func() {
			So(hw.startTime, ShouldBeZeroValue)
			beforeInvokeTime := time.Now()
			hw.Invoke(expectedContext, expectedPayload)
			So(hw.startTime, ShouldHappenBetween, beforeInvokeTime, time.Now())
		})

		Convey("Passes arguments unchanged to wrappedHandler", func() {
			hw.Invoke(expectedContext, expectedPayload)
			So(receivedContext, ShouldEqual, expectedContext)
			So(receivedPayload, ShouldEqual, expectedPayload)
		})

		Convey("Returns identical values to wrappedHandler", func() {
			actualResponse, actualError := hw.Invoke(expectedContext, expectedPayload)
			So(actualResponse, ShouldEqual, expectedResponse)
			So(actualError, ShouldEqual, expectedError)
		})

		Convey("Panics when wrappedHandler panics", func() {
			hw.wrappedHandler = wrapperHandlerThatPanics
			So(func() {
				hw.Invoke(expectedContext, expectedPayload)
			}, ShouldPanicWith, fmt.Sprintf("meow-%d", expectedResponse))
		})

		Convey("WrappedHandler panic sends report with matching error message", func() {
			var actualMessage string

			hw.wrappedHandler = wrapperHandlerThatPanics
			hw.reporter = func(report *Report) error {
				actualMessage = report.Errors.Message
				return nil
			}

			So(func() {
				hw.Invoke(expectedContext, expectedPayload)
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

			hw.deadline = time.Now().Add(100 * time.Millisecond)
			hw.wrappedHandler = wrapperHandlerThatSleeps
			hw.reporter = func(report *Report) error {
				reporterCalled = true
				return nil
			}
			hw.Invoke(expectedContext, expectedPayload)

			So(reporterCalled, ShouldBeFalse)
		})

		Convey("No timeout after Invoke succeeds", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// deadline is 100 ms from now
			// no timeWindow
			// after wrappedHandler succeeds 50 ms left for deadline
			// sleep for 60ms more
			var reporterCalled bool

			hw.deadline = time.Now().Add(100 * time.Millisecond)
			hw.wrappedHandler = wrapperHandlerThatSleeps
			hw.reporter = func(report *Report) error {
				reporterCalled = true
				return nil
			}
			hw.Invoke(expectedContext, expectedPayload)

			time.Sleep(60 * time.Millisecond)

			So(reporterCalled, ShouldBeTrue)
		})

		Convey("Report deadline subtracts timeWindow", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// deadline is 100ms from now
			// timeWindow is 60ms
			// report sent with timeout exceeded error

			var actualMessage string

			timeOutWindow := 60 * time.Millisecond
			hw.deadline = time.Now().Add(100 * time.Millisecond)
			hw.wrappedHandler = wrapperHandlerThatSleeps
			hw.agent = &Agent{Config: &Config{
				TimeoutWindow: &timeOutWindow,
			}}
			hw.reporter = func(report *Report) error {
				actualMessage = report.Errors.Message
				return nil
			}

			hw.Invoke(expectedContext, expectedPayload)
			hw.PostInvoke(nil)

			So(actualMessage, ShouldEqual, "timeout exceeded")
		})

		Convey("Unset deadline results no timeout regardless of timeOutWindow", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// no deadline
			// timeWindow is 2000 ms
			// no report sent
			var reporterCalled bool

			timeOutWindow := 2000 * time.Millisecond

			hw.wrappedHandler = wrapperHandlerThatSleeps
			hw.deadline = time.Time{}
			hw.agent = &Agent{Config: &Config{
				TimeoutWindow: &timeOutWindow,
			}}
			hw.reporter = func(report *Report) error {
				reporterCalled = true
				return nil
			}
			hw.Invoke(expectedContext, expectedPayload)

			time.Sleep(50 * time.Millisecond)

			So(reporterCalled, ShouldBeFalse)
		})
	})
}

func TestHandlerWrapper_RunHook(t *testing.T) {
	Convey("Runs hooks on all the plugins", t, func() {
		plugin1 := &testPlugin{}
		plugin2 := &testPlugin{}

		hw := HandlerWrapper{
			plugins: []Plugin{
				plugin1,
				plugin2,
			},
		}

		Convey("All plugins should receive RunHook", func() {
			hw.RunHook("test-hook")

			So(plugin1.LastHook, ShouldEqual, "test-hook")
			So(plugin2.LastHook, ShouldEqual, "test-hook")
		})

		Convey("Nil plugin does not result in panic", func() {
			hw.plugins = append(hw.plugins, nil)

			So(func() {
				hw.RunHook("not-test-hook")
			}, ShouldNotPanic)
			So(plugin1.LastHook, ShouldEqual, "not-test-hook")
			So(plugin2.LastHook, ShouldEqual, "not-test-hook")
		})

	})
}
