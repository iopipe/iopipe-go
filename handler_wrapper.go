package iopipe

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/lambdacontext"
	log "github.com/sirupsen/logrus"
)

// HandlerWrapper is the IOpipe handler wrapper
type HandlerWrapper struct {
	agent           *Agent
	deadline        time.Time
	lambdaContext   *lambdacontext.LambdaContext
	originalHandler interface{}
	report          *Report
	wrappedHandler  lambdaHandler

	Log *log.Logger
}

// NewHandlerWrapper creates a new IOpipe handler wrapper
func NewHandlerWrapper(handler interface{}, agent *Agent) *HandlerWrapper {
	return &HandlerWrapper{
		agent:           agent,
		originalHandler: handler,
		wrappedHandler:  newHandler(handler),
		Log:             agent.log,
	}
}

// Invoke invokes the wrapped handler, handling panics and timeouts
func (hw *HandlerWrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	lc, _ := lambdacontext.FromContext(ctx)
	hw.lambdaContext = lc

	cw := NewContextWrapper(lc, hw)
	ctx = NewContext(ctx, cw)
	hw.deadline, _ = ctx.Deadline()

	hw.report = NewReport(hw)

	hw.preInvoke(ctx, payload)

	// Handle and report a panic if it occurs
	defer func() {
		if panicErr := recover(); panicErr != nil {
			hw.Label("@iopipe/error")
			hw.report.prepare(NewPanicInvocationError(panicErr))
			hw.report.send()
			panic(panicErr)
		}
	}()

	// Start the timeout clock and handle timeouts
	go func() {
		if hw.deadline.IsZero() {
			hw.Log.Debug("Deadline is zero, disabling timeout handling")
			return
		}

		timeoutWindow := 0 * time.Millisecond

		if hw.agent != nil && hw.agent.TimeoutWindow != nil {
			timeoutWindow = *hw.agent.TimeoutWindow
		}

		timeoutDuration := hw.deadline.Add(-timeoutWindow)

		// If timeout duration is in the past, disable timeout handling
		if time.Now().After(timeoutDuration) {
			hw.Log.Debug("Timeout deadline is in the past, disabling timeout handling")
			return
		}

		hw.Log.Debug("Setting function to timeout in ", time.Until(timeoutDuration).String())

		timeoutChannel := time.After(time.Until(timeoutDuration))

		select {
		// We're within the timeout window
		case <-timeoutChannel:
			hw.Log.Debug("Function is about to timeout, sending report")
			hw.Label("@iopipe/timeout")
			hw.report.prepare(fmt.Errorf("Timeout Exceeded"))
			hw.report.send()
			return
		case <-ctx.Done():
			return
		}
	}()

	response, err = hw.wrappedHandler(ctx, payload)

	if coldStart {
		hw.Label("@iopipe/coldstart")
	}

	coldStart = false

	if err != nil {
		hw.Label("@iopipe/error")
	}

	hw.postInvoke(ctx, payload)

	if hw.report != nil {
		hw.report.prepare(err)
		hw.report.send()
	}

	return response, err
}

// Error adds an  error to the report
func (hw *HandlerWrapper) Error(err error) {
	if hw.report == nil {
		hw.Log.Warn("Attempting to add error before function decorated with IOpipe. This error will not be recorded.")
		return
	}

	hw.Label("@iopipe/error")
	hw.report.prepare(err)
	hw.report.send()
}

// Label adds a label to the report
func (hw *HandlerWrapper) Label(name string) {
	if hw.report == nil {
		hw.Log.Warn("Attempting to add labels before function decorated with IOpipe. This metric will not be recorded.")
		return
	}

	if utf8.RuneCountInString(name) > 128 {
		hw.Log.Warn(fmt.Sprintf("Label name %s is longer than allowed limit of 128 characters. This label will not be recorded.", name))
		return
	}

	// USing map to ensure uniqueness of labels
	if _, ok := hw.report.labels[name]; !ok {
		hw.report.labels[name] = struct{}{}
	}
}

// Metric adds a custom metric to the report
func (hw *HandlerWrapper) Metric(name string, value interface{}) {
	if hw.report == nil {
		hw.Log.Warn("Attempting to add metrics before function decorated with IOpipe. This metric will not be recorded.")
		return
	}

	if utf8.RuneCountInString(name) > 128 {
		hw.Log.Warn(fmt.Sprintf("Metric of name %s is longer than allowed limit of 128 characters. This metric will not be recorded.", name))
		return
	}

	s := coerceString(value)
	if s != nil {
		hw.Label("@iopipe/metrics")
		hw.report.CustomMetrics = append(hw.report.CustomMetrics, CustomMetric{Name: name, S: s})
	}

	n := coerceNumeric(value)
	if n != nil {
		hw.Label("@iopipe/metrics")
		hw.report.CustomMetrics = append(hw.report.CustomMetrics, CustomMetric{Name: name, N: n})
	}
}

// preInvoke runs the PreInvoke hooks
func (hw *HandlerWrapper) preInvoke(ctx context.Context, payload interface{}) {
	var wg sync.WaitGroup
	wg.Add(len(hw.agent.plugins))

	for _, plugin := range hw.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.PreInvoke(ctx, payload)
			}
		}(plugin)
	}

	wg.Wait()
}

// postInvoke runs the PostInvoke hooks
func (hw *HandlerWrapper) postInvoke(ctx context.Context, payload interface{}) {
	var wg sync.WaitGroup
	wg.Add(len(hw.agent.plugins))

	for _, plugin := range hw.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil && plugin.Enabled() {
				plugin.PostInvoke(ctx, payload)
			}
		}(plugin)
	}

	wg.Wait()
}
