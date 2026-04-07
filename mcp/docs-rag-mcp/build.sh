#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Building Docs-RAG MCP Server ==="
echo "Working directory: $(pwd)"
echo

# Step 1: Download dependencies
echo "1. Fetching dependencies..."
go mod tidy 2>&1 | head -20
echo "   ✓ Dependencies fetched"
echo

# Step 2: Run tests
echo "2. Running tests..."
go test ./internal/indexer -v -cover 2>&1 | head -100
GO_TEST_EXIT=$?
echo

# Step 3: Build the binary
echo "3. Building MCP binary..."
mkdir -p bin
go build -o bin/docs-rag-mcp main.go 2>&1 | head -20
if [ -f "bin/docs-rag-mcp" ]; then
    echo "   ✓ Binary built: bin/docs-rag-mcp"
    ls -lh bin/docs-rag-mcp
else
    echo "   ✗ Binary build failed"
fi
echo

# Step 4: Check code quality
echo "4. Checking code quality..."
go fmt ./... 2>&1 | head -10
go vet ./... 2>&1 | head -20
echo "   ✓ Code quality check complete"
echo

echo "=== Build Summary ==="
if [ $GO_TEST_EXIT -eq 0 ]; then
    echo "✓ Tests passed"
else
    echo "⚠ Some tests failed (see output above)"
fi
echo "✓ Build process complete"
