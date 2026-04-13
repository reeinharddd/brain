---
type: design-doc
id: infrastructure-baseline-canonical
title: Canonical Infrastructure Stack Baseline
version: 1.0.0
status: active
date_created: 2026-04-12
language: en
category: architecture
related:
  - stack-and-implementation-baseline
  - deployment-profiles-and-infrastructure-baseline
  - identity-policy-security
  - memory-and-knowledge-architecture
  - ai-runtime-and-context-optimization
  - environment-configuration
---

## Overview

This document freezes the concrete infrastructure technologies Brain uses across local personal, self-hosted corporate, and hosted cloud profiles.

It answers four questions for each category:

- what we use
- when we use it
- why we chose it
- what the rest of the app expects from it

The goal is to keep the stack explicit without turning every architecture doc into a vendor catalog.

## Baseline principles

- The daemon owns canonical state.
- Local mode must boot without heavy services.
- Self-hosted and cloud profiles should share the same core contracts.
- External systems are deployment choices, not hidden product behavior.
- If a service is optional in local mode, that must be stated explicitly.

## Profile summary

| Category           | Local personal                                            | Self-hosted corporate                      | Hosted cloud                               | Canonical technology                                     |
| ------------------ | --------------------------------------------------------- | ------------------------------------------ | ------------------------------------------ | -------------------------------------------------------- |
| Identity and auth  | Local bootstrap or dev Logto; no mandatory enterprise SSO | Logto or another tested OIDC IdP           | Brain-hosted Logto                         | Logto + OIDC PKCE                                        |
| Primary database   | SQLite or other local-first store                         | PostgreSQL                                 | PostgreSQL                                 | SQLite local, PostgreSQL elsewhere                       |
| Cache and sessions | In-memory first, optional Redis                           | Redis                                      | Redis                                      | Redis                                                    |
| Semantic memory    | Optional or disabled by default                           | Qdrant                                     | Qdrant                                     | Qdrant                                                   |
| Object storage     | Local filesystem                                          | S3-compatible private storage              | S3-compatible private storage              | S3-compatible storage                                    |
| CDN and WAF        | None by default                                           | Optional customer-controlled edge          | Cloudflare or equivalent edge layer        | Cloudflare semantics                                     |
| Observability      | Console logs and health checks                            | Structured logs, metrics, traces, alerting | Structured logs, metrics, traces, alerting | slog + OpenTelemetry + Prometheus + Grafana stack        |
| Backups and DR     | Local export or snapshot backup                           | Automated backups and restore tests        | Automated backups, PITR, restore tests     | Snapshot plus WAL-aware backups                          |
| Secrets            | Env files or OS keychain                                  | Vault or equivalent secret manager         | Cloud secret manager                       | Environment vars locally, managed secret store elsewhere |
| Tenant isolation   | Single-user only                                          | Optional shared tenancy                    | Required shared tenancy                    | Tenant-scoped records with Postgres RLS                  |

## Identity and auth

- Canonical technology: Logto with OIDC authorization code flow and PKCE.
- Local use: bootstrap login or a local/dev Logto instance is fine; the local profile does not require enterprise SSO.
- Self-hosted use: the operator can run Logto or another tested OIDC provider inside their own boundary.
- Hosted cloud use: Brain operates the Logto deployment in the cloud posture.
- Why this choice: Logto is open source, self-hostable, cloud-capable, and fits a standard OIDC flow without custom password storage.
- Why not other options: custom auth would increase risk and maintenance, Auth0 adds SaaS lock-in, and Keycloak is viable but heavier operationally.
- What the app expects: login handoff, session validation, repo linking, and policy context all depend on an identity provider that speaks standard OIDC.

## Primary database

- Canonical technology: SQLite for local personal mode, PostgreSQL for self-hosted and hosted cloud.
- Local use: SQLite keeps the single-user path simple, offline-friendly, and easy to bootstrap.
- Self-hosted/cloud use: PostgreSQL is the canonical transactional store for artifacts, policies, audit trails, workspace state, and profile-scoped metadata.
- Why this choice: SQLite minimizes startup overhead locally; PostgreSQL gives concurrency, JSON support, migrations, row-level security, and a strong backup story for larger deployments.
- Why not other options: MySQL does not give Brain a better fit than Postgres here, and document databases do not match the relational policy and artifact model.
- What the app expects: the daemon treats the database as source of truth for persisted state; Redis and Qdrant do not replace it.

## Cache and sessions

- Canonical technology: Redis, with in-memory fallback in local personal mode.
- Local use: run in-memory by default and keep Redis optional so the machine can boot without extra services.
- Self-hosted/cloud use: Redis handles cache, transient session state, rate limiting, and temporary coordination data.
- Why this choice: Redis is more useful than Memcached here because Brain needs TTLs, counters, transient sets, and flexible ephemeral state.
- Why not other options: DB-backed cache tables are too heavy for ephemeral data, and Memcached is too limited for the current control-plane needs.
- What the app expects: anything in Redis must be disposable; canonical state belongs in PostgreSQL or artifact storage.

## Semantic memory

- Canonical technology: Qdrant.
- Local use: optional or disabled by default, so the local profile stays lightweight.
- Self-hosted/cloud use: Qdrant backs vector recall, similarity clustering, and retrieval workflows.
- Why this choice: Qdrant is open source, self-hostable, and purpose-built for semantic retrieval without pretending to be the system of record.
- Why not other options: Pinecone adds lock-in, and a generic relational store is not the right retrieval engine for similarity-heavy workflows.
- What the app expects: Qdrant stores vectors and retrieval hints only; it must never become canonical artifact or policy storage.

