# Brain Repo — reeinharrrd
> A portable AI development environment that works on any OS, any IDE, any AI agent.

## What this is

A version-controlled "second brain" that:
- **Defines how I work** — universal coding principles, NOT tied to any language or tool
- **Configures every AI agent** — one source of truth, auto-adapted to Cursor, Windsurf, Claude Code, Gemini CLI, OpenCode, and Aider
- **Persists memory** — cross-session, cross-machine via Engram
- **Contains reusable agents** — orchestrator, researcher, planner, debugger, reviewer, etc.
- **Defines slash commands** — `/plan`, `/review`, `/research`, `/handover`, `/standup`, `/update-brain`

## Structure

```
~/.brain/
├── rules/              # Single source of truth for all rules
│   ├── canonical.md    # Core principles (edit this)
│   └── modules/        # communication, code-style, git, security, workflow
├── adapters/           # Generated configs for each agent/IDE
│   ├── generate.sh     # ← Run this after editing rules/
│   ├── claude-code/    # CLAUDE.md + settings.json
│   ├── cursor/         # .cursorrules
│   ├── windsurf/       # .windsurfrules
│   ├── gemini/         # GEMINI.md
│   ├── opencode/       # opencode.json
│   ├── cline/          # custom instructions
│   └── aider/          # .aider.conf.yml
├── agents/             # 9 global AI agents
├── commands/           # 6 global slash commands
├── mcp/                # MCP registry + profiles
├── providers/          # Model routing config
├── memory/             # Cross-session memory (Engram)
├── hooks/              # Claude Code pre/post tool-use hooks
├── install.sh          # Bootstrap (OS-aware)
├── doctor.sh           # Full diagnostic
└── update.sh           # Pull + regen + re-link
```

## Install

```bash
# Clone
git clone git@github.com:reeinharrrd/brain.git ~/.brain

# Bootstrap (one command does everything)
bash ~/.brain/install.sh

# Verify
bash ~/.brain/doctor.sh
```

## Operating Modes

The stable runtime path is:

- Core MCPs via `stdio` or docker on-demand
- Qdrant as the only persistent helper service

Bring up the persistent helper:

```bash
bash ~/.brain/docker/start.sh up
```

Check health:

```bash
bash ~/.brain/docker/start.sh status
bash ~/.brain/scripts/test-docker-mcp.sh
```

The legacy SSE compose stack is experimental and should not be treated as the default production path.

## Real Environment Validation

Before a push intended for real-world testing:

```bash
bash ~/.brain/scripts/smoke-real-env.sh
bash ~/.brain/scripts/benchmark-brain.sh
```

Artifacts are written under:

- `~/.brain/logs/real-runs/`
- `~/.brain/logs/benchmarks/`

Use `docs/reports/real-env-metrics-template.md` to compare runs and track regressions.

## Update rules

```bash
# 1. Edit the source of truth
vim ~/.brain/rules/canonical.md

# 2. Regenerate all agent adapters
bash ~/.brain/adapters/generate.sh

# 3. Commit
git -C ~/.brain commit -am "brain: updated [rule name]"
```

## Agents

| Agent | Purpose |
|-------|---------|
| `orchestrator` | Coordinates all other agents, reads providers.yml for model routing |
| `researcher` | Investigates tech/libraries/patterns with citations |
| `planner` | Turns goals into executable plans + ADRs |
| `designer` | UI/UX specs and component systems |
| `reviewer` | Code review with severity levels |
| `debugger` | Systematic root cause investigation |
| `refactor` | Safe incremental code improvement |
| `documenter` | README, API docs, ADRs, comments |
| `guardian` | Security audits and pre-tool-use blocking |

## Commands

| Command | Use when |
|---------|---------|
| `/plan` | Starting any task > 30 min |
| `/review` | Before merging anything significant |
| `/research` | Need a well-sourced recommendation |
| `/handover` | End of session, switching context |
| `/standup` | Start of session |
| `/update-brain` | Learning something that should be global |

## Supported agents/IDEs

✅ Claude Code · ✅ Cursor · ✅ Windsurf · ✅ OpenCode  
✅ Gemini CLI · ✅ Aider · ✅ Cline (VS Code extension)

## Philosophy

Language is a tool, not an identity. The brain repo contains **engineering principles**,  
not syntax guides. The goal is to think clearly, delegate intelligently, and ship quality work —  
regardless of the stack.

---

*"The real craft is in understanding what needs to be built, not in knowing every function name."*
