package iopipe

import (
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContext_NewContextWrapper(t *testing.T) {
	Convey("A context wrapper should wrap a lambda context", t, func() {
		lc := lambdacontext.LambdaContext{}
		hw := &HandlerWrapper{}
		cw := NewContextWrapper(&lc, hw)

		Convey("And embed the lambda context", func() {
			So(cw.AwsRequestID, ShouldNotBeNil)
		})

		Convey("And provide the handler wrapper", func() {
			So(cw.iopipe, ShouldEqual, hw)
		})
	})
}
