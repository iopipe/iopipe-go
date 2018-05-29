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

// Wrapper is an IOpipe wrapper
type Wrapper struct {
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

// NewWrapper creates a new IOpipe wrapper
func NewWrapper(handler interface{}, agentInstance *Agent) *Wrapper {

	w := &Wrapper{
		originalHandler: handler,
		wrappedHandler:  newHandler(handler),
		agent:           agentInstance,
	}

	var plugins []Plugin
	pluginInstantiators := agentInstance.PluginInstantiators

	if pluginInstantiators != nil {
		plugins = make([]Plugin, len(pluginInstantiators))
		for index, pluginInstantiator := range pluginInstantiators {
			plugins[index] = pluginInstantiator(w)
		}
	}

	w.plugins = plugins
	w.reporter = reportToIOpipe

	w.RunHook(HookPreSetup)

	w.RunHook(HookPostSetup)
	return w
}

// PreInvoke runs pre invoke hooks
func (w *Wrapper) PreInvoke(context context.Context) {
	w.deadline, _ = context.Deadline()
	w.lambdaContext, _ = lambdacontext.FromContext(context)
	w.RunHook(HookPreInvoke)
}

// RunHook runs the specified hooks
func (w *Wrapper) RunHook(hook string) {
	var wg sync.WaitGroup
	wg.Add(len(w.plugins))

	for _, plugin := range w.plugins {
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
func (w *Wrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	w.startTime = time.Now()
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// the stack trace the client gets will be a bit verbose with mentions of the wrapper
			w.PostInvoke(NewPanicInvocationError(panicErr))
			panic(panicErr)
		}
	}()

	go func() {
		timeoutWindow := 0 * time.Millisecond
		if w.agent != nil && w.agent.TimeoutWindow != nil {
			timeoutWindow = *w.agent.TimeoutWindow
		}

		if w.deadline.IsZero() {
			return
		}

		select {
		// naturally the deadline should occur before the context is closed
		case <-time.After(time.Until(w.deadline.Add(-timeoutWindow))):
			w.PostInvoke(fmt.Errorf("timeout exceeded"))
		case <-ctx.Done():
			return // returning not to leak the goroutine
		}
	}()

	return w.wrappedHandler(ctx, payload)
}

// PostInvoke runs the post invoke hooks
func (w *Wrapper) PostInvoke(err error) {
	if w.reportSending {
		return
	}
	w.reportSending = true

	w.RunHook(HookPostInvoke)
	w.RunHook(HookPreReport)

	w.endTime = time.Now()

	var (
		ok   bool
		hErr *InvocationError
	)

	if hErr, ok = err.(*InvocationError); !ok {
		hErr = NewInvocationError(err)
	}
	w.prepareReport(hErr)

	// PostInvoke
	if w.reporter != nil {
		err := w.reporter(w.report)
		if err != nil {
			// TODO: figure out what to do when invoke errors
			fmt.Println("Reporting errored: ", err)
		}
	}

	w.RunHook(HookPostReport)
}

func wrapHandler(handler interface{}, agentInstance *Agent) lambdaHandler {
	// decorate the handler

	return func(context context.Context, payload interface{}) (interface{}, error) {
		handlerWrapper := NewWrapper(handler, agentInstance)

		handlerWrapper.PreInvoke(context)
		response, err := handlerWrapper.Invoke(context, payload)
		handlerWrapper.PostInvoke(err)

		ColdStart = false
		return response, err
	}
}

func (w *Wrapper) prepareReport(invErr *InvocationError) {
	if w.report != nil {
		return
	}

	startTime := w.startTime
	endTime := w.endTime
	deadline := w.deadline
	lc := w.lambdaContext
	if lc == nil {
		lc = &lambdacontext.LambdaContext{
			AwsRequestID: "ERROR",
		}
	}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pluginsMeta := make([]interface{}, len(w.plugins))
	for index, plugin := range w.plugins {
		pluginsMeta[index] = plugin.Meta()
	}

	var errs interface{}
	errs = &struct{}{}
	if invErr != nil {
		errs = invErr
	}

	token := ""
	if w.agent != nil && w.agent.Token != nil {
		token = *w.agent.Token
	}

	w.report = &Report{
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
