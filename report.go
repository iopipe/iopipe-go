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
	sent      bool
	startTime time.Time

	ClientID      string             `json:"client_id"`
	InstallMethod string             `json:"installMethod"`
	Duration      int                `json:"duration"`
	ProcessID     string             `json:"processId"`
	Timestamp     int                `json:"timestamp"`
	TimestampEnd  int                `json:"timestampEnd"`
	AWS           *ReportAWS         `json:"aws"`
	Disk          *ReportDisk        `json:"disk"`
	Environment   *ReportEnvironment `json:"environment"`
	ColdStart     bool               `json:"coldstart"`
	Errors        interface{}        `json:"errors"`
	CustomMetrics []CustomMetric     `json:"custom_metrics"`
	labels        map[string]struct{}
	Labels        []string     `json:"labels"`
	Plugins       []PluginMeta `json:"plugins"`
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
	FreeMem  uint64                    `json:"freemem"`
	Hostname string                    `json:"hostname"`
	TotalMem uint64                    `json:"totalmem"`
	UsedMem  uint64                    `json:"usedmem"`
	CPUs     []ReportEnvironmentOSCPU  `json:"cpus"`
	Linux    *ReportEnvironmentOSLinux `json:"linux"`
}

// ReportEnvironmentOSCPU contains cpu information
type ReportEnvironmentOSCPU struct {
	Times ReportEnvironmentOSCPUTimes `json:"times"`
}

// ReportEnvironmentOSCPUTimes contains cpu times
type ReportEnvironmentOSCPUTimes struct {
	Idle uint64 `json:"idle"`
	Irq  uint64 `json:"irq"`
	Nice uint64 `json:"nice"`
	Sys  uint64 `json:"sys"`
	User uint64 `json:"user"`
}

// ReportEnvironmentOSLinux contains linux system information
type ReportEnvironmentOSLinux struct {
	PID *ReportEnvironmentOSLinuxPID `json:"pid"`
}

// ReportEnvironmentOSLinuxPID contains linux process information
type ReportEnvironmentOSLinuxPID struct {
	Self *ReportEnvironmentOSLinuxPIDSelf `json:"self"`
}

// ReportEnvironmentOSLinuxPIDSelf contains current process information
type ReportEnvironmentOSLinuxPIDSelf struct {
	Stat      *ReportEnvironmentOSLinuxPIDSelfStat   `json:"stat"`
	StatStart *ReportEnvironmentOSLinuxPIDSelfStat   `json:"stat_start"`
	Status    *ReportEnvironmentOSLinuxPIDSelfStatus `json:"status"`
}

// ReportEnvironmentOSLinuxPIDSelfStat contains process stats
type ReportEnvironmentOSLinuxPIDSelfStat struct {
	Cstime uint64 `json:"cstime"`
	Cutime uint64 `json:"cutime"`
	Stime  uint64 `json:"stime"`
	Utime  uint64 `json:"utime"`
}

// ReportEnvironmentOSLinuxPIDSelfStatus contains process status
type ReportEnvironmentOSLinuxPIDSelfStatus struct {
	FDSize  int32  `json:"FDSize"`
	Threads int32  `json:"Threads"`
	VMRSS   uint64 `json:"VmRSS"`
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
	statStart := readPIDStat()

	agent := handler.agent

	lc := handler.lambdaContext
	if lc == nil {
		lc = &lambdacontext.LambdaContext{
			AwsRequestID:       generateUUID(),
			InvokedFunctionArn: fmt.Sprintf("arn:aws:lambda:local:0:function:%s", lambdacontext.FunctionName),
		}
	}

	pluginsMeta := make([]PluginMeta, 0)

	token := ""
	if agent != nil && agent.Token != nil {
		token = *agent.Token
	}

	return &Report{
		agent:     agent,
		sent:      false,
		startTime: startTime,

		ClientID:      token,
		InstallMethod: "manual",
		ProcessID:     processID,
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
				LoadTime: loadTime,
			},
			Host: &ReportEnvironmentHost{
				BootID: bootID,
			},
			OS: &ReportEnvironmentOS{
				Hostname: hostname,
				Linux: &ReportEnvironmentOSLinux{
					PID: &ReportEnvironmentOSLinuxPID{
						Self: &ReportEnvironmentOSLinuxPIDSelf{
							Stat: &ReportEnvironmentOSLinuxPIDSelfStat{},
							StatStart: &ReportEnvironmentOSLinuxPIDSelfStat{
								Cstime: statStart.cstime,
								Cutime: statStart.cutime,
								Stime:  statStart.stime,
								Utime:  statStart.utime,
							},
						},
					},
				},
			},
			Runtime: &ReportEnvironmentRuntime{
				Name:    RUNTIME,
				Version: runtimeVersion,
			},
		},
		ColdStart:     coldStart,
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

	r.Plugins = make([]PluginMeta, len(r.agent.plugins))
	for index, plugin := range r.agent.plugins {
		r.Plugins[index] = *plugin.Meta()
	}

	statEnd := readPIDStat()
	r.Environment.OS.Linux.PID.Self.Stat.Cstime = statEnd.cstime
	r.Environment.OS.Linux.PID.Self.Stat.Cutime = statEnd.cutime
	r.Environment.OS.Linux.PID.Self.Stat.Stime = statEnd.stime
	r.Environment.OS.Linux.PID.Self.Stat.Utime = statEnd.utime

	status := readPIDStatus()
	r.Environment.OS.Linux.PID.Self.Status = &ReportEnvironmentOSLinuxPIDSelfStatus{
		FDSize:  status.fdSize,
		Threads: status.threads,
		VMRSS:   status.vmRss,
	}

	cpus := readSystemStat()
	cpuTimes := make([]ReportEnvironmentOSCPU, len(cpus))
	for index, cpu := range cpus {
		cpuTimes[index].Times = ReportEnvironmentOSCPUTimes{
			Idle: cpu.idle,
			Irq:  cpu.irq,
			Nice: cpu.nice,
			Sys:  cpu.sys,
			User: cpu.user,
		}
	}
	r.Environment.OS.CPUs = cpuTimes

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
}

// send sends the report to IOpipe
func (r *Report) send() {
	if r.sent {
		return
	}

	r.sent = true

	r.preReport()

	if r.agent != nil && r.agent.Reporter != nil {
		err := r.agent.Reporter(r)

		if err != nil {
			r.agent.log.Debug("Reporting error: ", err)
		}
	}

	r.postReport()
}

// preReport runs the PreReport hooks
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

// postReport runs the PostReport hooks
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
