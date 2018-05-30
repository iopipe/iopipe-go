package iopipe

import (
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// ContextWrapper wraps the AWS lambda context
type ContextWrapper struct {
	lambdacontext.LambdaContext
}
