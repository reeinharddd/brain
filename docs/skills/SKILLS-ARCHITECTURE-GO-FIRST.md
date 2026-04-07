# Skills Architecture: Go First

## Purpose

The skills system must be validated and operated through the Brain daemon, CLI, and UI. Production-facing automation stays in Go; shell or Python helpers are development-only if they exist at all.

## Core Rule

- The daemon owns skills registry validation and sync checks.
- The CLI exposes skill operations such as `brain skills validate`.
- The UI shows sync status and operational alerts.
- No production workflow depends on a manual shell script.

## Validation Layers

### 1. Daemon

- Loads the skills registry on startup.
- Fails fast when the registry and filesystem diverge in a production-sensitive way.
- Rechecks periodically so drift does not accumulate silently.

### 2. CLI

- Surfaces the current sync state to users.
- Keeps the user-facing workflow thin.
- Delegates all actual validation to the daemon.

### 3. UI

- Polls the daemon for status.
- Makes drift visible without requiring a terminal.
- Helps users notice issues before they become operational problems.

## Operational Policy

- Production helpers that ship with the product must be Go executables.
- Development-only helpers may exist for local iteration, but they must stay outside the shipped runtime.
- The source of truth is the registry plus the daemon's validation logic.

## What This Replaces

- Standalone validation scripts that required manual execution.
- Ad hoc checks that lived outside the main control plane.
- Duplicate skill logic across surfaces.

## Outcome

This approach keeps the skills system portable, observable, and safe for production use while still giving developers fast feedback during local work.
