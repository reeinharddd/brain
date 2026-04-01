#!/bin/bash
# Script separator
#  install.sh - Complete Brain repo setup (consolidated from 4 scripts)
#  
#  Integrates: OS bootstrap + persistent setup + CLI setup + autostart
#
#  Usage:
#    bash ~/.brain/scripts/install.sh              # Full install
#    bash ~/.brain/scripts/install.sh --bootstrap  # OS packages only
#    bash ~/.brain/scripts/install.sh --persistent # Setup without bootstrap
# Script separator

set -e

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
OS="unknown"
PACKAGE_MANAGER="unknown"
ERRORS=0
DRY_RUN=0

# -- Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

ok()      { echo -e "  ${GREEN}[ok]${NC} $1"; }
warn()    { echo -e "  ${YELLOW}[warn]${NC} $1"; }
fail()    { echo -e "  ${RED}[fail]${NC} $1"; ((ERRORS++)); }
info()    { echo -e "\n${BOLD}-- $1${NC}"; }
section() { echo -e "\n${BOLD}$1${NC}"; }

run_cmd() {
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "[dry-run] $*"
  else
    "$@" || fail "Command failed: $*"
  fi
}

# Parse arguments
for arg in "$@"; do
  case "$arg" in
    --dry-run)  DRY_RUN=1 ;;
  esac
done

# Script separator
# PHASE 1: OS Bootstrap
# Script separator

section "Phase 1: OS Bootstrap"

# Detect OS and package manager
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  OS="Linux"
  if command -v apt &>/dev/null; then
    PACKAGE_MANAGER="apt"
  elif command -v dnf &>/dev/null; then
    PACKAGE_MANAGER="dnf"
  fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
  OS="macOS"
  PACKAGE_MANAGER="brew"
else
  OS="Unknown"
fi

info "Detected OS: $OS ($PACKAGE_MANAGER)"

case "$PACKAGE_MANAGER" in
  apt)
    run_cmd sudo apt-get update
    run_cmd sudo apt-get install -y python3 python3-pip bash zsh git docker.io docker-compose
    ;;
  dnf)
    run_cmd sudo dnf install -y python3 python3-pip bash zsh git docker docker-compose
    ;;
  brew)
    run_cmd brew install python3 git bash zsh docker docker-compose
    ;;
  *)
    warn "Unsupported package manager - please install dependencies manually"
    ;;
esac

ok "OS bootstrap complete"

# Script separator
# PHASE 2: Generate Adapters & Link IDE Rules
# Script separator

section "Phase 2: Adapter Generation & IDE Setup"

if [ -f "$BRAIN_DIR/adapters/generate.sh" ]; then
  run_cmd bash "$BRAIN_DIR/adapters/generate.sh"
  ok "Adapters generated"
fi

# Link IDE config files
[ -f "$BRAIN_DIR/adapters/cursor/.cursorrules" ] && \
  run_cmd ln -sf "$BRAIN_DIR/adapters/cursor/.cursorrules" "$HOME/.cursorrules" && \
  ok "Linked .cursorrules"

[ -f "$BRAIN_DIR/adapters/windsurf/.windsurfrules" ] && \
  run_cmd ln -sf "$BRAIN_DIR/adapters/windsurf/.windsurfrules" "$HOME/.windsurfrules" && \
  ok "Linked .windsurfrules"

[ -f "$BRAIN_DIR/adapters/aider/.aider.conf.yml" ] && \
  run_cmd ln -sf "$BRAIN_DIR/adapters/aider/.aider.conf.yml" "$HOME/.aider.conf.yml" && \
  ok "Linked aider config"

ok "IDE adapters configured"

# Script separator
# PHASE 3: Persistent Setup (env, shell integration, git hooks)
# Script separator

section "Phase 3: Persistent Setup"

# Create brain.env from example
ENV_FILE="$BRAIN_DIR/brain.env"
ENV_EXAMPLE="$BRAIN_DIR/brain.env.example"

