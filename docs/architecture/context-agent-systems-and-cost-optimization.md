---
type: design-doc
id: context-agent-systems-and-cost-optimization
title: Context, Agent Systems, and Cost Optimization
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document defines the future-facing model for context handling, subagents, agent teams, pooling, and cost-efficient orchestration in Brain.

## Purpose

As Brain grows, the system must maximize useful context while minimizing duplication, latency, and cost.

## Core Principle

Context is a scarce resource. Brain should treat context as something to resolve, compress, and layer rather than simply accumulate.

## Context Layers

Brain context should be assembled in this order:

1. hard policy and safety baseline
2. shared organizational or user baseline
3. workspace and project context
4. task-local artifacts
5. ephemeral runtime observations

## Resolution Rules

- broader reusable guidance should live at broader scope
- project-specific artifacts should contain only what is truly local
- duplicate guidance should be promoted or consolidated
- the final bundle should be explainable and size-aware

## Agent System Concepts

### Single Agent

One agent handles the full task with tool and context support.

### Subagents

A parent agent delegates bounded work to smaller or specialized agents.

Use when:

- work can be split safely
- context can be narrowed
- the delegated result is independent enough to merge later

### Agent Teams

Multiple specialists coordinate around one task or program of work.

Use when:

- planning, implementation, review, and research need distinct roles
- parallel work reduces latency
- each role has different context needs

### Agent Pooling

Brain may maintain a reusable pool of known agent profiles or runtime workers for recurring task classes.

The pool must remain policy-aware and context-bounded.

## Cost Optimization Principles

- use the smallest capable model for each task
- move repeated guidance out of low-scope artifacts
- avoid sending whole documents when resolved summaries are enough
- use lightweight curator or optimizer jobs for deduplication
- use enhanced orchestration only when the task complexity justifies it

## Old and New Model Support

For older models:

- keep instructions explicit
- reduce indirection
- avoid requiring deep implicit reasoning
- use smaller resolved bundles

For newer models:

- enable parallel or delegated workflows
- exploit stronger reasoning and tool use
- still keep canonical context compact and structured

## Brain Direction

Brain should become a context compiler and orchestration layer that:

- expands capability for weaker models
- constrains waste for stronger models
- preserves a single canonical source of truth across all capability tiers

This direction depends on a formal memory subsystem:

- structured memory for typed continuity
- semantic memory for fuzzy recall and deduplication
- Qdrant-backed retrieval for compact relevance-first recall

Memory improves context resolution, but Brain must still compile a bounded context bundle instead of forwarding raw memory stores into prompts.
