package iopipe

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Report contains an IOpipe report
type Report struct {
	agent     *Agent
	sending   bool
	startTime time.Time

	ClientID      string             `json:"client_id"`
	InstallMethod string             `json:"installMethod"`
	Duration      int                `json:"duration"`
	ProcessID     string             `json:"processId"`
	Timestamp     int                `json:"timestamp"`
	TimestampEnd  int                `json:"timestampEnd"`
	AWS           *ReportAWS         `json:"aws"`
	Disk          *ReportDisk        `json:"disk"`
	Memory        *ReportMemory      `json:"memory"`
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

// ReportDisk contains disk usage information
type ReportDisk struct {
	TotalMiB       float64 `json:"totalMiB"`
	UsedMiB        float64 `json:"usedMiB"`
	UsedPercentage float64 `json:"usedPercentage"`
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
	FreeMem  uint64 `json:"freemem"`
	Hostname string `json:"hostname"`
	TotalMem uint64 `json:"totalmem"`
	UsedMem  uint64 `json:"usedmem"`
}

// ReportEnvironmentRuntime contains runtime information
type ReportEnvironmentRuntime struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ReportMemory contains memory usage information
type ReportMemory struct {
	TotalMiB       float64 `json:"totalMiB"`
	UsedMiB        float64 `json:"rssMiB"`
	UsedPercentage float64 `json:"rssTotalPercentage"`
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
		agent:     agent,
		sending:   false,
		startTime: startTime,

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

// prepare prepares an IOpipe report to be sent
func (r *Report) prepare(err error) {
	endTime := time.Now()
	r.TimestampEnd = int(endTime.UnixNano() / 1e6)
	r.Duration = int(endTime.Sub(r.startTime).Nanoseconds())

	if err != nil {
		r.Errors = coerceInvocationError(err)
	}

	for label := range r.labels {
		r.Labels = append(r.Labels, label)
	}

	disk := readDisk()
	r.Disk = &ReportDisk{
		TotalMiB:       disk.totalMiB,
		UsedMiB:        disk.usedMiB,
		UsedPercentage: disk.usedPercentage,
	}

	mem := readMemInfo()
	r.Environment.OS.FreeMem = mem.free
	r.Environment.OS.TotalMem = mem.total
	r.Environment.OS.UsedMem = mem.used
	r.Memory = &ReportMemory{
		TotalMiB:       mem.totalMiB,
		UsedMiB:        mem.usedMiB,
		UsedPercentage: mem.usedPercentage,
	}
}

func (r *Report) send() {
	if r.sending {
		return
	}

	r.sending = true

	r.preReport()

	if r.agent != nil && r.agent.Reporter != nil {
		err := r.agent.Reporter(r)

		if err != nil {
			fmt.Println("Reporting error: ", err)
		}
	}

	r.postReport()
}

func (r *Report) preReport() {
	var wg sync.WaitGroup
	wg.Add(len(r.agent.plugins))

	for _, plugin := range r.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil && plugin.Enabled() {
				plugin.PreReport(r)
			}
		}(plugin)
	}

	wg.Wait()
}

func (r *Report) postReport() {
	var wg sync.WaitGroup
	wg.Add(len(r.agent.plugins))

	for _, plugin := range r.agent.plugins {
		go func(plugin Plugin) {
			defer wg.Done()

			if plugin != nil && plugin.Enabled() {
				plugin.PostReport(r)
			}
		}(plugin)
	}

	wg.Wait()
}
