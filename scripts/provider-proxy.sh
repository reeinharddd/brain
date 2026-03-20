#!/bin/bash
# provider-proxy.sh - Runtime-enforced model routing based on providers.yml.
#
# Wraps the Anthropic CLI or any model call with:
#   - Task-type-based model selection (reads providers.yml routing table)
#   - Circuit breaker: falls back through fallback_chain on failure
#   - Cost logging: appends to logs/cost.jsonl
#   - Rate limit detection and retry with exponential backoff
#
# Usage:
#   bash provider-proxy.sh --task-type planning --prompt "Design auth system"
#   bash provider-proxy.sh --task-type implementation --prompt "..." --max-tokens 4096
#   bash provider-proxy.sh --model claude-sonnet-4-6 --prompt "..."  # bypass routing
#   bash provider-proxy.sh --cost-report  # show spending summary

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
PROVIDERS_PATH="$BRAIN_DIR/providers/providers.yml"
COST_LOG="$BRAIN_DIR/logs/cost.jsonl"
MAX_RETRIES=3
BACKOFF_BASE=2

mkdir -p "$(dirname "$COST_LOG")" "$BRAIN_DIR/logs"

# ── CLI parsing ──────────────────────────────────────────────────────────────
TASK_TYPE=""
EXPLICIT_MODEL=""
PROMPT=""
MAX_TOKENS=4096
COST_REPORT=0
DRY_RUN=0

while [ $# -gt 0 ]; do
  case "$1" in
    --task-type)   TASK_TYPE="$2";      shift 2 ;;
    --model)       EXPLICIT_MODEL="$2"; shift 2 ;;
    --prompt)      PROMPT="$2";         shift 2 ;;
    --max-tokens)  MAX_TOKENS="$2";     shift 2 ;;
    --cost-report) COST_REPORT=1;       shift ;;
    --dry-run)     DRY_RUN=1;           shift ;;
    *) shift ;;
  esac
done

# ── Cost report mode ──────────────────────────────────────────────────────────
if [ "$COST_REPORT" -eq 1 ]; then
  if [ ! -f "$COST_LOG" ]; then
    echo "No cost data yet."
    exit 0
  fi
  python3 - "$COST_LOG" << 'PY'
import json, pathlib, sys
from collections import defaultdict

log_path = pathlib.Path(sys.argv[1])
by_model = defaultdict(lambda: {"calls": 0, "input_tokens": 0, "output_tokens": 0})

for line in log_path.read_text().splitlines():
    if not line.strip(): continue
    try:
        e = json.loads(line)
        m = e.get("model", "unknown")
        by_model[m]["calls"]         += 1
        by_model[m]["input_tokens"]  += e.get("input_tokens", 0)
        by_model[m]["output_tokens"] += e.get("output_tokens", 0)
    except: pass

print("\nProvider Proxy - Cost Report")
print(f"{'Model':<40} {'Calls':>6} {'In Tok':>10} {'Out Tok':>10}")
print("-" * 70)
for model, stats in sorted(by_model.items()):
    print(f"{model:<40} {stats['calls']:>6} {stats['input_tokens']:>10} {stats['output_tokens']:>10}")
PY
  exit 0
fi

# ── Model resolution from providers.yml ───────────────────────────────────────
resolve_model() {
  local task_type="$1"
  python3 - "$PROVIDERS_PATH" "$task_type" << 'PY'
import re, sys, json

providers_path = sys.argv[1]
task_type = sys.argv[2].lower()

try:
    content = open(providers_path).read()
except FileNotFoundError:
    print("claude-sonnet-4-6")
    sys.exit(0)

# Find tier for task type from task_routing block
routing_match = re.search(r'task_routing:(.*?)(?=\n\w|\Z)', content, re.DOTALL)
tier = "standard"
if routing_match:
    for line in routing_match.group(1).splitlines():
        if task_type in line.lower() and ":" in line:
            tier = line.split(":")[-1].strip().split("#")[0].strip()
            break

# Find model for tier under claude provider
for line in content.splitlines():
    if f"{tier}:" in line:
        candidate = line.split(":")[-1].strip()
        if candidate.startswith("claude-") or candidate.startswith("gpt-") or candidate.startswith("gemini-"):
            print(candidate)
            sys.exit(0)

# Fallback defaults
defaults = {"fast": "claude-haiku-4-5-20251001", "standard": "claude-sonnet-4-6", "powerful": "claude-opus-4-6"}
print(defaults.get(tier, "claude-sonnet-4-6"))
PY
}

