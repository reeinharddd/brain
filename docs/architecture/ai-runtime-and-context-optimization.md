---
type: design-doc
id: ai-runtime-and-context-optimization
title: AI Runtime and Context Optimization
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the `ai/` subsystem for Brain v2.

The subsystem covers model registries, local and cloud runtimes, routing policy, and an internal curator process that keeps artifact context compact and reusable.

## Motivation

Brain currently treats providers and model selection as a narrow routing concern. That is not sufficient for a product that must:

- run local open-source models
- support hosted APIs
- provide context natively to self-hosted runtimes
- optimize and deduplicate artifact context over time

## Goals

- Introduce `artifacts/ai/` as a first-class domain.
- Support local, cloud, and hybrid AI runtimes.
- Keep Brain context native to self-hosted model flows.
- Add an internal curator model for deduplication and optimization tasks.

## Non-Goals

- Building a full training system.
- Replacing the main task models with a single internal model.
- Solving every provider-specific transport detail now.

## High-Level Design

### AI Domain Contents

The `ai/` domain contains:

- model definitions
- runtime definitions
- routing policies
- embedding backends
- optimizer and curator jobs

The AI subsystem depends on, but does not replace, Brain's memory and knowledge subsystem. In that relationship:

- embeddings are produced and consumed by AI-capable runtimes
- semantic retrieval is backed by Qdrant or another approved vector backend
- structured memory remains a separate concern from model execution

### Runtime Classes

- `hosted-api`
  - OpenAI-compatible and provider-native APIs
- `local-managed`
  - Ollama, vLLM, llama.cpp, LM Studio, local inference servers
- `remote-self-hosted`
  - organization-managed remote runtimes

### Curator Subsystem

Brain includes a smaller internal model or model profile for maintenance tasks:

- duplicate detection across artifacts
- similarity clustering
- context compaction
- promotion of reusable principles from project scope to user or org scope
- stale artifact cleanup proposals

The curator proposes changes. It does not silently rewrite canonical artifacts.

### Native Context Strategy

For self-hosted and managed runtimes, Brain should inject resolved context through a structured bundle rather than by appending every artifact verbatim to prompts.

Context bundle layers:

- organization and security baseline
- user baseline
- workspace and project specifics
- task-local context

### Environment and Safety

AI runtimes must declare:

- capabilities
- privacy posture
- allowed data classes
- tool support
- network expectations

## API / Interface

### Data Structures

```typescript
interface AIRuntime {
  id: string;
  class: "hosted-api" | "local-managed" | "remote-self-hosted";
  endpoint?: string;
  transport: string;
  privacy: "local" | "hybrid" | "cloud";
}

interface AIModelProfile {
  id: string;
  runtimeId: string;
  useCases: string[];
  supportsTools: boolean;
  supportsEmbeddings: boolean;
}
```

### Key Functions

```go
func ResolveModelForTask(input TaskRoutingInput) (*ModelSelection, error)
func BuildContextBundle(input ContextBundleInput) (*ContextBundle, error)
func RunCuratorScan(input CuratorScanInput) (*CuratorReport, error)
```

## Implementation Strategy

### Phase 1: AI Registry

- Add `ai` domain metadata and schemas.
- Separate providers from models and runtimes.

### Phase 2: Context Bundle Compiler

- Replace naive prompt concatenation with structured bundle assembly.

### Phase 3: Curator Reports

- Add dry-run scans for deduplication and promotion suggestions.

### Phase 4: Managed Local Runtime Integrations

- Add runtime adapters and policy-aware routing.

## Trade-Offs

- Aspect: Put AI runtime details in `providers/`
- Option A: Extend existing providers only
- Option B: Create dedicated `ai/` domain
- Chosen: B
- Rationale: Providers, models, runtimes, embeddings, and curator jobs are related but not identical concerns.

## Risks & Mitigation

- Risk: Curator suggestions become noisy
- Severity: Medium
- Mitigation: Start in dry-run mode with explicit review workflows.

- Risk: Native context injection becomes opaque
- Severity: Medium
- Mitigation: Keep bundle explainability output and token accounting.

## Testing Strategy

- [ ] Routing tests for local, hybrid, and cloud policies
- [ ] Context bundle size and ordering tests
- [ ] Curator dry-run regression tests

## Success Metrics

- Self-hosted runtimes receive compact structured context instead of oversized prompt dumps.
- Repeated cross-project guidance can be promoted and reused.
- Local and hosted runtimes can coexist under one policy-aware routing model.

## Deployment Plan

Ship as registry-first and dry-run-first. Runtime execution integrations follow after contracts are stable.

## Monitoring & Observability

Track:

- model selection decisions
- bundle token size
- semantic retrieval backend usage and health
- curator scan results
- deduplication candidates
- promotion suggestions accepted or rejected

## Related Decisions

- `docs/adr/ADR-0011-ai-runtime-and-curator-subsystem.md`
- `docs/architecture/brain-v2-target-architecture.md`

---

**Status**: Active
**Reviewed by**: Brain Architecture Team
**Target completion**: 2026-04-21 for subsystem approval
