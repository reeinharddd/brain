#!/bin/bash
# scripts/autostart-setup.sh
# Registers the brain environment for automatic startup.

BRAIN_DIR="$HOME/.brain"
INIT_SCRIPT="$BRAIN_DIR/scripts/init.sh"
OS_TYPE=$(uname -s)

echo "Setting up autostart for OS: $OS_TYPE"

case "$OS_TYPE" in
  Linux*)
    if [ -d "$HOME/.config/systemd/user" ]; then
      # Create systemd user service
      mkdir -p "$HOME/.config/systemd/user"
      cat <<EOF > "$HOME/.config/systemd/user/brain.service"
[Unit]
Description=Brain Environment Initializer
After=network.target

[Service]
Type=oneshot
ExecStart=/bin/bash $INIT_SCRIPT
RemainAfterExit=yes

[Install]
WantedBy=default.target
EOF
      systemctl --user daemon-reload
      systemctl --user enable brain.service
      echo "[OK] Registered systemd user service: brain.service"
    fi
    ;;
  Darwin*)
    # macOS LaunchAgent placeholder
    echo "[INFO] macOS detected. Manual registration in LaunchAgents may be required."
    ;;
  *)
    echo "[WARN] Unsupported OS for automated autostart. Please run $INIT_SCRIPT manually."
    ;;
esac

# Add to shell profile if not present
SHELL_RC="$HOME/.bashrc"
if [ -n "$ZSH_VERSION" ]; then SHELL_RC="$HOME/.zshrc"; fi

if ! grep -q "BRAIN_READY" "$SHELL_RC"; then
  cat <<EOF >> "$SHELL_RC"

# Brain Environment
if [ -f "$INIT_SCRIPT" ]; then
  source "$INIT_SCRIPT"
fi
EOF
  echo "[OK] Added brain initialization to $SHELL_RC"
fi
