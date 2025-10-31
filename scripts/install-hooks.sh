#!/usr/bin/env bash
#
# Install git hooks for cursor-rules project
#

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

HOOKS_DIR=".git/hooks"

if [ ! -d "$HOOKS_DIR" ]; then
    echo "❌ Not a git repository or hooks directory not found"
    exit 1
fi

echo "Installing git hooks..."

# Install pre-commit hook
cp -f scripts/pre-commit.sh "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Git hooks installed successfully"
echo ""
echo "Installed hooks:"
echo "  - pre-commit: Runs fmt, vet, lint, and tests before commit"
echo ""
echo "To skip hooks for a specific commit, use: git commit --no-verify"

