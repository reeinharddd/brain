## <!-- markdownlint-disable-file -->

# INSTRUCTIONS TEMPLATE - IDE-specific guidance files

name: "[Instructions Name]"
version: "1.0.0"
status: "stable"
ide: "[copilot|claude-desktop|cursor|windsurf|cline|aider]"
applyTo: "[glob patterns: *.go, *.ts, src/**]"

# DEPRECATION

deprecated: false
deprecated_in_favor_of: "[If applicable]"
sunset_date: "[If deprecated]"

# RAG

keywords: ["[domain]", "[when-to-use]"]
applies_to_languages: ["[go|typescript|python|...]"]

---

# [IDE] Instructions

**IDE**: [IDE Name]  
**Version**: [Version]  
**Status**: [Stable/Beta/Deprecated]

---

## When These Apply

These instructions are loaded when:

- **File Pattern**: `[*.go]`
- **Workspace**: [Global/Project-specific]
- **Condition**: [When automatically applied]

## What They Do

[Description of guidance provided]

## Key Behaviors

This instruction file establishes:

### 1. [Behavior 1]

[Explanation]

### 2. [Behavior 2]

[Explanation]

### 3. [Behavior 3]

[Explanation]

## Override/Custom

User can override by:

```
[How user can customize]
```

## Related

Links to related:

- [Rule]: [Link]
- [Skill]: [Link]
- [Agent]: [Link]

---

**Last Updated**: 2026-04-03
