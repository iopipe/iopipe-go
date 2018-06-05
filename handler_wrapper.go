package iopipe

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// HandlerWrapper is the IOpipe handler wrapper
type HandlerWrapper struct {
	agent           *Agent
	deadline        time.Time
	lambdaContext   *lambdacontext.LambdaContext
	originalHandler interface{}
	report          *Report
	reportSending   bool
	wrappedContext  *ContextWrapper
	wrappedHandler  lambdaHandler
}

// NewHandlerWrapper creates a new IOpipe handler wrapper
func NewHandlerWrapper(handler interface{}, agent *Agent) *HandlerWrapper {
	return &HandlerWrapper{
		agent:           agent,
		originalHandler: handler,
		wrappedHandler:  newHandler(handler),
	}
}

// Invoke invokes the wrapped handler
// Invoke invokes the wrapped handler, handling panics and timeouts
func (hw *HandlerWrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	hw.deadline, _ = ctx.Deadline()
	hw.lambdaContext, _ = lambdacontext.FromContext(ctx)
	hw.wrappedContext = NewContextWrapper(hw.lambdaContext, hw)
	hw.report = NewReport(hw)
	hw.preInvoke(hw.wrappedContext, payload)

	// Handle and report a panic if it occurs
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// the stack trace the client gets will be a bit verbose with mentions of the wrapper
			hw.postInvoke(hw.wrappedContext, payload)
			hw.report.prepare(NewPanicInvocationError(panicErr))
			hw.preReport(hw.report)
			hw.preReport(hw.report)
			hw.report.send()
			hw.postReport(hw.report)
			panic(panicErr)
		}
	}()

	// Start the timeout clock and handle timeouts
	go func() {
		timeoutWindow := 0 * time.Millisecond
		if hw.agent != nil && hw.agent.TimeoutWindow != nil {
			timeoutWindow = *hw.agent.TimeoutWindow
		}

		if hw.deadline.IsZero() {
			return
		}

		timeoutChannel := time.After(time.Until(hw.deadline.Add(-timeoutWindow)))

		select {
		// We have timed out
		case <-timeoutChannel:
			// Report the timeout
			hw.postInvoke(hw.wrappedContext, payload)
			hw.report.prepare(fmt.Errorf("timeout exceeded"))
			hw.preReport(hw.report)
			hw.report.send()
			hw.postReport(hw.report)
		case <-ctx.Done():
			hw.postInvoke(hw.wrappedContext, payload)
			hw.report.prepare(nil)
			hw.preReport(hw.report)
			hw.report.send()
			hw.postReport(hw.report)
		}
	}()

	response, err = hw.wrappedHandler(ctx, payload)

	ColdStart = false

	hw.postInvoke(hw.wrappedContext, payload)
	hw.report.prepare(err)
	hw.preReport(hw.report)
	hw.report.send()
	hw.postReport(hw.report)

	return response, err
}

// Label adds a label to the report
func (hw *HandlerWrapper) Label(name string) {
	if hw.report == nil {
		// TODO: Do the equivalent of a warning here
		return
	}

	if utf8.RuneCountInString(name) > 128 {
		// TODO: Do the equivalent of a warning here
		return
	}

	if _, ok := hw.report.labels[name]; !ok {
		hw.report.labels[name] = struct{}{}
	}
}

// Metric adds a custom metric to the report
func (hw *HandlerWrapper) Metric(name string, value interface{}) {
	if hw.report == nil {
		// TODO: Do the equivalent of a warning here
		return
	}

	if utf8.RuneCountInString(name) > 128 {
		// TODO: Do the equivalent of a warning here
		return
	}

	s := coerceString(value)
	if s != nil {
		hw.report.CustomMetrics = append(hw.report.CustomMetrics, CustomMetric{Name: name, S: s})
	}

	n := coerceNumeric(value)
	if n != nil {
		hw.report.CustomMetrics = append(hw.report.CustomMetrics, CustomMetric{Name: name, N: n})
	}
}

func (hw *HandlerWrapper) preInvoke(ctx *ContextWrapper, payload interface{}) {
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

func (hw *HandlerWrapper) postInvoke(ctx *ContextWrapper, payload interface{}) {
	var wg sync.WaitGroup
	wg.Add(len(hw.agent.plugins))

	for _, plugin := range hw.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.PostInvoke(ctx, payload)
			}
		}(plugin)
	}

	wg.Wait()
}

func (hw *HandlerWrapper) preReport(report *Report) {
	var wg sync.WaitGroup
	wg.Add(len(hw.agent.plugins))

	for _, plugin := range hw.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.PreReport(report)
			}
		}(plugin)
	}

	wg.Wait()
}

func (hw *HandlerWrapper) postReport(report *Report) {
	var wg sync.WaitGroup
	wg.Add(len(hw.agent.plugins))

	for _, plugin := range hw.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.PostReport(report)
			}
		}(plugin)
	}

	wg.Wait()
}
