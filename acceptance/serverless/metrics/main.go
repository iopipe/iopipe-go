package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{Debug: iopipe.True()})

func handler(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	context.IOpipe.Metric("time", int(time.Now().UnixNano()/1e6))
	context.IOpipe.Metric("a-metric", "value")

	return "Invocation successful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
