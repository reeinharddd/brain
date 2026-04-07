---
type: integration-guide
id: mcp-ide-integration
title: MCP Docs-RAG Integration for IDEs and LLMs
status: active
version: 1.0
date_created: 2026-04-04
language: en
category: architecture
---

## MCP Docs-RAG Integration Guide

## Status: ✅ OPERATIONAL

The docs-rag-mcp server is fully functional and ready for real-world testing through IDEs, CLI tools, and LLM interfaces.

## Quick Start

### Test the MCP Directly (No IDE)

```bash
# Test 1: Check index status
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"docs_status"}}' | \
  ~/.brain/bin/docs-rag-mcp

# Test 2: Search for documents
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"docs_search","params":{"query":"daemon","limit":5}}}' | \
  ~/.brain/bin/docs-rag-mcp

# Test 3: Rebuild index (dev-only)
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"docs_rebuild","params":{}}}' | \
  ~/.brain/bin/docs-rag-mcp
```

## Integration Paths

### 1. Claude Desktop (Easiest)

**Location**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "brain-docs": {
      "command": "~/.brain/bin/docs-rag-mcp",
      "args": [],
      "env": {
        "BRAIN_ENV": "development",
        "BRAIN_ROOT": "/home/reeinharrrd/.brain"
      }
    }
  }
}
```

**Then restart Claude Desktop.** The `docs_search`, `docs_status`, and `docs_rebuild` tools will be available.

### 2. VSCode + GitHub Copilot

**MCP configuration location:**  
`~/.config/Code/User/globalStorage/GitHub.copilot/claude-dev/mcp/global-config.json`

```json
{
  "mcpServers": {
    "brain-docs": {
      "command": "~/.brain/bin/docs-rag-mcp",
      "args": [],
      "env": {
        "BRAIN_ENV": "development"
      }
    }
  }
}
```

**Then reload VS Code extension (Copilot Chat).**

### 3. Cline (VSCode Extension)

**Create or edit**: `.cline/mcp.json` in workspace root

```json
{
  "mcpServers": {
    "brain-docs": {
      "command": "~/.brain/bin/docs-rag-mcp",
      "args": [],
      "env": {
        "BRAIN_ENV": "development"
      }
    }
  }
}
```

### 4. Custom Tools / Scripts

Any tool that implements the MCP protocol (JSON-RPC 2.0 over stdio) can use the MCP:

```bash
#!/bin/bash
# Simple wrapper script
QUERY="${1:-daemon}"
echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"docs_search\",\"params\":{\"query\":\"$QUERY\",\"limit\":5}}}" | \
  ~/.brain/bin/docs-rag-mcp | jq .result.results
```

## Available Tools

### docs_search

Search documentation by query string (semantic embeddings).

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "docs_search",
    "params": {
      "query": "string (required)",
      "limit": "number (optional, default 5)",
      "domain": "string (optional: adr, architecture, skills, testing)"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "results": [
      {
        "document_id": "string",
        "title": "string",
        "path": "string",
        "score": 0.95,
        "rag_priority": "high",
        "snippet": "string (excerpt)"
      }
    ],
    "metadata": {
      "total_indexed": 2,
      "query_time_ms": 15,
      "index_status": "ready",
      "results_count": 5
    }
  }
}
```

### docs_status

