# Guardian Architecture

The Guardian follows a git-native Prowler-style split:

- `checks/`: deterministic rules with explicit severity
- `providers/`: execution backends or enrichment layers
- `outputs/`: presentation adapters for CLI or CI consumers
- `run.sh`: thin orchestrator over staged or diff-only changes

## Execution modes

- Local shift-left: run the repository security checks on the staged diff.
- CI PR mode: run the same security checks on the pull request diff range.
- Guardian resolves the repo root from its own location, so it works both in `~/.brain` and in a normal checkout
- Local fallback: when `--staged` has no staged files, Guardian falls back to `HEAD` unless `--no-fallback-head` is passed

## Current checks

- hardcoded secrets
- explicit `any` in TypeScript
- tracked `.env` files in the diff
- non-ASCII characters in added lines

## Output contracts

- `text`: human-readable findings for local hooks
- `json`: machine-readable findings for CI or dashboards
