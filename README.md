# Brain Repository

> A portable AI development environment: one source of truth, compiled to every IDE and LLM

---

## What is Brain?

**Brain** is a version-controlled developer environment that:

- **Defines how you work** — Universal engineering principles + IDE-agnostic configurations
- **Configures every AI agent** — One source of truth (`rules/canonical.md`), auto-compiled to Cursor, Windsurf, Claude Code, Copilot, Gemini, Aider, OpenCode, Cline
- **Persists knowledge** — Cross-session memory graph via MCP server + vector search (Qdrant)
- **Enforces security** — Automatic guardian checks on commits, prevents secrets leakage
- **Runs agents** — 12 specialized agents (researcher, planner, debugger, etc.) callable from terminal or IDE
- **Evaluates itself** — Benchmark suite measuring memory recall, security, rule consistency

**In 30 seconds**: Clone it, run `brain install-global` and `brain sync`, and every IDE reads the same rules.

---

## Quick Start (from a fresh clone)

### 1. Clone

```bash
git clone https://github.com/reeinharrrd/brain.git ~/.brain
cd ~/.brain
```

### 2. Install global commands (production-like)

```bash
go run ./cli/cmd/brain/main.go install-global
```

This installs:

- `~/.local/bin/brain`
- `~/.local/bin/braind`
- `~/.config/brain/root` (points to your cloned repo root)

If `brain` is not found in a new shell, add:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 3. Verify

```bash
brain daemon-start
brain status
brain sync
```

### 4. Start Using

```bash
# Start daemon
brain daemon-start

# Use your IDE
code /my/project
cursor /my/project
windsurf /my/project

# Stop daemon
brain daemon-stop
```

---

## Keyboard Commands

**In any IDE**, you can use these **slash commands** (via CLI or IDE integration):

| Command          | What it does                                    |
| ---------------- | ----------------------------------------------- |
| `/plan`          | Break down a complex task into steps            |
| `/review`        | Code review of your changes                     |
| `/research`      | Deep research with web search                   |
| `/handover`      | Save context for next session                   |
| `/standup`       | Quick project status summary                    |
| `/memory-search` | Query your knowledge graph                      |
| `/consolidate`   | Clean up old memories, detect patterns          |
| `/update-brain`  | Pull latest Brain updates + regenerate adapters |

---

## Core Commands

Use the `brain` CLI (always available):

```bash
# Global install
brain install-global        # Build/install brain + braind to ~/.local/bin

# Daemon lifecycle
brain daemon-start          # Start daemon in background
brain daemon-stop           # Stop daemon
brain status                # Show daemon status and process count

# Runtime operations
brain sync                  # Trigger unified config sync through daemon
brain ps                    # List managed sub-process states
brain logs                  # Stream real-time daemon logs via WebSocket

# Desktop UI
brain ui                    # Start desktop dev UI (and ensure daemon is running)

# Process control
brain start <id> <cmd> ...  # Start a managed process
brain stop <id>             # Stop a managed process
```

---

## Desktop Control Plane

The **Brain Desktop UI** (Tauri + React) provides a real-time control plane for daemon orchestration:

### Features

- **Manager Status Dashboard** — Real-time indicators for Docker, Qdrant, Ollama, MCP Registry
- **Daemon Controls** — One-click Start/Stop for entire infrastructure
- **LLM Provider Detection** — Auto-discover available providers (Claude, Gemini, Ollama, OpenAI)
- **Live Daemon Logs** — WebSocket-based real-time log streaming (100-entry buffer)
- **Process Management** — View and control running sub-processes

### Starting the Desktop UI

```bash
# Option 1: From CLI
brain ui

# Option 2: Manual (development mode)
cd ~/.brain/desktop
npm run dev
# Opens http://localhost:5173
```

### What You See

```text
┌─────────────────────────────────────────────┐
│ Status: Running/Stopped                     │
│ [START] [STOP] [REFRESH PROVIDERS]         │
├─────────────────────────────────────────────┤
│ Docker  │ Qdrant  │ Ollama  │ MCP  │Providers
│ Ready   │ :6333   │ :11434  │Synced│ 3 avail
├─────────────────────────────────────────────┤
│ Available Providers:                        │
│ Claude 3.5 Sonnet | Gemini 2.5 Pro         │
│ Ollama (local)    | OpenAI GPT-4o          │
├─────────────────────────────────────────────┤
│ Processes         │ Live Daemon Logs       │
│ [web_ping Running]│ [Daemon logs stream]   │
└─────────────────────────────────────────────┘
```

### API Integration

The desktop UI communicates with the daemon via HTTP + WebSocket on port 9090:

**Status Endpoints** (polled every 5 seconds):

- `GET /api/status` — Overall daemon status
- `GET /api/docker/status` — Docker manager
- `GET /api/qdrant/status` — Qdrant vector DB
- `GET /api/ollama/status` — Ollama LLM runtime
- `GET /api/mcp/status` — MCP registry sync

**Provider Detection**:

- `GET /api/providers/available` — Available LLM providers

