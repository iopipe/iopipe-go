package iopipe

import (
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContextWrapper_NewContextWrapper(t *testing.T) {
	Convey("A context wrapper should wrap a lambda context", t, func() {
		lc := lambdacontext.LambdaContext{}
		hw := &HandlerWrapper{}
		cw := NewContextWrapper(lc, hw)

		Convey("And embed the lambda context", func() {
			So(cw.AwsRequestID, ShouldNotBeNil)
		})

		Convey("And provide the handler wrapper", func() {
			So(cw.handler, ShouldEqual, hw)
		})
	})
}

func TestContextWrapper_Metric(t *testing.T) {
	Convey("A context wrapper allows custom metrics to be added to a report", t, func() {
		lc := lambdacontext.LambdaContext{}
		hw := &HandlerWrapper{}
		cw := NewContextWrapper(lc, hw)

		Convey("Doesnot panic if there is no report", func() {
			So(cw.handler.report, ShouldBeNil)
			So(func() {
				cw.Metric("foo", "bar")
			}, ShouldNotPanic)
		})

		Convey("Add a custom metric to the report", func() {
			r := NewReport(hw)
			cw.handler.report = r

			So(len(cw.handler.report.CustomMetrics), ShouldEqual, 0)
			cw.Metric("foo", "bar")
			So(len(cw.handler.report.CustomMetrics), ShouldEqual, 1)
		})

		Convey("Does not add metric if name is too long", func() {
			r := NewReport(hw)
			cw.handler.report = r

			So(len(cw.handler.report.CustomMetrics), ShouldEqual, 0)
			cw.Metric(strings.Repeat("X", 129), "bar")
			So(len(cw.handler.report.CustomMetrics), ShouldEqual, 0)
		})
	})
}

func TestContextWrapper_coerceNumeric(t *testing.T) {
	Convey("Numeric values should be coerced to their type or nil", t, func() {
		So(coerceNumeric("not a number"), ShouldBeNil)
		So(coerceNumeric(true), ShouldBeNil)

		So(coerceNumeric(123), ShouldEqual, 123)
		So(coerceNumeric(123.456), ShouldEqual, 123.456)
	})
}

func TestContextWrapper_coerceString(t *testing.T) {
	Convey("String values should be coerced to strings or nil", t, func() {
		So(coerceString("i'm a string"), ShouldEqual, "i'm a string")

		So(coerceString(true), ShouldBeNil)
		So(coerceString(123), ShouldBeNil)
		So(coerceString(123.456), ShouldBeNil)
	})
}
