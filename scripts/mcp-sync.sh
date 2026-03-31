#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/scripts/mcp-sync.sh
#  Synchronizes MCP configurations from global-config.json
#  to all supported IDEs (Cursor, VS Code, Claude Code).
#  Updated for MCP Gateway architecture.
# ═══════════════════════════════════════════════════════════

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
GLOBAL_CONFIG="$BRAIN_DIR/mcp/global-config.json"
GLOBAL_STDIO_CONFIG="$BRAIN_DIR/mcp/global-config-stdio.json"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; RESET='\033[0m'
ok()   { echo -e "  ${GREEN}✓${RESET} $1"; }
info() { echo -e "  ${BOLD}── $1${RESET}"; }
warn() { echo -e "  ${YELLOW}!${RESET} $1"; }

# ── Substitute environment variables in config ─────────────────
substitute_env_vars() {
    local content="$1"
    # Replace ${VAR} and $VAR with actual values
    content=$(echo "$content" | sed "s|\\\${BRAIN_DIR}|${BRAIN_DIR}|g")
    content=$(echo "$content" | sed "s|\\\${HOME}|${HOME}|g")
    echo "$content"
}

if [ ! -f "$GLOBAL_CONFIG" ] && [ ! -f "$GLOBAL_STDIO_CONFIG" ]; then
    echo "ERROR: Config files not found at $GLOBAL_CONFIG" >&2
    exit 1
fi

# Determine source for IDEs (prefer stdio for native experience)
SYNC_SOURCE="$GLOBAL_STDIO_CONFIG"
if [ ! -f "$SYNC_SOURCE" ]; then
    SYNC_SOURCE="$GLOBAL_CONFIG"
fi

info "Synchronizing MCP configurations from $SYNC_SOURCE"

# Check if MCP Gateway is running
if curl -sf http://localhost:3000/health >/dev/null 2>&1; then
    ok "MCP Gateway detected at localhost:3000"
else
    warn "MCP Gateway not detected. Run 'brain up' first for full functionality."
fi

# 1. Claude Code
CLAUDE_SETTINGS="$HOME/.claude/settings.json"
if [ -f "$CLAUDE_SETTINGS" ] || [ -d "$(dirname "$CLAUDE_SETTINGS")" ]; then
    mkdir -p "$(dirname "$CLAUDE_SETTINGS")"
    
    python3 -c "
import json, os, re

brain_dir = os.environ.get('BRAIN_DIR', os.path.expanduser('~/.brain'))

with open('$SYNC_SOURCE', 'r') as f:
    global_mcp = json.load(f).get('mcpServers', {})

# Substitute environment variables
def substitute_vars(obj):
    if isinstance(obj, str):
        obj = obj.replace('\\${BRAIN_DIR}', brain_dir)
        obj = obj.replace('\\$BRAIN_DIR', brain_dir)
        obj = obj.replace('\\${HOME}', os.path.expanduser('~'))
        return obj
    elif isinstance(obj, list):
        return [substitute_vars(item) for item in obj]
    elif isinstance(obj, dict):
        return {k: substitute_vars(v) for k, v in obj.items()}
    return obj

def normalize_entry(config):
    # Ensure command is a string and args is a list when possible
    if not isinstance(config, dict):
        return config
    cmd = config.get('command')
    if isinstance(cmd, list) and len(cmd) > 0:
        # convert list form to command + args
        config['command'] = cmd[0]
        if 'args' not in config:
            config['args'] = cmd[1:]
    # Recursively normalize nested dicts
    for k, v in list(config.items()):
        if isinstance(v, dict):
            config[k] = normalize_entry(v)
    return config

global_mcp = substitute_vars(global_mcp)

# Normalize entries so IDEs that expect a string 'command' don't crash
for name in list(global_mcp.keys()):
    global_mcp[name] = normalize_entry(global_mcp[name])

# Read or create settings
if os.path.exists('$CLAUDE_SETTINGS'):
    with open('$CLAUDE_SETTINGS', 'r') as f:
        try:
            settings = json.load(f)
        except json.JSONDecodeError:
            settings = {}
