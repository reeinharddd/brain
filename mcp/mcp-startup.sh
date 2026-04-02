#!/bin/bash
# MCP Server Startup Script with Port Management
# Solves connection conflicts by using dedicated ports

set -e

# Source port allocation
source_ports() {
    MEMORY_PORT=8001
    FILESYSTEM_PORT=8002
    CONTEXT7_PORT=8003
    SEQUENTIAL_THINKING_PORT=8004
    GITHUB_PORT=8005
}

# Kill existing MCP processes
cleanup_mcp() {
    echo "Cleaning up existing MCP processes..."
    pkill -f "supergateway" 2>/dev/null || true
    sleep 2
}

# Start MCP server with dedicated port
start_mcp_server() {
    local server_name=$1
    local server_cmd=$2
    local port=$3
    
    echo "Starting $server_name on port $port..."
    nohup bunx --bun supergateway \
        --port $port \
        --stdio "$server_cmd" \
        > /tmp/mcp-${server_name}.log 2>&1 &
    
    echo "$server_name PID: $!"
    sleep 1
}

# Main startup sequence
main() {
    source_ports
    cleanup_mcp
    
    echo "Starting MCP servers with dedicated ports..."
    
    # Start core servers
    start_mcp_server "memory" "bunx --bun @modelcontextprotocol/server-memory \"$HOME/.brain/memory\"" $MEMORY_PORT
    start_mcp_server "filesystem" "bunx --bun @modelcontextprotocol/server-filesystem \"$HOME\"" $FILESYSTEM_PORT
    start_mcp_server "context7" "bunx --bun @upstash/context7-mcp@latest" $CONTEXT7_PORT
    start_mcp_server "sequential-thinking" "bunx --bun @modelcontextprotocol/server-sequential-thinking" $SEQUENTIAL_THINKING_PORT

    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        start_mcp_server "github" "bunx --bun @modelcontextprotocol/server-github" $GITHUB_PORT
    else
        echo "Skipping github on port $GITHUB_PORT (GITHUB_TOKEN not set)"
    fi
    
    echo "MCP servers started successfully!"
    echo "Check logs: tail -f /tmp/mcp-*.log"
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
