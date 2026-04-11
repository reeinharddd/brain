---
type: design-doc
id: unified-capability-enhancement-roadmap
title: Unified Capability Enhancement Roadmap
version: 1.0.0
status: active
date_created: 2026-04-11
language: en
category: architecture
relationships:
  - brain-v2-target-architecture
  - ide-cli-integration-strategy
  - implementation-gap-roadmap
  - capability-control-plane-roadmap
---

## Overview

This document defines the complete enhancement roadmap for Brain as the **single unified control plane** for all AI coding tools, IDEs, and CLIs.

Brain's mission: **One daemon to govern all AI development surfaces**. Every IDE, CLI, TUI, and desktop app reads from and writes to the same canonical state — artifacts, context, policy, memory, models, skills, MCPs, and agents. No duplication. No missing features on one surface. No inconsistent behavior. No resource waste.

This document extends the base architecture with capabilities drawn from deep market research across 20+ tools, frameworks, and emerging patterns in 2026.

## Core Philosophy

```
┌─────────────────────────────────────────────────────────────┐
│                    BRAIN DAEMON (Core)                       │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌───────┐ │
│  │Artifact │ │ Context │ │ Policy  │ │ Memory  │ │ Model │ │
│  │Registry │ │Compiler │ │ Engine  │ │ System  │ │Router │ │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └───┬───┘ │
│       │           │           │           │          │     │
│  ┌────┴───────────┴───────────┴───────────┴──────────┴───┐ │
│  │           Unified State & Event Bus                   │ │
│  └────────────────────────┬──────────────────────────────┘ │
│                           │                                │
│  ┌────────────────────────┴──────────────────────────────┐ │
│  │          Self-Improvement Agent (AutoEvolve)          │ │
│  │  Monitors usage → proposes improvements → applies     │ │
│  └───────────────────────────────────────────────────────┘ │
└───────────────────────────┬─────────────────────────────────┘
                            │ HTTP/MCP/ACP/WS
        ┌───────────────────┼───────────────────┐
        │           │       │       │           │
   ┌────┴───┐ ┌─────┴──┐ ┌─┴────┐ ┌┴──────┐ ┌─┴────┐
   │VS Code │ │ Claude │ │Cursor│ │ Qwen  │ │Codex │
   │   IDE  │ │  Code  │ │  IDE │ │  CLI  │ │ CLI  │
   └────────┘ └────────┘ └──────┘ └───────┘ └──────┘
   ┌────────┐ ┌────────┐ ┌──────┐ ┌───────┐ ┌──────┐
   │Windsurf│ │Continue│ │Cline │ │ Zed   │ │JetBr │
   │   IDE  │ │  IDE   │ │  IDE │ │  IDE  │ │ains │
   └────────┘ └────────┘ └──────┘ └───────┘ └──────┘
   ┌────────┐ ┌────────┐ ┌──────┐ ┌───────┐
   │OpenCode│ │ Gemini │ │Aider │ │Neovim │
   │  CLI   │ │  CLI   │ │ CLI  │ │Editor │
   └────────┘ └────────┘ └──────┘ └───────┘
```

**Rule**: Every capability flows through daemon. Every surface gets the same features. No surface-specific logic.

## Capability Inventory

Brain must provide these capabilities to **all** connected surfaces:

### 1. Artifact Management (Unified Registry)
### 2. Context Compilation & Optimization
### 3. Hierarchical Policy Resolution
### 4. Multi-Agent Orchestration & Pooling
### 5. Model Routing & Cost Optimization
### 6. Memory System (Cross-Session Continuity)
### 7. Skill Registry & Marketplace
### 8. MCP Server Hub
### 9. Self-Improvement Engine (AutoEvolve)
### 10. Token Efficiency Engine
### 11. Workflow Orchestration (DAG-based)
### 12. Governance & Access Control
### 13. Observability & Telemetry
### 14. Sync & Collaboration
### 15. Future Integration Framework

---

## 1. Artifact Management (Enhanced)

**Current**: Basic artifact envelope with lifecycle states.

**Enhancement**: Full registry with dependency tracking, version resolution, and cross-artifact relationships.

```go
// core/artifacts/registry.go

type ArtifactRegistry struct {
    artifacts    map[ArtifactKey]*ArtifactRecord
    dependencies map[ArtifactKey][]ArtifactKey  // artifact → depends-on
    reverseDeps  map[ArtifactKey][]ArtifactKey  // artifact → required-by
    indexes      map[string]ArtifactIndex       // kind → index
}

type ArtifactRecord struct {
    Envelope       ArtifactEnvelope
    Dependencies   []ArtifactDependency
    Compatibility  CompatibilityMatrix
    UsageMetrics   UsageMetrics
    VersionHistory []VersionEntry
}

type ArtifactDependency struct {
    Kind        string   // skill, mcp, rule, agent
    ID          string
    VersionReq  string   // semver constraint: ">=1.0.0, <2.0.0"
    Optional    bool
    Description string
}

type CompatibilityMatrix struct {
    ClaudeCode  VersionRange
    CodexCLI    VersionRange
    Cursor      VersionRange
    VSCode      VersionRange
    GeminiCLI   VersionRange
    OpenCode    VersionRange
    // ... all supported surfaces
}

type UsageMetrics struct {
    TotalActivations    int
    LastActivated       time.Time
    AvgSessionDuration  time.Duration
    SurfacesUsed        map[string]int  // surface → activation count
    SuccessRate         float64
    FailurePatterns     []FailurePattern
}
```

**New Capabilities**:
- **Dependency Resolution**: Skills can depend on other skills, MCPs, rules. Daemon resolves full dependency graph before activation.
- **Version Compatibility Matrix**: Each artifact declares which surfaces/versions it works with. Daemon rejects incompatible installs.
- **Usage Analytics**: Track which artifacts are used, by which surfaces, how often. Feeds AutoEvolve engine.
- **Hot Reload**: Artifacts can be updated without daemon restart. Surfaces receive live update events.
- **Rollback**: Every artifact change is versioned. One-command rollback to any previous state.

**Daemon API**:
```
POST /api/v1/artifacts/install
  Body: { "kind": "skill", "id": "code-refactoring", "version": ">=1.0.0" }
  Resolves: dependencies, compatibility, downloads if needed

POST /api/v1/artifacts/upgrade
  Body: { "kind": "skill", "id": "code-refactoring", "strategy": "latest|compatible|pinned" }

GET /api/v1/artifacts/{kind}/{id}/dependencies
  Returns: Full dependency tree with resolution status

GET /api/v1/artifacts/{kind}/{id}/usage
  Returns: Usage metrics, surfaces used, success rate

POST /api/v1/artifacts/{kind}/{id}/rollback
  Body: { "to_version": "1.2.0" }
```

