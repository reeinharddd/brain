---
type: documentation
id: readme-adr
title: Architecture Decision Records
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: documentation
---

## Architecture Decision Records (ADRs)

This folder contains all architecture decision records for the project. ADRs document important decisions with their context, alternatives considered, and rationale.

## What's Here

Architecture decisions about:

- System design and architecture
- Technology choices
- Major API decisions
- Integration patterns
- Infrastructure decisions

## How to Find What You Need

**Browse by status**:

- Active: Decisions currently in effect
- Deprecated: Old decisions (see what replaced them)
- Archived: Historical decisions

**Browse by topic**:

- Use grep: `grep -r "category: " *`
- Check related links in YAML frontmatter
- Look at "Related ADRs" section in each ADR

## Creating a New ADR

1. **Use template**: `templates/adr-template.md`
2. **Number it**: Next number in sequence (e.g., ADR-0006-databasechoice)
3. **Follow structure**: Context → Options → Decision → Rationale → Consequences
4. **Consider trade-offs**: Be honest about costs
5. **Link related**: Update related field when it affects other decisions
6. **Keep it English-only**: titles, sections, and rationale must stay in English

## ADR Index

- [ADR-0001](ADR-0001-combined-prompt-alignment.md) - Combine the Two Brain Ecosystem Prompts Selectively - accepted - 2026-03-31
- [ADR-0002](ADR-0002-centralized-brain-cli.md) - Centralized Brain CLI for Multi-IDE Service Orchestration - accepted - 2026-03-31
- [ADR-0003](ADR-0003-central-daemon-orchestration.md) - Centralization of Orchestration in Daemon - accepted - 2026-04-02
- [ADR-0004](ADR-0004-skill-system-contract.md) - Portable Skill Contract with Surface Adapters - accepted - 2026-04-03
- [ADR-0005](ADR-0005-strict-development-and-production-boundary.md) - Strict Development and Production Boundary - accepted - 2026-04-03
- [ADR-0006](ADR-0006-docs-rag-mcp.md) - Docs-RAG MCP - Semantic Search Over Documentation - approved - 2026-04-03
- [ADR-0007](ADR-0007-unified-capability-control-plane.md) - Unified Capability Control Plane with Hierarchical Context Resolution - approved - 2026-04-04

## Validation Expectations

- ADR filenames use the `ADR-0001-slug.md` format.
- ADR numbers are consecutive and start at `ADR-0001`.
- ADR content follows the template sections and English-only rule.
- Any new ADR must be added through the same sequence and validation path.

## Best Practices

✅ **DO**:

- Document decisions early (before implementation)
- Include rationale and trade-offs
- Link related decisions
- Update status when decision changes
- Keep future ADRs aligned with the automatic validator and manifest

❌ **DON'T**:

- Hide trade-offs or risks
- Ignore related decisions
- Change status without documenting why

---

**Last updated**: 2026-04-03
