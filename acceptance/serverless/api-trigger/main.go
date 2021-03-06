package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{Debug: iopipe.True()})

func handler() (interface{}, error) {
	url := os.Getenv("API_GATEWAY_URL")
	if url == "" {
		return nil, fmt.Errorf("No API_GATEWAY_URL provided")
	}

	_, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return "Invocation successful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
