#!/bin/bash
# setup-persistent.sh
# One-time setup that makes the brain repo fully automatic and persistent.
#
# What this does:
#   1. Sources brain.env automatically on every shell session (zsh/bash)
#   2. Links all IDE adapters (Claude Code, Cursor, Windsurf, Gemini, OpenCode, Aider)
#   3. Syncs MCP servers to all IDEs
#   4. Installs Guardian git hook globally (runs on every commit in any repo)
#   5. Installs cron jobs (daily validation, weekly memory consolidation, weekly evals)
#   6. Creates brain.env from example if it doesn't exist yet
#   7. Registers the brain-rules MCP server in Claude Code settings
#
# Run once after install. Safe to re-run (idempotent).
#
# Usage:
#   bash ~/.brain/scripts/setup-persistent.sh
#   bash ~/.brain/scripts/setup-persistent.sh --dry-run

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
DRY_RUN=0

for arg in "$@"; do
  [ "$arg" = "--dry-run" ] && DRY_RUN=1
done

G='\033[0;32m'; Y='\033[1;33m'; R='\033[0;31m'; B='\033[1m'; RS='\033[0m'
ok()      { echo -e "  ${G}[ok]${RS}   $1"; }
warn()    { echo -e "  ${Y}[warn]${RS} $1"; }
fail()    { echo -e "  ${R}[fail]${RS} $1"; }
info()    { echo -e "\n${B}-- $1${RS}"; }
skip()    { echo -e "  ${Y}[skip]${RS} $1 (dry-run)"; }
run_cmd() { [ "$DRY_RUN" -eq 1 ] && skip "$*" || eval "$@"; }

echo ""
echo -e "${B}Brain Repo - Persistent Setup${RS}"
[ "$DRY_RUN" -eq 1 ] && echo -e "${Y}DRY RUN - no changes will be made${RS}"
echo ""

# ---- 0. Sanity checks --------------------------------------------------------
if [ ! -d "$BRAIN_DIR/.git" ]; then
  fail "Brain dir not found or not a git repo: $BRAIN_DIR"
  fail "Clone first: git clone git@github.com:reeinharddd/brain.git ~/.brain"
  exit 1
fi

# ---- 1. Create brain.env from example ----------------------------------------
info "User configuration (brain.env)"

ENV_FILE="$BRAIN_DIR/brain.env"
ENV_EXAMPLE="$BRAIN_DIR/brain.env.example"

if [ ! -f "$ENV_FILE" ]; then
  if [ -f "$ENV_EXAMPLE" ]; then
    run_cmd cp "$ENV_EXAMPLE" "$ENV_FILE"
    ok "brain.env created from example"
    echo "     Edit it to set your API keys: $ENV_FILE"
  else
    warn "brain.env.example not found - run brain-upgrade-installer.sh first"
  fi
else
  ok "brain.env already exists"
fi

# Add brain.env to .gitignore if not already there
GITIGNORE="$BRAIN_DIR/.gitignore"
if [ -f "$GITIGNORE" ] && ! grep -q "brain.env$" "$GITIGNORE"; then
  run_cmd bash -c "echo 'brain.env' >> '$GITIGNORE'"
  ok ".gitignore: brain.env added"
fi

# ── 2. Shell auto-sourcing (zsh + bash) ───────────────────────────────────────
info "Shell integration (auto-source brain.env)"

SHELL_SNIPPET='
# brain repo - auto-source config
[ -f "$HOME/.brain/brain.env" ] && set -a && . "$HOME/.brain/brain.env" && set +a
export BRAIN_DIR="$HOME/.brain"
'

