---
type: architecture
id: github-operating-model
title: GitHub Operating Model for Brain
version: 1.0.0
status: active
date_created: 2026-04-04
language: en
category: architecture
keywords:
  - github
  - projects
  - issues
  - pull-requests
  - actions
rag_priority: high
chunk_strategy: section
---

## GitHub Operating Model for Brain

## Overview

Brain will use GitHub as the operating system for planning, execution, review, and release.

The goal is to keep the project professional, scalable, and easy to run as a solo developer today, while remaining ready for future collaborators.

This model uses GitHub Projects as the central board, GitHub Issues as the unit of work, pull requests as the merge gate, Actions as automation, and ADRs as the decision record.

## Why GitHub

GitHub gives this project a single place for code, tasks, reviews, automation, and release controls.

Relevant capabilities from GitHub's official docs:

- Projects can be viewed as a table, kanban board, or roadmap.
- Projects can track issues, pull requests, and draft issues.
- Projects support custom fields, iterations, charts, templates, and built-in automations.
- Issues support sub-issues, dependencies, labels, milestones, branches, and slash commands.
- Pull requests support reviews, requested reviewers, required reviews, and merge controls.
- Branch protection can require reviews, status checks, conversation resolution, signed commits, linear history, merge queues, and deployments.
- CODEOWNERS can request reviews automatically for changed paths.
- GitHub Actions can automate checks, labels, workflows, and project updates.

## Core Principle

One source of truth per concern:

- Vision and decisions live in docs and ADRs.
- Work items live in issues.
- Prioritization and flow live in Projects.
- Code changes live in branches and pull requests.
- Validation lives in Actions and branch protection.
- Ownership lives in CODEOWNERS and repository permissions.

## Domain Structure

Use domains to organize responsibility instead of organizing by people.

Recommended domains for this repository:

| Domain | What belongs here |
| --- | --- |
| platform | daemon, internal services, sync engine, orchestration |
| frontend | desktop UI, components, pages, client state |
| cli | commands, flags, user-facing text, thin client behavior |
| ci-cd | GitHub Actions, releases, packaging, automation |
| docs | ADRs, architecture docs, indexes, guides |
| testing | unit tests, integration tests, validation flows |
| security | branch protections, code owners, secrets, policy |
| integrations | MCPs, external APIs, GitHub automation, webhooks |

## Recommended GitHub Surfaces

### Issues

Use issues for atomic work items.

Each issue should describe one outcome and one acceptance path.

Use issue types such as:

- feature
- bug
- refactor
- docs
- test
- chore

Useful issue structures:

- a small bug fix as one issue
- a feature split into multiple sub-issues
- a technical spike as a draft issue first
- a release checklist as a milestone-linked issue set

### Projects

Use one primary GitHub Project for the whole repository.

Recommended views:

- Backlog table
- Kanban board
- Roadmap timeline
- Blocked items view
- By domain view

Recommended fields:

- Domain: single select
- Type: single select
- Priority: single select
- Status: built-in project status
- Effort: number or iteration
- Target release: date or iteration
- Risk: single select
- Depends on: text or issue links

### Pull Requests

Use pull requests for every meaningful change.

Rules:

- one PR per logical change
- PR must reference the issue number
- PR must include a summary and validation notes
- PR should be small enough to review quickly
- PR should be merged only after all checks pass

### Actions

Use GitHub Actions for repetitive validation and automation.

Good uses:

- lint and test runs
- documentation validation
- status badge generation
- auto-labeling
- project field updates
- release packaging
- deployment checks

### Discussions

Use GitHub Discussions for open-ended questions, design debates, and team alignment that are not ready to become issues.

When a discussion becomes scoped work, convert it into an issue.

### Labels and Milestones

Use labels for categorization and milestones for delivery windows.

Recommended labels:

- domain:platform
- domain:frontend
- domain:cli
- domain:ci-cd
- domain:docs
- domain:testing
- domain:security
- domain:integrations
- type:feature
- type:bug
- type:refactor
- type:docs
- type:test
- type:chore
- priority:high
- priority:medium
- priority:low
- blocked
- needs-review

Recommended milestones:

- phase-1-foundation
- phase-2-productivity
- phase-3-integration
- phase-4-hardening

## Standard Working Order

This is the default loop to use every time.

### 1. Capture the work

Create or refine an issue.

If the work is unclear, create a draft issue or discussion first.

