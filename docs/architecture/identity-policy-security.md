---
type: design-doc
id: identity-policy-security
title: Identity, Policy, and Security Architecture
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the hierarchy, authorization, trust, and runtime security model for Brain v2.

## Motivation

Brain is moving beyond a personal environment manager into a control plane suitable for teams and organizations. That requires explicit identity and policy layers rather than implicit local trust.

## Goals

- Define hierarchy across organization, team, user, workspace, project, and session.
- Support mandatory organization policy with lower-scope extensions.
- Add trust and permission checks for third-party artifacts.
- Separate security of stored data, imported artifacts, and runtime execution.

## Non-Goals

- Picking a specific external identity provider in this phase.
- Designing the full UI for account administration.
- Replacing operating-system security controls.

## High-Level Design

### Hierarchy

Brain resolves state using these scopes:

- `platform`
- `organization`
- `team`
- `user`
- `workspace`
- `project`
- `session`

Each artifact and policy record declares:

- scope
- owner
- precedence
- inheritance mode
- visibility

### Policy Classes

- `hard`
  - mandatory and non-overridable except by authorized break-glass workflows
- `guarded`
  - overridable only through explicit approval and audit logging
- `soft`
  - preference-level guidance

### Security Layers

1. Identity and access
   - login, sessions, service tokens, RBAC
2. Artifact trust
   - origin, signature or checksum, review status, permissions
3. Runtime isolation
   - sandboxing, allowed actions, scoped execution
4. Data protection
   - secret storage, encryption, sync integrity, audit trails

### Third-Party Artifact Controls

Every imported or installed artifact must declare:

- source origin
- trust status
- requested permissions
- review status
- environment visibility

Default policy is deny-by-default for privileged capabilities.

## API / Interface

### Data Structures

```typescript
interface PolicyRecord {
  id: string;
  scope: string;
  enforcement: "hard" | "guarded" | "soft";
  selectors: string[];
  rules: Record<string, unknown>;
}

interface ArtifactPermissions {
  fs: "none" | "read" | "write";
  net: "none" | "restricted" | "full";
  exec: "none" | "sandboxed" | "full";
  secrets: "none" | "scoped";
}
```

### Key Functions

```go
func Authenticate(input AuthInput) (*Session, error)
func Authorize(input AuthorizationInput) (*Decision, error)
func VerifyArtifactTrust(input TrustInput) (*TrustDecision, error)
func EvaluatePolicy(input PolicyEvaluationInput) (*PolicyDecision, error)
```

## Implementation Strategy

### Phase 1: Scope and Policy Metadata

- Add scope and enforcement fields to registries.
- Add explainable resolution traces.

### Phase 2: Identity Foundation

- Add account, org, workspace, and token data models.

### Phase 3: Trust and Permission Enforcement

- Add artifact verification and permission checks.
- Gate execution of untrusted or over-privileged artifacts.

### Phase 4: Encryption and Sync Security

- Encrypt sensitive local state.
- Add signed sync payloads and audit events.

## Trade-Offs

- Aspect: Local convenience vs enterprise control
- Option A: Trust local artifacts by default
- Option B: Treat origin and permissions as first-class policy data
- Chosen: B
- Rationale: Brain must scale to organizational use where local convenience cannot override governance.

## Risks & Mitigation

- Risk: Added security controls slow onboarding
- Severity: Medium
- Mitigation: Provide trust presets and explicit development profiles.

- Risk: Policy hierarchy becomes opaque
- Severity: Medium
- Mitigation: All decisions must produce explainability output.

## Testing Strategy

- [ ] Policy resolution tests across hierarchy scenarios
- [ ] Trust verification tests for local and imported artifacts
- [ ] Authorization tests for admin, user, and workspace roles

## Success Metrics

- Organization policy is always enforced across surfaces.
- Third-party artifacts cannot execute privileged actions without declared permissions.
- Every blocked action returns a deterministic and auditable reason.

## Deployment Plan

Introduce metadata and dry-run decisions first. Enforce blocking only after observability is in place.

## Monitoring & Observability

Security-relevant events must include:

- authentication
- authorization decisions
- policy overrides
- artifact verification results
- blocked executions
- secret access events

## Related Decisions

- `docs/adr/ADR-0010-hierarchical-identity-policy-and-security-model.md`
- `docs/adr/ADR-0005-strict-development-and-production-boundary.md`

---

**Status**: Active
**Reviewed by**: Brain Architecture Team
**Target completion**: 2026-04-21 for policy model approval
