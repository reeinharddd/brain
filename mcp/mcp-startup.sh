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
    pkill -f "mcp-" 2>/dev/null || true
    sleep 2
}

# Start MCP server with dedicated port
start_mcp_server() {
    local server_name=$1
    local server_cmd=$2
    local port=$3
    
    echo "Starting $server_name on port $port..."
    nohup npx -y supergateway \
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
    start_mcp_server "memory" "@modelcontextprotocol/server-memory" $MEMORY_PORT
    start_mcp_server "filesystem" "@modelcontextprotocol/server-filesystem /workspace" $FILESYSTEM_PORT
    start_mcp_server "context7" "@upstash/context7-mcp@latest" $CONTEXT7_PORT
    start_mcp_server "sequential-thinking" "@modelcontextprotocol/server-sequential-thinking" $SEQUENTIAL_THINKING_PORT
    
    echo "MCP servers started successfully!"
    echo "Check logs: tail -f /tmp/mcp-*.log"
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
