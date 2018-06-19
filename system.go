package iopipe

import (
	"os"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

type cpuTimes struct {
	idle uint64
	nice uint64
	sys  uint64
	user uint64
	irq  uint64
}

type diskInfo struct {
	totalMiB       float64
	usedMiB        float64
	usedPercentage float64
}

type memInfo struct {
	available      uint64
	free           uint64
	total          uint64
	totalMiB       float64
	used           uint64
	usedMiB        float64
	usedPercentage float64
}

type pidStat struct {
	cstime uint64
	cutime uint64
	stime  uint64
	utime  uint64
}

type pidStatus struct {
	fdSize  int32
	threads int32
	vmRss   uint64
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

// readMemInfo returns memory usage stats
func readMemInfo() *memInfo {
	stat, _ := mem.VirtualMemory()
	return &memInfo{
		available:      stat.Available,
		free:           stat.Free,
		total:          stat.Total,
		totalMiB:       float64(stat.Total) / float64(1<<20),
		used:           stat.Used,
		usedMiB:        float64(stat.Used) / float64(1<<20),
		usedPercentage: stat.UsedPercent,
	}
}

// readPIDStat returns process cpu times
func readPIDStat() *pidStat {
	pid := os.Getpid()

	proc, _ := process.NewProcess(int32(pid))
	times, _ := proc.Times()

	var (
		childSystem uint64
		childUser   uint64
	)

	childSystem, childUser = 0, 0
	children, _ := proc.Children()

	// TODO: Investigate a more efficient way to do this
	if children != nil && len(children) > 0 {
		for child := range children {
			childProc, _ := process.NewProcess(int32(child))
			childTimes, _ := childProc.Times()

			childSystem = childSystem + uint64(childTimes.System)
			childUser = childUser + uint64(childTimes.User)
		}
	}

	return &pidStat{
		cstime: childSystem,
		cutime: childUser,
		stime:  uint64(times.System),
		utime:  uint64(times.User),
	}
}

func readPIDStatus() *pidStatus {
	pid := os.Getpid()

	proc, _ := process.NewProcess(int32(pid))
	memInfo, _ := proc.MemoryInfo()
	threads, _ := proc.NumThreads()

	fdSize := 0
	openFiles, _ := proc.OpenFiles()
	if openFiles != nil {
		fdSize = len(openFiles)
	}

	return &pidStatus{
		fdSize:  int32(fdSize),
		threads: threads,
		vmRss:   memInfo.RSS / 1024,
	}
}

func readSystemStat() []cpuTimes {
	cpus, _ := cpu.Times(true)
	times := make([]cpuTimes, len(cpus))
	for index, time := range cpus {
		times[index] = cpuTimes{
			idle: uint64(time.Idle),
			irq:  uint64(time.Irq),
			nice: uint64(time.Nice),
			sys:  uint64(time.System),
			user: uint64(time.User),
		}
	}

	return times
}
