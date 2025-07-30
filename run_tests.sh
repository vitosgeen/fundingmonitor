#!/bin/bash

# Test runner script for fundingmonitor

echo "🧪 Funding Monitor Test Runner"
echo "================================"

case "$1" in
    "unit")
        echo "Running unit tests..."
        go test ./internal/... -v
        ;;
    "integration")
        echo "Running integration tests..."
        go test ./tests/integration/... -v
        ;;
    "e2e")
        echo "Running end-to-end tests..."
        go test ./tests/integration/... -run "TestE2E" -v
        ;;
    "all")
        echo "Running all tests..."
        go test ./... -v
        ;;
    "short")
        echo "Running short tests (excluding E2E)..."
        go test ./internal/... ./tests/integration/... -short -v
        ;;
    *)
        echo "Usage: $0 {unit|integration|e2e|all|short}"
        echo ""
        echo "  unit       - Run unit tests only"
        echo "  integration - Run integration tests only"
        echo "  e2e        - Run end-to-end tests only"
        echo "  all        - Run all tests"
        echo "  short      - Run tests excluding E2E (for CI)"
        echo ""
        echo "Examples:"
        echo "  $0 unit"
        echo "  $0 integration"
        echo "  $0 all"
        exit 1
        ;;
esac 