**Daemon Control**:

- `POST /api/daemon/start` — Start all services
- `POST /api/daemon/stop` — Stop all services

**Real-time Logs**:

- `WebSocket /ws` — Live log stream (no polling)

### Keyboard Shortcuts

| Action               | Method                           |
| -------------------- | -------------------------------- |
| Start all services   | Click [START] button             |
| Stop all services    | Click [STOP] button              |
| Refresh providers    | Click [REFRESH PROVIDERS] button |
| View process details | Click process row                |

---

## Architecture

### Static Layer (Configuration)

These are read-only files that configure IDEs and agents:

```text
rules/canonical.md          <- Single source of truth (edit here)
├─ Compiled to all IDEs by: brain sync
│  ├─ adapters/claude-code/CLAUDE.md
│  ├─ adapters/cursor/.cursorrules
│  ├─ adapters/windsurf/.windsurfrules
│  ├─ adapters/copilot/copilot-instructions.md
│  └─ ... (more adapters)
│
agents/*.md                 <- 12 agent prompts (researcher, planner, debugger, etc.)
commands/*.md               <- Slash command specs (/plan, /review, /research, etc.)
mcp/registry.yml            <- MCP server catalog
providers/providers.yml      <- Model routing table (task-type → model)
```

### Runtime Layer (Execution)

Go commands and runtime components:

```text
cli/cmd/brain/main.go       <- Central orchestrator (all commands)
daemon/cmd/braind/main.go   <- Runtime daemon and sync engine
desktop/src/App.tsx         <- Visual control plane
commands/*.md               <- Slash command definitions
mcp/registry.yml            <- MCP catalog and routing
```

### Knowledge Layer (Memory)

Persistent across sessions:

```text
memory/
├─ vector-config.json       <- Qdrant configuration
├─ manifest.json            <- Memory graph schema
└─ chunks/                  <- Vector search index

logs/
├─ brain-cli.log            <- CLI execution log
├─ telemetry.ndjson         <- Performance metrics
└─ cron/                    <- Automated task logs
```

---

## Configuration

All settings live in **one file**: `~/.brain/.brain.config`

```bash
# Auto-generated on first run, edit manually to customize:

# Autostart on boot (true|false)
AUTOSTART=false

# Memory backend (qdrant|local)
MEMORY_BACKEND=qdrant

# MCP server ports
MCP_PORT_BASE=8001

# Logging level (DEBUG|INFO|WARN|ERROR)
LOG_LEVEL=INFO

# Optional: Cloud sync
MEMORY_CLOUD_SYNC=false
MEMORY_CLOUD_PROVIDER=none
```

Edit with:

```bash
brain config                # Opens in your editor
```

---

## Multi-IDE Usage

**All IDEs use the same rules and services — no conflicts:**

```bash
# Start services once
brain start

# Open multiple IDEs (they share the same servers)
code /project &
cursor /project &
windsurf /project &

# Each IDE reads:
# - ~/.cursorrules        (Cursor)
# - ~/.windsurfrules      (Windsurf)
# - ~/.claude/commands/   (Claude Code)
# - etc.

# All communicate via ONE MCP gateway on ports 8001-8005
```

**No setup per IDE.** No startup conflicts. Everything synchronized.

---

## File Structure

```text
~/.brain/
├── README.md                    ← You are here
├── brain.env, brain.env.example ← Configuration
├── CLAUDE.md                    ← Root compatibility instruction file
│
├── rules/
│   └── canonical.md             ← THE SOURCE OF TRUTH
│
├── adapters/                    ← IDE-specific compiled rules
│   ├── claude-code/, cursor/, windsurf/, copilot/, etc.
│   └── brain sync               ← Rebuilds generated configs from canonical sources
│
├── agents/                      ← Agent prompt definitions
│   ├── researcher.md, planner.md, debugger.md, ...
│   └── brain agents             ← Inspect and route agents from the CLI
│
├── commands/                    ← Slash command specs
│   └── plan.md, review.md, research.md, ...
│
├── cli/                         ← Brain CLI entrypoints
│   └── cmd/brain/main.go        ← Main orchestrator
├── daemon/                      ← Runtime daemon and sync engine
│   └── cmd/braind/main.go       ← Background control plane
├── desktop/                     ← Tauri + React UI
├── commands/                    ← Slash command specs
├── mcp/                         ← MCP registry and server
│
├── mcp/                         ← MCP server setup
│   ├── registry.yml
│   ├── brain-mcp-server/        ← Custom tools
│   └── docker-compose.yml
│
├── memory/                      ← Knowledge graph storage
│
├── docs/                        ← Detailed documentation
│   ├── guides/                  ← How-to guides
│   └── adr/                     ← Architecture Decision Records
│
├── guardian/                    ← Security checks
│
├── hooks/                       ← Git pre/post hooks
│
├── evals/                       ← Benchmark suite
│
└── tests/                       ← Test suite
```

---

## Installation Details

### Full Installation

For detailed setup:

