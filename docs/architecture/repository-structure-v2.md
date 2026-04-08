---
type: design-doc
id: repository-structure-v2
title: Repository Structure v2
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the target repository layout for Brain v2.

The main objective is a clean root with obvious subsystem boundaries. The repository root should explain the product in one screen instead of listing every artifact domain as a peer concept.

## Motivation

The current repository mixes executable applications, canonical artifacts, development helpers, and generated or support material near the root. That shape makes sense during early exploration but does not scale well once Brain includes:

- multiple user surfaces
- hierarchical policy and identity
- cloud sync
- AI runtimes
- internal-only and production-safe artifacts

## Goals

- Keep the root small and stable.
- Separate executable product code from managed artifacts.
- Create clear boundaries between public runtime, internal tooling, and deployment.
- Support future plugin and marketplace workflows without root sprawl.

## Non-Goals

- Fully rewriting all paths immediately.
- Forcing a single deploy shape for every environment.
- Hiding important concepts behind unclear names.

## High-Level Design

### Root Layout

```text
/
├── apps/
├── core/
├── artifacts/
├── internal/
├── docs/
├── utils/
├── deploy/
├── README.md
├── manifest.yml
└── testconfig.yml
```

### Directory Responsibilities

- `apps/`
  - all user-facing executable surfaces
  - `cli/`, `daemon/`, `desktop/`, future `tui/`

- `core/`
  - shared product subsystems
  - registry, sync, policy, identity, security, context, observability, projection

- `artifacts/`
  - canonical source-managed assets governed by Brain
  - rules, skills, agents, commands, MCPs, providers, AI runtimes, templates

- `internal/`
  - development-only implementation support
  - migrations, bootstrap, testdata, generators, compatibility shims

- `docs/`
  - only valid project documentation

- `utils/`
  - convenience scripts and non-product helper assets

- `deploy/`
  - environment-specific deployment and release assets

### Artifact Subtree

```text
artifacts/
├── rules/
├── skills/
├── agents/
├── commands/
├── mcps/
├── providers/
├── ai/
└── templates/
```

### Transitional Mapping from Current Structure

- `cli/` -> `apps/cli/`
- `daemon/` -> `apps/daemon/`
- `desktop/` -> `apps/desktop/`
- `rules/` -> `artifacts/rules/`
- `skills/` -> `artifacts/skills/`
- `agents/` -> `artifacts/agents/`
- `commands/` -> `artifacts/commands/`
- `mcp/` -> `artifacts/mcps/`
- `providers/` -> `artifacts/providers/`

## Implementation Strategy

### Phase 1: Introduce New Structure in Parallel

- Create new directories.
- Add compatibility reads from old and new paths.
- Move canonical operational Markdown into `artifacts/`.
- Keep legacy directories as symlink-based shims only.

### Phase 2: Move Canonical Sources

- Migrate artifacts first.
- Update registry loading to prefer the new paths.

### Phase 3: Move Application Code

- Relocate CLI, daemon, and desktop modules.
- Update build and test paths.

### Phase 4: Remove Transitional Shims

- Drop old path support after validation.

## Trade-Offs

- Aspect: Explicit `artifacts/` subtree
- Option A: Keep each domain at root
- Option B: Group all managed domains under `artifacts/`
- Chosen: B
- Rationale: Brain manages many artifact kinds, so the repository should make that explicit instead of flattening them.

- Aspect: `internal/` versus `utils/`
- Option A: Use one helper folder for everything
- Option B: Separate product-internal support from general utilities
- Chosen: B
- Rationale: Internal code and convenience scripts have different stability and exposure requirements.

## Risks & Mitigation

- Risk: Path churn breaks existing commands
- Severity: High
- Mitigation: Add compatibility resolution and migrate incrementally.

- Risk: Team confusion during transition
- Severity: Medium
- Mitigation: Update docs and provide one migration map.

## Testing Strategy

- [ ] Path resolution unit tests
- [ ] Integration tests for registry loading from new paths
- [ ] CLI and daemon smoke tests after relocation

## Success Metrics

- Root-level directories remain under a controlled set of product categories.
- New features can be placed without inventing new root conventions.
- Old and new paths can coexist during migration without drift.

## Deployment Plan

Introduce new structure first, then gradually move sources. No big-bang move.

## Monitoring & Observability

During migration, daemon logs must report which path set was used for:

- artifact loading
- sync projection
- compatibility fallback

## Related Decisions

- `docs/adr/ADR-0009-clean-repository-structure-and-domain-boundaries.md`
- `docs/architecture/brain-v2-target-architecture.md`

---

**Status**: Active
**Reviewed by**: Brain Architecture Team
**Target completion**: 2026-04-21 for migration plan approval
