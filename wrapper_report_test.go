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

func TestWrapper_prepareReport(t *testing.T) {
	t.Skip()
	Convey("Prepare report using information found inside wrapper instance", t, func() {
		w := Wrapper{}

		Convey("Report generated on empty wrapper adheres to spec", func() {
			w.prepareReport(nil)
			reportJSONBytes, _ := json.Marshal(w.report)
			var actualReportJSON interface{}
			_ = json.Unmarshal(reportJSONBytes, &actualReportJSON)

			emptyReportVersioned := fmt.Sprintf(emptyReport, RuntimeVersion, RuntimeVersion)
			var expectedReportJSON interface{}
			_ = json.Unmarshal([]byte(executeTemplateString(emptyReportVersioned, w.report)), &expectedReportJSON)

			So(actualReportJSON, ShouldResemble, expectedReportJSON)
		})

		Convey("Is idempotent", func() {
			w.prepareReport(nil)
			report1 := w.report

			w.prepareReport(nil)
			report2 := w.report

			So(report1, ShouldResemble, report2)
		})
	})
}

const emptyReport = `
{
  "client_id": "",
  "installMethod": "manual",
  "duration": 0,
  "processId": "{{.ProcessId}}",
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
      "Runtime": "go",
      "Version": "0.1.1",
      "LoadTime": {{.Environment.Agent.LoadTime}}
    },
    "go": {
      "version": "%s",
      "memoryUsage": {
        "alloc": {{.Environment.Go.MemoryUsage.Alloc}},
        "totalAlloc": {{.Environment.Go.MemoryUsage.TotalAlloc}},
        "sys": {{.Environment.Go.MemoryUsage.Sys}},
        "numGC": {{.Environment.Go.MemoryUsage.NumGC}}
      }
    },
    "runtime": {
      "name": "go",
      "version": "%s"
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
