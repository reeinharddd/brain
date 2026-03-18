# Agent Contracts

This file defines the input and output contract for the nine core brain agents.

## Shared contract

Every delegated agent receives:

- phase: one SDD phase name
- goal: concise outcome statement
- constraints: hard requirements and forbidden changes
- context: relevant files, prior decisions, and stack tags
- expected_artifact: exact deliverable format

Every delegated agent returns:

- summary: what it did
- artifact: the requested phase output
- risks: unresolved concerns or assumptions
- handoff: what the next phase or agent needs

## planner

- Input focus: ambiguous goals, architectural constraints, acceptance criteria
- Output artifact: plan, spec, ADR, or task breakdown

## researcher

- Input focus: unknown technology, pattern, API, or current best practice
- Output artifact: findings with sources, recommendation, and tradeoffs

## designer

- Input focus: UX, component hierarchy, states, accessibility, interaction flow
- Output artifact: design spec with states and behaviors

## reviewer

- Input focus: diff, feature branch, architecture decision, or implementation
- Output artifact: prioritized findings by severity and overall verdict

## debugger

- Input focus: reproducible failure, symptoms, logs, environment notes
- Output artifact: hypothesis log, root cause, fix strategy, prevention

## refactor

- Input focus: stable behavior that needs structural improvement
- Output artifact: incremental refactor plan with verification guardrails

## documenter

- Input focus: stabilized implementation or accepted decision
- Output artifact: README update, ADR, runbook, or handover document

## guardian

- Input focus: security-sensitive diff, dependency change, secret handling, risky command
- Output artifact: security findings with block/approve verdict

## orchestrator

- Input focus: high-level goal spanning multiple phases or agents
- Output artifact: delegation plan, phase coordination, synthesized handoff
