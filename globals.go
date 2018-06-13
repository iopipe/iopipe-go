package iopipe

import (
	"crypto/rand"
	"io"
	"runtime"
	"time"
)

var (
	// bootID is the kernel's boot_id
	bootID = readBootID()

	// coldStart is true if this is a cold start
	coldStart = true

	// hostname is the system's hostname
	hostname = readHostname()

	// loadTime is the unix time the module was loaded
	loadTime = int(time.Now().UnixNano() / 1e6)

	// processID is the ID for this process
	processID = getProcessID()

	// RuntimeVersion is the golang runtime version (minus "go" prefix)
	runtimeVersion = runtime.Version()[2:]
)

const (
	// VERSION is the version of the IOpipe agent
	VERSION = "0.1.1"

	// RUNTIME is the runtime of the IOpipe agent
	RUNTIME = "go"
)

// getProcessID returns a unique identifier for this process
func getProcessID() string {
	var uuid UUID
	io.ReadFull(rand.Reader, uuid[:])

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return uuid.String()
}
