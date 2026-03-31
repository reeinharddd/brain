# Local AI Configuration

> Version: 1.0.0 | Last updated: 2026-03-22
> This module defines the standard for local AI model management and access within the Brain Station.

---

## Architecture

The Brain Station uses a hybrid local AI architecture to maximize compatibility and performance:

1.  **Ollama (Host-based)**: Primarily used by `ai-local` internal services (Open WebUI, etc.) and low-latency local tasks.
2.  **Docker Model Runner (DMR)**: Provides a standardized, Docker-native API for external IDEs and tools.

---

## Access Points

| Service | Endpoint | Purpose |
| :--- | :--- | :--- |
| **Ollama Bridge** | `http://localhost:11435/v1` | Legacy and high-performance access |
| **Docker Model Runner** | `http://localhost:12434/engines/v1` | Standardized IDE integration (Cursor, VS Code) |
| **Open WebUI** | `http://localhost:3000` | Human-friendly model interaction |

---

## Model Management

### Docker Model Runner (DMR)
Models must be pulled using the `docker model` CLI:
```bash
docker model pull ai/qwen2.5-coder:7b
```
Use the full model identifier (e.g., `ai/qwen2.5-coder`) in your IDE configuration.

### Ollama
Models must be managed via the `ollama` CLI:
```bash
ollama pull llama3
```

---

## IDE Integration Guidelines

### Cursor / VS Code (via OpenAI Provider)
- **Base URL**: `http://localhost:12434/engines/v1`
- **API Key**: `not-needed` (any string)
- **Model Name**: Use the DMR identifier (e.g., `ai/qwen2.5-coder`)

### Cline / Continue (Ollama Provider)
- **Base URL**: `http://localhost:11434` (Ollama Direct) or `http://localhost:12434` (DMR)

---

## Troubleshooting

- **Server not reachable**: Ensure `ai-local` services are up: `bash start-brain.sh up`.
- **GPU Issues**: DMR is configured to run in CPU-only mode by default on Linux to avoid `/dev/dri` conflicts.
- **Model not found**: Verify the model has been pulled for the specific runner you are using.
