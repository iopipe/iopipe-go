package iopipe

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlerWrapper_NewHandlerWrapper(t *testing.T) {
	Convey("Should wrap a handler function", t, func() {
		agent := NewAgent(Config{})
		handler := func(ctx context.Context, payload interface{}) (interface{}, error) {
			return "hi there", nil
		}
		wrappedHandler := newHandler(handler)
		handlerWrapper := NewHandlerWrapper(handler, agent)

		So(handlerWrapper.agent, ShouldResemble, agent)
		So(handlerWrapper.originalHandler, ShouldHaveSameTypeAs, handler)
		So(handlerWrapper.wrappedHandler, ShouldHaveSameTypeAs, wrappedHandler)
	})
}

func TestHandlerWrapper_Error(t *testing.T) {
	Convey("A handler wrapper allows errors to be added to a report", t, func() {
		a := NewAgent(Config{})
		hw := &HandlerWrapper{agent: a}

		Convey("Doesnot panic if there is no report", func() {
			So(hw.report, ShouldBeNil)
			So(func() {
				hw.Error(fmt.Errorf("whoops"))
			}, ShouldNotPanic)
		})

		Convey("Add an error to the report", func() {
			r := NewReport(hw)
			hw.report = r

			So(r.Errors, ShouldResemble, &struct{}{})
			hw.Error(fmt.Errorf("Whoops"))
			So(r.Errors, ShouldHaveSameTypeAs, &InvocationError{})
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

		a := NewAgent(Config{
			PluginInstantiators: []PluginInstantiator{
				TestPlugin(TestPluginConfig{}),
			},
		})
		a.Reporter = func(report *Report) error {
			return nil
		}

		hw := NewHandlerWrapper(wrappedHandler, a)

		Convey("Pre and post invoke hooks are called", func() {
			plugin, _ := a.plugins[0].(*testPlugin)

			So(plugin.preInvokeCalled, ShouldEqual, 0)
			So(plugin.postInvokeCalled, ShouldEqual, 0)

			hw.Invoke(expectedContext, expectedPayload)

			So(plugin.preInvokeCalled, ShouldEqual, 1)
			So(plugin.postInvokeCalled, ShouldEqual, 1)
		})

		Convey("Passes context and payload to wrappedHandler", func() {
			hw.Invoke(expectedContext, expectedPayload)
			lc, _ := lambdacontext.FromContext(expectedContext)
			cw := NewContextWrapper(lc, hw)
			iopipeContext := NewContext(expectedContext, cw)

			So(receivedContext, ShouldResemble, iopipeContext)
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
			a.Reporter = func(report *Report) error {
				actualMessage = report.Errors.(*InvocationError).Message
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
			// wrapperHandlerThatSleeps succeeds
			var reporterCalled bool

			deadlineContext, deadlineCancel := context.WithDeadline(expectedContext, time.Now().Add(100*time.Millisecond))
			defer deadlineCancel()
			hw.wrappedHandler = wrapperHandlerThatSleeps
			a.Reporter = func(report *Report) error {
				reporterCalled = true
				return nil
			}

			hw.Invoke(deadlineContext, expectedPayload)

			So(reporterCalled, ShouldBeTrue)
		})

		Convey("No timeout after Invoke succeeds", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// deadline is 100 ms from now
			// no timeWindow
			// after wrappedHandler succeeds 50 ms left for deadline
			// sleep for 60ms more
			var reporterCalled bool

			deadlineContext, deadlineCancel := context.WithDeadline(expectedContext, time.Now().Add(100*time.Millisecond))
			defer deadlineCancel()
			hw.wrappedHandler = wrapperHandlerThatSleeps
			a.Reporter = func(report *Report) error {
				reporterCalled = true
				return nil
			}

			hw.Invoke(deadlineContext, expectedPayload)

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

			deadlineContext, deadlineCancel := context.WithDeadline(expectedContext, time.Now().Add(100*time.Millisecond))
			defer deadlineCancel()
			hw.wrappedHandler = wrapperHandlerThatSleeps

			a.TimeoutWindow = &timeOutWindow
			a.Reporter = func(report *Report) error {
				actualMessage = report.Errors.(*InvocationError).Message
				return nil
			}

			hw.Invoke(deadlineContext, expectedPayload)

			So(actualMessage, ShouldEqual, "Timeout Exceeded")
		})

		Convey("Unset deadline results no timeout regardless of timeOutWindow", func() {
			// wrapperHandlerThatSleeps takes 50ms
			// no deadline
			// timeWindow is 2000 ms
			var (
				reportError    interface{}
				reporterCalled bool
			)

			timeOutWindow := 2000 * time.Millisecond

			hw.wrappedHandler = wrapperHandlerThatSleeps

			a.TimeoutWindow = &timeOutWindow
			a.Reporter = func(report *Report) error {
				reportError = report.Errors
				reporterCalled = true

				return nil
			}

			hw.Invoke(expectedContext, expectedPayload)

			time.Sleep(50 * time.Millisecond)

			So(reporterCalled, ShouldBeTrue)
			So(reportError, ShouldResemble, &struct{}{})
		})
	})
}

func TestHandlerWrapper_Label(t *testing.T) {
	Convey("A handler wrapper allows labels to be added to a report", t, func() {
		a := NewAgent(Config{})
		hw := &HandlerWrapper{agent: a}

		Convey("Doesnot panic if there is no report", func() {

			So(hw.report, ShouldBeNil)
			So(func() {
				hw.Label("foo")
			}, ShouldNotPanic)
		})

		Convey("Add a label to the report", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(r.labels), ShouldEqual, 0)
			hw.Label("foobar")
			So(len(r.labels), ShouldEqual, 1)

			r.prepare(nil)
			So(len(r.Labels), ShouldEqual, 1)
		})

		Convey("Does not add label if name is too long", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(r.labels), ShouldEqual, 0)
			hw.Label(strings.Repeat("X", 129))
			So(len(r.labels), ShouldEqual, 0)
		})
	})
}

func TestHandlerWrapper_Metric(t *testing.T) {
	Convey("A handler wrapper allows custom metrics to be added to a report", t, func() {
		a := NewAgent(Config{})
		hw := &HandlerWrapper{agent: a}

		Convey("Doesnot panic if there is no report", func() {
			So(hw.report, ShouldBeNil)
			So(func() {
				hw.Metric("foo", "bar")
			}, ShouldNotPanic)
		})

		Convey("Add a custom string metric to the report", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
			hw.Metric("foo", "bar")
			So(len(hw.report.CustomMetrics), ShouldEqual, 1)
		})

		Convey("Add a custom numeric metric to the report", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
			hw.Metric("meaning of life", 42)
			So(len(hw.report.CustomMetrics), ShouldEqual, 1)
		})

		Convey("Does not add metric if name is too long", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
			hw.Metric(strings.Repeat("X", 129), "bar")
			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
		})

		Convey("Does not add metric if value is not string or number", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
			hw.Metric("foo", true)
			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
		})
	})
}
