package iopipe

// Report contains an IOpipe report
type Report struct {
	ClientID      string            `json:"client_id"`
	InstallMethod string            `json:"installMethod"`
	Duration      int               `json:"duration"`
	ProcessID     string            `json:"processId"`
	Timestamp     int               `json:"timestamp"`
	TimestampEnd  int               `json:"timestampEnd"`
	AWS           ReportAWS         `json:"aws"`
	Environment   ReportEnvironment `json:"environment"`
	ColdStart     bool              `json:"coldstart"`
	Errors        interface{}       `json:"errors"`
	CustomMetrics []CustomMetric    `json:"custom_metrics"`
	Plugins       []interface{}     `json:"plugins"`
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
	Agent   ReportEnvironmentAgent   `json:"agent"`
	Host    ReportEnvironmentHost    `json:"host"`
	OS      ReportEnvironmentOS      `json:"os"`
	Runtime ReportEnvironmentRuntime `json:"runtime"`
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
