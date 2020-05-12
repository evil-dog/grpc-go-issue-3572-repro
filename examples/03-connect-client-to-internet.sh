#!/bin/bash

docker network connect bridge examples_greeter-client_1
docker-compose start greeter-client
docker-compose logs -f greeter-client
