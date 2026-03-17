#!/bin/bash
set -euo pipefail

# ═══════════════════════════════════════════════════════════
#  brain/install.sh — OS-aware bootstrap for the brain repo
#  Supports: Linux · macOS · WSL
#  Usage: bash ~/.brain/install.sh
# ═══════════════════════════════════════════════════════════

BRAIN_DIR="$HOME/.brain"
OS="unknown"
PACKAGE_MANAGER="unknown"
ERRORS=0

# ── Colors ───────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; RESET='\033[0m'

ok()   { echo -e "  ${GREEN}✓${RESET} $1"; }
warn() { echo -e "  ${YELLOW}⚠${RESET}  $1"; }
fail() { echo -e "  ${RED}✗${RESET} $1"; ((ERRORS++)); }
info() { echo -e "  ${BLUE}→${RESET} $1"; }
section() { echo -e "\n${BOLD}── $1${RESET}"; }

# ── OS Detection ─────────────────────────────────────────────
detect_os() {
  section "Detecting environment"
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if grep -q Microsoft /proc/version 2>/dev/null; then
      OS="wsl"; info "Windows Subsystem for Linux (WSL)"
    else
      OS="linux"; info "Linux"
    fi
    # Detect package manager
    if command -v apt &>/dev/null;    then PACKAGE_MANAGER="apt"
    elif command -v dnf &>/dev/null;  then PACKAGE_MANAGER="dnf"
    elif command -v pacman &>/dev/null; then PACKAGE_MANAGER="pacman"
    elif command -v zypper &>/dev/null; then PACKAGE_MANAGER="zypper"
    fi
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"; PACKAGE_MANAGER="brew"; info "macOS"
  else
    warn "Unknown OS: $OSTYPE — proceeding with best effort"
    OS="unknown"
  fi
  ok "OS: $OS | Package manager: $PACKAGE_MANAGER"
}

# ── Generate adapters ─────────────────────────────────────────
run_generate() {
  section "Generating rule adapters"
  if [ -f "$BRAIN_DIR/adapters/generate.sh" ]; then
    bash "$BRAIN_DIR/adapters/generate.sh"
    ok "All adapters generated"
  else
    fail "adapters/generate.sh not found — skipping adapter generation"
  fi
}

