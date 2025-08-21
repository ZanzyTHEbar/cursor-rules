BINARY=cursor-rules
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: build
build:
	go build -o bin/$(BINARY) ./cmd/cursor-rules

.PHONY: run
run:
	go run ./cmd/cursor-rules $(args)

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt ./...

# TODO: add a target to build the extension, and a target to install it in editor

# TODO: add help target
