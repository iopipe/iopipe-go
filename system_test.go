package iopipe

import (
	"testing"

	"github.com/shirou/gopsutil/cpu"
	. "github.com/smartystreets/goconvey/convey"
)

func BenchmarkSystem_ReadBootID(b *testing.B) {
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

func BenchmarkSystem_ReadDisk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readDisk()
	}
}

func TestSystem_readDisk(t *testing.T) {
	Convey("readDisk should return stats about /tmp", t, func() {
		d := readDisk()

		So(d.totalMiB, ShouldNotBeNil)
		So(d.usedMiB, ShouldNotBeNil)
		So(d.usedPercentage, ShouldNotBeNil)
	})
}

func BenchmarkSystem_ReadHostname(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readDisk()
	}
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

func BenchmarkSystem_ReadMemInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readMemInfo()
	}
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

func BenchmarkSystem_ReadPIDStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readPIDStat()
	}
}

func TestSystem_readPIDStat(t *testing.T) {
	Convey("readPIDStat should return process cpu times", t, func() {
		p := readPIDStat()

		So(p.cstime, ShouldNotBeNil)
		So(p.cutime, ShouldNotBeNil)
		So(p.stime, ShouldNotBeNil)
		So(p.utime, ShouldNotBeNil)
	})
}

func BenchmarkSystem_ReadPIDStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readPIDStatus()
	}
}

func TestSystem_readPIDStatus(t *testing.T) {
	Convey("readPIDStatus should reutrn fd, thread count and rss", t, func() {
		p := readPIDStatus()

		So(p.fdSize, ShouldNotBeNil)
		So(p.threads, ShouldNotBeNil)
		So(p.vmRss, ShouldNotBeNil)
	})
}

func BenchmarkSystem_ReadSystemStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readSystemStat()
	}
}

func TestSystem_readSystemStat(t *testing.T) {
	Convey("readSystemStat should read return system cpu times", t, func() {
		s := readSystemStat()

		c, _ := cpu.Counts(false)
		So(len(s), ShouldEqual, c)

		for _, t := range s {
			So(t.idle, ShouldNotBeNil)
			So(t.irq, ShouldNotBeNil)
			So(t.nice, ShouldNotBeNil)
			So(t.sys, ShouldNotBeNil)
			So(t.user, ShouldNotBeNil)
		}
	})
}
