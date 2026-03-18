#!/bin/bash

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

printf '{"dependencies":{"react":"18.0.0","next":"15.0.0"}}' > "$TMPDIR/package.json"
printf '{"compilerOptions":{"strict":true}}' > "$TMPDIR/tsconfig.json"
mkdir -p "$TMPDIR/app"

OUTPUT="$(bash "$BRAIN_DIR/scripts/render-skill-context.sh" "$TMPDIR")"
echo "$OUTPUT" | grep -q 'Next.js App'
echo "$OUTPUT" | grep -q 'React UI'
