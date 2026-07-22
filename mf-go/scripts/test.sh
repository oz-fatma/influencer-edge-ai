#!/usr/bin/env bash
#
# test.sh - Run tests with coverage and formatting
#
# Usage:
#   ./scripts/test.sh          - Run all tests
#   ./scripts/test.sh -v        - Verbose output
#   ./scripts/test.sh -cover     - With coverage report
#   ./scripts/test.sh ./path     - Run tests in specific path
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

ARGS=()
COVER=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -cover|--cover)
            COVER=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            ARGS+=("-v")
            shift
            ;;
        *)
            ARGS+=("$1")
            shift
            ;;
    esac
done

if [[ "$COVER" == true ]]; then
    echo "🧪 Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out -covermode=atomic "${ARGS[@]}" ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ Coverage report: coverage.html"
else
    echo "🧪 Running tests..."
    go test -race -count=1 "${ARGS[@]}" ./...
fi

echo "✅ Tests complete"
