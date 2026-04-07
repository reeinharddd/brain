# Skills Enforcement Strategy

## Summary

The Brain skills system uses automatic enforcement instead of manual shell scripts. The daemon owns validation, the CLI exposes status, and the UI displays sync health so users do not need to run checks by hand.

## Enforcement Layers

### 1. Daemon Validation

- Validates the skills registry on startup.
- Revalidates on a periodic timer.
- Fails fast when the registry and filesystem diverge in a production-sensitive way.

### 2. CLI Validation

- Exposes `brain skills validate` for explicit checks.
- Delegates all actual validation to the daemon.
- Prints clear errors for orphan or missing skills.

### 3. UI Visibility

- Shows sync status in the desktop control panel.
- Polls the daemon at regular intervals.
- Makes drift visible without requiring a terminal.

### 4. Documentation and Rules

- Documents the master-registry-first workflow.
- Explains the safe sequence for adding and removing skills.
- Keeps the development flow understandable and auditable.

## Operational Policy

- Production-facing automation must be implemented in Go.
- Shell or Python helpers are allowed only for development convenience and transitional work.
- If a helper is operational, the Go implementation is the source of truth.

## Safe Workflow

### Add a Skill

1. Register it in the master registry.
2. Create the filesystem directory.
3. Let the daemon validate the sync.
4. Commit only after validation passes.

### Remove a Skill

1. Remove the registry entry first.
2. Remove the filesystem directory second.
3. Validate that no orphans remain.
4. Commit only after validation passes.

## Anti-Patterns

- Creating a directory without registering it.
- Removing a registry entry without deleting the filesystem directory.
- Shipping a production helper as a shell or Python script when a Go executable should exist instead.

## Outcome

This strategy keeps the skills system deterministic, auditable, and safe for production while preserving fast developer feedback.
