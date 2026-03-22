# Brain Repo - reeinharrrd

> A portable AI development environment: one source of truth, compiled to every IDE and agent.

## What this actually is

A version-controlled developer brain that:

- **Defines how I work** - universal engineering principles in `rules/canonical.md`, compiled to every IDE automatically
- **Configures every AI agent** - one source of truth, auto-adapted to Cursor, Windsurf, Claude Code, Gemini CLI, OpenCode, Aider, Cline, Copilot
- **Persists memory** - cross-session knowledge graph via MCP memory server, with vector search via Qdrant
- **Contains executable agents** - 13 specialized agents runnable via `scripts/agent-runner.py` or as IDE prompts
- **Enforces security** - guardian checks on every commit and pre-tool-use hook in Claude Code
- **Defines slash commands** - `/plan`, `/review`, `/research`, `/handover`, `/standup`, `/update-brain`, `/memory-search`, `/consolidate`
- **Evaluates itself** - benchmark suite measuring memory recall, guardian accuracy, and adapter correctness

## Architecture

```
rules/canonical.md          <- Single source of truth (edit here)
        |
adapters/generate.sh        <- Compiles rules to all targets
        |
        +-- adapters/claude-code/CLAUDE.md
        +-- adapters/cursor/.cursorrules
        +-- adapters/windsurf/.windsurfrules
        +-- adapters/gemini/GEMINI.md
        +-- adapters/opencode/opencode.json
        +-- adapters/aider/system-prompt.md
        +-- adapters/cline/cline_custom_instructions.md
        +-- adapters/copilot/copilot-instructions.md

agents/*.md                 <- Agent prompt definitions (loaded by agent-runner.py)
commands/*.md               <- Slash command specifications
mcp/registry.yml            <- MCP server catalog
mcp/brain-mcp-server/       <- Custom MCP: exposes rules/agents/routing as tools
providers/providers.yml     <- Model routing table (task-type -> model tier)
memory/                     <- Memory manifest and vector config
hooks/                      <- Claude Code pre/post tool-use hooks
guardian/                   <- Security audit checks
evals/                      <- Benchmark suite
scripts/                    <- All operational scripts
```

## System components

### Static (configuration layer)
These are files read by IDEs and agents as context. No runtime required.

| Component | Purpose |
| :--- | :--- |
| `rules/canonical.md` | Core engineering principles |
| `rules/modules/*.md` | Modular rule sets (git, security, memory, code-style...) |
| `agents/*.md` | Agent definitions with role, methodology, anti-patterns |
| `commands/*.md` | Slash command protocols |
| `providers/providers.yml` | Model routing: task type -> model tier -> model name |

### Runtime (execution layer)
These are scripts with actual execution logic.

| Script | Purpose |
| :--- | :--- |
| `scripts/agent-runner.py` | Execute agents: calls API, injects memory+rules, supports pipelines |
| `scripts/provider-proxy.sh` | Runtime model routing with circuit breaker and cost logging |
| `scripts/consolidate-memory.sh` | Memory consolidation: detect contradictions, surface rule candidates |
| `scripts/embed.py` | Embedding backend: OpenAI > sentence-transformers > hash fallback |
| `scripts/validate-schema.py` | Schema validation for canonical.md before adapter generation |
| `scripts/cron-setup.sh` | Install automated maintenance tasks (daily validation, weekly consolidation) |
| `mcp/brain-mcp-server/server.py` | Custom MCP server: 7 tools for rules, agents, routing, search |

### Memory layer

| Component | What it does |
| :--- | :--- |
| MCP memory server | Knowledge graph with entities and relations (via `@modelcontextprotocol/server-memory`) |
| Qdrant (optional) | Vector search for semantic codebase context retrieval |
| `scripts/embed.py` | Embedding backend with graceful degradation |
| `memory/manifest.json` | Stats and metadata about memory state |
| `rules/modules/memory-types.md` | Entity type schema for classified memory storage |

### Evaluation layer

| Benchmark | What it measures |
| :--- | :--- |
| `evals/benchmarks/memory-retrieval.sh` | Recall rate: does memory surface the right entities? |
| `evals/benchmarks/guardian-coverage.sh` | Guardian accuracy: catches bad code, passes good code |
| `evals/benchmarks/adapter-schema.sh` | Adapter correctness: all outputs valid and non-empty |
| `evals/skills/*.sh` | Skills evals (stack detection, render-skill-context) |

## Install

```bash
# Clone
git clone git@github.com:reeinharddd/brain.git ~/.brain

# Bootstrap
bash ~/.brain/scripts/install.sh

# Verify
bash ~/.brain/scripts/doctor.sh

# Optional: install automated maintenance tasks
bash ~/.brain/scripts/cron-setup.sh
```

