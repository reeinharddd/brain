---
type: design-doc
id: implementation-gap-roadmap
title: Implementation Gap Roadmap
version: 1.0.0
status: active
date_created: 2026-04-11
language: en
category: architecture
relationships:
  - brain-v2-target-architecture
  - ide-cli-integration-strategy
  - capability-control-plane-roadmap
---

## Overview

This document consolidates all identified gaps and improvement opportunities from market research against the current Brain architecture. Each gap includes source (industry pattern), priority, implementation plan, and acceptance criteria.

Gaps are organized by priority: 🔴 Critical (must implement), 🟡 Important (should implement), 🟢 Future (nice to have).

## Gap Inventory

### 🔴 GAP-001: Agent Delegation Graph

**Source**: Claude Code (Fork/Teammate/Worktree subagents), Codex CLI (6 concurrent threads), CrewAI (role-based delegation), AutoGen (conversational multi-agent)

**Problem**: Brain's architecture mentions agents and subagents but does not define delegation protocol, DAG structure, budget controls, or fallback chains.

**Impact**: Without this, Brain cannot orchestrate complex multi-agent workflows. Users are limited to single-agent interactions.

**Implementation**:

Create `core/delegation/` module:

```
core/delegation/
├── graph.go           # DAG of agent delegation
├── budget.go          # Token/time/cost budgets per delegation
├── fallback.go        # Fallback chain on agent failure
├── executor.go        # Delegation execution engine
├── telemetry.go       # Trace per delegation
└── delegation_test.go
```

**Data Structures**:

```go
// core/delegation/graph.go

type DelegationMode string

const (
    DelegationFork      DelegationMode = "fork"       // Isolated context (Claude Code pattern)
    DelegationTeammate  DelegationMode = "teammate"   // Shared context (Claude Code pattern)
    DelegationWorktree  DelegationMode = "worktree"   // Parallel git worktrees (Claude Code pattern)
    DelegationConcurrent DelegationMode = "concurrent" // Parallel threads (Codex CLI pattern)
)

type DelegationGraph struct {
    ID          string
    RootAgent   string
    Nodes       map[string]*AgentNode
    Edges       map[string][]string  // parent → children
    Mode        DelegationMode
    MaxDepth    int
    MaxParallel int
    Budget      DelegationBudget
    Fallback    FallbackChain
}

type AgentNode struct {
    ID          string
    AgentID     string
    Role        string         // "architect", "builder", "reviewer", etc.
    Input       TaskInput
    Constraints []Constraint
    Timeout     time.Duration
    RetryPolicy RetryPolicy
    Metadata    map[string]string
}

type DelegationBudget struct {
    MaxTokens     int
    MaxCostUSD    float64
    MaxDuration   time.Duration
    MaxRetries    int
    PerAgentLimit *AgentBudgetLimit // Optional per-agent sub-budget
}

type FallbackChain struct {
    Steps []FallbackStep
}

type FallbackStep struct {
    Condition   FailureCondition  // timeout, error, token_limit, policy_violation
    Action      FallbackAction    // retry, escalate, simplify, abort
    TargetAgent string            // Agent to delegate to
    Parameters  map[string]string
}
```

**Daemon API Endpoints**:

```
POST /api/v1/delegation/execute
  Body: { "graph_id": "...", "root_agent": "...", "task": {...}, "budget": {...} }
  Returns: { "execution_id": "...", "status": "running" }

GET /api/v1/delegation/{execution_id}/status
  Returns: { "status": "running|completed|failed", "progress": {...}, "results": [...] }

GET /api/v1/delegation/{execution_id}/trace
  Returns: OpenTelemetry-compatible trace of delegation execution

POST /api/v1/delegation/{execution_id}/cancel
  Cancels running delegation
```

**Acceptance Criteria**:
- [ ] Delegation graph can be defined with 3+ agents
- [ ] Fork mode creates isolated context per agent
- [ ] Teammate mode shares context between agents
- [ ] Worktree mode creates parallel git worktrees
- [ ] Budget enforcement: token, cost, duration limits
- [ ] Fallback chain activates on failure (retry → escalate → simplify → abort)
- [ ] Telemetry trace available for debugging
- [ ] Integration tests with mock agents
- [ ] Brain CLI can submit delegation jobs

