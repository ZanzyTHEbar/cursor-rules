BINARY=cursor-rules
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: build
build:
	go build -o bin/$(BINARY) ./cmd/cursor-rules

.PHONY: run
run:
	go ./cmd/cursor-rules $(args)

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: ext-build
ext-build:
	# Build VSCode extension: compile TypeScript from extension/src to extension/out
	cd extension && pnpm ci || true && pnpm build

.PHONY: ext-install
ext-install:
	# Package and install extension locally (user must have code CLI available)
	cd extension && pnpm build && npx @vscode/vsce package && code --install-extension *.vsix || true

.PHONY: ext-test
ext-test:
	cd extension && pnpm ci || true && pnpm build && pnpm test

.PHONY: help
help:
	@echo "Available targets: build test fmt ext-build ext-install"

.PHONY: all
all: build ext-build ext-install

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf extension/out/
	rm -rf extension/node_modules/
	rm -rf extension/package-lock.json