package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{Debug: iopipe.True()})

func handler(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	context.IOpipe.Label("dont-label-me")

	return "Invocation successful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
