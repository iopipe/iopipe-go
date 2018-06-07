package iopipe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"text/template"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReport_NewReport(t *testing.T) {
	Convey("Prepare report using information found inside wrapper instance", t, func() {
		a := NewAgent(Config{})
		hw := &HandlerWrapper{agent: a}

		Convey("Report generated on empty wrapper adheres to spec", func() {
			r := NewReport(hw)
			r.prepare(nil)

			reportJSONBytes, err := json.Marshal(r)

			So(reportJSONBytes, ShouldNotBeNil)
			So(err, ShouldBeNil)

			var actualReportJSON interface{}
			err = json.Unmarshal(reportJSONBytes, &actualReportJSON)

			So(err, ShouldBeNil)
			So(actualReportJSON, ShouldNotBeNil)

			fmt.Println("Actual Report JSON:", actualReportJSON)

			var expectedReportJSON interface{}
			_ = json.Unmarshal([]byte(executeTemplateString(emptyReport, r)), &expectedReportJSON)

			So(actualReportJSON, ShouldResemble, expectedReportJSON)
		})
	})
}

const emptyReport = `
{
  "client_id": "",
  "installMethod": "manual",
  "duration": {{.Duration}},
  "processId": "{{.ProcessID}}",
  "timestamp": {{.Timestamp}},
  "timestampEnd": {{.TimestampEnd}},
  "aws": {
    "functionName": "",
    "functionVersion": "",
    "awsRequestId": "ERROR",
    "invokedFunctionArn": "",
    "logGroupName": "",
    "logStreamName": "",
    "memoryLimitInMB": 0,
    "getRemainingTimeInMillis": -9223372036854,
    "traceId": ""
  },
  "disk": {
	  "totalMiB": {{.Disk.TotalMiB}},
	  "usedMiB": {{.Disk.UsedMiB}},
	  "usedPercentage": {{.Disk.UsedPercentage}}
  },
  "environment": {
    "agent": {
      "runtime": "go",
      "version": "0.1.1",
      "load_time": {{.Environment.Agent.LoadTime}}
    },
	"host": {
		"boot_id": "{{.Environment.Host.BootID}}"
	},
	"os": {
		"hostname": "{{.Environment.OS.Hostname}}",
		"totalmem": {{.Environment.OS.TotalMem}},
		"freemem": {{.Environment.OS.FreeMem}},
		"usedmem": {{.Environment.OS.UsedMem}},
		"cpus": [
		{{range $i, $c := .Environment.OS.CPUs}}
		{{if $i}},{{end}}{
			"times": {
				"idle": {{$c.Times.Idle}},
				"irq": {{$c.Times.Irq}},
				"nice": {{$c.Times.Nice}},
				"sys": {{$c.Times.Sys}},
				"user": {{$c.Times.User}}
			}
		}
		{{end}}
		],
		"linux": {
			"pid": {
				"self": {
					"stat": {
						"cstime": {{.Environment.OS.Linux.PID.Self.Stat.Cstime}},
						"cutime": {{.Environment.OS.Linux.PID.Self.Stat.Cutime}},
						"stime": {{.Environment.OS.Linux.PID.Self.Stat.Stime}},
						"utime": {{.Environment.OS.Linux.PID.Self.Stat.Utime}}
					},
					"stat_start": {
						"cstime": {{.Environment.OS.Linux.PID.Self.StatStart.Cstime}},
						"cutime": {{.Environment.OS.Linux.PID.Self.StatStart.Cutime}},
						"stime": {{.Environment.OS.Linux.PID.Self.StatStart.Stime}},
						"utime": {{.Environment.OS.Linux.PID.Self.StatStart.Utime}}
					},
					"status": {
						"FDSize": {{.Environment.OS.Linux.PID.Self.Status.FDSize}},
						"Threads": {{.Environment.OS.Linux.PID.Self.Status.Threads}},
						"VmRSS": {{.Environment.OS.Linux.PID.Self.Status.VMRSS}}
					}
				}
			}
		}
	},
    "runtime": {
      "name": "go",
      "version": "{{.Environment.Runtime.Version}}"
    }
  },
  "coldstart": true,
  "custom_metrics": [],
  "labels": [],
  "errors": {},
  "plugins": []
}
`

func executeTemplateString(tmplString string, data interface{}) string {
	tmpl := template.New("test")

	tmpl, err := tmpl.Parse(tmplString)
	if err != nil {
		log.Fatal("Parse: ", err)
		return ""
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, data)
	if err != nil {
		log.Fatal("Execute: ", err)
		return ""
	}

	return tpl.String()
}
