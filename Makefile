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
	@go build -ldflags "-X 'github.com/ZanzyTHEbar/cursor-rules/internal/cli.Version=$${VERSION:-dev}'" -o bin/$(BINARY) ./cmd/cursor-rules

.PHONY: install
install:
	@go install -ldflags "-X 'github.com/ZanzyTHEbar/cursor-rules/internal/cli.Version=$${VERSION:-dev}'" ./cmd/cursor-rules

.PHONY: run
run:
	@./bin/$(BINARY) $(args)

# ==============================================================================
# Testing Targets
# ==============================================================================

# Run all tests with verbose output
.PHONY: test
test:
	@echo "Running all tests..."
	@go test ./... -v

# Run all tests (silent mode)
.PHONY: test-silent
test-silent:
	@go test ./...

# Run unit tests only (fast, no integration/e2e)
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@go test ./internal/... ./cli/... -v -short

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test ./cmd/cursor-rules/... -v

# Run E2E tests (requires binary to be built)
.PHONY: test-e2e
test-e2e: build
	@echo "Running E2E tests..."
	@go test ./e2e_test.go -v

# Run quick tests (short mode, skips slow tests)
.PHONY: test-quick
test-quick:
	@echo "Running quick tests..."
	@go test ./... -short -v

# Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

# Run tests with race detector
.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	@go test ./... -race -v

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem -run=^$$

# Run benchmarks with CPU profiling
.PHONY: bench-cpu
bench-cpu:
	@echo "Running benchmarks with CPU profiling..."
	@go test ./... -bench=. -benchmem -cpuprofile=cpu.prof -run=^$$
	@echo "CPU profile saved to cpu.prof"
	@echo "View with: go tool pprof cpu.prof"

# Run benchmarks with memory profiling
.PHONY: bench-mem
bench-mem:
	@echo "Running benchmarks with memory profiling..."
	@go test ./... -bench=. -benchmem -memprofile=mem.prof -run=^$$
	@echo "Memory profile saved to mem.prof"
	@echo "View with: go tool pprof mem.prof"

# Run tests with verbose output and show coverage
.PHONY: test-verbose
test-verbose:
	@go test ./... -v -cover

# Run tests and open coverage report in browser
.PHONY: test-coverage-view
test-coverage-view: test-coverage
	@echo "Opening coverage report in browser..."
	@which open > /dev/null && open coverage.html || \
	 which xdg-open > /dev/null && xdg-open coverage.html || \
	 echo "Please open coverage.html manually"

# Run all test suites (unit, integration, e2e)
.PHONY: test-all
test-all: test-unit test-integration test-e2e
	@echo "All test suites completed successfully!"

# Run tests in CI mode (with coverage and race detection)
.PHONY: test-ci
test-ci:
	@echo "Running tests in CI mode..."
	@go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out

# Clean test artifacts
.PHONY: test-clean
test-clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -f cpu.prof mem.prof
	@find . -name "*.test" -delete
	@echo "Test artifacts cleaned"

# Watch tests (requires entr: brew install entr or apt-get install entr)
.PHONY: test-watch
test-watch:
	@echo "Watching for changes and running tests..."
	@find . -name "*.go" | entr -c make test-quick

.PHONY: fmt
fmt:
	@go fmt ./...

# ==============================================================================
# Code Quality Targets
# ==============================================================================

# Run golangci-lint
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run

# Run golangci-lint with auto-fix
.PHONY: lint-fix
lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run --fix

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run all quality checks
.PHONY: check
check: fmt vet lint test
	@echo "All quality checks passed!"

.PHONY: build-all
build-all:
	@make build

.PHONY: release-artifacts
release-artifacts:
	@make release-binaries VERSION=$(VERSION)

.PHONY: release-binaries
release-binaries:
	@rm -rf $(DIST_DIR) && mkdir -p $(DIST_DIR)
	@version=$${VERSION:-dev}; \
	for p in $(RELEASE_PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		echo "Building $$os/$$arch version $$version"; \
		out="$(DIST_DIR)/$(BINARY)_$${version}_$${os}_$${arch}"; \
		[ "$$os" = "windows" ] && out="$$out.exe"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w -X 'github.com/ZanzyTHEbar/cursor-rules/internal/cli.Version=$$version'" -o "$$out" ./cmd/$(BINARY); \
	done

.PHONY: help
help:
	@echo "=== Cursor Rules Makefile ==="
	@echo ""
	@echo "Build Targets:"
	@echo "  build              - Build the CLI binary"
	@echo "  install            - Install the CLI binary to GOPATH/bin"
	@echo "  build-all          - Build CLI binary"
	@echo "  release-binaries   - Build release binaries for all platforms"
	@echo ""
	@echo "Test Targets:"
	@echo "  test               - Run all tests (verbose)"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-e2e           - Run end-to-end tests"
	@echo "  test-quick         - Run quick tests (short mode)"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-race          - Run tests with race detector"
	@echo "  test-all           - Run all test suites"
	@echo "  test-ci            - Run tests in CI mode"
	@echo "  bench              - Run benchmarks"
	@echo ""
	@echo "Code Quality Targets:"
	@echo "  fmt                - Format Go code"
	@echo "  lint               - Run golangci-lint"
	@echo "  lint-fix           - Run golangci-lint with auto-fix"
	@echo "  vet                - Run go vet"
	@echo "  check              - Run all quality checks (fmt, vet, lint, test)"
	@echo "  install-hooks      - Install git pre-commit hooks"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean              - Clean build artifacts"
	@echo "  test-clean         - Clean test artifacts"
	@echo "  help               - Show this help message"

# ==============================================================================
# Git Hooks
# ==============================================================================

.PHONY: install-hooks
install-hooks:
	@echo "Installing git hooks..."
	@bash scripts/install-hooks.sh

.PHONY: all
all: build install

.PHONY: clean
clean:
	@rm -rf bin/
