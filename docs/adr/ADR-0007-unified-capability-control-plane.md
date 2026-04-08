---
id: ADR-0007
title: Unified Capability Control Plane with Hierarchical Context Resolution
status: approved
date_created: 2026-04-04
language: en
type: architecture-decision
category: adr
version: 1.0.0
---

## ADR-0007: Unified Capability Control Plane with Hierarchical Context Resolution

Status: APPROVED  
Deciders: Brain Architecture Team  
Date: 2026-04-04

## Context

Brain already provides strong foundations:

- Daemon is the orchestration source of truth.
- Skills support CRUD and sync propagation to configured targets.
- CLI and UI are active consumers of daemon state.

However, the repository still lacks an official, end-to-end model for final-state capability governance across:

- skills
- agents
- MCP definitions
- rules
- provider routing policies

Current behavior is functional but not yet formalized as a single hierarchical resolution model with:

- explicit precedence across global, organization, team, user, project, and task scopes
- policy-aware conflict resolution
- token-budget-aware context assembly
- deterministic explainability for why a capability is active

Without this formal model, long-term scaling risks include policy drift, duplicated logic, inconsistent behavior across surfaces, and unclear security boundaries.

## Options Considered

### Option 1: Keep current domain-by-domain evolution

Pros:

- Lowest short-term implementation cost
- No major contract changes immediately required

Cons:

- Resolution logic remains fragmented
- Harder to enforce consistent policy and security
- Higher long-term maintenance cost

Trade-offs: Preserves velocity now, increases architectural entropy later.

### Option 2: Move source of truth to generated per-surface artifacts

Pros:

- Fast local adaptation per IDE/surface
- Simple surface-specific optimization

Cons:

- Violates single-daemon orchestration principle
- Creates multiple competing sources of truth
- Increases drift and rollback complexity

Trade-offs: Local convenience at the cost of governance correctness and reliability.

### Option 3: Adopt a unified capability control plane in daemon (Chosen)

Pros:

- Preserves single-daemon principle
- Enables one policy and precedence model for all domains
- Supports secure, explainable, and deterministic resolution
- Scales to multi-tenant and multi-surface use

Cons:

- Requires new contracts and migration work
- Requires phased rollout and compatibility shims

Trade-offs: Higher initial coordination, lower long-term complexity and risk.

## Decision

Brain will adopt a unified capability control plane model in the daemon.

The model will standardize capability resolution across skills, agents, MCPs, rules, and provider policies using:

1. Unified capability catalog (canonical metadata model)
2. Hierarchical precedence and policy enforcement
3. Context compiler with token-budget-aware assembly
4. Deterministic projection to CLI, UI, and IDE targets

The rollout will be skills-first, then generalized to other domains.

## Rationale

This decision aligns with existing repository constraints and proven ecosystem patterns:

- It reinforces daemon-centric orchestration instead of fragmenting logic.
- It supports explicit security controls and least-privilege policy layering.
- It makes behavior explainable and testable across all surfaces.
- It keeps portability high by separating canonical source from generated projections.

The chosen model is the simplest approach that preserves future options while reducing long-term operational risk.

## Consequences

### Positive Outcomes

- One consistent capability resolution model across all domains
- Better security posture through policy and scope boundaries
- Clear explainability for active capabilities and conflict outcomes
- Cleaner scaling path for organizations, teams, and projects
- Reduced duplication across daemon, CLI, UI, and adapters

### Risks and Mitigation

- Risk: Migration complexity during transition to unified contracts
  - Probability: Medium
  - Impact: High
  - Mitigation: Skills-first rollout, compatibility layer, phased enforcement gates

- Risk: Performance regressions from added resolution steps
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Benchmark gates, caching strategy, token budget limits, profiling before promotion

- Risk: Policy overreach blocking legitimate workflows
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Audit logs, dry-run policy mode, explicit override workflow with traceability

## Implementation Notes

Implementation contract and migration path are documented in:

- `docs/architecture/capability-control-plane-roadmap.md`

Initial rollout phases:

- Phase 1: Skills contract hardening and hierarchy metadata
- Phase 2: Daemon context resolution endpoint for skills
- Phase 3: Policy and explainability layer
- Phase 4: Generalization to agents, MCPs, rules, and providers

## Related ADRs

- ADR-0002: Centralized Brain CLI for Multi-IDE Service Orchestration (relationship: enables)
- ADR-0003: Centralization of Orchestration in Daemon (relationship: depends_on)
- ADR-0005: Strict Development and Production Boundary (relationship: constrains)
- ADR-0008: Unified Artifact Packaging and Lifecycle (relationship: extends)
- ADR-0010: Hierarchical Identity, Policy, and Security Model (relationship: constrained_by)

Status: Active  
Decided by: Brain Architecture Team  
Decision date: 2026-04-04  
Review date: 2026-06-30
