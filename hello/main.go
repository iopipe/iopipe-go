package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"time"
	"github.com/iopipe/iopipe-go"
	"fmt"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.Body {
	case "error":
		return events.APIGatewayProxyResponse{Body: "it's over 9000" , StatusCode: 90001}, fmt.Errorf("it's over 9000")
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
	token := "TODO: some-token"
	enabled := true
	timeoutWindow := 100 * time.Millisecond

	agent := iopipe.NewAgent(iopipe.AgentConfig{
		Token:         &token,
		TimeoutWindow: &timeoutWindow,
		Enabled:       &enabled,
		PluginInstantiators: &[]iopipe.PluginInstantiator{
			iopipe.LoggerPlugin(
				iopipe.LoggerPluginConfig{
					Key: "meow",
				},
			),
		},
	})
	lambda.Start(agent.WrapHandler(Handler))
}
