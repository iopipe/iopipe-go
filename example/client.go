package main

import (
    "github.com/iopipe/iopipe-go/example/golambdainvoke"
    "fmt"
    "github.com/aws/aws-lambda-go/events"
)

func main()  {
    response, err := golambdainvoke.Run(8001, events.APIGatewayProxyRequest{
        Body: "simon",
    })
    if err == nil {
        fmt.Println("no errror")
        fmt.Println(string(response))
    } else {
        fmt.Println("with error")
        fmt.Println(err)
    }
}
