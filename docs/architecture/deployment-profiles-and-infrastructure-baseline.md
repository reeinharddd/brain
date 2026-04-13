---
type: design-doc
id: deployment-profiles-and-infrastructure-baseline
title: Deployment Profiles and Infrastructure Baseline
version: 1.0.0
status: active
date_created: 2026-04-12
language: en
category: architecture
related:
  - brain-v2-target-architecture
  - stack-and-implementation-baseline
  - identity-policy-security
  - memory-and-knowledge-architecture
  - ai-runtime-and-context-optimization
---

## Overview

Brain must support a single product codebase across three deployment profiles:

- local personal
- self-hosted corporate
- hosted cloud

The goal is to keep the local experience lightweight while preserving the same enterprise-grade capabilities for teams, companies, and hosted SaaS users. Infrastructure changes by profile; core domain behavior does not.

This document records the profile matrix we are actively closing so later work does not accidentally turn Brain into either a heavy local-only tool or a locked-in cloud platform.

Detailed technology choices, rationale, and alternatives live in `docs/architecture/infrastructure-baseline-canonical.md`. This profile matrix stays focused on posture and defaults.

## Principles

- One codebase, multiple deployment profiles.
- Core identity, policy, artifact, and projection semantics stay the same across profiles.
- Local usage must be easy to start and cheap to run.
- Self-hosted and cloud deployments must retain full governance, security, and scale paths.
- Important product data stays in Brain-controlled storage, not in a third-party system of record.
- Extra infrastructure is additive and opt-in, not required for a single-user setup.

## Profile matrix

| Capability              | Local personal                                                              | Self-hosted corporate                                 | Hosted cloud                                       |
| ----------------------- | --------------------------------------------------------------------------- | ----------------------------------------------------- | -------------------------------------------------- |
| Primary goal            | Single user, fast start, low overhead                                       | Team/company control, private infrastructure          | Public SaaS, scale, multi-tenant                   |
| Identity                | Lightweight local bootstrap or local/dev Logto; no mandatory enterprise SSO | Logto self-hosted or compatible corporate IdP         | Logto hosted or Brain-operated Logto in our cloud  |
| Canonical DB            | SQLite or other local-first store                                           | PostgreSQL in customer-owned infra                    | PostgreSQL in Brain-operated cloud account         |
| Cache / ephemeral state | In-memory or optional Redis                                                 | Redis                                                 | Redis                                              |
| Semantic memory         | Optional local Qdrant or disabled by default                                | Qdrant self-hosted or Brain-operated in private infra | Qdrant self-hosted in Brain-operated cloud account |
| Object storage          | Local filesystem or local-compatible object store                           | Private S3-compatible object storage                  | Private S3-compatible object storage               |
| CDN / WAF               | None by default                                                             | Optional; customer-controlled                         | Cloudflare or equivalent edge protection           |
| Observability           | Console logs and basic health checks                                        | Structured logs, metrics, traces, alerting            | Structured logs, metrics, traces, alerting         |
| Backups                 | Local export / snapshot backup                                              | Automated backups and restore tests                   | Automated backups, PITR, restore tests             |
| Multi-tenancy           | Off by default                                                              | Optional, if the operator needs it                    | Required                                           |
| GitHub repo linking     | Optional GitHub App when needed                                             | GitHub App                                            | GitHub App                                         |
| Corporate SSO           | Not required                                                                | Supported                                             | Supported                                          |

## Local personal profile

Local mode should behave like a real product, not a toy cloud replica.

Default expectations:

- the daemon can run on one machine
- the user can work with a single account
- most heavy services are optional
- the runtime should start cleanly even if CDN, Redis, Qdrant, or enterprise auth are absent
- the local profile should still support the same core artifact and policy model as larger deployments

What stays optional:

- CDN and WAF
- Redis cluster
- Qdrant service
- enterprise SSO
- directory sync
- multi-region failover
- billing
- high-retention audit systems

## Self-hosted corporate profile

Self-hosted mode is for organizations that want control over infrastructure while keeping modern identity and governance.

Default expectations:

- the operator owns the database and object storage
- the operator can connect corporate identity providers
- the same governance and audit model used in cloud still applies
- the deployment can run on Docker Compose for small teams or Kubernetes for larger environments
- private networking and data isolation are preserved

Typical components:

- Logto self-hosted
- PostgreSQL
- Redis
- Qdrant
- private object storage
- structured logs and metrics
- backups and restore procedures

## Hosted cloud profile

Hosted cloud is the public SaaS posture.

Default expectations:

- Brain operates the environment
- tenants remain isolated
- scale and availability are managed explicitly
- edge protection and observability are enabled by default
- the same identity and policy semantics apply, but the operator owns the service lifecycle

Typical components:

- Logto in Brain-operated cloud
- managed PostgreSQL in Brain-operated cloud account
- Redis
- Qdrant in Brain-operated cloud account
- private object storage
- Cloudflare or equivalent edge layer
- metrics, tracing, alerting, and backup automation

## Invariants across all profiles

The following must remain true regardless of deployment profile:

- the daemon remains the source of truth for runtime state
- policy and authorization stay inside Brain
- GitHub repository linking uses a GitHub App model
- canonical data and metadata stay in Brain-controlled storage
- semantic memory is a retrieval layer, not the source of truth
- infrastructure differences are configuration-driven, not product forks

## Guardrails

To keep local mode lightweight without losing enterprise capability:

- do not require CDN/WAF in local mode
- do not require Redis or Qdrant in local mode unless the user enables them
- do not turn a BaaS into the canonical source of truth for core product data
- do not split the repository into separate local and enterprise codebases
- do not hide profile-specific behavior behind undocumented defaults
- do not treat cloud-only services as mandatory for personal use

## Relationship to other docs

- `docs/architecture/brain-v2-target-architecture.md` defines the conceptual architecture.
- `docs/architecture/stack-and-implementation-baseline.md` defines the baseline technology choices.
- `docs/architecture/identity-policy-security.md` defines hierarchy, trust, and policy.
- `docs/architecture/memory-and-knowledge-architecture.md` defines retrieval and memory responsibilities.
- `docs/architecture/ai-runtime-and-context-optimization.md` defines runtime and context routing.

## Success criteria

- a single user can run Brain locally without dragging enterprise infrastructure along
- a self-hosted company can enable security, storage, and governance without changing the product model
- the hosted cloud path can scale without a separate architecture branch
- infrastructure defaults are obvious, documented, and profile-specific

## Related documents

- `docs/architecture/infrastructure-baseline-canonical.md`
- `docs/architecture/stack-and-implementation-baseline.md`
- `docs/architecture/environment-configuration.md`