Get current index and server status.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "docs_status"
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "index_status": {
      "state": "ready",
      "document_count": 2,
      "last_rebuild_time": "2026-04-04T09:57:31Z",
      "index_size_bytes": 12345,
      "qdrant_health": "healthy",
      "errors": []
    }
  }
}
```

### docs_rebuild

Rebuild the documentation index (development only).

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "docs_rebuild",
    "params": {
      "domains": ["adr", "architecture"]
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "success": true,
    "message": "Index rebuilt",
    "document_count": 2,
    "chunk_count": 14,
    "duration_ms": 3456
  }
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│ IDE / LLM / Tool (Claude Desktop, VSCode, Custom Tool) │
└──────────────────────┬──────────────────────────────────┘
                       │ JSON-RPC 2.0
                       │ (stdio)
                       ▼
┌─────────────────────────────────────────────────────────┐
│ docs-rag-mcp (MCP Server)                               │
│ - Handles JSON-RPC protocol                             │
│ - Routes to tool handlers                               │
│ - Manages index lifecycle                               │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
        ┌──────────────────────────┐
        │  Index Manager           │
        │  - Lazy-load             │
        │  - Cache results         │
        │  - Status tracking       │
        └──────────────────────────┘
                       │
         ┌─────────────┼─────────────┐
         ▼             ▼             ▼
    ┌────────┐   ┌────────┐   ┌────────┐
    │ Indexer│   │Manifest│   │Qdrant  │
    │(build) │   │(schema)│   │(search)│
    └────────┘   └────────┘   └────────┘
         │             │             │
         └─────────────┼─────────────┘
                       ▼
         /home/user/.brain/docs/
         (documentation source)
```

## Current Development State

### ✅ Complete
- MCP server binary: `~/.brain/bin/docs-rag-mcp` (3.8M)
- JSON-RPC 2.0 protocol implementation
- Tools: docs_search, docs_status, docs_rebuild
- Manifest parsing (JSON, with correct schema)
- Index initialization (lazy-load on first search)
- Error handling and logging
- BRAIN_ENV detection (dev/prod safety)

### 📋 Manifest/Content Issues (Non-Blocking)
- 98 of 100 documents lack YAML frontmatter
- Only 2 documents fully valid for indexing
- Expected for templates (don't need frontmatter)
- Fixable: Add frontmatter to core docs/* files

### 🎯 Testing Ready For
- ✅ Claude Desktop (direct stdio)
- ✅ VSCode + Copilot (MCP protocol)
- ✅ Custom scripts / tools (JSON-RPC)
- ✅ Command-line testing (stdio)

## Troubleshooting

### MCP not responding
```bash
# Test directly
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"docs_status"}}' | \
  ~/.brain/bin/docs-rag-mcp

# Should return status within 2 seconds
```

### Index shows "not-initialized"
```bash
# Trigger index rebuild
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"docs_rebuild"}}' | \
  ~/.brain/bin/docs-rag-mcp
```

### Documents not indexed
- Check: `~/brain/docs-manifest.json` exists ✓
- Check: `~/.brain/docs/` has markdown files ✓
- Check: Files have valid YAML frontmatter (optional for templates)

### "command not found: docs-rag-mcp"
```bash
# Ensure binary exists and is executable
ls -lh ~/.brain/bin/docs-rag-mcp
# If not, rebuild:
cd ~/.brain/mcp/docs-rag-mcp && go build -o ~/.brain/bin/docs-rag-mcp .
```

## Files Involved

| File | Purpose | Status |
|------|---------|--------|
| `mcp/docs-rag-mcp/main.go` | MCP server entrypoint | ✅ |
| `mcp/docs-rag-mcp/internal/indexer/` | Index & search logic | ✅ |
| `mcp/docs-rag-mcp/internal/indexer/loader.go` | Manifest & doc loading | ✅ Fixed |
| `mcp/docs-rag-mcp/internal/indexer/types.go` | Data structures | ✅ Fixed |
| `docs-manifest.json` | Documentation schema | ✅ |
| `~/.brain/bin/docs-rag-mcp` | Compiled binary | ✅ |

## Next Steps

1. **Test in Claude Desktop**
   - Add config (see "Claude Desktop" section)
   - Ask Claude about docs (e.g., "Search for daemon architecture")
   
2. **Test in VSCode**
   - Configure MCP in global-config.json
   - Use Copilot Chat to access docs_search

3. **Measure Performance**
   - First search: ~2-5 seconds (index builds)
   - Subsequent searches: <200ms (cached)

4. **Add Frontmatter** (future)
   - Make all docs indexable with full metadata
   - Enables domain filtering and priority ranking

## Version History

- 2026-04-04: Phase 5 Complete - MCP operational, JSON parsing fixed, ready for IDE testing
- 2026-04-03: Phase 5 handler integration, HTTP endpoints added

