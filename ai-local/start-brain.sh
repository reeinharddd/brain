#!/bin/bash
# Brain Station Unified Startup Script
# Usage: bash start-brain.sh [up|down|status|logs]

set -euo pipefail

BASE_DIR="/home/reeinharrrd/ai-local"
SOCAT_PID_FILE="/tmp/socat_brain.pid"
SOCAT_LOG="/tmp/socat_brain.log"

# Colors for output (ASCII only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if .env exists
if [ ! -f "$BASE_DIR/.env" ]; then
    log_warn ".env file not found. Copying from .env.example..."
    cp "$BASE_DIR/.env.example" "$BASE_DIR/.env"
    log_warn "Please edit .env with your actual values before running again."
    exit 1
fi

# Load environment variables
export $(grep -v '^#' "$BASE_DIR/.env" | xargs) 2>/dev/null || true

# Ensure workspace directory exists
if [ -n "${WORKSPACE_DIR:-}" ]; then
    mkdir -p "$WORKSPACE_DIR"
fi

start_bridge() {
    # Check if socat is already running
    if [ -f "$SOCAT_PID_FILE" ] && kill -0 "$(cat "$SOCAT_PID_FILE")" 2>/dev/null; then
        log_info "Ollama Bridge already active (PID: $(cat "$SOCAT_PID_FILE"))"
        return 0
    fi

    # Check if socat is available
    if ! command -v socat &> /dev/null; then
        log_error "socat is not installed. Install with: sudo apt-get install socat"
        exit 1
    fi

    log_info "Starting Ollama Bridge (Port 11435 -> 11434)..."

    # Start socat with proper process management
    socat TCP-LISTEN:11435,fork,reuseaddr,bind=127.0.0.1 TCP:127.0.0.1:11434 > "$SOCAT_LOG" 2>&1 &
    SOCAT_PID=$!
    echo $SOCAT_PID > "$SOCAT_PID_FILE"

    # Wait a moment and verify it's running
    sleep 1
    if kill -0 "$SOCAT_PID" 2>/dev/null; then
        log_info "Ollama Bridge started successfully (PID: $SOCAT_PID)"
    else
        log_error "Failed to start Ollama Bridge. Check $SOCAT_LOG"
        exit 1
    fi
}

stop_bridge() {
    if [ -f "$SOCAT_PID_FILE" ]; then
        PID=$(cat "$SOCAT_PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID" 2>/dev/null || true
            log_info "Ollama Bridge stopped"
        fi
        rm -f "$SOCAT_PID_FILE"
    fi
}

start_services() {
    cd "$BASE_DIR" || exit 1

    # Check if Docker is running
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi

    # Create workspace if it doesn't exist
    mkdir -p "${WORKSPACE_DIR:-./workspace}"

    log_info "Starting Docker Services..."
    docker compose up -d

    log_info "Waiting for services to be healthy..."
    sleep 5

    # Check service health
    SERVICES=("open-webui" "brain-qdrant" "brain-mcp-memory" "brain-mcp-filesystem" "brain-mcp-sequential")
    for service in "${SERVICES[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${service}$"; then
            log_info "Service $service is running"
        else
            log_warn "Service $service may not be running. Check with: docker logs $service"
        fi
    done
}

stop_services() {
    cd "$BASE_DIR" || exit 1
    log_info "Stopping Docker Services..."
    docker compose down
    stop_bridge
}

show_status() {
    echo ""
    echo "=== Brain Station Status ==="
    echo ""

    # Check Ollama Bridge
    if [ -f "$SOCAT_PID_FILE" ] && kill -0 "$(cat "$SOCAT_PID_FILE")" 2>/dev/null; then
        echo "Ollama Bridge: RUNNING (PID: $(cat "$SOCAT_PID_FILE"))"
    else
        echo "Ollama Bridge: STOPPED"
    fi

    # Check Docker services
    echo ""
    echo "Docker Services:"
    docker compose ps 2>/dev/null || echo "  Not running"

    echo ""
    echo "=== Access Points ==="
    echo "Web UI:        http://localhost:3000"
    echo "Ollama API:    http://localhost:11435/v1"
    echo "Qdrant:        http://localhost:6333"
    echo "Memory MCP:    http://localhost:3001"
    echo "Filesystem MCP: http://localhost:3002"
    echo "Sequential MCP: http://localhost:3003"
}

show_logs() {
    cd "$BASE_DIR" || exit 1
    docker compose logs -f
}

# Main command handler
case "${1:-up}" in
    up|start)
        start_bridge
        start_services
        echo ""
        log_info "Brain Station is UP!"
        show_status
        ;;
    down|stop)
        stop_services
        log_info "Brain Station stopped"
        ;;
    restart)
        stop_services
        sleep 2
        start_bridge
        start_services
        log_info "Brain Station restarted"
        show_status
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    *)
        echo "Usage: $0 {up|down|restart|status|logs}"
        exit 1
        ;;
esac
