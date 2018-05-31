package iopipe

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
	"text/template"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlerWrapper_prepareReport(t *testing.T) {
	Convey("Prepare report using information found inside wrapper instance", t, func() {
		hw := HandlerWrapper{}

		Convey("Report generated on empty wrapper adheres to spec", func() {
			hw.prepareReport(nil)
			reportJSONBytes, _ := json.Marshal(hw.report)

			var actualReportJSON interface{}
			_ = json.Unmarshal(reportJSONBytes, &actualReportJSON)

			var expectedReportJSON interface{}
			_ = json.Unmarshal([]byte(executeTemplateString(emptyReport, hw.report)), &expectedReportJSON)

			So(actualReportJSON, ShouldResemble, expectedReportJSON)
		})

		Convey("Is idempotent", func() {
			hw.prepareReport(nil)
			report1 := hw.report

			hw.prepareReport(nil)
			report2 := hw.report

			So(report1, ShouldResemble, report2)
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
    "runtime": {
      "name": "go",
      "version": "{{.Environment.Runtime.Version}}"
    }
  },
  "coldstart": true,
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
