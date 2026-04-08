---
type: design-doc
id: artifact-migration-and-compatibility-plan
title: Artifact Migration and Compatibility Plan
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document records the first live migration step from the legacy repository layout to the Brain v2 artifact layout.

## Goal

Move operational markdown artifacts into `artifacts/` without breaking current runtime behavior that still resolves legacy paths.

## Current Strategy

Brain now stores the canonical markdown artifacts in:

- `artifacts/agents/`
- `artifacts/commands/`
- `artifacts/rules/`
- `artifacts/adapters/`
- `artifacts/mcps/`
- `artifacts/hooks/`
- `artifacts/internal/guardian/`

Legacy locations remain present only as compatibility shims.

## Compatibility Rule

During migration:

- legacy directories may remain
- markdown files in legacy directories must be symlinks only
- canonical content lives in `artifacts/`

This preserves the current daemon and CLI behavior while the loaders are updated.

## Why This Approach

- avoids a flag day refactor
- keeps existing runtime path assumptions working
- gives the project a clean canonical storage layout immediately
- allows gradual code migration and test updates

## Next Code Changes

1. update agent loading code to read from `artifacts/agents/`
2. update sync engine source paths to the new artifact tree
3. update adapters and config generators to emit `artifacts/` references
4. remove compatibility symlinks after all loaders and tests are migrated

## Validation

The repository now includes an automated boundary validator:

- `utils/scripts/validate-markdown-boundary.sh`

This script enforces that:

- canonical docs live under `docs/`
- operational markdown artifacts live under `artifacts/`
- legacy locations contain symlink shims only

## Related Documents

- `docs/architecture/repository-structure-v2.md`
- `docs/architecture/documentation-boundary-and-markdown-policy.md`
- `docs/architecture/artifact-system-contract.md`
