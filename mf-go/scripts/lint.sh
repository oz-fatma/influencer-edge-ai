#!/usr/bin/env bash
#
# lint.sh - Run linters and formatters
#
# Usage:
#   ./scripts/lint.sh          - Run all linters
#   ./scripts/lint.sh -fix     - Auto-fix issues
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

FIX=false
if [[ "${1:-}" == "-fix" ]]; then
    FIX=true
fi

echo "🔍 Running linters..."

# Format check
echo "  📝 Checking formatting..."
if ! gofmt -l . | grep -q .; then
    if [[ "$FIX" == true ]]; then
        echo "    Fixing formatting..."
        gofmt -w .
        echo "    ✅ Formatting fixed"
    else
        echo "    ❌ Code is not formatted. Run: gofmt -w ."
        exit 1
    fi
else
    echo "    ✅ Code is formatted"
fi

# Vet check
echo "  🔎 Running go vet..."
if go vet ./...; then
    echo "    ✅ go vet passed"
else
    echo "    ❌ go vet found issues"
    exit 1
fi

# golangci-lint (if available)
if command -v golangci-lint &>/dev/null; then
    echo "  🔧 Running golangci-lint..."
    if [[ "$FIX" == true ]]; then
        golangci-lint run --fix ./...
    else
        golangci-lint run ./...
    fi
    echo "    ✅ golangci-lint passed"
else
    echo "    ⚠️  golangci-lint not installed (optional)"
fi

echo "✅ Linting complete"
