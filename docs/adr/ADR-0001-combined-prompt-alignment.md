---
type: adr
id: ADR-0001-combined-prompt-alignment
title: Combine the Two Brain Ecosystem Prompts Selectively
version: 1.0.0
status: accepted
date_created: 2026-03-31
language: en
category: architecture
related: []
keywords:
  - prompt-alignment
  - architecture
  - compatibility
rag_priority: high
chunk_strategy: section
---

## ADR-0001: Combine the Two Brain Ecosystem Prompts Selectively

## Context

Two implementation prompts describe a similar target architecture for the Brain repository, but they differ in strictness and scope. One prompt emphasizes staged batches and a gentler implementation flow. The other defines a broader implementation backlog with many explicit files and scripts.

Applying both prompts literally would duplicate concepts and create more ceremony than the current Bash-and-Markdown-first design needs.

## Decision Drivers

- Preserve the existing rules-first architecture
- Keep the repository portable and Markdown-centric
- Add compatibility where it improves clarity or onboarding
- Avoid duplicate systems for orchestration, memory, and validation
- Prefer reversible changes over large rewrites

## Options Considered

### Option 1: Implement both prompts literally

Pros:

- Maximum prompt compliance

Cons:

- Duplicates files and protocols
- Increases maintenance cost
- Risks losing the repository's current simplicity

### Option 2: Keep the current repository unchanged

Pros:

- Lowest risk

Cons:

- Misses useful structure from both prompts
- Leaves obvious gaps in tooling and validation

### Option 3: Selective alignment with one canonical implementation

Pros:

- Captures the highest-value ideas while preserving coherence
- Keeps the repository understandable
- Supports gradual adoption

Cons:

- Some prompt-specific files become compatibility layers rather than exact replicas

## Decision

Adopt Option 3: selective alignment with one canonical implementation.

## Rationale

Selective alignment keeps the repository coherent while still absorbing the highest-value ideas from both prompts. It avoids duplicate systems, preserves the current operational style, and leaves room for compatibility layers where they are genuinely useful.

## Consequences

### Positive

- The repository aligns better with both prompts without splitting its identity
- Missing conceptual pieces get first-class files where useful
- Existing scripts remain the operational source of truth

### Negative

- Some prompt checklists will still show partial rather than full compliance
- There will be intentional compatibility layers instead of exact replicas

## Implementation Notes

- Add or keep the architect and implementer agent docs
- Add explicit memory protocol documentation where needed
- Keep the existing scripts as the execution layer
- Prefer compatibility wrappers only when they materially reduce friction

## Related ADRs

- ADR-0002: Centralized Brain CLI for Multi-IDE Service Orchestration
- ADR-0003: Centralization of Orchestration in Daemon
- ADR-0004: Portable Skill Contract with Surface Adapters
- ADR-0005: Strict Development and Production Boundary
