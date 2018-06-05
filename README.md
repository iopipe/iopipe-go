# IOpipe Agent for Go (alpha)
This package provides analytics and distributed tracing for event-driven applications running on AWS Lambda.

_WARNING! This library is in an alpha state, use at your own risk!_

- [Installation](#installation)
- [Contributing](#contributing)
- [License](#license)

## Installation

Set the `IOPIPE_TOKEN` environment variable to [your project token](https://dashboard.iopipe.com/install),
import this library, instantiate an agent, and wrap your handler that you expose
to AWS. An example follows:

```go
import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)


var agent = iopipe.NewAgent(iopipe.Config{})

func hello() (string, error) {
	return "Hello ƛ!", nil
}

func main() {
	lambda.Start(agent.WrapHandler(hello))
}
```

The `Config` struct offers further options for configuring how your function
interacts with IOpipe, please refer to the [godoc](https://godoc.org/github.com/iopipe/iopipe-go#Config)
for more information.

## Contributing

Please refer to our [code of conduct](https://github.com/iopipe/iopipe-go/blob/master/CODE_OF_CONDUCT.md). Please follow it in all your interactions with the project.

## License

Apache 2.0