---

## 2. Context Compilation & Optimization (Enhanced)

**Current**: Context bundle design defined, not implemented.

**Enhancement**: Full context engineering pipeline with all industry-best techniques.

### Context Layers (Compiled in Order)

```
Layer 0: Hard Policy & Safety      (never compressed, always included)
Layer 1: Identity & Security       (org/user identity, trust boundaries)
Layer 2: Org Baseline              (org-wide conventions, shared patterns)
Layer 3: User Baseline             (personal preferences, coding style)
Layer 4: Workspace Context         (project architecture, conventions)
Layer 5: Project Context           (module structure, key decisions)
Layer 6: Task-Local Context        (current goal, active files, recent changes)
Layer 7: Active Skills             (progressive disclosure: frontmatter → full on trigger)
Layer 8: Active MCP Tools          (tool definitions, progressive disclosure)
Layer 9: Memory - Structured       (user preferences, past decisions)
Layer 10: Memory - Semantic        (Qdrant similarity matches, recency-weighted)
Layer 11: Memory - Episodic        (recent session events, temporal context)
Layer 12: Runtime Ephemeral        (current session state, conversation history)
```

### Optimization Techniques (All Applied by Daemon)

| Technique | Savings | Implementation |
|-----------|---------|----------------|
| **Just-in-time retrieval** | 40-60% | Fetch only needed data per step, not full repo dumps |
| **Progressive disclosure** | 30-50% | Load skill frontmatter first, full content only on trigger |
| **Prompt cache breakpoints** | 90% on cached tokens | Stable prefix (system + tools) first, variable content last |
| **Chain of Draft (CoD)** | 92.4% on reasoning | ~5-word step drafts instead of verbose chain-of-thought |
| **Server-side summarization** | 60-80% | Use provider compaction APIs for long sessions |
| **Semantic caching** | 100% on cache hit | Match semantically similar queries, skip API call entirely |
| **Differential outputs** | 50-70% | Request diffs, not full rewrites |
| **Output format optimization** | 50% | YAML/TSV instead of JSON for structured data |
| **On-demand tool loading** | 85% tool overhead | Don't load idle MCP/tool definitions |
| **Context resets between phases** | 30-50% | Separate discovery → implementation → verification |
| **Algorithmic compression (LLMLingua)** | 20x | Compress prompts while preserving accuracy |
| **KV-cache aware assembly** | 15-25% | Order content to maximize KV-cache reuse |

### Context Compiler Implementation

```go
// core/context/compiler.go

type ContextCompiler struct {
    config       CompilerConfig
    cache        PromptCache         // Multi-tier cache
    retriever    ContextRetriever    // Just-in-time retrieval
    compactor    ContextCompactor    // Summarization + pruning
    costTracker  *CostTracker        // Token counting and estimation
}

type CompiledBundle struct {
    Layers        []ContextLayer
    TotalTokens   int
    TokenLimit    int
    Utilization   float64
    CacheHitRate  float64
    Optimizations []AppliedOptimization
    CostEstimate  float64
}

type AppliedOptimization struct {
    Type         string  // "progressive_disclosure", "semantic_cache", "compaction", ...
    Layer        string
    SavingsTokens int
    SavingsUSD   float64
    AccuracyRisk float64  // 0.0-1.0, how much accuracy might be affected
}

func (c *ContextCompiler) Compile(ctx context.Context, req CompileRequest) (*CompiledBundle, error) {
    bundle := &CompiledBundle{}
    
    // Phase 1: Assemble mandatory layers (0-1, never compressed)
    mandatory, err := c.assembleMandatory(req)
    if err != nil {
        return nil, err
    }
    bundle.Layers = append(bundle.Layers, mandatory...)
    
    // Phase 2: Check semantic cache
    if cached := c.cache.LookupSemantic(req.Query); cached != nil {
        return cached, nil  // 100% savings on this request
    }
    
    // Phase 3: Just-in-time retrieval for remaining layers
    for layerNum := 2; layerNum <= 12; layerNum++ {
        layer, err := c.retriever.Retrieve(ctx, req, layerNum)
        if err != nil {
            continue
        }
        
        // Apply optimization based on remaining budget
        remaining := req.TokenLimit - bundle.TotalTokens
        if layer.TokenCount > remaining {
            layer = c.compactor.Compress(layer, remaining)
        }
        
        // Progressive disclosure: inject frontmatter, full content on trigger
        if layer.SupportsProgressiveDisclosure {
            layer = layer.ToSummary()  // Replace with summary, full on demand
        }
        
        bundle.Layers = append(bundle.Layers, layer)
    }
    
    // Phase 4: Optimize ordering for KV-cache efficiency
    bundle.Layers = c.optimizeForKVCaching(bundle.Layers)
    
    // Phase 5: Calculate totals and estimates
    bundle.TotalTokens = c.countTokens(bundle.Layers)
    bundle.Utilization = float64(bundle.TotalTokens) / float64(req.TokenLimit)
    bundle.CostEstimate = c.costTracker.Estimate(bundle, req.Model)
    
    // Phase 6: Store in semantic cache
    c.cache.StoreSemantic(req.Query, bundle)
    
    return bundle, nil
}
```

**Daemon API**:
```
POST /api/v1/context/compile
  Body: { "scope_chain": [...], "task": "refactor auth module", "model": "gpt-4", "token_limit": 8000 }
  Returns: CompiledBundle with all optimizations applied

GET /api/v1/context/optimizations
  Returns: Available optimization techniques and their effectiveness rates

POST /api/v1/context/cache/clear
  Clears semantic cache

GET /api/v1/context/cache/stats
  Returns: Cache hit rate, size, savings
```

---

## 3. Multi-Agent Orchestration & Pooling (NEW)

**Source**: Claude Code (Fork/Teammate/Worktree), AutoGen (conversational), CrewAI (role-based teams), OpenClaw (specialized agents), Flyte 2.0 (DAG-based parallel execution)

**Vision**: Brain manages a **pool of specialized agents** that any IDE/CLI can request. Instead of each tool having its own generic agent, Brain provides the right agent for the job.

### Agent Pool Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Agent Pool Manager                     │
│                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │ Architect   │  │ Builder     │  │ Reviewer        │ │
│  │ (Tier 3)    │  │ (Tier 2)    │  │ (Tier 2)        │ │
│  │ System design│ │ Code gen    │  │ Code review     │ │
│  └─────────────┘  └─────────────┘  └─────────────────┘ │
│                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │ Debugger    │  │ Refactorer  │  │ Tester          │ │
│  │ (Tier 2)    │  │ (Tier 2)    │  │ (Tier 2)        │ │
│  │ Debug flows │  │ Refactor    │  │ Generate tests  │ │
│  └─────────────┘  └─────────────┘  └─────────────────┘ │
│                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │ Documenter  │  │ Migrator    │  │ Security Auditor│ │
│  │ (Tier 1-2)  │  │ (Tier 2)    │  │ (Tier 3)        │ │
│  │ Write docs  │  │ Migrate code│  │ Security review │ │
│  └─────────────┘  └─────────────┘  └─────────────────┘ │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │           Orchestrator (Supervisor)               │  │
│  │  Routes tasks → agents, manages DAG, handles      │  │
│  │  failures, merges results, tracks costs           │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Agent Definition

