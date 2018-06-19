package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{Debug: iopipe.True()})

func handler() (string, error) {
	return "", fmt.Errorf("Invocation error")
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
