---
type: design-doc
id: research-foundations
title: Research Foundations and Theoretical Inputs
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document records the research and theory areas that inform the Brain architecture.

## Purpose

Brain should not be designed only from local preference. Its architecture needs to be grounded in durable ideas from modern AI tooling, software architecture, security, and context management.

## Foundation Areas

### Capability Control Planes

Brain is influenced by the idea that many surfaces should consume one coordinated control layer rather than inventing local truth.

### Hierarchical Policy Systems

The scope model is influenced by organizational policy layering:

- mandatory baseline
- role or user-specific extension
- workspace and project specialization

### Artifact Normalization

The artifact model is informed by packaging, registry, and lifecycle patterns where source origin, trust, metadata, and payload must travel together.

### Context Engineering

Brain assumes that prompt quality depends more on context selection, layering, and retrieval discipline than on sheer context volume.

### Agentic Systems

Brain treats subagents, agent teams, and orchestration as structured delegation problems, not as unbounded prompt recursion.

### Cost-Aware Inference

Brain treats token usage, model tier selection, and context compaction as first-class architectural concerns.

## External Reference Categories

Research and product validation for Brain should continue to draw from:

- official LLM provider documentation
- IDE and assistant integration docs
- MCP specifications and security guidance
- software architecture and control plane patterns
- context engineering and retrieval best practices
- agentic system design patterns

## Rule

If a future design depends on external behavior that may change, the canonical architecture should record the principle, not just the current vendor detail.
