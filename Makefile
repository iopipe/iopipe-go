build:
    dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/hello hello/main.go
dev:
	dep ensure
	bash
