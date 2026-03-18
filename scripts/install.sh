#!/bin/bash
# set -euo pipefail  <- removed for robustness in multi-platform environments

# ═══════════════════════════════════════════════════════════
#  brain/scripts/install.sh — OS-aware bootstrap for the brain repo
#  Supports: Linux · macOS · WSL
#  Usage: bash ~/.brain/scripts/install.sh
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

  # Docker Environment
  if [ ! -f "$BRAIN_DIR/docker/.env" ] && [ -f "$BRAIN_DIR/docker/.env.example" ]; then
    cp "$BRAIN_DIR/docker/.env.example" "$BRAIN_DIR/docker/.env"
    sed -i "s|HOST_HOME=.*|HOST_HOME=$HOME|g" "$BRAIN_DIR/docker/.env"
    ok "Docker .env initialized from example"
  fi

  # Claude Code directories
  mkdir -p "$HOME/.claude/agents" "$HOME/.claude/commands"

  # Claude Code settings
  # Default to Docker Persistent for a warm-start experience
  if [ -f "$BRAIN_DIR/adapters/claude-code/settings.persistent.json" ]; then
    ln -sf "$BRAIN_DIR/adapters/claude-code/settings.persistent.json" "$HOME/.claude/settings.json"
    ok "Claude Code: Persistent Mode activated"
  elif [ -f "$BRAIN_DIR/adapters/claude-code/settings.json" ]; then
    # Keep existing if linked, otherwise default to settings.json
    if [ ! -L "$HOME/.claude/settings.json" ]; then
        ln -sf "$BRAIN_DIR/adapters/claude-code/settings.json" "$HOME/.claude/settings.json"
        ok "Claude Code settings.json linked"
    else
        ok "Claude Code settings.json already linked (currently: $(readlink "$HOME/.claude/settings.json"))"
    fi
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
  # Standard tools (individual checks)
  if command -v git &>/dev/null; then ok "git"; else warn "git not found"; fi
  if command -v node &>/dev/null; then ok "node"; else warn "node not found"; fi

  # AI agents (at least one should exist)
  local agents_found=0
  for agent_cmd in claude opencode aider gemini; do
    if command -v "$agent_cmd" &>/dev/null; then
      ok "AI agent: $agent_cmd"
      ((agents_found++))
    fi
  done

  if [ $agents_found -eq 0 ]; then
    warn "No AI agent found in PATH — install at least one (claude, opencode, aider, gemini)"
  fi

  # Docker Pre-pull for Persistent MCPs
  if command -v docker &>/dev/null; then
    section "Pre-pulling Docker MCPs (Bootstrap Speedup)"
    local images=(
      "node:20-alpine"
      "mcp/github"
      "mcp/duckduckgo"
      "mcp/sequentialthinking"
      "mcp/google-maps"
    )
    for img in "${images[@]}"; do
      if [[ "$(docker images -q "$img" 2>/dev/null)" == "" ]]; then
        info "Pulling $img..."
        docker pull -q "$img" &>/dev/null || warn "Failed to pull $img (skipping)"
      else
        ok "$img already present"
      fi
    done
  fi
}

# ── Git init ──────────────────────────────────────────────────
init_git() {
  section "Initializing git"
  if [ ! -d "$BRAIN_DIR/.git" ]; then
    git -C "$BRAIN_DIR" init -q
    git -C "$BRAIN_DIR" config user.name "reeinharrrd"
    git -C "$BRAIN_DIR" config user.email "reeinharrrd@users.noreply.github.com"
    git -C "$BRAIN_DIR" add -A
    git -C "$BRAIN_DIR" commit -q -m "brain: initial setup"
    ok "Git repo initialized"
  else
    ok "Git repo already initialized"
  fi
}

# ── Summary ───────────────────────────────────────────────────
print_summary() {
  section "Result"
  if [ $ERRORS -eq 0 ]; then
    echo -e "\n  ${GREEN}${BOLD}✓ Brain repo scripts installed on $OS${RESET}"
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
  [ -L "$HOME/.claude/commands/test.md" ] && echo "    ✓ Test Command" || echo "    - Test Command (not linked)"
  [ -L "$HOME/.claude/commands/init.md" ] && echo "    ✓ Init Command" || echo "    - Init Command (not linked)"

  echo ""
  echo "  Next steps:"
  echo "    1. Push to GitHub:     cd ~/.brain && git remote add origin git@github.com:reeinharrrd/brain.git && git push -u origin main"
  echo "    2. Run doctor:         ~/.brain/scripts/doctor.sh"
  echo "    3. Run validation:     ~/.brain/scripts/validate.sh"
  echo "    4. Update architecture: edit ~/.brain/rules/canonical.md → run ~/.brain/adapters/generate.sh"
  echo ""
}

# ── Main ──────────────────────────────────────────────────────
detect_os
run_generate
link_common
link_os_specific
# Centralize and sync MCP configs
if [ -f "$BRAIN_DIR/scripts/mcp-sync.sh" ]; then
  bash "$BRAIN_DIR/scripts/mcp-sync.sh"
fi
check_tools
init_git
print_summary

# Auto-run doctor
if [ -f "$BRAIN_DIR/scripts/doctor.sh" ]; then
  bash "$BRAIN_DIR/scripts/doctor.sh"
fi
