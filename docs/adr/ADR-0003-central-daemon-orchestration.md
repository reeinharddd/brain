---
type: adr
id: ADR-0003-central-daemon-orchestration
title: Centralization of Orchestration in Daemon
version: 1.0.0
status: accepted
date_created: 2026-04-02
language: en
category: architecture
related: []
keywords:
  - daemon
  - orchestration
  - mcp
  - providers
  - docker
rag_priority: high
chunk_strategy: section
---

## ADR-0003: Centralization of Orchestration in Daemon

## Context

The Brain repository had several orchestration problems:

- duplicate Docker Compose files created drift and port conflicts
- Open WebUI was included but not part of the current product direction
- MCP services ran in multiple modes, which made behavior harder to reason about
- the daemon was passive instead of acting as the active orchestrator
- LLM routing existed in configuration but was not centrally enforced
- users had to manage service startup manually

## Decision Drivers

- Keep Brain centered on one orchestrator
- Reduce configuration drift and operational duplication
- Prefer stdio MCP execution over scattered bridges
- Make provider routing and fallback behavior explicit
- Preserve the daemon/CLI/UI separation of concerns

## Options Considered

### Option 1: Keep orchestration distributed

Pros:

- Lower initial change

Cons:

- Duplicated control paths
- More drift and harder debugging

### Option 2: Use Kubernetes instead of Docker Compose

Pros:

- Strong orchestration model

Cons:

- Too much complexity for a local developer environment
- Harder to maintain and debug

### Option 3: Central daemon with shared runtime managers

Pros:

- One control plane
- Easier to validate and observe
- Keeps provider and MCP routing centralized

Cons:

- Daemon becomes more complex

## Decision

Adopt a central daemon that actively orchestrates Docker, MCP sync, provider selection, and runtime health.

## Rationale

A central daemon makes the runtime easier to observe, reduces configuration drift, and keeps service management in one place. It also gives the CLI and UI a stable control plane instead of independent orchestration logic.

## Consequences

### Positive

- Single source of truth for orchestration
- Reduced configuration drift
- Better observability and cleaner control flow
- Easier to add new services without changing every client surface

### Negative

- The daemon carries more responsibility
- Startup and health management must be implemented carefully

## Implementation Notes

- Use dedicated managers for Docker, MCP registry, providers, and health
- Keep CLI commands thin and request-driven
- Prefer stdio MCP mode for Brain-owned integrations
- Fail closed for production-sensitive orchestration paths

## Related ADRs

- ADR-0002: Centralized Brain CLI for Multi-IDE Service Orchestration
- ADR-0005: Strict Development and Production Boundary
