# Brain Repository: GitHub Copilot Instructions

**Version**: 2.0.0 | **Last Updated**: 2026-04-03 | **Scope**: ~/.brain only

---

## Executive Summary

Brain is a portable AI development environment compiled to 8 IDEs + 1 CLI + 1 Desktop UI. It features a centralized Go daemon orchestrator, thin CLI client, and TypeScript React UI control plane. 

**Single Source of Truth**: Daemon (braind) at port 9090 coordinates all services. CLI and UI are consumers only.

---

## Project Architecture at a Glance

```
Brain Repository (~/brain/)
├── daemon/              Go orchestrator (braind @ 9090)
│   ├── cmd/braind/      HTTP API + WebSocket server
│   └── internal/        Business logic (Docker, Qdrant, MCP, Skills, Providers)
├── cli/                 Thin client (HTTP to daemon)
│   └── cmd/brain/       CLI commands
├── desktop/             Tauri + React control panel
│   └── src/             React components + WebSocket client
├── rules/               Global development principles
│   ├── canonical.md     Single source of truth for all adaptation
│   └── modules/         Code style, security, memory, workflow
├── config/              Control plane definitions
│   ├── agents.json      12 AI agents
│   ├── mcps.json        11 MCP servers
│   └── providers.yml    LLM routing + fallback
├── docker/              Persistent services
│   └── compose.yml      Qdrant + optional Ollama
├── skills/              Portable skill packages
│   ├── registry.yml     Metadata-only discovery index
│   ├── contexts/        Reusable stack guidance
│   ├── dynamic-registry.tsv  Context packs index
│   └── <id>/SKILL.md    Individual skill definitions
├── agents/              Agent prompt definitions
├── mcp/                 MCP server configurations
└── docs/                Architecture decisions (ADRs)
```

---

## Core Development Constraints

### 1. Language Rules (STRICT)

- **Go ONLY** for daemon, CLI, and automation
  - daemon/ : braind service, managers, API handlers
  - cli/cmd/brain/ : user-facing commands
  - scripts/ : Go executables, NOT bash/Python
  
- **TypeScript/React** for desktop UI (Tauri + React)
  - desktop/src/ : React components
  - Strict mode enforced (noUnusedLocals, noUnusedParameters, noImplicitAny)

- **YAML/JSON** for configuration
  - providers.yml, agents.json, config/, artifacts/skills/registry.yml

- **Forbidden**: Bash scripts (.sh), Python scripts (.py) in ~/.brain
  - Previous cleanup removed all shell/Python orchestration
  - Exception: Project-specific scripts in local projects are OK

- **English ONLY** - 100% requirement for entire repository
  - ALL code comments MUST be in English
  - ALL variable names, function names, types MUST be in English
  - ALL documentation (README, ADRs, comments, docstrings) MUST be in English
  - ALL commit messages MUST be in English
  - ALL error messages and log output MUST be in English
  - NO Spanish, NO other languages in any code or documentation
  - This applies to: daemon, CLI, UI, config files, documentation, comments
  - Exception: User content in project-specific repos outside ~/.brain can use any language

### 2. Architecture Rules (NON-NEGOTIABLE)

**Single Daemon Principle**
- Daemon (braind @ 9090) is the ONLY active orchestrator
- CLI is thin client: HTTP calls to daemon, no direct service access
- Desktop UI is control panel: WebSocket to daemon for real-time updates
- No duplicate docker-compose.yml files
- MCP servers run in stdio mode (launched by daemon)

**100% Implementation Rule (Critical)**
Every feature MUST be implemented 100% across ALL three surfaces:
1. **Daemon**: Go API endpoint exposing data/logic
2. **CLI**: Command to read/write that data
3. **UI**: React component displaying data + user interactions

NO EXCEPTIONS for partial features. If feature exists on daemon but not UI, it's incomplete.

**Master Registry Principle**
- artifacts/skills/registry.yml is source of truth for executable skills
- CLI writes to registry first, then filesystem second
- Daemon validates sync every 5 minutes (logs orphans if found)
- No orphan skills allowed (hardcoded rule, enforced by daemon ticker)


