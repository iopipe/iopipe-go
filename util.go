package iopipe

import (
	"reflect"
	"runtime"
)

func getFuncName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
