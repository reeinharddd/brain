---
type: design-doc
id: artifact-system-contract
title: Unified Artifact System Contract
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the canonical contract for Brain-managed artifacts.

An artifact is any managed unit of behavior, context, policy, or runtime capability governed by Brain. Artifacts are first-class entities regardless of whether they originate locally, are synced from cloud, are imported from third-party systems, or are installed from package ecosystems.

## Motivation

Today the repository treats domains such as skills, agents, MCP definitions, rules, and commands as separate folder conventions with partially independent lifecycle rules. That makes imports, sync, validation, security review, and explainability harder than necessary.

Brain v2 needs one artifact model so that all domains can share:

- metadata
- trust and security checks
- sync state
- scope and ownership
- lifecycle transitions

## Goals

- Define one shared metadata envelope for all artifact kinds.
- Support manual creation, import, install, link, and sync flows.
- Support complex payloads such as scripts, prompts, assets, and manifests.
- Enable hierarchical resolution and security gating without domain-specific hacks.

## Non-Goals

- Making every artifact kind identical in payload format.
- Removing domain-specific metadata extensions.
- Solving marketplace protocol details in this document.

## High-Level Design

### Common Envelope

Every artifact has a canonical manifest:

```yaml
apiVersion: brain/v1
kind: skill
id: angular-testing
displayName: Angular Testing
scope: project
owner: user:123
visibility: workspace
source:
  type: manual
  uri: local://artifacts/skills/angular-testing
sync:
  mode: hybrid
  cloudEnabled: true
security:
  trust: verified
  reviewStatus: approved
  permissions:
    fs: read
    net: none
    exec: none
content:
  entrypoint: prompt.md
  files:
    - artifact.yml
    - prompt.md
    - scripts/check.sh
```

### Supported Artifact Kinds

- `rule`
- `skill`
- `agent`
- `command`
- `mcp`
- `provider`
- `ai-model`
- `ai-runtime`
- `template`

### Artifact Package Shape

Complex artifacts live as folders:

```text
artifacts/skills/<id>/
├── artifact.yml
├── prompt.md
├── scripts/
├── assets/
└── tests/
```

Simple artifacts may be manifest-only if no extra payload is required.

### Supported Acquisition Flows

- `create`
  - native Brain artifact creation
- `import`
  - normalize an existing local or remote artifact into Brain format
- `install`
  - fetch and materialize from package or registry ecosystem
- `link`
  - reference external artifact without copying the full payload
- `sync`
  - reconcile local and cloud state

### Lifecycle States

- `draft`
- `active`
- `disabled`
- `deprecated`
- `blocked`
- `archived`

## API / Interface

### Data Structures

```typescript
interface ArtifactRef {
  kind: string;
  id: string;
  version?: string;
}

interface ArtifactSource {
  type: "manual" | "local" | "cloud" | "git" | "registry" | "npm" | "oci";
  uri: string;
}

interface ArtifactEnvelope {
  apiVersion: "brain/v1";
  kind: string;
  id: string;
  scope: string;
  owner: string;
  visibility: string;
  source: ArtifactSource;
}
```

### Key Functions

```go
func CreateArtifact(input CreateArtifactInput) (*ArtifactRecord, error)
func ImportArtifact(input ImportArtifactInput) (*ArtifactRecord, error)
func InstallArtifact(input InstallArtifactInput) (*ArtifactRecord, error)
func SyncArtifact(input SyncArtifactInput) (*ArtifactRecord, error)
func ValidateArtifact(input ValidateArtifactInput) error
```

## Implementation Strategy

### Phase 1: Envelope and Schemas

- Define base schema and per-kind extensions.
- Add validation hooks.

### Phase 2: Registry Normalization

- Load all current domains through the common envelope.

### Phase 3: Lifecycle and Trust

- Add state transitions, source metadata, permissions, and trust verification.

### Phase 4: Cross-Device Sync

- Add local/cloud reconciliation and conflict reporting.

## Trade-Offs

- Aspect: Domain-specific registries
- Option A: Keep one parser and lifecycle per domain
- Option B: Use one envelope plus domain extensions
- Chosen: B
- Rationale: The shared envelope removes duplicated logic while preserving kind-specific metadata where needed.

## Risks & Mitigation

- Risk: Common envelope becomes too broad
- Severity: Medium
- Mitigation: Keep the base schema minimal and use extensions per kind.

- Risk: Existing local formats require migration work
- Severity: Medium
- Mitigation: Add compatibility readers before hard enforcement.

## Testing Strategy

- [ ] Schema validation tests
- [ ] Round-trip tests for create/import/install flows
- [ ] Regression tests for backward-compatible reads

## Success Metrics

- Every managed domain can be loaded as an artifact record.
- Import and install flows no longer require ad hoc code paths per domain.
- Security and sync metadata are available for all artifact kinds.

## Deployment Plan

Start with read-path normalization, then add write-path support.

## Monitoring & Observability

Artifact events must be auditable:

- created
- imported
- installed
- linked
- synced
- blocked
- deprecated

## Related Decisions

- `docs/adr/ADR-0008-unified-artifact-packaging-and-lifecycle.md`
- `docs/architecture/brain-v2-target-architecture.md`

---

**Status**: Active
**Reviewed by**: Brain Architecture Team
**Target completion**: 2026-04-18 for contract approval
