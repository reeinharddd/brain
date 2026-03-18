#!/usr/bin/env bash
# guardian.sh - compatibility wrapper for local Guardian execution

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
GUARDIAN_RUNNER="$BRAIN_DIR/guardian/run.sh"
PROVIDER_ENV="$BRAIN_DIR/providers/guardian.env"
MODE="staged"
OUTPUT="text"
THRESHOLD="critical"
PASS_THROUGH=()
DRY_RUN=0
VERBOSE=0

for arg in "$@"; do
  case "$arg" in
    --dry-run)
      DRY_RUN=1
      ;;
    --verbose)
      VERBOSE=1
      ;;
    --staged)
      MODE="staged"
      PASS_THROUGH+=("$arg")
      ;;
    --diff-range)
      MODE="range"
      PASS_THROUGH+=("$arg")
      ;;
    --output)
      OUTPUT="custom"
      PASS_THROUGH+=("$arg")
      ;;
    --threshold)
      THRESHOLD="custom"
      PASS_THROUGH+=("$arg")
      ;;
    *)
      PASS_THROUGH+=("$arg")
      ;;
  esac
done

if [ -f "$PROVIDER_ENV" ]; then
  # shellcheck disable=SC1090
  . "$PROVIDER_ENV"
fi

if [ "$DRY_RUN" -eq 1 ]; then
  echo "Guardian dry run"
  echo "- runner: $GUARDIAN_RUNNER"
  echo "- mode: $MODE"
  echo "- provider env: ${PROVIDER_ENV}"
  exit 0
fi

if [ ! -x "$GUARDIAN_RUNNER" ]; then
  echo "ERROR: guardian runner not found at $GUARDIAN_RUNNER" >&2
  exit 1
fi

if [ "$VERBOSE" -eq 1 ]; then
  echo "Guardian wrapper invoking: $GUARDIAN_RUNNER ${PASS_THROUGH[*]}" >&2
fi

exec bash "$GUARDIAN_RUNNER" "${PASS_THROUGH[@]}"
