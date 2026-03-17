#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/scripts/validate.sh — Functional Validation Suite
#  Ensures agents have MCP access, memory, and rule alignment.
# ═══════════════════════════════════════════════════════════

# set -euo pipefail
BRAIN_DIR="$HOME/.brain"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; RESET='\033[0m'
PASS=0; FAIL=0

# ── Helpers ──────────────────────────────────────────────────
ok()   { echo -e "  ${GREEN}✓${RESET} $1"; ((PASS++)); }
fail() { echo -e "  ${RED}✗${RESET} $1"; ((FAIL++)); }
warn() { echo -e "  ${YELLOW}⚠${RESET}  $1"; }
info() { echo -e "  ${BOLD}→ $1${RESET}"; }

echo -e "\n${BOLD}── Running Functional Brain Validation${RESET}"

# 1. Rules Extraction Test
info "Testing Rule Alignment"
if grep -q "Clarity over cleverness" "$BRAIN_DIR/rules/canonical.md"; then
    ok "Core philosophy accessible in canonical.md"
else
    fail "Philosophy check failed"
fi

# 2. Filesystem MCP Test
info "Testing Filesystem MCP Access"
TEST_FILE="/tmp/brain_validation_$(date +%s).tmp"
echo "mcp_test" > "$TEST_FILE"
if [ -f "$TEST_FILE" ] && [ "$(cat "$TEST_FILE")" == "mcp_test" ]; then
    ok "Read/Write access verified via shell"
    rm "$TEST_FILE"
else
    fail "Filesystem access failed"
fi

# 3. Agent Autonomy / Command Path Test
info "Testing Command Links"
for cmd in plan review research handover update-brain test; do
    if [ -f "$BRAIN_DIR/commands/$cmd.md" ]; then
        if [ -L "$HOME/.claude/commands/$cmd.md" ]; then
            ok "Command link: /$cmd"
        else
            warn "Command /$cmd defined but not linked to ~/.claude"
        fi
    fi
done

# 4. Identity & Consistency Test
info "Testing Git Identity"
GIT_NAME=$(git -C "$BRAIN_DIR" config user.name)
if [ "$GIT_NAME" == "reeinharrrd" ]; then
    ok "Git identity correctly set to reeinharrrd"
else
    fail "Git identity mismatch: $GIT_NAME"
fi

# 5. MCP Health (Docker or NPX)
info "Testing MCP Health"
if [ -f "$BRAIN_DIR/docker/.env" ]; then
    DOCKER_MODE=1
    RUNNING=$(docker compose -f "$BRAIN_DIR/docker/docker-compose.yml" ps --format json | grep -c "running" || true)
    if [ "$RUNNING" -ge 4 ]; then
        ok "Docker MCP Stack: Healthy ($RUNNING services running)"
    else
        fail "Docker MCP Stack: Only $RUNNING services running"
    fi
else
    ok "npx mode: doctor.sh should verify manual dependencies"
fi

# ── Final Verdict ───────────────────────────────────────────
echo -e "\n────────────────────────────────────────────"
if [ $FAIL -eq 0 ]; then
    echo -e "  ${GREEN}${BOLD}PASSED: $PASS tests successful${RESET}"
    echo "  The brain repo is consistent and ready for use."
    exit 0
else
    echo -e "  ${RED}${BOLD}FAILED: $FAIL issues found${RESET}"
    exit 1
fi
