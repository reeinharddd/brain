# Guardian Architecture

The Guardian follows a git-native Prowler-style split:

- `checks/`: deterministic rules with explicit severity
- `providers/`: execution backends or enrichment layers
- `outputs/`: presentation adapters for CLI or CI consumers
- `run.sh`: thin orchestrator over staged or diff-only changes

## Execution modes

- Local shift-left: `bash ~/.brain/guardian/run.sh --staged --threshold critical`
- CI PR mode: `bash guardian/run.sh --diff-range <base...head> --pr-mode`
- Local fallback: when `--staged` has no staged files, Guardian falls back to `HEAD` unless `--no-fallback-head` is passed

## Current checks

- hardcoded secrets
- explicit `any` in TypeScript
- tracked `.env` files in the diff
- non-ASCII characters in added lines

## Output contracts

- `text`: human-readable findings for local hooks
- `json`: machine-readable findings for CI or dashboards
