---
type: architecture
id: documentation-enforcement-system
title: Brain Documentation Enforcement System
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: architecture
keywords:
  - documentation
  - enforcement
  - validation
  - automation
  - pre-commit
rag_priority: high
chunk_strategy: section
---

## Brain Documentation Enforcement System

## Purpose

Documentation in Brain is validated automatically before commit. The goal is to keep the documentation baseline consistent, English-only, and easy to navigate.

## Enforcement Layers

- Pre-commit hook
- Incremental validator
- Manifest baseline
- Documentation schema and structure rules
- CI workflow for pull requests and main-branch pushes

## What Is Enforced

- Required frontmatter fields
- English-only text
- File naming conventions
- Markdown syntax correctness
- Template and section structure
- Domain `README.md` files are reference-level navigation hubs, not strict design docs
- `docs/templates/` remains reference-level guidance and examples

## Where to Fix Issues

- `docs/README.md`
- `docs/STRUCTURE.md`
- `docs/INDEX.md`
- `docs/testing/README.md`
- `docs/testing/testing-strategy.md`
- `docs/metadata/docs-manifest.json`
- `docs/metadata/docs-changelog.jsonl`
- `.githooks/pre-commit`
- `.github/workflows/docs-validation.yml`
- `utils/scripts/validate-docs-incremental.sh`
- `utils/scripts/validate-markdown-boundary.sh`
- `utils/scripts/validate-artifact-manifests.py`

## Outcome

The documentation system stays clean, predictable, and safe to change without manual review of every file.
