#!/bin/bash
# test-memory.sh — Verify MCP Memory Server (Engram)

[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" # Load nvm

echo "Testing Brain Memory (MCP Server-Memory)..."

# 1. Store a test entity using the 'call_tool' (MCP Spec)
# Note: In stdio mode, we just pipe the JSON-RPC
JSON_CALL='{"jsonrpc":"2.0","method":"tools/call","params":{"name":"create_entities","arguments":{"entities":[{"name":"BrainTestEntity","entityType":"Verification","observations":["Verified at '$(date +%Y-%m-%d)'"]}]}},"id":1}'

echo "$JSON_CALL" | npx -y @modelcontextprotocol/server-memory > /tmp/memory_test_out.json 2>&1

if grep -q "BrainTestEntity" /tmp/memory_test_out.json || grep -q "added" /tmp/memory_test_out.json; then
    echo "✓ Entity creation verified via Tools API."
else
    echo "✗ Failed to create entity."
    cat /tmp/memory_test_out.json
    exit 1
fi

echo "✓ Memory (Engram/MCP) Functional."
exit 0