Issue should include:

- problem statement
- expected outcome
- domain
- acceptance criteria
- dependencies
- priority

### 2. Place it on the Project

Add the issue to the main project.

Set:

- status = Todo
- domain
- priority
- effort
- target release if known

If it is not ready, move it to a clarification column instead of starting work.

### 3. Split if needed

If the item is larger than one focused implementation block, split it into sub-issues.

Recommended split patterns:

- design
- implementation
- tests
- documentation
- release automation

### 4. Branch for execution

Create a topic branch from the issue.

Branch naming suggestion:

- `feat/<short-topic>`
- `fix/<short-topic>`
- `docs/<short-topic>`
- `test/<short-topic>`
- `chore/<short-topic>`

### 5. Implement in small commits

Keep commits focused and reviewable.

Each commit should represent one meaningful step, not a grab bag.

### 6. Open a pull request

Open a PR as soon as the work is in a reviewable state.

The PR should:

- link the issue
- describe what changed and why
- mention tests run
- call out risks or follow-ups

### 7. Review and validate

Require checks and review before merge.

Use branch protection to enforce:

- required reviews
- required status checks
- conversation resolution
- linear history if desired
- signed commits if required
- merge queue if the branch is busy

### 8. Merge and close

Merge only when the branch is green and reviewable.

After merge:

- move project status to Done
- close or convert dependent issues
- update changelog or release notes if needed

## Branch Protection Model

Protect the default branch and any release branches.

Recommended protections:

- require pull request reviews before merging
- require status checks before merging
- require conversation resolution before merging
- require signed commits if the team can support it
- require linear history
- require merge queue if multiple PRs land often
- require deployments to succeed before merging if you deploy from GitHub
- restrict who can push directly to protected branches
- do not allow bypassing the above settings for admins unless there is a strong reason

This keeps the main branch as a reliable release line.

## CODEOWNERS Strategy

Use CODEOWNERS to route review requests by domain.

Suggested path ownership:

- `daemon/` -> platform owner
- `desktop/` -> frontend owner
- `cli/` -> cli owner
- `docs/` -> docs owner
- `mcp/` -> integrations owner
- `rules/` -> security or governance owner
- `skills/` -> skills owner

For a solo project, this still has value because it creates explicit review intent and makes future team growth easy.

## Automation Model

Use Actions or Project automations for repetitive work.

Recommended automations:

- when issue is labeled `type:bug`, add to Bugs view
- when issue is linked to PR, move project status to In Progress
- when PR is merged, mark issue Done
- when PR receives required approvals and checks pass, add to merge queue or notify reviewer
- when doc files change, run documentation validation
- when release is tagged, run packaging and publish workflow

## Release and Delivery Rhythm

Use a simple weekly cadence.

### Weekly rhythm

- Monday: triage backlog and choose top priorities
- Tuesday to Thursday: execute focused implementation work
- Friday: review, merge, clean backlog, and update roadmap

### Daily rhythm

- pick one main issue
- update project status
- commit in small increments
- open PR early
- finish with validation and notes

## Definition of Ready

An issue is ready only if it has:

- a clear problem statement
- acceptance criteria
- a domain
- a priority
- enough context to start
- dependencies identified

## Definition of Done

Work is done only if:

- code is merged
- tests pass
- review is complete
- project item is updated
- docs are updated if needed
- release notes or changelog are updated if relevant

## Scaling Path

### Solo stage

Use one Project, one backlog, one branch strategy, and one weekly review.

### Small team stage

Add CODEOWNERS, required reviews, and domain ownership.

### Larger team stage

Add multiple Projects by area, stronger branch protections, merge queue, and GitHub Actions automation for field updates and release gates.

## Practical Default Flow for This Repository

Use this as the always-on sequence:

1. capture idea or bug in GitHub Issues
2. classify by domain and type
3. place in GitHub Projects backlog
4. clarify acceptance criteria if needed
5. move to ready
6. create topic branch
7. implement in one focused PR
8. run checks with GitHub Actions
9. require review and resolve comments
10. merge through protected branch rules
11. update roadmap and close item

## If You Want to Work Like a Mature OSS Project

The discipline is not in the tool, it is in the repetition:

- every idea becomes an issue or discussion
- every real task lands in the project
- every meaningful change becomes a PR
- every PR gets reviewed and validated
- every merged item updates the board
- every decision is documented

That is how the project stays organized even when it grows.
