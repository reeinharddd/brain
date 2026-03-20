#!/bin/bash
# set -euo pipefail  <- removed for robustness in multi-platform environments

# ===========================================================
#  brain/scripts/install.sh - OS-aware bootstrap for the brain repo
#  Supports: Linux / macOS / WSL
#  Usage: bash ~/.brain/scripts/install.sh
# ===========================================================

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
OS="unknown"
PACKAGE_MANAGER="unknown"
ERRORS=0
DRY_RUN=0
INTERACTIVE=0
CLAUDE_MODE="persistent"

# -- Colors --------------------------------------------------------
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; RESET='\033[0m'

ok()   { echo -e "  ${GREEN}[ok]${RESET} $1"; }
warn() { echo -e "  ${YELLOW}[warn]${RESET}  $1"; }
fail() { echo -e "  ${RED}[fail]${RESET} $1"; ((ERRORS++)); }
info() { echo -e "  ${BLUE}[info]${RESET} $1"; }
section() { echo -e "\n${BOLD}-- $1${RESET}"; }

run_cmd() {
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

should_run_step() {
  local label="$1"
  if [ "$INTERACTIVE" -eq 0 ]; then
    return 0
  fi
  read -r -p "$label? [Y/n] " answer
  case "${answer:-y}" in
    n|N) return 1 ;;
    *) return 0 ;;
  esac
}

choose_claude_mode() {
  if [ "$INTERACTIVE" -eq 0 ]; then
    return
  fi

  echo "Select Claude Code mode:"
  echo "  1. Persistent"
  echo "  2. Standard"
  read -r -p "Choice [1/2]: " answer
  case "${answer:-1}" in
    2) CLAUDE_MODE="standard" ;;
    *) CLAUDE_MODE="persistent" ;;
  esac
}

# -- OS Detection -------------------------------------------------
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
    warn "Unknown OS: $OSTYPE - proceeding with best effort"
    OS="unknown"
  fi
  ok "OS: $OS | Package manager: $PACKAGE_MANAGER"
}

# -- Generate adapters ---------------------------------------------
run_generate() {
  section "Generating rule adapters"
  if [ -f "$BRAIN_DIR/adapters/generate.sh" ]; then
    run_cmd bash "$BRAIN_DIR/adapters/generate.sh"
    ok "All adapters generated"
  else
    fail "adapters/generate.sh not found - skipping adapter generation"
  fi
}

