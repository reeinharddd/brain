---
type: design-doc
id: skill-artifact-authoring
title: Skill Artifact Authoring and Scope Rules
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines how skills fit into the Brain v2 artifact model.

## Purpose

Skills are no longer treated as an isolated special system. They are one artifact kind inside the unified Brain artifact platform.

The purpose of a skill is to package reusable capability guidance, context, and optional support files in a way that can be:

- created manually
- imported from an external source
- installed from a package or registry
- synchronized across devices and scopes

## Scope Rules

Skills may exist at:

- organization scope
- team scope
- user scope
- workspace scope
- project scope

Skills must not duplicate baseline guidance that belongs at a broader scope. Reusable cross-project principles should be promoted upward and consumed through hierarchical resolution.

## Packaging Rules

Each skill artifact should be stored as a folder with:

- `artifact.yml` for canonical metadata
- primary content such as `SKILL.md` or equivalent entry file
- optional `scripts/`
- optional `assets/`
- optional `tests/`

## Behavioral Rules

- Skills are resolved by daemon policy, not directly by the client surface.
- Skills may declare trust, permissions, and environment visibility.
- Skills may be active locally, cloud-synced, or hybrid.
- Skills may not bypass organization hard policy.

## Authoring Guidance

- Keep skill payload focused.
- Move broad software principles to shared rules or higher-scope artifacts.
- Avoid duplicating framework-neutral guidance in every project.
- Prefer composition through resolution over copy-paste proliferation.