---

### 🔴 GAP-002: Observability Stack (OpenTelemetry)

**Source**: LangSmith (tracing, evaluation, monitoring), industry standard: 80% of production readiness = observability

**Problem**: `core/observability/` is defined as domain but has no implementation. No structured logging, event traces, or health status.

**Impact**: Cannot debug daemon issues, cannot monitor production, cannot provide visibility to users.

**Implementation**:

Create `core/observability/` module with OpenTelemetry compatibility:

```
core/observability/
├── tracer.go          # OpenTelemetry tracer
├── metrics.go         # Prometheus-compatible metrics
├── logger.go          # Structured logging (slog/zap)
├── exporter.go        # OTLP exporter configuration
├── health.go          # Health check endpoint
├── trace_context.go   # Trace context propagation
└── observability_test.go
```

**Implementation Details**:

```go
// core/observability/tracer.go

package observability

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/sdk/trace"
)

type TracerConfig struct {
    ServiceName    string
    OTLPEndpoint   string
    SampleRate     float64  // 0.0-1.0, 1.0 = sample all
    Enabled        bool
}

func InitTracer(ctx context.Context, cfg TracerConfig) (*trace.TracerProvider, error) {
    if !cfg.Enabled {
        return nil, nil
    }
    
    exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(
        otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
    ))
    if err != nil {
        return nil, err
    }
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(cfg.SampleRate))),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}

// Daemon-spanning attributes
const (
    AttrArtifactKind   = "brain.artifact.kind"
    AttrArtifactID     = "brain.artifact.id"
    AttrScope          = "brain.scope"
    AttrPolicyClass    = "brain.policy.class"
    AttrModelID        = "brain.model.id"
    AttrCapabilityTier = "brain.model.capability_tier"
)
```

```go
// core/observability/metrics.go

package observability

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Artifact metrics
    ArtifactResolutionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "brain_artifact_resolution_duration_seconds",
            Help: "Time to resolve artifacts hierarchically",
        },
        []string{"kind", "scope"},
    )
    
    ArtifactLoadErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "brain_artifact_load_errors_total",
            Help: "Total artifact load failures",
        },
        []string{"kind", "reason"},
    )
    
    // Context metrics
    ContextBundleSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "brain_context_bundle_size_tokens",
            Help: "Size of compiled context bundles",
        },
        []string{"scope_chain"},
    )
    
    ContextCompressionRatio = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "brain_context_compression_ratio",
            Help: "Original/compressed token ratio",
        },
        []string{"bundle_id"},
    )
    
    // Policy metrics
    PolicyResolutionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "brain_policy_resolution_duration_seconds",
            Help: "Time to resolve policy hierarchy",
        },
        []string{"scope_depth"},
    )
    
    PolicyOverrides = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "brain_policy_overrides_total",
            Help: "Total policy overrides applied",
        },
        []string{"class", "scope"},
    )
    
    // Model routing metrics
    ModelRoutingDecisions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "brain_model_routing_decisions_total",
            Help: "Total model routing decisions",
        },
        []string{"model", "capability_tier", "reason"},
    )
    
    ModelCostUSD = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "brain_model_cost_usd_total",
            Help: "Total USD spent on model API calls",
        },
        []string{"model", "workspace"},
    )
    
    // Daemon health metrics
    DaemonUptime = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "brain_daemon_uptime_seconds",
            Help: "Daemon uptime in seconds",
        },
    )
    
    ActiveSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "brain_active_sessions",
            Help: "Number of active client sessions",
        },
    )
)
```

**Daemon API Endpoints**:

```
GET /api/v1/health
  Returns: { "status": "healthy", "uptime": 12345, "version": "0.1.0" }

GET /metrics
  Returns: Prometheus-compatible metrics

GET /api/v1/traces/{execution_id}
  Returns: Trace timeline for delegation/job execution

POST /api/v1/health/check
  Returns: Detailed health check (artifact db, policy engine, memory, sync)
```

