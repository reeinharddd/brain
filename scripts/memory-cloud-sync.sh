#!/bin/bash
# memory-cloud-sync.sh - Sync brain memory to/from cloud storage.
#
# Supports 4 backends (configured in brain.env):
#   A) Mem0 Cloud     - MEM0_API_KEY (best: managed, cross-device, semantic)
#   B) Qdrant Cloud   - QDRANT_URL + QDRANT_API_KEY (self-service vector DB)
#   C) Upstash Vector - UPSTASH_VECTOR_URL + UPSTASH_VECTOR_TOKEN (serverless)
#   D) Git sync       - BRAIN_MEMORY_GIT_SYNC=true (free, uses your git remote)
#
# Usage:
#   bash memory-cloud-sync.sh status
#   bash memory-cloud-sync.sh push      # local -> cloud
#   bash memory-cloud-sync.sh pull      # cloud -> local (git only)
#   bash memory-cloud-sync.sh setup     # interactive setup guide

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
COMMAND="${1:-status}"

# Load brain.env
[ -f "$BRAIN_DIR/brain.env" ] && set -a && . "$BRAIN_DIR/brain.env" && set +a

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; RESET='\033[0m'
ok()   { echo -e "  ${GREEN}[ok]${RESET}   $1"; }
warn() { echo -e "  ${YELLOW}[warn]${RESET} $1"; }
fail() { echo -e "  ${RED}[fail]${RESET} $1"; }
info() { echo -e "  ${BOLD}[-->]${RESET}  $1"; }