else:
    settings = {}

if 'mcpServers' not in settings:
    settings['mcpServers'] = {}

# Merge mcpServers
for name, config in global_mcp.items():
    settings['mcpServers'][name] = config

with open('$CLAUDE_SETTINGS', 'w') as f:
    json.dump(settings, f, indent=2)
"
    ok "Claude Code synced (~/.claude/settings.json)"
fi

# 2. IDEs (VS Code, Cursor, Windsurf)
# We look for standard paths and filenames across common IDEs
PATHS=(
    # Extension-specific (Cline/Roo-Code)
    "$HOME/.config/Cursor/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json"
    "$HOME/.config/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json"
    "$HOME/.config/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings/cline_mcp_settings.json"
    "$HOME/.config/Code/User/globalStorage/rooveterinaryinc.roo-cline/settings/cline_mcp_settings.json"
    
    # Native IDE integrations (Standard)
    "$HOME/.config/Code/User/mcp.json"
    "$HOME/.config/Cursor/User/mcp.json"
    "$HOME/.codeium/windsurf/mcp_config.json"

    # Claude Desktop
    "$HOME/.config/Claude/claude_desktop_config.json"
)

# Erroneous/Confusing filenames to cleanup
CLEANUP_PATHS=(
    "$HOME/.config/Code/User/mcp-config.json"
    "$HOME/.config/Cursor/User/mcp-config.json"
)

for target in "${PATHS[@]}"; do
    if [ -f "$target" ] || [ -d "$(dirname "$target")" ]; then
        mkdir -p "$(dirname "$target")"
        if [ ! -f "$target" ]; then
            echo '{"mcpServers": {}}' > "$target"
        fi
        
        info "Syncing to $target"
        python3 -c "
import json, os

brain_dir = os.environ.get('BRAIN_DIR', os.path.expanduser('~/.brain'))

with open('$SYNC_SOURCE', 'r') as f:
    global_mcp = json.load(f).get('mcpServers', {})

# Substitute environment variables
def substitute_vars(obj):
    if isinstance(obj, str):
        obj = obj.replace('\\${BRAIN_DIR}', brain_dir)
        obj = obj.replace('\\$BRAIN_DIR', brain_dir)
        obj = obj.replace('\\${HOME}', os.path.expanduser('~'))
        return obj
    elif isinstance(obj, list):
        return [substitute_vars(item) for item in obj]
    elif isinstance(obj, dict):
        return {k: substitute_vars(v) for k, v in obj.items()}
    return obj

def normalize_entry(config):
    # Ensure command is a string and args is a list when possible
    if not isinstance(config, dict):
        return config
    cmd = config.get('command')
    if isinstance(cmd, list) and len(cmd) > 0:
        config['command'] = cmd[0]
        if 'args' not in config:
            config['args'] = cmd[1:]
    for k, v in list(config.items()):
        if isinstance(v, dict):
            config[k] = normalize_entry(v)
    return config

global_mcp = substitute_vars(global_mcp)

# Normalize so targets receive predictable shapes
for name in list(global_mcp.keys()):
    global_mcp[name] = normalize_entry(global_mcp[name])

with open('$target', 'r') as f:
    try:
        settings = json.load(f)
    except json.JSONDecodeError:
        settings = {}

# Handle different key names
key = 'mcpServers'
if 'servers' in settings and 'mcpServers' not in settings:
    key = 'servers'

if key not in settings:
    settings[key] = {}

# Merge global configs into target
for name, config in global_mcp.items():
    settings[key][name] = config

with open('$target', 'w') as f:
    json.dump(settings, f, indent=2)
"
    fi
done

# Cleanup erroneous files
for bad_file in "${CLEANUP_PATHS[@]}"; do
    if [ -f "$bad_file" ]; then
        rm "$bad_file"
        ok "Removed erroneous config: $bad_file"
    fi
done

# 3. Global registry exposure (as a check)
ok "Global MCP registry: $GLOBAL_CONFIG"
echo ""
echo -e "  ${GREEN}${BOLD}✓ Synchronization complete${RESET}"
