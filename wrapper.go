package iopipe

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"time"
	"os"
	"runtime"
	"sync"
	"reflect"
)

type wrapper struct {
	report          *Report
	reporter        Reporter
	originalHandler interface{}
	wrappedHandler  lambdaHandler
	startTime       time.Time
	endTime         time.Time
	deadline        time.Time
	lambdaContext   *lambdacontext.LambdaContext
	reportSending   bool
	agent           *agent
	plugins         []Plugin
}

func NewWrapper(handler interface{}, agentInstance *agent) *wrapper {

	w := &wrapper{
		originalHandler: handler,
		wrappedHandler:  newHandler(handler),
		agent:           agentInstance,
	}

	var plugins []Plugin
	pluginInstantiators := *agentInstance.PluginInstantiators

	if pluginInstantiators != nil {
		plugins = make([]Plugin, len(pluginInstantiators))
		for index, pluginInstantiator := range pluginInstantiators {
			plugins[index] = pluginInstantiator(w)
		}
	}

	w.plugins = plugins

	w.RunHook(HOOK_PRE_SETUP)

	w.RunHook(HOOK_POST_SETUP)
	return w
}

func (w *wrapper) PreInvoke(context context.Context) {
	w.deadline, _ = context.Deadline()
	w.lambdaContext, _ = lambdacontext.FromContext(context)
	w.RunHook(HOOK_PRE_INVOKE)
}

func (w *wrapper) RunHook(hook string) {
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

func (w *wrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
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
		case <-time.After(time.Until(w.deadline.Add(- timeoutWindow))):
			w.PostInvoke(fmt.Errorf("timeout exceeded"))
		case <-ctx.Done():
			return // returning not to leak the goroutine
		}
	}()

	w.startTime = time.Now()
	return w.wrappedHandler(ctx, payload)
}

func (w *wrapper) PostInvoke(err error) {
	if w.reportSending {
		return
	}
	w.reportSending = true

	w.RunHook(HOOK_POST_INVOKE)
	w.RunHook(HOOK_PRE_REPORT)

	w.endTime = time.Now()

	var (
		ok bool
		hErr *invocationError
	)

	if hErr, ok = err.(*invocationError); !ok {
		hErr = NewInvocationError(err)
	}
	w.prepareReport(hErr)

	// PostInvoke
	if w.reporter != nil {
		w.reporter(w.report)
	}

	w.RunHook(HOOK_POST_REPORT)
}

func wrapHandler(handler interface{}, agentInstance *agent) lambdaHandler {
	// decorate the handler

	return func(context context.Context, payload interface{}) (interface{}, error) {
		handlerWrapper := NewWrapper(handler, agentInstance)

		handlerWrapper.PreInvoke(context)
		response, err := handlerWrapper.Invoke(context, payload)
		handlerWrapper.PostInvoke(err)

		COLD_START = false
		return response, err
	}
}

func (w *wrapper) prepareReport(invErr *invocationError) {
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
	errs = invErr
	if reflect.ValueOf(errs).IsNil() {
		errs = &struct {}{}
	}

	w.report = &Report{
		ClientID:      "TODO: some-client-id",
		ProjectId:     nil,
		InstallMethod: "TODO: manual",
		Duration:      int(endTime.Sub(startTime).Nanoseconds()),
		ProcessId:     PROCESS_ID,
		Timestamp:     int(startTime.UnixNano() / 1e6),
		TimestampEnd:  int(endTime.UnixNano() / 1e6),
		AWS: AWSReportDetails{
			FunctionName:             lambdacontext.FunctionName,
			FunctionVersion:          lambdacontext.FunctionVersion,
			AWSRequestId:             lc.AwsRequestID,
			InvokedFunctionArn:       lc.InvokedFunctionArn,
			LogGroupName:             lambdacontext.LogGroupName,
			LogStreamName:            lambdacontext.LogStreamName,
			MemoryLimitInMB:          lambdacontext.MemoryLimitInMB,
			GetRemainingTimeInMillis: int(time.Until(deadline).Nanoseconds() / 1e6),
			TraceId:                  os.Getenv("_X_AMZN_TRACE_ID"),
		},
		Environment: EnvironmentReportDetails{
			Agent: EnvironmentAgent{
				Runtime:  RUNTIME,
				Version:  VERSION,
				LoadTime: MODULE_LOAD_TIME,
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
		ColdStart: COLD_START,
		Errors:    errs,
		Plugins:   pluginsMeta,
	}
}
