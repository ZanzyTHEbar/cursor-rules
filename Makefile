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

.PHONY: ext-build
ext-build:
	# Build VSCode extension: compile TypeScript from extension/src to extension/out
	cd extension && pnpm build

.PHONY: ext-install
ext-install:
	# Package and install extension locally (user must have code CLI available)
	cd extension && pnpm build && vsce package && code --install-extension *.vsix || true

.PHONY: help
help:
	@echo "Available targets: build run test fmt ext-build ext-install"
