---
type: architecture
id: development-methodology
title: Brain Development Methodology and Working Rules
version: 1.0.0
status: active
date_created: 2026-04-04
language: en
category: architecture
keywords:
  - methodology
  - workflow
  - planning
  - tracking
  - quality
rag_priority: high
chunk_strategy: section
---

## Brain Development Methodology and Working Rules

## Purpose

This document defines how Brain work should be planned, tracked, implemented,
reviewed, and released in a professional way.

The goal is to keep the work:

- clear before coding starts
- traceable while it is in progress
- testable before merge
- reversible when it fails
- consistent across daemon, CLI, desktop UI, docs, and MCP work

## Core Principles

1. One source of truth per concern.
   - Decisions live in ADRs.
   - Work items live in GitHub Issues.
   - Progress lives in GitHub Projects.
   - Code changes live in branches and PRs.
   - Validation lives in tests and GitHub Actions.

2. Smallest effective change.
   - Prefer narrow, reviewable diffs.
   - Avoid mixing unrelated work in one PR.

3. No implementation without a plan.
   - If the task is not trivial, define scope, risks, and acceptance criteria first.

4. No merge without proof.
   - A change is not done until tests, review, and validation pass.

5. Keep the runtime and the workflow separate.
   - Production runtime stays clean.
   - Planning and tracking stay in GitHub and docs.

## Work Size Model

Use the smallest workflow that still gives enough control.

### Type 1: Tiny Fix

Use when:

- typo fix
- one-line bug
- a very small doc correction
- a simple rename with no architectural impact

Flow:

1. Inspect the change.
2. Make the minimal edit.
3. Run the most relevant validation.
4. Commit and close.

### Type 2: Normal Task

Use when:

- a feature, bug, refactor, docs update, or test change needs multiple steps
- the task affects more than one file
- the result needs validation or review

Flow:

1. Create or refine the issue.
2. Put it on the project board.
3. Define acceptance criteria.
4. Branch from the issue.
5. Implement in small steps.
6. Validate.
7. Open PR.
8. Review and merge.

### Type 3: Large Change

Use when:

- the work spans multiple domains
- there is an architecture decision to make
- the change affects runtime behavior, security, or packaging
- the work is likely to take more than one focused session

Flow:

1. Explore.
2. Propose options.
3. Write or update the spec or ADR.
4. Break into tasks.
5. Implement incrementally.
6. Verify with tests and checks.
7. Archive the result.

## Standard Workflow

### 1. Capture

Record the work as an issue or discussion.

Include:

- problem statement
- expected outcome
- domain
- dependencies
- priority
- acceptance criteria

### 2. Plan

If the work is not trivial, create a short design note or ADR before coding.

The plan should answer:

- What are we changing?
- Why now?
- What can fail?
- What is out of scope?

### 3. Track

Put the issue on the main GitHub Project.

Use fields:

- domain
- type
- priority
- effort
- target release
- risk
- dependencies

Use statuses such as:

- Backlog
- Ready
- In progress
- Blocked
- Review
- Done

### 4. Branch

Create one topic branch per issue.

Branch naming:

- `feat/<short-topic>`
- `fix/<short-topic>`
- `docs/<short-topic>`
- `test/<short-topic>`
- `refactor/<short-topic>`

### 5. Implement

Work in small, verifiable steps.

Rules:

- keep commits focused
- do not mix unrelated changes
- do not guess on ambiguous behavior
- stop if the spec must change

### 6. Validate

Every meaningful change must have evidence.

At minimum, validate:

- build or compile status
- unit tests for changed behavior
- lint or formatting where applicable
- manual verification if the change affects UX or workflow

### 7. Review

Open a PR early enough to review the shape of the change, not only the final result.

PRs should include:

- what changed
- why it changed
- tests run
- known trade-offs
- follow-up work, if any

### 8. Release or Archive

Once merged:

- update the project status
- close or link dependent issues
- update docs if behavior changed
- archive old notes if the work is complete

## Rule Set by Case

### Features

Use this when adding new behavior.

Required flow:

1. Define user outcome.
2. Confirm data flow and boundaries.
3. Decide whether an ADR is needed.
4. Break into implementation tasks.
5. Implement the smallest usable slice first.
6. Add tests before or with the code.

### Bugs

Use this when behavior is wrong or unstable.

Required flow:

1. Reproduce the bug.
2. Isolate the smallest failing path.
3. Form 1-3 testable hypotheses.
4. Fix the root cause.
5. Add a regression test.

### Refactors

Use this when the code works but is hard to maintain.

Required flow:

1. Preserve behavior.
2. Add characterization tests if needed.
3. Make one structural improvement at a time.
4. Validate after each step.

### Documentation

Use this when the change is mostly docs or workflow.

Required flow:

1. Keep docs factual.
2. Update navigation if discoverability changes.
3. Link to the source of truth.
4. Avoid duplicating content.

### Security-Sensitive Work

Use this when the change touches secrets, auth, packaging, or production boundaries.

Required flow:

1. Classify the risk.
2. Check the boundary rules.
3. Review the impact on production and dev-only surfaces.
4. Validate explicitly.
5. Keep the change minimal and auditable.

## Planning Rules

Planning should stay practical.

Good planning:

- names the outcome clearly
- lists assumptions
- identifies dependencies
- separates must-have from nice-to-have
- includes success criteria

Bad planning:

- vague goals
- hidden dependencies
- no validation plan
- scope that keeps expanding

## Tracking Rules

Use tracking to reduce ambiguity, not to create bureaucracy.

Track:

- ownership
- status
- domain
- priority
- dependencies
- release target

Do not track:

- the full implementation detail
- duplicated notes already present in the PR
- process noise with no decision value

## Code Quality Rules

Every change should satisfy these checks:

- clear naming
- small functions where practical
- explicit errors
- tests for changed behavior
- no dead code
- no hardcoded secrets
- no hidden state changes

### Quality gates

For code changes, the default gate is:

1. format
2. lint
3. unit tests
4. integration checks if relevant
5. review

### Review standard

Review must answer:

- does it do what it says?
- does it break existing behavior?
- is it safe to merge?
- is the scope still clean?

## How to Work in Each Scenario

### Scenario A: Small bug fix

1. Open issue.
2. Branch.
3. Fix.
4. Add regression test.
5. PR.
6. Merge.

### Scenario B: New feature

1. Issue with acceptance criteria.
2. Board placement.
3. Design note or ADR if needed.
4. Branch.
5. Implement in steps.
6. Validate.
7. PR and review.

### Scenario C: Refactor

1. Add characterization tests.
2. Split work into safe steps.
3. Change structure without changing behavior.
4. Validate after each step.

### Scenario D: Docs or workflow change

1. Update the authoritative document.
2. Update navigation.
3. Avoid duplicate explanations.
4. Verify links and consistency.

### Scenario E: Security or production boundary change

1. Treat as high sensitivity.
2. Minimize the blast radius.
3. Review boundary impact.
4. Validate behavior in the correct environment.

## Recommended Tooling Map

- GitHub Issues: work capture
- GitHub Projects: tracking and status
- GitHub PRs: execution and review
- GitHub Actions: automated validation
- CODEOWNERS: review routing
- ADRs: durable decisions
- Docs: operator and contributor reference

## Validation Plan

A methodology is only useful if it can be checked.

Validate the methodology by asking:

- Was the work captured before coding?
- Was the work tracked on the board?
- Was the implementation kept small?
- Were tests added or updated?
- Was the PR reviewable?
- Was the result documented or archived?

## Final Rule

If a task is not yet clear, do not rush to code.

First make the problem concrete, then define the path, then execute with evidence.
