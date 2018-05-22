# IOpipe Agent for Go (alpha)
This package provides analytics and distributed tracing for event-driven applications running on AWS Lambda.

_WARNING! This library is in an alpha state, use at your own risk!_

- [Installation](#installation)
- [Contributing](#contributing)
- [License](#license)

## Installation

```
import (
	"github.com/aws/aws-lambda-go/lambda"
	iopipe "github.com/iopipe/iopipe-go"
)

func hello() (string, error) {
	return "Hello ƛ!", nil
}

func main() {
	agent := iopipe.NewAgent(iopipe.AgentConfig{})

	lambda.Start(agent.WrapHandler(hello))
}
```

## Contributing

Please refer to our [code of conduct](https://github.com/iopipe/iopipe-go/blob/master/CODE_OF_CONDUCT.md). Please follow it in all your interactions with the project.

## License

Apache 2.0