**Acceptance Criteria**:
- [ ] Structured logging with levels (info, warn, error) + context (artifact ID, scope, trace ID)
- [ ] OpenTelemetry tracing across all daemon operations
- [ ] Prometheus metrics exported and scrapeable
- [ ] Health check endpoint with detailed component status
- [ ] Trace context propagation across HTTP requests
- [ ] OTLP exporter configuration (supports Jaeger, Tempo, Lightstep)
- [ ] Integration tests verifying trace/metric emission

---

### 🔴 GAP-003: Context Curator Implementation

**Source**: Industry context engineering patterns (Write/Select/Compress/Isolate), Hindsight memory (retain/recall/reflect), KAIROS autoDream

**Problem**: Curator is designed (proposes deduplication, promotion, cleanup) but not implemented. No actual context optimization runs.

**Impact**: Context bundles grow unbounded. No deduplication, no compression, no memory hygiene.

**Implementation**:

Create `core/context/curator/` module:

```
core/context/
├── curator/
│   ├── curator.go          # Main curator service
│   ├── deduplication.go    # Duplicate detection (hash + semantic)
│   ├── compaction.go       # Context compaction (summarization)
│   ├── promotion.go        # Context promotion (short-term → long-term)
│   ├── cleanup.go          # Cleanup recommendations (dry-run)
│   ├── auto_dream.go       # Idle-time consolidation (KAIROS-inspired)
│   └── curator_test.go
├── bundle.go               # Context bundle compiler
└── context_test.go
```

**Implementation Details**:

```go
// core/context/curator/curator.go

package curator

type CuratorConfig struct {
    Enabled              bool
    IdleThreshold        time.Duration  // Trigger consolidation after X idle
    MinSessionsForDream  int            // Minimum sessions before autoDream (KAIROS: 5)
    MinHoursSinceDream   int            // Minimum hours between autoDream (KAIROS: 24)
    MaxMemoryLines       int            // Cap on MEMORY.md (KAIROS: 200)
    MaxMemoryBytes       int            // Cap on memory size (KAIROS: 25KB)
    ConsolidationLock    bool           // Prevent concurrent consolidation
    DryRun               bool           // Never silently rewrite
}

type CuratorService struct {
    config     CuratorConfig
    detector   *DeduplicationDetector
    compactor  *Compactor
    promoter   *Promoter
    cleaner    *CleanupAdvisor
    dreamer    *AutoDreamService
}

// CuratorReport is returned from all curator operations
type CuratorReport struct {
    RunAt           time.Time
    DryRun          bool
    Duplicates      []DuplicateFinding
    Compactions     []CompactionSuggestion
    Promotions      []PromotionSuggestion
    Cleanups        []CleanupSuggestion
    MemoryState     MemoryState
    TokenSavings    TokenSavingsEstimate
}

// Run executes full curator analysis (dry-run by default)
func (c *CuratorService) Run(ctx context.Context) (*CuratorReport, error) {
    report := &CuratorReport{
        RunAt:  time.Now(),
        DryRun: c.config.DryRun,
    }
    
    // Phase 1: Detect duplicates
    dups, err := c.detector.Detect(ctx)
    if err != nil {
        return nil, err
    }
    report.Duplicates = dups
    
    // Phase 2: Suggest compactions
    compactions, err := c.compactor.Analyze(ctx)
    if err != nil {
        return nil, err
    }
    report.Compactions = compactions
    
    // Phase 3: Suggest promotions
    promotions, err := c.promoter.Analyze(ctx)
    if err != nil {
        return nil, err
    }
    report.Promotions = promotions
    
    // Phase 4: Cleanup recommendations
    cleanups, err := c.cleaner.Analyze(ctx)
    if err != nil {
        return nil, err
    }
    report.Cleanups = cleanups
    
    // Phase 5: Memory state assessment
    memState, err := c.assessMemoryState(ctx)
    if err != nil {
        return nil, err
    }
    report.MemoryState = memState
    
    // Phase 6: Estimate token savings
    report.TokenSavings = c.estimateSavings(report)
    
    return report, nil
}

// Apply executes curator actions (NOT dry-run, requires explicit approval)
func (c *CuratorService) Apply(ctx context.Context, actions []CuratorAction) (*CuratorReport, error) {
    if c.config.DryRun {
        return nil, fmt.Errorf("curator is in dry-run mode, cannot apply actions")
    }
    
    // Log all actions for auditability (KAIROS pattern: append-only)
    if err := c.logActions(actions); err != nil {
        return nil, err
    }
    
    // Execute each action
    for _, action := range actions {
        if err := c.executeAction(ctx, action); err != nil {
            return nil, fmt.Errorf("action %s failed: %w", action.Type, err)
        }
    }
    
    return c.Run(ctx) // Re-run to get updated report
}
```

