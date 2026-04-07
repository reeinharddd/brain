# Skills Compatibility Matrix

## Supported Types

| Type         | Description                                      | Examples                                |
| ------------ | ------------------------------------------------ | --------------------------------------- |
| Skill        | Reusable capability with a `SKILL.md` definition | debugging-methodology, code-refactoring |
| Context Pack | Stack-specific guidance injected when relevant   | go-service, react-ui                    |

## Supported Surfaces

| Surface       | Read | Write   | Validation |
| ------------- | ---- | ------- | ---------- |
| Daemon        | Yes  | Yes     | Yes        |
| CLI           | Yes  | Yes     | Yes        |
| UI            | Yes  | Limited | Yes        |
| Documentation | Yes  | Yes     | Yes        |

## Validation Flow

1. The daemon loads the registry.
2. The daemon checks filesystem sync.
3. The CLI exposes `brain skills validate`.
4. The UI polls the daemon for sync state.
5. Production-facing automation stays in Go.

## Compatibility Rules

- A skill must have a registry entry and a filesystem definition.
- A context pack must have a registry entry and a valid context path.
- Development-only helpers must not be required by production.
- Any operational helper that ships with the product must be implemented in Go.

## What to Avoid

- Shell-based validation as a production dependency.
- Hidden manual steps between the registry and the filesystem.
- Duplicate skill definitions across surfaces.

## Outcome

This matrix keeps skill support predictable across the daemon, CLI, UI, and documentation while preserving the Go-first production boundary.
