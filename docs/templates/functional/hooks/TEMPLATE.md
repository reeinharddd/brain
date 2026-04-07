## <!-- markdownlint-disable-file -->

# HOOK TEMPLATE - Git and system automation

id: "[hook-id]"
name: "[Hook Name]"
version: "1.0.0"
type: "hook"
trigger: "pre-commit|post-commit|pre-push|on-startup"
language: "bash|go|node"
status: "stable"

# BEHAVIOR

behavior:
blocking: true # Blocks operation if fails
repeatable: true # Safe to run multiple times
timeout_seconds: 30
logs_to: "logs/hooks.log"

---

# Hook: [Name]

## What It Does

[Clear description of what this hook does]

## When It Runs

Triggered: `[event]`

## What It Checks

This hook checks for:

- [Check 1]
- [Check 2]
- [Check 3]

## Success Criteria

Hook passes when:

- ✅ [Condition 1]
- ✅ [Condition 2]

Hook fails when:

- ❌ [Condition 1]
- ❌ [Condition 2]

## Error Message

When this hook fails:

```
ERROR: [Friendly error message]
Fix: [What the user should do]
```

## Examples

### Example 1: Passes

```bash
# User does:
git commit -m "feat: add feature"

# Hook checks: ✅ No hardcoded secrets
# Result: Commit accepted
```

### Example 2: Fails

```bash
# User does:
git commit -m "feat: add api key"
# But file contains: API_KEY="secret123"

# Hook checks: ❌ Found hardcoded secret at line 15
# Result: Commit blocked
```

## Configuration

**Location**: `.git/hooks/[trigger]`

**Invocation**: Automatic on git [trigger]

## Idempotency

This hook is idempotent: safe to run multiple times

```bash
# Same input → Same result
$ [hook-script] ✅
$ [hook-script] ✅  (idempotent)
```

## Logging

Logs output to: `logs/hooks.log`

Sample log entry:

```
[2026-04-03T15:30:22] pre-commit: [hook-id]
[2026-04-03T15:30:22] Checking 5 files...
[2026-04-03T15:30:23] ✅ All checks passed
```

---

**Last Updated**: 2026-04-03
