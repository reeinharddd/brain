---
type: design-doc
id: llm-and-ide-operating-model
title: LLM and IDE Operating Model
version: 1.0.0
status: active
date_created: 2026-04-07
language: en
category: architecture
---

## Overview

This document explains how LLMs and IDE assistants work from Brain's perspective and how Brain should integrate with them for maximum leverage.

## Purpose

Brain cannot optimize prompts, skills, commands, context, or agent systems well unless it understands the execution model of:

- IDE-based assistants
- CLI-based assistants
- hosted LLM APIs
- self-hosted runtimes

## Core Model

An LLM does not understand the repository by magic. It operates over a bounded input window and whatever tools, instructions, and retrieved context are made available at inference time.

In practice, most modern assistant systems combine:

- system or developer instructions
- user task input
- selected repository context
- tool results
- prior conversation state
- model-specific runtime behavior

## IDE Model

Most IDE assistants add Brain-adjacent value through:

- instruction files
- commands
- tool integrations
- project-aware file reading
- conversation memory inside the IDE session

Limitations:

- IDEs differ in instruction precedence
- many clients inject proprietary wrappers before the user sees anything
- some clients have weak tool orchestration or inconsistent command support

Implication for Brain:

- Brain should not assume identical behavior across IDEs
- Brain should compile target-aware projections instead of shipping one naive prompt everywhere

## CLI Model

CLI assistants tend to be:

- more deterministic
- more tool-capable
- easier to orchestrate
- better suited for scripted and daemon-backed flows

Implication for Brain:

- CLI is the best baseline for reproducible orchestration
- daemon and CLI contracts should define the truth that other surfaces consume

## Hosted Model Behavior

Hosted APIs usually provide:

- strong reasoning quality
- fast model evolution
- optional tool calling
- token-based cost and context constraints

Limitations:

- opaque provider-side wrappers
- cost sensitivity
- privacy concerns
- variable support for advanced workflows

## Self-Hosted Model Behavior

Self-hosted models offer:

- lower marginal cost after setup
- privacy control
- runtime customization
- stronger integration opportunities with Brain

Limitations:

- more variability in instruction-following
- weaker tool use on smaller models
- lower context quality in older or cheaper models

Implication for Brain:

- Brain should not rely on giant flat prompts
- Brain should use structured context bundles and tiered context injection

## Prompting Principles for Brain

- separate baseline rules from task-local context
- inject only what the current task needs
- prefer compact resolved bundles over concatenating many files
- make commands and skills explicit enough for weaker models
- avoid assuming advanced reasoning features exist

## Brain Integration Strategy

Brain should improve both old and new models by providing:

- clean canonical instructions
- resolved artifact bundles
- task-aware context ordering
- reusable commands and skills
- daemon-mediated tool access
- deterministic fallback behavior

## Compatibility Principle

Advanced models may support:

- subagents
- agent teams
- parallel tool orchestration
- deeper planning

Older models may not.

Brain must therefore support two operating tiers:

- baseline compatible mode
- enhanced orchestration mode

Both should consume the same canonical artifacts, but projections may differ by capability tier.