```yaml
# Agent pool definition
apiVersion: brain/v1
kind: agent
id: architect-core
name: Architect Core
version: 1.0.0
role: architect  # architect | builder | reviewer | debugger | refactore | tester | documenter | migrator | security-auditor
tier: 3  # capability tier required
model_requirements:
  min_capability_tier: 3
  preferred_models: [claude-opus-4.6, gpt-4]
  max_tokens: 4000
  requires_tools: true

capabilities:
  - system_design
  - architecture_review
  - tech_debt_assessment
  - api_design
  - data_modeling

constraints:
  never_writes_code: true  # Only produces plans, specs, recommendations
  output_format: markdown  # Structured output
  max_concurrent_tasks: 3  # Parallel capacity

pool_config:
  min_instances: 1
  max_instances: 5  # Auto-scale up to 5 under load
  idle_timeout: 5m  # Scale down after 5 minutes idle
  queue_capacity: 20  # Max queued tasks

cost_budget:
  max_per_task_usd: 0.50
  max_daily_usd: 10.00
```

### Orchestration Patterns

Brain supports 5 orchestration patterns, selected automatically based on task complexity:

| Pattern | When Used | Example |
|---------|-----------|---------|
| **Direct LLM** | Simple, single-step queries | "What does this function do?" |
| **ReAct** | Dynamic, exploratory tasks | Debugging, navigating unfamiliar code |
| **Plan-and-Execute** | Structured, multi-step tasks | Refactoring, scaffolding new features |
| **Orchestrator-Worker** | Parallelizable subtasks | Code review + test gen + doc gen simultaneously |
| **Hierarchical Teams** | Complex, multi-domain workflows | Full feature: design → build → test → review → deploy |

### DAG-Based Workflow Orchestration

```go
// core/orchestration/dag.go

type WorkflowDAG struct {
    ID       string
    Name     string
    Nodes    map[string]*WorkflowNode
    Edges    map[string][]string  // node → depends-on
    Parallel bool  // Allow parallel execution of independent nodes
}

type WorkflowNode struct {
    ID       string
    Agent    string  // Agent pool member to execute
    Input    TaskInput
    Output   *TaskOutput  // Filled after execution
    Timeout  time.Duration
    Retry    RetryPolicy
    Status   NodeStatus  // pending | running | completed | failed | skipped
}

// Execution engine runs DAG with dependency-aware scheduling
func (e *ExecutionEngine) RunDAG(ctx context.Context, dag *WorkflowDAG) (*ExecutionResult, error) {
    queue := e.findReadyNodes(dag)  // Nodes with all dependencies met
    
    for len(queue) > 0 {
        // Execute ready nodes in parallel (up to agent pool capacity)
        results := e.executeParallel(ctx, queue)
        
        // Update DAG with results
        for nodeID, result := range results {
            dag.Nodes[nodeID].Status = result.Status
            dag.Nodes[nodeID].Output = result.Output
            
            if result.Status == "failed" {
                if dag.Nodes[nodeID].Retry.ShouldRetry() {
                    queue = append(queue, nodeID)  // Retry
                } else {
                    e.markDownstreamSkipped(dag, nodeID)  // Skip dependents
                }
            }
        }
        
        // Find next batch of ready nodes
        queue = e.findReadyNodes(dag)
    }
    
    return e.assembleResult(dag), nil
}
```

**Pre-built Workflows** (available to all surfaces):

| Workflow | DAG | Use Case |
|----------|-----|----------|
| **feature-dev** | architect → (builder + tester) → reviewer → documenter | New feature development |
| **bug-fix** | debugger → builder → tester → reviewer | Bug investigation and fix |
| **refactor** | refactore → (builder + tester) → reviewer | Code refactoring |
| **code-review** | reviewer + security-auditor (parallel) | Comprehensive code review |
| **migration** | migrator → tester → documenter | Framework/library migration |
| **full-release** | feature-dev → migration → code-review → documenter | End-to-end release prep |

**Daemon API**:
```
POST /api/v1/agents/pool/list
  Returns: All available agents with status, capacity, current load

POST /api/v1/agents/execute
  Body: { "agent": "architect-core", "task": {...}, "budget_usd": 0.50 }
  Returns: { "execution_id": "...", "result": {...} }

POST /api/v1/workflows/execute
  Body: { "workflow": "feature-dev", "input": {...}, "budget_usd": 5.00 }
  Returns: { "execution_id": "...", "status": "running" }

GET /api/v1/workflows/{execution_id}/status
  Returns: DAG execution progress, completed/failed nodes

POST /api/v1/workflows/{execution_id}/cancel
  Cancels running workflow
```

---

## 4. Skill Registry & Marketplace (NEW)

**Source**: skills.sh, ClawHub (OpenClaw, 13,700+ skills), Agensi marketplace, Anthropic skills roadmap

**Vision**: Brain operates a **canonical skill registry** that all surfaces share. Install once, use everywhere. Skills are security-scanned, compatibility-checked, and usage-tracked.

### Registry Architecture

```
┌─────────────────────────────────────────────────────┐
│              Brain Skill Registry                    │
│                                                     │
│  ┌──────────────┐  ┌───────────────┐  ┌──────────┐ │
│  │ Discovery    │  │ Security Scan │  │ Install  │ │
│  │ - Search     │  │ - 8-point check│  │ - Download│ │
│  │ - Browse     │  │ - Static analysis│ │ - Verify │ │
│  │ - Ratings    │  │ - Secret detection│ │ - Deploy │ │
│  │ - Requests   │  │ - Injection check │ │ - Activate│ │
│  └──────────────┘  └───────────────┘  └──────────┘ │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │            Skill Catalog (Local + Cloud)       │  │
│  │  - Official Brain skills (curated)            │  │
│  │  - Community skills (security-reviewed)        │  │
│  │  - Enterprise skills (private registry)        │  │
│  │  - Auto-generated skills (by AutoEvolve)       │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### 8-Point Security Scan (Every Skill Must Pass)

| Check | What It Detects | Severity |
|-------|----------------|----------|
| **File Structure** | Unexpected files, hidden files, symlink attacks | Critical |
| **Dangerous Commands** | rm -rf, chmod 777, curl | pip, eval() | Critical |
| **Hardcoded Secrets** | API keys, tokens, passwords in skill files | Critical |
| **Env Variable Harvesting** | Reads sensitive env vars (AWS_SECRET_KEY, etc.) | High |
| **Network Access** | Outbound HTTP/HTTPS to unknown endpoints | High |
| **Obfuscation** | Base64-encoded scripts, minified content, unicode tricks | High |
| **Prompt Injection** | Instructions that override org hard policy | Critical |
| **Privilege Escalation** | Requests permissions beyond declared scope | Critical |

### Skill Metadata

```yaml
apiVersion: brain/v1
kind: skill
id: code-refactoring
name: code-refactoring
version: 1.2.0

