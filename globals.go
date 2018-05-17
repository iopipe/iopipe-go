package iopipe

import (
	"crypto/rand"
	"io"
	"runtime"
	"time"
)

var (
	PROCESS_ID       = ProcessId()
	COLD_START       = true
	RUNTIME_VERSION  = runtime.Version()
	MODULE_LOAD_TIME = int(time.Now().UnixNano() / 1e6)
)

const (
	VERSION = "0.1.1"
	RUNTIME = "go"
)

func ProcessId() string {
	var uuid UUID
	io.ReadFull(rand.Reader, uuid[:])

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid.String()
}
