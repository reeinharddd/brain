# Testing Quick Reference

## What the Testing System Does

- Runs only the tests affected by your change when possible.
- Uses the Brain daemon as the source of truth.
- Exposes test status through the CLI and UI.
- Keeps production-facing automation in Go.

## Core Commands

- `brain test` - run the configured test suites.
- `brain test --watch` - rerun affected tests while files change.
- `brain test --ci-mode` - run in CI-friendly mode.
- `brain test --check` - validate before commit.

## Typical Workflow

1. Change code.
2. Let the daemon detect the affected tests.
3. Run the smallest useful test set.
4. Review the result in the CLI or UI.
5. Commit only after the test run passes.

## Document Map

- `TESTING-IMPLEMENTATION-GUIDE.md` - build and operation details.
- `TESTING-ARCHITECTURE-PROPOSAL.md` - why the system is structured this way.
- `TESTING-INDUSTRY-VALIDATION.md` - supporting evidence and comparisons.
- `UI-TESTING-GUIDE.md` - UI-specific testing guidance.

## Rules of Thumb

- Prefer affected tests over full-suite runs.
- Keep tests deterministic and isolated.
- Avoid shell-based production workflows.
- Use the daemon, CLI, and UI as the supported control surfaces.

## Goal

The testing system should be fast, explicit, and safe enough that developers can trust it without running manual helper scripts.
