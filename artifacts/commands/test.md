---
name: test
description: Run a full functional validation of the brain repo.
---

# /test — Brain Repo Validation

Use this command to verify that your environment is fully operational, MCPs are connected, and you are following the rules.

## How to invoke

```bash
/test
```

## What this command does

1. **Runs `~/.brain/scripts/validate.sh`**: Checks for file system health, rule accessibility, and identity.
2. **Performs a Rule Adherence Check**: Reads `canonical.md` and explains how the agent will apply "Smallest effective change" in the current context.
3. **Tests Memory**: Attempts to save a test bit of information and retrieve it (if Memory MCP is active).
4. **Verifies Autonomy**: Confirms that the agent can spawn a sub-agent or command if needed.

## Success Criteria

- All `validate.sh` tests pass (marked with ✓).
- Agent can correctly cite a rule from `~/.brain/rules/canonical.md`.
- No broken symlinks in `~/.claude/commands` or `~/.claude/agents`.

---
Run this if you suspect the environment is drifting or after an update.
