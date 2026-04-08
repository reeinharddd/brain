---
type: guidance
id: advanced-skill-design-and-composition
title: Advanced Skill Design and Composition
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: documentation
---

## Overview

This guide explains how Brain skills should be designed once the system supports hierarchical scope, deduplication, and context compilation.

## Skill Role

A skill is not a dumping ground for everything known about a stack.

A skill should package only the guidance, examples, scripts, and metadata required for a specific reusable capability.

## What Belongs in a Skill

- stack-specific guidance that is not globally reusable
- concrete procedures for repeated task classes
- tightly related examples
- optional scripts or helpers declared through metadata
- compatibility and cost metadata

## What Does Not Belong in a Skill

- organization-wide policy
- personal work style preferences that apply across stacks
- general software engineering principles
- duplicated security or workflow baselines
- large unrelated knowledge dumps

## Promotion and Consolidation

When the same guidance appears in multiple skills:

- move the shared principle to a broader rule or context layer
- keep only the stack-specific delta inside each skill
- let Brain inject the broader baseline before the skill-specific layer

This is required to reduce token waste and improve maintainability.

## Composition Model

Skills should be composed with:

1. policy baseline
2. broader reusable guidance
3. skill-specific instructions
4. project-local context

The skill should not attempt to impersonate all four layers at once.

## Packaging Guidance

Use a folder-based artifact layout when the skill contains multiple assets:

- `artifact.yml`
- `prompt.md`
- `examples/`
- `scripts/`
- `tests/`

If the skill is small, keep it minimal. Do not invent folder depth without need.

## Versioning

Each skill should carry:

- stable id
- version
- compatibility targets
- scope defaults
- trust and origin metadata

## Validation Checklist

- the skill has one clear responsibility
- it does not duplicate broader guidance
- examples are specific and realistic
- failure modes are documented
- metadata is sufficient for routing and compatibility decisions

## Anti-Patterns

- "framework encyclopedia" skills
- mixing prompts, policies, and personal notes in one file
- hiding required scripts outside declared metadata
- copying the same testing and security advice into every stack skill

## Related Documents

- `docs/skills/skill-artifact-authoring.md`
- `docs/skills/prompt-engineering-for-brain-artifacts.md`
- `docs/architecture/artifact-system-contract.md`
