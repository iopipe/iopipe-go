package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func BenchmarkReadBootID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ReadBootID()
	}
}

func TestSystem_ReadBootID(t *testing.T) {
	Convey("ReadBootId should return the system boot ID", t, func() {
		bootID := ReadBootID()

		So(bootID, ShouldNotEqual, nil)
		So(bootID, ShouldNotEqual, "")
	})
}

func BenchmarkReadHostname(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ReadHostname()
	}
}

func TestSystem_ReadHostname(t *testing.T) {
	Convey("ReadHostname should return the system hostname", t, func() {
		hostname := ReadHostname()

		So(hostname, ShouldNotEqual, nil)
		So(hostname, ShouldNotEqual, "")
	})
}