```go
// core/context/curator/auto_dream.go

package curator

// AutoDreamService implements KAIROS-inspired idle-time consolidation
type AutoDreamService struct {
    config         CuratorConfig
    lastDreamTime  time.Time
    sessionCounter int
    activityMonitor *ActivityMonitor
}

// Tick is called periodically by daemon. Evaluates whether to trigger autoDream.
func (ads *AutoDreamService) Tick(ctx context.Context) error {
    // Check conditions (KAIROS pattern: 3 conditions must align)
    if time.Since(ads.lastDreamTime) < time.Duration(ads.config.MinHoursSinceDream)*time.Hour {
        return nil // Too soon
    }
    if ads.sessionCounter < ads.config.MinSessionsForDream {
        return nil // Not enough sessions
    }
    if ads.config.ConsolidationLock {
        return nil // Another consolidation in progress
    }
    if !ads.activityMonitor.IsIdle(ads.config.IdleThreshold) {
        return nil // User still active
    }
    
    // Trigger autoDream
    return ads.executeDream(ctx)
}

// executeDream runs the four-phase consolidation (KAIROS pattern)
func (ads *AutoDreamService) executeDream(ctx context.Context) error {
    // Phase 1: Orient - assess current memory state
    orient, err := ads.orient(ctx)
    if err != nil {
        return err
    }
    
    // Phase 2: Gather Recent Signal - extract insights from recent sessions
    signals, err := ads.gatherRecentSignal(ctx, orient.LastConsolidation)
    if err != nil {
        return err
    }
    
    // Phase 3: Consolidate - merge new knowledge with existing memory
    consolidated, err := ads.consolidate(ctx, orient.Memory, signals)
    if err != nil {
        return err
    }
    
    // Phase 4: Prune & Index - remove redundancies, cap memory size
    pruned, err := ads.pruneAndIndex(ctx, consolidated)
    if err != nil {
        return err
    }
    
    // Update state
    ads.lastDreamTime = time.Now()
    ads.sessionCounter = 0
    
    // Append-only log entry (KAIROS pattern: daemon cannot erase its own logs)
    if err := ads.logDreamRun(pruned); err != nil {
        return err
    }
    
    return nil
}
```

**Daemon API Endpoints**:

```
POST /api/v1/context/curator/run
  Query: ?dry_run=true
  Returns: CuratorReport with duplicates, compactions, promotions, cleanups

POST /api/v1/context/curator/apply
  Body: { "actions": [{ "type": "deduplicate", "target": "skill:abc123" }] }
  Returns: Updated CuratorReport after applying actions

GET /api/v1/context/curator/report
  Returns: Last curator report

GET /api/v1/context/curator/audit
  Returns: Append-only log of all curator actions (including autoDream)
```

**Acceptance Criteria**:
- [ ] Duplicate detection: hash-based (exact) + semantic similarity (fuzzy)
- [ ] Compaction analysis: identifies oversized context layers, suggests summaries
- [ ] Promotion analysis: identifies short-term context that should become long-term
- [ ] Cleanup recommendations: identifies stale/deprecated artifacts for archival
- [ ] AutoDream: triggers during idle, 4-phase cycle (Orient→Gather→Consolidate→Prune)
- [ ] Memory caps: MEMORY.md capped at 200 lines / 25KB
- [ ] Dry-run default: curator never silently rewrites
- [ ] Append-only audit log: all curator actions logged, daemon cannot modify logs
- [ ] Token savings estimation in every report

