#!/bin/bash

docker-compose -f docker-compose.orig-client.yml up -d
docker-compose -f docker-compose.orig-client.yml logs -f greeter-client