#### Requirements

- **Bash 4+** (macOS users: `brew install bash`)
- **Docker** & **Docker Compose** (optional, for Qdrant memory backend)
- **Python 3.8+** (for agents and embedding)
- **Git** (for repo management)
- An API key for your preferred LLM (Claude, OpenAI, Gemini, etc.)

#### Step 1: Clone Repository

```bash
git clone https://github.com/reeinharrrd/brain.git ~/.brain
cd ~/.brain
```

#### Step 2: Create Configuration

```bash
cp brain.env.example brain.env
# Edit brain.env with your API keys:
nano brain.env
```

Required environment variables:

- `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` (depending on your LLM)

#### Step 3: Install the CLI

```bash
brain install-global
```

This:

- Builds and installs `brain` and `braind`
- Adds `brain` to your PATH when needed
- Saves the resolved repo root for the CLI and daemon

#### Step 4: Start Services (Optional)

```bash
brain daemon-start    # Launches the daemon
brain status          # Verify everything is running
```

#### Step 5: Initialize a Project

In any project directory, use the native project command or refresh the synced context:

```text
/init
brain sync
```

This:

- Refreshes generated Brain configs
- Updates the project context files used by supported IDEs
- Applies the current Brain rules to the project

#### Step 6: Use Your IDE

The brain automatically integrates with:

- **VS Code** — Reads `~/.vscode/copilot-instructions.md`
- **Cursor** — Reads `~/.cursorrules`
- **Windsurf** — Reads `~/.windsurfrules`
- **Claude Code** — Reads `~/.claude/commands/`
- **Copilot Chat** — Reads VS Code instructions
- **Gemini CLI** — Reads `~/.gemini/GEMINI.md`
- **Aider** — Reads `~/.aider.conf.yml`
- **OpenCode** — Reads `~/.config/opencode/opencode.json`

No additional setup needed — just open an IDE.

---

## Common Use Cases

### I want to change my coding standards

Edit `~/.brain/rules/canonical.md`, then run:

```bash
brain sync
```

### I want to add a new agent

1. Create `~/.brain/agents/my-agent.md`
2. Define the prompt and capabilities
3. Reference it in `/commands/my-command.md`
4. Use the `brain agents` commands to inspect or synchronize the agent set

### I want to share context between sessions

Use `/handover` command:

```bash
# Saves current session state
/handover
```

Next session, read it back:

```bash
/standup          # Show what you worked on before
/memory-search    # Find relevant context
```

### I want to enforce security checks

Edit `~/.brain/guardian/checks/` to add new security rules. They run:

- On every `git commit`
- When pushing to remote
- During `brain health` checks

### I want to use local inference (no API)

Replace adapters with local model configs:

```bash
# Use Ollama locally
ANTHROPIC_API_KEY=local brain start
```

---

## Troubleshooting

### Services not starting

```bash
# Full diagnostics
brain health

# Check logs
brain logs 50

# Reset and try again
brain reset
brain start
```

### IDE not reading rules

```bash
brain sync

# Verify symlinks
ls -l ~/.cursorrules ~/.windsurfrules ~/.claude/
```

### Memory not working

```bash
# Check Qdrant health
curl http://localhost:6333/health

# Restart memory backend
docker restart brain-qdrant
```

### Python imports failing

```bash
# Ensure brain.env is sourced
source ~/.brain/brain.env

# Install Python deps (if needed)
python3 -m pip install qdrant-client python-dotenv
```

---

## Architecture Decisions

Why is Brain structured this way?

**Why one source of truth (`canonical.md`)?**

- Reduces consistency bugs (one rule change, all IDEs update)
- Easier to audit (one file to review)
- Avoids IDE-specific workarounds
- Version-controlled history

**Why compile to adapters instead of inline?**

- IDEs understand their native formats (.cursorrules, CLAUDE.md, etc.)
- Works offline (no runtime translation)
- Better IDE integration (syntax highlighting, validation)
- Faster IDE startup

**Why agents instead of only IDE context?**

- Agents can be:
  - Run from CLI independently
  - Chained in pipelines
  - Used across tools (Aider, Cline, etc.)
  - Versioned separately from rules

**Why MCP instead of direct integration?**

- IDE-agnostic (works anywhere that speaks MCP)
- Secure (confined execution environment)
- Easy to extend (just add new tools)
- Future-proof (MCP becoming standard)

See `docs/adr/` for more decisions.

---

## Contributing

This is a personal brain repo, but if you find bugs or have ideas:

1. Open an issue at [brain/issues](https://github.com/reeinharrrd/brain/issues)
2. Submit a PR with improvements
3. Test changes: `bash scripts/validate.sh`

---

## License

MIT License — Use freely, modify, redistribute.

---

## Support

- **Read** — Check `docs/guides/`
- **Ask** — Use `/research` to find answers
- **Debug** — Run `brain doctor` for diagnostics
- **Contribute** — Open a PR

---

**Version**: Refactored March 31, 2026  
**Status**: Production-ready with ongoing cleanup
