#!/bin/bash
# evals/run.sh - Run the full eval suite for brain repo.
# Includes skills evals, benchmarks, and schema validation.
#
# Usage:
#   bash evals/run.sh [--json] [--only benchmarks|skills|schema]

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
EVALS_DIR="$BRAIN_DIR/evals"
JSON_OUTPUT=0
ONLY=""
RUN_ID="$(date -u '+%Y%m%dT%H%M%SZ')"
LOGS_DIR="$BRAIN_DIR/logs/evals"
mkdir -p "$LOGS_DIR"

for arg in "$@"; do
  case "$arg" in
    --json)  JSON_OUTPUT=1 ;;
    --only)  shift; ONLY="${1:-}" ;;
  esac
done

TOTAL_PASS=0; TOTAL_FAIL=0
declare -a ALL_RESULTS=()

run_eval() {
  local script="$1" label="$(basename "$1" .sh)"
  if [ ! -x "$script" ]; then return; fi

  local json_flag=""
  [ "$JSON_OUTPUT" -eq 1 ] && json_flag="--json"

  if output=$("$script" $json_flag 2>&1); then
    TOTAL_PASS=$((TOTAL_PASS + 1))
    ALL_RESULTS+=("{\"name\":\"$label\",\"status\":\"pass\"}")
    [ "$JSON_OUTPUT" -eq 0 ] && echo "  PASS $label"
    [ "$JSON_OUTPUT" -eq 0 ] && [ -n "$output" ] && echo "$output" | sed 's/^/    /'
  else
    TOTAL_FAIL=$((TOTAL_FAIL + 1))
    ALL_RESULTS+=("{\"name\":\"$label\",\"status\":\"fail\"}")
    [ "$JSON_OUTPUT" -eq 0 ] && echo "  FAIL $label"
    [ "$JSON_OUTPUT" -eq 0 ] && [ -n "$output" ] && echo "$output" | sed 's/^/    /'
  fi
}

[ "$JSON_OUTPUT" -eq 0 ] && echo "" && echo "Brain Eval Suite - $RUN_ID"

# Skills evals
if [ -z "$ONLY" ] || [ "$ONLY" = "skills" ]; then
  [ "$JSON_OUTPUT" -eq 0 ] && echo "" && echo "-- Skills"
  for f in "$EVALS_DIR/skills/"*.sh; do
    [ -f "$f" ] && run_eval "$f"
  done
fi

# Benchmarks
if [ -z "$ONLY" ] || [ "$ONLY" = "benchmarks" ]; then
  [ "$JSON_OUTPUT" -eq 0 ] && echo "" && echo "-- Benchmarks"
  for f in "$EVALS_DIR/benchmarks/"*.sh; do
    [ -f "$f" ] && run_eval "$f"
  done
fi

# Schema validation
if [ -z "$ONLY" ] || [ "$ONLY" = "schema" ]; then
  [ "$JSON_OUTPUT" -eq 0 ] && echo "" && echo "-- Schema"
  if [ -f "$BRAIN_DIR/scripts/validate-schema.py" ]; then
    if python3 "$BRAIN_DIR/scripts/validate-schema.py" >/dev/null 2>&1; then
      TOTAL_PASS=$((TOTAL_PASS + 1))
      ALL_RESULTS+=("{\"name\":\"canonical-schema\",\"status\":\"pass\"}")
      [ "$JSON_OUTPUT" -eq 0 ] && echo "  PASS canonical-schema"
    else
      TOTAL_FAIL=$((TOTAL_FAIL + 1))
      ALL_RESULTS+=("{\"name\":\"canonical-schema\",\"status\":\"fail\"}")
      [ "$JSON_OUTPUT" -eq 0 ] && echo "  FAIL canonical-schema"
    fi
  fi
fi

TOTAL=$((TOTAL_PASS + TOTAL_FAIL))
STATUS="PASS"
[ "$TOTAL_FAIL" -gt 0 ] && STATUS="FAIL"

# Write to log
RESULT_JSON="{\"run_id\":\"$RUN_ID\",\"total\":$TOTAL,\"passed\":$TOTAL_PASS,\"failed\":$TOTAL_FAIL,\"status\":\"$STATUS\"}"
echo "$RESULT_JSON" > "$LOGS_DIR/${RUN_ID}.json"

if [ "$JSON_OUTPUT" -eq 1 ]; then
  IFS=","
  echo "{\"run_id\":\"$RUN_ID\",\"total\":$TOTAL,\"passed\":$TOTAL_PASS,\"failed\":$TOTAL_FAIL,\"status\":\"$STATUS\",\"results\":[${ALL_RESULTS[*]}]}"
else
  echo ""
  echo "Results: $TOTAL_PASS/$TOTAL passed | Status: $STATUS"
  echo "Log: $LOGS_DIR/${RUN_ID}.json"
fi

[ "$STATUS" = "PASS" ]
