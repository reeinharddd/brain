<!-- markdownlint-disable-file -->

---

artifact_type: hooks
version: 2.0.0

---

# Guide: Hook Artifacts (DO's & DON'Ts)

## What is a Hook?

**Hooks** are automation triggered by Git events (commit, push, merge).

Examples: pre-commit validation, pre-push testing, post-merge cleanup

---

## DO's

1. **Hook is fast** — Complete in <1s (pre-commit), <5s (pre-push). Slow hooks block workflow.
2. **Document trigger** — What event launches this hook? (pre-commit, post-merge, etc)
3. **Provide bypass mechanism** — `git commit --no-verify` for emergency bypass (log it)
4. **Show clear error message** — User knows EXACTLY what went wrong and how to fix
5. **Log outcome** — All hook runs logged to `~/.brain/logs/hooks.log`
6. **Idempotent** — Running twice should be safe (no side effects)

---

## DON'Ts

1. **Don't run slow operations** — Pre-commit hooks <1s, or developers ignore them
2. **Don't hide the error** — Don't just fail silently, tell user what went wrong
3. **Don't modify files without asking** — Pre-commit hooks should validate, not auto-fix
4. **Don't assume tools are installed** — Pre-commit hook might run on new checkout (no jq yet?)
5. **Don't make hook optional** — Use pre-commit framework (enforces consistency)
6. **Don't forget stderr output** — All messages to STDERR so user sees them clearly

---

## Common Mistakes

| Mistake                                | Why Bad                                        | Fix                              |
| -------------------------------------- | ---------------------------------------------- | -------------------------------- |
| **Hook runs in 10s**                   | Dev disables it: `git commit --no-verify`      | Use fast tools, cache results    |
| **Error message: "validation failed"** | Dev doesn't know what to fix                   | Show specific issue + solution   |
| **Hook auto-fixes code**               | Dev shocked by changed files                   | Only validate, never auto-fix    |
| **Hook requires npm install**          | Fails on fresh checkout                        | Check tool exists, or install it |
| **No bypass possible**                 | Dev frustrated, forces push bypasses entire CI | Provide `--no-verify` + logging  |

---

## Template Checklist

- [ ] Trigger event defined (pre-commit, pre-push, etc)
- [ ] Execution time <1s (pre-commit) or <5s (pre-push)
- [ ] Error message is clear (specific issue + fix)
- [ ] Bypass available (`--no-verify`) with logging
- [ ] Tools are checked for existence before use
- [ ] No auto-fixes (validate only)
- [ ] Idempotent (safe to run twice)

---

## Examples to Reference

- pre-commit-validation — Checks YAML/JSON schemas
- pre-push-tests — Runs test suite before push

Location: `docs/templates/functional/hooks/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
