build:
	go get -u github.com/aws/aws-lambda-go/lambda
	go get -u github.com/iopipe/iopipe-go

	env GOOS=linux go build -ldflags="-s -w" -o bin/api-gateway api-gateway/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/api-trigger api-trigger/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/baseline baseline/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/baseline-coldstart baseline-coldstart/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/coldstart coldstart/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/error error/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/event event/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/handled-error handled-error/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/labels labels/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/metrics metrics/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/panic panic/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/success success/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/timeout timeout/main.go