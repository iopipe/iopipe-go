#!/usr/bin/env bash

PROJECT_NAME=iopipe-go
PROJECT_DIR=github.com/iopipe/iopipe-go

docker build -t $PROJECT_NAME-builder:latest .
docker run -it --rm -p 8080:8080 -v $PWD:/go/src/$PROJECT_DIR $PROJECT_NAME-builder:latest make dev
