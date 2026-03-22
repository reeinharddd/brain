#!/bin/bash
# pre-tool-use/inject-global-rules.sh
# Validator: Ensures adapter files are up-to-date with canonical rules.

BRAIN_DIR="$HOME/.brain"
CANONICAL="$BRAIN_DIR/rules/canonical.md"
ADAPTERS=("$BRAIN_DIR/CLAUDE.md" "$BRAIN_DIR/.cursorrules" "$BRAIN_DIR/.windsurfrules")

# Check if adapters are fresh
for adapter in "${ADAPTERS[@]}"; do
  if [ -f "$adapter" ]; then
    if [ "$CANONICAL" -nt "$adapter" ]; then
      echo "[WARN] Adapter $(basename "$adapter") is outdated vs canonical.md."
      echo "      Run 'bash ~/.brain/adapters/generate.sh' to update."
    fi
  else
    echo "[WARN] Adapter $(basename "$adapter") is missing."
  fi
done

# The mechanism of universal rules is now via adapter files (CLAUDE.md, etc.)
# and NOT via environment variables, to ensure consistency across IDEs.
