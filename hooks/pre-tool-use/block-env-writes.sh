#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/hooks/pre-tool-use/block-env-writes.sh
#  Pre-tool-use hook for Claude Code.
#  Blocks AI agents from writing secrets to environment files.
#
#  Called with: bash block-env-writes.sh "$TOOL_INPUT"
#  Exit 0 = allow, Exit 2 = block with message
# ═══════════════════════════════════════════════════════════

TOOL_INPUT="${1:-}"

# ── Patterns to block ─────────────────────────────────────────

# Block writing to .env files (prevent secret leakage)
if echo "$TOOL_INPUT" | grep -qE '\.env(\.|$|/)' 2>/dev/null; then
  echo "🛡️  GUARDIAN BLOCKED: Attempt to write to .env file detected."
  echo "   Agents should never modify .env files directly."
  echo "   To add a secret: edit .env manually, or use a secrets manager."
  exit 2
fi

# Block commands that set environment variables to files
if echo "$TOOL_INPUT" | grep -qE 'echo.*API_KEY|echo.*SECRET|echo.*PASSWORD|echo.*TOKEN' 2>/dev/null; then
  if echo "$TOOL_INPUT" | grep -qE '>>?\s*\.(env|bashrc|zshrc|profile)' 2>/dev/null; then
    echo "🛡️  GUARDIAN BLOCKED: Attempt to write a secret to a shell config file."
    echo "   Manage secrets manually or via a secrets manager."
    exit 2
  fi
fi

# Block writing private keys directly
if echo "$TOOL_INPUT" | grep -qE 'BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY' 2>/dev/null; then
  echo "🛡️  GUARDIAN BLOCKED: Private key content detected in tool input."
  echo "   Never let an AI agent handle raw private key material."
  exit 2
fi

# Allow everything else
exit 0