---

### 🔴 GAP-004: Model Capability Tier Router

**Source**: LiteLLM (model routing), industry trend: cost-aware routing to smallest capable model

**Problem**: Capability tiers defined in docs (Tier 1=constrained, Tier 2=standard, Tier 3=advanced) but no routing implementation.

**Impact**: All requests use default model. No cost optimization, no capability-aware routing, no fallback.

**Implementation**:

Create `core/runtime/router/` module:

```
core/runtime/
├── router/
│   ├── router.go           # Model router service
│   ├── capability_tier.go  # Tier detection and routing logic
│   ├── cost_estimator.go   # Token/cost estimation
│   ├── budget.go           # Per-user/workspace budget enforcement
│   └── router_test.go
├── model.go                # Model definitions
└── job.go                  # Job execution
```

**Implementation Details**:

```go
// core/runtime/router/capability_tier.go

package router

type CapabilityTier int

const (
    Tier1Constrained CapabilityTier = 1  // Small context, limited tools, no complex reasoning
    Tier2Standard   CapabilityTier = 2  // Standard context, tool use, moderate reasoning
    Tier3Advanced   CapabilityTier = 3  // Full context, advanced reasoning, complex orchestration
)

type ModelCapability struct {
    ModelID           string
    Provider          string
    Tier              CapabilityTier
    MaxContextTokens  int
    MaxOutputTokens   int
    SupportsTools     bool
    SupportsParallel  bool
    CostPer1KInput    float64
    CostPer1KOutput   float64
    LatencyP50        time.Duration
    LatencyP99        time.Duration
}

// RouteRequest defines requirements for model routing
type RouteRequest struct {
    MinCapabilityTier CapabilityTier
    PreferredModels   []string  // Ordered preference list
    FallbackModels    []string  // Fallback if preferred unavailable
    MaxTokens         int
    RequiresTools     bool
    RequiresParallel  bool
    BudgetUSD         float64  // Maximum willing to spend
    PreferLocal       bool     // Prefer local models if capable
    LatencySLA        time.Duration  // Maximum acceptable latency
}

// RouteResponse is the router's decision
type RouteResponse struct {
    SelectedModel  *ModelCapability
    Reason         string
    CostEstimate   float64
    FallbackChain  []string  // Models to try if selected fails
    CompressionNeeded bool   // Whether context needs compression before sending
}

// Route selects the best model based on requirements and policy
func (r *ModelRouter) Route(ctx context.Context, req RouteRequest) (*RouteResponse, error) {
    candidates := r.filterCandidates(req)
    
    if len(candidates) == 0 {
        return nil, fmt.Errorf("no models meet requirements: tier>=%d, tools=%v", 
            req.MinCapabilityTier, req.RequiresTools)
    }
    
    // Score candidates
    scored := r.scoreCandidates(candidates, req)
    
    // Select best
    best := scored[0]
    
    return &RouteResponse{
        SelectedModel: best.Model,
        Reason: best.Reason,
        CostEstimate: r.estimateCost(best.Model, req.MaxTokens),
        FallbackChain: r.buildFallbackChain(scored, best),
        CompressionNeeded: req.MaxTokens > best.Model.MaxContextTokens,
    }, nil
}

// scoreCandidates ranks models based on requirements
func (r *ModelRouter) scoreCandidates(models []*ModelCapability, req RouteRequest) []ScoredModel {
    var scored []ScoredModel
    for _, m := range models {
        score := 0.0
        
        // Capability match (higher tier = higher score for complex tasks)
        score += float64(m.Tier) * 10
        
        // Cost preference (lower cost = higher score if budget-conscious)
        avgCost := (m.CostPer1KInput + m.CostPer1KOutput) / 2
        score += (1.0 / avgCost) * 5
        
        // Local preference
        if req.PreferLocal && isLocalModel(m.ModelID) {
            score += 20
        }
        
        // Latency preference
        if m.LatencyP50 < req.LatencySLA {
            score += 10
        }
        
        // Policy override (user/org preference)
        if r.policyPrefersModel(m.ModelID) {
            score += 15
        }
        
        scored = append(scored, ScoredModel{Model: m, Score: score, 
            Reason: fmt.Sprintf("tier=%d, cost=%.4f/1K, latency=%v", m.Tier, avgCost, m.LatencyP50)})
    }
    
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })
    return scored
}
```

