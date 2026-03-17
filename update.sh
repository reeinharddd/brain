#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/update.sh — Pull latest changes and re-apply
#  Usage: ~/.brain/scripts/update.sh
# ═══════════════════════════════════════════════════════════

set -euo pipefail
BRAIN_DIR="$HOME/.brain"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; RESET='\033[0m'

echo -e "\n${BOLD}── Updating brain repo${RESET}"

# Check for uncommitted changes
if [ -d "$BRAIN_DIR/.git" ]; then
  if ! git -C "$BRAIN_DIR" diff --quiet 2>/dev/null; then
    echo -e "  ${YELLOW}⚠${RESET}  Uncommitted changes detected. Stashing..."
    git -C "$BRAIN_DIR" stash push -m "update.sh auto-stash $(date +%Y%m%d-%H%M%S)"
    STASHED=true
  else
    STASHED=false
  fi

  # Pull latest
  echo "  → Pulling latest from origin..."
  if git -C "$BRAIN_DIR" remote get-url origin &>/dev/null; then
    git -C "$BRAIN_DIR" pull --rebase origin "$(git -C "$BRAIN_DIR" branch --show-current)" && echo -e "  ${GREEN}✓${RESET} Pulled latest changes"
  else
    echo -e "  ${YELLOW}⚠${RESET}  No remote origin set — skipping pull (local update only)"
  fi

  # Restore stash
  if [ "${STASHED:-false}" = "true" ]; then
    git -C "$BRAIN_DIR" stash pop && echo -e "  ${GREEN}✓${RESET} Stash restored"
  fi
fi

# Regenerate adapters
echo "  → Regenerating rule adapters..."
bash "$BRAIN_DIR/adapters/generate.sh"

# Re-run links (idempotent)
echo "  → Re-applying symlinks..."
bash "$BRAIN_DIR/install.sh" 2>/dev/null | grep -E "(✓|✗|⚠)" || true

echo ""
echo -e "  ${GREEN}${BOLD}✓ Brain repo updated${RESET}"
echo "  Run ~/.brain/scripts/doctor.sh to verify"
echo ""
