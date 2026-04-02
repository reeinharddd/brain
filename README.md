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

**In 30 seconds**: Clone it, run `brain setup`, and every IDE reads the same rules.

---

## Quick Start (5 Minutes)

### 1. Clone

```bash
git clone https://github.com/reeinharrrd/brain.git ~/.brain
cd ~/.brain
```

### 2. Setup

```bash
brain setup
```

This creates:

- `~/.brain/.brain.config` — Central configuration
- Symlinks for IDE integration (Cursor, Windsurf, Claude Code, etc.)
- Optional systemd autostart (automated)
- Git hooks for security checks

### 3. Verify

```bash
brain status      # See service status
brain health      # Run health checks
```

### 4. Start Using

```bash
# Start all services (Docker, MCP, memory)
brain start

# Use your IDE
code /my/project
cursor /my/project
windsurf /my/project

# Stop services
brain stop
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
# Service control
brain start                 # Launch Docker + MCP + memory backend
brain stop                  # Stop all services
brain restart               # Restart everything
brain status                # Show what's running
brain health                # Full diagnostics
brain doctor                # Full integrity checks
brain logs [N]              # Show last N log lines

# Repository and config maintenance
brain setup                 # Bootstrap or repair the Brain repo locally
brain init                  # Initialize the current project
brain generate              # Regenerate adapters and derived outputs
brain sync-mcp              # Sync MCP configs to supported IDEs
brain validate              # Validate rules and configuration
brain update                # Pull updates and refresh outputs

# Configuration
brain config                # Show/edit configuration
brain dashboard             # Interactive menu (TUI)

# Optional autostart
brain autostart-enable      # Start on boot
brain autostart-disable     # Manual start only
brain autostart-status      # Check status

# Debugging
brain reset                 # Clear state and logs
```

---

## Architecture

### Static Layer (Configuration)

These are read-only files that configure IDEs and agents:

```text
rules/canonical.md          <- Single source of truth (edit here)
├─ Compiled to all IDEs by: adapters/generate.sh
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

Scripts that run services and execute tasks:

```text
scripts/
├─ brain-cli.sh             <- Central orchestrator (all commands)
├─ init.sh                  <- Per-project initialization
├─ doctor.sh                <- Health check & diagnostics
├─ lib/
│  ├─ common.sh            <- Shared utilities (logging, errors, docker, etc.)
│  ├─ colors.sh            <- ANSI color definitions
│  ├─ logging.sh           <- Consistent logging
│  ├─ docker.sh            <- Docker utilities
│  └─ assert.sh            <- Assertion helpers
│
mcp/brain-mcp-server/       <- Custom MCP: exposes rules, agents, memory as tools
docker-compose.yml          <- Services (Qdrant memory, optional gateway)
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
│   └── generate.sh              ← Rebuilds all from canonical.md
│
├── agents/                      ← Agent prompt definitions
│   ├── researcher.md, planner.md, debugger.md, ...
│   └── agent-runner.py          ← Execute agents programmatically
│
├── commands/                    ← Slash command specs
│   └── plan.md, review.md, research.md, ...
│
├── scripts/                     ← All operational scripts
│   ├── brain-cli.sh             ← Main orchestrator
│   ├── init.sh                  ← Per-project init
│   ├── doctor.sh                ← Health check
│   ├── lib/                     ← Reusable modules
│   │   ├── common.sh, colors.sh, logging.sh, docker.sh, assert.sh
│   │   └── ...
│   └── (30+ other utilities)
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

#### Step 3: Run Setup

```bash
brain setup
```

This:

- Makes scripts executable
- Adds brain command to PATH
- Creates `.brain.config`
- Sets up IDE symlinks
- Tests basic functionality

#### Step 4: Start Services (Optional)

```bash
brain start           # Launches Docker containers if configured
brain status          # Verify everything is running
```

#### Step 5: Initialize a Project

In any project directory:

```bash
cd your/project
bash ~/.brain/scripts/init.sh
```

This:

- Links project-specific rules
- Installs git hooks
- Initializes `.env.example`
- Shows next steps

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
# Regenerates all IDE configs automatically
bash ~/.brain/adapters/generate.sh
```

### I want to add a new agent

1. Create `~/.brain/agents/my-agent.md`
2. Define the prompt and capabilities
3. Reference it in `/commands/my-command.md`
4. Run in terminal:

```bash
python3 ~/.brain/scripts/agent-runner.py my-agent --input "task" --memory
```

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
# Regenerate adapters
bash ~/.brain/adapters/generate.sh

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
