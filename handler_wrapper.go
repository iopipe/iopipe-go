package iopipe

import (
	"context"
	"fmt"
	"os"
	"runtime"
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
	hw.reporter = reportToIOpipe

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

// Invoke invokes the wrapped handler
func (hw *HandlerWrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	hw.startTime = time.Now()
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// the stack trace the client gets will be a bit verbose with mentions of the wrapper
			hw.PostInvoke(NewPanicInvocationError(panicErr))
			panic(panicErr)
		}
	}()

	go func() {
		timeoutWindow := 0 * time.Millisecond
		if hw.agent != nil && hw.agent.TimeoutWindow != nil {
			timeoutWindow = *hw.agent.TimeoutWindow
		}

		if hw.deadline.IsZero() {
			return
		}

		select {
		// naturally the deadline should occur before the context is closed
		case <-time.After(time.Until(hw.deadline.Add(-timeoutWindow))):
			hw.PostInvoke(fmt.Errorf("timeout exceeded"))
		case <-ctx.Done():
			return // returning not to leak the goroutine
		}
	}()

	return hw.wrappedHandler(ctx, payload)
}

// PostInvoke runs the post invoke hooks
func (hw *HandlerWrapper) PostInvoke(err error) {
	if hw.reportSending {
		return
	}

	hw.reportSending = true
	hw.RunHook(HookPostInvoke)
	hw.RunHook(HookPreReport)
	hw.endTime = time.Now()

	var (
		ok   bool
		hErr *InvocationError
	)

	if hErr, ok = err.(*InvocationError); !ok {
		hErr = NewInvocationError(err)
	}

	hw.prepareReport(hErr)

	// PostInvoke
	if hw.reporter != nil {
		err := hw.reporter(hw.report)
		if err != nil {
			// TODO: Don't we want to pass the error on to AWS?
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

func (hw *HandlerWrapper) prepareReport(invErr *InvocationError) {
	if hw.report != nil {
		return
	}

	startTime := hw.startTime
	endTime := hw.endTime
	deadline := hw.deadline

	lc := hw.lambdaContext
	if lc == nil {
		lc = &lambdacontext.LambdaContext{
			AwsRequestID: "ERROR",
		}
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pluginsMeta := make([]interface{}, len(hw.plugins))
	for index, plugin := range hw.plugins {
		pluginsMeta[index] = plugin.Meta()
	}

	var errs interface{}
	errs = &struct{}{}
	if invErr != nil {
		errs = invErr
	}

	token := ""
	if hw.agent != nil && hw.agent.Token != nil {
		token = *hw.agent.Token
	}

	hw.report = &Report{
		ClientID:      token,
		InstallMethod: "manual",
		Duration:      int(endTime.Sub(startTime).Nanoseconds()),
		ProcessID:     ProcessID,
		Timestamp:     int(startTime.UnixNano() / 1e6),
		TimestampEnd:  int(endTime.UnixNano() / 1e6),
		AWS: AWSReportDetails{
			FunctionName:             lambdacontext.FunctionName,
			FunctionVersion:          lambdacontext.FunctionVersion,
			AWSRequestID:             lc.AwsRequestID,
			InvokedFunctionArn:       lc.InvokedFunctionArn,
			LogGroupName:             lambdacontext.LogGroupName,
			LogStreamName:            lambdacontext.LogStreamName,
			MemoryLimitInMB:          lambdacontext.MemoryLimitInMB,
			GetRemainingTimeInMillis: int(time.Until(deadline).Nanoseconds() / 1e6),
			TraceID:                  os.Getenv("_X_AMZN_TRACE_ID"),
		},
		Environment: EnvironmentReportDetails{
			Agent: EnvironmentAgent{
				Runtime:  RUNTIME,
				Version:  VERSION,
				LoadTime: LoadTime,
			},
			Go: EnvironmentGo{
				Version: runtime.Version(),
				MemoryUsage: EnvironmentGoMemoryUsage{
					Alloc:      int(memStats.Alloc),
					TotalAlloc: int(memStats.TotalAlloc),
					Sys:        int(memStats.Sys),
					NumGC:      int(memStats.NumGC),
				},
			},
			Runtime: EnvironmentRuntime{
				Name:    RUNTIME,
				Version: runtime.Version(),
			},
		},
		ColdStart: ColdStart,
		Errors:    errs,
		Plugins:   pluginsMeta,
	}
}
