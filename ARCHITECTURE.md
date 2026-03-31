# Brain Repo Unified Architecture

> Single entry point, portable, consistent AI development environment.

## Overview

The Brain Repo has been re-architected to solve critical fragmentation issues:

1. **Single Entry Point**: `brain` CLI orchestrates everything
2. **Unified MCP Management**: Gateway pattern handles npm-only and Docker MCPs
3. **LLM Strategy**: Ollama primary, Docker Model Runner optional
4. **Zero Hardcoded Paths**: All configuration via environment variables
5. **IDE Agnostic**: Same behavior across Cursor, Windsurf, Claude Code, VS Code

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      brain CLI                               │
│         (single entry point for all operations)            │
└──────────────────────┬────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
   ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
   │  up     │   │  doctor │   │  mcp    │
   │  down   │   │  audit  │   │  llm    │
   │  status │   │  sync   │   │  logs   │
   └────┬────┘   └────┬────┘   └────┬────┘
        │              │              │
        └──────────────┼──────────────┘
                       │
┌──────────────────────▼────────────────────────────────────┐
│                 Docker Compose Stack                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │
│  │MCP Gateway  │  │   Qdrant    │  │  Open WebUI     │  │
│  │  (port 3000)│  │ (port 6333) │  │  (port 8080)    │  │
│  └──────┬──────┘  └─────────────┘  └─────────────────┘  │
│         │                                                │
│  ┌──────▼──────────────────────────────────────────┐  │
│  │  Optional: Docker Model Runner (profile)         │  │
│  │  (port 12434)                                    │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
   ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
   │  npm    │   │  docker │   │  local  │
   │  mcps   │   │  mcps   │   │  mcps   │
   │(gateway)│   │(gateway)│   │(stdio)  │
   └─────────┘   └─────────┘   └─────────┘
```

## Components

### 1. Brain CLI (`bin/brain`)

Unified command interface:

```bash
brain up              # Start all services
brain down            # Stop all services
brain status          # Show system status
brain doctor          # Run health checks
brain sync            # Sync MCP configs to IDEs
brain mcp list        # List available MCPs
brain mcp add <name>  # Start an MCP
brain llm list        # List LLM models
brain logs [service]  # View logs
```

### 2. MCP Gateway (`mcp/gateway/`)

Solves the npm-only MCP problem (like skill-ninja):

- **Container-based**: Runs in Docker with access to host Docker socket
- **Dynamic spawning**: Can start any MCP from npm or Docker
- **Unified interface**: All MCPs exposed via HTTP/SSE on port 3000
- **Health monitoring**: Auto-restart failed MCPs

Registry format (`mcp-registry.json`):
```json
{
  "mcps": {
    "skill-ninja": {
      "type": "npm",
      "package": "aktsmm/skill-ninja-mcp-server",
      "required": false
    },
    "duckduckgo": {
      "type": "docker",
      "image": "mcp/duckduckgo:latest"
    }
  }
}
```

### 3. LLM Strategy (`bin/brain-llm`)

Unified LLM management:

- **Primary**: Ollama (recommended, most models, mature)
- **Secondary**: Docker Model Runner (optional, optimized models)
- **Auto-routing**: Uses best available for requested model

Configuration via environment:
```bash
BRAIN_LLM_PRIMARY=ollama
BRAIN_LLM_FALLBACK=docker-model-runner
OLLAMA_HOST=host.docker.internal:11434
```

### 4. Portable Configuration

All paths use environment variables:

```bash
BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"          # Brain repo location
BRAIN_DATA_DIR="${BRAIN_DATA_DIR:-$HOME/.brain-data}"  # Data storage
BRAIN_LOGS_DIR="${BRAIN_LOGS_DIR:-$BRAIN_DATA_DIR/logs}"  # Logs
```

No hardcoded paths in generated configs.

## Migration from Old Architecture

### What Changed

| Before | After |
|--------|-------|
| Multiple docker-compose files | Single unified docker-compose.yml |
| MCPs duplicated (stdio + SSE) | MCP Gateway unifies all |
| skill-ninja didn't work | Works via gateway npm proxy |
| Hardcoded paths | Environment variable substitution |
| Ollama + Model Runner confusion | Clear primary/fallback strategy |
| Different configs per IDE | Unified via mcp-sync.sh |

### Migration Steps

1. **Backup**: `cp -r ~/.brain ~/.brain-backup`
2. **Update**: `git -C ~/.brain pull` (get new architecture)
3. **Configure**: Copy `brain.env.example` to `brain.env`, fill in values
4. **Start**: `brain up`
5. **Verify**: `brain doctor`
6. **Sync**: `brain sync`

## File Structure

```
~/.brain/
├── bin/
│   ├── brain              # Main CLI entry point
│   ├── brain-llm          # LLM management
│   └── brain-audit        # Code quality audit
├── mcp/
│   ├── gateway/           # MCP Gateway (Docker + Node)
│   │   ├── Dockerfile
│   │   ├── server.js
│   │   ├── package.json
│   │   └── mcp-registry.json
│   ├── brain-mcp-server/ # Custom brain MCP
│   ├── global-config.json # IDE configs (with env vars)
│   └── registry.yml       # Legacy registry (deprecated)
├── ai-local/
│   └── docker-compose.yml # Unified infrastructure
├── scripts/
│   ├── mcp-sync.sh        # Sync to all IDEs
│   ├── doctor.sh          # Health checks
│   └── install.sh         # Bootstrap (updated)
├── brain.env.example      # Configuration template
└── ARCHITECTURE.md        # This file
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BRAIN_DIR` | `~/.brain` | Brain repo location |
| `BRAIN_DATA_DIR` | `~/.brain-data` | Data storage |
| `BRAIN_LLM_PRIMARY` | `ollama` | Primary LLM provider |
| `BRAIN_LLM_FALLBACK` | `docker-model-runner` | Fallback provider |
| `OLLAMA_HOST` | `host.docker.internal:11434` | Ollama endpoint |
| `MODEL_RUNNER_URL` | `http://localhost:12434` | Model Runner endpoint |
| `GITHUB_TOKEN` | - | For GitHub MCP |
| `WEBUI_SECRET_KEY` | - | For Open WebUI |

## Ports

| Service | Port | Description |
|---------|------|-------------|
| MCP Gateway | 3000 | All MCPs via HTTP/SSE |
| Qdrant | 6333 | Vector database |
| Open WebUI | 8080 | Chat interface |
| Docker Model Runner | 12434 | Optional LLM backend |

## Best Practices

1. **Always use `brain` CLI**: Don't start services manually
2. **Check status first**: Run `brain status` before debugging
3. **Use environment variables**: Never hardcode paths
4. **One source of truth**: `brain.env` contains all configuration
5. **IDE sync**: Run `brain sync` after changing MCP configs

## Troubleshooting

### MCP Gateway not starting
```bash
brain logs mcp-gateway
```

### IDE can't connect to MCPs
```bash
brain status        # Check gateway is running
brain sync          # Re-sync configs
```

### Ollama not detected
```bash
ollama serve        # Start Ollama
brain status        # Verify connection
```

### Skill-ninja not working
Now works via gateway: `curl http://localhost:3000/mcp/skill-ninja/sse`

## Future Enhancements

- [ ] Kubernetes deployment manifests
- [ ] Cloud provider integrations (AWS/GCP/Azure)
- [ ] Web-based admin dashboard
- [ ] Model performance telemetry
- [ ] Multi-device sync protocols
