package iopipe

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// ContextWrapper wraps the AWS lambda context
type ContextWrapper struct {
	*lambdacontext.LambdaContext
	IOpipe *HandlerWrapper
}

type key struct{}

var contextKey = &key{}

// NewContext returns a new Context that contains the IOpipe context wrapper
func NewContext(parent context.Context, cw *ContextWrapper) context.Context {
	return context.WithValue(parent, contextKey, cw)
}

// FromContext returns the context wrapper stored in ctx
func FromContext(ctx context.Context) (*ContextWrapper, bool) {
	cw, ok := ctx.Value(contextKey).(*ContextWrapper)
	return cw, ok
}

// NewContextWrapper returns a new context wrapper
func NewContextWrapper(ctx *lambdacontext.LambdaContext, handler *HandlerWrapper) *ContextWrapper {
	return &ContextWrapper{ctx, handler}
}
