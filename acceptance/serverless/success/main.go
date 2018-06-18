package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func handler() (string, error) {
	return "Invocation successful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
