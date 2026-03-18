#!/bin/bash
# smoke-real-env.sh - Collect a reproducible smoke test bundle for a real environment run.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
RUN_ID="$(date -u '+%Y%m%dT%H%M%SZ')"
OUT_DIR="$BRAIN_DIR/logs/real-runs/$RUN_ID"
mkdir -p "$OUT_DIR"

run_capture() {
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
  bash "$BRAIN_DIR/scripts/telemetry.sh" record "smoke-$name" "$status" "$duration" "$OUT_DIR/$name.out" >/dev/null 2>&1 || true
  printf '%s\t%s\t%s\n' "$name" "$status" "$duration" >> "$OUT_DIR/summary.tsv"
}

: > "$OUT_DIR/summary.tsv"

run_capture doctor_json bash "$BRAIN_DIR/scripts/doctor.sh" --json --verbose
run_capture docker_status bash "$BRAIN_DIR/docker/start.sh" status
run_capture memory_test bash "$BRAIN_DIR/scripts/test-memory.sh"
run_capture stdio_memory bash "$BRAIN_DIR/scripts/test-stdio-mcp.sh" memory
run_capture stdio_filesystem bash "$BRAIN_DIR/scripts/test-stdio-mcp.sh" filesystem
run_capture stdio_sequential bash "$BRAIN_DIR/scripts/test-stdio-mcp.sh" sequential
run_capture stdio_context7 bash "$BRAIN_DIR/scripts/test-stdio-mcp.sh" context7
run_capture vector_sync bash "$BRAIN_DIR/scripts/vector-sync-qdrant.sh" "$BRAIN_DIR/.brain/codebase-context.ndjson"
run_capture context_pack bash "$BRAIN_DIR/skills/codebase-contextualizer/contextualize.sh" "$BRAIN_DIR"
run_capture docker_helper bash "$BRAIN_DIR/scripts/test-docker-mcp.sh"

cat > "$OUT_DIR/README.md" <<EOF
# Real Environment Smoke Run

- Run ID: $RUN_ID
- Project: $BRAIN_DIR
- Summary: \`summary.tsv\`
- Outputs: one \`.out\` and \`.err\` file per check
EOF

echo "$OUT_DIR"
