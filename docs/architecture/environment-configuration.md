---
type: architecture
id: environment-configuration
title: Environment Configuration and Production Boundary
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: architecture
---

## Environment Configuration and Production Boundary

This document defines the boundary between development-only tooling and production runtime for the Brain repository.

## Goal

The repository must keep the production artifact minimal, auditable, and safe.

Development-only systems such as documentation RAG tooling, test harnesses, local workflows, and developer support files must remain available to the team, but they must not be packaged or exposed in production.

## Environment model

Brain uses an explicit environment contract:

- `development` for local authoring, debugging, and internal tooling
- `staging` for pre-production validation
- `production` for the shipped runtime

When the environment is not explicit, the runtime must fail closed or use the safest production-safe behavior.

## Classification rules

Every runtime-facing component must declare its visibility:

- `dev-only` for development tooling that must never run in production
- `prod-safe` for components that are allowed in the shipped runtime

Examples of `dev-only` scope include documentation RAG MCPs, test harnesses, inspection tools, and developer workflows.

Examples of `prod-safe` scope include core daemon services, required MCPs used by Brain itself, and runtime configuration that the user needs for normal operation.

## Production packaging rules

The production artifact must be allowlisted, not denylisted.

Production builds should include only:

- runtime binaries
- required runtime configuration
- explicitly approved assets
- operational dependencies that are part of the shipped product

Production builds must exclude:

- documentation used only by developers
- tests and fixtures
- CI/CD workflows
- development scripts and helper tooling
- dev-only MCPs
- editor or IDE support files

Operational scripts and helper workflows that are expected to ship with the product must be implemented as Go executables. Shell or Python scripts are allowed only for local development convenience, transitional compatibility, or documentation examples, and they must remain outside the production artifact.

## Runtime enforcement

The daemon is responsible for enforcing the environment boundary.

At startup it must:

1. Read the active environment.
2. Load only components allowed in that environment.
3. Refuse or skip dev-only entries in production.
4. Emit a clear log entry when a forbidden component is detected.

This keeps the source repository flexible while ensuring the shipped runtime remains minimal.

Development-only helpers may still exist in the repository, but the production path must always be Go-based when the helper is operational rather than purely illustrative.

## Validation approach

The boundary is considered correct only if all of the following are true:

- production startup does not load dev-only tooling
- production artifacts do not contain development-only paths
- the team can still use the same repo for development
- CI can detect accidental leakage before release

## Relationship to deployment profiles

This document defines the dev/staging/production boundary. It works alongside the deployment profile baseline, which defines where Brain runs and which infrastructure pieces are available in each mode.

- Environment = what is allowed to load (`dev-only` vs `prod-safe`)
- Profile = what infrastructure is present or optional (`local personal`, `self-hosted corporate`, `hosted cloud`)

See `docs/architecture/deployment-profiles-and-infrastructure-baseline.md` for the profile matrix.

The concrete technology choices for database, cache, identity, storage, CDN/WAF, observability, backups, and secrets live in `docs/architecture/infrastructure-baseline-canonical.md`.

## Decision summary

The recommended approach is a deny-by-default runtime policy plus allowlisted production packaging.

This is the safest option because it prevents accidental exposure, keeps production clean, and still allows the repository to remain a rich development environment.
