id: skill-system-contract-decision
date: 2026-04-03
status: approved
scope: brain:architecture

<!-- markdownlint-disable-file -->

# Decision: Portable Skill System Contract

## Problem

Skill library needed unified definition across Brain surfaces (daemon, CLI, UI, IDEs). Current: inconsistent formats, no central registry, hard to discover.

## Options Considered

**Option A**: Monolithic registry (all skills in one YAML file)

- Pro: Single source of truth, easy to index
- Con: Doesn't scale (hundreds of skills), all-or-nothing updates

**Option B**: Distributed skill folders (each skill is a directory with SKILL.md)

- Pro: Modular, scales to 1000+ skills, versioning per skill
- Con: Harder to discover, requires filesystem scanner

**Option C**: Hybrid (YAML registry + context packs TSV)

- Pro: Best of both: discovery index + modular storage
- Con: Slight complexity (two files to sync)

## Selected Approach

**Option C: Hybrid Registry**

Skills stored as:

1. `skills/{skill-id}/SKILL.md` (executable definition)
2. `skills/registry.yml` (discovery index)
3. `skills/context-packs.tsv` (reusable context references)

Daemon loads all three, indexes by type/scope/keywords, exposes via API.

## Rationale

- **Discovery**: Registry.yml enables fast search without reading filesystem
- **Scalability**: Filesystem organization supports 1000+ skills
- **Modularity**: Each skill is independent directory (git cherry-pick friendly)
- **RAG Integration**: Index enables LLM context injection by keywords
- **Versioning**: Per-skill semver in registry decouples from global versioning

## Consequences

**Positive**:

- Skills discoverable across all surfaces
- LLMs receive relevant context automatically
- Future: Marketplace capable (publish/subscribe model)
- Easy to archive/deprecate skills

**Negative**:

- Must sync registry.yml with filesystem (daemon validates)
- User must understand folder structure
- Requires daemon orchestration (no standalone scripts)

## Validation

| Validation      | Method                   | Status |
| --------------- | ------------------------ | ------ |
| Discovery works | CLI: `brain skills list` | ✅     |
| Registry valid  | Schema check in CI       | ✅     |
| Sync enforced   | Daemon startup check     | ✅     |
| Scaling tested  | 100+ test skills         | ✅     |

## Approval

- Approved by: @tech-lead (2026-04-03)
- Reviewed by: @architect, @security-lead
- Implements: Rule "Master Registry Drives Filesystem"

## Related Decisions

- [Skills Validation Architecture](decision-skills-validation.md)
- [Go-Only Orchestration](decision-go-only.md)

---

**Created**: 2026-04-03  
**Modified**: 2026-04-03  
**Review Date**: 2026-07-03
