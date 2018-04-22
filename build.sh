#!/usr/bin/env bash

PROJECT_NAME=iopipe-go
PROJECT_DIR=github.com/iopipe/$PROJECT_NAME

docker build -t $PROJECT_NAME-builder:latest .
docker run -it --rm -v $PWD:/go/src/$PROJECT_DIR $PROJECT_NAME-builder:latest make
