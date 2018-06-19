package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func handler(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	context.IOpipe.Error(fmt.Errorf("oh noes"))

	return "Invocation successful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
