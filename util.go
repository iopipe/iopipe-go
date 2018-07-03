package iopipe

import (
	"crypto/rand"
	"io"
	"reflect"
	"runtime"
)

func coerceNumeric(x interface{}) interface{} {
	switch x := x.(type) {
	case uint8:
		return int64(x)
	case int8:
		return int64(x)
	case uint16:
		return int64(x)
	case int16:
		return int64(x)
	case uint32:
		return int64(x)
	case int32:
		return int64(x)
	case uint64:
		return int64(x)
	case int64:
		return int64(x)
	case int:
		return int64(x)
	case float32:
		return float64(x)
	case float64:
		return float64(x)
	default:
		return nil
	}
}

func coerceString(x interface{}) interface{} {
	switch x := x.(type) {
	case string:
		return string(x)
	default:
		return nil
	}
}

func getFuncName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// False returns a pointer to false
func False() *bool {
	b := false
	return &b
}

// True returns a pointer to true
func True() *bool {
	b := true
	return &b
}

func strToBool(s string) *bool {
	if s == "true" || s == "1" {
		return True()
	}

	return False()
}

func generateUUID() string {
	var uuid UUID
	io.ReadFull(rand.Reader, uuid[:])

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return uuid.String()
}
