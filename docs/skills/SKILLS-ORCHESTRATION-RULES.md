# Skills Orchestration Rules

## Core Principle

The skills registry is the source of truth. The filesystem must always match the registry, and the daemon enforces that rule automatically.

## Rules

1. Master registry drives the filesystem.
2. No orphan skill directories.
3. Register first, then create the directory.
4. Remove the registry entry first, then delete the directory.
5. Production-facing automation stays in Go.

## Validation

- The daemon validates sync on startup.
- The daemon rechecks periodically.
- The CLI exposes `brain skills validate`.
- The UI shows sync status in real time.

## What Is Allowed

- Go-based operational helpers.
- Development-only scripts only if they are clearly non-production and transitional.
- Skills that have a matching registry entry and a matching filesystem directory.

## What Is Not Allowed

- Orphan directories.
- Registry entries without a matching directory.
- Production workflows that require shell scripts.
- Hidden manual steps that the daemon cannot validate.

## Safe Workflow

### Add a Skill

1. Register it.
2. Create the folder.
3. Validate with the daemon.
4. Commit only when sync is green.

### Remove a Skill

1. Remove the registry entry.
2. Remove the folder.
3. Validate with the daemon.
4. Commit only when sync is green.

## Outcome

This keeps skills deterministic, auditable, and safe for production while preserving a fast developer workflow.
