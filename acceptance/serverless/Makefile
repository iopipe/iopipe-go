build:
	go get -u github.com/aws/aws-lambda-go/lambda
	go get -u github.com/iopipe/iopipe-go

	env GOOS=linux go build -ldflags="-s -w" -o bin/error error/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/labels labels/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/metrics metrics/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/panic panic/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/success success/main.go