package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func handler() (string, error) {
	panic("at the disco")
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
