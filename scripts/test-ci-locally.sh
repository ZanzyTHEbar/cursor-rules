#!/usr/bin/env bash
#
# Test CI locally - simulates GitHub Actions CI workflow
# Run this before pushing to ensure CI will pass
#

set -e

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

echo "════════════════════════════════════════════════════════════════"
echo "  Testing CI Locally - Simulating GitHub Actions Workflow"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Track failures
FAILURES=0

# Helper function to run a step
run_step() {
    local name="$1"
    local cmd="$2"
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "▶ $name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if eval "$cmd"; then
        echo "✅ PASS: $name"
        echo ""
        return 0
    else
        echo "❌ FAIL: $name"
        echo ""
        ((FAILURES++))
        return 1
    fi
}

# ============================================================================
# Job 1: Test Go Application
# ============================================================================
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Job 1: Test Go Application                                  ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

run_step "Download dependencies" "go mod download" || true
run_step "Run go vet" "go vet ./..." || true
run_step "Run tests" "make test" || true
run_step "Run tests with coverage" "make test-coverage" || true
run_step "Build binary" "make build" || true

# ============================================================================
# Job 2: Lint Go Code
# ============================================================================
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Job 2: Lint Go Code                                         ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

if command -v golangci-lint >/dev/null 2>&1; then
    run_step "Run golangci-lint" "golangci-lint run --timeout=10m" || true
else
    echo "⚠️  golangci-lint not found - install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    echo "❌ FAIL: Lint (skipped - tool not available)"
    ((FAILURES++))
fi

# ============================================================================
# Job 3: Security Audit
# ============================================================================
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Job 3: Security Audit                                       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

if command -v gosec >/dev/null 2>&1; then
    run_step "Run gosec (security scan)" "gosec -exclude=G104,G301,G304,G306 -fmt=sarif -out=gosec.sarif ./..." || true
    
    # Check SARIF file was created
    if [ -f gosec.sarif ]; then
        echo "✅ SARIF file created successfully"
        if command -v jq >/dev/null 2>&1; then
            issues=$(jq -r '.runs[0].results | length' gosec.sarif 2>/dev/null || echo "unknown")
            echo "   Security issues found: $issues"
        fi
        rm -f gosec.sarif
    else
        echo "❌ SARIF file not created"
        ((FAILURES++))
    fi
else
    echo "⚠️  gosec not found - install: go install github.com/securego/gosec/v2/cmd/gosec@latest"
    echo "❌ FAIL: Security audit (skipped - tool not available)"
    ((FAILURES++))
fi

# ============================================================================
# Job 4: Integration Tests
# ============================================================================
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Job 4: Integration Tests                                    ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

run_step "Run integration tests" "bash scripts/integration_test.sh" || true

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  CI Simulation Complete"
echo "════════════════════════════════════════════════════════════════"
echo ""

if [ $FAILURES -eq 0 ]; then
    echo "✅ SUCCESS: All critical CI checks passed!"
    echo ""
    echo "✓ Tests: PASS"
    echo "✓ Linting: PASS"
    echo "✓ Security: PASS"
    echo "✓ Integration: PASS"
    echo ""
    echo "Your Go code changes are ready to push. CI should pass successfully."
    echo ""
    exit 0
else
    echo "❌ FAILURE: $FAILURES critical check(s) failed"
    echo ""
    echo "Please fix the failing checks before pushing."
    echo "CI will likely fail with these issues."
    echo ""
    exit 1
fi

