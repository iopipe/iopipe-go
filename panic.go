package iopipe

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"encoding/json"
)

const defaultErrorFrameCount = 32
const framesToPanicInfo = 3 // (top-of-stack) Callers, getPanicStack -> getPanicInfo -> beyond

func getErrorType(err interface{}) string {
	if errorType := reflect.TypeOf(err); errorType.Kind() == reflect.Ptr {
		return errorType.Elem().Name()
	} else {
		return errorType.Name()
	}
}

type invocationError struct {
	Message    string                  `json:"message"`
	Type       string                  `json:"type"`
	StackTrace []*panicErrorStackFrame `json:"stackTrace,omitempty"`
}

func (h *invocationError) Error() string  {
	errorJSON, _ := json.Marshal(h)
	return string(errorJSON)
}

func NewPanicInvocationError(err interface{}) *invocationError {
	if err == nil {
		return nil
	}

	const framesToHide = framesToPanicInfo + 4 // here (NewPanicInvocationError) -> handler defer func -> 2 for panic -> actual error
	panicInfo := getPanicInfo(err, framesToHide)
	return &invocationError{
		Message:    panicInfo.Message,
		Type:       getErrorType(err),
		StackTrace: panicInfo.StackTrace,
	}
}

func NewInvocationError(err error) *invocationError {
	return &invocationError{
		Message: getErrorMessage(err),
		Type:    getErrorType(err),
	}
}

type panicErrorStackFrame struct {
	Path     string `json:"path"`
	Line     int32  `json:"line"`
	Function string `json:"function"`
}

type panicInfo struct {
	Message    string                  // Value passed to panic call, converted to string
	StackTrace []*panicErrorStackFrame // Stack trace of the panic
}

func getPanicInfo(value interface{}, framesToHide int) panicInfo {
	message := getErrorMessage(value)
	stack := getPanicStack(framesToHide)

	return panicInfo{Message: message, StackTrace: stack}
}

func getErrorMessage(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func getPanicStack(framesToHide int) []*panicErrorStackFrame {
	s := make([]uintptr, defaultErrorFrameCount)
	n := runtime.Callers(framesToHide, s)
	if n == 0 {
		return make([]*panicErrorStackFrame, 0)
	}

	s = s[:n]

	return convertStack(s)
}

func convertStack(s []uintptr) []*panicErrorStackFrame {
	var converted []*panicErrorStackFrame
	frames := runtime.CallersFrames(s)

	for {
		frame, more := frames.Next()

		formattedFrame := formatFrame(frame)
		converted = append(converted, formattedFrame)

		if !more {
			break
		}
	}
	return converted
}

func formatFrame(inputFrame runtime.Frame) *panicErrorStackFrame {
	path := inputFrame.File
	line := int32(inputFrame.Line)
	function := inputFrame.Function

	// Strip GOPATH from path by counting the number of seperators in label & path
	//
	// For example given this:
	//     GOPATH = /home/user
	//     path   = /home/user/src/pkg/sub/file.go
	//     label  = pkg/sub.Type.Method
	//
	// We want to set:
	//     path  = pkg/sub/file.go
	//     label = Type.Method

	i := len(path)
	for n, g := 0, strings.Count(function, "/")+2; n < g; n++ {
		i = strings.LastIndex(path[:i], "/")
		if i == -1 {
			// Something went wrong and path has less seperators than we expected
			// Abort and leave i as -1 to counteract the +1 below
			break
		}
	}

	path = path[i+1:] // Trim the initial /

	// Strip the path from the function name as it's already in the path
	function = function[strings.LastIndex(function, "/")+1:]
	// Likewise strip the package name
	function = function[strings.Index(function, ".")+1:]

	return &panicErrorStackFrame{
		Path:     path,
		Line:     line,
		Function: function,
	}
}
