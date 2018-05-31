package iopipe

import (
	"crypto/rand"
	"io"
	"runtime"
	"time"
)

var (
	// ProcessID is the ID for this process
	ProcessID = GetProcessID()
	// ColdStart is true if this is a cold start
	ColdStart = true
	// RuntimeVersion is the golang runtime version (minus "go" prefix)
	RuntimeVersion = runtime.Version()[2:]
	// LoadTime is the unix time the module was loaded
	LoadTime = int(time.Now().UnixNano() / 1e6)
	// BootID is the kernel's boot_id
	BootID = ReadBootID()
)

const (
	// VERSION is the version of the IOpipe agent
	VERSION = "0.1.1"
	// RUNTIME is the runtime of the IOpipe agent
	RUNTIME = "go"
)

// GetProcessID returns a unique identifier for this process
func GetProcessID() string {
	var uuid UUID
	io.ReadFull(rand.Reader, uuid[:])

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return uuid.String()
}
