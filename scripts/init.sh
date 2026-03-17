#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/scripts/init.sh — Initialize local project with Brain rules
#  Usage: bash ~/.brain/scripts/init.sh
# ═══════════════════════════════════════════════════════════

set -euo pipefail
BRAIN_DIR="$HOME/.brain"

GREEN='\033[0;32m'; BOLD='\033[1m'; RESET='\033[0m'

echo -e "\n${BOLD}── Initializing Project with Brain Rules${RESET}"

# 1. CursorRules
if [ -f "$BRAIN_DIR/adapters/cursor/.cursorrules" ]; then
    ln -sf "$BRAIN_DIR/adapters/cursor/.cursorrules" "./.cursorrules"
    echo "  ✓ .cursorrules linked"
fi

# 2. WindsurfRules
if [ -f "$BRAIN_DIR/adapters/windsurf/.windsurfrules" ]; then
    ln -sf "$BRAIN_DIR/adapters/windsurf/.windsurfrules" "./.windsurfrules"
    echo "  ✓ .windsurfrules linked"
fi

# 3. GitHub Copilot
mkdir -p .github
if [ -f "$BRAIN_DIR/adapters/copilot/copilot-instructions.md" ]; then
    ln -sf "$BRAIN_DIR/adapters/copilot/copilot-instructions.md" "./.github/copilot-instructions.md"
    echo "  ✓ .github/copilot-instructions.md linked"
fi

# 4. CLAUDE.md (Local project context)
if [ -f "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" ]; then
    ln -sf "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" "./CLAUDE.md"
    echo "  ✓ CLAUDE.md linked"
fi

echo -e "\n  ${GREEN}${BOLD}✓ Project successfully connected to Brain Repo${RESET}"
