package iopipe

import (
	"os"

	"github.com/shirou/gopsutil/host"
)

// ReadBootID returns the /proc/sys/kernel/random/boot_id
func ReadBootID() string {
	infoStat, _ := host.Info()
	return infoStat.HostID
}

// ReadHostname returns the system hostname
func ReadHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
