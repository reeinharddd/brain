#!/bin/bash
# ===========================================================
#  brain/scripts/init.sh - Initialize local project with Brain rules
#  Usage: bash ~/.brain/scripts/init.sh
# ===========================================================

set -euo pipefail
BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"

GREEN='\033[0;32m'; BOLD='\033[1m'; RESET='\033[0m'

echo -e "\n${BOLD}-- Initializing Project with Brain Rules${RESET}"

# 1. CursorRules
if [ -f "$BRAIN_DIR/adapters/cursor/.cursorrules" ]; then
    ln -sf "$BRAIN_DIR/adapters/cursor/.cursorrules" "./.cursorrules"
    echo "  [ok] .cursorrules linked"
fi

# 2. WindsurfRules
if [ -f "$BRAIN_DIR/adapters/windsurf/.windsurfrules" ]; then
    ln -sf "$BRAIN_DIR/adapters/windsurf/.windsurfrules" "./.windsurfrules"
    echo "  [ok] .windsurfrules linked"
fi

# 3. GitHub Copilot
mkdir -p .github
if [ -f "$BRAIN_DIR/adapters/copilot/copilot-instructions.md" ]; then
    ln -sf "$BRAIN_DIR/adapters/copilot/copilot-instructions.md" "./.github/copilot-instructions.md"
    echo "  [ok] .github/copilot-instructions.md linked"
fi

# 4. CLAUDE.md (Local project context)
if [ -f "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" ]; then
    ln -sf "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" "./CLAUDE.md"
    echo "  [ok] CLAUDE.md linked"
fi

# 5. Dynamic stack-aware skill context
if [ -x "$BRAIN_DIR/scripts/render-skill-context.sh" ]; then
    CONTEXT_PATH="$(bash "$BRAIN_DIR/scripts/render-skill-context.sh" --write .)"
    echo "  [ok] Skill context generated at $CONTEXT_PATH"
fi

# 6. Git-native Guardian hook
if [ -d ".git/hooks" ]; then
    cat > ".git/hooks/pre-commit" <<HOOK
#!/usr/bin/env bash
set -euo pipefail
bash "$BRAIN_DIR/guardian/run.sh" --staged --threshold critical
HOOK
    chmod +x ".git/hooks/pre-commit"
    echo "  [ok] Git pre-commit hook installed"
fi

echo -e "\n  ${GREEN}${BOLD}[ok] Project successfully connected to Brain Repo${RESET}"
