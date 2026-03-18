#!/bin/bash

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
PROJECT_ROOT="${1:-$PWD}"
OUTPUT_DIR="$PROJECT_ROOT/.brain"
VECTOR_CONFIG="$BRAIN_DIR/memory/vector-config.json"
GENERATED_AT="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

mkdir -p "$OUTPUT_DIR"

NAMESPACE="$(bash "$BRAIN_DIR/scripts/memory-namespace.sh" "$PROJECT_ROOT")"
STACK_CONTEXT="$(bash "$BRAIN_DIR/scripts/render-skill-context.sh" "$PROJECT_ROOT")"
INDEX_PATH="$(bash "$BRAIN_DIR/scripts/vector-context-index.sh" "$PROJECT_ROOT")"
INDEX_DOC_COUNT="$(wc -l < "$INDEX_PATH" | tr -d ' ')"
VECTOR_STATUS="disabled"
VECTOR_DETAILS="vector backend disabled"

if [ -f "$VECTOR_CONFIG" ] && python3 - "$VECTOR_CONFIG" <<'PY'
import json, sys
config = json.load(open(sys.argv[1], "r", encoding="utf-8"))
raise SystemExit(0 if config.get("enabled") else 1)
PY
then
  if VECTOR_OUTPUT="$(bash "$BRAIN_DIR/scripts/vector-sync-qdrant.sh" "$INDEX_PATH" 2>&1)"; then
    VECTOR_STATUS="live"
    VECTOR_DETAILS="$VECTOR_OUTPUT"
  else
    VECTOR_STATUS="unavailable"
    VECTOR_DETAILS="$(printf '%s\n' "$VECTOR_OUTPUT" | tail -n 1)"
  fi
fi

{
  echo "# Codebase Context Pack"
  echo
  echo "- Generated at: $GENERATED_AT"
  echo "- Project root: $PROJECT_ROOT"
  echo "- Namespace: $NAMESPACE"
  echo "- Context index: $INDEX_PATH"
  echo "- Context documents: $INDEX_DOC_COUNT"
  echo "- Vector status: $VECTOR_STATUS"
  echo "- Vector details:"
  printf '%s\n' "$VECTOR_DETAILS" | sed 's/^/  /'
  echo
  echo "$STACK_CONTEXT"
} > "$OUTPUT_DIR/codebase-context.md"

echo "$OUTPUT_DIR/codebase-context.md"
