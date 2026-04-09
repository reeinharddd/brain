---
type: guidance
id: structure-map
title: Documentation Structure Map
version: 2.2.0
status: active
date_created: 2026-04-07
language: en
category: documentation
---

## Structure Map

## Root Level

```text
docs/
├── README.md
├── QUICK-START.md
├── STRUCTURE.md
├── DOCUMENTATION-UNIFIED-SCHEMA.md
└── INDEX.md
```

## Organization by Domain

```text
docs/
├── adr/
│   └── ADR-*.md
├── architecture/
│   ├── brain-v2-target-architecture.md
│   ├── repository-structure-v2.md
│   ├── artifact-system-contract.md
│   ├── identity-policy-security.md
│   ├── ai-runtime-and-context-optimization.md
│   └── other architecture docs
├── skills/
│   ├── README.md
│   └── skill-artifact-authoring.md
├── testing/
│   ├── README.md
│   └── testing-strategy.md
└── templates/
```

## What Goes Where

- Why a decision was made: `adr/`
- How the system should work: `architecture/`
- Skill-specific guidance: `skills/`
- Validation guidance: `testing/`
- Historical material: keep in git history unless a future archive contract is explicitly restored
- Reusable templates: `templates/`

## Current Navigation Rule

If a document defines active project direction, it must be reachable from:

- `docs/README.md`
- `docs/INDEX.md`
- the relevant domain README
- the relevant domain README for testing, skills, architecture, or ADRs

## Current Priority Path

The canonical path for the current architecture closure is:

1. `docs/architecture/brain-v2-target-architecture.md`
2. `docs/adr/ADR-0007-unified-capability-control-plane.md`
3. `docs/adr/ADR-0008-unified-artifact-packaging-and-lifecycle.md`
4. `docs/adr/ADR-0009-clean-repository-structure-and-domain-boundaries.md`
5. `docs/adr/ADR-0010-hierarchical-identity-policy-and-security-model.md`
6. `docs/adr/ADR-0011-ai-runtime-and-curator-subsystem.md`