### 3. File and Code Organization

**File Locations**
```
daemon/internal/manager/skills.go        -> Business logic
daemon/internal/api/handlers.go          -> HTTP endpoints
daemon/cmd/braind/main.go                -> Service initialization
cli/cmd/brain/registry.go                -> CLI commands
desktop/src/components/SkillsList.tsx    -> UI components
```

**Go File Structure** (daemon + CLI)
```go
package main

import (
    "stdlib"               // Go standard library
    "external/package"     // Third-party
    "internal/module"      // Internal modules
)

const (
    DAEMON_PORT = 9090
    BRAIN_ROOT  = "/home/user/.brain"
)

type Manager struct {
    field1 string
    field2 int
}

func (m *Manager) Method() error {
    // max 30 lines soft limit
}
```

**TypeScript File Structure** (React components)
```typescript
import React from 'react';
import { useEffect, useState } from 'react';
import { fetchData } from '../api/client';

const ComponentName: React.FC = () => {
  const [state, setState] = useState();
  
  useEffect(() => {
    // setup
  }, []);
  
  return <div></div>;
};

export default ComponentName;
```

**Max File Sizes**
- Go files: 300 lines (split into internal/ subdirectories if larger)
- React components: 150 lines (extract hooks/utils if larger)
- Config files: 500 lines (split into separate files if needed)

**Function Complexity**
- Max cyclomatic complexity: 10 per function
- Max nesting depth: 3 levels (use early returns)
- Ternary operators: Simple cases only, never nested

---

## Development Workflow

### Before You Start

1. **Understand the shape**: What is the goal? Who is it for? What are constraints?
2. **Check memory**: Does a related decision already exist? (ADRs in docs/adr/)
3. **Identify phase**: Is this a <30 min fix or >2 hour feature?
   - Tiny fixes: Just do it, no formal plan
   - Small tasks (30-120 min): Brief plan comment
   - Large features (>2 hrs): Full SDD phase breakdown

### The SDD DAG (for features >2 hours)

Follow these phases in order. Each phase produces an artifact.

1. **Explore** → Analyze codebase, identify constraints, detect stack → outputs: constraints, assumptions
2. **Propose** → Draft approach(es) with tradeoffs → outputs: selected approach + rationale
3. **Spec** → Write formal acceptance criteria and boundaries → outputs: spec document
4. **Design** → Architecture, flows, component interfaces → outputs: design doc + diagrams
5. **Tasks** → Break into atomic executable units → outputs: task list (GitHub issues)
6. **Implement** → Code changes, one task at a time → outputs: working code + tests
7. **Verify** → Tests pass, no lint warnings, behavior validated → outputs: test evidence
8. **Archive** → Docs updated, memory persisted, context recorded → outputs: closing summary

**Rule**: Do NOT skip phases. If you find yourself writing code in the Propose phase, STOP and complete Spec first.

### Commit Workflow

**Three Steps Before Committing**
1. Review diff: `git diff --staged` (inspect ALL changes)
2. Run tests: Go tests + TS tests must pass
3. Check lint: No new warnings allowed

**Commit Message Format** (Conventional Commits)
```
<type>(<scope>): <imperative description>

Optional body explaining WHY (not WHAT - diff shows what).
Line wrap at 80 characters.

Closes #123
```

**Commit Types**
- `feat(skills)`: New feature for skills system
- `fix(api)`: Bug fix in API handler
- `docs(README)`: Documentation update
- `refactor(daemon)`: Code improvement, no behavior change
- `test(manager)`: Add/update tests
- `chore(deps)`: Dependency updates
- `perf(sync)`: Performance improvement
- `ci(github)`: GitHub Actions changes
- `revert(abc123)`: Revert previous commit

**Brain Repo Specific**
- Always use prefix: `brain: <message>` (e.g., `brain: update copilot instructions`)
- After editing rules/: Always run `adapters/generate.sh` before committing
- Never commit environment-specific state (no hardcoded paths, no secrets)

---

## Communication with Copilot (Agent Delegation)

When asking Copilot to work on a task, use this structure:

