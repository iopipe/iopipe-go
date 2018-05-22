package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.Body {
	case "error":
		return events.APIGatewayProxyResponse{Body: "it's over 9000", StatusCode: 90001}, fmt.Errorf("it's over 9000")
		break
	case "panic":
		panic("panic at the disco")
		break
	case "sleep":
		time.Sleep(10 * time.Minute)
		break
	default:
		break
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	token := "your IOpipe token"
	enabled := true
	timeoutWindow := 100 * time.Millisecond

	agent := iopipe.NewAgent(iopipe.Config{
		Token:         &token,
		TimeoutWindow: &timeoutWindow,
		Enabled:       &enabled,
		PluginInstantiators: []iopipe.PluginInstantiator{
			iopipe.LoggerPlugin(
				iopipe.LoggerPluginConfig{
					Key: "meow",
				},
			),
		},
	})
	lambda.Start(agent.WrapHandler(Handler))
}