## Autostart

The brain environment can be configured to start automatically upon login:

```bash
# Register autostart service (Linux/WSL)
bash ~/.brain/scripts/autostart-setup.sh
```

This ensures that MCPs, memory servers, and core rule validators are active in every terminal and IDE session without manual intervention.

## Update rules (the core loop)

```bash
# 1. Edit the source of truth
vim ~/.brain/rules/canonical.md

# 2. Validate schema before generating
python3 ~/.brain/scripts/validate-schema.py

# 3. Regenerate all adapters
bash ~/.brain/adapters/generate.sh

# 4. Commit
git -C ~/.brain commit -am "brain: updated [rule name]"
```

## Agent execution

Agents are markdown definitions that can be used in two ways:

**As IDE context** - paste the agent content as a system prompt in Claude Code, Cursor, etc.

**As executable agents** (via agent-runner.py):
```bash
# List all agents
python3 ~/.brain/scripts/agent-runner.py --list

# Run a single agent
python3 ~/.brain/scripts/agent-runner.py \
  --agent researcher \
  --task "Compare Qdrant vs Weaviate for production vector search" \
  --memory

# Run a pipeline (output of each feeds next)
python3 ~/.brain/scripts/agent-runner.py \
  --pipeline "planner->implementer->reviewer" \
  --task "Add JWT authentication to the Express API"
```

Note: agent-runner.py requires `ANTHROPIC_API_KEY` to be set.

## Memory usage

```bash
# Search memory before starting a task
/memory-search [query]

# Save end-of-session context
/handover

# Run consolidation (monthly or after large projects)
bash ~/.brain/scripts/consolidate-memory.sh

# Preview consolidation without writing
bash ~/.brain/scripts/consolidate-memory.sh --dry-run
```

## Run evaluations

```bash
# Full eval suite
bash ~/.brain/evals/run.sh

# JSON output for CI
bash ~/.brain/evals/run.sh --json

# Only benchmarks
bash ~/.brain/evals/run.sh --only benchmarks

# Individual benchmarks
bash ~/.brain/evals/benchmarks/memory-retrieval.sh
bash ~/.brain/evals/benchmarks/guardian-coverage.sh
bash ~/.brain/evals/benchmarks/adapter-schema.sh
```

## Brain MCP server

The custom MCP server exposes brain internals as tools for any MCP-compatible client:

```json
{
  "mcpServers": {
    "brain-rules": {
      "command": "python3",
      "args": ["${HOME}/.brain/mcp/brain-mcp-server/server.py"]
    }
  }
}
```

Available tools: `brain_get_rules`, `brain_get_agent`, `brain_list_agents`,
`brain_get_command`, `brain_route_task`, `brain_search_rules`, `brain_get_provider`

## Agents

| Agent | Tier | Purpose |
| :--- | :--- | :--- |
| `orchestrator` | powerful | Coordinates all agents, reads providers.yml for routing |
| `researcher` | standard | Investigates tech/libraries with citations |
| `planner` | powerful | Turns goals into executable plans and ADRs |
| `architect` | powerful | Designs components, compares options, documents trade-offs |
| `implementer` | standard | Scoped implementation from accepted specs |
| `designer` | powerful | UI/UX specs and component systems |
| `reviewer` | standard | Code review with severity levels |
| `debugger` | standard | Systematic root cause investigation |
| `refactor` | standard | Safe incremental code improvement |
| `documenter` | fast | README, API docs, ADRs, comments |
| `guardian` | standard | Security audits and pre-tool-use blocking |
| `configurator` | fast | Environment and tooling setup |

## Commands

| Command | When to use |
| :--- | :--- |
| `/plan` | Starting any task > 30 min |
| `/review` | Before merging anything significant |
| `/research` | Need a well-sourced recommendation |
| `/handover` | End of session, switching context |
| `/standup` | Start of session |
| `/update-brain` | Learning something that should be global |
| `/memory-search` | Before a task - retrieve relevant past context |
| `/consolidate` | Monthly or after a large project - clean up memory |

## Supported IDEs and agents

Claude Code - Cursor - Windsurf - OpenCode - Gemini CLI - Aider - Cline - GitHub Copilot

## Philosophy

Language is a tool, not an identity. The brain repo contains engineering principles,
not syntax guides. The goal is to model problems clearly, delegate intelligently,
and produce quality outcomes regardless of the stack.

"The real craft is in understanding what needs to be built, not in knowing every function name."
