---
type: guidance
id: structure-map
title: Documentation Structure Map
version: 2.1.0
status: active
date_created: 2026-04-03
language: en
category: documentation
---

## Structure Map

## Root Level (5 Essential Files)

```text
docs/
├── README.md                          ← Start here
├── QUICK-START.md                     ← 5-min introduction
├── STRUCTURE.md                       ← This file
├── DOCUMENTATION-UNIFIED-SCHEMA.md    ← Standards reference
└── INDEX.md                           ← Master navigation
```

## Organization by Domain (6 Folders)

```text
docs/
├── adr/                 # Architecture Decision Records
│   └── *.md            # Technical decisions with context & rationale
│
├── architecture/        # System design, contracts & integration
│   ├── daemon-orchestration.md
│   ├── cli-integration.md
│   ├── agent-contracts.md              ← Agent input/output contracts
│   ├── skill-system-contract.md       ← Skill system architecture
│   └── README.md
│
├── skills/             # LLM agent skill definitions
│   ├── *-architecture.md
│   ├── *-enforcement.md
│   └── *-quick-ref.md
│
├── testing/            # Testing frameworks & guidelines
│   └── *.md           # Test patterns, strategies, guides
│
├── archive/           # Historical docs and handovers
│   └── *.md           # Reference material kept out of the active root

└── templates/          # Copy-paste ready templates
    ├── functional/     # 7 types: agents, skills, rules, commands, mcps, hooks, instructions
    ├── internal/       # 3 types: decisions, project-docs, learning
    └── REFERENCE FILES # Navigation for templates
```

## What Goes Where

- **Why a decision was made** - `adr/`
- **How components connect** - `architecture/`
- **Agent/Skill contracts** - `architecture/`
- **Skill definitions** - `skills/`
- **Test patterns** - `testing/`
- **A template to copy** - `templates/`
- **Historical reference** - `archive/`
- **How to structure docs** - `DOCUMENTATION-UNIFIED-SCHEMA.md`
- **Everything at a glance** - `INDEX.md`

## File Count Summary

```text
Root:         5 files (navigation + standards reference)
adr/:         6 files
architecture/: 7 files
skills/:      8 files
testing/:     5 files
archive/:     2 files
templates/:   49 files

Total:       81 markdown files across 6 domains
```

## Philosophy

- **Factual**: Only documented decisions and specifications
- **Current**: Describes what IS, not what WAS
- **Useful**: Every file answers a real question
- **Organized**: By domain, not by process or phase
- **No Noise**: No reports, handovers, or process meta-docs

**This structure is production-ready and maintained daily.**

## Validation Baseline

- [`../docs-manifest.json`](../docs-manifest.json) defines the current structural contract for the repository.
- [`../docs-changelog.jsonl`](../docs-changelog.jsonl) records documentation changes for future context injection.

## Current Inventory

- Root docs: 5 files
- `adr/`: 6 files
- `architecture/`: 7 files
- `skills/`: 8 files
- `testing/`: 5 files
- `archive/`: 2 files
- `templates/`: 49 files
- Total: 81 markdown files