**Tiny Fix** (< 30 min)
```
Fix: [Describe the issue]
Expected: [What success looks like]
Constraints: [Any rules to follow]
```

**Feature Task** (30 min - 2 hrs)
```
Task: [What needs to be built]
Goal: [Why we're building it]
Context: [Relevant files or background]
Acceptance Criteria: [How to verify success]
Constraints:
  - No shell scripts (Go only)
  - Implement in all 3 surfaces (daemon + CLI + UI)
  - Tests must pass
  - No new lint warnings
```

**Large Feature** (> 2 hrs)
Delegate to Planner agent first to create SDD breakdown, then delegate to Implementer for each phase.

**Rules for Any Delegation**
1. Be specific about scope: "Don't refactor X" or "Only change Y"
2. State constraints clearly: "Must use library Z, must support Node 18"
3. Specify what NOT to do: "Don't change database schema" or "Don't delete files"
4. Reference acceptance criteria: "Tests pass, TypeScript strict mode, runs locally"

---

## Required Tests for Any Change

### Test-First Requirement

- **BEFORE** implementing a feature, write the test that verifies it
- **Tests are not optional** - they're part of the feature definition
- **Target coverage**: 100% for new code (80% minimum acceptable)

### Test Examples

**Go (daemon/CLI)**
```go
func TestResolveBrainRoot(t *testing.T) {
    // Arrange
    os.Setenv("BRAIN_ROOT", "/custom/path")
    
    // Act
    result := resolveBrainRoot()
    
    // Assert
    if result != "/custom/path" {
        t.Fatalf("Expected /custom/path, got %s", result)
    }
}
```

**TypeScript (React)**
```typescript
describe('SkillsList', () => {
  it('renders skills from API response', async () => {
    // Arrange
    const mockSkills = [{ id: '1', name: 'Test' }];
    jest.spyOn(window, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({ skills: mockSkills })
    });
    
    // Act
    render(<SkillsList />);
    
    // Assert
    expect(await screen.findByText('Test')).toBeInTheDocument();
  });
});
```

### Validation Checklist

- [ ] Tests written and passing (`go test ./...` or `bun run test`)
- [ ] No TypeScript errors (strict mode compliance)
- [ ] No new lint warnings (eslint, gofmt)
- [ ] Function complexity <= 10 cyclomatic
- [ ] Code coverage >= 80%
- [ ] Error handling is explicit (no silent failures)
- [ ] Security audit passed (no hardcoded secrets)
- [ ] Works locally without manual setup steps
- [ ] Documentation updated

---

## Security Rules (ABSOLUTE)

### Secrets Management

- **NEVER hardcode secrets** - API keys, tokens, passwords, private URLs
- Use **environment variables** for all secrets
- **`.env` file**: Always in `.gitignore`, NEVER committed
- **`.env.example`**: ALWAYS exists with placeholder values, committed to repo
- Use **Guardian agent** pre-commit checks to block hardcoded secrets

Example `.env.example`:
```
ANTHROPIC_API_KEY=sk-xxx-YOUR-KEY-HERE
OPENAI_API_KEY=sk-xxx-YOUR-KEY-HERE
BRAIN_ROOT=$HOME/.brain
```

### Input Validation

- Validate ALL inputs from external sources (users, APIs, files, env vars)
- Sanitize before processing or passing to sub-systems
- Return error with context, never silent failures
- Log validation failures but don't expose details to end users

### Error Handling

- Explicit error returns in Go: `func() (result, error)`
- Wrap errors with context: `return fmt.Errorf("failed to load skills: %w", err)`
- No empty catch blocks in TypeScript
- All errors logged with meaningful context

---

## Critical Restrictions and Prohibitions

### NEVER DO (Absolute Rules)

