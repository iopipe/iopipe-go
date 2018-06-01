package iopipe

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
	"text/template"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReport_NewReport(t *testing.T) {
	Convey("Prepare report using information found inside wrapper instance", t, func() {
		hw := &HandlerWrapper{}

		Convey("Report generated on empty wrapper adheres to spec", func() {
			r := NewReport(hw)
			r.prepare(nil)

			reportJSONBytes, _ := json.Marshal(r)

			var actualReportJSON interface{}
			_ = json.Unmarshal(reportJSONBytes, &actualReportJSON)

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
  "duration": 0,
  "processId": "{{.ProcessID}}",
  "timestamp": -6795364578871,
  "timestampEnd": -6795364578871,
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
		"hostname": "{{.Environment.OS.Hostname}}"
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
