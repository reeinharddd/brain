---
type: documentation
id: readme-skills
title: Skills Documentation
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: documentation
---

## Skills Documentation

This folder contains definitions for reusable agent capabilities (skills). Each skill documents what it does, how to call it, error handling, and examples.

## What's Here

- Skill definitions with complete input/output contracts
- Error codes and handling procedures
- Usage examples and anti-patterns
- Compatibility information
- Performance and cost metrics
- Automatic validation and orchestration guidance for the Go-based skill system

## How to Find Skills

All skills in this folder follow the naming convention: `skill-[name].md`

**By category**: Check the `category` field in frontmatter

**By function**: Browse the skills list below

**By validation status**: Use the daemon-backed check exposed through `brain skills validate`

## Available Skills

- [Enforcement Rules](ENFORCEMENT-RULES.md) — Validation & integrity rules

## Validation Model

The skill system is validated automatically by the Brain daemon and exposed through the CLI and UI.

- The daemon owns the source of truth for registry-to-filesystem sync
- The CLI surfaces validation results with `brain skills validate`
- The UI polls the daemon for status so users do not need to run manual scripts
- Production-facing automation must remain Go-based; shell helpers are development-only if they exist at all

## Creating a New Skill

1. **Use template**: `../templates/skill-template.md`
2. **Name it**: `skill-[descriptive-name].md`
3. **Define contracts**: Input schema → Output schema
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

✅ **DO**:

- Document all error codes
- Provide examples for every skill feature
- Update compatibility when models change
- Include performance expectations

❌ **DON'T**:

- Leave examples out (few-shot learning essential for LLMs)
- Ignore error cases (LLMs hallucinate without them)
- Skip anti-patterns (prevents misuse)

---

**Last updated**: 2026-04-03
