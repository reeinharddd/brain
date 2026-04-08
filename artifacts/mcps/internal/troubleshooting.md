# MCP Troubleshooting Guide

## Common Issues & Solutions

### 1. "Already connected to a transport" Error
**Cause**: Multiple MCP servers using same port (8000)  
**Solution**: Use dedicated ports via `mcp-startup.sh`

### 2. Frequent Server Restarts
**Cause**: Port conflicts and connection state issues  
**Solution**: Clean restart with port allocation

### 3. SSE Connection Failures
**Cause**: Gateway process conflicts  
**Solution**: Kill all supergateway processes before restart

## Recovery Commands

```bash
# Emergency cleanup
pkill -f "supergateway"
pkill -f "mcp-"
lsof -i :8000 | grep -v PID | awk '{print $2}' | xargs kill

# Clean restart
~/.brain/mcp/mcp-startup.sh

# Check status
ps aux | grep -E "(supergateway|mcp-)" | grep -v grep
lsof -i :8001-8005
```

## Port Allocation

| Server | Port | Purpose |
|---------|------|---------|
| memory | 8001 | Knowledge graph storage |
| filesystem | 8002 | File operations |
| context7 | 8003 | Documentation retrieval |
| sequential-thinking | 8004 | Structured reasoning |
| github | 8005 | GitHub API |

## Log Locations

- Memory: `/tmp/mcp-memory.log`
- Filesystem: `/tmp/mcp-filesystem.log`
- Context7: `/tmp/mcp-context7.log`
- Sequential Thinking: `/tmp/mcp-sequential-thinking.log`

## Verification

```bash
# Test connections
curl -s http://localhost:8001/sse | head -1
curl -s http://localhost:8002/sse | head -1
curl -s http://localhost:8003/sse | head -1
curl -s http://localhost:8004/sse | head -1
```
