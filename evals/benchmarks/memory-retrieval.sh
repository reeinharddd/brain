#!/bin/bash
# evals/benchmarks/memory-retrieval.sh
# Measures memory retrieval quality: does the system surface relevant context?
#
# Methodology:
# 1. Seeds memory with known entities
# 2. Queries with semantically related but not identical queries  
# 3. Measures recall (were seeded entities retrieved?)
# 4. Reports precision and recall scores
#
# Usage:
#   bash memory-retrieval.sh [--json]

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
JSON_OUTPUT=0
[ "${1:-}" = "--json" ] && JSON_OUTPUT=1

resolve_npx() {
  [ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" 2>/dev/null || true
  command -v npx-nvm >/dev/null 2>&1 && npx-nvm -v >/dev/null 2>&1 && { echo "npx-nvm"; return; }
  command -v npx >/dev/null 2>&1 && { echo "npx"; return; }
  echo "ERROR: npx not found" >&2; exit 1
}
NPX="$(resolve_npx)"

TMP_MEM="$(mktemp -d)"
trap 'rm -rf "$TMP_MEM"' EXIT

# ── Seed known entities ──────────────────────────────────────────────────────
seed_entity() {
  local name="$1" obs="$2"
  printf '%s\n' \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"eval","version":"1.0.0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}' \
    "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"create_entities\",\"arguments\":{\"entities\":[{\"name\":\"$name\",\"entityType\":\"Decision\",\"observations\":[\"$obs\"]}]}}}" \
  | "$NPX" -y @modelcontextprotocol/server-memory "$TMP_MEM" >/dev/null 2>&1
}

query_memory() {
  local query="$1"
  printf '%s\n' \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"eval","version":"1.0.0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}' \
    "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"search_nodes\",\"arguments\":{\"query\":\"$query\"}}}" \
  | "$NPX" -y @modelcontextprotocol/server-memory "$TMP_MEM" 2>/dev/null
}

# Seed test entities
seed_entity "AuthDecision" "JWT with RS256 for authentication due to stateless scalability"
seed_entity "PostgresChoice" "PostgreSQL chosen over MongoDB for relational queries and ACID"
seed_entity "ErrorHandlingRule" "All errors wrapped with context using Result pattern not exceptions"
seed_entity "DeploymentConfig" "Docker Compose on staging Kubernetes on production environment"
seed_entity "TypeScriptPreference" "Strict mode mandatory no explicit any types allowed in codebase"

TOTAL=5
RETRIEVED=0
declare -a RESULT_LINES=()

check_retrieval() {
  local query="$1" expected="$2"
  local output
  output="$(query_memory "$query")"
  if echo "$output" | grep -qi "$expected"; then
    RETRIEVED=$((RETRIEVED + 1))
    RESULT_LINES+=("PASS | $query")
  else
    RESULT_LINES+=("FAIL | $query (expected: $expected)")
  fi
}

check_retrieval "authentication token approach" "AuthDecision"
check_retrieval "database selection rationale" "PostgresChoice"
check_retrieval "how should errors be handled" "ErrorHandlingRule"
check_retrieval "deployment infrastructure" "DeploymentConfig"
check_retrieval "typescript configuration strict" "TypeScriptPreference"

RECALL=$(python3 -c "print(round($RETRIEVED / $TOTAL * 100, 1))")
STATUS="PASS"
[ "$RETRIEVED" -lt 3 ] && STATUS="FAIL"

if [ "$JSON_OUTPUT" -eq 1 ]; then
  python3 -c "
import json
items = []
for line in '''$(printf '%s\n' "${RESULT_LINES[@]}")'''.splitlines():
    if not line.strip(): continue
    status, query = line.split(' | ', 1)
    items.append({'status': status.strip(), 'query': query.strip()})
print(json.dumps({'benchmark': 'memory-retrieval', 'total': $TOTAL, 'retrieved': $RETRIEVED, 'recall_pct': $RECALL, 'status': '$STATUS', 'cases': items}, indent=2))
"
else
  echo ""
  echo "Memory Retrieval Benchmark"
  echo "  Total   : $TOTAL"
  echo "  Retrieved: $RETRIEVED"
  echo "  Recall  : ${RECALL}%"
  echo "  Status  : $STATUS"
  echo ""
  for line in "${RESULT_LINES[@]}"; do
    echo "  $line"
  done
fi

[ "$STATUS" = "PASS" ]
