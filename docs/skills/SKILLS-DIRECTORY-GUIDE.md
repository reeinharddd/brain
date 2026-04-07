# Skills Directory Guide

## Purpose

This directory holds the registry, context packs, and skill metadata that power the Brain skills system.

## Current Structure

- `registry.yml` - master registry for skills
- `dynamic-registry.tsv` - master registry for context packs
- `contexts/` - reusable context packs for technology stacks
- skill folders - portable skill definitions with `SKILL.md`

## What Each Part Does

### registry.yml

- Stores the canonical skill metadata
- Is the source of truth for skill entries
- Must stay in sync with the filesystem

### dynamic-registry.tsv

- Stores context-pack metadata
- Supports stack-aware context injection
- Must stay in sync with the filesystem

### contexts/

- Contains reusable guidance packs by stack
- Example topics: Go service, React UI, Markdown, Python service
- Used when the daemon detects a matching project stack

## Validation Model

- The daemon validates registry-to-filesystem sync.
- The CLI exposes `brain skills validate`.
- The UI shows the current sync state.
- Production-facing helpers must be Go-based.

## Safe Workflow

1. Register the item.
2. Create the matching file or directory.
3. Validate with the daemon.
4. Commit only when sync is green.

## Do Not

- Create orphan directories.
- Add registry entries without matching files.
- Rely on manual shell scripts for production validation.
- Let development helpers leak into production packaging.

## Outcome

The directory stays auditable, deterministic, and safe for production while still remaining convenient for development.