```go
// core/runtime/router/cost_estimator.go

package router

type CostEstimator struct {
    models map[string]*ModelCapability
    budgets map[string]Budget  // user/workspace → budget
}

type Budget struct {
    DailyUSD    float64
    MonthlyUSD  float64
    CurrentSpend float64
    ResetAt     time.Time
}

// EstimateCost calculates estimated cost for a request
func (ce *CostEstimator) EstimateCost(modelID string, inputTokens, outputTokens int) float64 {
    model := ce.models[modelID]
    if model == nil {
        return 0
    }
    
    inputCost := (float64(inputTokens) / 1000.0) * model.CostPer1KInput
    outputCost := (float64(outputTokens) / 1000.0) * model.CostPer1KOutput
    return inputCost + outputCost
}

// CheckBudget verifies request is within budget
func (ce *CostEstimator) CheckBudget(userID, workspaceID string, estimatedCost float64) error {
    // Check user budget
    if budget, ok := ce.budgets[userID]; ok {
        remaining := budget.DailyUSD - budget.CurrentSpend
        if estimatedCost > remaining {
            return fmt.Errorf("user budget exceeded: %.2f > %.2f remaining", 
                estimatedCost, remaining)
        }
    }
    
    // Check workspace budget
    if budget, ok := ce.budgets[workspaceID]; ok {
        remaining := budget.MonthlyUSD - budget.CurrentSpend
        if estimatedCost > remaining {
            return fmt.Errorf("workspace budget exceeded: %.2f > %.2f remaining",
                estimatedCost, remaining)
        }
    }
    
    return nil
}
```

**Behavior by Tier**:

```go
// Router applies different strategies based on model tier:

func (r *ModelRouter) PrepareContext(ctx ContextBundle, model *ModelCapability) ContextBundle {
    switch model.Tier {
    case Tier1Constrained:
        // Aggressively compress: keep only hard policy + task-local
        return r.aggressiveCompress(ctx)
    case Tier2Standard:
        // Full context bundle, moderate compression if needed
        if ctx.Totals.TokenCount > model.MaxContextTokens {
            return r.moderateCompress(ctx)
        }
        return ctx
    case Tier3Advanced:
        // Full context, no compression. Enable orchestration features
        return ctx
    default:
        return ctx
    }
}
```

**Daemon API Endpoints**:

```
POST /api/v1/runtime/route
  Body: { "min_tier": 2, "max_tokens": 4000, "requires_tools": true, "budget_usd": 0.05 }
  Returns: { "model": "gpt-4", "reason": "...", "cost_estimate": 0.03, "fallback": ["claude-sonnet"] }

GET /api/v1/runtime/models
  Returns: List of registered models with capabilities

GET /api/v1/runtime/budget/{user_or_workspace_id}
  Returns: Current budget status and spending

POST /api/v1/runtime/budget/{user_or_workspace_id}
  Body: { "daily_usd": 1.0, "monthly_usd": 30.0 }
  Sets budget limits
```

**Acceptance Criteria**:
- [ ] Router selects model based on capability tier, cost, latency, policy
- [ ] Tier 1 models get aggressively compressed context
- [ ] Tier 2 models get full context with moderate compression if needed
- [ ] Tier 3 models get full context, orchestration enabled
- [ ] Cost estimation accurate within 10% of actual
- [ ] Budget enforcement: rejects requests over budget
- [ ] Fallback chain: if selected model fails, tries next automatically
- [ ] Local model preference when policy allows
- [ ] Prometheus metrics: model routing decisions, cost tracking

---

### 🟡 GAP-005: Memory Sync with Conflict Resolution

**Source**: Qdrant (semantic memory), Hindsight (persistent memory), KAIROS autoDream, industry cross-session continuity

