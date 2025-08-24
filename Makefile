BINARY=cursor-rules
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CURSOR_PATH?=$(shell which cursor)
.DEFAULT_GOAL := all

# Release matrix
RELEASE_PLATFORMS=linux/amd64 linux/arm64 darwin/arm64 windows/amd64
DIST_DIR=dist

.PHONY: build
build:
	@go build -ldflags "-X 'main.Version=$${VERSION:-dev}'" -o bin/$(BINARY) ./cmd/cursor-rules

.PHONY: install
install:
	@go install -ldflags "-X 'main.Version=$${VERSION:-dev}'" ./cmd/cursor-rules

.PHONY: run
run:
	@./bin/$(BINARY) $(args)

.PHONY: test
test:
	@go test ./...

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: ext-build
ext-build:
	@# Build VSCode extension: compile TypeScript from extension/src to extension/out
	@cd extension && pnpm install --no-frozen-lockfile && pnpm build

.PHONY: ext-prepare
ext-prepare:
	@# Ensure LICENSE exists for VSCE packaging
	@cp -f LICENSE extension/LICENSE

.PHONY: ext-version
ext-version:
	@# Update extension/package.json version when VERSION is provided
	@node -e "const fs=require('fs');const p=require('./extension/package.json');const v=process.env.VERSION; if(v){p.version=v; fs.writeFileSync('./extension/package.json', JSON.stringify(p,null,2)+'\n'); console.log('Set extension version to', v);} else { console.log('VERSION not set; leaving extension version as', p.version);}"

.PHONY: ext-install
ext-install:
	@# Package and install extension locally (user must have code CLI available)
	@echo "Installing the cursor-rules cli binary in $(GOPATH)/bin"
	@make install
	@echo "Cursor-rules binary: $$(ls -1 $(GOPATH)/bin/cursor-rules)"
	@# Ensure the extension is packaged (creates .vsix)
	@make ext-package
	@cd extension && VSIX_FILE=$$(ls -1t *.vsix | head -n1) && \
	BASE_NAME=$$(node -e 'console.log(require("./package.json").name)') && \
	echo "VSIX packaged at: $$(pwd)/$$VSIX_FILE" && \
	echo "Versionless VSIX available at: $$(pwd)/$$BASE_NAME.vsix" && \
	echo "Cursor CLI does not support silent VSIX install. Install manually:" && \
	echo "In Cursor: Command Palette → Extensions: Install from VSIX... → select the VSIX above or use $$BASE_NAME.vsix."
    
.PHONY: ext-test
ext-test:
	@cd extension && pnpm install --no-frozen-lockfile && pnpm build && pnpm test

.PHONY: ext-package
.PHONY: ext-package
ext-package: ext-prepare ext-build
	@cd extension && npx @vscode/vsce package --no-dependencies
	@cd extension && VSIX_FILE=$$(ls -1t *.vsix | head -n1); \
	BASE_NAME=$$(node -e 'console.log(require("./package.json").name)'); \
	cp "$$VSIX_FILE" "$$BASE_NAME.vsix"; \
	# Remove the versioned VSIX to keep only the versionless file
	rm -f "$$VSIX_FILE"; \
	echo "Created versionless VSIX at: $$(pwd)/$$BASE_NAME.vsix"

.PHONY: build-all
build-all:
	@make build
	@make ext-build

.PHONY: release-artifacts
release-artifacts:
	@make release-binaries VERSION=$(VERSION)
	@make ext-version VERSION=$(VERSION)
	@make ext-package

.PHONY: release-binaries
release-binaries:
	@rm -rf $(DIST_DIR) && mkdir -p $(DIST_DIR)
	@version=$${VERSION:-dev}; \
	for p in $(RELEASE_PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		echo "Building $$os/$$arch version $$version"; \
		out="$(DIST_DIR)/$(BINARY)_$${version}_$${os}_$${arch}"; \
		[ "$$os" = "windows" ] && out="$$out.exe"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w -X 'main.Version=$$version'" -o "$$out" ./cmd/$(BINARY); \
	done

.PHONY: help
help:
	@echo "Available targets: build test fmt ext-build ext-install"

.PHONY: all
all: build ext-prepare ext-build ext-install

.PHONY: clean
clean:
	@rm -rf bin/
	@rm -rf extension/out/
	@rm -rf extension/node_modules/
	@rm -rf extension/package-lock.json