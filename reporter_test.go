package iopipe

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Region struct {
	Region string
	URL    string
}

func TestReporter_getCollectorURL(t *testing.T) {
	regions := []Region{
		Region{Region: "", URL: "https://metrics-api.iopipe.com/v0/event"},
		Region{Region: "ap-northeast-1", URL: "https://metrics-api.ap-northeast-1.iopipe.com/v0/event"},
		Region{Region: "ap-southeast-2", URL: "https://metrics-api.ap-southeast-2.iopipe.com/v0/event"},
		Region{Region: "eu-west-1", URL: "https://metrics-api.eu-west-1.iopipe.com/v0/event"},
		Region{Region: "us-east-1", URL: "https://metrics-api.iopipe.com/v0/event"},
		Region{Region: "us-east-2", URL: "https://metrics-api.us-east-2.iopipe.com/v0/event"},
		Region{Region: "us-west-1", URL: "https://metrics-api.us-west-1.iopipe.com/v0/event"},
		Region{Region: "us-west-2", URL: "https://metrics-api.us-west-2.iopipe.com/v0/event"},
	}

	Convey("getCollectorURL should return the correct URL for the region", t, func() {
		for _, region := range regions {
			So(getCollectorURL(region.Region), ShouldEqual, region.URL)
		}
	})
}

func TestReporter_sendReport(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(res, "")
	}))
	defer ts.Close()

	oldRegion := os.Getenv("AWS_REGION")
	defer os.Setenv("AWS_REGION", oldRegion)

	os.Setenv("AWS_REGION", "mock")
	os.Setenv("MOCK_SERVER", ts.URL)

	a := NewAgent(Config{})
	hw := &HandlerWrapper{agent: a}
	r := NewReport(hw)
	r.prepare(nil)

	Convey("sendReport should send a report", t, func() {
		err := sendReport(r)

		So(err, ShouldBeNil)
	})
}