# -- Status -------------------------------------------------------------------
cmd_status() {
  echo ""
  echo -e "${BOLD}Memory Cloud Sync Status${RESET}"
  echo ""

  # Mem0
  if [ -n "${MEM0_API_KEY:-}" ]; then
    if python3 -c "
import urllib.request, json, os, sys
req = urllib.request.Request('https://api.mem0.ai/v1/memories/', headers={'Authorization': 'Token ' + os.environ['MEM0_API_KEY']})
try:
    resp = json.loads(urllib.request.urlopen(req, timeout=5).read())
    print('ok:' + str(len(resp.get('results',[]))))
except: print('fail')
" 2>/dev/null | grep -q "^ok:"; then
      COUNT=$(python3 -c "
import urllib.request, json, os
req = urllib.request.Request('https://api.mem0.ai/v1/memories/', headers={'Authorization': 'Token ' + os.environ['MEM0_API_KEY']})
resp = json.loads(urllib.request.urlopen(req, timeout=5).read())
print(len(resp.get('results',[])))
" 2>/dev/null)
      ok "Mem0 Cloud: connected ($COUNT memories)"
    else
      fail "Mem0 Cloud: API key set but unreachable"
    fi
  else
    warn "Mem0 Cloud: not configured (MEM0_API_KEY not set)"
    info "  Setup: https://app.mem0.ai -> API Keys -> add MEM0_API_KEY to brain.env"
  fi

  # Qdrant Cloud
  if [ -n "${QDRANT_URL:-}" ] && [ "${QDRANT_URL}" != "http://localhost:6333" ]; then
    QDRANT_API_KEY="${QDRANT_API_KEY:-}"
    if curl -sf "${QDRANT_URL}/collections" \
         ${QDRANT_API_KEY:+-H "api-key: $QDRANT_API_KEY"} >/dev/null 2>&1; then
      ok "Qdrant Cloud: connected ($QDRANT_URL)"
    else
      fail "Qdrant Cloud: URL set but unreachable"
    fi
  else
    warn "Qdrant Cloud: not configured (using localhost)"
    info "  Setup: https://cloud.qdrant.io -> free cluster -> add QDRANT_URL + QDRANT_API_KEY"
  fi

  # Upstash Vector
  if [ -n "${UPSTASH_VECTOR_URL:-}" ]; then
    if curl -sf "${UPSTASH_VECTOR_URL}/info" \
         -H "Authorization: Bearer ${UPSTASH_VECTOR_TOKEN:-}" >/dev/null 2>&1; then
      ok "Upstash Vector: connected"
    else
      fail "Upstash Vector: credentials set but unreachable"
    fi
  else
    warn "Upstash Vector: not configured"
    info "  Setup: https://console.upstash.com -> Vector -> add credentials to brain.env"
  fi

  # Git sync
  if [ "${BRAIN_MEMORY_GIT_SYNC:-false}" = "true" ]; then
    REMOTE="${BRAIN_MEMORY_GIT_REMOTE:-origin}"
    if git -C "$BRAIN_DIR" remote get-url "$REMOTE" >/dev/null 2>&1; then
      ok "Git memory sync: enabled (remote: $REMOTE)"
    else
      fail "Git memory sync: enabled but remote '$REMOTE' not found"
    fi
  else
    warn "Git memory sync: disabled"
    info "  Enable: set BRAIN_MEMORY_GIT_SYNC=true in brain.env"
  fi

  echo ""
}

# -- Push: local -> cloud ------------------------------------------------------
cmd_push() {
  echo ""
  info "Pushing memory to cloud..."
  PUSHED=0

  # Mem0: export MCP knowledge graph entities to Mem0 cloud
  if [ -n "${MEM0_API_KEY:-}" ]; then
    info "Syncing to Mem0..."
    python3 - "$BRAIN_DIR" << 'PY'
import json, os, urllib.request, pathlib, sys

brain_dir = pathlib.Path(sys.argv[1])
mem_json = brain_dir / "memory" / "brain-memories.json"

if not mem_json.exists():
    print("  [skip] No MCP memory JSON found at memory/brain-memories.json")
    sys.exit(0)

try:
    data = json.loads(mem_json.read_text())
    entities = data.get("entities", [])
    api_key = os.environ["MEM0_API_KEY"]
    user_id = os.environ.get("BRAIN_USER", "brain-repo")
    pushed = 0
    for ent in entities:
        text = f"[{ent.get('entityType','?')}] {ent.get('name','')}: " + " | ".join(ent.get("observations", []))
        payload = json.dumps({"messages": [{"role": "user", "content": text}], "user_id": user_id}).encode()
        req = urllib.request.Request(
            "https://api.mem0.ai/v1/memories/",
            data=payload,
            headers={"Authorization": f"Token {api_key}", "Content-Type": "application/json"},
            method="POST",
        )
        try:
            urllib.request.urlopen(req, timeout=10)
            pushed += 1
        except Exception:
            pass
    print(f"  [ok] Pushed {pushed}/{len(entities)} entities to Mem0")
except Exception as e:
    print(f"  [warn] Mem0 push failed: {e}")
PY
    PUSHED=1
  fi

  # Git sync: commit and push memory JSON
  if [ "${BRAIN_MEMORY_GIT_SYNC:-false}" = "true" ]; then
    info "Syncing memory via git..."
    REMOTE="${BRAIN_MEMORY_GIT_REMOTE:-origin}"
    MEM_FILES=("memory/brain-memories.json" "memory/manifest.json")
    CHANGED=0
    for f in "${MEM_FILES[@]}"; do
      if [ -f "$BRAIN_DIR/$f" ]; then
        git -C "$BRAIN_DIR" add "$f" 2>/dev/null && CHANGED=1
      fi
    done
    if [ "$CHANGED" -eq 1 ]; then
      git -C "$BRAIN_DIR" diff --cached --quiet 2>/dev/null || {
        git -C "$BRAIN_DIR" commit -m "brain: memory sync $(date -u '+%Y-%m-%dT%H:%M:%SZ')" --quiet
        git -C "$BRAIN_DIR" push "$REMOTE" HEAD --quiet 2>/dev/null \
          && ok "Memory pushed to git remote: $REMOTE" \
          || warn "Git push failed - check remote auth (SSH vs HTTPS)"
      }
    else
      ok "Git sync: nothing changed"
    fi
    PUSHED=1
  fi

  [ "$PUSHED" -eq 0 ] && warn "No cloud backend configured. Edit brain.env to enable cloud sync."
  echo ""
}

# -- Pull: cloud -> local (git only, Mem0 is read via MCP) -------------------
cmd_pull() {
  echo ""
  info "Pulling memory from cloud..."

  if [ "${BRAIN_MEMORY_GIT_SYNC:-false}" = "true" ]; then
    REMOTE="${BRAIN_MEMORY_GIT_REMOTE:-origin}"
    git -C "$BRAIN_DIR" fetch "$REMOTE" --quiet 2>/dev/null \
      && git -C "$BRAIN_DIR" checkout "$REMOTE/main" -- memory/brain-memories.json 2>/dev/null \
      && ok "Memory pulled from git remote: $REMOTE" \
      || warn "Git pull failed - check remote auth"
  else
    warn "Pull only works with BRAIN_MEMORY_GIT_SYNC=true"
    info "For Mem0/Qdrant: memory is read in real-time via MCP - no pull needed"
  fi
  echo ""
}

# -- Setup guide --------------------------------------------------------------
cmd_setup() {
  echo ""
  echo -e "${BOLD}Cloud Memory Setup Guide${RESET}"
  echo ""
  echo "  Pick ONE option and add to ~/.brain/brain.env:"
  echo ""
  echo -e "  ${BOLD}Option A: Mem0 (recommended - managed, semantic, free tier)${RESET}"
  echo "    1. Go to: https://app.mem0.ai"
  echo "    2. Settings -> API Keys -> copy your key"
  echo "    3. Add to brain.env:  MEM0_API_KEY=\"m0-...\""
  echo "    4. Add MCP server:    see mcp/registry.yml -> mem0 section"
  echo ""
  echo -e "  ${BOLD}Option B: Qdrant Cloud (self-service vector DB, free 1GB)${RESET}"
  echo "    1. Go to: https://cloud.qdrant.io"
  echo "    2. Create free cluster -> copy URL and API key"
  echo "    3. Add to brain.env:"
  echo "         QDRANT_URL=\"https://xxx.qdrant.io\""
  echo "         QDRANT_API_KEY=\"your-key\""
  echo "    4. Runs automatically on next: bash scripts/vector-sync-qdrant.sh"
  echo ""
  echo -e "  ${BOLD}Option C: Upstash Vector (serverless, pay-per-use, free 10k vectors)${RESET}"
  echo "    1. Go to: https://console.upstash.com -> Vector"
  echo "    2. Create index -> copy REST URL and token"
  echo "    3. Add to brain.env:"
  echo "         UPSTASH_VECTOR_URL=\"https://xxx.upstash.io\""
  echo "         UPSTASH_VECTOR_TOKEN=\"your-token\""
  echo ""
  echo -e "  ${BOLD}Option D: Git-backed sync (free, zero external service)${RESET}"
  echo "    1. Add to brain.env:  BRAIN_MEMORY_GIT_SYNC=\"true\""
  echo "    2. Memory JSON commits with your brain repo on /handover"
  echo "    3. Works across devices as long as git remote is accessible"
  echo ""
  echo "  After setup, run: bash ~/.brain/scripts/memory-cloud-sync.sh status"
  echo ""
}

case "$COMMAND" in
  status) cmd_status ;;
  push)   cmd_push   ;;
  pull)   cmd_pull   ;;
  setup)  cmd_setup  ;;
  *)
    echo "Usage: bash memory-cloud-sync.sh [status|push|pull|setup]"
    exit 1
    ;;
esac