# Marketplace metadata
marketplace:
  category: development
  tags: [refactoring, code-quality, clean-code]
  rating: 4.8
  total_installs: 12500
  creator: brain-team
  license: MIT
  price: free  # free | paid (with price_usd)
  
# Compatibility
compatibility:
  min_brain_version: "0.1.0"
  min_capability_tier: 2
  surfaces:
    claude_code: ">=1.0.0"
    codex_cli: ">=1.0.0"
    cursor: ">=1.0.0"
    vscode: ">=1.0.0"
    gemini_cli: ">=1.0.0"
    opencode: ">=1.0.0"
    # ... all surfaces

# Security
security_scan:
  passed_at: "2026-04-11T10:00:00Z"
  checks:
    file_structure: pass
    dangerous_commands: pass
    hardcoded_secrets: pass
    env_harvesting: pass
    network_access: pass
    obfuscation: pass
    prompt_injection: pass
    privilege_escalation: pass

# Usage
usage:
  total_activations: 45000
  success_rate: 0.96
  avg_token_cost: 350
  avg_duration: 2m30s
  top_surfaces:
    vscode: 35%
    claude_code: 25%
    cursor: 20%
    qwen_cli: 15%
    other: 5%
```

**Daemon API**:
```
GET /api/v1/skills/search?q=refactoring&category=development
  Returns: Matching skills with ratings, compatibility, install count

GET /api/v1/skills/{id}/versions
  Returns: All versions with changelog

POST /api/v1/skills/install
  Body: { "id": "code-refactoring", "version": ">=1.0.0", "scope": "workspace" }
  Runs: Security scan → compatibility check → install → activate

POST /api/v1/skills/scan
  Body: { "path": "/path/to/skill" }
  Returns: 8-point security scan results

GET /api/v1/skills/requests
  Returns: Most requested skills (community demand signals)
```

---

## 5. MCP Server Hub (NEW)

**Source**: MCP best practices, 2026 MCP server ecosystem, enterprise deployment patterns

**Vision**: Brain operates as a **unified MCP hub**. Instead of each IDE connecting to 10 different MCP servers, they all connect to Brain, which proxies and manages all MCP connections.

### MCP Hub Architecture

```
┌──────────────────────────────────────────────────────────┐
│                   Brain MCP Hub                          │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │              MCP Server Registry                    │ │
│  │                                                     │ │
│  │  Official (Brain-maintained):                      │ │
│  │  ├── filesystem      (read/write project files)    │ │
│  │  ├── git             (git operations, history)     │ │
│  │  ├── github          (PRs, issues, reviews)        │ │
│  │  ├── database        (PostgreSQL, SQLite queries)  │ │
│  │  ├── browser         (Playwright, web testing)     │ │
│  │  ├── terminal        (shell execution, sandboxed)  │ │
│  │  ├── semgrep         (security scanning)           │ │
│  │  ├── playwright      (E2E testing)                 │ │
│  │  └── knowledge       (Qdrant semantic memory)     │ │
│  │                                                     │ │
│  │  Community (security-reviewed):                    │ │
│  │  ├── figma           (design system integration)   │ │
│  │  ├── linear          (issue tracking)              │ │
│  │  ├── stripe          (payment APIs)                │ │
│  │  └── ...             (extensible)                  │ │
│  │                                                     │ │
│  │  Enterprise (private):                             │ │
│  │  ├── internal-api   (company-specific services)    │ │
│  │  ├── internal-db    (company databases)            │ │
│  │  └── ...                                           │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │            Connection Manager                        │ │
│  │  - Multiplexes: 1 Brain MCP → many clients         │ │
│  │  - Rate limits: per-tool, per-client, per-workspace │ │
│  │  - Circuit breakers: graceful degradation           │ │
│  │  - Health checks: auto-reconnect on failure         │ │
│  └────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

### MCP Tool Design (Best Practices Applied)

Every MCP tool exposed by Brain follows these rules:

1. **Bounded Context**: Each server has a clear, narrow purpose
2. **Strict Schemas**: LLM-friendly input/output with clear types
3. **Machine-Verifiable**: Errors are structured, not free-text
4. **Rate Limited**: Prevents abuse, respects upstream limits
5. **Sandboxed**: Tool execution isolated, no host compromise
6. **Observable**: Every invocation logged with latency, status, result size
7. **Versioned**: Additive versioning, no breaking changes

### Pre-built MCP Servers

| Server | Tools | Resources | Use Case |
|--------|-------|-----------|----------|
| **brain-filesystem** | `read_file`, `write_file`, `list_dir`, `search` | `file://` URIs | File operations for all surfaces |
| **brain-git** | `git_log`, `git_diff`, `git_blame`, `git_status` | `git://` refs | Git history and operations |
| **brain-github** | `create_pr`, `review_pr`, `list_issues`, `add_comment` | `gh://` refs | GitHub integration |
| **brain-database** | `query`, `schema`, `migrate` | `db://` tables | Database operations |
| **brain-browser** | `navigate`, `screenshot`, `click`, `type`, `evaluate` | `http://` pages | Web testing and scraping |
| **brain-terminal** | `execute`, `output`, `cancel` | N/A | Sandshell shell execution |
| **brain-semgrep** | `scan`, `rule`, `fix` | `rule://` definitions | Security scanning |
| **brain-playwright** | `test`, `trace`, `report` | `test://` specs | E2E testing |
| **brain-knowledge** | `search`, `store`, `recall`, `forget` | `mem://` entries | Semantic memory via Qdrant |
| **brain-context** | `get_bundle`, `compile`, `optimize` | `ctx://` bundles | Context bundle access |
| **brain-policy** | `resolve`, `check`, `explain` | `pol://` rules | Policy resolution |

**Daemon API**:
```
GET /api/v1/mcp/servers
  Returns: All registered MCP servers with status, health, tool count

GET /api/v1/mcp/servers/{id}/tools
  Returns: Tools exposed by server with schemas

POST /api/v1/mcp/servers/register
  Body: { "id": "my-server", "command": "my-mcp-server", "args": [...] }
  Registers new MCP server (stdio or HTTP)

POST /api/v1/mcp/servers/{id}/call
  Body: { "tool": "read_file", "arguments": { "path": "..." } }
  Calls MCP tool through Brain proxy
```