# ── Common symlinks (all OS) ──────────────────────────────────
link_common() {
  section "Linking common files"

  # Claude Code directories
  mkdir -p "$HOME/.claude/agents" "$HOME/.claude/commands"

  # Claude Code settings
  if [ -f "$BRAIN_DIR/adapters/claude-code/settings.json" ]; then
    ln -sf "$BRAIN_DIR/adapters/claude-code/settings.json" "$HOME/.claude/settings.json"
    ok "Claude Code settings.json linked"
  else
    warn "claude-code/settings.json not found — skipping"
  fi

  # Claude Code CLAUDE.md
  if [ -f "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" ]; then
    ln -sf "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" "$HOME/.claude/CLAUDE.md"
    ok "CLAUDE.md linked"
  else
    warn "claude-code/CLAUDE.md not found — run adapters/generate.sh"
  fi

  # Agents
  if [ -d "$BRAIN_DIR/agents" ]; then
    for agent in "$BRAIN_DIR/agents"/*.md; do
      [ -f "$agent" ] || continue
      ln -sf "$agent" "$HOME/.claude/agents/$(basename "$agent")"
    done
    ok "Agents linked → ~/.claude/agents/"
  fi

  # Commands
  if [ -d "$BRAIN_DIR/commands" ]; then
    for cmd in "$BRAIN_DIR/commands"/*.md; do
      [ -f "$cmd" ] || continue
      ln -sf "$cmd" "$HOME/.claude/commands/$(basename "$cmd")"
    done
    ok "Commands linked → ~/.claude/commands/"
  fi

  # Aider (universal)
  if [ -f "$BRAIN_DIR/adapters/aider/.aider.conf.yml" ]; then
    ln -sf "$BRAIN_DIR/adapters/aider/.aider.conf.yml" "$HOME/.aider.conf.yml"
    ok "Aider config linked"
  fi
}

# ── OS-specific symlinks ──────────────────────────────────────
link_os_specific() {
  section "Linking OS-specific files ($OS)"

  # Cursor (.cursorrules goes to HOME on all platforms)
  if [ -f "$BRAIN_DIR/adapters/cursor/.cursorrules" ]; then
    ln -sf "$BRAIN_DIR/adapters/cursor/.cursorrules" "$HOME/.cursorrules"
    ok ".cursorrules linked"
  fi

  # Windsurf
  if [ -f "$BRAIN_DIR/adapters/windsurf/.windsurfrules" ]; then
    ln -sf "$BRAIN_DIR/adapters/windsurf/.windsurfrules" "$HOME/.windsurfrules"
    ok ".windsurfrules linked"
  fi

  # Gemini CLI
  if [ -f "$BRAIN_DIR/adapters/gemini/GEMINI.md" ]; then
    mkdir -p "$HOME/.gemini"
    ln -sf "$BRAIN_DIR/adapters/gemini/GEMINI.md" "$HOME/.gemini/GEMINI.md"
    ok "GEMINI.md linked"
  fi

  case $OS in
    linux|wsl)
      # OpenCode
      if [ -f "$BRAIN_DIR/adapters/opencode/opencode.json" ]; then
        mkdir -p "$HOME/.config/opencode"
        ln -sf "$BRAIN_DIR/adapters/opencode/opencode.json" "$HOME/.config/opencode/opencode.json"
        ok "OpenCode config linked"
      fi
      ;;
    macos)
      # OpenCode on macOS uses ~/Library/Application Support/opencode
      if [ -f "$BRAIN_DIR/adapters/opencode/opencode.json" ]; then
        mkdir -p "$HOME/Library/Application Support/opencode"
        ln -sf "$BRAIN_DIR/adapters/opencode/opencode.json" "$HOME/Library/Application Support/opencode/opencode.json"
        ok "OpenCode config linked (macOS)"
      fi
      ;;
  esac
}

# ── Tool check ────────────────────────────────────────────────
check_tools() {
  section "Checking dependencies"
  command -v git  &>/dev/null && ok "git" || warn "git not found — install it"
  command -v node &>/dev/null && ok "node" || warn "node not found (needed for npm-based MCPs)"

  # AI agents (at least one should exist)
  local agents_found=0
  for agent_cmd in claude opencode aider gemini; do
    command -v "$agent_cmd" &>/dev/null && { ok "AI agent: $agent_cmd"; ((agents_found++)); }
  done
  [ $agents_found -eq 0 ] && warn "No AI agent found in PATH — install at least one (claude, opencode, aider, gemini)"
}

# ── Git init ──────────────────────────────────────────────────
init_git() {
  section "Initializing git"
  if [ ! -d "$BRAIN_DIR/.git" ]; then
    git -C "$BRAIN_DIR" init -q
    git -C "$BRAIN_DIR" add -A
    git -C "$BRAIN_DIR" commit -q -m "brain: initial setup"
    ok "Git repo initialized and first commit made"
  else
    ok "Git repo already initialized"
  fi
}

# ── Summary ───────────────────────────────────────────────────
print_summary() {
  section "Result"
  if [ $ERRORS -eq 0 ]; then
    echo -e "\n  ${GREEN}${BOLD}✓ Brain repo installed on $OS${RESET}"
  else
    echo -e "\n  ${YELLOW}${BOLD}⚠ Brain repo installed with $ERRORS issue(s)${RESET}"
    echo "  Run ~/.brain/scripts/doctor.sh for details"
  fi

  echo ""
  echo "  Active adapters:"
  [ -L "$HOME/.claude/CLAUDE.md" ]   && echo "    ✓ Claude Code" || echo "    - Claude Code (not linked)"
  [ -L "$HOME/.cursorrules" ]         && echo "    ✓ Cursor" || echo "    - Cursor (not linked)"
  [ -L "$HOME/.windsurfrules" ]       && echo "    ✓ Windsurf" || echo "    - Windsurf (not linked)"
  [ -L "$HOME/.gemini/GEMINI.md" ]    && echo "    ✓ Gemini CLI" || echo "    - Gemini CLI (not linked)"
  [ -L "$HOME/.aider.conf.yml" ]      && echo "    ✓ Aider" || echo "    - Aider (not linked)"

  echo ""
  echo "  Next steps:"
  echo "    1. Push to GitHub: cd ~/.brain && git remote add origin git@github.com:reeinharrrd/brain.git && git push -u origin main"
  echo "    2. Run doctor:     ~/.brain/scripts/doctor.sh"
  echo "    3. Update rules:   edit ~/.brain/rules/canonical.md → run ~/.brain/adapters/generate.sh"
  echo ""
}

# ── Main ──────────────────────────────────────────────────────
detect_os
run_generate
link_common
link_os_specific
check_tools
init_git
print_summary
