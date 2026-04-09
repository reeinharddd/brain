---
type: design-doc
id: testing-strategy
title: Testing, Linting, and CI Strategy
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the future-facing quality baseline for Brain.

## Goals

- One coherent quality model across CLI, daemon, desktop, artifacts, and docs
- Fast local feedback for development
- Deterministic checks in CI
- Clear separation between required gates and optional deep validation

## Non-Goals

- Replacing all ad hoc developer judgment with automation
- Turning documentation validation into a full semantic review of prose quality
- Enforcing release-only policies on local development-only scratch files

## Required Quality Gates

- formatting
- linting
- unit tests
- integration tests for daemon-centered flows
- documentation validation
- security and secret scanning

## How to Test

Local validation:

- Run the docs boundary check before committing changes that touch `docs/` or `artifacts/`
- Run the docs validator against changed files to confirm manifest alignment and section coverage

CI validation:

- Confirm pull requests fail when docs break the manifest, frontmatter, or structure rules
- Confirm the merge gate runs the same validation path as the pre-commit hook

## Execution Model

Local development should optimize for speed and confidence.

CI should optimize for determinism, reproducibility, and merge safety.

Production release validation should only use production-safe tooling and artifacts.

## Layering

### Local

- fast format and lint checks
- focused unit and integration runs
- docs validation

### CI

- full lint
- full unit test matrix
- integration tests
- artifact schema validation
- policy and security checks

### Release

- production profile validation
- environment boundary enforcement
- release packaging verification

## Rules

- All required quality commands must be invokable through Brain-owned workflows.
- Docs checks must validate canonical docs, not generated working notes.
- Dev-only tools must not be required to validate production-safe releases.
- Quality gates must be defined before refactors land, not after.
