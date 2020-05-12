#!/bin/bash

docker-compose up -d
docker-compose logs -f greeter-client
