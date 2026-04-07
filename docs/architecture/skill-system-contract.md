# Skill System Contract

This document defines the replicable contract for creating, generating, modifying, and syncing skills in Brain.

The goal is to keep the base portable and broad, while letting the CLI, daemon, UI, and IDEs project their own output formats without becoming separate sources of truth.

## Goals

- Define a portable skill format that works across compatible agents
- Keep the registry small and metadata-only
- Separate reusable context packs from executable skills
- Keep surface-specific behavior out of the portable core
- Let the daemon compile and validate skills once
- Let CLI, UI, and IDEs consume derived artifacts only
- Make updates deterministic, incremental, and auditable

## Non-goals

- Creating a single monolithic "skill master"
- Duplicating skill logic in CLI, daemon, and UI
- Hardcoding IDE-specific paths inside portable skill content
- Storing generated output as source of truth
- Making skills rewrite their own source files directly

## Terms

### Portable skill

A skill folder that can travel across surfaces with minimal change. It has one required entrypoint file, `SKILL.md`, and optional supporting files.

### Context pack

Reusable background knowledge for a stack or domain. In this repo, `skills/contexts/*.md` are context packs, not full skills.

### Registry

The metadata index in `skills/registry.yml`. It tells the system what exists and where it lives. It should not contain full skill bodies.

### Surface adapter

The layer that renders skill data into the shape a surface needs. Examples:

- CLI tables or JSON
- daemon API responses and compiled caches
- UI cards and status views
- IDE-specific instruction bundles

### Generated artifact

Any file or cache created from the portable source by the daemon or an approved sync process. Generated artifacts are derived data, not source.

## Canonical layout

Preferred structure:

```text
skills/
  registry.yml
  dynamic-registry.tsv
  contexts/
    go-service.md
    react-ui.md
    node-typescript.md
  code-refactoring/
    SKILL.md
    references/
    scripts/
    assets/
```

Rules:

- One skill per folder
- One `SKILL.md` per skill folder
- Supporting files stay one level deep from `SKILL.md`
- Keep long references in separate files
- Keep stack guidance in `skills/contexts/`

## Portable skill contract

### Required frontmatter

Use the smallest stable contract that still discovers the skill well:

- `name`: required
- `description`: required

Recommended optional fields:

- `metadata.version`
- `license`
- other portable metadata only if it is not surface-specific

### Required body behavior

`SKILL.md` should answer these questions clearly:

1. What does this skill do?
2. When should it be used?
3. What is the default workflow?
4. What edge cases matter?
5. What support files should be read next?

### Style rules

- Keep the body short and focused
- Prefer direct instructions over long explanations
- Use one-level-deep references only
- Put deterministic operations in scripts when helpful
- Avoid surface-specific instructions in the portable core

### Forbidden in the portable core

These belong in surface adapters, not in the portable skill contract:

- Claude Code invocation controls
- `context: fork`
- `agent`
- `disable-model-invocation`
- `user-invocable`
- `paths`
- `allowed-tools`
- IDE-specific file paths or prompts
- CLI output formatting rules
- daemon API routing logic
- UI presentation rules

## Registry rules

`skills/registry.yml` should act as a discovery index only.

Recommended fields:

- `id`
- `name`
- `path`
- `description`
- `version`
- `tags`
- `type`
- `targets`
- `status`
- `requires`

Registry should not contain:

- the full skill body
- large examples
- surface overrides
- generated summaries that can drift

If a skill needs a large amount of detail, put that detail in the skill folder and let the daemon generate the surface-specific views.

## Surface separation

### CLI

The CLI is a thin textual client.

It may:

- list skills
- search skills
- show skill metadata
- trigger sync

It should not:

- parse raw skill files independently
- duplicate daemon validation logic
- become a second registry implementation

### Daemon

The daemon is the compiler, validator, and sync engine.

It should:

- watch skill sources and registry changes
- validate frontmatter and references
- normalize skill data
- expose API endpoints
- generate derived artifacts atomically
- maintain hashes and dependency edges for incremental sync

It should not:

- hand-author skills as a primary workflow
- rely on the UI to keep state correct

### UI

The UI is for browsing, status, and triggering actions.

It should:

- show skill state
- show sync state
- show search results
- trigger sync when needed

It should not:

- be a source of truth
- encode portable skill logic

### IDEs

IDE targets may have their own compiled instruction files or managed sections.

They should:

- consume generated outputs
- keep user-owned sections separate from managed sections
- avoid hand-editing generated skill projections

## Sync and autoupdate model

The safest model is single-writer, multi-reader.

### Source of truth

The source of truth is the portable skill package plus the registry index.

### Update flow

1. Edit a portable skill or a context pack.
2. The daemon notices the change or receives an explicit sync request.
3. Validate the skill:
   - folder name matches `name`
   - `SKILL.md` frontmatter is valid
   - referenced files exist
   - forbidden surface-specific fields are absent from the portable core
4. Normalize the skill into an internal representation.
5. Rebuild affected indexes and derived outputs.
6. Write generated files atomically.
7. Back up the previous state so rollback stays possible.

### Invalidation rules

- Changing one skill invalidates that skill and its surface projections
- Changing a shared context pack invalidates every skill that references it
- Changing the registry invalidates discovery and routing, not the raw skill body

### Autoupdate rule

Skills may auto-update their derived artifacts through the daemon or approved hooks.

Skills should not directly mutate their own canonical source in place.

## Replicable authoring workflow

Use this process when creating or modifying a skill:

1. Define the task in one sentence.
2. Decide whether it is a portable skill or a reusable context pack.
3. Create or update the portable `SKILL.md`.
4. Add supporting files one level deep if needed.
5. Register only metadata and paths in `skills/registry.yml`.
6. Validate the skill through the daemon or a dedicated validator.
7. Sync generated outputs for the relevant surfaces.
8. Test on one surface first, then confirm the other consumers still work.
9. Promote shared content into `skills/contexts/` only if it is broadly reusable.

## Practical templates

### Portable skill template

```text
---
name: code-review
description: Reviews code changes for correctness, maintainability, and test coverage. Use when reviewing diffs or planning a refactor.
---

# Code Review

## When to use

Use this skill when the task is reviewing code, evaluating a refactor, or checking a change for regressions.

## Workflow

1. Read the relevant files.
2. Check behavior, boundaries, and tests.
3. Summarize risks and required fixes.
4. Suggest the smallest safe change.

## Supporting references

- `references/checklist.md`
- `scripts/validate.sh`
```

### Registry entry template

```yaml
skills:
  code-review:
    name: code-review
    path: skills/code-review/
    description: Reviews code changes for correctness, maintainability, and test coverage
    version: 1.0.0
    tags: [review, quality]
    type: portable
    targets: [cli, daemon, ide]
    status: active
```

## Anti-patterns

- A single "skill master" file that mixes every surface
- Copying the same skill logic into CLI, daemon, UI, and IDE configs
- Storing generated projections in the same place as source
- Putting Claude Code-only frontmatter into a portable skill without a reason
- Making the UI or CLI parse the raw source tree as a second registry

## Validation checklist

- `SKILL.md` exists and is the skill entrypoint
- `name` and `description` are present and meaningful
- folder name matches the skill name
- supporting files stay one level deep
- registry points to a real folder
- no surface-specific fields leak into the portable core
- daemon can validate and compile the skill
- CLI and UI consume derived outputs only
- updates are atomic and reversible

## What belongs where

| Place                 | Belongs here                           | Does not belong here                   |
| --------------------- | -------------------------------------- | -------------------------------------- |
| `skills/`             | portable skill packages, context packs | UI logic, IDE rules, daemon routing    |
| `skills/registry.yml` | metadata, paths, targets, status       | full skill bodies, generated summaries |
| `skills/contexts/`    | reusable stack/domain guidance         | invocation controls or permissions     |
| `daemon`              | validation, sync, compilation, API     | manual authoring of skills             |
| `cli`                 | list, search, info, sync UX            | raw skill parsing                      |
| `ui`                  | browsing, state, triggers              | source of truth                        |

## References

- `docs/adr/ADR-006-skill-system-contract.md`
- `docs/sdd/agent-contracts.md`
- `docs/sdd/flow.md`
- `skills/registry.yml`
- `skills/dynamic-registry.tsv`
