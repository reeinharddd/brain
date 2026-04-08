---
type: design-doc
id: control-plane-domain-map
title: Control Plane Domain Map
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the primary Brain v2 control-plane domains and their responsibilities.

## Domains

### Execution Surfaces

- `apps/cli`
- `apps/daemon`
- `apps/desktop`

### Shared Core

- `core/runtime`
  - root resolution
  - configured root persistence
  - workspace identity for the local machine
- `core/artifacts`
  - canonical domain resolution
  - legacy compatibility lookup
  - artifact file and directory location helpers

### Artifact Domains

- `artifacts/agents`
- `artifacts/commands`
- `artifacts/rules`
- `artifacts/adapters`
- `artifacts/mcps`
- `artifacts/skills`
- `artifacts/providers`
- `artifacts/memory`
- `artifacts/ai`
- `artifacts/identity`
- `artifacts/policy`

## Domain Rules

- Every governed domain should have:
  - a canonical artifact directory
  - a `manifests/` directory
  - schema-backed manifests for managed units
- `artifacts/` contains operational Markdown and data needed by Brain.
- `docs/` contains canonical human documentation.
- Legacy directories may exist temporarily only as compatibility shims.

## Current Transition Status

- `agents`, `commands`, `rules`, `adapters`, `mcps`, `skills`, and `providers` already have canonical artifact roots.
- `memory`, `ai`, `identity`, and `policy` are now formal v2 domains with baseline manifests and schemas.
- Runtime migration to remove legacy path assumptions is still in progress.

## Related Documents

- `docs/architecture/brain-v2-target-architecture.md`
- `docs/architecture/repository-structure-v2.md`
- `docs/architecture/artifact-system-contract.md`
- `docs/architecture/memory-and-knowledge-architecture.md`
- `docs/architecture/identity-policy-security.md`
