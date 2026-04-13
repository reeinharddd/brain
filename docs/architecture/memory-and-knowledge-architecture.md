---
type: design-doc
id: memory-and-knowledge-architecture
title: Memory and Knowledge Architecture
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines Brain's memory and knowledge architecture.

Brain does not treat memory as an optional side feature. Memory is a first-class subsystem that improves continuity, context reuse, policy application, and cross-device operation.

## Purpose

Brain must coordinate more than static artifacts. It must also manage evolving knowledge:

- cross-session memory
- reusable summaries
- semantic recall
- deduplicated context
- cloud-synced knowledge state

## Core Principle

Canonical artifacts remain the source of truth. Memory is a supporting knowledge layer that helps Brain retrieve, compress, and reuse the right information at the right scope.

Memory must never silently replace canonical rules, skills, policies, or other governed artifacts.

## Memory Layers

Brain memory operates in multiple layers:

1. structured memory
   - decisions, observations, handovers, entity relationships, typed notes
2. semantic memory
   - embeddings and vector retrieval for fuzzy recall
3. derived memory
   - summaries, promoted guidance, duplication reports, curator output
4. ephemeral runtime memory
   - current session state, temporary observations, live orchestration state

## Storage Model

Brain should support more than one backing store, but the architecture must distinguish their roles.

### Structured Memory Store

Used for:

- typed entities
- relationships
- observations
- session handovers
- explicit classifications

This layer may be backed by MCP memory servers, local stores, or future managed services.

### Semantic Memory Store

Used for:

- vector search
- semantic recall
- similarity clustering
- deduplication assistance
- retrieval for docs and knowledge augmentation

`Qdrant` is the primary semantic memory backend currently recognized by Brain.

Qdrant is not the whole memory system. It is the vector retrieval component inside the larger memory architecture.

## Qdrant Position in Brain

Qdrant must be treated as:

- an optional but first-class semantic memory backend
- a reusable service for memory retrieval, artifact similarity, and RAG-style indexing
- a component that can run locally, in managed cloud, or in organization-hosted deployments

The profile-specific choices for database, cache, object storage, and backups are recorded in the canonical infrastructure baseline. This document stays focused on memory behavior and retrieval rules.

Qdrant must not be treated as:

- the canonical source of truth for artifacts
- the only memory system
- a hidden implementation detail that bypasses policy, scope, or audit

## Memory Responsibilities

Brain memory should support:

- recall of relevant prior work
- cross-device continuity when sync is enabled
- project and user namespacing
- deduplication detection across artifacts
- promotion of reusable guidance from project to broader scope
- support for internal curator workflows
- support for agent teams and subagent handoff efficiency

## Scope and Namespacing

Memory entries must be classified by scope:

- organization
- team
- user
- workspace
- project
- session

Memory retrieval must respect both scope and policy. Broader memory may inform narrower work, but project-specific memory must not leak across boundaries unless policy explicitly allows it.

## Sync Modes

Brain memory supports these modes:

- local
  - all memory remains on-device
- cloud-synced
  - memory state syncs across devices through approved remote backends
- hybrid
  - selected memory classes sync while sensitive classes remain local

Sync mode must be explicit and auditable.

## Retrieval Rules

Memory should be retrieved in layers:

1. policy and safety baseline
2. structured high-signal memory
3. semantic matches from Qdrant or equivalent vector backend
4. derived summaries
5. task-local observations

Retrieval must remain bounded. Brain should not dump raw memory into prompts without filtering, ranking, and compaction.

## Relation to the AI Subsystem

The `ai/` subsystem and the memory subsystem are related but separate.

- `ai/` manages runtimes, models, routing, embeddings, and curator jobs
- memory manages what is stored, recalled, summarized, and synchronized
- Qdrant sits at the boundary as the main semantic retrieval backend

## Security and Governance

Memory must follow the same governance model as artifacts:

- classification by scope
- trust boundaries for imported data
- audit trails for sync and promotion events
- encryption for sensitive state
- policy checks before retrieval or cross-scope reuse

## Implementation Direction

### Phase 1

- Keep current Qdrant integration visible as semantic memory infrastructure.
- Formalize memory contracts and scope metadata.

### Phase 2

- Add explicit memory registry and backend configuration under the future architecture.
- Standardize retrieval and compaction pipelines.

### Phase 3

- Add cross-device sync controls, audit, and promotion workflows.
- Integrate curator scans with semantic similarity results.

## Success Criteria

- Brain can explain where retrieved memory came from.
- Qdrant-backed semantic recall improves retrieval without becoming the source of truth.
- Memory remains scope-aware, policy-aware, and compact.
- Cross-device continuity works without flattening all context into one unsafe shared store.

## Related Documents

- `docs/architecture/brain-v2-target-architecture.md`
- `docs/architecture/infrastructure-baseline-canonical.md`
- `docs/architecture/ai-runtime-and-context-optimization.md`
- `docs/architecture/context-agent-systems-and-cost-optimization.md`
- `docs/architecture/identity-policy-security.md`
