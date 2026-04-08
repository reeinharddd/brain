---
id: ADR-0010
title: Hierarchical Identity, Policy, and Security Model
status: approved
date_created: 2026-04-07
language: en
type: architecture-decision
category: adr
version: 1.0.0
---

## ADR-0010: Hierarchical Identity, Policy, and Security Model

Status: APPROVED  
Deciders: Brain Architecture Team  
Date: 2026-04-07

## Context

Brain is intended to support not only personal environments but also teams and organizations. In that setting, capability activation cannot be based on local convention alone. The system needs explicit hierarchy, mandatory policy, artifact trust controls, and auditable authorization.

## Options Considered

### Option 1: Keep local-first trust with limited optional policy

Pros:

- Faster to build
- Minimal friction for personal usage

Cons:

- Weak fit for organizational governance
- Third-party artifact risk remains too implicit
- Hard to enforce company-wide requirements

Trade-offs: Convenience over control.

### Option 2: Adopt explicit hierarchy, policy classes, and trust enforcement (Chosen)

Pros:

- Supports organizational and user-level coexistence
- Enables least-privilege execution and auditability
- Makes third-party artifact handling explicit

Cons:

- Higher implementation complexity
- Requires explainability to stay usable

Trade-offs: More system complexity in exchange for enterprise-grade governance and safety.

## Decision

Brain will use a hierarchical identity, policy, and security model with explicit scope layers, policy classes, trust metadata, and auditable authorization outcomes.

## Rationale

This decision is required for Brain to safely support:

- organization-wide mandatory baselines
- user-specific preferences
- workspace and project specificity
- third-party artifact installation and execution
- cloud sync across devices

## Consequences

### Positive Outcomes

- Organization policy can be enforced without destroying user flexibility
- Artifact trust and requested permissions become visible and enforceable
- Execution decisions become auditable across all surfaces

### Risks and Mitigation

- Risk: Policy engine becomes difficult to understand
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Require explainability output for every policy decision

- Risk: Security defaults slow adoption
  - Probability: Medium
  - Impact: Medium
  - Mitigation: Use dry-run mode and profile-based onboarding

## Implementation Notes

Implementation details are documented in:

- `docs/architecture/identity-policy-security.md`

## Related ADRs

- ADR-0005: Strict Development and Production Boundary (relationship: extends)
- ADR-0007: Unified Capability Control Plane with Hierarchical Context Resolution (relationship: enables)
- ADR-0008: Unified Artifact Packaging and Lifecycle (relationship: depends_on)

Status: Active  
Decided by: Brain Architecture Team  
Decision date: 2026-04-07  
Review date: 2026-06-30
