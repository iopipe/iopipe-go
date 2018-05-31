package iopipe

import (
	"github.com/shirou/gopsutil/host"
)

func ReadBootTime() uint64 {
	bootTime, _ := host.BootTime()
	return bootTime
}