---

## 6. Self-Improvement Engine: AutoEvolve (NEW)

**Source**: KAIROS autoDream, Hindsight (retain/recall/reflect), recursive self-improvement loops, self-evolving agents

**Vision**: Brain runs a **periodic self-improvement agent** that monitors its own usage, identifies gaps, proposes improvements, and applies them (with approval). This is the meta-agent that makes Brain better over time.

### AutoEvolve Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  AutoEvolve Agent                        │
│                                                          │
│  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐ │
│  │ MONITOR      │  │ ANALYZE       │  │ PROPOSE      │ │
│  │ (Continuous) │  │ (Periodic)    │  │ (On-demand)  │ │
│  │              │  │               │  │              │ │
│  │ - Tool usage │  │ - Usage       │  │ - New skills │ │
│  │ - Skill hits  │  │   patterns   │  │ - MCP servers│ │
│  │ - Failures   │  │ - Failure     │  │ - Rules      │ │
│  │ - Latency    │  │   patterns   │  │ - Agents     │ │
│  │ - Token cost │  │ - Token waste│  │ - Optimizations│ │
│  │ - Gaps       │  │ - Missing     │  │ - Deprecations│ │
│  │              │  │   artifacts  │  │              │ │
│  └───────┬──────┘  └───────┬───────┘  └──────┬───────┘ │
│          │                 │                  │         │
│          └─────────────────┼──────────────────┘         │
│                            │                            │
│                    ┌───────┴────────┐                   │
│                    │    REVIEW      │                   │
│                    │  (Human-in-loop)│                   │
│                    │                │                   │
│                    │ - Show report  │                   │
│                    │ - Explain why  │                   │
│                    │ - Recommend    │                   │
│                    │ - Get approval │                   │
│                    └───────┬────────┘                   │
│                            │                            │
│                    ┌───────┴────────┐                   │
│                    │    APPLY       │                   │
│                    │  (Approved only)│                   │
│                    │                │                   │
│                    │ - Install skill│                   │
│                    │ - Update config│                   │
│                    │ - Add MCP      │                   │
│                    │ - Modify rules │                   │
│                    │ - Rollback if  │                   │
│                    │   regression   │                   │
│                    └────────────────┘                   │
└─────────────────────────────────────────────────────────┘
```

### Monitoring (Continuous)

```go
// core/autoevolve/monitor.go

type UsageTelemetry struct {
    Timestamp        time.Time
    Surface          string  // which IDE/CLI
    ActionType       string  // skill_used, mcp_called, policy_checked, ...
    ArtifactKind     string  // skill, mcp, rule, agent
    ArtifactID       string
    Success          bool
    Duration         time.Duration
    TokensUsed       int
    TokensWasted     int  // Unused context, failed retries
    ErrorType        string
    UserSatisfaction *int  // If surface provides feedback (1-5)
}

type GapDetection struct {
    Type         string  // missing_skill, missing_mcp, high_failure_rate, ...
    Evidence     []UsageTelemetry
    Confidence   float64  // 0.0-1.0
    Impact       string   // high | medium | low
    FirstSeen    time.Time
    Occurrences  int
}

// Monitor processes telemetry in real-time
func (m *Monitor) Process(telemetry UsageTelemetry) {
    m.accumulator.Add(telemetry)
    
    // Detect anomalies
    if !telemetry.Success && telemetry.ErrorType != "" {
        m.failureTracker.Record(telemetry)
    }
    
    // Detect missing artifacts
    if telemetry.ActionType == "skill_search" && telemetry.Success == false {
        m.gapDetector.RecordMissingSkill(telemetry)
    }
    
    // Detect token waste
    if telemetry.TokensWasted > telemetry.TokensUsed {
        m.wasteTracker.Record(telemetry)
    }
    
    // Detect underused artifacts
    if artifact := m.registry.Get(telemetry.ArtifactID); artifact != nil {
        artifact.UsageMetrics.Record(telemetry)
    }
}
```

### Analysis (Periodic, triggers like autoDream)

```go
// core/autoevolve/analyzer.go

type AnalysisReport struct {
    Period         time.Range
    TopSkills      []SkillUsage  // Most used skills
    FailedSkills   []SkillUsage  // Skills with high failure rate
    MissingSkills  []MissingSkill  // Search queries with no results
    TokenWaste     []TokenWasteFinding  // Where tokens are wasted
    MCPPatterns    []MCPPattern  // MCP usage trends
    PolicyViolations []PolicyViolation  // Policy enforcement stats
    Recommendations []Recommendation  // What to improve
}

type Recommendation struct {
    Type         string  // new_skill, new_mcp, update_skill, deprecate_skill, optimize_context
    Title        string
    Description  string
    Evidence     []UsageTelemetry
    Impact       string  // high | medium | low
    Effort       string  // trivial | small | medium | large
    Confidence   float64
    ProposedArtifact *ProposedArtifact  // If applicable
}

type ProposedArtifact struct {
    Kind    string  // skill, mcp, rule
    ID      string
    Draft   string  // Markdown draft of the artifact
    Source  string  // "auto-generated by AutoEvolve"
}

// Analyzer runs periodically (like autoDream) or on-demand
func (a *Analyzer) Analyze(ctx context.Context) (*AnalysisReport, error) {
    report := &AnalysisReport{Period: a.period}
    
    // Phase 1: Analyze skill usage
    report.TopSkills = a.topN(a.skillUsage, 10)
    report.FailedSkills = a.failureRate(a.skillUsage, 0.3)  // >30% failure
    
    // Phase 2: Detect missing skills
    missing := a.gapDetector.GetMissingSkills()
    for _, m := range missing {
        if m.SearchCount >= 5 {  // At least 5 searches for same missing skill
            report.MissingSkills = append(report.MissingSkills, m)
            
            // Auto-generate skill proposal
            if draft, err := a.proposeSkill(m); err == nil {
                report.Recommendations = append(report.Recommendations, Recommendation{
                    Type: "new_skill",
                    Title: fmt.Sprintf("Create skill: %s", m.Query),
                    ProposedArtifact: draft,
                    Confidence: 0.7,
                    Impact: "high",
                })
            }
        }
    }
    
    // Phase 3: Analyze token waste
    waste := a.findTokenWaste()
    for _, w := range waste {
        report.Recommendations = append(report.Recommendations, Recommendation{
            Type: "optimize_context",
            Title: fmt.Sprintf("Reduce token waste in %s", w.Surface),
            Description: w.Description,
            Impact: w.Impact,
            Confidence: w.Confidence,
        })
    }
    
    // Phase 4: MCP usage patterns
    report.MCPPatterns = a.analyzeMCPUsage()
    
    // Phase 5: Policy effectiveness
    report.PolicyViolations = a.analyzePolicyEnforcement()
    
    // Phase 6: Generate improvement proposals
    report.Recommendations = a.prioritizeRecommendations(report.Recommendations)
    
    return report, nil
}

