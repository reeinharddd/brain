#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/adapters/generate.sh
#  Reads rules/canonical.md + rules/modules/*.md and produces
#  adapter files for every supported AI agent/IDE.
#
#  Usage: bash ~/.brain/adapters/generate.sh
#  Idempotent: safe to run multiple times.
# ═══════════════════════════════════════════════════════════

set -euo pipefail

BRAIN_DIR="$HOME/.brain"
SOURCE="$BRAIN_DIR/rules/canonical.md"
MODULES_DIR="$BRAIN_DIR/rules/modules"
NPX_CMD="/home/reeinharrrd/.local/bin/npx-nvm" # Use wrapper to avoid ENOENT in IDEs
BUILD_RULES_SCRIPT="$BRAIN_DIR/scripts/build-rules.sh"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; RESET='\033[0m'
ok()   { echo -e "  ${GREEN}✓${RESET} $1"; }
warn() { echo -e "  ${YELLOW}⚠${RESET}  $1"; }

echo -e "\n${BOLD}── Generating rule adapters${RESET}"

# ── Assemble full rules from canonical + modules ─────────────
if [ ! -x "$BUILD_RULES_SCRIPT" ]; then
  echo "ERROR: $BUILD_RULES_SCRIPT not found or not executable." >&2
  exit 1
fi

FULL_RULES="$("$BUILD_RULES_SCRIPT" --stdout)"
"$BUILD_RULES_SCRIPT" >/dev/null

TIMESTAMP="Generated on $(date '+%Y-%m-%d %H:%M:%S') · Source: ~/.brain/rules/canonical.md + modules/"
HEADER_COMMENT="<!-- AUTO-GENERATED — DO NOT EDIT DIRECTLY -->\n<!-- $TIMESTAMP -->\n\n"

# ── Helper: write file ────────────────────────────────────────
write_adapter() {
  local path="$1"
  local content="$2"
  mkdir -p "$(dirname "$path")"
  printf "%b" "$content" > "$path"
}

# ═══════════════════════════════════════════════════════════
#  1. Claude Code — ~/.brain/adapters/claude-code/CLAUDE.md
# ═══════════════════════════════════════════════════════════
CLAUDE_CONTENT="${HEADER_COMMENT}${FULL_RULES}"
write_adapter "$BRAIN_DIR/adapters/claude-code/CLAUDE.md" "$CLAUDE_CONTENT"
ok "claude-code/CLAUDE.md"

# ═══════════════════════════════════════════════════════════
#  2. Cursor — ~/.brain/adapters/cursor/.cursorrules
# ═══════════════════════════════════════════════════════════
CURSOR_CONTENT="${HEADER_COMMENT}${FULL_RULES}"
write_adapter "$BRAIN_DIR/adapters/cursor/.cursorrules" "$CURSOR_CONTENT"
ok "cursor/.cursorrules"

# ═══════════════════════════════════════════════════════════
#  3. Windsurf — ~/.brain/adapters/windsurf/.windsurfrules
# ═══════════════════════════════════════════════════════════
WINDSURF_CONTENT="${HEADER_COMMENT}${FULL_RULES}"
write_adapter "$BRAIN_DIR/adapters/windsurf/.windsurfrules" "$WINDSURF_CONTENT"
ok "windsurf/.windsurfrules"

# ═══════════════════════════════════════════════════════════
#  4. Gemini CLI — ~/.brain/adapters/gemini/GEMINI.md
# ═══════════════════════════════════════════════════════════
GEMINI_CONTENT="${HEADER_COMMENT}${FULL_RULES}"
write_adapter "$BRAIN_DIR/adapters/gemini/GEMINI.md" "$GEMINI_CONTENT"
ok "gemini/GEMINI.md"

# ═══════════════════════════════════════════════════════════
#  5. Cline (VS Code extension) — custom instructions
# ═══════════════════════════════════════════════════════════
CLINE_CONTENT="${HEADER_COMMENT}# Cline Custom Instructions\n\n${FULL_RULES}\n\n---\n*To apply: Open Cline settings → Custom Instructions → paste the content of this file.*"
write_adapter "$BRAIN_DIR/adapters/cline/cline_custom_instructions.md" "$CLINE_CONTENT"
ok "cline/cline_custom_instructions.md"

# ═══════════════════════════════════════════════════════════
#  6. Aider — system prompt + conf
# ═══════════════════════════════════════════════════════════
AIDER_PROMPT_PATH="$BRAIN_DIR/adapters/aider/system-prompt.md"
write_adapter "$AIDER_PROMPT_PATH" "${FULL_RULES}"

cat > "$BRAIN_DIR/adapters/aider/.aider.conf.yml" <<YAML
# AUTO-GENERATED — DO NOT EDIT DIRECTLY
# Source: ~/.brain/rules/canonical.md + modules/
# Generated: $(date '+%Y-%m-%d %H:%M:%S')

# Model (override per-project as needed)
# model: claude-sonnet-4-5

