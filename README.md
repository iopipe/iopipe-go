# IOpipe Agent for Go (beta)

[![GoDoc][1]][2] [![Apache 2.0 licensed][3]][4] [![Go Report Card][5]][6] [![Slack][7]][8]

[1]: https://godoc.org/github.com/iopipe/iopipe-go?status.svg
[2]: https://godoc.org/github.com/iopipe/iopipe-go
[3]: https://img.shields.io/badge/License-Apache%202.0-blue.svg
[4]: LICENSE
[5]: https://goreportcard.com/badge/github.com/iopipe/iopipe-go
[6]: https://goreportcard.com/report/github.com/iopipe/iopipe-go
[7]: https://img.shields.io/badge/chat-slack-ff69b4.svg
[8]: https://iopipe.now.sh)

This package provides analytics and distributed tracing for event-driven applications running on AWS Lambda.

- [Installation](#installation)
- [Usage](#usage)
  - [Configuration](#configuration)
  - [Contexts](#contexts)
  - [Custom Metrics](#custom-metrics)
  - [Labels](#labels)
  - [Reporting Errors](#reporting-errors)
- [Running Tests](#running-tests)
- [Contributing](#contributing)
- [License](#license)

## Installation

Using `go get`:

```bash
go get github.com/iopipe/iopipe-go
```

Using `dep`:

```bash
dep ensure -add github.com/iopipe/iopipe-go
```

## Usage

Set the `IOPIPE_TOKEN` environment variable to [your project token](https://dashboard.iopipe.com/install),

import this library, instantiate an agent, and wrap your handler that you expose to AWS:

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

The `iopipe.Config` struct offers further options for configuring how your function interacts with IOpipe, please refer
to the [godoc](https://godoc.org/github.com/iopipe/iopipe-go#Config)for more information.

### Configuration

The following may be set via the `iopipe.Config{}` struct passed to the `iopipe.NewAgent()` initializer:

#### `Token` (*string: required)

Your IOpipe project token. If not supplied, the environment variable `IOPIPE_TOKEN` will be used if present. [Find your project token](https://dashboard.iopipe.com/install)

#### `Debug` (*bool: optional = false)

Debug mode will log all data sent to IOpipe servers. This is also a good way to evaluate the sort of data that IOpipe is receiving from your application. If not supplied, the environment variable `IOPIPE_DEBUG` will be used if present.

#### `TimeoutWindow` (*time.Duration: optional = 150)

By default, IOpipe will capture timeouts by exiting your function 150 milliseconds early from the AWS configured timeout, to allow time for reporting. You can disable this feature by setting `timeout_window` to `0` in your configuration. If not supplied, the environment variable `IOPIPE_TIMEOUT_WINDOW` will be used if present.

#### `Enabled` (*bool: optional = true)

Conditionally enable/disable the agent. The environment variable `IOPIPE_ENABLED` will also be checked.

### Contexts

The IOpipe agent wraps the `lambdacontext.LambdaContext`. So instead of doing this:

```go
import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func hello(ctx context.Context) (string, error) {
	context, _ := lambdacontext.FromContext(ctx)

	return fmt.Sprintf("My requestId is %s", context.AwsRequestID), nil
}

func main() {
	lambda.Start(agent.WrapHandler(hello))
}
```

You can do this:

```go
import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func hello(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	return fmt.Sprintf("My requestId is %s", context.AwsRequestID), nil
}

func main() {
	lambda.Start(agent.WrapHandler(hello))
}
```

And the `lambdacontext.LambdaContext` will be embedded in `context`. In addition to this, `iopipe.FromContext()` also
attaches `context.IOpipe` which exposes methods to instrument your functions. See the sections below for examples.

### Custom Metrics

You can log custom values in the data sent upstream to IOpipe using the following syntax:

```go
import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func hello(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	// numerical (int, float) and string types supported for values
	context.IOpipe.Metric("my_metric", 42)

	return "Hello ƛ!", nil
}

func main() {
	lambda.Start(agent.WrapHandler(hello))
}
```

Metric key names are limited to 128 characters, and string values are limited to 1024 characters.

### Labels

Invocation labels can be sent to IOpipe by calling the `Label` method with a string (limit of 128 characters):

```go
import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func hello(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	// the name of the label must be a string
	context.IOpipe.Label("this-invocation-is-special")

	return "Hello ƛ!", nil
}

func main() {
	lambda.Start(agent.WrapHandler(hello))
}
```

### Reporting Errors

The IOpipe agent will automatically recover, trace and re-panic any unhandled panics in your function. If you want to trace errors in your case, you can use the `.Error(err)` method. This will add the error to the current report.

```go
import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iopipe/iopipe-go"
)

var agent = iopipe.NewAgent(iopipe.Config{})

func hello(ctx context.Context) (string, error) {
	context, _ := iopipe.FromContext(ctx)

	thing, err := doSomething()

	if err != nil {
	  context.IOpipe.Error(err)
	}

	return "Hello ƛ!", nil
}

func main() {
	lambda.Start(agent.WrapHandler(hello))
}
```

It is important to note that a report is sent to IOpipe when `Error()` is called. So you should only record exceptions this way for failure states. For caught exceptions that are not a failure state, it is recommended to use custom metrics.

You also don't need to use `Error()` if the error is being returned as the second return value of the function. IOpipe will add that error to the report for you automatically.

## Running Tests

The tests use [Convey](https://github.com/smartystreets/goconvey/), so make sure that is installed:

```bash
go get github.com/smartystreets/goconvey
```

We use `dep` for package dependencies, to installed them:

```bash
dep ensure
```

To run the tests:

```bash
go test -v
```

## Contributing

Please refer to our [code of conduct](https://github.com/iopipe/iopipe-go/blob/master/CODE_OF_CONDUCT.md). Please follow it in all your interactions with the project.

## License

Apache 2.0
