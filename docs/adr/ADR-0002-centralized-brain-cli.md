---
type: adr
id: ADR-0002-centralized-brain-cli
title: Centralized Brain CLI for Multi-IDE Service Orchestration
version: 1.0.0
status: accepted
date_created: 2026-03-31
language: en
category: architecture
related: []
keywords:
  - cli
  - orchestration
  - autostart
  - multi-ide
  - mcp-gateway
rag_priority: high
chunk_strategy: section
---

## ADR-0002: Centralized Brain CLI for Multi-IDE Service Orchestration

## Context

The Brain repository provides a unified knowledge base and tool infrastructure across multiple IDEs and LLM providers. Service orchestration was fragmented: boot-time initialization was global, MCP servers launched independently, autostart was not opt-in, and service status was hard to inspect.

## Decision Drivers

- Preserve the repository's rules-first architecture
- Keep the CLI thin and user-facing
- Avoid port conflicts across IDEs
- Make autostart explicit and reversible
- Improve health checks and observability

## Options Considered

### Option 1: Leave orchestration distributed

Pros:

- Lowest immediate change

Cons:

- Keeps port conflicts and inconsistent state
- Hard to debug or audit

### Option 2: Centralize everything in the CLI

Pros:

- One command can control the system

Cons:

- The CLI would become a second source of truth
- Harder to support the daemon and UI split

### Option 3: Centralized CLI with a daemon-backed control plane

Pros:

- Keeps the CLI thin
- Gives one stable entry point for users
- Supports shared state, health checks, and opt-in autostart

Cons:

- Requires daemon coordination and a little more wiring

## Decision

Adopt a centralized `brain` CLI that delegates orchestration to the daemon and exposes a single control surface for start, stop, health, status, configuration, and logs.

## Rationale

A daemon-backed CLI preserves a single entry point for users while keeping the CLI itself thin. That avoids duplicate orchestration logic and aligns with the daemon/CLI/UI separation used across the repository.

## Consequences

### Positive

- Single point of control for service management
- Opt-in autostart instead of forced boot behavior
- No port conflicts when multiple IDEs connect to the same MCP gateway
- Better debugging through centralized health checks and logs

### Negative

- Users must learn the `brain` command
- The daemon becomes more important operationally

## Implementation Notes

- Use configuration-driven behavior for autostart and health intervals
- Keep project initialization separate from boot-time service startup
- Use the shared MCP gateway model for all IDEs
- Log service state changes to the telemetry stream

## Related ADRs

- ADR-0003: Centralization of Orchestration in Daemon
- ADR-0005: Strict Development and Production Boundary
