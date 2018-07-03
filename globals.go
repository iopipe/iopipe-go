package iopipe

import (
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
	processID = generateUUID()

	// RuntimeVersion is the golang runtime version (minus "go" prefix)
	runtimeVersion = runtime.Version()[2:]
)

const (
	// VERSION is the version of the IOpipe agent
	VERSION = "0.1.1"

	// RUNTIME is the runtime of the IOpipe agent
	RUNTIME = "go"
)
