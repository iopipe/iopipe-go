package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{Debug: iopipe.True()})

func handler(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	// Send the report before exiting
	context.IOpipe.Error(nil)

	os.Exit(1)

	return "Invocation successful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
