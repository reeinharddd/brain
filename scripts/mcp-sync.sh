#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/scripts/mcp-sync.sh
#  Synchronizes MCP configurations from global-config.json
#  to all supported IDEs (Cursor, VS Code, Claude Code).
# ═══════════════════════════════════════════════════════════

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
GLOBAL_CONFIG="$BRAIN_DIR/mcp/global-config.json"
GLOBAL_STDIO_CONFIG="$BRAIN_DIR/mcp/global-config-stdio.json"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; RESET='\033[0m'
ok()   { echo -e "  ${GREEN}✓${RESET} $1"; }
info() { echo -e "  ${BOLD}── $1${RESET}"; }

if [ ! -f "$GLOBAL_CONFIG" ] && [ ! -f "$GLOBAL_STDIO_CONFIG" ]; then
    echo "ERROR: Config files not found. Run generate.sh first." >&2
    exit 1
fi

# Determine source for IDEs (prefer stdio for native experience)
SYNC_SOURCE="$GLOBAL_STDIO_CONFIG"
if [ ! -f "$SYNC_SOURCE" ]; then
    SYNC_SOURCE="$GLOBAL_CONFIG"
fi

info "Synchronizing MCP configurations from $SYNC_SOURCE"

# 1. Claude Code
CLAUDE_SETTINGS="$HOME/.claude/settings.json"
if [ -f "$CLAUDE_SETTINGS" ]; then
    # Update claude settings with MCP servers from global config
    # We use python3 for robust JSON merging
    python3 -c "
import json, sys
with open('$SYNC_SOURCE', 'r') as f:
    global_mcp = json.load(f).get('mcpServers', {})
with open('$CLAUDE_SETTINGS', 'r') as f:
    settings = json.load(f)

# Merge mcpServers
if 'mcpServers' not in settings:
    settings['mcpServers'] = {}

# We only update if it's SSE based on our docker stack
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
with open('$SYNC_SOURCE', 'r') as f:
    global_mcp = json.load(f).get('mcpServers', {})

with open('$target', 'r') as f:
    try:
        settings = json.load(f)
    except json.JSONDecodeError:
        settings = {}

# Handle different key names: 'mcpServers' (vsc/cursor/cline) vs 'servers' (native vsc/windsurf)
# We prioritize 'mcpServers' if it exists, otherwise check 'servers'
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