## Object storage

- Canonical technology: filesystem storage locally, S3-compatible object storage for self-hosted and cloud profiles.
- Local use: use the local filesystem for exports, attachments, and developer-friendly persistence.
- Self-hosted/cloud use: use S3-compatible private storage for artifacts, exports, snapshots, and generated assets.
- Why this choice: S3-compatible semantics keep the app portable across environments and make signed URLs, private buckets, and lifecycle policies easy to reason about.
- Why not other options: vendor-specific blob APIs create avoidable lock-in and make migration harder.
- What the app expects: object storage is for binary payloads and exports, not for canonical policy or registry data.

## CDN and WAF

- Canonical technology: none in local personal mode, customer-controlled edge in self-hosted mode, Cloudflare or equivalent edge in hosted cloud.
- Local use: do not require CDN or WAF to start the app.
- Self-hosted use: allow the operator to bring their own edge or use Cloudflare if they want a standard reference.
- Hosted cloud use: Cloudflare is the default because it combines DNS, CDN, WAF, and DDoS protection in one operational layer.
- Why this choice: the app only needs edge protection and caching at the boundary; it should not depend on edge features for correctness.
- Why not other options: forcing a CDN into local mode would violate the lightweight profile, and hard-coding a single cloud-native edge would reduce portability.

## Observability

- Canonical technology: Go slog for structured logs, OpenTelemetry for tracing, Prometheus-compatible metrics, and Grafana-family backends for aggregation in self-hosted or cloud deployments.
- Local use: console logs and health endpoints are enough for the smallest profile, with the same instrumentation contract underneath.
- Self-hosted/cloud use: structured logs, metrics, traces, and alerting become first-class operational requirements.
- Why this choice: OpenTelemetry and Prometheus are open standards, fit the Go toolchain, and keep the instrumentation portable across deployment styles.
- Why not other options: ad hoc logs alone do not give enough visibility, and vendor-specific telemetry would make the control plane harder to move.
- What the app expects: health screens, status endpoints, workflow traces, and audit events all depend on this layer.

## Backups and disaster recovery

- Canonical technology: snapshot or export backups in local mode; PostgreSQL backups with WAL archiving and restore testing in self-hosted/cloud mode; managed PITR when the cloud deployment uses a managed database service.
- Local use: simple exports or snapshots are enough for personal mode.
- Self-hosted/cloud use: backups must be automated, restorable, and tested regularly.
- Why this choice: Brain needs an explicit restore path, not a collection of manual scripts that nobody has rehearsed.
- Why not other options: ad hoc copies and one-off dumps are too fragile for the operational surface Brain is moving toward.
- What the app expects: restore workflows, migration safety, and recovery documentation must exist before production use.

## Secrets and sensitive config

- Canonical technology: environment variables and OS keychain for local mode; Vault or an equivalent secret manager for self-hosted; cloud secret manager for hosted cloud.
- Local use: `.env` files are acceptable for development convenience, but not for shared production secrets.
- Self-hosted/cloud use: secrets must live outside the repo and be retrievable through a managed secret store.
- Why this choice: it keeps secrets out of source control and supports rotation, audit, and least privilege.
- Why not other options: checked-in secrets and app-owned encrypted blobs create avoidable operational and security risk.
- What the app expects: API keys, signing keys, database credentials, and service tokens must all be sourced from the environment or a managed secret backend.

## Tenant isolation

- Canonical technology: tenant-scoped records with PostgreSQL row-level security as the default shared-deployment pattern.
- Local use: no tenant isolation is needed because the profile is single-user.
- Self-hosted/cloud use: shared deployments should enforce tenant boundaries at the database layer and through policy metadata.
- Why this choice: it lines up with Brain's hierarchy model and keeps the shared cloud posture manageable without forcing a separate codebase.
- Why not other options: schema-per-tenant by default adds operational sprawl, and no isolation policy at all would conflict with Brain's governance model.
- What the app expects: policy resolution, audit queries, and sync boundaries must all respect tenant scope.

## How the profiles use the stack

- Local personal: SQLite or other local-first storage, in-memory or optional Redis, optional Qdrant, local filesystem storage, no mandatory CDN/WAF, and lightweight auth.
- Self-hosted corporate: PostgreSQL, Redis, Qdrant, S3-compatible storage, Logto or compatible OIDC, customer-owned edge, and managed secrets.
- Hosted cloud: PostgreSQL, Redis, Qdrant, S3-compatible storage, Brain-hosted Logto, Cloudflare edge, managed secrets, and automated backup/restore.

## Why this stack instead of others

- It keeps the daemon as the source of truth and avoids putting canonical state into the UI or CLI.
- It uses open protocols at the boundaries that matter most: OIDC, S3 semantics, OpenTelemetry, Prometheus.
- It favors self-hostable infrastructure for the control points where Brain needs portability and enterprise trust.
- It keeps local mode simple enough that a single user can run the product without inheriting cloud-only dependencies.

## Related documents

- `docs/architecture/stack-and-implementation-baseline.md`
- `docs/architecture/deployment-profiles-and-infrastructure-baseline.md`
- `docs/architecture/brain-v2-target-architecture.md`
- `docs/architecture/identity-policy-security.md`
- `docs/architecture/memory-and-knowledge-architecture.md`
- `docs/architecture/ai-runtime-and-context-optimization.md`
- `docs/architecture/environment-configuration.md`
