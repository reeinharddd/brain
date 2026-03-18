#!/bin/bash
# Run lightweight evals for brain skills and platform scripts.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
FAIL=0

for eval_script in "$BRAIN_DIR"/evals/skills/*.sh; do
  [ -x "$eval_script" ] || continue
  echo "Running $(basename "$eval_script")"
  if ! "$eval_script"; then
    FAIL=1
  fi
done

exit "$FAIL"
