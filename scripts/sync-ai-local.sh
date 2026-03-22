#!/bin/bash
# brain/scripts/sync-ai-local.sh
# Synchronizes configuration between .brain and ai-local
# Usage: bash ~/.brain/scripts/sync-ai-local.sh [--check|--sync]

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
AI_LOCAL_DIR="${AI_LOCAL_DIR:-$HOME/ai-local}"
MODE="${1:-check}"

# Colors (ASCII only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Validate directories exist
if [ ! -d "$BRAIN_DIR" ]; then
    log_error "Brain directory not found: $BRAIN_DIR"
    exit 1
fi

if [ ! -d "$AI_LOCAL_DIR" ]; then
    log_error "AI-Local directory not found: $AI_LOCAL_DIR"
    log_info "Clone or create ai-local first"
    exit 1
fi

check_sync_status() {
    local status=0

    echo ""
    echo "=== Sync Status Check ==="
    echo ""

    # Check if ai-local has .env
    if [ -f "$AI_LOCAL_DIR/.env" ]; then
        log_info ".env file exists in ai-local"
    else
        log_warn ".env file missing in ai-local (copy from .env.example)"
        status=1
    fi

    # Check if Qdrant config matches
    if [ -f "$BRAIN_DIR/memory/vector-config.json" ]; then
        local qdrant_url
        qdrant_url=$(grep -o '"url": "[^"]*"' "$BRAIN_DIR/memory/vector-config.json" | cut -d'"' -f4)
        if [[ "$qdrant_url" == *"localhost:6333"* ]]; then
            log_info "Qdrant config points to localhost:6333"
        else
            log_warn "Qdrant config may not match ai-local"
            status=1
        fi
    fi

    # Check if MCP ports are consistent
    local mcp_ports=(3001 3002 3003)
    for port in "${mcp_ports[@]}"; do
        if grep -q "127.0.0.1:$port" "$AI_LOCAL_DIR/docker-compose.yml" 2>/dev/null; then
            log_info "MCP port $port bound to localhost"
        else
            log_warn "MCP port $port may be exposed to network"
            status=1
        fi
    done

    # Check if ai-local workspace is configured
    if [ -d "$AI_LOCAL_DIR/workspace" ]; then
        log_info "Workspace directory exists"
    else
        log_warn "Workspace directory missing"
        status=1
    fi

    # Check if start-brain.sh is executable
    if [ -x "$AI_LOCAL_DIR/start-brain.sh" ]; then
        log_info "start-brain.sh is executable"
    else
        log_warn "start-brain.sh is not executable"
        status=1
    fi

    echo ""
    if [ $status -eq 0 ]; then
        log_info "All checks passed - systems are synchronized"
    else
        log_warn "Some checks failed - run with --sync to fix"
    fi

    return $status
}

sync_configuration() {
    echo ""
    echo "=== Syncing Configuration ==="
    echo ""

    # Ensure workspace directory exists
    if [ ! -d "$AI_LOCAL_DIR/workspace" ]; then
        mkdir -p "$AI_LOCAL_DIR/workspace"
        log_info "Created workspace directory"
    fi

    # Ensure .env exists
    if [ ! -f "$AI_LOCAL_DIR/.env" ]; then
        if [ -f "$AI_LOCAL_DIR/.env.example" ]; then
            cp "$AI_LOCAL_DIR/.env.example" "$AI_LOCAL_DIR/.env"
            log_info "Created .env from .env.example"
            log_warn "Please edit $AI_LOCAL_DIR/.env with your actual values"
        else
            log_error ".env.example not found"
            exit 1
        fi
    fi

    # Update .env with current user info
    if [ -w "$AI_LOCAL_DIR/.env" ]; then
        # Add UID/GID if not present
        if ! grep -q "^HOST_UID=" "$AI_LOCAL_DIR/.env"; then
            echo "" >> "$AI_LOCAL_DIR/.env"
            echo "HOST_UID=$(id -u)" >> "$AI_LOCAL_DIR/.env"
            echo "HOST_GID=$(id -g)" >> "$AI_LOCAL_DIR/.env"
            log_info "Added HOST_UID/HOST_GID to .env"
        fi

        # Add WORKSPACE_DIR if not present
        if ! grep -q "^WORKSPACE_DIR=" "$AI_LOCAL_DIR/.env"; then
            echo "WORKSPACE_DIR=./workspace" >> "$AI_LOCAL_DIR/.env"
            log_info "Added WORKSPACE_DIR to .env"
        fi
    fi

    # Ensure start-brain.sh is executable
    if [ -f "$AI_LOCAL_DIR/start-brain.sh" ]; then
        chmod +x "$AI_LOCAL_DIR/start-brain.sh"
        log_info "Made start-brain.sh executable"
    fi

    log_info "Sync complete"
}

show_integration_help() {
    echo ""
    echo "=== Integration Guide ==="
    echo ""
    echo "To use ai-local with .brain:"
    echo ""
    echo "1. Start ai-local stack:"
    echo "   cd ~/ai-local && bash start-brain.sh"
    echo ""
    echo "2. Verify services are running:"
    echo "   docker compose ps"
    echo ""
    echo "3. Access services:"
    echo "   - OpenWebUI: http://localhost:3000"
    echo "   - Qdrant:    http://localhost:6333"
    echo "   - Ollama:    http://localhost:11435"
    echo ""
    echo "4. Use with Claude Code:"
    echo "   claude --model ollama/qwen2.5-coder:7b"
    echo ""
    echo "5. MCP endpoints (for debugging):"
    echo "   - Memory:     http://localhost:3001"
    echo "   - Filesystem: http://localhost:3002"
    echo "   - Sequential: http://localhost:3003"
    echo ""
}

# Main
case "$MODE" in
    --check|check)
        check_sync_status
        ;;
    --sync|sync)
        sync_configuration
        check_sync_status
        ;;
    --help|help)
        echo "Usage: $0 [--check|--sync|--help]"
        echo ""
        echo "  --check  Check if .brain and ai-local are synchronized (default)"
        echo "  --sync   Fix synchronization issues"
        echo "  --help   Show this help"
        echo ""
        show_integration_help
        ;;
    *)
        log_error "Unknown mode: $MODE"
        echo "Usage: $0 [--check|--sync|--help]"
        exit 1
        ;;
esac