# Determine which shell files to patch
SHELL_FILES=()
[ -f "$HOME/.zshrc" ]             && SHELL_FILES+=("$HOME/.zshrc")
[ -f "$HOME/.bashrc" ]            && SHELL_FILES+=("$HOME/.bashrc")
[ -f "$HOME/.bash_profile" ]      && SHELL_FILES+=("$HOME/.bash_profile")
[ ${#SHELL_FILES[@]} -eq 0 ]      && SHELL_FILES+=("$HOME/.zshrc")

MARKER="# brain repo - auto-source config"
for shell_file in "${SHELL_FILES[@]}"; do
  if grep -q "$MARKER" "$shell_file" 2>/dev/null; then
    ok "$(basename $shell_file): already configured"
  else
    run_cmd bash -c "printf '%s' '$SHELL_SNIPPET' >> '$shell_file'"
    ok "$(basename $shell_file): brain.env auto-source added"
  fi
done

# ── 3. Regenerate adapters ────────────────────────────────────────────────────
info "Regenerating adapters (canonical.md -> all IDEs)"

if [ "$DRY_RUN" -eq 0 ]; then
  BRAIN_DIR="$BRAIN_DIR" bash "$BRAIN_DIR/adapters/generate.sh" 2>&1 | \
    grep -E "\[ok\]|All adapters|ERROR" | head -15
  ok "Adapters generated"
else
  skip "BRAIN_DIR=$BRAIN_DIR bash $BRAIN_DIR/adapters/generate.sh"
fi

# ── 4. Link adapters to IDE locations ─────────────────────────────────────────
info "Linking adapters to IDE config locations"

# Claude Code
mkdir -p "$HOME/.claude/agents" "$HOME/.claude/commands"

link() {
  local src="$1" dst="$2" label="$3"
  if [ "$DRY_RUN" -eq 1 ]; then
    skip "ln -sf $src $dst"
  elif [ -f "$src" ] || [ -d "$src" ]; then
    mkdir -p "$(dirname "$dst")"
    ln -sf "$src" "$dst"
    ok "$label"
  else
    warn "$label: source not found ($src)"
  fi
}

# Claude Code - rules
link "$BRAIN_DIR/adapters/claude-code/CLAUDE.md"             "$HOME/.claude/CLAUDE.md"             "Claude Code: CLAUDE.md"
link "$BRAIN_DIR/adapters/claude-code/settings.persistent.json" "$HOME/.claude/settings.json"     "Claude Code: settings.json"

# Claude Code - agents (one symlink per agent)
if [ "$DRY_RUN" -eq 0 ]; then
  for agent in "$BRAIN_DIR/agents/"*.md; do
    [ -f "$agent" ] || continue
    ln -sf "$agent" "$HOME/.claude/agents/$(basename "$agent")" 2>/dev/null || true
  done
  ok "Claude Code: agents linked (~/.claude/agents/)"
fi

# Claude Code - commands (one symlink per command)
if [ "$DRY_RUN" -eq 0 ]; then
  for cmd in "$BRAIN_DIR/commands/"*.md; do
    [ -f "$cmd" ] || continue
    ln -sf "$cmd" "$HOME/.claude/commands/$(basename "$cmd")" 2>/dev/null || true
  done
  ok "Claude Code: commands linked (~/.claude/commands/)"
fi

# Cursor
link "$BRAIN_DIR/adapters/cursor/.cursorrules"               "$HOME/.cursorrules"                  "Cursor: .cursorrules"

# Windsurf
link "$BRAIN_DIR/adapters/windsurf/.windsurfrules"           "$HOME/.windsurfrules"                "Windsurf: .windsurfrules"

# Gemini CLI
mkdir -p "$HOME/.gemini" 2>/dev/null || true
link "$BRAIN_DIR/adapters/gemini/GEMINI.md"                  "$HOME/.gemini/GEMINI.md"             "Gemini CLI: GEMINI.md"

# OpenCode (Linux path)
if [ "$(uname)" = "Linux" ] || uname -r 2>/dev/null | grep -qi microsoft; then
  mkdir -p "$HOME/.config/opencode" 2>/dev/null || true
  link "$BRAIN_DIR/adapters/opencode/opencode.json"          "$HOME/.config/opencode/opencode.json" "OpenCode: opencode.json (Linux)"
fi

# OpenCode (macOS path)
if [ "$(uname)" = "Darwin" ]; then
  mkdir -p "$HOME/Library/Application Support/opencode" 2>/dev/null || true
  link "$BRAIN_DIR/adapters/opencode/opencode.json"          "$HOME/Library/Application Support/opencode/opencode.json" "OpenCode: opencode.json (macOS)"
fi

# Aider
link "$BRAIN_DIR/adapters/aider/.aider.conf.yml"             "$HOME/.aider.conf.yml"               "Aider: .aider.conf.yml"

# Copilot (VS Code workspace instructions)
VSCODE_COPILOT="$HOME/.vscode/copilot-instructions.md"
link "$BRAIN_DIR/adapters/copilot/copilot-instructions.md"   "$VSCODE_COPILOT"                     "Copilot: instructions.md"

# ── 5. Inject brain-rules MCP server into Claude Code settings ────────────────
info "Registering brain-rules MCP server in Claude Code"

CLAUDE_SETTINGS="$HOME/.claude/settings.json"

if [ -f "$CLAUDE_SETTINGS" ] && [ "$DRY_RUN" -eq 0 ]; then
  python3 - "$CLAUDE_SETTINGS" "$BRAIN_DIR" << 'PY'
import json, sys, pathlib

settings_path = pathlib.Path(sys.argv[1])
brain_dir = sys.argv[2]

try:
    settings = json.loads(settings_path.read_text())
except Exception:
    settings = {}

if "mcpServers" not in settings:
    settings["mcpServers"] = {}

brain_mcp_entry = {
    "command": "python3",
    "args": [f"{brain_dir}/mcp/brain-mcp-server/server.py"]
}

if settings["mcpServers"].get("brain-rules") != brain_mcp_entry:
    settings["mcpServers"]["brain-rules"] = brain_mcp_entry
    settings_path.write_text(json.dumps(settings, indent=2))
    print("  brain-rules MCP server registered in Claude Code settings")
else:
    print("  brain-rules MCP server already registered")
PY
  ok "Claude Code: brain-rules MCP registered"
elif [ "$DRY_RUN" -eq 1 ]; then
  skip "inject brain-rules MCP into $CLAUDE_SETTINGS"
else
  warn "Claude Code settings.json not found at $CLAUDE_SETTINGS"
fi

# ── 6. MCP sync to all IDEs ───────────────────────────────────────────────────
info "Syncing MCP config to all IDEs"

if [ -f "$BRAIN_DIR/scripts/mcp-sync.sh" ]; then
  if [ "$DRY_RUN" -eq 0 ]; then
    BRAIN_DIR="$BRAIN_DIR" bash "$BRAIN_DIR/scripts/mcp-sync.sh" 2>&1 | \
      grep -E "ok|warn|ERROR" | head -10
    ok "MCP servers synced to all IDEs"
  else
    skip "bash $BRAIN_DIR/scripts/mcp-sync.sh"
  fi
fi

# ── 7. Guardian git hook (global - all repos) ─────────────────────────────────
info "Guardian: global git pre-commit hook"

if [ "$DRY_RUN" -eq 0 ]; then
  bash "$BRAIN_DIR/scripts/install-hooks.sh" --global 2>&1 && \
    ok "Guardian hook installed globally (every new git repo will have it)" || \
    warn "Global hook install failed - install per-repo: bash ~/.brain/scripts/install-hooks.sh"
else
  skip "bash $BRAIN_DIR/scripts/install-hooks.sh --global"
fi

# ── 8. Cron jobs (automated maintenance) ─────────────────────────────────────
info "Automated maintenance (cron)"

if command -v crontab >/dev/null 2>&1; then
  if crontab -l 2>/dev/null | grep -q "brain-repo-managed"; then
    ok "Cron jobs already installed"
  else
    if [ "$DRY_RUN" -eq 0 ]; then
      bash "$BRAIN_DIR/scripts/cron-setup.sh" 2>&1
      ok "Cron jobs installed"
    else
      skip "bash $BRAIN_DIR/scripts/cron-setup.sh"
    fi
  fi
else
  warn "crontab not available on this system"
fi

# ── 9. Verify everything ──────────────────────────────────────────────────────
info "Running doctor to verify setup"

if [ "$DRY_RUN" -eq 0 ]; then
  BRAIN_DIR="$BRAIN_DIR" bash "$BRAIN_DIR/scripts/doctor.sh" 2>/dev/null
else
  skip "bash $BRAIN_DIR/scripts/doctor.sh"
fi

# ── 10. Summary ───────────────────────────────────────────────────────────────
echo ""
echo -e "${B}Setup complete.${RS}"
echo ""
echo -e "${B}What is now automatic (no manual action needed):${RS}"
echo ""
echo "  On every shell open:"
echo "    - brain.env loaded (API keys available everywhere)"
echo "    - BRAIN_DIR set"
echo ""
echo "  On every IDE open:"
echo "    - Claude Code reads ~/.claude/CLAUDE.md (symlink -> brain rules)"
echo "    - Claude Code has all agents in ~/.claude/agents/"
echo "    - Claude Code has all commands in ~/.claude/commands/"
echo "    - Claude Code settings.json has all MCP servers configured"
echo "    - Cursor reads ~/.cursorrules (symlink -> brain rules)"
echo "    - Windsurf reads ~/.windsurfrules (symlink -> brain rules)"
echo "    - Gemini CLI reads ~/.gemini/GEMINI.md (symlink -> brain rules)"
echo "    - OpenCode reads its config (symlink -> brain opencode.json)"
echo "    - Aider reads ~/.aider.conf.yml (symlink -> brain aider config)"
echo ""
echo "  On every git commit (any repo):"
echo "    - Guardian pre-commit hook runs automatically"
echo "    - Blocks: hardcoded secrets, explicit any, tracked .env files"
echo ""
echo "  Scheduled (cron):"
echo "    - Daily  02:00 - schema validation"
echo "    - Daily  02:30 - doctor check"
echo "    - Weekly Sun - memory consolidation + eval suite"
echo "    - Monthly 1st - vector index rebuild"
echo ""
echo -e "${B}What you need to do manually (one time):${RS}"
echo ""
echo "  1. Set your API key in brain.env:"
echo "     echo 'ANTHROPIC_API_KEY=\"sk-ant-...\"' >> ~/.brain/brain.env"
echo ""
echo "  2. Reload your shell:"
echo "     source ~/.zshrc   # or source ~/.bashrc"
echo ""
echo "  3. (Optional) Cloud memory sync - pick one:"
echo "     bash ~/.brain/scripts/memory-cloud-sync.sh setup"
echo ""
echo "  After that - just open your IDE and start working."
echo "  Nothing else to start, no daemons, no docker required."
echo ""
