#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/scripts/update.sh — Pull latest and regenerate
#  Usage: bash ~/.brain/scripts/update.sh
# ═══════════════════════════════════════════════════════════

set -euo pipefail
BRAIN_DIR="$HOME/.brain"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; RESET='\033[0m'

echo -e "\n${BOLD}── Updating Brain Repo${RESET}"

# 1. Pull latest
if [ -d "$BRAIN_DIR/.git" ]; then
    echo "  Pulling latest changes..."
    git -C "$BRAIN_DIR" stash || true
    git -C "$BRAIN_DIR" pull --rebase || echo "  No remote configured yet."
    git -C "$BRAIN_DIR" stash pop || true
fi

# 2. Regenerate adapters
if [ -f "$BRAIN_DIR/adapters/generate.sh" ]; then
    bash "$BRAIN_DIR/adapters/generate.sh"
fi

# 3. Re-apply symlinks via install.sh
if [ -f "$BRAIN_DIR/scripts/install.sh" ]; then
    bash "$BRAIN_DIR/scripts/install.sh"
fi

echo -e "\n  ${GREEN}${BOLD}✓ Update complete${RESET}"
