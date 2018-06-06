package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func benchmarkReadBootID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readBootID()
	}
}

func TestSystem_readBootID(t *testing.T) {
	Convey("readBootId should return the system boot ID", t, func() {
		bootID := readBootID()

		So(bootID, ShouldNotEqual, nil)
		So(bootID, ShouldNotEqual, "")
	})
}

func TestSystem_readDisk(t *testing.T) {
	Convey("readDisk should return stats about /tmp", t, func() {
		d := readDisk()

		So(d.totalMiB, ShouldNotBeNil)
		So(d.usedMiB, ShouldNotBeNil)
		So(d.usedPercentage, ShouldNotBeNil)
	})
}

func benchmarkReadHostname(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readHostname()
	}
}

func TestSystem_readHostname(t *testing.T) {
	Convey("readHostname should return the system hostname", t, func() {
		hostname := readHostname()

		So(hostname, ShouldNotEqual, nil)
		So(hostname, ShouldNotEqual, "")
	})
}

func TestSystem_readMemInfo(t *testing.T) {
	Convey("readMemInfo should return memory stats", t, func() {
		m := readMemInfo()

		So(m.available, ShouldNotBeNil)
		So(m.free, ShouldNotBeNil)
		So(m.total, ShouldNotBeNil)
		So(m.totalMiB, ShouldNotBeNil)
		So(m.used, ShouldNotBeNil)
		So(m.usedMiB, ShouldNotBeNil)
		So(m.usedPercentage, ShouldNotBeNil)
	})
}
