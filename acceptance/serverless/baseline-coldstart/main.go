package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler() (string, error) {
	defer os.Exit(1)

	return "Invocation successful", nil
}

func main() {
	lambda.Start(handler)
}
