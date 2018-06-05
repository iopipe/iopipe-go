package iopipe

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlerWrapper_Label(t *testing.T) {
	Convey("A handler wrapper allows labels to be added to a report", t, func() {
		a := NewAgent(Config{})
		hw := &HandlerWrapper{agent: a}

		Convey("Doesnot panic if there is no report", func() {

			So(hw.report, ShouldBeNil)
			So(func() {
				hw.Label("foo")
			}, ShouldNotPanic)
		})

		Convey("Add a label to the report", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(r.labels), ShouldEqual, 0)
			hw.Label("foobar")
			So(len(r.labels), ShouldEqual, 1)

			r.prepare(nil)
			So(len(r.Labels), ShouldEqual, 1)
		})

		Convey("Does not add label if name is too long", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(r.labels), ShouldEqual, 0)
			hw.Label(strings.Repeat("X", 129))
			So(len(r.labels), ShouldEqual, 0)
		})
	})
}

func TestHandlerWrapper_Metric(t *testing.T) {
	Convey("A handler wrapper allows custom metrics to be added to a report", t, func() {
		a := NewAgent(Config{})
		hw := &HandlerWrapper{agent: a}

		Convey("Doesnot panic if there is no report", func() {
			So(hw.report, ShouldBeNil)
			So(func() {
				hw.Metric("foo", "bar")
			}, ShouldNotPanic)
		})

		Convey("Add a custom metric to the report", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
			hw.Metric("foo", "bar")
			So(len(hw.report.CustomMetrics), ShouldEqual, 1)
		})

		Convey("Does not add metric if name is too long", func() {
			r := NewReport(hw)
			hw.report = r

			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
			hw.Metric(strings.Repeat("X", 129), "bar")
			So(len(hw.report.CustomMetrics), ShouldEqual, 0)
		})
	})
}
