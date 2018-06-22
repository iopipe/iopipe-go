package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{Debug: iopipe.True()})

func handler(ctx context.Context, event interface{}) (interface{}, error) {
	return event, nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
