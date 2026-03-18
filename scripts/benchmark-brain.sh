#!/bin/bash
# benchmark-brain.sh - Capture simple timing benchmarks for the brain workflow.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
RUN_ID="$(date -u '+%Y%m%dT%H%M%SZ')"
OUT_DIR="$BRAIN_DIR/logs/benchmarks/$RUN_ID"
mkdir -p "$OUT_DIR"
RESULTS_JSON="$OUT_DIR/results.json"

measure() {
  local name="$1"
  shift
  local start end duration status
  start="$(date +%s%3N 2>/dev/null || date +%s000)"
  if "$@" >"$OUT_DIR/$name.out" 2>"$OUT_DIR/$name.err"; then
    status="ok"
  else
    status="fail"
  fi
  end="$(date +%s%3N 2>/dev/null || date +%s000)"
  duration="$((end - start))"
  printf '%s\t%s\t%s\n' "$name" "$status" "$duration" >> "$OUT_DIR/results.tsv"
  bash "$BRAIN_DIR/scripts/telemetry.sh" record "benchmark-$name" "$status" "$duration" "$OUT_DIR/$name.out" >/dev/null 2>&1 || true
}

: > "$OUT_DIR/results.tsv"

measure doctor bash "$BRAIN_DIR/scripts/doctor.sh" --json --verbose
measure context_pack bash "$BRAIN_DIR/skills/codebase-contextualizer/contextualize.sh" "$BRAIN_DIR"
measure vector_sync bash "$BRAIN_DIR/scripts/vector-sync-qdrant.sh" "$BRAIN_DIR/.brain/codebase-context.ndjson"
measure memory_test bash "$BRAIN_DIR/scripts/test-memory.sh"
measure docker_helper bash "$BRAIN_DIR/scripts/test-docker-mcp.sh"

python3 - "$OUT_DIR/results.tsv" "$RESULTS_JSON" <<'PY'
import json
import sys
from pathlib import Path

tsv_path = Path(sys.argv[1])
json_path = Path(sys.argv[2])
rows = []
for line in tsv_path.read_text(encoding="utf-8").splitlines():
    if not line.strip():
        continue
    name, status, duration = line.split("\t")
    rows.append({"name": name, "status": status, "duration_ms": int(duration)})
json_path.write_text(json.dumps({"benchmarks": rows}, indent=2), encoding="utf-8")
PY

cat > "$OUT_DIR/README.md" <<EOF
# Brain Benchmark Run

- Run ID: $RUN_ID
- Results table: \`results.tsv\`
- Results JSON: \`results.json\`
EOF

echo "$OUT_DIR"
