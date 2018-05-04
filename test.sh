#!/usr/bin/env bash
set -e

./build.sh

if [[ $(serverless info) ]]; then
   echo "Testing on existing deployment"
else
    serverless deploy -v
    echo "Testing on a shiny new deployment :)"
fi

serverless deploy function -f hello
serverless invoke -f hello -l -d '{"body": "hello"}'
