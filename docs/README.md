---
type: documentation-index
id: docs-main-index
title: Brain Repository Documentation
version: 2.1.0
status: active
date_created: 2026-04-03
language: en
category: documentation
---

## Documentation Home

## Development and Production Boundary

Brain keeps development-only tooling separate from the production runtime.

Start here if you need the policy and enforcement model:

- [architecture/environment-configuration.md](architecture/environment-configuration.md) — environment contract, dev-only classification, and production packaging rules

The documentation validation baseline lives alongside these docs:

- [../docs-manifest.json](../docs-manifest.json) — immutable structure contract for the current docs layout
- [../docs-changelog.jsonl](../docs-changelog.jsonl) — append-only log of documentation changes for future RAG context

## Starting Points

- **Just arrived?** Read [QUICK-START.md](QUICK-START.md) for a 5-minute overview.
- **Need a map?** Read [STRUCTURE.md](STRUCTURE.md) for the layout.
- **Looking for something?** Open [INDEX.md](INDEX.md) for master navigation.

## Core Domains

- **adr/** - Architecture Decision Records. Use this for technical decisions and rationale.
- **architecture/** - System design, contracts, and integration.
- **skills/** - LLM agent capabilities and enforcement guidance.
- **testing/** - Testing frameworks and validation practices.
- **templates/** - Copy-paste ready artifact templates.

## ✅ Standards Reference

[**DOCUMENTATION-UNIFIED-SCHEMA.md**](DOCUMENTATION-UNIFIED-SCHEMA.md) — How to structure any document in this repo. Metadata, body sections, and RAG optimization.

[**documentation-enforcement-system.md**](architecture/documentation-enforcement-system.md) — Automated validation system. Explains enforcement rules, how to fix violations, and how the system works.

[**development-methodology.md**](architecture/development-methodology.md) — Planning, tracking, quality gates, and case-based working rules.

### Automated Enforcement

All documentation is automatically validated on every commit:

- **Frontmatter**: Required YAML metadata (type, id, title, version, status, date_created, language, category)
- **Language**: English only - no Spanish or mixed languages allowed
- **Emojis**: None - plain ASCII text only
- **Filenames**: Lowercase with hyphens (kebab-case)
- **Markdown**: Valid syntax, closed code blocks, matched brackets
- **Structure**: Clear headings and sections

**How it works**:

1. Git pre-commit hook validates before every commit
2. If validation fails, commit is BLOCKED with clear error message
3. Run `./scripts/validate-documentation.sh docs/` to validate manually

**To fix violations**, see [documentation-enforcement-system.md](architecture/documentation-enforcement-system.md).

---

## Index of Key Documents

### Decisions

- Architecture decisions with context, options, and rationale.
- Each ADR documents why something was chosen.

### Architecture

- Daemon orchestration design.
- CLI integration pattern.
- GitHub operating model for planning and delivery.
- Agent input/output contracts.
- Skill system contract and architecture.
- Capability control plane target-state roadmap.
- Environment configuration and production boundary.

### Skills

- Skill definitions and contracts.
- Enforcement rules and validation strategies.
- Compatibility matrices.
- Quick reference guides.

### Testing

- Unit test patterns.
- Integration test frameworks.
- UI testing guidelines.
- Test coverage expectations.

### Templates

- Agent templates.
- Skill templates.
- Rule templates.
- Decision templates.
- More template material as needed.

## Quality Principles

This documentation follows these rules:

1. Factual: no opinions, only documented decisions and specifications.
2. Current: describes what is, not what was.
3. Useful: every document answers a real need.
4. Organized: by domain, not by process or phase.
5. Linked: cross-references connect related docs.

## How to Contribute

When adding documentation:

1. Choose the right domain: adr/, architecture/, skills/, testing/, or templates/.
2. Follow [DOCUMENTATION-UNIFIED-SCHEMA.md](DOCUMENTATION-UNIFIED-SCHEMA.md) for structure.
3. Link from [INDEX.md](INDEX.md) so it is discoverable.
4. Keep it factual: no process notes, reports, or historical handovers here.

## Find Things Quickly

### By Question

- How do I propose a change? Create an ADR in `adr/`.
- How do I understand the architecture? Read [architecture/](architecture/).
- How do I write a skill? Copy [templates/skill-template.md](templates/skill-template.md).
- How do I add a test? See [testing/](testing/).
- How do I find an example? Browse [examples/](examples/).
- How do I understand the API? Check [specifications/](specifications/).

### By Topic

- Architecture: [architecture/](architecture/), [adr/](adr/), [specifications/](specifications/)
- Development: [skills/](skills/), [testing/](testing/), [examples/](examples/)
- Reference: [templates/](templates/), [archive/](archive/)

## Documentation Standards

All documents in this repository follow a unified standard with:

- YAML frontmatter for metadata.
- Consistent structure for each document type.
- Clear examples before abstract descriptions.
- English only.
- Complete information, including error handling and trade-offs.

See [DOCUMENTATION-UNIFIED-SCHEMA.md](DOCUMENTATION-UNIFIED-SCHEMA.md) for complete standards.

## Using Templates

For every new document:

1. Choose your document type (ADR, Design Doc, Skill, and so on).
2. Copy the template from [templates/](templates/).
3. Follow the structure and fill in content.
4. Validate with the checklist in [DOCUMENTATION-UNIFIED-SCHEMA.md](DOCUMENTATION-UNIFIED-SCHEMA.md).
5. Place it in the appropriate domain folder.

Example: Creating a new ADR

```bash
cp templates/adr-template.md adr/ADR-NNN-my-decision.md
edit adr/ADR-NNN-my-decision.md
```

## Document Types

- ADR - record decisions. Folder: `/adr/`. Example: `ADR-005-daemon-orchestration.md`.
- Design Doc - plan features. Folder: `/architecture/`. Example: `design-api-gateway.md`.
- Skill - define capability. Folder: `/skills/`. Example: `skill-code-analysis.md`.
- Agent - define LLM role. Folder: coming soon. Example: `agent-architect.md`.
- Tool - document a function. Folder: `/specifications/`. Example: `tool-search-code.md`.
- Example - show usage. Folder: `/examples/`. Example: `example-create-adr.md`.
- Test Guide - testing practice. Folder: `/testing/`. Example: `implementation-guide.md`.
- Spec - define contracts. Folder: `/specifications/`. Example: `specification-api.md`.

## Search Tips

Use grep to find things:

```bash
grep -r "authentication" /docs --include="*.md"
grep -r "status: active" /docs --include="*.md"
find /docs/adr -name "ADR-*.md"
grep -r "database" /docs --include="*.md" | head -20
```

## Document Statistics

- Core architecture: 5+ active.
- Architecture decisions: 4+ active.
- Skill definitions: 10+ active.
- Test guides: 3+ active.
- Examples: 10+ growing.
- Specifications: 5+ active.
- Archived documents: 40+ reference.

## Important Rules

Must:

- Use YAML frontmatter on all docs.
- Write in English only.
- Include examples in all specs and guides.
- Document your decisions.

Must not:

- Use shell scripts (.sh) in documentation.
- Embed metadata in prose.
- Leave dead links.
- Duplicate content.

## Key Documents

- [DOCUMENTATION-UNIFIED-SCHEMA.md](DOCUMENTATION-UNIFIED-SCHEMA.md) - standard for all documents.
- [PROJECT-RESTRUCTURING-PLAN.md](PROJECT-RESTRUCTURING-PLAN.md) - how this came together.
- [QUICK-START.md](QUICK-START.md) - 5-minute orientation.
- [STRUCTURE.md](STRUCTURE.md) - visual documentation map.
- [architecture/capability-control-plane-roadmap.md](architecture/capability-control-plane-roadmap.md) - final-state architecture and phased delivery plan.
- [adr/ADR-0007-unified-capability-control-plane.md](adr/ADR-0007-unified-capability-control-plane.md) - official decision for unified capability governance.
- [architecture/github-operating-model.md](architecture/github-operating-model.md) - GitHub board, PR, automation, and branch protection operating model.

## Help and Feedback

- Something unclear? Check the QUICK-START or the relevant domain README.
- Can't find something? Try [searching with grep](#search-tips).
- Want to improve docs? Create an ADR or update guidance.
- Found a bug or broken link? File an issue.

## Quick Links

- [Quick Start](QUICK-START.md) - 5-minute orientation.
- [Visual Map](STRUCTURE.md) - see folder structure.
- [Standards](DOCUMENTATION-UNIFIED-SCHEMA.md) - our rules.
- [Templates](templates/) - copy-paste ready.
- [Archive](archive/) - historical docs.
- [Architecture](architecture/) - system design.
- [ADRs](adr/) - decisions.
- [Skills](skills/) - capabilities.

**Last Updated**: 2026-04-03  
**Total Documents**: 86  
**Coverage**: Complete  
**Status**: Active & Growing
