#!/usr/bin/env bash
#
# Pre-commit hook for cursor-rules
# Runs go fmt, go vet, golangci-lint, and tests before commit
#

set -e

echo "🔍 Running pre-commit checks..."

# Change to repository root
cd "$(git rev-parse --show-toplevel)"

# 1. Format check
echo "→ Running gofmt..."
unformatted=$(gofmt -l . 2>&1 | grep -v "^vendor/" || true)
if [ -n "$unformatted" ]; then
    echo "❌ Files need formatting (run 'gofmt -w .'):"
    echo "$unformatted"
    exit 1
fi

# 2. Import check
echo "→ Running goimports..."
if command -v goimports >/dev/null 2>&1; then
    unformatted=$(goimports -l . 2>&1 | grep -v "^vendor/" || true)
    if [ -n "$unformatted" ]; then
        echo "❌ Files need import formatting (run 'goimports -w .'):"
        echo "$unformatted"
        exit 1
    fi
else
    echo "⚠️  goimports not found, skipping (install: go install golang.org/x/tools/cmd/goimports@latest)"
fi

# 3. Go vet
echo "→ Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet failed"
    exit 1
fi

# 4. golangci-lint (if available)
echo "→ Running golangci-lint..."
if command -v golangci-lint >/dev/null 2>&1; then
    if ! golangci-lint run --timeout=5m; then
        echo "❌ golangci-lint failed"
        exit 1
    fi
else
    echo "⚠️  golangci-lint not found, skipping (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)"
fi

# 5. Run tests
echo "→ Running tests..."
if ! go test ./... -short; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✅ All pre-commit checks passed!"

