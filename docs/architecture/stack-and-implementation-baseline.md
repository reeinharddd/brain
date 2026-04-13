---
type: design-doc
id: stack-and-implementation-baseline
title: Stack and Implementation Baseline
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the intended stack baseline for Brain.

## Product Stack

Brain is built around three product surfaces:

- daemon
- CLI
- desktop UI

## Preferred Technology Baseline

### Core Runtime

- Go for daemon, CLI, sync, validation, policy, and other production-facing operational logic

### Desktop Surface

- TypeScript
- React
- Tauri

### Artifact Storage and Configuration

- YAML and JSON for machine-readable metadata
- Markdown for human-authored artifact payloads and canonical docs

### Observability

- structured logs
- event traces
- health status endpoints
- future metrics and audit streams

### AI Runtime Integration

- provider-backed hosted APIs
- OpenAI-compatible runtimes
- local managed runtimes such as Ollama and future equivalents

For concrete vendor and service choices for identity, database, cache, storage, edge, observability, backups, and secrets, see `docs/architecture/infrastructure-baseline-canonical.md`.

## Stack Usage Matrix

The chosen stack is not just a list of tools; it defines where work happens and which
surfaces are allowed to own behavior.

| Stack area                            | Local personal                                                                               | Self-hosted corporate                                                      | Hosted cloud                                                                        | Real impact on the app                                                                       |
| ------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Go daemon and core services           | Runs as the single source of truth on one machine, with the smallest viable dependency set.  | Runs the same code with operator-owned services and stricter governance.   | Runs the same code against Brain-operated infrastructure and multi-tenant controls. | Keeps business logic, policy, sync, MCP orchestration, and validation out of the UI and CLI. |
| Go CLI                                | Talks to the local daemon and stays thin.                                                    | Talks to the self-hosted daemon and inherits its auth and policy.          | Talks to the hosted daemon and inherits tenant and billing context.                 | Prevents command behavior from drifting away from daemon state.                              |
| TypeScript + React + Tauri desktop UI | Renders localhost-backed projections and can run with optional local-only services disabled. | Renders the same projections against private infrastructure.               | Renders the same projections against Brain-operated infrastructure.                 | Makes the desktop a client of daemon state, not a second orchestrator.                       |
| YAML, JSON, and Markdown artifacts    | Stores registries, configs, and docs in repo-friendly formats that can survive offline use.  | Stores the same canonical artifacts with enterprise sync and review flows. | Stores the same canonical artifacts with cloud sync and audit paths.                | Keeps artifact schemas stable so registry, sync, docs, and migration code all agree.         |
| Observability stack                   | Starts with console logs and health checks; metrics can remain minimal.                      | Expands to structured logs, metrics, traces, and alerts.                   | Expands to the same observability stack with hosted-scale retention and routing.    | Drives the status screens, health endpoints, and troubleshooting paths in the app.           |
| AI runtime integration                | Can use local models or hosted APIs, depending on the user setup.                            | Usually uses organization-approved providers and local fallback models.    | Usually uses managed provider chains and budget-aware routing.                      | Affects context optimization, cost tracking, model selection, and fallback behavior.         |

## Cross-Cutting Impact

The stack choices above affect more than deployment. They shape the rest of the repository in practical ways:

- `identity-policy-security.md` depends on the daemon being the enforcement point, so auth and policy stay server-side.
- `memory-and-knowledge-architecture.md` depends on the storage layer being profile-aware, so local deployments can stay light while larger profiles add Qdrant and backup flows.
- `ai-runtime-and-context-optimization.md` depends on the AI runtime row, so provider routing, token budgeting, and fallbacks are treated as daemon responsibilities.
- `artifact-system-contract.md` depends on YAML, JSON, and Markdown staying canonical, so sync, validation, and migrations can share one schema contract.
- `brain-v2-target-architecture.md` and `environment-configuration.md` depend on the profile and environment split, so production packaging stays allowlisted while local mode stays flexible.
- The desktop app must remain a projection layer, so new UI sections should request data from daemon endpoints instead of inventing their own state model.
- The CLI must remain thin, so command behavior should reflect daemon state and not duplicate orchestration logic.

## Implementation Rules

- production-facing operational workflows should be implemented in Go
- desktop behavior should remain a client of daemon state, not a second orchestrator
- no surface should own canonical state outside daemon-governed contracts
- machine-readable schemas should exist for all important registries and artifact types

## Deployment Profile Baseline

Brain must support one codebase with multiple deployment profiles:

- local personal: lightweight, single-user, minimal infrastructure
- self-hosted corporate: operator-owned storage, auth, and private networking
- hosted cloud: Brain-operated cloud services with multi-tenant scale and edge protection

Local mode should stay intentionally small. Heavy services such as CDN, Redis, Qdrant, and enterprise SSO should remain optional in local deployments rather than becoming unconditional startup requirements.

See `docs/architecture/deployment-profiles-and-infrastructure-baseline.md` for the full profile matrix.

## Long-Term Structure

The implementation target is organized around:

- `apps/`
- `core/`
- `artifacts/`
- `internal/`
- `deploy/`

## Compatibility Goal

Brain should remain usable across older and newer model families by:

- keeping canonical instructions structured and compact
- separating baseline context from task-local context
- preferring resolution and projection over prompt bloat
- degrading gracefully when advanced agent features are unavailable
