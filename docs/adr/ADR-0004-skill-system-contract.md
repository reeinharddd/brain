---
type: adr
id: ADR-0004-portable-skill-contract-with-surface-adapters
title: Portable Skill Contract with Surface Adapters
version: 1.0.0
status: accepted
date_created: 2026-04-03
language: en
category: architecture
related: []
keywords:
  - skills
  - registry
  - adapters
  - portability
  - validation
rag_priority: high
chunk_strategy: section
---

## ADR-0004: Portable Skill Contract with Surface Adapters

## Context

The Brain repository already has a skill registry, reusable context packs, a daemon that loads skills, a CLI that consumes the daemon API, and a desktop UI that observes sync status. The remaining problem is how to separate portable skill content from surface-specific behavior.

The current mix is too broad:

- portable skill packages
- reusable stack context
- registry metadata
- surface-specific behavior
- older master-registry assumptions

That mixture makes it hard to answer what belongs everywhere and what belongs only in one surface.

## Decision Drivers

- Keep skills portable across surfaces
- Keep the registry small and easy to sync
- Prevent a monolithic skill blob
- Make updates deterministic and easy to validate
- Preserve the repository rule that broad shared rules stay in the global base
- Keep CLI, daemon, and UI outputs surface-specific without duplicating source of truth

## Options Considered

### Option 1: One master skill with everything inside

Pros:

- Simple at first glance

Cons:

- Mixes portable content with surface-specific behavior
- Hard to validate and port cleanly

### Option 2: Duplicate skills per surface

Pros:

- Each surface can tune its own output

Cons:

- Drift and maintenance cost grow quickly

### Option 3: Portable skill core plus surface adapters

Pros:

- Clean separation of concerns
- Portable and testable
- Easy to sync and regenerate

Cons:

- Adds a compilation and validation layer

## Decision

Adopt a portable skill core plus surface adapters.

## Rationale

This model keeps reusable skill content stable while allowing each surface to present the right execution shape. It is simpler to validate than duplicating skills and safer than merging every concern into one master blob.

## Consequences

### Positive

- Skills become portable and auditable
- The registry stays metadata-only and easy to sync
- CLI, daemon, UI, and IDE-specific outputs can evolve independently
- Surface-specific behavior no longer leaks into portable skill content

### Negative

- A compile and validation step is required
- The conceptual model becomes one layer richer

## Implementation Notes

- Portable skills live as folders with a required `SKILL.md`
- Reusable context packs are support material, not standalone skills
- `skills/registry.yml` should index metadata and paths only
- The daemon owns validation, compilation, sync, and derived indexes
- CLI and UI consume derived data only
- Surface-specific controls stay out of the portable core

## Related ADRs

- ADR-0002: Centralized Brain CLI for Multi-IDE Service Orchestration
- ADR-0003: Centralization of Orchestration in Daemon
