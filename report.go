package iopipe

type AWSReportDetails struct {
	FunctionName             string `json:"functionName""`
	FunctionVersion          string `json:"functionVersion"`
	AWSRequestId             string `json:"awsRequestId"`
	InvokedFunctionArn       string `json:"invokedFunctionArn"`
	LogGroupName             string `json:"logGroupName"`
	LogStreamName            string `json:"logStreamName"`
	MemoryLimitInMB          int    `json:"memoryLimitInMB"`
	GetRemainingTimeInMillis int    `json:"getRemainingTimeInMillis"`
	TraceId                  string `json:"traceId"`
}

type Report struct {
	ClientID      string                   `json:"client_id"`
	ProjectId     *string                  `json:"projectId,omitempty"`
	InstallMethod string                   `json:"installMethod"`
	Duration      int                      `json:"duration"`
	ProcessId     string                   `json:"processId"`
	Timestamp     int                      `json:"timestamp"`
	TimestampEnd  int                      `json:"timestampEnd"`
	AWS           AWSReportDetails         `json:"aws"`
	Environment   EnvironmentReportDetails `json:"environment"`
	ColdStart     bool                     `json:"coldstart"`
	Errors        interface{}              `json:"errors"`
	Plugins       []interface{}            `json:"plugins"`
}

type EnvironmentReportDetails struct {
	Agent   EnvironmentAgent   `json:"agent"`
	Go      EnvironmentGo      `json:"go"`
	Runtime EnvironmentRuntime `json:"runtime"`
}

type EnvironmentAgent struct {
	Runtime  string `json:"runtime`
	Version  string `json:"version`
	LoadTime int    `json:"load_time`
}

type EnvironmentRuntime struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type EnvironmentGo struct {
	Version     string                   `json:"version"`
	MemoryUsage EnvironmentGoMemoryUsage `json:"memoryUsage"`
}

type EnvironmentGoMemoryUsage struct {
	Alloc      int `json:"alloc"`
	TotalAlloc int `json:"totalAlloc"`
	Sys        int `json:"sys"`
	NumGC      int `json:"numGC"`
}

type Reporter func(report *Report)

//{
// ...
//  "environment": {
//    "agent": {
//      "runtime": "s",
//      "version": "s",
//      "load_time": "n"
//    },
//    "runtime": {
//      "name": "s",
//      "version": "s",
//      "vendor": "s",
//      "vmVersion": "s",
//      "vmVendor": "s"
//    },
//    "nodejs": {
//      "version": "s",
//      "memoryUsage": {
//        "rss": "n"
//      }
//    },
//    "python": {
//      "version": "s"
//    },
//    "host": {
//      "boot_id": "s"
//    },
//    "os": {
//      "hostname": "s",
//      "totalmem": "n",
//      "freemem": "n",
//      "usedmem": "n",
//      "cpus": [
//        {
//          "times": {
//            "idle": "n",
//            "irq": "n",
//            "sys": "n",
//            "user": "n",
//            "nice": "n"
//          }
//        }
//      ],
//      "linux": {
//        "pid": {
//          "self": {
//            "stat": {
//              "utime": "n",
//              "stime": "n",
//              "cutime": "n",
//              "cstime": "n"
//            },
//            "stat_start": {
//              "utime": "n",
//              "stime": "n",
//              "cutime": "n",
//              "cstime": "n"
//            },
//            "status": {
//              "VmRSS": "n",
//              "Threads": "n",
//              "FDSize": "n"
//            }
//          }
//        }
//      }
//    }
//  },
//  "disk": {
//    "totalMiB": "n",
//    "usedMiB": "n",
//    "usedPercentage": "n"
//  },
//  "memory": {
//    "rssMiB": "n",
//    "totalMiB": "n",
//    "rssTotalPercentage": "n"
//  },
// "errors": {
//   "stack": "s",
//   "name": "s",
//   "message": "s",
//   "stackHash": "s",
//   "count": "n"
// },
//  "custom_metrics": [
//    {
//      "name": "s",
//      "s": "s",
//      "n": "n"
//    }
//  ],
//  "labels": ["s"],
//  "performanceEntries": [
//    {
//      "name": "s",
//      "startTime": "n",
//      "duration": "n",
//      "entryType": "s",
//      "timestamp": "n"
//    }
//  ],
//  "plugins": [
//    {
//      "name": "s",
//      "version": "s",
//      "homepage": "s",
//      "Enabled": "b",
//      "uploads": ["s"]
//    }
//  ]
//}
