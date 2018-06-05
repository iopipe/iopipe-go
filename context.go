package iopipe

import (
	"context"
)

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
