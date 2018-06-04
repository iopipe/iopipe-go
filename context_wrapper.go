package iopipe

import (
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// ContextWrapper wraps the AWS lambda context
type ContextWrapper struct {
	*lambdacontext.LambdaContext
	iopipe *HandlerWrapper
}

// NewContextWrapper returns a new context wrapper
func NewContextWrapper(ctx *lambdacontext.LambdaContext, handler *HandlerWrapper) *ContextWrapper {
	return &ContextWrapper{ctx, handler}
}

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
