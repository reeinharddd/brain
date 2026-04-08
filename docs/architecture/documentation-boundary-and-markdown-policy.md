---
type: design-doc
id: documentation-boundary-and-markdown-policy
title: Documentation Boundary and Markdown Policy
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the canonical boundary for Markdown content in the Brain repository.

## Purpose

Brain uses Markdown for two different purposes:

1. canonical documentation
2. operational artifact payloads

Those two categories must not be confused.

## Canonical Rule

All canonical documentation must live under `docs/`.

Canonical documentation includes:

- architecture and design documents
- ADRs
- documentation rules
- future-facing skill and testing guidance
- reusable templates

## Non-Canonical Markdown

Markdown outside `docs/` may still exist only when it is part of the runtime or artifact payload layer.

The canonical operational location is now `artifacts/`.

Examples:

- `artifacts/agents/*.md`
- `artifacts/commands/*.md`
- `artifacts/rules/**/*.md`
- `artifacts/adapters/**/*.md`
- `artifacts/mcps/**/*.md`

These files are not treated as documentation pages. They are operational source artifacts consumed by Brain or by external tools.

## Long-Term Direction

The long-term target is:

- documentation Markdown only in `docs/`
- operational artifacts migrated into the unified `artifacts/` system

Until that migration is complete, legacy paths such as `agents/`, `commands/`, `rules/`, `adapters/`, `mcp/`, `guardian/`, and `hooks/` may still exist only as compatibility shims.

Legacy locations must not host canonical Markdown content directly.

## Prohibited Cases

Markdown outside `docs/` is not allowed for:

- ad hoc planning notes
- generated summaries
- delivery reports
- obsolete implementation diaries
- duplicate copies of canonical docs

Those belong either in `docs/` if canonical or must be removed if temporary.

## Naming Rules

- canonical documents under `docs/` should use stable names and kebab-case where practical
- established index names such as `README.md` and `INDEX.md` are allowed as navigation anchors
- template entry files may use `TEMPLATE.md` where that convention improves discoverability
- operational artifact Markdown may follow runtime-driven naming conventions until migrated

## Enforcement Direction

Future documentation validation should enforce:

- no non-canonical working docs outside `docs/`
- no duplicated canonical content between `docs/` and runtime artifact paths
- explicit classification for Markdown outside `docs/`

The repository now enforces this boundary with:

- `utils/scripts/validate-markdown-boundary.sh`
- `.github/workflows/docs-boundary.yml`

## Decision

Brain will treat `docs/` as the only canonical documentation root, while allowing operational Markdown under `artifacts/` and temporary legacy symlink shims during migration.
