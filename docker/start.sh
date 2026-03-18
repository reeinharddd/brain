#!/bin/bash
# ===========================================================
#  brain/docker/start.sh - Persistent helper services manager
#  Usage:
#    ~/.brain/docker/start.sh up         # start persistent helper services
#    ~/.brain/docker/start.sh down       # stop all
#    ~/.brain/docker/start.sh status     # show running helper services
#    ~/.brain/docker/start.sh logs [svc] # view logs
# ===========================================================

set -euo pipefail

DOCKER_DIR="$HOME/.brain/docker"
ENV_FILE="$DOCKER_DIR/.env"
VECTOR_CONFIG="$HOME/.brain/memory/vector-config.json"
QDRANT_CONTAINER="brain-qdrant"
QDRANT_VOLUME="brain-qdrant-data"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'
BOLD='\033[1m'; RESET='\033[0m'

ok()   { echo -e "  ${GREEN}[ok]${RESET} $1"; }
warn() { echo -e "  ${YELLOW}[warn]${RESET}  $1"; }
fail() { echo -e "  ${RED}[fail]${RESET} $1"; }

vector_enabled() {
  [ -f "$VECTOR_CONFIG" ] || return 1
  python3 - "$VECTOR_CONFIG" <<'PY'
import json, sys
config = json.load(open(sys.argv[1], "r", encoding="utf-8"))
raise SystemExit(0 if config.get("enabled") else 1)
PY
}

# -- Ensure .env exists -------------------------------------------
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

# -- Commands -----------------------------------------------------
cmd_up() {
  ensure_env

  echo -e "\n${BOLD}-- Starting Persistent Helpers${RESET}"
  docker volume create "$QDRANT_VOLUME" >/dev/null
  if docker ps --format '{{.Names}}' | grep -qx "$QDRANT_CONTAINER"; then
    ok "qdrant already running"
  elif docker ps -a --format '{{.Names}}' | grep -qx "$QDRANT_CONTAINER"; then
    docker start "$QDRANT_CONTAINER" >/dev/null
    ok "qdrant container started"
  else
    docker run -d \
      --name "$QDRANT_CONTAINER" \
      --restart unless-stopped \
      -p 6333:6333 \
      -v "$QDRANT_VOLUME:/qdrant/storage" \
      qdrant/qdrant:latest >/dev/null
    ok "qdrant container created"
  fi

  if vector_enabled; then
    echo ""
    ok "Vector backend is enabled and managed by docker-compose.yml"
  fi

  echo ""
  echo -e "  ${BOLD}Persistent helpers:${RESET}"
  echo "    qdrant: http://localhost:6333"
  echo ""
  echo -e "  ${BOLD}Next:${RESET} use hybrid persistent settings:"
  echo "    ln -sf ~/.brain/adapters/claude-code/settings.persistent.json ~/.claude/settings.json"
  echo ""

  ok "Persistent helpers started"
}

cmd_down() {
  echo -e "\n${BOLD}-- Stopping Persistent Helpers${RESET}"
  if docker ps --format '{{.Names}}' | grep -qx "$QDRANT_CONTAINER"; then
    docker stop "$QDRANT_CONTAINER" >/dev/null
    ok "qdrant stopped"
  else
    warn "qdrant is not running"
  fi
}

cmd_status() {
  echo -e "\n${BOLD}-- Persistent Helpers Status${RESET}\n"
  docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -E 'NAMES|brain-qdrant' || true

  echo ""
  echo -e "  ${BOLD}Health check:${RESET}"
  if vector_enabled; then
    if curl -sf --max-time 5 --retry 5 --retry-delay 1 "http://localhost:6333/collections" >/dev/null 2>&1; then
      ok "qdrant (port 6333)"
    else
      warn "qdrant (port 6333) - vector enabled in config, but service is not responding"
    fi
  fi

  echo ""
  echo "  Core MCPs are expected to run via stdio or docker-on-demand from adapter settings."
}

cmd_logs() {
  local svc="${1:-}"
  local target="${svc:-$QDRANT_CONTAINER}"
  docker logs --follow --tail=50 "$target"
}

cmd_reset_memory() {
  echo -e "\n${BOLD}-- Reset MCP Memory${RESET}"
  warn "This will not reset stdio memory backends in helper-only mode."
  read -r -p "  Are you sure? (y/N): " confirm
  [ "${confirm:-N}" != "y" ] && { echo "  Cancelled."; exit 0; }

  warn "No docker-managed memory volume is configured in helper-only mode."
}

# -- Main ---------------------------------------------------------
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
    echo "  up              Start persistent helper services"
    echo "  down            Stop helper services"
    echo "  status          Show running status and health"
    echo "  logs [service]  Tail logs (service: qdrant)"
    echo "  reset-memory    Not used in helper-only mode"
    ;;
esac
