# Brain - Development Kickoff Prompt

## Context

You are working on **Brain**, a daemon-centered control plane for AI engineering environments. Brain is the **single unified control point** for all AI coding tools (IDEs, CLIs, TUIs, desktop apps), providing shared artifacts, context, policy, memory, models, skills, MCPs, and agents.

**Repository**: `/mnt/main1tb/work/Personal/brain`
**Language**: Go 1.24.4
**Architecture**: Daemon-centered, multi-client, hierarchical policy, artifact-first
**Modules**: `core/` (shared library), `apps/cli/`, `apps/daemon/`, `apps/desktop/`, `apps/tui/`

## Canonical Documentation

All architecture decisions are in `docs/`. Read these in order before implementing anything:

1. `docs/architecture/brain-v2-target-architecture.md` — Overall vision
2. `docs/architecture/ide-cli-integration-strategy.md` — 16 IDEs/CLIs integration plan
3. `docs/architecture/implementation-gap-roadmap.md` — 9 gaps with Go code specs
4. `docs/architecture/unified-capability-enhancement-roadmap.md` — 15 capability domains, full implementation specs

## Current State

- ✅ Architecture fully documented (78 docs, 4,800+ lines)
- ✅ Go workspace: `core/`, `apps/cli/`, `apps/daemon/`
- ✅ Core has: artifact paths, runtime basics
- ✅ Daemon has: 29 Go files, foundational infrastructure
- ✅ CLI has: 7 Go files, skeleton
- ❌ **Nothing implemented yet** from the capability roadmap

## Implementation Roadmap (6 Phases)

### Phase 1: Foundation (Weeks 1-4) — START HERE
1. **Observability Stack** (GAP-002): OpenTelemetry, Prometheus, structured logging, health checks
2. **Artifact Registry Enhanced**: Dependencies, version resolution, compatibility matrix, usage analytics, hot reload, rollback
3. **Token Efficiency Engine**: Multi-tier cache (exact + semantic), prompt caching, context compaction, cost tracking
4. **Context Compiler**: 12-layer bundle assembly, progressive disclosure, optimization techniques

### Phase 2: Intelligence (Weeks 5-8)
5. **Model Capability Tier Router** (GAP-004): 3-tier routing, cost estimation, budget enforcement, fallback chains
6. **Context Curator** (GAP-003): Deduplication, compaction, promotion, cleanup, autoDream (KAIROS-inspired)
7. **Memory Sync** (GAP-005): Cloud sync, conflict resolution (5 strategies), audit trail, encryption
8. **MCP Server Hub**: First 5 official servers (filesystem, git, github, terminal, knowledge), connection manager
9. **Governance Engine**: RBAC + ABAC + OPA/Rego, policy as code, audit trail

### Phase 3: Orchestration (Weeks 9-12)
10. **Agent Delegation Graph** (GAP-001): DAG, 3 delegation modes (fork/teammate/worktree), budgets, fallback chains
11. **Agent Pool Manager**: 9 specialized agents, auto-scaling, capacity management
12. **DAG Workflow Orchestration**: Dependency-aware scheduling, parallel execution, 6 pre-built workflows
13. **Skill Registry + Marketplace**: Discovery, 8-point security scan, install, compatibility check
14. **Pre-built Workflows**: feature-dev, bug-fix, refactor, code-review, migration, full-release

### Phase 4: Self-Improvement (Weeks 13-16)
15. **AutoEvolve Engine**: Monitor (telemetry), Analyze (patterns), Propose (improvements), Apply (human-in-loop)
16. **Review & Apply UI**: Web + CLI interface for AutoEvolve recommendations
17. **Auto-generated Skill Proposals**: LLM-based skill generation from search patterns
18. **Token Waste Analysis**: Identify and report optimization opportunities
19. **Cost Optimization Engine** (GAP-007): Dashboards, budgets, alerts, per-surface tracking

### Phase 5: Client Surfaces (Weeks 17-22)
20. **VS Code Extension**: TypeScript, @brain chat participant, custom views, file watchers
21. **Brain Desktop**: Tauri + React, artifact/context/policy views, real-time events
22. **ACP Server**: Zed + JetBrains integration
23. **Qwen Code, Codex CLI, OpenCode, Continue.dev** integrations
24. **Brain TUI** (GAP-009)

### Phase 6: Advanced (Weeks 23+)
25. **Agent Teams Auto-Scaling** (GAP-008)
26. **Future Integration SDK** (Go, TypeScript, Rust, Python)
27. **Enterprise Features**: SSO, SCIM, compliance reports
28. **Multi-region Deployment**
29. **Collaborative Features**: Shared context, pair programming

## Implementation Rules

1. **Read docs first**: Always read relevant architecture docs before writing code
2. **Follow Go conventions**: Production code only in Go, no shell/Python for runtime paths
3. **Test-driven**: Write tests before or alongside implementation
4. **Observability first**: Every module emits structured logs, metrics, and traces
5. **API-first**: Define daemon HTTP API endpoints before implementing
6. **Backward compatible**: No breaking changes to existing artifact paths or APIs
7. **Security by default**: Input validation, no hardcoded secrets, least privilege
8. **Small commits**: Each commit is a logical, reviewable unit

## How to Work

For each capability:

1. Read the relevant section in `unified-capability-enhancement-roadmap.md`
2. Read the relevant GAP in `implementation-gap-roadmap.md`
3. Implement the Go code following the specs in those docs
4. Write unit tests (>80% coverage on core logic)
5. Write integration tests
6. Run `go vet`, `golangci-lint`, `go test ./...`
7. Update docs if implementation reveals design changes
8. Commit with clear message

## Starting Point

**Begin with Phase 1, Item 1: Observability Stack**

This is the foundation for everything else. Without observability, you cannot debug, monitor, or validate any other capability.

Location: `core/observability/`
Spec: See `implementation-gap-roadmap.md` → GAP-002 and `unified-capability-enhancement-roadmap.md` → Token Efficiency section

After observability, proceed in order through Phase 1.

## Key Data Structures

All canonical data structures are defined in the architecture docs. The most important:

- **Artifact Envelope**: `docs/architecture/ide-cli-integration-strategy.md` (standardized skill format)
- **Context Bundle**: Same doc (JSON structure all surfaces consume)
- **Policy Resolution**: Same doc (JSON structure all surfaces enforce)
- **Delegation Graph**: `docs/architecture/implementation-gap-roadmap.md` (GAP-001)
- **Agent Pool**: `docs/architecture/unified-capability-enhancement-roadmap.md` (Agent Pool Architecture)
- **Workflow DAG**: Same doc (DAG-Based Workflow Orchestration)

## Success Criteria

After completing all phases, Brain must:

1. Every IDE/CLI has the same capabilities via daemon
2. Zero duplication: install once, use everywhere
3. >50% token cost reduction through optimizations
4. AutoEvolve proposes and applies weekly improvements
5. 100% policy compliance across all surfaces
6. Multi-agent workflows 3x faster than single-agent
7. New IDEs integrate in <1 day via SDK
8. Full cost visibility per surface, user, workspace
9. Cross-surface memory continuity
10. All capabilities pass: unit tests, integration tests, security review, linting, docs

## Begin

Start by reading the architecture docs, then implement Phase 1 in order. Ask for clarification if architecture is ambiguous. Do not skip phases. Do not implement out of order unless a dependency requires it.
