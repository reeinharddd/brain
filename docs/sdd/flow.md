# SDD Flow

## Goal

This document defines the canonical Spec-Driven Development flow for the brain repo.
It is plain Markdown so every IDE, agent, and hook can consume the same workflow.

## DAG

```text
Explore -> Propose -> Spec -> Design -> Tasks -> Implement -> Verify -> Archive
```

## Phase details

### 1. Explore

- Inputs: user goal, repository state, prior memory, constraints
- Activities: inspect codebase, detect stack, identify assumptions, surface risks
- Outputs:
  - problem summary
  - constraints list
  - detected stack tags
  - open questions or explicit assumptions

### 2. Propose

- Inputs: exploration summary
- Activities: compare options, identify tradeoffs, recommend one direction
- Outputs:
  - chosen approach
  - rejected alternatives
  - decision rationale

### 3. Spec

- Inputs: accepted proposal
- Activities: define boundaries, acceptance criteria, non-goals
- Outputs:
  - spec document or section
  - success criteria
  - failure conditions

### 4. Design

- Inputs: spec
- Activities: map architecture, interfaces, data flow, UX states where relevant
- Outputs:
  - design notes
  - interfaces and responsibilities
  - operational constraints

### 5. Tasks

- Inputs: design
- Activities: split into atomic, verifiable units
- Outputs:
  - task list
  - validation plan
  - execution order

### 6. Implement

- Inputs: task list
- Activities: make the smallest effective change for each task
- Outputs:
  - code or documentation changes
  - local notes for verification

### 7. Verify

- Inputs: implementation
- Activities: run tests, lint, scripts, and targeted manual checks
- Outputs:
  - validation evidence
  - residual risks
  - follow-up items if needed

### 8. Archive

- Inputs: verified result
- Activities: update docs, handover notes, memory, and decision logs
- Outputs:
  - session summary
  - ADR or handover if needed
  - archived learnings
