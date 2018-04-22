package iopipe

import (
	"fmt"
	"reflect"
	"runtime"
)

const framesToHide = 0 // TODO: this (Callers) -> NewPanicError -> getPanicStack -> getPanicInfo -> defer func -> 2 for panic -> observeAndInvokeF -> 2 for panic -> f
const defaultErrorFrameCount = 32

func getErrorType(err interface{}) string {
	if errorType := reflect.TypeOf(err); errorType.Kind() == reflect.Ptr {
		return errorType.Elem().Name()
	} else {
		return errorType.Name()
	}
}

type handlerError struct {
	Message    string                  `json:"message"`
	Type       string                  `json:"type"`
	StackTrace []*panicErrorStackFrame `json:"stackTrace,omitempty"`
}

func NewHandlerError(err interface{}, isPanic bool) *handlerError {
	if err == nil {
		return nil
	}

	if isPanic {
		panicInfo := getPanicInfo(err)
		return &handlerError{
			Message:    panicInfo.Message,
			Type:       getErrorType(err),
			StackTrace: panicInfo.StackTrace,
		}
	} else {
		return &handlerError{
			Message: getErrorMessage(err),
			Type:    getErrorType(err),
		}
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

func getPanicInfo(value interface{}) panicInfo {
	message := getErrorMessage(value)
	stack := getPanicStack()

	return panicInfo{Message: message, StackTrace: stack}
}

func getErrorMessage(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func getPanicStack() []*panicErrorStackFrame {
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

	return &panicErrorStackFrame{
		Path:     path,
		Line:     line,
		Function: function,
	}
}
