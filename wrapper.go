package iopipe

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"time"
	"os"
	"runtime"
)

type wrapper struct {
	report          Report
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
}

func (w *wrapper) RunHook(hook string) {
	for _, plugin := range w.plugins {
		if plugin != nil {
			plugin.RunHook(hook)
		}
	}
}

func (w *wrapper) Invoke(ctx context.Context, payload interface{}) (interface{}, error) {
	w.RunHook(HOOK_PRE_INVOKE)
	defer func() {
		if err := recover(); err != nil {
			// the stack trace the client gets will be a bit verbose with mentions of the wrapper
			w.Report(NewHandlerError(err, true))

			panic(err)
		}
	}()

	go func() {
		timeoutWindow := 0 * time.Millisecond
		if w.agent.TimeoutWindow != nil {
			timeoutWindow = *w.agent.TimeoutWindow
		}

		<-time.After(time.Until(w.deadline.Add(- timeoutWindow)))
		w.Report(NewHandlerError(fmt.Errorf("timeout exceeded"), false))
	}()

	w.startTime = time.Now()
	response, err := w.wrappedHandler(ctx, payload)

	return response, err
}

func (w *wrapper) PostInvoke() {
	w.RunHook(HOOK_POST_INVOKE)
}

func (w *wrapper) Report(err *handlerError) {
	if w.reportSending {
		return
	}
	w.RunHook(HOOK_POST_INVOKE)

	w.RunHook(HOOK_PRE_REPORT)

	w.endTime = time.Now()
	w.reportSending = true

	w.prepareReport(err)

	reportJSON, _ := json.MarshalIndent(w.report, "", "  ")

	// Report to AWS
	UploadFileToS3(fmt.Sprintf("%s.json", w.report.AWS.AWSRequestId), string(reportJSON))

	w.RunHook(HOOK_POST_REPORT)
}

func wrapHandler(handler interface{}, agentInstance *agent) lambdaHandler {
	// decorate the handler

	return func(context context.Context, payload interface{}) (interface{}, error) {
		handlerWrapper := NewWrapper(handler, agentInstance)

		handlerWrapper.PreInvoke(context)
		response, err := handlerWrapper.Invoke(context, payload)
		handlerWrapper.PostInvoke()
		handlerWrapper.Report(NewHandlerError(err, false))

		COLD_START = false
		return response, err
	}
}

func (w *wrapper) prepareReport(err *handlerError) {
	startTime := w.startTime
	endTime := w.endTime
	deadline := w.deadline
	lc := w.lambdaContext
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pluginsMeta := make([]interface{}, len(w.plugins))
	for index, plugin := range w.plugins {
		pluginsMeta[index] = plugin.Meta()
	}

	w.report = Report{
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
		Errors:    err,
		Plugins:   pluginsMeta,
	}
}