// proposeSkill uses an LLM to draft a skill based on search patterns
func (a *Analyzer) proposeSkill(missing MissingSkill) (*ProposedArtifact, error) {
    // Use a Tier 3 model to generate skill content from search patterns
    prompt := fmt.Sprintf(`
Based on the following search queries that returned no results, 
draft a skill definition that would address this need:

Search queries (last 30 days):
%s

Total searches: %d
Top surfaces requesting this: %s

Draft a skill with:
- Clear name and description
- Step-by-step methodology
- Examples of usage
- Common mistakes to avoid
`, missing.Queries, missing.SearchCount, missing.TopSurfaces)

    content, err := a.llm.Generate(prompt, llm.Options{
        Model: "claude-opus-4.6",  // Use best model for skill generation
        MaxTokens: 4000,
    })
    if err != nil {
        return nil, err
    }
    
    return &ProposedArtifact{
        Kind:   "skill",
        ID:     slugify(missing.Query),
        Draft:  content,
        Source: "auto-generated by AutoEvolve",
    }, nil
}
```

### Review & Apply (Human-in-Loop)

```
┌─────────────────────────────────────────────────────┐
│              AutoEvolve Review UI                    │
│              (Web, CLI, IDE notification)            │
│                                                     │
│  📊 Weekly Improvement Report                        │
│  ─────────────────────────────                       │
│                                                     │
│  Usage Summary (Last 7 days):                        │
│  • 1,247 skill activations (↑12% from last week)    │
│  • 96% success rate (↑2%)                           │
│  • 45,000 tokens saved by optimizations (↑30%)      │
│  • 23 failed skill executions (↓15%)               │
│                                                     │
│  🔍 3 Improvement Proposals:                         │
│  ─────────────────────────────                       │
│                                                     │
│  1. [HIGH IMPACT] Create skill: "docker-compose-debug"│
│     Searched 18 times, no matching skill found.      │
│     Top surfaces: VS Code (12), Claude Code (6)     │
│     Auto-generated draft available for review.       │
│     [View Draft] [Approve] [Reject] [Modify]        │
│                                                     │
│  2. [MEDIUM] Optimize: "code-refactoring" context   │
│     Wasting ~200 tokens/session on unused examples.  │
│     Suggestion: Move examples to progressive         │
│     disclosure (load on-demand, not upfront).       │
│     Estimated savings: 3,000 tokens/day.            │
│     [View Analysis] [Apply] [Reject]               │
│                                                     │
│  3. [LOW] Deprecate: "legacy-js-patterns" skill     │
│     Not used in 30 days. All surfaces migrated to   │
│     TypeScript. Recommendation: archive skill.      │
│     [View Usage] [Archive] [Keep]                  │
│                                                     │
│  [Apply All Approved] [Review Individually]         │
└─────────────────────────────────────────────────────┘
```

**AutoEvolve Triggers**:

| Trigger | Frequency | Action |
|---------|-----------|--------|
| **Usage Telemetry** | Continuous | Accumulate metrics, detect anomalies |
| **Analysis Cycle** | Weekly (configurable) | Generate improvement report |
| **Idle Detection** | During user inactivity (KAIROS pattern) | Run autoDream-style consolidation |
| **Failure Threshold** | Immediate (>5 failures/hour for same artifact) | Alert + propose fix |
| **Gap Detection** | Daily | Identify missing skills/MCPs |
| **Token Waste Alert** | Daily | Identify optimization opportunities |
| **Manual** | On-demand | `/autoevolve analyze` |

**Daemon API**:
```
GET /api/v1/autoevolve/report
  Returns: Latest analysis report with recommendations

GET /api/v1/autoevolve/report/{id}
  Returns: Historical analysis report

POST /api/v1/autoevolve/apply
  Body: { "recommendation_ids": [1, 2], "auto_approve_low": true }
  Applies approved recommendations

POST /api/v1/autoevolve/analyze
  Triggers on-demand analysis

GET /api/v1/autoevolve/config
  Returns/Sets: Analysis frequency, thresholds, auto-approve rules
```

---

## 7. Token Efficiency Engine (NEW)

**Source**: Industry token optimization techniques, provider-specific APIs, prompt caching, KV-cache management

**Vision**: Brain tracks, optimizes, and minimizes token usage across ALL surfaces. Every API call goes through the efficiency engine, which applies all known optimization techniques.

### Efficiency Techniques Applied

```go
// core/efficiency/engine.go

type TokenEfficiencyEngine struct {
    cache        PromptCache       // Multi-tier: exact, semantic, KV
    compactor    ContextCompactor  // Compression algorithms
    router       ModelRouter       // Route to cheapest capable model
    tracker      CostTracker       // Real-time cost monitoring
}

type OptimizationResult struct {
    OriginalTokens int
    OptimizedTokens int
    SavingsPercent  float64
    SavingsUSD     float64
    Techniques      []AppliedTechnique
    AccuracyRisk   float64  // Risk that optimization affects quality
}

type AppliedTechnique struct {
    Name        string
    Description string
    TokensSaved int
    RiskLevel   string  // none | low | medium | high
}

