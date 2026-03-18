# ADR-0001: Combine the Two Brain Ecosystem Prompts Selectively

## Status

Accepted

## Context

Two implementation prompts describe a similar target architecture for the brain
repo, but they differ in strictness and scope. One prompt emphasizes the
"Gentleman-Brain" architecture and staged batches. The other defines an
enterprise implementation backlog with many explicit files and scripts.

Applying both prompts literally would overload the repo with duplicate concepts,
parallel protocols, and more ceremony than the current Bash-and-Markdown-first
design needs.

## Decision Drivers

- Preserve the existing "rules first" architecture
- Keep the repo portable and Bash/Markdown-centric
- Add compatibility where it improves clarity or onboarding
- Avoid duplicate systems for SDD, Guardian, and memory
- Prefer reversible changes over large rewrites

## Considered Options

### Option 1: Implement both prompts literally

- Pros: maximum prompt compliance
- Cons: duplicates files and protocols, increases maintenance cost, risks losing
  the repo's current simplicity

### Option 2: Keep current repo unchanged

- Pros: lowest risk
- Cons: misses useful structure from both prompts and leaves obvious gaps

### Option 3: Selective alignment with one canonical implementation

- Pros: captures the highest-value ideas while preserving coherence
- Cons: some prompt-specific files become compatibility layers rather than full
  implementations

## Decision

Adopt **Option 3: selective alignment with one canonical implementation**.

## What We Adopt

- Canonical rules remain the single source of truth
- Dynamic stack detection and skill-context injection
- Delegate-first SDD as the default coordination model
- Memory progressive disclosure and project namespaces
- Git-native Guardian with local-first validation
- Persistent runtime validation and context-pack generation
- Compatibility shims where prompt-specific paths improve onboarding

## What We Reject or Defer

- Strict "do not advance batches without human approval" as a hard repo rule
- Mandatory `.specs/` artifacts for every small task
- LLM-dependent Guardian as the default enforcement path
- Large stack-matrix skill expansion before there is real usage pressure
- Replacing existing working scripts only to match prompt naming exactly

## Consequences

### Positive

- The repo aligns better with both prompts without splitting its identity
- Missing conceptual pieces get first-class files where useful
- Existing scripts remain the operational source of truth

### Negative

- Some prompt checklists will still show partial rather than full compliance
- There will be intentional compatibility layers instead of exact replicas

## Implementation Notes

- Add `architect` and `implementer` agent docs
- Add `memory-protocol.md` as an explicit module
- Add Guardian compatibility wrapper and hook installer
- Improve handover and initialization around persistent operation
