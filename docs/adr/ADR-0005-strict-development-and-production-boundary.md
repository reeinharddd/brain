---
type: adr
id: ADR-0005-strict-development-and-production-boundary
title: Strict Development and Production Boundary
version: 1.0.0
status: accepted
date_created: 2026-04-03
language: en
category: architecture
related: []
keywords:
  - environment
  - production
  - development
  - visibility
  - packaging
rag_priority: high
chunk_strategy: section
---

## ADR-0005: Strict Development and Production Boundary

## Context

Brain is both a product runtime and a development platform. That makes it easy for tooling intended for developers to leak into production if the boundary is not explicit.

The repository includes or may include:

- developer-only documentation and architecture notes
- tests and fixtures
- CI/CD workflows
- internal tooling such as RAG-backed documentation MCPs
- runtime components that must remain available in production

Without an explicit boundary, the project risks shipping development-only artifacts, exposing internal information, or making the production runtime harder to audit.

## Options Considered

### Option 1: Keep development and production implicit

Pros:

- Minimal setup

Cons:

- Easy to leak development-only tools into production
- Hard to audit or reason about the runtime boundary

### Option 2: Separate development and production into different repositories

Pros:

- Very clear separation

Cons:

- Higher maintenance burden
- Harder to keep shared logic aligned

### Option 3: Single repository with explicit environment classification

Pros:

- Keeps shared logic in one place
- Lets the daemon enforce the boundary directly
- Works well with allowlists and packaging checks

Cons:

- Requires explicit classification for every new component

## Decision

Treat development and production as separate operational modes with explicit component classification.

## Rationale

A single repository with explicit environment classification gives Brain the best trade-off between maintainability and safety. It avoids duplicate repositories while still allowing the daemon and packaging layer to fail closed when development-only content is about to reach production.

## Decision Rules

- The runtime must read an explicit environment value.
- Components must declare whether they are `dev-only` or `prod-safe`.
- The daemon must refuse to load `dev-only` components in production.
- Production packaging must use an allowlist and must not ship development-only files.
- Developer tooling such as the documentation RAG MCP must remain development-only unless a separate production-safe use case is explicitly approved.
- Any operational script or script-like helper that is intended to ship with production must be implemented as a Go executable.
- Shell or Python scripts are allowed only for development convenience, local experimentation, or transitional compatibility and must remain outside the production runtime.
- If a script-like workflow is needed during development, the Go implementation is the source of truth and any wrapper script must be clearly marked as dev-only.

## Consequences

### Positive

- Production remains minimal and easier to audit
- Development tools stay available without contaminating the shipped runtime
- New components can be classified centrally instead of being hidden in scattered conditionals
- Security review becomes simpler because the production bundle has a smaller attack surface

### Negative

- Build and runtime logic becomes slightly more explicit
- New components must be classified before they are accepted
- Release packaging must be validated instead of assumed
- Development helpers may still exist, but only when they are not part of the production surface area

## Implementation Notes

- Use explicit environment configuration
- Keep component metadata centralized
- Fail closed in production
- Add build and CI checks that reject accidental dev-only inclusion
- Prefer Go binaries for any future production-facing automation, including validators, sync jobs, and operational helpers
- Treat existing shell helpers as development-only until they are replaced or wrapped by Go executables

## Related ADRs

- ADR-0003: Centralization of Orchestration in Daemon