func (e *TokenEfficiencyEngine) Optimize(ctx context.Context, req OptimizationRequest) (*OptimizationResult, error) {
    result := &OptimizationResult{
        OriginalTokens: e.countTokens(req.Input),
    }
    
    // Technique 1: Exact cache lookup (100% savings on hit)
    if cached := e.cache.LookupExact(req.Input); cached != nil {
        result.OptimizedTokens = 0
        result.SavingsPercent = 100
        result.Techniques = append(result.Techniques, AppliedTechnique{
            Name: "exact_cache_hit",
            Description: "Identical request found in cache",
            TokensSaved: result.OriginalTokens,
            RiskLevel: "none",
        })
        return result, nil
    }
    
    // Technique 2: Semantic cache lookup (skip API call)
    if cached := e.cache.LookupSemantic(req.Input, 0.95); cached != nil {
        result.OptimizedTokens = 0
        result.SavingsPercent = 100
        result.Techniques = append(result.Techniques, AppliedTechnique{
            Name: "semantic_cache_hit",
            Description: "Semantically similar request found in cache",
            TokensSaved: result.OriginalTokens,
            RiskLevel: "low",
        })
        return result, nil
    }
    
    // Technique 3: Prompt cache optimization (90% discount on cached prefix)
    optimized := e.optimizePromptCache(req)
    result.Techniques = append(result.Techniques, optimized.CacheOptimization...)
    
    // Technique 4: Context compaction (if over budget)
    if optimized.TokenCount > req.TokenLimit {
        compacted := e.compactor.Compact(optimized, req.TokenLimit)
        result.Techniques = append(result.Techniques, AppliedTechnique{
            Name: "context_compaction",
            Description: "Compressed context to fit token limit",
            TokensSaved: optimized.TokenCount - compacted.TokenCount,
            RiskLevel: "medium",
        })
        optimized = compacted
    }
    
    // Technique 5: Progressive disclosure (defer loading full content)
    optimized = e.applyProgressiveDisclosure(optimized)
    result.Techniques = append(result.Techniques, AppliedTechnique{
        Name: "progressive_disclosure",
        Description: "Deferred full content loading",
        TokensSaved: optimized.DeferredTokens,
        RiskLevel: "low",
    })
    
    // Technique 6: Chain of Draft (replace verbose reasoning)
    if req.Model.SupportsThinking {
        optimized = e.applyChainOfDraft(optimized)
        result.Techniques = append(result.Techniques, AppliedTechnique{
            Name: "chain_of_draft",
            Description: "Replaced verbose reasoning with ~5-word drafts",
            TokensSaved: optimized.ReasoningTokens * 0.924,
            RiskLevel: "low",
        })
    }
    
    // Technique 7: Model right-sizing (cheapest capable model)
    routeResult := e.router.Route(ctx, RouteRequest{
        MinCapabilityTier: req.MinTier,
        MaxTokens: optimized.TokenCount,
        BudgetUSD: req.BudgetUSD,
    })
    if routeResult.SelectedModel.CostPer1KInput < req.DefaultModel.CostPer1KInput {
        result.Techniques = append(result.Techniques, AppliedTechnique{
            Name: "model_right_sizing",
            Description: fmt.Sprintf("Routed to %s instead of %s", 
                routeResult.SelectedModel.ModelID, req.DefaultModel.ModelID),
            TokensSaved: 0,  // Same tokens, lower cost
            RiskLevel: "none",
        })
    }
    
    result.OptimizedTokens = optimized.TokenCount
    result.SavingsPercent = float64(result.OriginalTokens-result.OptimizedTokens) / float64(result.OriginalTokens) * 100
    result.SavingsUSD = e.calculateSavings(result)
    
    return result, nil
}
```

**Cost Dashboard** (available in all surfaces):

```
Token Efficiency Report (Last 7 days)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total tokens processed:     2,450,000
Total cost:                 $48.75
Avg cost/session:           $2.43

Optimization Savings:
  Exact cache hits:         340 requests  ($12.50 saved)
  Semantic cache hits:      180 requests   ($8.20 saved)
  Prompt caching:           890 requests   ($15.30 saved)
  Context compaction:       45 requests    ($3.10 saved)
  Progressive disclosure:   520 requests   ($4.80 saved)
  Model right-sizing:       120 requests   ($5.60 saved)
  Chain of Draft:           200 requests   ($2.90 saved)
  ─────────────────────────────────────────
  Total saved:                             $52.40 (51.8% reduction)

Top Cost Centers:
  1. VS Code → code-refactoring:  $12.30 (340k tokens)
  2. Claude Code → debug-session:  $8.50 (220k tokens)
  3. Cursor → feature-dev:         $6.80 (180k tokens)

Wasted Tokens (opportunity for improvement):
  1. Unused context in sessions:   45,000 tokens ($1.20)
  2. Failed retries (malformed):   12,000 tokens ($0.35)
  3. Stale session history:        28,000 tokens ($0.80)
```

**Daemon API**:
```
POST /api/v1/efficiency/optimize
  Body: { "input_tokens": 5000, "model": "gpt-4", "budget_usd": 0.10 }
  Returns: OptimizationResult with all techniques applied

GET /api/v1/efficiency/report
  Query: ?period=7d&surface=vscode
  Returns: Cost dashboard with savings breakdown

GET /api/v1/efficiency/cost-by-surface
  Returns: Cost breakdown per connected surface

POST /api/v1/efficiency/cache/clear
  Clears all caches

GET /api/v1/efficiency/cache/stats
  Returns: Cache hit rates, size, savings
```

---

## 8. Governance & Access Control (Enhanced)

**Source**: ABAC-style constraints, AI agent access control boundaries, enterprise governance

**Vision**: Brain enforces fine-grained access control across all surfaces. Every action is authorized by policy. Every surface respects the same boundaries.

### Access Control Model

```
┌──────────────────────────────────────────────────────┐
│              Governance Engine                         │
│                                                       │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐│
│  │ RBAC        │  │ ABAC         │  │ Policy       ││
│  │ (Role-based)│  │ (Attribute-  │  │ as Code      ││
│  │             │  │  based)      │  │              ││
│  │ admin       │  │ user.org=X  │  │ OPA/Rego     ││
│  │ developer   │  │ resource.   │  │ policies     ││
│  │ viewer      │  │  scope=Y    │  │              ││
│  │ agent       │  │ action.type │  │              ││
│  └──────┬──────┘  └──────┬───────┘  └──────┬───────┘│
│         │                │                 │        │
│         └────────────────┼─────────────────┘        │
│                          │                          │
│              ┌───────────┴────────────┐             │
│              │  Authorization Decision │             │
│              │  PERMIT | DENY | OBLIGE│             │
│              │  (with obligations)     │             │
│              └─────────────────────────┘             │
└──────────────────────────────────────────────────────┘
```

### Policy as Code (OPA/Rego)

```rego
# core/policy/rego/brain.rego

package brain

# Hard policy: No hardcoded secrets
deny[msg] {
    input.action.type == "artifact_create"
    input.artifact.kind == "skill"
    contains_secrets(input.artifact.content)
    msg := "Hardcoded secrets detected in artifact. Use environment variables or secret manager."
}

# Guarded policy: Model routing prefers local-first
oblige[msg] {
    input.action.type == "model_route"
    not input.policy.model_preference == "local-first"
    msg := "Policy requires local-first model routing. Override needs approval."
}

# Soft policy: Code style prefers gofmt
advise[msg] {
    input.action.type == "code_generate"
    input.workspace.language == "go"
    not uses_gofmt(input.artifact.content)
    msg := "Code does not follow gofmt. Run 'gofmt -w' to comply."
}

# Access control: Only developers can modify artifacts
allow {
    input.user.role == "developer"
    input.action.type == "artifact_update"
}

# Access control: Agents can read but not write artifacts
allow {
    input.user.role == "agent"
    input.action.type == "artifact_read"
}

