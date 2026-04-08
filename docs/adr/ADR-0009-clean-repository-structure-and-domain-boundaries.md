---
id: ADR-0009
title: Clean Repository Structure and Domain Boundaries
status: approved
date_created: 2026-04-07
language: en
type: architecture-decision
category: adr
version: 1.0.0
---

## ADR-0009: Clean Repository Structure and Domain Boundaries

Status: APPROVED  
Deciders: Brain Architecture Team  
Date: 2026-04-07

## Context

The repository root currently exposes many domain folders directly. That shape makes the system feel like a collection of concepts rather than a product with clear boundaries.

Brain is now explicitly defined as:

- multiple client surfaces
- one daemon-centered control plane
- one managed artifact system
- one governance and security layer

That system model is better served by a root organized around product boundaries than by a flat list of artifact types.

## Options Considered

### Option 1: Keep artifact domains at repository root

Pros:

- Minimal immediate change
- Easy to browse individual domains in a small prototype

Cons:

- Root grows noisier as new domains are added
- Harder to separate executable product, artifacts, and internal tooling
- Poor fit for long-term onboarding and governance

Trade-offs: Simpler today, more confusing tomorrow.

### Option 2: Consolidate repository into `apps/`, `core/`, `artifacts/`, and supporting roots (Chosen)

Pros:

- Clean product-oriented root
- Clear split between applications, canonical artifacts, and internal support
- Better fit for multi-surface product evolution

Cons:

- Requires path migration work
- Existing docs and scripts need updates

Trade-offs: Higher migration effort in exchange for a clearer and more scalable repository shape.

## Decision

Brain will adopt a clean root organized around applications, core subsystems, artifacts, internal support, and deployment concerns.

## Rationale

This structure aligns with the actual architecture:

- `apps/` for executable surfaces
- `core/` for shared system subsystems
- `artifacts/` for managed assets
- `internal/` for non-public support code

The new structure reflects how Brain behaves, not just how it started.

## Consequences

### Positive Outcomes

- The root remains readable as the system grows
- New domains can be added without inventing new root conventions
- Runtime-safe and development-only concerns become easier to separate

### Risks and Mitigation

- Risk: Existing commands and docs break during migration
  - Probability: High
  - Impact: High
  - Mitigation: Use compatibility shims and move incrementally

- Risk: Migration distracts from feature delivery
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Finish architecture closure first, then refactor by subsystem

## Implementation Notes

Implementation details are documented in:

- `docs/architecture/repository-structure-v2.md`

## Related ADRs

- ADR-0002: Centralized Brain CLI for Multi-IDE Service Orchestration (relationship: complements)
- ADR-0003: Centralization of Orchestration in Daemon (relationship: complements)
- ADR-0008: Unified Artifact Packaging and Lifecycle (relationship: depends_on)

Status: Active  
Decided by: Brain Architecture Team  
Decision date: 2026-04-07  
Review date: 2026-06-30