if [ ! -f "$ENV_FILE" ] && [ -f "$ENV_EXAMPLE" ]; then
  run_cmd cp "$ENV_EXAMPLE" "$ENV_FILE"
  ok "brain.env created (add API keys manually)"
else
  ok "brain.env already exists"
fi

# Shell integration
SHELL_RC="$HOME/.zshrc"
[ ! -f "$SHELL_RC" ] && [ -f "$HOME/.bashrc" ] && SHELL_RC="$HOME/.bashrc"

if [ -f "$SHELL_RC" ] && ! grep -q "brain.env" "$SHELL_RC"; then
  run_cmd bash <<EOF
cat >> "$SHELL_RC" <<'SHELL_EOF'

# Brain repo - auto-source config
[ -f "\$HOME/.brain/brain.env" ] && set -a && . "\$HOME/.brain/brain.env" && set +a
export BRAIN_DIR="\$HOME/.brain"
SHELL_EOF
EOF
  ok "Shell profile updated"
fi

# Git hooks
if [ -f "$BRAIN_DIR/hooks/pre-commit.sh" ]; then
  run_cmd bash "$BRAIN_DIR/hooks/pre-commit.sh"
  ok "Git hooks installed"
fi

ok "Persistent setup complete"

# Script separator
# PHASE 4: CLI Setup (make executable, PATH, test)
# Script separator

section "Phase 4: CLI Setup"

run_cmd chmod +x "$BRAIN_DIR/scripts/brain-cli.sh" "$BRAIN_DIR/scripts/init.sh" "$BRAIN_DIR/bin/brain"
ok "Scripts made executable"

run_cmd mkdir -p "$HOME/.local/bin"
run_cmd ln -sf "$BRAIN_DIR/bin/brain" "$HOME/.local/bin/brain"

if ! echo "$PATH" | grep -q "${HOME}/.local/bin"; then
  warn "${HOME}/.local/bin not in PATH - add to ~/.zshrc or ~/.bashrc)"
fi

# Test
if bash "$BRAIN_DIR/scripts/brain-cli.sh" --version >/dev/null 2>&1; then
  ok "brain-cli.sh works"
else
  fail "brain-cli.sh test failed"
fi

ok "CLI setup complete"

# Script separator
# PHASE 5: Autostart Setup (systemd / shell profile)
# Script separator

section "Phase 5: Autostart Registration"

OS_TYPE=$(uname -s)

if [[ "$OS_TYPE" == "Linux"* ]]; then
  mkdir -p "$HOME/.config/systemd/user"
  run_cmd bash <<EOF
cat > "$HOME/.config/systemd/user/brain.service" <<'SYSEOF'
[Unit]
Description=Brain Environment
After=network.target

[Service]
Type=oneshot
ExecStart=$BRAIN_DIR/scripts/init.sh
RemainAfterExit=yes

[Install]
WantedBy=default.target
SYSEOF
EOF
  if [ "$DRY_RUN" -eq 0 ]; then
    systemctl --user daemon-reload 2>/dev/null || true
    systemctl --user enable brain.service 2>/dev/null || true
  fi
  ok "Systemd user service registered"
elif [[ "$OS_TYPE" == "Darwin"* ]]; then
  warn "macOS: Manual LaunchAgent registration required"
fi

ok "Autostart registration complete"

# Script separator
# Summary
# Script separator

echo ""
section "Installation Complete!"

if [ "$ERRORS" -eq 0 ]; then
  ok "All phases completed successfully"
  echo ""
  echo "Next steps:"
  echo "  1. Set API keys: edit $BRAIN_DIR/brain.env"
  echo "  2. Test installation: brain status"
  echo "  3. Start services: brain start"
else
  warn "Completed with $ERRORS error(s)"
  echo "  Run script/doctor.sh for diagnosis"
  exit 1
fi

echo ""
