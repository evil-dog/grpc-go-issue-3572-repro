#!/bin/bash

docker build -f Dockerfile-client -t grpc-example-client .
docker build -f Dockerfile-client-orig -t grpc-example-client-orig .
docker build -f Dockerfile-server -t grpc-example-server .
