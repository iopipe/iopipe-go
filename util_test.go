package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUtil_coerceNumeric(t *testing.T) {
	Convey("Numeric values should be coerced to their type or nil", t, func() {
		Convey("Strings are not numeric", func() {
			So(coerceNumeric("not a number"), ShouldBeNil)
		})

		Convey("Booleans are not numeric", func() {
			So(coerceNumeric(true), ShouldBeNil)
		})

		Convey("Integers should be coerced to int64", func() {
			So(coerceNumeric(int(123)), ShouldEqual, int64(123))
			So(coerceNumeric(int8(123)), ShouldEqual, int64(123))
			So(coerceNumeric(uint8(123)), ShouldEqual, int64(123))
			So(coerceNumeric(int16(123)), ShouldEqual, int64(123))
			So(coerceNumeric(uint16(123)), ShouldEqual, int64(123))
			So(coerceNumeric(int32(123)), ShouldEqual, int64(123))
			So(coerceNumeric(uint32(123)), ShouldEqual, int32(123))
			So(coerceNumeric(int64(123)), ShouldEqual, int64(123))
			So(coerceNumeric(uint64(123)), ShouldEqual, int64(123))
		})

		Convey("Floats should be coerced to floats", func() {
			So(coerceNumeric(float32(123.456)), ShouldEqual, float32(123.456))
			So(coerceNumeric(float64(123.456)), ShouldEqual, float64(123.456))
		})
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
