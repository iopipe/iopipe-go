FROM golang:1.8
MAINTAINER Kumbirai Tanekha

RUN apt-get update && apt-get install jq
RUN go get -u github.com/jstemmer/go-junit-report
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get github.com/smartystreets/goconvey

ENV PROJECT_DIR=/go/src/github.com/iopipe/iopipe-go
RUN mkdir -p $PROJECT_DIR
WORKDIR $PROJECT_DIR

ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080
