#!/bin/bash
# test-memory.sh - Verify MCP Memory Server (Engram)

set -euo pipefail

[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" # Load nvm

resolve_npx_cmd() {
  if [ -n "${NPX_CMD:-}" ] && command -v "${NPX_CMD%% *}" >/dev/null 2>&1; then
    printf '%s\n' "$NPX_CMD"
    return 0
  fi
  if command -v npx-nvm >/dev/null 2>&1 && npx-nvm -v >/dev/null 2>&1; then
    printf '%s\n' "npx-nvm"
    return 0
  fi
  if command -v npx >/dev/null 2>&1; then
    printf '%s\n' "npx"
    return 0
  fi
  echo "ERROR: neither npx-nvm nor npx is available" >&2
  exit 1
}

NPX_CMD="$(resolve_npx_cmd)"
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
