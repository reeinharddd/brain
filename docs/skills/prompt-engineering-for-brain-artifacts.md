---
type: guidance
id: prompt-engineering-for-brain-artifacts
title: Prompt Engineering for Brain Artifacts
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: documentation
---

## Overview

This guide defines how prompts should be written for Brain-managed artifacts such as agents, commands, adapter instructions, and reusable context bundles.

## Core Principle

Brain artifacts should not try to win by being long. They should win by being:

- explicit
- scoped
- composable
- explainable
- compatible with weak and strong models

## Prompt Layers

Prompt material should be split by function:

1. identity
   - what this artifact is
2. responsibility
   - what it owns
3. constraints
   - what it must not do
4. procedure
   - ordered steps when the task needs repeatable execution
5. output contract
   - exact form of the result
6. anti-patterns
   - how the artifact commonly fails

## Design Rules

- Put stable guidance in broader-scope artifacts.
- Put only truly local guidance in project-specific artifacts.
- Prefer short, ordered instructions over dense narrative prose.
- State hard constraints before examples.
- Use examples only when they materially reduce ambiguity.
- Avoid duplicating the same baseline in many artifacts.

## Do

- define the role in one sentence
- define success criteria
- define forbidden behavior
- define when to stop and escalate
- define output shape
- define fallback behavior for degraded environments

## Do Not

- dump entire process manuals into one prompt
- mix identity, policy, and examples randomly
- rely on hidden assumptions
- write prompts that only strong frontier models can interpret
- duplicate global rules inside every artifact

## Compatibility Guidance

For older models:

- make instructions direct
- reduce cross-references
- use explicit step ordering
- keep output schemas simple

For newer models:

- allow delegated work and richer tool usage
- keep constraints explicit anyway
- avoid assuming the model will infer policy correctly without it being stated

## Recommended Structure

```markdown
# [Artifact Name]

## Identity
[one short paragraph]

## Responsibilities
- ...

## Constraints
- ...

## Procedure
1. ...
2. ...

## Output Contract
- ...

## Anti-Patterns
- ...
```

## Brain-Specific Guidance

- Rules belong in rules artifacts, not hidden inside every agent.
- Team behavior belongs in orchestrator or planner-level artifacts, not in every implementer.
- Adapter instructions should express integration-specific constraints only.
- Commands should stay action-oriented and not become generic manuals.

## Quality Check

Before accepting a prompt artifact, verify:

- can it be understood by a weaker model
- does it duplicate a broader rule
- is the output shape testable
- is there any instruction that belongs in policy instead of prompt text

## Related Documents

- `docs/skills/skill-artifact-authoring.md`
- `docs/architecture/context-agent-systems-and-cost-optimization.md`
- `docs/architecture/llm-and-ide-operating-model.md`