**Problem**: Qdrant local defined, cloud sync not implemented. No conflict resolution, no audit trail, no encryption.

**Impact**: Memory is local-only. No cross-device continuity, no cloud backup, no collaborative memory.

**Implementation**:

Enhance `core/sync/` with memory sync support:

```
core/sync/
├── engine.go            # Sync engine interface
├── push_pull.go         # Push/pull with conflict resolution
├── conflict.go          # Conflict resolution strategies
├── audit.go             # Audit trail for all sync operations
├── encryption.go        # Encryption in-transit + at-rest
├── modes.go             # Local, cloud-synced, hybrid modes
└── sync_test.go
```

**Conflict Resolution Strategies**:

```go
// core/sync/conflict.go

package sync

type ConflictResolutionStrategy string

const (
    LastWriteWins     ConflictResolutionStrategy = "last_write_wins"
    Manual            ConflictResolutionStrategy = "manual"
    Merge             ConflictResolutionStrategy = "merge"
    LocalPreferred    ConflictResolutionStrategy = "local_preferred"
    RemotePreferred   ConflictResolutionStrategy = "remote_preferred"
)

type ConflictResolver struct {
    defaultStrategy ConflictResolutionStrategy
    typeStrategies  map[string]ConflictResolutionStrategy  // per artifact type
}

type Conflict struct {
    LocalArtifact  Artifact
    RemoteArtifact Artifact
    ConflictType   string  // "content", "metadata", "lifecycle_state"
    DetectedAt     time.Time
}

type Resolution struct {
    Conflict   Conflict
    Strategy   ConflictResolutionStrategy
    Winner     string  // "local", "remote", "merged"
    MergedArtifact *Artifact  // If strategy is "merge"
    ResolvedAt time.Time
    RequiresUserApproval bool
}

func (cr *ConflictResolver) Resolve(conflict Conflict) (*Resolution, error) {
    strategy := cr.getStrategy(conflict.LocalArtifact.Kind)
    
    switch strategy {
    case LastWriteWins:
        return cr.lastWriteWins(conflict)
    case Merge:
        return cr.merge(conflict)
    case LocalPreferred:
        return &Resolution{Conflict: conflict, Strategy: strategy, Winner: "local"}, nil
    case RemotePreferred:
        return &Resolution{Conflict: conflict, Strategy: strategy, Winner: "remote"}, nil
    case Manual:
        return &Resolution{
            Conflict: conflict, 
            Strategy: strategy, 
            RequiresUserApproval: true,
        }, nil
    default:
        return cr.lastWriteWins(conflict)
    }
}
```

**Acceptance Criteria**:
- [ ] Push/pull with cloud backend
- [ ] Conflict resolution: last-write-wins, merge, local-preferred, remote-preferred, manual
- [ ] Audit trail: all sync operations logged with timestamp, source, outcome
- [ ] Encryption: TLS in-transit, AES-256 at-rest
- [ ] Sync modes: local (no sync), cloud-synced (auto), hybrid (explicit approval)
- [ ] Offline handling: queues operations, syncs when reconnected

---

### 🟡 GAP-006: Desktop App (Tauri)

**Source**: Cursor (IDE-first wins), industry trend: visual artifact management

**Problem**: Desktop app planned but not scaffolded.

**Impact**: Non-CLI users have no Brain interface.

**Implementation**:

```
apps/desktop/
├── src/
│   ├── main.ts           # Tauri setup
│   ├── App.tsx           # Root component
│   ├── components/
│   │   ├── Artifacts.tsx  # Artifact browser
│   │   ├── Context.tsx    # Context bundle viewer
│   │   ├── Policy.tsx     # Policy resolver view
│   │   └── Events.tsx     # Real-time event log
│   ├── clients/
│   │   └── daemon.ts      # HTTP + WebSocket client
│   └── utils/
├── src-tauri/
│   ├── Cargo.toml
│   ├── tauri.conf.json
│   └── src/
│       └── main.rs        # Rust backend (daemon communication)
├── package.json
└── tsconfig.json
```

