package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSystem_ReadBootID(t *testing.T) {
	Convey("ReadBootId should return the system boot ID", t, func() {
		bootId := ReadBootID()

		So(bootId, ShouldNotEqual, nil)
		So(bootId, ShouldNotEqual, "")
	})
}

func TestSystem_ReadHostname(t *testing.T) {
	Convey("ReadHostname should return the system hostname", t, func() {
		hostname := ReadHostname()

		So(hostname, ShouldNotEqual, nil)
		So(hostname, ShouldNotEqual, "")
	})
}
