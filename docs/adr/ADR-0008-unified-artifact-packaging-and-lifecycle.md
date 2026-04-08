---
id: ADR-0008
title: Unified Artifact Packaging and Lifecycle
status: approved
date_created: 2026-04-07
language: en
type: architecture-decision
category: adr
version: 1.0.0
---

## ADR-0008: Unified Artifact Packaging and Lifecycle

Status: APPROVED  
Deciders: Brain Architecture Team  
Date: 2026-04-07

## Context

Brain manages multiple domains that are currently represented with inconsistent contracts and lifecycle rules:

- rules
- skills
- agents
- commands
- MCP definitions
- providers

The product direction now includes cloud sync, local and remote installation flows, trust verification, and hierarchical resolution. Those capabilities become brittle if each domain continues to invent its own lifecycle and metadata shape.

## Options Considered

### Option 1: Keep per-domain packaging and lifecycle rules

Pros:

- Lowest short-term migration effort
- Maximum local flexibility inside each domain

Cons:

- Duplicated parsing and lifecycle logic
- Harder sync, security, and explainability
- More edge cases for import and install flows

Trade-offs: Preserves local freedom but increases long-term system complexity.

### Option 2: Use a common artifact envelope with domain extensions (Chosen)

Pros:

- One lifecycle model across all managed domains
- Shared trust, sync, and scope metadata
- Cleaner support for create, import, install, link, and sync flows

Cons:

- Requires schema work and compatibility shims
- Some domains need migration from simple registries to package folders

Trade-offs: More upfront structure in exchange for lower long-term operational cost.

## Decision

Brain will use a unified artifact envelope and lifecycle model for all managed domains, with domain-specific extensions where needed.

## Rationale

This approach best supports the product direction:

- cloud-syncable user and organization state
- secure import and installation of third-party content
- deterministic hierarchical resolution
- explainable activation and rejection decisions

The shared envelope gives Brain one language for governance while preserving differences in payload shape.

## Consequences

### Positive Outcomes

- Create, import, install, link, and sync become platform features instead of per-domain inventions
- Security and trust metadata become universal
- Hierarchical and policy-aware resolution becomes easier to implement consistently

### Risks and Mitigation

- Risk: Migration burden for current registries
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Add compatibility readers before enforcing write-path migration

- Risk: Common envelope becomes bloated
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Keep the base schema minimal and push specifics into domain extensions

## Implementation Notes

Implementation details are documented in:

- `docs/architecture/artifact-system-contract.md`

Initial rollout:

- define base schema
- normalize read paths
- add trust and lifecycle metadata
- migrate write paths incrementally

## Related ADRs

- ADR-0007: Unified Capability Control Plane with Hierarchical Context Resolution (relationship: enables)
- ADR-0009: Clean Repository Structure and Domain Boundaries (relationship: complements)

Status: Active  
Decided by: Brain Architecture Team  
Decision date: 2026-04-07  
Review date: 2026-06-30
