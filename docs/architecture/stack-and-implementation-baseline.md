---
type: design-doc
id: stack-and-implementation-baseline
title: Stack and Implementation Baseline
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the intended stack baseline for Brain.

## Product Stack

Brain is built around three product surfaces:

- daemon
- CLI
- desktop UI

## Preferred Technology Baseline

### Core Runtime

- Go for daemon, CLI, sync, validation, policy, and other production-facing operational logic

### Desktop Surface

- TypeScript
- React
- Tauri

### Artifact Storage and Configuration

- YAML and JSON for machine-readable metadata
- Markdown for human-authored artifact payloads and canonical docs

### Observability

- structured logs
- event traces
- health status endpoints
- future metrics and audit streams

### AI Runtime Integration

- provider-backed hosted APIs
- OpenAI-compatible runtimes
- local managed runtimes such as Ollama and future equivalents

## Implementation Rules

- production-facing operational workflows should be implemented in Go
- desktop behavior should remain a client of daemon state, not a second orchestrator
- no surface should own canonical state outside daemon-governed contracts
- machine-readable schemas should exist for all important registries and artifact types

## Long-Term Structure

The implementation target is organized around:

- `apps/`
- `core/`
- `artifacts/`
- `internal/`
- `deploy/`

## Compatibility Goal

Brain should remain usable across older and newer model families by:

- keeping canonical instructions structured and compact
- separating baseline context from task-local context
- preferring resolution and projection over prompt bloat
- degrading gracefully when advanced agent features are unavailable
