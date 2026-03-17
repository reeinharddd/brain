#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/doctor.sh — Full diagnostic of the brain repo chain
#  Usage: ~/.brain/scripts/doctor.sh [--fix]
#  Exit codes: 0 = all good, 1 = failures found
# ═══════════════════════════════════════════════════════════

BRAIN_DIR="$HOME/.brain"
PASS=0; WARN=0; FAIL=0

# ── Colors ───────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; RESET='\033[0m'

section() { echo -e "\n${BOLD}── $1${RESET}"; }

check() {
  local label="$1"; local cmd="$2"
  if eval "$cmd" &>/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${RESET} $label"; ((PASS++))
  else
    echo -e "  ${RED}✗${RESET} $label"; ((FAIL++))
  fi
}

opt() {
  local label="$1"; local cmd="$2"
  if eval "$cmd" &>/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${RESET} $label"; ((PASS++))
  else
    echo -e "  ${YELLOW}⚠${RESET}  $label ${YELLOW}(optional)${RESET}"; ((WARN++))
  fi
}

# ── Core tools ───────────────────────────────────────────────
section "Core tools"
check "git"           "command -v git"
check "node >= 18"    "node -e 'process.exit(parseInt(process.versions.node) >= 18 ? 0 : 1)'"
opt   "go"            "command -v go"
opt   "uv (python)"   "command -v uv"
opt   "docker"        "command -v docker"

# ── Brain repo integrity ──────────────────────────────────────
section "Brain repo integrity"
check "~/.brain exists"         "test -d $BRAIN_DIR"
check "rules/canonical.md"      "test -f $BRAIN_DIR/rules/canonical.md"
check "adapters/generate.sh"    "test -f $BRAIN_DIR/adapters/generate.sh"
check "install.sh"              "test -f $BRAIN_DIR/install.sh"
check "git repo initialized"    "test -d $BRAIN_DIR/.git"
opt   "remote origin set"       "git -C $BRAIN_DIR remote get-url origin"

# ── Rule modules ─────────────────────────────────────────────
section "Rule modules"
for module in communication code-style git security workflow; do
  check "modules/$module.md" "test -f $BRAIN_DIR/rules/modules/$module.md"
done

# ── Generated adapters ───────────────────────────────────────
section "Generated adapters"
check "claude-code/CLAUDE.md"        "test -f $BRAIN_DIR/adapters/claude-code/CLAUDE.md"
check "cursor/.cursorrules"          "test -f $BRAIN_DIR/adapters/cursor/.cursorrules"
check "windsurf/.windsurfrules"      "test -f $BRAIN_DIR/adapters/windsurf/.windsurfrules"
check "gemini/GEMINI.md"             "test -f $BRAIN_DIR/adapters/gemini/GEMINI.md"
opt   "opencode/opencode.json"       "test -f $BRAIN_DIR/adapters/opencode/opencode.json"
opt   "aider/.aider.conf.yml"        "test -f $BRAIN_DIR/adapters/aider/.aider.conf.yml"
opt   "cline instructions"           "test -f $BRAIN_DIR/adapters/cline/cline_custom_instructions.md"

# ── Symlinks ─────────────────────────────────────────────────
section "Active symlinks"
check "CLAUDE.md → ~/.claude/CLAUDE.md"           "test -L $HOME/.claude/CLAUDE.md"
opt   "settings.json → ~/.claude/settings.json"   "test -L $HOME/.claude/settings.json"
check ".cursorrules → ~/.cursorrules"             "test -L $HOME/.cursorrules"
check ".windsurfrules → ~/.windsurfrules"         "test -L $HOME/.windsurfrules"
check "GEMINI.md → ~/.gemini/GEMINI.md"           "test -L $HOME/.gemini/GEMINI.md"
opt   "aider → ~/.aider.conf.yml"                 "test -L $HOME/.aider.conf.yml"
opt   "opencode → ~/.config/opencode/"            "test -L $HOME/.config/opencode/opencode.json"

# ── AI agents ────────────────────────────────────────────────
section "AI agents (at least one required)"
opt "claude (Claude Code)"     "command -v claude"
opt "opencode"                 "command -v opencode"
opt "aider"                    "command -v aider"
opt "gemini (Gemini CLI)"      "command -v gemini"
opt "cursor"                   "command -v cursor"
opt "cline (VS Code ext)"      "code --list-extensions 2>/dev/null | grep -q 'saoudrizwan.claude-dev'"

# ── Agents defined ───────────────────────────────────────────
section "Global agents defined"
for agent in orchestrator researcher planner designer reviewer debugger refactor documenter guardian; do
  opt "agents/$agent.md" "test -f $BRAIN_DIR/agents/$agent.md"
done

# ── Commands defined ─────────────────────────────────────────
section "Global commands defined"
for cmd in plan review research handover update-brain standup; do
  opt "commands/$cmd.md" "test -f $BRAIN_DIR/commands/$cmd.md"
done

# ── MCP config ───────────────────────────────────────────────
section "MCP configuration"
opt "registry.yml"            "test -f $BRAIN_DIR/mcp/registry.yml"
opt "profiles/minimal.json"   "test -f $BRAIN_DIR/mcp/profiles/minimal.json"
opt "profiles/standard.json"  "test -f $BRAIN_DIR/mcp/profiles/standard.json"
opt "settings.json references MCP" "grep -q 'mcpServers' $HOME/.claude/settings.json 2>/dev/null"

# ── Memory ───────────────────────────────────────────────────
section "Memory"
opt "memory/manifest.json"    "test -f $BRAIN_DIR/memory/manifest.json"
opt "memory/chunks/ exists"   "test -d $BRAIN_DIR/memory/chunks"
opt "engram installed"        "command -v engram"

# ── Providers ────────────────────────────────────────────────
section "Providers"
opt "providers/providers.yml" "test -f $BRAIN_DIR/providers/providers.yml"

# ── Hooks ────────────────────────────────────────────────────
section "Hooks"
opt "pre-tool-use/block-env-writes.sh"   "test -f $BRAIN_DIR/hooks/pre-tool-use/block-env-writes.sh"
opt "post-tool-use/run-linter.sh"        "test -f $BRAIN_DIR/hooks/post-tool-use/run-linter.sh"

# ── Summary ──────────────────────────────────────────────────
echo ""
echo "────────────────────────────────────────────"
if [ $FAIL -gt 0 ]; then
  echo -e "  ${RED}${BOLD}✗ Result: $PASS passed · $WARN warnings · $FAIL failed${RESET}"
  echo ""
  echo "  To fix failures:"
  echo "    bash ~/.brain/install.sh"
  echo ""
  exit 1
elif [ $WARN -gt 0 ]; then
  echo -e "  ${YELLOW}${BOLD}⚠ Result: $PASS passed · $WARN warnings · 0 failed${RESET}"
  echo ""
  echo "  Warnings are optional tools/integrations not yet set up."
  echo "  Add them when needed. Run install.sh if you just installed something."
  exit 0
else
  echo -e "  ${GREEN}${BOLD}✓ Result: $PASS passed · 0 warnings · 0 failed${RESET}"
  echo ""
  echo "  Brain repo is fully operational."
  exit 0
fi
