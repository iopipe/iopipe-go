package iopipe

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContext_NewContext(t *testing.T) {
	Convey("NewContext should return a new context", t, func() {
		lc := lambdacontext.LambdaContext{}
		hw := &HandlerWrapper{}
		cw := NewContextWrapper(&lc, hw)

		ctx := context.Background()
		newCtx := NewContext(ctx, cw)

		So(newCtx, ShouldNotBeNil)
	})
}

func TestContext_FromContext(t *testing.T) {
	Convey("FromContext should return context value", t, func() {
		lc := lambdacontext.LambdaContext{}
		hw := &HandlerWrapper{}
		cw := NewContextWrapper(&lc, hw)

		ctx := context.Background()
		newCtx := NewContext(ctx, cw)
		cwValue, _ := FromContext(newCtx)

		So(cwValue, ShouldResemble, cw)
	})
}

func TestContext_NewContextWrapper(t *testing.T) {
	Convey("A context wrapper should wrap a lambda context", t, func() {
		lc := lambdacontext.LambdaContext{}
		hw := &HandlerWrapper{}
		cw := NewContextWrapper(&lc, hw)

		Convey("And embed the lambda context", func() {
			So(cw.AwsRequestID, ShouldNotBeNil)
		})

		Convey("And provide the handler wrapper", func() {
			So(cw.IOpipe, ShouldEqual, hw)
		})
	})
}