# Multi-tenant isolation: Users only see their org's artifacts
allow {
    input.action.type == "artifact_list"
    input.artifact.scope.org == input.user.org
}

# MCP access: Third-party MCPs require allowlist
allow {
    input.action.type == "mcp_connect"
    input.mcp.server_id in input.policy.mcp_allowlist
}
```

**Daemon API**:
```
POST /api/v1/policy/authorize
  Body: { "user": {...}, "action": {...}, "resource": {...} }
  Returns: { "decision": "PERMIT|DENY|OBLIGE", "obligations": [...] }

POST /api/v1/policy/evaluate
  Body: { "rego_policy": "...", "input": {...} }
  Returns: Policy evaluation result

GET /api/v1/policy/audit
  Returns: Recent authorization decisions (for audit trail)
```

---

## 9. Future Integration Framework (NEW)

**Vision**: Brain is designed for integrations that don't exist yet. New IDEs, CLIs, protocols, or AI surfaces can connect by implementing a standard interface.

### Integration Interface

Any new surface can connect to Brain by implementing:

```go
// core/integration/interface.go

type BrainClient interface {
    // Required: Authenticate with daemon
    Authenticate(ctx context.Context, creds Credentials) error
    
    // Required: Fetch context bundle
    GetContextBundle(ctx context.Context, req ContextRequest) (*ContextBundle, error)
    
    // Required: Resolve policy
    ResolvePolicy(ctx context.Context, scope ScopeChain) (*PolicyResult, error)
    
    // Required: Resolve artifacts
    ResolveArtifact(ctx context.Context, kind, id string) (*Artifact, error)
    
    // Required: Listen for events
    SubscribeEvents(ctx context.Context, handler EventHandler) error
    
    // Optional: Execute agent jobs
    ExecuteJob(ctx context.Context, job JobRequest) (*JobResult, error)
    
    // Optional: Call MCP tools
    CallMCP(ctx context.Context, server, tool string, args map[string]any) (any, error)
}

// Brain provides SDK implementations for each surface:
// - brain-sdk-go (for Go-based CLIs)
// - brain-sdk-ts (for TypeScript IDE extensions)
// - brain-sdk-rs (for Rust-based tools)
// - brain-sdk-py (for Python integrations)
```

### Protocol Support Matrix

| Protocol | Status | Used By | Brain Implementation |
|----------|--------|---------|---------------------|
| HTTP REST | ✅ Current | All surfaces | `api/v1/*` endpoints |
| WebSocket | ✅ Current | Real-time surfaces | Event streaming |
| MCP (stdio) | ✅ Current | Local IDEs/CLIs | MCP server |
| MCP (HTTP) | ✅ Current | Remote IDEs/CLIs | MCP server |
| ACP | 🔧 Planned | Zed, JetBrains | ACP server |
| LSP | 🔧 Planned | Neovim, editors | LSP server |
| gRPC | 🔧 Future | High-performance clients | gRPC service |
| GraphQL | 🔧 Future | Web clients | GraphQL API |
| SSE | 🔧 Future | Browser clients | Server-Sent Events |

---

## Consolidated Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)
- [x] Architecture documentation (complete)
- [ ] GAP-002: Observability Stack (OpenTelemetry)
- [ ] Artifact Registry with dependencies and version resolution
- [ ] Token Efficiency Engine (cache + compaction basics)
- [ ] Context Compiler (basic bundle assembly)

### Phase 2: Intelligence (Weeks 5-8)
- [ ] GAP-004: Model Capability Tier Router
- [ ] GAP-003: Context Curator (dedup, compaction, autoDream)
- [ ] GAP-005: Memory Sync (cloud sync + conflict resolution)
- [ ] MCP Server Hub (first 5 official servers)
- [ ] Governance Engine (RBAC + ABAC + OPA/Rego)

### Phase 3: Orchestration (Weeks 9-12)
- [ ] GAP-001: Agent Delegation Graph
- [ ] Agent Pool Manager (first 5 agents)
- [ ] DAG-Based Workflow Orchestration
- [ ] Pre-built Workflows (feature-dev, bug-fix, code-review)
- [ ] Skill Registry + Marketplace (discovery + install + security scan)

### Phase 4: Self-Improvement (Weeks 13-16)
- [ ] AutoEvolve Engine (monitor + analyze + propose)
- [ ] Review & Apply UI (web + CLI)
- [ ] Auto-generated skill proposals
- [ ] Token waste analysis
- [ ] GAP-007: Cost Optimization Engine (dashboards + budgets)

### Phase 5: Client Surfaces (Weeks 17-22)
- [ ] MCP Server Hub (community + enterprise servers)
- [ ] VS Code Extension
- [ ] Brain Desktop (Tauri)
- [ ] ACP Server (Zed + JetBrains)
- [ ] Qwen Code, Codex CLI, OpenCode, Continue.dev integrations
- [ ] GAP-006: Desktop App
- [ ] GAP-009: TUI

### Phase 6: Advanced (Weeks 23+)
- [ ] GAP-008: Agent Teams/Pooling (auto-scaling)
- [ ] Future Integration Framework (SDK for new surfaces)
- [ ] Advanced AutoEvolve (self-modifying configs, ML-based proposals)
- [ ] Enterprise features (SSO, SCIM, compliance reports)
- [ ] Multi-region deployment
- [ ] Collaborative features (shared context, pair programming)

---

## Success Metrics

Brain is successful when:

1. **Every IDE/CLI has the same capabilities**: No feature gap between VS Code, Claude Code, Cursor, Qwen Code, etc.
2. **Zero duplication**: Skills, MCPs, rules installed once, available everywhere
3. **Token efficiency**: >50% cost reduction through caching, compaction, routing
4. **Self-improving**: AutoEvolve proposes and applies improvements weekly
5. **Policy enforcement**: 100% policy compliance across all surfaces
6. **Agent orchestration**: Multi-agent workflows complete 3x faster than single-agent
7. **Zero missing features**: AutoEvolve detects gaps and fills them automatically
8. **Cross-surface memory**: Context learned in VS Code is available in Claude Code
9. **Cost visibility**: Every user knows exactly what they're spending and where
10. **Extensibility**: New IDEs/CLIs can integrate in <1 day using SDK

---

## Related Documents

- Target Architecture: `docs/architecture/brain-v2-target-architecture.md`
- IDE/CLI Integration: `docs/architecture/ide-cli-integration-strategy.md`
- Gap Roadmap: `docs/architecture/implementation-gap-roadmap.md`
- Capability Control Plane: `docs/architecture/capability-control-plane-roadmap.md`
- AI Runtime: `docs/architecture/ai-runtime-and-context-optimization.md`

---

**Status**: Active
**Last updated**: 2026-04-11
**Author**: Brain Architecture Team
