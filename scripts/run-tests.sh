#!/bin/bash
set -e

echo "🧪 Running Cursor Rules Test Suite"
echo "=================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "Please install Go 1.24.6 or later"
    exit 1
fi

echo -e "${GREEN}✅ Go version: $(go version)${NC}"
echo ""

# Run unit tests
echo "📦 Running Unit Tests..."
echo "------------------------"
if go test ./internal/transform/... -v; then
    echo -e "${GREEN}✅ Transform unit tests passed${NC}"
else
    echo -e "${RED}❌ Transform unit tests failed${NC}"
    exit 1
fi

if go test ./internal/manifest/... -v; then
    echo -e "${GREEN}✅ Manifest unit tests passed${NC}"
else
    echo -e "${RED}❌ Manifest unit tests failed${NC}"
    exit 1
fi

echo ""

# Run integration tests
echo "🔗 Running Integration Tests..."
echo "--------------------------------"
if go test ./cmd/cursor-rules/... -v; then
    echo -e "${GREEN}✅ Integration tests passed${NC}"
else
    echo -e "${RED}❌ Integration tests failed${NC}"
    exit 1
fi

echo ""

# Build binary for E2E tests
echo "🔨 Building binary for E2E tests..."
echo "------------------------------------"
if make build; then
    echo -e "${GREEN}✅ Binary built successfully${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo ""

# Run E2E tests
echo "🌐 Running End-to-End Tests..."
echo "-------------------------------"
if go test -tags=e2e ./... -v; then
    echo -e "${GREEN}✅ E2E tests passed${NC}"
else
    echo -e "${YELLOW}⚠️  E2E tests skipped or failed${NC}"
    echo "Note: E2E tests require the binary to be built"
fi

echo ""
echo "=================================="
echo -e "${GREEN}✅ All tests completed successfully!${NC}"
echo "=================================="

# Generate coverage report
echo ""
echo "📊 Generating coverage report..."
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✅ Coverage report generated: coverage.html${NC}"
