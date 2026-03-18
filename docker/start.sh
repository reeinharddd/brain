#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/docker/start.sh — MCP Stack manager
#  Usage:
#    ~/.brain/docker/start.sh up         # start standard stack
#    ~/.brain/docker/start.sh up --github  # + GitHub MCP
#    ~/.brain/docker/start.sh down       # stop all
#    ~/.brain/docker/start.sh status     # show running MCPs
#    ~/.brain/docker/start.sh logs [svc] # view logs
# ═══════════════════════════════════════════════════════════

set -euo pipefail

DOCKER_DIR="$HOME/.brain/docker"
ENV_FILE="$DOCKER_DIR/.env"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'
BOLD='\033[1m'; RESET='\033[0m'

ok()   { echo -e "  ${GREEN}✓${RESET} $1"; }
warn() { echo -e "  ${YELLOW}⚠${RESET}  $1"; }
fail() { echo -e "  ${RED}✗${RESET} $1"; }

# ── Ensure .env exists ───────────────────────────────────────
ensure_env() {
  if [ ! -f "$ENV_FILE" ]; then
    cp "$DOCKER_DIR/.env.example" "$ENV_FILE"
    warn ".env created from template"
  fi
  # Always ensure HOST_HOME is set to the real value (not the placeholder)
  if grep -q "your-username\|HOST_HOME=$" "$ENV_FILE" 2>/dev/null; then
    sed -i "s|HOST_HOME=.*|HOST_HOME=$HOME|" "$ENV_FILE"
    ok "HOST_HOME set to $HOME"
  fi
}

# ── Commands ─────────────────────────────────────────────────
cmd_up() {
  local github_profile=""
  [ "${1:-}" = "--github" ] && github_profile="--profile github"

  ensure_env

  echo -e "\n${BOLD}── Starting MCP Stack${RESET}"
  docker compose -f "$DOCKER_DIR/docker-compose.yml" --env-file "$ENV_FILE" \
    $github_profile up -d --pull missing

  echo ""
  echo -e "  ${BOLD}MCP endpoints:${RESET}"
  echo "    memory:   http://localhost:3001/sse"
  echo "    fs:       http://localhost:3002/sse"
  echo "    sequential: http://localhost:3003/sse"
  echo "    github:     http://localhost:3004/sse"
  echo "    context7:   http://localhost:3005/sse"
  echo "    ninja:      http://localhost:3006/sse"
  echo "    duckgo:     http://localhost:3007/sse"
  echo "    crawl4ai:   http://localhost:3008/sse"
  echo "    awesome:    http://localhost:3009/sse"
  echo ""
  echo -e "  ${BOLD}Next:${RESET} switch your Claude Code settings to Docker mode:"
  echo "    ln -sf ~/.brain/adapters/claude-code/settings.docker.json ~/.claude/settings.json"
  echo ""

  ok "MCP stack started"
}

cmd_down() {
  echo -e "\n${BOLD}── Stopping MCP Stack${RESET}"
  docker compose -f "$DOCKER_DIR/docker-compose.yml" --env-file "$ENV_FILE" down
  ok "Stack stopped (memory data preserved in Docker volume)"
}

cmd_status() {
  echo -e "\n${BOLD}── MCP Stack Status${RESET}\n"
  docker compose -f "$DOCKER_DIR/docker-compose.yml" --env-file "$ENV_FILE" ps 2>/dev/null || {
    warn "Stack is not running. Start with: ~/.brain/docker/start.sh up"
    return
  }

  echo ""
  echo -e "  ${BOLD}Health check:${RESET}"
  local ports=(3001 3002 3003 3004 3005 3006 3007 3008 3009)
  local names=("memory" "filesystem" "sequential" "github" "context7" "ninja" "duckgo" "crawl4ai" "awesome")
  for i in "${!ports[@]}"; do
    local port="${ports[$i]}"
    local name="${names[$i]}"
    if curl -sf "http://localhost:$port/sse" &>/dev/null; then
      ok "$name (port $port)"
    else
      warn "$name (port $port) — not responding yet"
    fi
  done
}

cmd_logs() {
  local svc="${1:-}"
  docker compose -f "$DOCKER_DIR/docker-compose.yml" --env-file "$ENV_FILE" \
    logs --follow --tail=50 $svc
}

cmd_reset_memory() {
  echo -e "\n${BOLD}── Reset MCP Memory${RESET}"
  warn "This will permanently delete all stored memory."
  read -r -p "  Are you sure? (y/N): " confirm
  [ "${confirm:-N}" != "y" ] && { echo "  Cancelled."; exit 0; }

  docker compose -f "$DOCKER_DIR/docker-compose.yml" --env-file "$ENV_FILE" \
    down mcp-memory 2>/dev/null || true
  docker volume rm brain-mcp-memory-data 2>/dev/null || true
  ok "Memory volume removed. Start again with: start.sh up"
}

# ── Main ─────────────────────────────────────────────────────
case "${1:-help}" in
  up)      shift; cmd_up "$@" ;;
  down)    cmd_down ;;
  status)  cmd_status ;;
  logs)    shift; cmd_logs "${1:-}" ;;
  reset-memory) cmd_reset_memory ;;
  *)
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  up [--github]   Start MCP stack (add --github to include GitHub MCP)"
    echo "  down            Stop all MCP services"
    echo "  status          Show running status and health"
    echo "  logs [service]  Tail logs (service: mcp-memory, mcp-filesystem, etc.)"
    echo "  reset-memory    Delete persistent memory volume (destructive!)"
    ;;
esac
