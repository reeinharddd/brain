#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REQUESTED_BRAIN_DIR="${BRAIN_DIR:-}"

if [ -n "$REQUESTED_BRAIN_DIR" ] && [ -d "$REQUESTED_BRAIN_DIR" ]; then
  REQUESTED_BRAIN_DIR="$(cd "$REQUESTED_BRAIN_DIR" && pwd)"
fi

if [ -n "$REQUESTED_BRAIN_DIR" ] && [ "$REQUESTED_BRAIN_DIR" = "$REPO_ROOT" ]; then
  BRAIN_DIR="$REQUESTED_BRAIN_DIR"
else
  BRAIN_DIR="$REPO_ROOT"
fi

GUARDIAN_DIR="$BRAIN_DIR/guardian"
CHECKS_DIR="$GUARDIAN_DIR/checks"
OUTPUTS_DIR="$GUARDIAN_DIR/outputs"
GUARDIAN_LIB="$GUARDIAN_DIR/lib.sh"
GUARDIAN_REPO_ROOT="${GUARDIAN_REPO_ROOT:-$BRAIN_DIR}"
START_TS="$(date +%s%3N 2>/dev/null || date +%s000)"

MODE="staged"
OUTPUT_FORMAT="text"
THRESHOLD="critical"
GUARDIAN_DIFF_RANGE=""
AUTO_FALLBACK_TO_HEAD=1

usage() {
  cat <<'EOF'
Usage:
  bash ~/.brain/guardian/run.sh --staged
  bash ~/.brain/guardian/run.sh --diff-range origin/main...HEAD --pr-mode
Options:
  --staged
  --diff-range <range>
  --diff-only
  --pr-mode
  --threshold <critical|high|medium|low>
  --output <text|json>
  --no-fallback-head
EOF
}

severity_rank() {
  local sev="$(echo "$1" | tr '[:upper:]' '[:lower:]')"
  case "$sev" in
    critical) echo 4 ;;
    high) echo 3 ;;
    medium) echo 2 ;;
    low) echo 1 ;;
    *) echo 0 ;;
  esac
}

while [ $# -gt 0 ]; do
  case "$1" in
    --staged)
      MODE="staged"
      shift
      ;;
    --diff-range)
      MODE="range"
      GUARDIAN_DIFF_RANGE="${2:-}"
      shift 2
      ;;
    --diff-only|--pr-mode)
      shift
      ;;
    --threshold)
      THRESHOLD="${2:-critical}"
      shift 2
      ;;
    --output)
      OUTPUT_FORMAT="${2:-text}"
      shift 2
      ;;
    --no-fallback-head)
      AUTO_FALLBACK_TO_HEAD=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if ! git -C "$GUARDIAN_REPO_ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "ERROR: guardian must run inside a git repository" >&2
  exit 1
fi

if [ "$MODE" = "range" ] && [ -z "$GUARDIAN_DIFF_RANGE" ]; then
  echo "ERROR: --diff-range requires a git diff range" >&2
  exit 1
fi

FILES_FILE="$(mktemp)"
RESULTS_FILE="$(mktemp)"
trap 'rm -f "$FILES_FILE" "$RESULTS_FILE"' EXIT

if [ "$MODE" = "staged" ]; then
  git -C "$GUARDIAN_REPO_ROOT" diff --cached --name-only --diff-filter=ACMR > "$FILES_FILE"
  if [ ! -s "$FILES_FILE" ] && [ "$AUTO_FALLBACK_TO_HEAD" -eq 1 ]; then
    MODE="range"
    GUARDIAN_DIFF_RANGE="HEAD"
    git -C "$GUARDIAN_REPO_ROOT" diff --name-only --diff-filter=ACMR "$GUARDIAN_DIFF_RANGE" > "$FILES_FILE"
  fi
else
  git -C "$GUARDIAN_REPO_ROOT" diff --name-only --diff-filter=ACMR "$GUARDIAN_DIFF_RANGE" > "$FILES_FILE"
fi

if [ ! -s "$FILES_FILE" ]; then
  echo "Guardian verdict: PASS (no files to check)"
  exit 0
fi

export GUARDIAN_MODE="$MODE"
export GUARDIAN_DIFF_RANGE
export GUARDIAN_FILES_FILE="$FILES_FILE"
export GUARDIAN_LIB
export GUARDIAN_REPO_ROOT
export GUARDIAN_FALLBACK_TO_HEAD_USED="$([ "$MODE" = "range" ] && [ "${GUARDIAN_DIFF_RANGE:-}" = "HEAD" ] && printf '1' || printf '0')"

for check in "$CHECKS_DIR"/*.sh; do
  [ -x "$check" ] || continue
  "$check" >> "$RESULTS_FILE"
done

FAIL_COUNT=0
THRESHOLD_RANK="$(severity_rank "$THRESHOLD")"

if [ -s "$RESULTS_FILE" ]; then
  while IFS=$'\t' read -r severity _rest; do
    if [ "$(severity_rank "$severity")" -ge "$THRESHOLD_RANK" ]; then
      FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
  done < "$RESULTS_FILE"
fi

case "$OUTPUT_FORMAT" in
  text)
    "$OUTPUTS_DIR/text.sh" "$RESULTS_FILE" "$FAIL_COUNT"
    ;;
  json)
    "$OUTPUTS_DIR/json.sh" "$RESULTS_FILE" "$FAIL_COUNT"
    ;;
  *)
    echo "ERROR: unsupported output format: $OUTPUT_FORMAT" >&2
    exit 1
    ;;
esac

if [ "$FAIL_COUNT" -gt 0 ]; then
  if [ -x "$BRAIN_DIR/scripts/telemetry.sh" ]; then
    END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
    DURATION_MS="$((END_TS - START_TS))"
    bash "$BRAIN_DIR/scripts/telemetry.sh" record "guardian-run" "block" "$DURATION_MS" "$MODE" >/dev/null 2>&1 || true
  fi
  exit 1
fi

if [ -x "$BRAIN_DIR/scripts/telemetry.sh" ]; then
  END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
  DURATION_MS="$((END_TS - START_TS))"
  STATUS="ok"
  if [ -s "$RESULTS_FILE" ]; then
    STATUS="warn"
  fi
  bash "$BRAIN_DIR/scripts/telemetry.sh" record "guardian-run" "$STATUS" "$DURATION_MS" "$MODE" >/dev/null 2>&1 || true
fi
