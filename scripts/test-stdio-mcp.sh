#!/bin/bash
# test-stdio-mcp.sh - Verify stdio MCP startup and initialize handshake.

set -euo pipefail

[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh"

NPX_CMD="${NPX_CMD:-$HOME/.local/bin/npx-nvm}"
SERVICE="${1:-}"
OUT_FILE="$(mktemp)"
trap 'rm -f "$OUT_FILE"' EXIT

if [ -z "$SERVICE" ]; then
  echo "Usage: bash ~/.brain/scripts/test-stdio-mcp.sh <memory|filesystem|sequential|context7|github>" >&2
  exit 1
fi

case "$SERVICE" in
  memory)
    CMD=("$NPX_CMD" -y @modelcontextprotocol/server-memory "$HOME/.brain/memory")
    EXPECT='\"result\"'
    ;;
  filesystem)
    CMD=("$NPX_CMD" -y @modelcontextprotocol/server-filesystem "$HOME")
    EXPECT='\"result\"'
    ;;
  sequential)
    CMD=("$NPX_CMD" -y @modelcontextprotocol/server-sequential-thinking)
    EXPECT='\"result\"'
    ;;
  context7)
    CMD=("$NPX_CMD" -y @upstash/context7-mcp@latest)
    EXPECT='\"result\"'
    ;;
  github)
    CMD=("$NPX_CMD" -y @modelcontextprotocol/server-github)
    EXPECT='\"result\"'
    ;;
  *)
    echo "ERROR: unsupported service '$SERVICE'" >&2
    exit 1
    ;;
esac

{
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"brain-stdio-test","version":"1.0.0"}}}'
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
} | timeout 20s "${CMD[@]}" > "$OUT_FILE" 2>&1 || true

if grep -Eq "$EXPECT" "$OUT_FILE"; then
  echo "[ok] $SERVICE stdio MCP initialize handshake succeeded."
  exit 0
fi

echo "[fail] $SERVICE stdio MCP initialize handshake failed."
cat "$OUT_FILE"
exit 1
