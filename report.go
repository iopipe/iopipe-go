package iopipe

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Report contains an IOpipe report
type Report struct {
	agent         *Agent
	reportSending bool
	startTime     time.Time

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
	startTime := time.Now()
	agent := handler.agent

	lc := handler.lambdaContext
	if lc == nil {
		lc = &lambdacontext.LambdaContext{
			AwsRequestID: "ERROR",
		}
	}

	pluginsMeta := make([]interface{}, len(agent.plugins))
	for index, plugin := range agent.plugins {
		pluginsMeta[index] = plugin.Meta()
	}

	token := ""
	if agent != nil && agent.Token != nil {
		token = *agent.Token
	}

	return &Report{
		agent:         agent,
		startTime:     startTime,
		ClientID:      token,
		InstallMethod: "manual",
		ProcessID:     ProcessID,
		Timestamp:     int(startTime.UnixNano() / 1e6),
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
func (r *Report) prepare(err error) {
	endTime := time.Now()
	r.TimestampEnd = int(endTime.UnixNano() / 1e6)
	r.Duration = int(endTime.Sub(r.startTime).Nanoseconds())

	if err != nil {
		r.Errors = coerceInvocationError(err)
	}
}

func (r *Report) send() {
	if r.reportSending {
		return
	}

	r.reportSending = true

	if r.agent.Reporter != nil {
		err := r.agent.Reporter(r)
		if err != nil {
			// TODO: We want to log an error during reporting
			fmt.Println("Reporting errored: ", err)
		}
	}
}