**Acceptance Criteria**:
- [ ] Tauri app scaffolds and builds
- [ ] Connects to daemon via HTTP + WebSocket
- [ ] Views: Artifacts (list/detail), Context (bundle visualization), Policy (resolved rules), Events (real-time)
- [ ] Auth: stores daemon auth tokens securely
- [ ] Offline: shows disconnected state, queues actions

---

### 🟡 GAP-007: Cost Optimization Engine

**Source**: Industry trend: smallest capable model routing, OpenCode dual-pipeline MoE

**Problem**: No cost tracking, no optimization recommendations.

**Impact**: Users cannot control spending. No visibility into model costs.

**Implementation**:

Create `core/cost/` module:

```
core/cost/
├── estimator.go           # Token/cost estimation per model
├── optimizer.go           # Cost optimization recommendations
├── budget.go              # Budget management
├── report.go              # Cost reports and analytics
└── cost_test.go
```

**Acceptance Criteria**:
- [ ] Cost estimation per model API call
- [ ] Budget management: daily, monthly, per-workspace
- [ ] Optimization recommendations: "this skill could use a cheaper model"
- [ ] Cost reports: spending by model, workspace, user, time period
- [ ] Alerts: budget threshold warnings

---

### 🟢 GAP-008: Agent Teams/Pooling

**Source**: AutoGen (dynamic teams), OpenClaw (specialized agents), CrewAI (role-based teams)

**Problem**: Agent pooling mentioned but not detailed. No auto-scaling by load.

**Impact**: Cannot dynamically create agent teams for complex tasks.

**Acceptance Criteria**:
- [ ] Agent Pool Manager with auto-scaling
- [ ] Dynamic team creation based on task complexity
- [ ] Team templates (architect + builder + reviewer)
- [ ] Load-based scaling: spawn more agents under heavy load

---

### 🟢 GAP-009: TUI (Terminal UI)

**Problem**: Terminal UI planned but not implemented.

**Acceptance Criteria**:
- [ ] Terminal UI with artifact browsing
- [ ] Context bundle visualization in terminal
- [ ] Policy resolution display
- [ ] Real-time event monitoring

---

## Implementation Priority Order

Based on market research and architectural dependencies:

```
Phase 1: Foundation (Weeks 1-3)
├── GAP-002: Observability Stack          ← Everything else needs visibility
├── GAP-004: Model Capability Tier Router ← Needed for runtime decisions
└── Artifact validation hardening          ← Already in progress

Phase 2: Context & Memory (Weeks 4-6)
├── GAP-003: Context Curator              ← Context hygiene is critical
├── GAP-005: Memory Sync                  ← Cross-session continuity
└── Context bundle compiler               ← Hierarchical resolution

Phase 3: Agent Orchestration (Weeks 7-9)
├── GAP-001: Agent Delegation Graph       ← Multi-agent workflows
├── Delegation execution engine           ← Run delegation graphs
└── GAP-007: Cost Optimization Engine     ← Cost visibility

Phase 4: Client Surfaces (Weeks 10-14)
├── GAP-006: Desktop App                  ← Visual interface
├── MCP Server                            ← Cursor, Windsurf, Continue
├── VS Code Extension                     ← Most popular IDE
└── ACP Server                            ← Zed, JetBrains

Phase 5: Additional Integrations (Weeks 15-18)
├── Qwen Code integration
├── Codex CLI integration
├── OpenCode plugin
├── Continue.dev bridge
└── GAP-009: TUI

Phase 6: Advanced Features (Weeks 19+)
├── GAP-008: Agent Teams/Pooling
├── Cloud sync advanced features
├── Collaborative memory
└── Enterprise features
```

## Validation Gates

All implementations must pass:

1. **Unit Tests**: >80% coverage on core logic
2. **Integration Tests**: End-to-end with mock clients
3. **Security Review**: No secrets, no privilege escalation, input validation
4. **Performance Benchmarks**: Meet latency SLAs (artifact resolution <100ms, context compilation <500ms)
5. **Documentation**: Updated architecture docs, ADR if decision required
6. **Linting**: `go vet`, `golangci-lint`, no warnings
7. **Schema Validation**: All data structures validated against JSON schemas
