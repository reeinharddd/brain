---
type: documentation-schema
id: unified-documentation-schema
title: Brain Repository - Unified Documentation Schema
version: 2.0.0
status: active
date_created: 2026-04-03
language: en
category: documentation-standards
maintained_by: documentation-team

keywords: [schema, templates, standards, documentation, rag, frontmatter]
rag_priority: critical
chunk_strategy: section

related:
  - type: guidance
    id: documentation-standards
    relationship: references

description: >
  Unified, industry-aligned documentation schema for internal and functional
  documentation across the Brain repository. All documents MUST use this schema.
  Covers 10 document types: ADRs, design docs, project status, agents, skills,
  tools, MCPs, hooks, state documentation, and examples.
---

<!-- markdownlint-disable-file -->

# Brain Repository: Unified Documentation Schema

**Version**: 2.0.0  
**Status**: Active - All new documents must comply  
**Objective**: Single source of truth for documentation format, metadata, and structure

---

## Table of Contents

1. [Core Principles](#core-principles)
2. [Universal Frontmatter Schema](#universal-frontmatter-schema)
3. [Document Types & Structures](#document-types--structures)
4. [Metadata Guidelines](#metadata-guidelines)
5. [Formatting Standards](#formatting-standards)
6. [RAG Optimization](#rag-optimization)
7. [Examples & Test Cases](#examples--test-cases)
8. [Anti-Patterns](#anti-patterns)
9. [Validation Checklist](#validation-checklist)

---

## Core Principles

### 1. Unified Metadata First

All documents start with **structured YAML frontmatter**. Metadata must not appear in prose (bold text, embedded headers).

```yaml
---
type: { document-type }
id: { unique-identifier }
title: { Human-readable title }
version: { semver }
status: active|draft|review|deprecated|archived
date_created: { YYYY-MM-DD }
language: en
category: { domain }
---
```

### 2. Machine-Readable & Human-Readable

- YAML frontmatter for machines (indexing, automation)
- Clear prose/structure for humans
- Examples before abstract descriptions
- No ambiguity in instructions

### 3. RAG-Optimized Chunking

Documents must be splittable into LLM context windows:

- Clear section boundaries (H2 headers)
- Self-contained sections (can be read independently)
- Explicit relationships between sections
- Keywords for semantic search

### 4. No Duplication of Content

Every fact appears in ONE place. If it appears in multiple docs, use YAML `related:` to link.

### 5. English Only

All documentation in English. No Spanish, no mixed languages.

### 6. Status Fields Enable Evolution

Track document lifecycle: draft -> review -> active -> deprecated -> archived

---

## Universal Frontmatter Schema

**ALL documents start with this tiered YAML frontmatter**:

### Tier 1: Core (Required for Every Type)

```yaml
---
type:
  {
    agent|skill|adr|design-doc|project-status|tool|mcp-server|hook|state|example,
  }
id: { unique-identifier }
title: { Human-readable title, 5-10 words }
version: { semver }
status: active|draft|review|deprecated|archived
date_created: { YYYY-MM-DD }
language: en
---
```

### Tier 2: Classification (Recommended)

```yaml
category:
  { documentation|architecture|execution|data|integration|operations|security }
tags: [tag1, tag2, tag3] # 2-5 tags for discovery
maintainer: { name or team }
visibility: internal|team|public
```

### Tier 3: Relationships (Highly Recommended)

```yaml
related:
  - type: adr
    id: ADR-015
    relationship: depends_on|enables|supersedes|based_on|implements
  - type: design-doc
    id: DESIGN-auth
    relationship: references
```

### Tier 4: Versioning & Evolution

```yaml
version: 1.0.0 # semver: major.minor.patch
previous_version: 0.9.0
next_review_date: 2026-07-03
deprecated_date: null
deprecation_notice: null
```

### Tier 5: RAG Optimization

```yaml
keywords: [auth, oauth2, security] # 3-5 terms for semantic search
rag_priority: critical|high|medium|low # Ranking in search results
chunk_strategy: section|sequential|example-first # How to split for LLM
chunk_boundaries: # Explicit section boundaries
  - "## Context"
  - "## Options Considered"
  - "## Decision"
estimated_reading_time: "8 minutes"
estimated_implementation_time: "4 hours" # If design/task doc
```

### Tier 6: Metadata (Context-Specific)

For specific types (filled only when relevant):

```yaml
# For ADRs
decision_maker: alice@company.com
affected_systems: [api-gateway, cache-layer, logging]

# For Design Docs
target_release: Q2-2026
impact_level: major|minor|patch
requires_migration: true|false

# For Skills/Tools
input_schema: ./schema/input.json
compatibility: ["claude-3-opus", "claude-3-sonnet"]
timeout_seconds: 30

# For Projects
member_count: 4
timeline_start: 2026-03-15
timeline_end: 2026-05-30
```

---

## Document Types & Structures

### 1. Architecture Decision Record (ADR)

**File**: `docs/adr/ADR-{NNN}-{slug}.md`  
**Purpose**: Record significant architectural choices and their rationale

````yaml
---
type: adr
id: ADR-015
title: OAuth 2.0 for Authentication
version: 1.0.0
status: active
date_created: 2026-04-01
decision_maker: security-team
category: architecture
tags: [security, auth, infrastructure]
affected_systems: [api-gateway, identity-service, client-sdk]

related:
  - type: adr
    id: ADR-008
    relationship: supersedes
  - type: design-doc
    id: DESIGN-auth-v2
    relationship: implements

keywords: [oauth2, authentication, authorization, openid-connect]
rag_priority: high
chunk_strategy: section
---

# ADR-015: OAuth 2.0 for Authentication

## Context

[Describe the business, technical, and regulatory drivers for this decision.]
[What problem are we solving? What constraints do we have?]
[Why now? Why this matters.]

## Options Considered

### Option 1: OAuth 2.0 with PKCE
**Pros**:
- Industry standard (IETF RFC 6749, RFC 7636)
- Supports multiple flows (authorization code, PKCE, client credentials)
- Mature ecosystem & libraries

**Cons**:
- More complex for simple internal APIs
- Requires additional infrastructure (token server)

**Trade-offs**:
- Complexity cost: Medium
- Security benefit: High
- Operational overhead: Medium

### Option 2: JWT with Shared Secret
**Pros**:
- Lightweight, stateless
- Simple to implement
**Cons**:
- No standard revocation mechanism
- Vulnerable to key compromise
- Not suitable for public APIs

**Trade-offs**:
- Security benefit: Low
- Operational overhead: High (key rotation)

### Option 3: Session-Based (Cookies)
**Pros**:
- Traditional, well-understood by teams

**Cons**:
- State management required
- Doesn't scale to multiple services
- Poor for mobile clients


**Chosen**: OAuth 2.0 with PKCE flows for all user-facing APIs.

**Rationale**:
- Balances security with modern architecture
- Supports web, mobile, and third-party integrations
- Standard protocol reduces lock-in risk
- Team expertise in OAuth ecosystem

## Consequences

### Positive Outcomes
- Enables third-party developer ecosystem
- Compliance with industry security standards
- Easier user onboarding across clients
- Auditability and consent tracking built-in

### Negative & Risks

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Increased infrastructure complexity | Medium | Use hosted solutions (Auth0, Okta) |
| Token expiry & refresh logic in clients | Medium | Provide SDK handling token refresh |
| Potential timing attacks on token validation | Medium | Use constant-time comparison, HTTP-only cookies |

## Implementation Notes

- **PR**: #2845 (Implemented auth server)
- **Deployment**: Staged rollout (shadow -> 10% -> GA)
- **Migration**: Existing session tokens migrated via temporary dual-auth

## Monitoring

- Track token issuance/redemption metrics
- Monitor token expiry failures
- Alert on suspicious login patterns

## Related ADRs

- ADR-008: Superseded by this decision
- ADR-012: Token refresh strategy (separate concern)
- ADR-020: Internal service-to-service auth (uses different flow)

---

### 2. Design Document

**File**: `docs/architecture/DESIGN-{slug}.md`
**Purpose**: Detailed technical proposal for a feature, system, or subsystem

```yaml
---
type: design-doc
id: DESIGN-cache-v2
title: Cache Layer Redesign for Sub-100ms Latency
version: 2.1.0
status: review
date_created: 2026-03-20
category: architecture
tags: [performance, caching, redis]
author: engineering-team
reviewers: [alice@company.com, bob@company.com]

target_release: Q2-2026
impact_level: major
requires_migration: true

related:
  - type: adr
    id: ADR-045-redis-choice
    relationship: implements
  - type: design-doc
    id: DESIGN-db-sharding
    relationship: depends_on

keywords: [cache, performance, redis, latency, architecture]
rag_priority: high
chunk_strategy: section
estimated_reading_time: "15 minutes"
estimated_implementation_time: "120 hours"
---

# Cache Layer Redesign

## Overview

Redesign cache layer to achieve <100ms p99 latency for read-heavy workloads while maintaining data consistency during cluster partitions.

## Motivation

Current cache hits 150-200ms p99. Analysis shows:
- Redis roundtrip: 5-10ms (network + serialization)
- Connection pool contention: 50-100ms (during spikes)
- Fallback penalty: 100-300ms (when cache misses)

Target: <100ms p99 across 99th percentile load.

## Goals

- **Primary**: P99 latency < 100ms (baseline: 180ms)
- **Secondary**: 95%+ cache hit rate
- Secondary: Zero data loss during failover
- Reduce operational complexity (vs. multi-region replicas)

## Non-Goals

- Support offline-first clients (separate project)
- Warm caches across data centers (future optimization)
- Custom consistency semantics (use eventual consistency)

## Proposed Solution

### Architecture

[Diagram showing layers: clients -> CDN -> Redis cluster -> DB]

```mermaid
graph TB
    Client["Clients"]
    CDN["Edge Cache<br/>(CloudFront)"]
    Redis["Redis Cluster<br/>4 nodes + sentinel"]
    DB["PostgreSQL<br/>Primary"]

    Client -->|HTTP| CDN
    CDN -->|Cache miss| Redis
    Redis -->|Cache hit| CDN
    Redis -->|Cache miss| DB
    DB -->|Write invalidation| Redis

    style Redis fill:#ffcccc
````

### Data Model

```yaml
cache_entries:
  key: string # User-defined key
  value: blob (max 1MB) # Serialized value (JSON)
  ttl: integer (seconds) # Auto-expire time
  version: integer # For invalidation
  created_at: timestamp
  last_accessed: timestamp
  fingerprint: hash # For consistency checking
```

### API & Interfaces

#### Get

```
GET /cache/{key}

Response 200:
{
  "value": <serialized>,
  "ttl_remaining": seconds,
  "version": 123
}

Response 404:
{
  "error": "KEY_NOT_FOUND"
}
```

#### Set

```
POST /cache/{key}

Body:
{
  "value": <serialized>,
  "ttl": seconds
}

Response 201:
{ "version": 124 }
```

#### Delete

```
DELETE /cache/{key}

Response 204: (no content)
```

### Deployment Strategy

| Phase     | Timeline | Actions                                      |
| --------- | -------- | -------------------------------------------- |
| 1. Setup  | Week 1   | Provision Redis cluster (4 nodes) in staging |
| 2. Shadow | Week 2-3 | Run parallel (redirect 5% read traffic)      |
| 3. Canary | Week 4   | Increase to 25% of production reads          |
| 4. GA     | Week 5   | Full production traffic; deprecate old cache |

### Trade-offs

| Aspect      | Chosen            | Alternative    | Why                         |
| ----------- | ----------------- | -------------- | --------------------------- |
| Consistency | Eventual (TTL)    | Strong (locks) | Latency (locks add 50ms)    |
| Storage     | Redis in-memory   | RocksDB (disk) | Latency (disk slow)         |
| Replication | Sentinel failover | Multi-region   | Complexity vs latency gains |

## Alternative Approaches

### Approach 1: Local In-Memory Cache + Distributed Tier

- Faster local hits (1-2ms)
- Problem: Invalidation nightmare, stale data across servers

### Approach 2: GraphQL DataLoader Pattern

- Batches multiple requests
- Problem: Doesn't help single-request latency

### Approach 3: HTTP/2 Server Push

- Preemptively send cached data
- Problem: Doesn't work for unpredictable access patterns

**Why chosen solution**:

- Simplest to operate
- Lowest latency variance
- Clear failure modes

## Risks & Mitigation

| Risk                                 | Severity | Mitigation                               |
| ------------------------------------ | -------- | ---------------------------------------- |
| Redis cluster partition -> stale data | High    | Use Redis Sentinel + TTL-based staleness |
| Hot keys cause uneven load           | Medium   | Monitor per-key access; shard hot keys   |
| Large values (>1MB) block others     | Medium   | Reject >1MB; suggest compression         |
| Thundering herd on cache miss        | Medium   | Implement probabilistic early expiry     |

## Success Metrics

- **Latency**: P99 < 100ms (baseline: 180ms)
- **Availability**: 99.95% uptime (Redis cluster SLA)
- **Hit rate**: 95%+ (baseline: 88%)
- **Operational MTTR**: <5 min recovery from node failure

## Monitoring & Alerts

```yaml
metrics:
  - name: cache_hit_ratio
    alert: < 90%
    dashboard: true

  - name: cache_latency_p99
    alert: > 150ms
    dashboard: true

  - name: redis_cluster_health
    alert: < 4 nodes healthy
    dashboard: true

dashboards:
  - grafana-cache-health.json
```

## Dependencies & Concerns

- [ ] Redis cluster provisioning (ops-team)
- [ ] Client library updates to handle cache invalidation
- [ ] Data serialization standard (JSON or MessagePack?)
- [ ] Fallback strategy when Redis is down

## Open Questions

- [ ] Do we need separate caches for different data sensitivity levels?
- [ ] Should we implement cache warming on startup?
- [ ] Is compression enabled by default or opt-in?

## Sign-off

- **Proposed by**: alice@company.com
- **Reviewed by**: bob@company.com (architect)
- **Approved by**: charlie@company.com (tech lead)
- **Date**: 2026-04-03

---

### 3. Project Status

**File**: `docs/status/PROJECT-{code}-{phase}.md`  
**Purpose**: Lightweight tracking of active projects, phases, and blockers

````yaml
---
type: project-status
id: project-cache-redesign
title: Cache Layer Redesign - Phase 2
status: in-progress
version: 1.0.0
date_created: 2026-03-20
date_updated: 2026-04-03

team: infrastructure
member_count: 4
phase: implementation
timeline_start: 2026-03-15
timeline_end: 2026-05-30

related:
  - type: design-doc
    id: DESIGN-cache-v2
    relationship: implements
  - type: adr
    id: ADR-045
    relationship: based_on

keywords: [cache, redis, infrastructure, performance]
rag_priority: medium
chunk_strategy: section
---

# Project: Cache Layer Redesign - Phase 2

## Current Status

**Phase**: Implementation (Weeks 7-10 of 10)
**Progress**: 65% complete
**Health**: On track

## This Week's Work

- [x] Redis cluster deployed to staging
- [x] Fallback cache strategy designed & approved
- [ ] Load testing suite configured (in progress, 60% done)
- [ ] Integration tests written (pending load test results)

## Active Blockers

| Blocker | Owner | ETA | Impact |
|---------|-------|-----|--------|
| Redis migration script for prod data | alice | Fri 4/6 | HIGH - blocks shadow testing |
| Test infra upgrade (versions) | bob | Wed 4/5 | MEDIUM - delays early detection |

## Upcoming Milestones

- [ ] **April 8**: Load tests complete
- [ ] **April 12**: Integration tests pass
- [ ] **April 15**: Shadow deployment (non-prod mirror)
- [ ] **April 22**: GA rollout (canary 10% -> GA)

## Risks (Next 2 Weeks)

- Alice on vacation April 9-12 (mitigation: Bob shadows on migration script)
- Load test environment instability (mitigation: pre-stage data migration)

## Key Decisions Made

- Chosen Redis (ADR-045) over Cassandra
- Using Sentinel for failover (simple operational model)
- Eventual consistency + TTL (not strong consistency with locks)

## Production Notes

- Cluster endpoint: `cache-prod.internal:6379`
- Failover IP: `10.0.1.100` (managed by DNS)
- Monitor: hit ratio, latency, cluster health

## Handover Notes

For next phase owner, see: `docs/handovers/HANDOVER-cache-design.md`

## Related Docs

- Design: `docs/architecture/DESIGN-cache-v2.md`
- ADR: `docs/adr/ADR-045-redis-choice.md`
- Runbook: `ops/cache-cluster-runbook.md`

---

### 4. Agent Definition

**File**: `artifacts/agents/{agent-id}.md`
**Purpose**: Define agent role, protocol, constraints for coordinated LLM behavior

```yaml
---
type: agent
id: implementer
title: Implementer Agent
version: 1.0.0
status: active
date_created: 2026-03-01
category: execution

tags: [coding, implementation, single-task]
maintainer: engineering-team
related:
  - type: agent
    id: planner
    relationship: works_with
  - type: agent
    id: reviewer
    relationship: works_with

keywords: [implementation, code, execution, task]
rag_priority: high
chunk_strategy: section
---

# Implementer Agent

## Role

You implement one bounded, well-specified task at a time. You write code, docs, or configurations following the approved design and spec.

**Core Contract**:
- One task per invocation (unless trivially inseparable)
- Smallest effective change to solve the task
- No scope expansion; escalate ambiguity to delegator
- Tests or validation alongside changes (not afterthoughts)

## When You Are Invoked

````

delegator_agent
├─ provides: task (with done condition)
├─ provides: spec artifact (what we're building)
├─ provides: design artifact (how to build it)
└─ triggers: implementer.execute(task)

```

You receive a task like:
> "Implement POST /api/users endpoint according to DESIGN-auth-v2.md, spec in users-spec.md. Tests must pass. Status code 201 on success, 400 on validation error."

## Mandatory Rules (Don't Negotiate These)

### Rule 1: Implement ONE Task
- If task has 3 sub-steps, ask if they're inseparable
- If separable, ask to split into 3 tasks
- If truly one unit, implement all 3 together

### Rule 2: Preserve Existing State
- Don't delete unrelated files
- Don't refactor code outside task scope
- Don't change configurations not required by this task

### Rule 3: Stop If Contract Changes Required
- If implementation needs to change the spec or design, STOP
- Escalate to delegator with explanation
- Do NOT silently modify the contract

### Rule 4: Success Criteria First
- Read the task's "done condition" before starting
- Code only to meet those criteria
- Validate against criteria before returning

### Rule 5: No Ambiguity Guessing
- If task is ambiguous, ask for clarification
- Don't use "standard practice" as a tiebreaker
- Document non-obvious decisions as comments

## Decision Protocol

When facing implementation choices not covered by spec/design:

1. Check the spec & design documents FIRST
2. If silent, apply rules in this order:
  - `artifacts/rules/modules/security.md` (security always first)
  - `artifacts/rules/modules/code-style.md` (consistency)
  - `artifacts/rules/modules/performance.md` (if design specifies performance goals)
3. If still ambiguous, ask the delegator
4. If you must guess, document it as a comment

## Protocol: State Machine

```

START
└─ receives: task with done condition
└─ -> ANALYZING

ANALYZING (read design, spec, task)
├─ ambiguity found? -> ASKING
├─ contract issues? -> ESCALATING
└─ ready to code? -> IMPLEMENTING

IMPLEMENTING (write code/docs)
├─ blocked or stuck? -> ASKING
├─ done? -> VALIDATING
└─ test failure? -> FIXING

VALIDATING (run tests,lint, check done condition)
├─ pass? -> RETURNING
└─ fail? -> IMPLEMENTING (loop)

ASKING (awaiting clarification)
└─ receive answer -> back to IMPLEMENTING

ESCALATING (issue with contract)
└─ receive updated task -> back to ANALYZING

RETURNING (task complete)
└─ return code/artifact + validation evidence

````

## Output Format

When done, return:

```markdown
# Implementation Complete: {Task Name}

## Changes Made
- File 1: [what changed]
- File 2: [what changed]

## Validation Evidence
- [ ] All tests pass
- [ ] No TypeScript errors
- [ ] Linting clean
- [ ] Done condition verified: {specific evidence}

## Non-Obvious Decisions
- Used approach X because [explained in design as requirement #3]

## Known Limitations
- [Anything deferred to future task]
````

## Examples

### Example 1: Simple Endpoint Implementation

**Task**:

> "Implement GET /api/users/:id endpoint. Returns user object on 200, 404 if not found. Must use current auth middleware. See DESIGN-api-v2.md and spec in api-spec.md."

**You do**:

1. Read spec.md for request/response format
2. Read DESIGN-api-v2.md for auth requirements
3. Implement handler + endpoint
4. Write tests for 200 and 404 cases
5. Validate against done condition
6. Return with evidence

**You do NOT**:

- Refactor the auth middleware (out of scope)
- Add caching (not in spec)
- Optimize database query (defer to perf task)

### Example 2: Escalating to Delegator

**Task**:

> "Add user role validation to auth middleware per ADR-024"

**You analyze and find**:

- ADR-024 doesn't specify HOW to validate (role enum? permission matrix? external service?)
- Spec is silent on validation method
- Design references ADR-024 but doesn't detail validation approach

**You do**:

- STOP implementation
- Message delegator: "ADR-024 is incomplete. Need clarification: validation method (enum vs matrix vs external service)?"
- WAIT for response
- Continue only when clarified

## Anti-Patterns You MUST NOT Do

 BAD: **Expand scope** — "While I'm here, I'll refactor the auth module"
GOOD: **Stay focused** — Implement ONE task

 BAD: **Guess on ambiguity** — "Usually we do X, so I'll do that"
GOOD: **Ask delegator** — "This is ambiguous, need clarification"

 BAD: **Silent contract changes** — Design says use PostgreSQL, you switch to MongoDB
GOOD: **Escalate blockers** — "Design doesn't support this requirement"

 BAD: **Test as afterthought** — Write code, then ask if we need tests
GOOD: **Test-driven** — Write tests alongside code

---

### 5. Skill Definition

**File**: `artifacts/skills/{skill-id}/SKILL.md`  
**Purpose**: Define a reusable capability for LLM agents with input/output contracts

````yaml
---
type: skill
id: codebase-contextualizer
name: Codebase Contextualizer Skill
version: 2.1.0
status: active
date_created: 2026-02-15

category: code-analysis
tags: [rag, code-understanding, context-extraction]
maintainer: research-team
compatibility: ["claude-3-opus", "claude-3-sonnet"]

inputs_schema: ./schemas/input.json
outputs_schema: ./schemas/output.json

execution:
  timeout_seconds: 30
  max_tokens_output: 8000
  retry_policy: "exponential_backoff_max_2"
  state_model: "stateless"

related:
  - type: example
    id: EXAMPLE-codebase-contextualizer
    relationship: references
  - type: agent
    id: researcher
    relationship: used_by

keywords: [codebase, context, rag, semantic, code-analysis]
rag_priority: high
chunk_strategy: section
estimated_cost: "$0.05 per invocation"
---

# Codebase Contextualizer Skill

## What It Does

Extracts semantic context from a codebase to create RAG-optimized summaries. Returns structured understanding of:
- Project structure and entry points
- Key modules and their responsibilities
- Dependencies (internal and external)
- Architectural patterns used
- File-to-concept mapping

## Input Contract

```json
{
  "codebase_path": "/absolute/path/to/repo",
  "depth": 3,
  "filters": ["ts", "js", "go", "py"],
  "include_tests": false,
  "max_files": 100
}
````

### Parameter Details

| Parameter       | Type    | Required | Default      | Notes                                   |
| --------------- | ------- | -------- | ------------ | --------------------------------------- |
| `codebase_path` | string  | YES      | -            | Absolute path, must be readable         |
| `depth`         | integer | NO       | 3            | Range: 1-5. Higher = more files, slower |
| `filters`       | array   | NO       | ["ts", "js"] | File extensions to include              |
| `include_tests` | boolean | NO       | false        | Include `*.test.*` and `*.spec.*` files |
| `max_files`     | integer | NO       | 100          | Stop after N files (prevent timeout)    |

## Output Contract

```json
{
  "status": "ok",
  "project": {
    "name": "my-app",
    "type": "react-typescript-app",
    "size_mb": 42,
    "file_count": 87
  },
  "structure": {
    "src": ["components", "api", "utils"],
    "entry_points": ["src/index.tsx"]
  },
  "modules": [
    {
      "path": "src/api/client.ts",
      "responsibility": "HTTP client for backend communication",
      "exports": ["apiClient", "request"]
    }
  ],
  "dependencies": {
    "external": ["react@18.2.0", "axios@1.4.0"],
    "internal": []
  },
  "patterns": ["singleton-http-client", "react-hooks", "barrel-exports"]
}
```

### Error Cases

| Code                | Message                        | Cause                             |
| ------------------- | ------------------------------ | --------------------------------- |
| `PATH_NOT_FOUND`    | Codebase path does not exist   | Invalid path provided             |
| `PERMISSION_DENIED` | Cannot read codebase directory | File permissions issue            |
| `DEPTH_TOO_HIGH`    | Depth > 5 not supported        | User input validation error       |
| `TIMEOUT`           | Analysis exceeded 30 seconds   | Codebase too large or slow system |
| `NO_FILES_FOUND`    | No files matching filters      | Filters too restrictive           |

**Handling**:
From these errors, return structured error response with code, message, and suggested actions.

## Examples

### Example 1: React Application

**Input**:

```json
{
  "codebase_path": "/home/user/react-app",
  "depth": 3,
  "filters": ["tsx", "ts"],
  "include_tests": false
}
```

**Output**:

```json
{
  "status": "ok",
  "project": {
    "name": "react-app",
    "type": "React + TypeScript",
    "size_mb": 25,
    "file_count": 45
  },
  "structure": {
    "src": ["components", "pages", "hooks", "utils"],
    "entry_points": ["src/index.tsx", "src/App.tsx"]
  },
  "modules": [
    {
      "path": "src/api/authClient.ts",
      "responsibility": "Authentication HTTP client",
      "exports": ["authClient", "useAuth", "withAuth"]
    },
    {
      "path": "src/components/Button.tsx",
      "responsibility": "Reusable button component",
      "exports": ["Button", "ButtonProps"]
    }
  ],
  "patterns": ["custom-hooks", "HOC-pattern", "barrel-exports"]
}
```

**Use Case**: New developer onboarding, codebase review before contributing

### Example 2: Go Microservice

**Input**:

```json
{
  "codebase_path": "/home/user/auth-service",
  "depth": 2,
  "filters": ["go"],
  "include_tests": true,
  "max_files": 50
}
```

**Output**:

```json
{
  "status": "ok",
  "project": {
    "name": "auth-service",
    "type": "Go HTTP Service",
    "size_mb": 8,
    "file_count": 12
  },
  "structure": {
    "cmd": ["main.go"],
    "internal": ["handlers", "models", "auth"]
  },
  "dependencies": {
    "external": [
      "github.com/gorilla/mux@v1.8.0",
      "github.com/dgrijalva/jwt-go@v3.2.0"
    ],
    "internal": []
  }
}
```

## Anti-Patterns & Mistakes

BAD: **Vague input**: `{path: "/repo"}`
-> User doesn't know if path must be absolute or relative, doesn't know about depth parameter

GOOD **Precise input schema**: Documented required fields, default values, validation rules

BAD: **Missing error cases**: "Might fail in some edge cases"
-> LLM doesn't know how to handle failures

GOOD **Enumerated errors**: Each code listed with cause, message, and handling procedure

BAD: **No timeouts**: Skill hangs if codebase is huge
-> LLM context blocks waiting for response

GOOD **Explicit timeouts**: 30 seconds max; graceful truncation if exceeded

BAD: **Ambiguous output**: "Returns information about the project"
-> LLM doesn't know what fields to expect

GOOD: **Structured output schema**: JSON with typed fields, consistent across invocations

## Compatibility

- **Tested on**: Claude 3 Opus, Claude 3 Sonnet, Claude 3 Haiku
- **Requires**: Minimum 4KB context tokens (for output)
- **Network**: Requires filesystem access
- **Performance**: Typical: 5-15 seconds, Max: 30 seconds

## Cost & Performance

- **Input tokens**: ~200 (schema + instructions)
- **Output tokens**: ~500-2000 (varies by codebase size)
- **Estimated cost**: $0.03-0.08 per invocation
- **Latency**: 5-15 seconds typical, 30 seconds max

---

### 6. Tool/Function Definition

**File**: `docs/tools/{tool-name}.md`  
**Purpose**: Document a single LLM-callable tool with input/output schemas

````yaml
---
type: tool
id: validate-email
name: Email Validator Tool
version: 1.2.0
status: active
date_created: 2026-03-10

category: validation
tags: [email, validation, utility]
maintainer: platform-team
compatibility: ["openai-gpt4", "anthropic-claude", "gemini-pro"]

related:
  - type: tool
    id: send-email
    relationship: complements

keywords: [email, validation, regex, dns]
rag_priority: medium
chunk_strategy: section
---

# Email Validator Tool

## Purpose

Validates email addresses with optional DNS MX record checking.

## Invocation Schema

```json
{
  "name": "validate_email",
  "description": "Validates an email address format and optionally checks DNS MX records",
  "input_schema": {
    "type": "object",
    "properties": {
      "email": {
        "type": "string",
        "description": "Email address to validate, e.g., user@example.com"
      },
      "check_dns": {
        "type": "boolean",
        "description": "Whether to perform DNS MX record lookup (slower, more thorough)",
        "default": false
      },
      "require_smtp": {
        "type": "boolean",
        "description": "Attempt SMTP verification (slow, may trigger rate limits)",
        "default": false
      }
    },
    "required": ["email"],
    "additionalProperties": false
  }
}
````

## Response Format

### Success Response (200)

```json
{
  "valid": true,
  "email": "user@example.com",
  "local_part": "user",
  "domain": "example.com",
  "format_valid": true,
  "dns_record_found": true,
  "smtp_verified": false,
  "disposable": false,
  "notes": []
}
```

### Error Response (400)

```json
{
  "valid": false,
  "error": {
    "code": "INVALID_FORMAT",
    "message": "Email address format is invalid",
    "details": {
      "reason": "Missing @ symbol"
    }
  }
}
```

## Error Codes

| Code                | Message              | Cause                       | Retry? |
| ------------------- | -------------------- | --------------------------- | ------ |
| `INVALID_FORMAT`    | Format is invalid    | Malformed email             | No     |
| `INVALID_DOMAIN`    | Domain is invalid    | Not a valid domain          | No     |
| `DNS_LOOKUP_FAILED` | DNS lookup failed    | Network issue or invalid MX | Yes    |
| `TIMEOUT`           | Validation timed out | DNS lookup too slow         | Yes    |

## Constraints

- **Max email length**: 254 characters (RFC 5321)
- **Timeout**: 5 seconds (DNS), 30 seconds (SMTP)
- **Rate limit**: 100 calls per minute per API key
- **State**: Stateless; can be called in parallel

---

## Metadata Guidelines

### Category Field

Use from this controlled list:

```
- documentation     (guides, tutorials, references)
- architecture      (design, decisions, patterns)
- execution         (agents, skills, tasks)
- data              (schemas, models, state)
- integration       (APIs, tools, MCPs, hooks)
- operations        (runbooks, monitoring, deployment)
- security          (policies, auth, compliance)
- performance       (benchmarks, optimization recommendations)
```

### Tags Field

2-5 short, searchable terms. Examples:

```yaml
tags: [redis, caching, performance]
tags: [oauth2, security, auth]
tags: [microservices, architecture, api]
```

### Status Field

Track document lifecycle:

```yaml
status: active         # In use, maintained
status: draft          # Under development
status: review         # Awaiting approval
status: deprecated     # No longer used; keep for history
status: archived       # Completed project; unlikely to change
```

### Related Field

Structure relationships between documents:

```yaml
related:
  - type: adr
    id: ADR-015
    relationship: depends_on # This doc depends on ADR-015

  - type: design-doc
    id: DESIGN-auth
    relationship: implements # This doc implements that design

  - type: agent
    id: implementer
    relationship: used_by # That agent uses this skill

  - type: tool
    id: validate-email
    relationship: complements # Works alongside this tool

  - type: adr
    id: ADR-008
    relationship: supersedes # Replaces old decision
```

---

## Formatting Standards

### Headings

- **H1 (#)**: Document title only (appears once)
- **H2 (##)**: Major sections (Context, Decision, Implementation, etc.)
- **H3 (###)**: Subsections (Pros, Cons, Error Cases, etc.)
- Never use H4+ unless necessary; usually indicates poor structure

### Lists

**Unordered** (when order doesn't matter):

```markdown
- Item 1
- Item 2
  - Nested item
  - Another nested
- Item 3
```

**Ordered** (when sequence matters):

```markdown
1. Step 1
2. Step 2
3. Step 3
   a. Substep A
   b. Substep B
```

### Tables

Use when comparing 3+ items across 2+ attributes:

```markdown
| Header 1 | Header 2 | Header 3 |
| -------- | -------- | -------- |
| Value A  | Value B  | Value C  |
```

### Code Blocks

Always specify language:

````markdown
```json
{
  "key": "value"
}
```

```python
def hello():
    print("world")
```
````

### Emphasis

- **Bold** for UI elements, file names, terms being defined
- _Italic_ for emphasis on important concepts
- `Code` for variable names, commands, file paths
- Never use BOTH bold + code: `**`code`**` (use one or the other)

---

## RAG Optimization

### Chunking Strategy

Documents are split at H2 boundaries for LLM context. Each chunk must be **self-contained**:

```yaml
chunk_strategy: section # Default: split at H2


# Example chunking:
# Chunk 1: [H1 title] + [## Context H2 + content]
# Chunk 2: [## Options H2 + content]
# Chunk 3: [## Decision H2 + content]
```

### Keywords

3-5 semantically relevant terms for semantic search:

```yaml
keywords: [oauth2, authentication, openid-connect]
```

These enable vector search to find the right document.

### Priority

Influences search ranking:

```yaml
rag_priority: critical  # Must always be included if relevant
rag_priority: high      # Usually included
rag_priority: medium    # Include if related to query
rag_priority: low       # Only if nothing else found
```

---

## Examples & Test Cases

### Pattern: Good Example

```markdown
## Examples

### Example 1: Simple User Lookup

**Scenario**: New user onboarding, needs to find user by email

**Input**:
```

GET /api/users?email=alice@company.com

````

**Output**:
```json
{
  "id": "user-123",
  "email": "alice@company.com",
  "name": "Alice"
}
````

**Explanation**: Query parameter filters the user list by email. Returns the user object if found.

````

### Pattern: Anti-Pattern Example

```markdown
## What NOT to Do

### Don't: Unvalidated User Input
```python
# WRONG - accepts any input without validation
@app.route("/api/users/<id>")
def get_user(id):
    return db.query(f"SELECT * FROM users WHERE id={id}")
````

**Why wrong**: SQL injection vulnerability. User ID is inserted directly into query.

### Do Instead

```python
# CORRECT - input is parameterized
@app.route("/api/users/<id>")
def get_user(id):
    return db.query("SELECT * FROM users WHERE id=?", (id,))
```

**Why correct**: Uses parameterized queries to prevent injection.

````

---

## Anti-Patterns

### Anti-Pattern 1: Metadata in Prose

BAD: **Wrong**:
```markdown
# Design Doc: Cache Layer

**Status**: In Review
**Author**: alice@company.com
**Date**: April 3, 2026
````

GOOD: **Correct**:

```yaml
---
type: design-doc
status: review
author: alice@company.com
date_created: 2026-04-03
---
# Design Doc: Cache Layer
```

### Anti-Pattern 2: Vague Input/Output Specs

BAD: **Wrong**:

```markdown
## Input

Takes a user object

## Output

Returns updated user
```

GOOD: **Correct**:

```yaml
inputs:
  user:
    type: object
    properties:
      id:
        type: string
        example: "user-123"
      name:
        type: string
        example: "Alice Smith"
```

### Anti-Pattern 3: Long, Non-Chunked Documents

BAD: **Wrong**: Single document with 200+ lines, no section breaks

GOOD: **Correct**: H2 sections at regular intervals (~300 words each); self-contained chunks

### Anti-Pattern 4: Duplicated Content

BAD: **Wrong**: Skill definition doc + separate skill usage guide

GOOD: **Correct**: One skill SKILL.md with examples embedded

### Anti-Pattern 5: No Examples

BAD: **Wrong**:

```markdown
## How to Use

Call the skill with appropriate parameters.
```

GOOD: **Correct**:

```markdown
## Examples

### Example 1: Simple Case

**Input**: {...}
**Output**: {...}
**Use Case**: When you need to...
```

---

## Validation Checklist

Use this before finalizing any document:

### Metadata

- [ ] `type` field present and correct
- [ ] `id` field present (unique, slugified)
- [ ] `title` field 5-10 words
- [ ] `version` follows semver
- [ ] `status` is one of: active, draft, review, deprecated, archived
- [ ] `date_created` in YYYY-MM-DD format
- [ ] `language` is "en"
- [ ] All text is English (no Spanish, no mixed)

### Content

- [ ] H1 title matches `title` field
- [ ] H2 sections are self-contained (can be read independently)
- [ ] No metadata in prose (all in frontmatter YAML)
- [ ] Examples before abstract concepts
- [ ] Anti-patterns documented with explanations
- [ ] No code without language specified
- [ ] Tables used appropriately (3+ rows, 2+ columns)

### Relationships

- [ ] `related` field populated if doc references others
- [ ] Cross-document links use Markdown links (`[text](path)`)
- [ ] No internal duplication of content
- [ ] Circular dependencies checked and avoided

### RAG Optimization

- [ ] `keywords` present (3-5 terms)
- [ ] `rag_priority` set
- [ ] `chunk_strategy` specified
- [ ] Sections are 300-500 words (RAG-optimal)

### Completeness

- [ ] Done condition clear (for tasks/projects)
- [ ] All required fields per document type included
- [ ] Examples provided (2-3 minimum)
- [ ] Error cases documented
- [ ] Related documents linked

---

## Quick Reference: Document Type Selection Matrix

| Need                              | Type        | Location                           |
| --------------------------------- | ----------- | ---------------------------------- |
| Record a major decision           | ADR         | `docs/adr/ADR-NNN-slug.md`         |
| Propose a feature/design          | Design Doc  | `docs/architecture/DESIGN-slug.md` |
| Define architecture direction     | Design Doc  | `docs/architecture/[name].md`      |
| Explain skill placement and rules | Skill Doc   | `docs/skills/[name].md`            |
| Define quality baseline           | Testing Doc | `docs/testing/[name].md`           |
| Define reusable authoring pattern | Template    | `docs/templates/.../TEMPLATE.md`   |

---

## Version History

| Version | Date       | Changes                                   |
| ------- | ---------- | ----------------------------------------- |
| 2.0.0   | 2026-04-03 | Initial unified schema; 10 document types |
| 1.0.0   | 2026-03-15 | Legacy multi-system docs (deprecated)     |

---

**Status**: Active - All documents created after 2026-04-03 MUST follow this schema.
