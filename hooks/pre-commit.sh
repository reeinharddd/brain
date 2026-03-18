#!/usr/bin/env bash

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
GUARDIAN_RUNNER="$BRAIN_DIR/scripts/guardian.sh"

echo "Running Pre-commit Guardian..."

if [ ! -x "$GUARDIAN_RUNNER" ]; then
  echo "ERROR: Guardian runner not found at $GUARDIAN_RUNNER"
  exit 1
fi

bash "$GUARDIAN_RUNNER" --staged --threshold critical

if git diff --cached --name-only | grep -q "^rules/"; then
  echo "Rules changed. Re-generating adapters..."
  ./adapters/generate.sh
  git add adapters/ rules/compiled/
fi

echo "Guardian Passed."
exit 0
