package iopipe

import (
	"github.com/shirou/gopsutil/host"
)

// ReadBootID returns the /proc/sys/kernel/random/boot_id
func ReadBootID() string {
	info, _ := host.Info()
	return info.HostID
}

// ReadHostname returns the system hostname
func ReadHostname() string {
	info, _ := host.Info()
	return info.Hostname
}
