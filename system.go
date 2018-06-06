package iopipe

import (
	"os"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
)

type diskInfo struct {
	totalMiB       float64
	usedMiB        float64
	usedPercentage float64
}

// readBootID returns the /proc/sys/kernel/random/boot_id
func readBootID() string {
	info, _ := host.Info()
	return info.HostID
}

// readDisk returns usage stats for /tmp
func readDisk() *diskInfo {
	// TODO: Make windows friendly
	stat, _ := disk.Usage("/tmp")
	return &diskInfo{
		totalMiB:       float64(stat.Total) / float64(1<<20),
		usedMiB:        float64(stat.Used) / float64(1<<20),
		usedPercentage: stat.UsedPercent,
	}
}

// readHostname returns the system hostname
func readHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
