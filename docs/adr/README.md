---
type: documentation
id: readme-adr
title: Architecture Decision Records
version: 2.0.0
status: active
date_created: 2026-04-07
language: en
category: documentation
---

## Architecture Decision Records

This folder contains the active ADR set that still governs Brain.

## Active ADRs

- [ADR-0002](ADR-0002-centralized-brain-cli.md)
- [ADR-0003](ADR-0003-central-daemon-orchestration.md)
- [ADR-0005](ADR-0005-strict-development-and-production-boundary.md)
- [ADR-0007](ADR-0007-unified-capability-control-plane.md)
- [ADR-0008](ADR-0008-unified-artifact-packaging-and-lifecycle.md)
- [ADR-0009](ADR-0009-clean-repository-structure-and-domain-boundaries.md)
- [ADR-0010](ADR-0010-hierarchical-identity-policy-and-security-model.md)
- [ADR-0011](ADR-0011-ai-runtime-and-curator-subsystem.md)

## Rules

- Only active decisions that still govern the product should remain in this folder.
- Historical numbering may contain gaps after cleanup.
- New ADRs must use the `ADR-NNNN-slug.md` format.
- If an ADR is superseded and no longer needed for active understanding, remove it from the canonical set.
