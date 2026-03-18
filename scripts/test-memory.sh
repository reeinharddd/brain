#!/bin/bash
# test-memory.sh - Verify MCP Memory Server (Engram)

[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" # Load nvm
NPX_CMD="${NPX_CMD:-$HOME/.local/bin/npx-nvm}"
TMP_MEMORY_DIR="$(mktemp -d)"
OUT_FILE="$(mktemp)"

trap 'rm -rf "$TMP_MEMORY_DIR" "$OUT_FILE"' EXIT

echo "Testing Brain Memory (MCP Server-Memory)..."

# Use an isolated temporary store and a full MCP handshake.
{
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"brain-test","version":"1.0.0"}}}'
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  printf '%s\n' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create_entities","arguments":{"entities":[{"name":"MemoryProtocolTest","entityType":"Verification","observations":["Memory protocol verified"]}]}}}'
  printf '%s\n' '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_nodes","arguments":{"query":"MemoryProtocolTest"}}}'
} | "$NPX_CMD" -y @modelcontextprotocol/server-memory "$TMP_MEMORY_DIR" > "$OUT_FILE" 2>&1

if grep -q "MemoryProtocolTest" "$OUT_FILE"; then
    echo "[ok] Entity creation and retrieval verified via MCP Tools API."
else
    echo "[fail] Failed to create or retrieve entity."
    cat "$OUT_FILE"
    exit 1
fi

echo "[ok] Memory (Engram/MCP) Functional."
exit 0
