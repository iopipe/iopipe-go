package iopipe

import (
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Report contains an IOpipe report
type Report struct {
	handler       *HandlerWrapper
	ClientID      string             `json:"client_id"`
	InstallMethod string             `json:"installMethod"`
	Duration      int                `json:"duration"`
	ProcessID     string             `json:"processId"`
	Timestamp     int                `json:"timestamp"`
	TimestampEnd  int                `json:"timestampEnd"`
	AWS           *ReportAWS         `json:"aws"`
	Environment   *ReportEnvironment `json:"environment"`
	ColdStart     bool               `json:"coldstart"`
	Errors        interface{}        `json:"errors"`
	CustomMetrics []CustomMetric     `json:"custom_metrics"`
	labels        map[string]struct{}
	Labels        []string      `json:"labels"`
	Plugins       []interface{} `json:"plugins"`
}

// ReportAWS contains AWS invocation details
type ReportAWS struct {
	FunctionName             string `json:"functionName"`
	FunctionVersion          string `json:"functionVersion"`
	AWSRequestID             string `json:"awsRequestId"`
	InvokedFunctionArn       string `json:"invokedFunctionArn"`
	LogGroupName             string `json:"logGroupName"`
	LogStreamName            string `json:"logStreamName"`
	MemoryLimitInMB          int    `json:"memoryLimitInMB"`
	GetRemainingTimeInMillis int    `json:"getRemainingTimeInMillis"`
	TraceID                  string `json:"traceId"`
}

// ReportEnvironment contains environment information
type ReportEnvironment struct {
	Agent   *ReportEnvironmentAgent   `json:"agent"`
	Host    *ReportEnvironmentHost    `json:"host"`
	OS      *ReportEnvironmentOS      `json:"os"`
	Runtime *ReportEnvironmentRuntime `json:"runtime"`
}

// ReportEnvironmentAgent contains information about the IOpipe agent
type ReportEnvironmentAgent struct {
	Runtime  string `json:"runtime"`
	Version  string `json:"version"`
	LoadTime int    `json:"load_time"`
}

// ReportEnvironmentHost contains host information
type ReportEnvironmentHost struct {
	BootID string `json:"boot_id"`
}

// ReportEnvironmentOS contains operating system information
type ReportEnvironmentOS struct {
	Hostname string `json:"hostname"`
}

// ReportEnvironmentRuntime contains runtime information
type ReportEnvironmentRuntime struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Reporter is the reporter interface
type Reporter func(report *Report) error

// CustomMetric is a custom metric
type CustomMetric struct {
	Name string      `json:"name"`
	S    interface{} `json:"s"`
	N    interface{} `json:"n"`
}

// NewReport instantiates a new IOpipe report
func NewReport(handler *HandlerWrapper) *Report {
	lc := handler.lambdaContext
	if lc == nil {
		lc = &lambdacontext.LambdaContext{
			AwsRequestID: "ERROR",
		}
	}

	pluginsMeta := make([]interface{}, len(handler.plugins))
	for index, plugin := range handler.plugins {
		pluginsMeta[index] = plugin.Meta()
	}

	token := ""
	if handler.agent != nil && handler.agent.Token != nil {
		token = *handler.agent.Token
	}

	return &Report{
		handler:       handler,
		ClientID:      token,
		InstallMethod: "manual",
		ProcessID:     ProcessID,
		Timestamp:     int(handler.startTime.UnixNano() / 1e6),
		AWS: &ReportAWS{
			FunctionName:             lambdacontext.FunctionName,
			FunctionVersion:          lambdacontext.FunctionVersion,
			AWSRequestID:             lc.AwsRequestID,
			InvokedFunctionArn:       lc.InvokedFunctionArn,
			LogGroupName:             lambdacontext.LogGroupName,
			LogStreamName:            lambdacontext.LogStreamName,
			MemoryLimitInMB:          lambdacontext.MemoryLimitInMB,
			GetRemainingTimeInMillis: int(time.Until(handler.deadline).Nanoseconds() / 1e6),
			TraceID:                  os.Getenv("_X_AMZN_TRACE_ID"),
		},
		Environment: &ReportEnvironment{
			Agent: &ReportEnvironmentAgent{
				Runtime:  RUNTIME,
				Version:  VERSION,
				LoadTime: LoadTime,
			},
			Host: &ReportEnvironmentHost{
				BootID: BootID,
			},
			OS: &ReportEnvironmentOS{
				Hostname: Hostname,
			},
			Runtime: &ReportEnvironmentRuntime{
				Name:    RUNTIME,
				Version: RuntimeVersion,
			},
		},
		ColdStart:     ColdStart,
		CustomMetrics: []CustomMetric{},
		labels:        make(map[string]struct{}, 0),
		Labels:        make([]string, 0),
		Errors:        &struct{}{},
		Plugins:       pluginsMeta,
	}
}

// Prepare prepares an IOpipe report to be sent
func (r *Report) prepare(invErr *InvocationError) {
	r.Duration = int(r.handler.endTime.Sub(r.handler.startTime).Nanoseconds())
	r.TimestampEnd = int(r.handler.endTime.UnixNano() / 1e6)

	if invErr != nil {
		r.Errors = invErr
	}
}
