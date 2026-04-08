---
type: design-doc
id: model-compatibility-and-capability-layering
title: Model Compatibility and Capability Layering
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

Brain must improve both older and newer models instead of optimizing only for frontier behavior.

This document defines the compatibility model used by Brain when building context, selecting artifacts, and enabling agent workflows.

## Core Principle

Do not assume all models reason equally well, follow long instructions equally well, or tolerate the same context size.

Brain should adapt the system to the model, not force every model through the same artifact payload.

## Capability Tiers

Brain should classify models by effective capability, not just vendor branding.

### Tier 1: constrained models

Traits:

- short stable context windows
- weaker instruction following
- limited tool orchestration
- more brittle formatting compliance

Use:

- direct prompts
- compact bundles
- explicit output contracts
- minimal indirection

### Tier 2: standard capable models

Traits:

- usable reasoning
- reliable structured output with guidance
- moderate tool use
- moderate context tolerance

Use:

- layered prompts
- selected examples
- light delegation
- bounded tool plans

### Tier 3: advanced reasoning models

Traits:

- stronger planning
- better tool orchestration
- better synthesis across artifacts
- improved long-horizon execution

Use:

- subagents where justified
- richer context graphs
- planner and reviewer flows
- curator-assisted context optimization

## Brain Responsibilities by Tier

For constrained models, Brain should:

- compress aggressively
- make goals explicit
- avoid scattered context
- resolve conflicts before injection

For standard capable models, Brain should:

- inject layered context
- preserve output structure
- expose only the relevant tools

For advanced models, Brain should:

- enable orchestration patterns
- maintain strict policy boundaries
- avoid wasting tokens on repeated baseline text

## Artifact Selection Rules

The same artifact should not always be injected in the same way.

Brain may:

- inject full text
- inject compact summary
- inject selected sections
- inject only metadata and keep the payload available through tools

## Subagents and Teams

Older and weaker models should not be forced into complex agent-team behavior.

Brain should reserve:

- multi-agent coordination
- subagent pooling
- richer delegation protocols

for models and runtimes that can support that complexity without losing reliability.

## Compatibility Metadata

Artifacts should eventually declare compatibility fields such as:

- minimum capability tier
- preferred capability tier
- context cost estimate
- tool dependency level
- output strictness

## Success Criteria

- weaker models become more reliable through better context packaging
- stronger models gain scale without uncontrolled context growth
- the same Brain knowledge base remains usable across different runtime classes

## Related Documents

- `docs/architecture/ai-runtime-and-context-optimization.md`
- `docs/architecture/context-agent-systems-and-cost-optimization.md`
- `docs/skills/prompt-engineering-for-brain-artifacts.md`
