---
id: ADR-0011
title: AI Runtime and Curator Subsystem
status: approved
date_created: 2026-04-07
language: en
type: architecture-decision
category: adr
version: 1.0.0
---

## ADR-0011: AI Runtime and Curator Subsystem

Status: APPROVED  
Deciders: Brain Architecture Team  
Date: 2026-04-07

## Context

Brain already manages providers and local runtimes indirectly, but the project now needs a first-class AI subsystem that can:

- run local open-source models
- integrate hosted APIs and remote self-hosted runtimes
- provide native Brain context to managed runtimes
- optimize and deduplicate artifact context over time

Without a dedicated subsystem, these concerns would remain fragmented across providers, daemon logic, and ad hoc scripts.

## Options Considered

### Option 1: Keep AI runtime concerns inside providers and daemon internals

Pros:

- Smallest immediate structural change
- Reuses existing provider configuration surface

Cons:

- Blurs providers, models, runtimes, and embeddings
- Leaves no clear home for context optimization jobs
- Harder to support self-hosted native flows cleanly

Trade-offs: Simpler near term, muddier architecture long term.

### Option 2: Create a first-class AI subsystem with runtimes, models, routing, and curator jobs (Chosen)

Pros:

- Clean separation of model and runtime concerns
- Supports local, hybrid, and cloud operation explicitly
- Provides a formal home for context optimization and curation

Cons:

- Adds a new domain and migration work
- Requires careful policy design for runtime data handling

Trade-offs: Slightly broader architecture now in exchange for better long-term clarity and product capability.

## Decision

Brain will introduce a first-class AI subsystem that manages runtimes, models, routing, embeddings, and a curator workflow for context optimization.

## Rationale

This is the smallest correct structure for the product direction. It allows Brain to improve user experience on self-hosted runtimes while keeping context efficient and reusable.

The curator is especially important because Brain's managed knowledge will grow over time. The system needs a formal mechanism to propose deduplication, promotion, and cleanup instead of relying on accidental manual maintenance.

## Consequences

### Positive Outcomes

- Local and hosted model flows are represented consistently
- Self-hosted runtime experience can become more native and efficient
- Curator scans can reduce context bloat across projects and scopes

### Risks and Mitigation

- Risk: Curator makes low-quality suggestions
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Start as dry-run with explicit user approval

- Risk: AI subsystem overlaps with provider config
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Treat providers as one artifact kind inside the wider AI domain

## Implementation Notes

Implementation details are documented in:

- `docs/architecture/ai-runtime-and-context-optimization.md`

## Related ADRs

- ADR-0007: Unified Capability Control Plane with Hierarchical Context Resolution (relationship: extends)
- ADR-0008: Unified Artifact Packaging and Lifecycle (relationship: depends_on)
- ADR-0010: Hierarchical Identity, Policy, and Security Model (relationship: constrained_by)

Status: Active  
Decided by: Brain Architecture Team  
Decision date: 2026-04-07  
Review date: 2026-06-30