1. **Never hardcode secrets** - Use environment variables
2. **Never commit .env files** - Always in .gitignore
3. **Never skip error handling** - All errors logged with context
4. **Never use @latest or unversioned dependencies** - Pin exact versions
5. **Never delete files without user approval** - Ask before destructive operations
6. **Never use shell scripts in ~/.brain** - Go only
7. **Never force-push to main/master** - Use revert commits instead
8. **Never leave commented-out code** - Delete or create issue/TODO
9. **Never implement features partially** - All 3 surfaces (daemon + CLI + UI) or incomplete
10. **Never silently fail** - Log errors, notify users, fail explicitly
11. **Never use emojis** - Use descriptive text only (no :rocket:, :check:, etc.)
12. **Never mix indentation** - Follow project standard (2 or 4 spaces, never mix)
13. **Never use non-English in code** - ALL code/docs/comments MUST be 100% English
    - No Spanish, no other languages in variable names, functions, comments, docs
    - All error messages and logs MUST be in English
    - Exception: User content in external projects (not in ~/.brain) can use other languages

---

## Project-Specific Patterns

### Error Handling Pattern (Go)

```go
func (m *Manager) DoSomething(ctx context.Context) error {
    // Early return on error, not deferred cleanup
    result, err := loadData(ctx)
    if err != nil {
        return fmt.Errorf("failed to load data: %w", err)
    }
    
    if result == nil {
        return errors.New("unexpected nil result")
    }
    
    return nil
}
```

### API Response Pattern

```go
type Response struct {
    Skills []CatalogItem `json:"skills"`
    Error  string        `json:"error,omitempty"`
}

// Handler
func (h *Handler) ListSkills(w http.ResponseWriter, r *http.Request) {
    skills := h.manager.GetSkills()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(Response{Skills: skills})
}
```

### React Hook Pattern

```typescript
import { useEffect, useState } from 'react';

interface SkillData {
  id: string;
  name: string;
}

export const useSkills = () => {
  const [data, setData] = useState<SkillData[]>([]);
  const [isLoading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    fetch('/api/skills')
      .then(r => r.json())
      .then(d => setData(d.skills))
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, []);
  
  return { data, isLoading, error };
};
```

### Config Loading Pattern (Go)

```go
type Config struct {
    BrainRoot string // from env: BRAIN_ROOT
    Providers map[string]Provider
}

func LoadConfig() (*Config, error) {
    brainRoot := os.Getenv("BRAIN_ROOT")
    if brainRoot == "" {
        brainRoot = resolveBrainRoot() // fallback
    }
    
    // Validate required paths exist
    if _, err := os.Stat(brainRoot); err != nil {
        return nil, fmt.Errorf("BRAIN_ROOT not found: %s", brainRoot)
    }
    
    return &Config{BrainRoot: brainRoot}, nil
}
```

---

## Tools and Environment

### Required Tools

| Tool | Version | Usage | Notes |
|------|---------|-------|-------|
| Go | 1.24.4+ | Daemon, CLI, automation | `go build`, `go test`, `gofmt` |
| TypeScript | 6.0.2+ | Desktop UI | Strict mode required |
| Bun | 1.3.10+ | JS/TS runtime | ONLY Bun, never npm/yarn |
| Docker | 29.3.0+ | Container orchestration | `docker compose` (space, not hyphen) |
| Git | 2.51.0+ | Version control | Pre-commit hooks enabled |

### Forbidden Tools
- **npm, yarn, pnpm**: Use Bun only
- **bash scripts**: Use Go executables
- **Python scripts**: Use Go executables
- **@latest or unversioned deps**: Pin all versions

### Environment Variables

**Required for Daemon**
```bash
BRAIN_ROOT=/home/user/.brain                 # Auto-detected, override if needed
ANTHROPIC_API_KEY=sk-xxx...                  # Default LLM provider
OPENAI_API_KEY=sk-xxx...                     # Fallback provider (optional)
GOOGLE_API_KEY=xxx...                        # Fallback provider (optional)
```

**Optional for Services**
```bash
OLLAMA_BASE_URL=http://localhost:11434       # Local LLM (optional)
QDRANT_API_KEY=xxx...                        # Qdrant auth (if using cloud)
```

---

## File Organization and Locations

**Key Config Files**
- `rules/canonical.md` - Source of truth for all principles
- `rules/modules/` - Code style, security, memory, workflow modules
- `providers.yml` - LLM routing config (SINGLE source of truth for providers)
- `config/agents.json` - Agent definitions
- `config/mcps.json` - MCP server configurations
- `artifacts/skills/registry.yml` - Skills metadata index
- `artifacts/skills/dynamic-registry.tsv` - Context packs index
- `docker/docker-compose.yml` - Infrastructure (Qdrant + Ollama)

