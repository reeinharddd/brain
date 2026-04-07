---
type: design-doc
id: capability-control-plane-roadmap
title: Capability Control Plane - Final State and Delivery Roadmap
version: 1.0.0
status: active
date_created: 2026-04-04
language: en
category: architecture
---

## Overview

This document defines the target final state for Brain as a mature capability control plane and the staged plan to reach that state.

It translates strategic direction into concrete architecture, constraints, and implementation phases.

## Motivation

Brain has grown into a multi-surface system (daemon, CLI, desktop UI, IDE adapters) with strong partial capability management, especially for skills.

To reach a fully mature final state, the project needs one explicit and secure model for:

- capability governance
- hierarchical context resolution
- policy enforcement
- explainable projections to all consumer surfaces

Without this model, future growth will increase drift, duplicated logic, and policy inconsistency.

## Goals

- Define what "final-state Brain" should be at maximum functional maturity.
- Establish one canonical capability model across skills, agents, MCPs, rules, and provider routing.
- Enforce secure-by-default policy layering with deterministic precedence.
- Preserve daemon as single operational source of truth.
- Provide a phased roadmap with measurable success gates.

## Non-Goals

- Replacing daemon orchestration with distributed per-surface orchestration.
- Introducing surface-local sources of truth for capability state.
- Rewriting all existing domains at once without migration safeguards.
- Expanding scope into unrelated product areas.

## High-Level Design

### Final-State System Model

At final maturity, Brain operates as a unified control plane with four core subsystems:

1. Canonical Capability Catalog
   - One normalized data contract for capabilities
   - Domain-agnostic metadata: ownership, scope, policy level, activation conditions, compatibility

2. Hierarchical Resolution Engine
   - Deterministic precedence across scope layers
   - Conflict handling between hard constraints and soft preferences

3. Context Compiler
   - Selects active capabilities for current task/project
   - Applies token-budget-aware assembly and ordering
   - Produces explainable resolution traces

4. Projection and Sync Plane
   - Generates surface-specific outputs from canonical resolved state
   - Keeps CLI, UI, and IDE adapters as consumers only

### Precedence Contract

Resolution uses this order from lowest to highest precedence:

- platform global
- organization or team policy
- user global
- project scope
- task or invocation scope
- runtime hooks

Conflict rules:

- Hard policy always overrides soft preference.
- More specific scope can override less specific scope only if it does not violate hard policy from higher governance layers.
- Every override is logged with reason metadata.

### Security Model

The final-state security model includes:

- deny-by-default policy enforcement
- least-privilege capability exposure
- explicit boundary between production-safe and development-only behavior
- auditable decision trace for capability activation and rejection

## Expected Final State

When the project is complete at maximum maturity, Brain should provide:

- Unified capability governance across all relevant domains
- Stable multi-tenant hierarchy support
- Explainable resolution outcomes for every request
- Token-efficient context assembly at scale
- Reliable projections to every supported surface
- Secure default operation with explicit, auditable overrides
- Continuous observability for quality, cost, and latency

## Delivery Roadmap

### Phase 1 - Skills Contract Hardening

Scope:

- Extend skills metadata to support hierarchy and enforcement fields.
- Keep backward compatibility with existing registry entries.

Exit Criteria:

- Skills schema supports ownership, scope, enforcement, and priority metadata.
- Existing skill workflows continue to operate without regression.

### Phase 2 - Skills Resolution API

Scope:

- Add daemon endpoint for context resolution of skills.
- Return resolved set plus explainability payload.

Exit Criteria:

- Deterministic resolution in daemon for representative scenarios.
- CLI and UI can render resolved results and reasons.

### Phase 3 - Policy and Explainability Layer

Scope:

- Introduce policy evaluation with hard and soft enforcement classes.
- Add dry-run policy mode and conflict diagnostics.

Exit Criteria:

- Policy outcomes are auditable and reproducible.
- Security policy violations are blocked with explicit reasons.

### Phase 4 - Domain Generalization

Scope:

- Apply the same model to agents, MCPs, rules, and provider routing.
- Unify projection contracts for all surfaces.

Exit Criteria:

- One resolution model covers all core capability domains.
- Surface outputs are derived-only and internally consistent.

### Phase 5 - Optimization and Operations

Scope:

- Add adaptive context optimization and budget control.
- Add observability dashboards and reliability SLOs.

Exit Criteria:

- Stable latency and token usage targets in production profiles.
- Operational alerts and dashboards cover policy, resolution, and sync health.

## Validation Plan

The roadmap is considered successful when all gates pass:

- Correctness: deterministic resolution tests pass across hierarchy scenarios.
- Security: hard policy constraints are always enforced.
- Maintainability: no duplicated resolution logic across surfaces.
- Performance: p95 latency and token budget targets are met.
- Operability: traces and logs explain all activation decisions.

## Key Metrics

- Resolution correctness rate
- Policy conflict rate and auto-resolution rate
- Token reduction percentage versus baseline
- p95 and p99 resolution latency
- Sync propagation success rate
- Explainability coverage percentage

## Dependencies

- Existing daemon orchestration and sync engine
- Stable metadata evolution strategy for capability registries
- Policy definitions and enforcement contract
- Surface clients updated to consume resolved outputs

## Related Decisions and References

- `docs/adr/ADR-0002-centralized-brain-cli.md`
- `docs/adr/ADR-0003-central-daemon-orchestration.md`
- `docs/adr/ADR-0004-skill-system-contract.md`
- `docs/adr/ADR-0005-strict-development-and-production-boundary.md`
- `docs/adr/ADR-0006-docs-rag-mcp.md`
- `docs/adr/ADR-0007-unified-capability-control-plane.md`
- `docs/architecture/skill-system-contract.md`
- `docs/architecture/environment-configuration.md`
