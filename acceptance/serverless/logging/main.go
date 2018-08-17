package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{
	Debug: iopipe.True(),
	Plugins: []iopipe.PluginInstantiator{
		iopipe.LoggerPlugin(iopipe.LoggerPluginConfig{}),
	},
})

func handler(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	context.IOpipe.Log.Debug("I'm a debug message")
	context.IOpipe.Log.Info("I'm an info message")
	context.IOpipe.Log.Warning("I'm a warning message")
	context.IOpipe.Log.Error("I'm an error message")

	return "Loggingsuccessful", nil
}

func main() {
	lambda.Start(agent.WrapHandler(handler))
}
