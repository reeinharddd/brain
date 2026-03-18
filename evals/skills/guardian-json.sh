#!/bin/bash

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

git -C "$TMPDIR" init -q
printf 'const value = 1;\n' > "$TMPDIR/index.ts"
git -C "$TMPDIR" add index.ts

OUTPUT="$(GUARDIAN_REPO_ROOT="$TMPDIR" bash "$BRAIN_DIR/guardian/run.sh" --staged --output json)"
echo "$OUTPUT" | grep -q '"verdict"' 
