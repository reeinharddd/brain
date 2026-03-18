---
name: codebase-contextualizer
description: Builds a project-scoped context pack from docs, READMEs, ADRs, and architecture notes, with optional vector-index handoff.
---

# Codebase Contextualizer

Use this skill when you need reusable project context before planning, implementation, or review.

## Capabilities

- Detect project namespace with `~/.brain/scripts/memory-namespace.sh`
- Render stack-aware skill context with `~/.brain/scripts/render-skill-context.sh`
- Build a lightweight codebase context index with `~/.brain/scripts/vector-context-index.sh`

## Usage

```bash
bash ~/.brain/skills/codebase-contextualizer/contextualize.sh [project_root]
```