# System prompt from brain repo
system-prompt: $AIDER_PROMPT_PATH

# Git settings
auto-commits: false
dirty-commits: true
commit-prompt: "Conventional Commits format: <type>(<scope>): <description>"

# Output
pretty: true
stream: true

# Safety
no-auto-accept-architect: true
YAML
ok "aider/.aider.conf.yml + system-prompt.md"

# ═══════════════════════════════════════════════════════════
#  7. OpenCode — opencode.json
# ═══════════════════════════════════════════════════════════
# Escape rules for JSON (basic escaping)
ESCAPED_RULES=$(echo "$FULL_RULES" | python3 -c "
import sys, json
content = sys.stdin.read()
print(json.dumps(content))
" 2>/dev/null || echo '"[rules — run generate.sh with python3 available]"')

OPENCODE_CONTENT=$(cat <<JSON
{
  "//": "AUTO-GENERATED — DO NOT EDIT DIRECTLY. Source: ~/.brain/rules/",
  "model": "anthropic/claude-sonnet-4-5",
  "systemPrompt": $ESCAPED_RULES,
  "mcpServers": {
    "memory": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-memory", "${BRAIN_DIR}/memory"]
    },
    "filesystem": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "${HOME}"]
    },
    "sequential-thinking": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    },
    "context7": {
      "command": "$NPX_CMD",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "github": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    },
    "skill-ninja": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "aktsmm/skill-ninja-mcp-server:latest"]
    },
    "duckduckgo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "mcp/duckduckgo:latest"]
    },
    "crawl4ai": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "unclecode/crawl4ai:latest"]
    },
    "context-awesome": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "bh-rat/context-awesome:latest"]
    }
  }
}
JSON
)
write_adapter "$BRAIN_DIR/adapters/opencode/opencode.json" "$OPENCODE_CONTENT"
ok "opencode/opencode.json"

# ═══════════════════════════════════════════════════════════
#  8. Global IDE MCP Config — mcp/global-config.json
# ═══════════════════════════════════════════════════════════
# This file is used by IDEs that support command-based MCP configs.
GLOBAL_MCP_CONTENT=$(cat <<JSON
{
  "mcpServers": {
    "memory": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-memory", "${BRAIN_DIR}/memory"]
    },
    "filesystem": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "${HOME}"]
    },
    "sequential-thinking": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    },
    "context7": {
      "command": "$NPX_CMD",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "github": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    },
    "skill-ninja": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "aktsmm/skill-ninja-mcp-server:latest"]
    },
    "duckduckgo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "mcp/duckduckgo:latest"]
    },
    "context-awesome": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "bh-rat/context-awesome:latest"]
    }
  }
}
JSON
)
write_adapter "$BRAIN_DIR/mcp/global-config.json" "$GLOBAL_MCP_CONTENT"
ok "mcp/global-config.json (Hybrid command mode)"

# ═══════════════════════════════════════════════════════════
#  9. Native MCP Config (stdio) — mcp/global-config-stdio.json
# ═══════════════════════════════════════════════════════════
# For use with Claude Desktop, Cursor (individual tools), etc.
GLOBAL_STDIO_MCP_CONTENT=$(cat <<JSON
{
  "mcpServers": {
    "memory": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-memory", "${BRAIN_DIR}/memory"]
    },
    "filesystem": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "${HOME}"]
    },
    "sequential-thinking": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    },
    "context7": {
      "command": "$NPX_CMD",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "github": {
      "command": "$NPX_CMD",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    },
    "skill-ninja": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "aktsmm/skill-ninja-mcp-server:latest"]
    },
    "duckduckgo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "mcp/duckduckgo:latest"]
    },
    "context-awesome": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "bh-rat/context-awesome:latest"]
    }
  }
}
JSON
)
write_adapter "$BRAIN_DIR/mcp/global-config-stdio.json" "$GLOBAL_STDIO_MCP_CONTENT"
write_adapter "$BRAIN_DIR/adapters/claude-desktop/claude_desktop_config.json" "$GLOBAL_STDIO_MCP_CONTENT"
ok "mcp/global-config-stdio.json + claude-desktop (Hybrid stdio)"

# ── GitHub Copilot ──────────────────────────────────────────
mkdir -p "$BRAIN_DIR/adapters/copilot"
echo "$FULL_RULES" > "$BRAIN_DIR/adapters/copilot/copilot-instructions.md"
ok "copilot/copilot-instructions.md"

# ═══════════════════════════════════════════════════════════
#  Summary
# ═══════════════════════════════════════════════════════════
echo ""
echo -e "  ${BOLD}Total rules size:${RESET} $(echo "$FULL_RULES" | wc -l) lines"
echo -e "  ${BOLD}Adapters generated:${RESET} claude-code · cursor · windsurf · gemini · cline · aider · opencode · copilot · test · init"
echo ""
echo -e "  ${GREEN}${BOLD}✓ All adapters up to date${RESET}"
echo "  Now run: bash ~/.brain/install.sh to apply symlinks"
echo ""