# -- Common symlinks (all OS) --------------------------------------
link_common() {
  section "Linking common files"

  # Docker Environment
  if [ ! -f "$BRAIN_DIR/docker/.env" ] && [ -f "$BRAIN_DIR/docker/.env.example" ]; then
    run_cmd cp "$BRAIN_DIR/docker/.env.example" "$BRAIN_DIR/docker/.env"
    if [ "$DRY_RUN" -eq 0 ]; then
      sed -i "s|HOST_HOME=.*|HOST_HOME=$HOME|g" "$BRAIN_DIR/docker/.env"
    else
      echo "[dry-run] sed -i s|HOST_HOME=.*|HOST_HOME=$HOME|g $BRAIN_DIR/docker/.env"
    fi
    ok "Docker .env initialized from example"
  fi

  # Claude Code directories
  run_cmd mkdir -p "$HOME/.claude/agents" "$HOME/.claude/commands"

  # Claude Code settings
  # Default to Docker Persistent for a warm-start experience
  if [ "$CLAUDE_MODE" = "persistent" ] && [ -f "$BRAIN_DIR/adapters/claude-code/settings.persistent.json" ]; then
    run_cmd ln -sf "$BRAIN_DIR/adapters/claude-code/settings.persistent.json" "$HOME/.claude/settings.json"
    ok "Claude Code: Persistent Mode activated"
  elif [ -f "$BRAIN_DIR/adapters/claude-code/settings.json" ]; then
    # Keep existing if linked, otherwise default to settings.json
    if [ ! -L "$HOME/.claude/settings.json" ]; then
        run_cmd ln -sf "$BRAIN_DIR/adapters/claude-code/settings.json" "$HOME/.claude/settings.json"
        ok "Claude Code settings.json linked"
    else
        ok "Claude Code settings.json already linked (currently: $(readlink "$HOME/.claude/settings.json"))"
    fi
  else
    warn "claude-code/settings.json not found - skipping"
  fi

  # Claude Code CLAUDE.md
  if [ -f "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" ]; then
    run_cmd ln -sf "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" "$HOME/.claude/CLAUDE.md"
    ok "CLAUDE.md linked"
  else
    warn "claude-code/CLAUDE.md not found - run adapters/generate.sh"
  fi

  # Agents
  if [ -d "$BRAIN_DIR/agents" ]; then
    for agent in "$BRAIN_DIR/agents"/*.md; do
      [ -f "$agent" ] || continue
      run_cmd ln -sf "$agent" "$HOME/.claude/agents/$(basename "$agent")"
    done
    ok "Agents linked -> ~/.claude/agents/"
  fi

  # Commands
  if [ -d "$BRAIN_DIR/commands" ]; then
    for cmd in "$BRAIN_DIR/commands"/*.md; do
      [ -f "$cmd" ] || continue
      run_cmd ln -sf "$cmd" "$HOME/.claude/commands/$(basename "$cmd")"
    done
    ok "Commands linked -> ~/.claude/commands/"
  fi

  # Aider (universal)
  if [ -f "$BRAIN_DIR/adapters/aider/.aider.conf.yml" ]; then
    run_cmd ln -sf "$BRAIN_DIR/adapters/aider/.aider.conf.yml" "$HOME/.aider.conf.yml"
    ok "Aider config linked"
  fi
}

# -- OS-specific symlinks ------------------------------------------
link_os_specific() {
  section "Linking OS-specific files ($OS)"

  # Cursor (.cursorrules goes to HOME on all platforms)
  if [ -f "$BRAIN_DIR/adapters/cursor/.cursorrules" ]; then
    run_cmd ln -sf "$BRAIN_DIR/adapters/cursor/.cursorrules" "$HOME/.cursorrules"
    ok ".cursorrules linked"
  fi

  # Windsurf
  if [ -f "$BRAIN_DIR/adapters/windsurf/.windsurfrules" ]; then
    run_cmd ln -sf "$BRAIN_DIR/adapters/windsurf/.windsurfrules" "$HOME/.windsurfrules"
    ok ".windsurfrules linked"
  fi

  # Gemini CLI
  if [ -f "$BRAIN_DIR/adapters/gemini/GEMINI.md" ]; then
    run_cmd mkdir -p "$HOME/.gemini"
    run_cmd ln -sf "$BRAIN_DIR/adapters/gemini/GEMINI.md" "$HOME/.gemini/GEMINI.md"
    ok "GEMINI.md linked"
  fi

  case $OS in
    linux|wsl)
      # OpenCode
      if [ -f "$BRAIN_DIR/adapters/opencode/opencode.json" ]; then
        run_cmd mkdir -p "$HOME/.config/opencode"
        run_cmd ln -sf "$BRAIN_DIR/adapters/opencode/opencode.json" "$HOME/.config/opencode/opencode.json"
        ok "OpenCode config linked"
      fi
      ;;
    macos)
      if [ -f "$BRAIN_DIR/adapters/opencode/opencode.json" ]; then
        run_cmd mkdir -p "$HOME/Library/Application Support/opencode"
        run_cmd ln -sf "$BRAIN_DIR/adapters/opencode/opencode.json" "$HOME/Library/Application Support/opencode/opencode.json"
        ok "OpenCode config linked (macOS)"
      fi
      ;;
  esac
}

# -- Tool check ----------------------------------------------------
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
    warn "No AI agent found in PATH - install at least one (claude, opencode, aider, gemini)"
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
        if [ "$DRY_RUN" -eq 0 ]; then
          docker pull -q "$img" &>/dev/null || warn "Failed to pull $img (skipping)"
        else
          echo "[dry-run] docker pull -q $img"
        fi
      else
        ok "$img already present"
      fi
    done
  fi
}

# -- Git init ------------------------------------------------------
init_git() {
  section "Initializing git"
  if [ ! -d "$BRAIN_DIR/.git" ]; then
    run_cmd git -C "$BRAIN_DIR" init -q
    run_cmd git -C "$BRAIN_DIR" config user.name "reeinharrrd"
    run_cmd git -C "$BRAIN_DIR" config user.email "reeinharrrd@users.noreply.github.com"
    run_cmd git -C "$BRAIN_DIR" add -A
    run_cmd git -C "$BRAIN_DIR" commit -q -m "brain: initial setup"
    ok "Git repo initialized"
  else
    ok "Git repo already initialized"
  fi
}

# -- ai-local ------------------------------------------------------
setup_ai_local() {
  section "Setting up ai-local (Ollama orchestrator)"
  if [ -d "$BRAIN_DIR/ai-local" ]; then
    ok "ai-local module present"
    info "To start local models, cd $BRAIN_DIR/ai-local && docker compose up -d"
  else
    warn "ai-local module not found"
  fi
}

# -- Summary -------------------------------------------------------
print_summary() {
  section "Result"
  if [ $ERRORS -eq 0 ]; then
    echo -e "\n  ${GREEN}${BOLD}[ok] Brain repo scripts installed on $OS${RESET}"
  else
    echo -e "\n  ${YELLOW}${BOLD}[warn] Brain repo installed with $ERRORS issue(s)${RESET}"
    echo "  Run ~/.brain/scripts/doctor.sh for details"
  fi

  echo ""
  echo "  Active adapters:"
  [ -L "$HOME/.claude/CLAUDE.md" ]   && echo "    [ok] Claude Code" || echo "    - Claude Code (not linked)"
  [ -L "$HOME/.cursorrules" ]         && echo "    [ok] Cursor" || echo "    - Cursor (not linked)"
  [ -L "$HOME/.windsurfrules" ]       && echo "    [ok] Windsurf" || echo "    - Windsurf (not linked)"
  [ -L "$HOME/.gemini/GEMINI.md" ]    && echo "    [ok] Gemini CLI" || echo "    - Gemini CLI (not linked)"
  [ -L "$HOME/.aider.conf.yml" ]      && echo "    [ok] Aider" || echo "    - Aider (not linked)"
  [ -L "$HOME/.claude/commands/test.md" ] && echo "    [ok] Test Command" || echo "    - Test Command (not linked)"
  [ -L "$HOME/.claude/commands/init.md" ] && echo "    [ok] Init Command" || echo "    - Init Command (not linked)"

  echo ""
  echo "  Next steps:"
  echo "    1. Push to GitHub:     cd ~/.brain && git remote add origin git@github.com:reeinharrrd/brain.git && git push -u origin main"
  echo "    2. Run doctor:         ~/.brain/scripts/doctor.sh"
  echo "    3. Run validation:     ~/.brain/scripts/validate.sh"
  echo "    4. Update architecture: edit ~/.brain/rules/canonical.md -> run ~/.brain/adapters/generate.sh"
  echo ""
}

# -- Args ----------------------------------------------------------
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --interactive) INTERACTIVE=1 ;;
  esac
done

# -- Main ----------------------------------------------------------
detect_os
choose_claude_mode
should_run_step "Generate adapters" && run_generate
should_run_step "Link common files" && link_common
should_run_step "Link OS-specific files" && link_os_specific
# Centralize and sync MCP configs
if [ -f "$BRAIN_DIR/scripts/mcp-sync.sh" ]; then
  should_run_step "Sync MCP configs" && run_cmd bash "$BRAIN_DIR/scripts/mcp-sync.sh"
fi
should_run_step "Check tools" && check_tools
should_run_step "Initialize git" && init_git
should_run_step "Setup ai-local (Ollama)" && setup_ai_local
print_summary

# Auto-run doctor
if [ -f "$BRAIN_DIR/scripts/doctor.sh" ]; then
  if [ "$DRY_RUN" -eq 0 ]; then
    bash "$BRAIN_DIR/scripts/doctor.sh"
  else
    echo "[dry-run] bash $BRAIN_DIR/scripts/doctor.sh"
  fi
fi

if [ "$ERRORS" -eq 0 ]; then
  exit 0
fi
exit 1
