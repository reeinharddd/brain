#!/bin/bash
# evals/fixtures/seed-memory-benchmark.sh
# Seeds the memory MCP server with known entities for the retrieval benchmark

set -euo pipefail

TMP_MEM="${1:-$(mktemp -d)}"

resolve_npx() {
  [ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" 2>/dev/null || true
  command -v npx-nvm >/dev/null 2>&1 && npx-nvm -v >/dev/null 2>&1 && { echo "npx-nvm"; return; }
  command -v npx >/dev/null 2>&1 && { echo "npx"; return; }
  echo "ERROR: npx not found" >&2; exit 1
}
NPX="$(resolve_npx)"

seed_entity() {
  local name="$1" obs="$2"
  printf '%s\n' \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"eval","version":"1.0.0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}' \
    "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"create_entities\",\"arguments\":{\"entities\":[{\"name\":\"$name\",\"entityType\":\"Decision\",\"observations\":[\"$obs\"]}]}}}" \
  | "$NPX" -y @modelcontextprotocol/server-memory "$TMP_MEM" >/dev/null 2>&1 || true
}

# Seed test entities
seed_entity "AuthDecision" "JWT with RS256 for authentication due to stateless scalability"
seed_entity "PostgresChoice" "PostgreSQL chosen over MongoDB for relational queries and ACID"
seed_entity "ErrorHandlingRule" "All errors wrapped with context using Result pattern not exceptions"
seed_entity "DeploymentConfig" "Docker Compose on staging Kubernetes on production environment"
seed_entity "TypeScriptPreference" "Strict mode mandatory no explicit any types allowed in codebase"

echo "Memory seeded in $TMP_MEM"
