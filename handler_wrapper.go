package iopipe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// HandlerWrapper is the IOpipe handler wrapper
type HandlerWrapper struct {
	report          *Report
	reporter        Reporter
	originalHandler interface{}
	wrappedHandler  lambdaHandler
	startTime       time.Time
	endTime         time.Time
	deadline        time.Time
	lambdaContext   *lambdacontext.LambdaContext
	reportSending   bool
	agent           *Agent
	plugins         []Plugin
}

// NewHandlerWrapper creates a new IOpipe handler wrapper
func NewHandlerWrapper(handler interface{}, agentInstance *Agent) *HandlerWrapper {
	hw := &HandlerWrapper{
		originalHandler: handler,
		wrappedHandler:  newHandler(handler),
		agent:           agentInstance,
	}

	// TODO: Plugins should be loaded at agent level and only once
	var plugins []Plugin
	pluginInstantiators := agentInstance.PluginInstantiators

	if pluginInstantiators != nil {
		plugins = make([]Plugin, len(pluginInstantiators))
		for index, pluginInstantiator := range pluginInstantiators {
			plugins[index] = pluginInstantiator(hw)
		}
	}

	hw.plugins = plugins
	hw.reporter = sendReport

	// TODO: This is supposed to happen during agent init
	hw.RunHook(HookPreSetup)
	hw.RunHook(HookPostSetup)

	return hw
}

// PreInvoke runs pre invoke hooks
func (hw *HandlerWrapper) PreInvoke(context context.Context) {
	hw.deadline, _ = context.Deadline()
	hw.lambdaContext, _ = lambdacontext.FromContext(context)
	hw.RunHook(HookPreInvoke)
}

// RunHook runs the specified hooks
func (hw *HandlerWrapper) RunHook(hook string) {
	var wg sync.WaitGroup
	wg.Add(len(hw.plugins))

	for _, plugin := range hw.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil {
				plugin.RunHook(hook)
			}
		}(plugin)
	}

	wg.Wait()
}

// Invoke invokes the wrapped handler, handling panics and timeouts
func (hw *HandlerWrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	hw.startTime = time.Now()
	hw.report = NewReport(hw)

	// Handle and report a panic if it occurs
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// the stack trace the client gets will be a bit verbose with mentions of the wrapper
			hw.PostInvoke(NewPanicInvocationError(panicErr))
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
			hw.PostInvoke(fmt.Errorf("timeout exceeded"))
		case <-ctx.Done():
			hw.PostInvoke(nil)
		}
	}()

	return hw.wrappedHandler(ctx, payload)
}

// PostInvoke prepares a report and sends it
func (hw *HandlerWrapper) PostInvoke(err error) {
	if hw.reportSending {
		return
	}

	hw.reportSending = true
	hw.endTime = time.Now()
	hw.RunHook(HookPostInvoke)

	var (
		ok   bool
		hErr *InvocationError
	)

	if hErr, ok = err.(*InvocationError); !ok {
		hErr = NewInvocationError(err)
	}

	hw.report.prepare(hErr)
	hw.RunHook(HookPreReport)

	// PostInvoke
	if hw.reporter != nil {
		err := hw.reporter(hw.report)
		if err != nil {
			// TODO: We want to log an error during reporting
			fmt.Println("Reporting errored: ", err)
		}
	}

	hw.RunHook(HookPostReport)
}

func wrapHandler(handler interface{}, agentInstance *Agent) lambdaHandler {
	// decorate the handler
	return func(context context.Context, payload interface{}) (interface{}, error) {
		handlerWrapper := NewHandlerWrapper(handler, agentInstance)
		handlerWrapper.PreInvoke(context)
		response, err := handlerWrapper.Invoke(context, payload)
		handlerWrapper.PostInvoke(err)

		ColdStart = false

		return response, err
	}
}
