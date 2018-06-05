package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUtil_coerceNumeric(t *testing.T) {
	Convey("Numeric values should be coerced to their type or nil", t, func() {
		So(coerceNumeric("not a number"), ShouldBeNil)
		So(coerceNumeric(true), ShouldBeNil)

		So(coerceNumeric(123), ShouldEqual, 123)
		So(coerceNumeric(123.456), ShouldEqual, 123.456)
	})
}

func TestUtil_coerceString(t *testing.T) {
	Convey("String values should be coerced to strings or nil", t, func() {
		So(coerceString("i'm a string"), ShouldEqual, "i'm a string")

		So(coerceString(true), ShouldBeNil)
		So(coerceString(123), ShouldBeNil)
		So(coerceString(123.456), ShouldBeNil)
	})
}
