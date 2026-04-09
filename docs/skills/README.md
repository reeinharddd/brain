---
type: documentation
id: readme-skills
title: Skills Documentation
version: 2.0.0
status: active
date_created: 2026-04-07
language: en
category: documentation
---

## Skills Documentation

This folder contains only the canonical documentation needed to explain how skills fit into Brain v2.

Skills are governed as artifacts, not as a separate documentation-heavy subsystem.

## Current Canonical Files

- [skill-artifact-authoring.md](/home/reeinharrrd/Work/Personal/brain/docs/skills/skill-artifact-authoring.md)
- [prompt-engineering-for-brain-artifacts.md](/home/reeinharrrd/Work/Personal/brain/docs/skills/prompt-engineering-for-brain-artifacts.md)
- [advanced-skill-design-and-composition.md](/home/reeinharrrd/Work/Personal/brain/docs/skills/advanced-skill-design-and-composition.md)

## Rules

- Cross-project reusable guidance should be promoted upward instead of duplicated in every skill.
- Skill behavior is resolved by daemon policy and scope hierarchy.
- Historical skill process notes and UI walkthroughs do not belong in this folder.

The skill system is validated automatically by the Brain daemon and exposed through the CLI and UI.

- The daemon owns the source of truth for registry-to-filesystem sync
- The CLI surfaces validation results with `brain skills validate`
- The UI polls the daemon for status so users do not need to run manual scripts
- Production-facing automation must remain Go-based; shell helpers are development-only if they exist at all

## Creating a New Skill

1. **Use template**: `docs/templates/functional/skills/TEMPLATE.md`
2. **Name it**: `skill-[descriptive-name].md`
3. **Define contracts**: Input schema -> Output schema
4. **Document errors**: Error codes with handling procedures
5. **Provide examples**: Minimum 2-3 usage scenarios
6. **Add anti-patterns**: What NOT to do
7. **Include metadata**: Compatibility, cost, performance
8. **Validate automatically**: Confirm the daemon, CLI, and UI all report the skill as synced before merge

## Skill Template Structure

```text
Type: skill
Input Contract: JSON Schema
Output Contract: JSON Schema
Error Cases: Table with codes
Examples: 2-3 scenarios
Anti-Patterns: Common mistakes
Compatibility: Model/system support
Cost/Performance: Metrics
```

## Best Practices

DO:

- Document all error codes
- Provide examples for every skill feature
- Update compatibility when models change
- Include performance expectations

DON'T:

- Leave examples out (few-shot learning essential for LLMs)
- Ignore error cases (LLMs hallucinate without them)
- Skip anti-patterns (prevents misuse)

---

**Last updated**: 2026-04-07