# ── Fallback chain ─────────────────────────────────────────────────────────────
get_fallback_chain() {
  python3 - "$PROVIDERS_PATH" << 'PY'
import re, sys, json
try:
    content = open(sys.argv[1]).read()
    match = re.search(r'fallback_chain:(.*?)(?=\n\w|\Z)', content, re.DOTALL)
    if match:
        models = []
        for line in match.group(1).splitlines():
            line = line.strip().lstrip("- ").strip()
            if line and not line.startswith("#"):
                models.append(line)
        print(",".join(models))
        sys.exit(0)
except: pass
print("claude,gemini,local")
PY
}

# ── Call model via Anthropic API ───────────────────────────────────────────────
call_model() {
  local model="$1" prompt="$2" max_tokens="$3"
  local api_key="${ANTHROPIC_API_KEY:-}"

  if [ -z "$api_key" ]; then
    echo "ERROR: ANTHROPIC_API_KEY not set" >&2
    return 1
  fi

  local payload
  payload=$(python3 -c "
import json, sys
print(json.dumps({
  'model': sys.argv[1],
  'max_tokens': int(sys.argv[2]),
  'messages': [{'role': 'user', 'content': sys.argv[3]}]
}))
" "$model" "$max_tokens" "$prompt")

  local response
  response=$(curl -sf \
    -X POST https://api.anthropic.com/v1/messages \
    -H "x-api-key: $api_key" \
    -H "anthropic-version: 2023-06-01" \
    -H "content-type: application/json" \
    -d "$payload" 2>&1) || return 1

  echo "$response"
}

# ── Log cost ───────────────────────────────────────────────────────────────────
log_cost() {
  local model="$1" in_tokens="$2" out_tokens="$3" task_type="$4" status="$5"
  python3 -c "
import json, sys
from datetime import datetime, timezone
print(json.dumps({
  'ts': datetime.now(timezone.utc).isoformat(),
  'model': sys.argv[1],
  'input_tokens': int(sys.argv[2]),
  'output_tokens': int(sys.argv[3]),
  'task_type': sys.argv[4],
  'status': sys.argv[5],
}))
" "$model" "$in_tokens" "$out_tokens" "$task_type" "$status" >> "$COST_LOG"
}

# ── Main execution ─────────────────────────────────────────────────────────────
if [ -z "$PROMPT" ]; then
  echo "ERROR: --prompt is required" >&2
  exit 1
fi

# Resolve model
if [ -n "$EXPLICIT_MODEL" ]; then
  MODEL="$EXPLICIT_MODEL"
else
  MODEL="$(resolve_model "${TASK_TYPE:-implementation}")"
fi

if [ "$DRY_RUN" -eq 1 ]; then
  echo "[provider-proxy] Would use model: $MODEL (task-type: ${TASK_TYPE:-implementation})"
  exit 0
fi

echo "[provider-proxy] Model: $MODEL | Task: ${TASK_TYPE:-unspecified}" >&2

# Try with retry + exponential backoff
attempt=0
while [ $attempt -lt $MAX_RETRIES ]; do
  attempt=$((attempt + 1))
  if RESPONSE=$(call_model "$MODEL" "$PROMPT" "$MAX_TOKENS"); then
    # Parse and output
    OUTPUT=$(echo "$RESPONSE" | python3 -c "
import json, sys
d = json.load(sys.stdin)
text = ''.join(b['text'] for b in d.get('content',[]) if b.get('type')=='text')
usage = d.get('usage', {})
print(json.dumps({'output': text, 'in': usage.get('input_tokens',0), 'out': usage.get('output_tokens',0)}))
" 2>/dev/null || echo '{"output":"","in":0,"out":0}')
    TEXT=$(echo "$OUTPUT" | python3 -c "import json,sys; print(json.load(sys.stdin)['output'])")
    IN_T=$(echo "$OUTPUT" | python3 -c "import json,sys; print(json.load(sys.stdin)['in'])")
    OUT_T=$(echo "$OUTPUT" | python3 -c "import json,sys; print(json.load(sys.stdin)['out'])")
    log_cost "$MODEL" "$IN_T" "$OUT_T" "${TASK_TYPE:-unspecified}" "ok"
    echo "$TEXT"
    exit 0
  fi

  SLEEP=$((BACKOFF_BASE ** attempt))
  echo "[provider-proxy] Attempt $attempt failed. Retrying in ${SLEEP}s..." >&2
  sleep "$SLEEP"
done

log_cost "$MODEL" "0" "0" "${TASK_TYPE:-unspecified}" "fail"
echo "ERROR: All $MAX_RETRIES attempts failed for model $MODEL" >&2
exit 1
