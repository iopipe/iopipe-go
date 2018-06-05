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

// Invoke invokes the wrapped handler, handling panics and timeouts
func (hw *HandlerWrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	hw.deadline, _ = ctx.Deadline()
	hw.report = NewReport(hw)

	lc, _ := lambdacontext.FromContext(ctx)
	hw.lambdaContext = lc

	cw := NewContextWrapper(lc, hw)
	ctx = NewContext(ctx, cw)

	hw.preInvoke(ctx, payload)

	// Handle and report a panic if it occurs
	defer func() {
		if panicErr := recover(); panicErr != nil {
			hw.postInvoke(ctx, payload)
			hw.Error(NewPanicInvocationError(panicErr))
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
			fmt.Println("Function is about to timeout, sending report")
			hw.postInvoke(ctx, payload)
			hw.Error(fmt.Errorf("timeout exceeded"))
		case <-ctx.Done():
			if hw.report != nil && !hw.report.sending {
				hw.postInvoke(ctx, payload)
				hw.report.prepare(nil)
				hw.preReport(hw.report)
				hw.report.send()
				hw.postReport(hw.report)
			}
		}
	}()

	response, err = hw.wrappedHandler(ctx, payload)

	ColdStart = false

	if hw.report != nil && !hw.report.sending {
		hw.postInvoke(ctx, payload)
		hw.report.prepare(err)
		hw.preReport(hw.report)
		hw.report.send()
		hw.postReport(hw.report)
	}

	return response, err
}

// Label adds a label to the report
// Error adds an  error to the report
func (hw *HandlerWrapper) Error(err error) {
	if hw.report == nil {
		fmt.Println("Attempting to add error before function decorated with IOpipe. This error will not be recorded.")
		return
	}

	if !hw.report.sending {
		hw.report.prepare(err)
		hw.preReport(hw.report)
		hw.report.send()
		hw.postReport(hw.report)
	}
}

func (hw *HandlerWrapper) Label(name string) {
	if hw.report == nil {
		fmt.Println("Attempting to add labels before function decorated with IOpipe. This metric will not be recorded.")
		return
	}

	if utf8.RuneCountInString(name) > 128 {
		fmt.Println(fmt.Sprintf("Label name %s is longer than allowed limit of 128 characters. This label will not be recorded.", name))
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
		fmt.Println("Attempting to add metrics before function decorated with IOpipe. This metric will not be recorded.")
		return
	}

	if utf8.RuneCountInString(name) > 128 {
		fmt.Println(fmt.Sprintf("Metric of name %s is longer than allowed limit of 128 characters. This metric will not be recorded.", name))
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
