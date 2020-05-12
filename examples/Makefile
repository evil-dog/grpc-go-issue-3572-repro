.PHONY: client server
VERSION := 1

client:
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/greeter-client helloworld/greeter_client_issue/main.go

client-orig:
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/greeter-client-orig helloworld/greeter_client/main.go

server:
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/greeter-server helloworld/greeter_server/main.go

dev-requirements:
	go mod download
	go install golang.org/x/lint/golint