**Important Paths**
```
~/.brain/daemon/cmd/braind/main.go       Main daemon entry point
~/.brain/daemon/internal/manager/        Business logic modules
~/.brain/cli/cmd/brain/main.go            CLI entry point
~/.brain/desktop/src/                    React components
~/.brain/docs/adr/                       Architecture Decision Records
~/.brain/artifacts/skills/registry.yml             Skills discovery index
~/.brain/rules/canonical.md              Development principles
```

---

## Frequently Enforced Warnings

### TypeScript/React

```
Warning: Unused variable 'x'
→ Fix: Remove unused variables, or prefix with _ if intentional (_x)

Warning: Unsafe any type
→ Fix: Add explicit types, enable noImplicitAny

Warning: Missing dependency in useEffect
→ Fix: Add to dependency array, or move inside effect
```

### Go

```
Warning: Unused import
→ Fix: Remove unused imports, gofmt will fix automatically

Warning: Unhandled error
→ Fix: Check error explicitly: if err != nil { return err }

Warning: Function too complex
→ Fix: Extract logic into helper functions, reduce nesting
```

### Common Lint Issues

- **Indentation**: Use 2-space or 4-space consistently (project uses 2 for TS, 4 for Go)
- **Line length**: Max 120 characters for code, 80 for comments
- **Naming**: CamelCase for types/classes, camelCase for variables, SCREAMING_SNAKE_CASE for constants
- **File names**: snake_case for JS/TS files imported as modules, PascalCase for React components

---

## Handling Special Scenarios

### Adding a New Skill

1. **Create folder**: `skills/<skill-id>/`
2. **Create SKILL.md**: Define skill interface, metadata
3. **Update registry.yml**: Add entry with id, name, type, path
4. **Wire daemon**: Add to SkillsRegistry.Load()
5. **Add CLI command**: `brain skills list` shows it
6. **Add UI component**: Dashboard displays new skill
7. **Test all 3 surfaces**: curl daemon → CLI → UI

### Modifying API Response Format

1. **Daemon first**: Update Go struct + handler
2. **CLI next**: Update CLI parser to handle new format
3. **UI last**: Update React component + TypeScript types
4. **Test round-trip**: Edit via UI → curl to verify → CLI reads it

### Adding a New Environment Variable

1. **Add to .env.example** with placeholder
2. **Document in README** and this file
3. **Validate in config loader** with helpful error if missing
4. **Use os.Getenv()** with fallback logic if optional
5. **Log at startup** what value was loaded (not the value itself if secret)

---

## Quick Reference: What Success Looks Like

- ✓ Code compiles (Go: `go build ./...`, TS: `tsc --noEmit`)
- ✓ Tests pass (`go test ./...`, `bun run test`)
- ✓ No lint warnings (gofmt, eslint, TypeScript strict)
- ✓ Feature works in all 3 surfaces (daemon API, CLI, UI)
- ✓ Commits follow Conventional Commits format
- ✓ Documentation updated (README, ADR if major decision)
- ✓ No hardcoded secrets
- ✓ Error handling is explicit
- ✓ Code is simple and readable
- ✓ Decision documented if architectural

---

## Where to Find More Info

- **Architecture**: See docs/adr/ for Architecture Decision Records
- **Code Standards**: See rules/modules/code-style.md
- **Security**: See rules/modules/security.md and guardian/
- **Workflow**: See rules/modules/workflow.md
- **Memory Protocol**: See rules/modules/memory.md
- **Agents**: See agents/ folder for prompt definitions
- **Skills**: See artifacts/skills/registry.yml for available skills

---

## This Document is Living

This file reflects the current state as of 2026-04-03. If you discover:
- A pattern that's repeated 3+ times, save as RuleCandidate in memory
- A rule that should change, propose it and update this file
- A contradiction between this and code behavior, report it

The brain learns from itself.
