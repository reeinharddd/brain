<!-- markdownlint-disable-file -->

---

artifact_type: commands
version: 2.0.0

---

# Guide: Command Artifacts (DO's & DON'Ts)

## What is a Command?

**Commands** are CLI entry points that delegate work to specialized agents or perform direct operations.

Examples: `/plan`, `/standup`, `/review`

---

## ✅ DO's

1. **Name is a verb phrase** — Start with action: `/plan`, `/debug`, `/standup`, not `/planner` or `/planning`
2. **Document input/output** — Show JSON examples of what it accepts and returns
3. **Delegate to agent** — `/plan` delegates to planner-agent (don't implement logic locally)
4. **Handle errors gracefully** — If daemon unreachable, show: "Daemon offline. Start with: `brain serve`"
5. **Keep output concise** — 5-10 lines max; use JSON for structured data
6. **Test response parsing** — Unit test that CLI parses agent's JSON response

---

## ❌ DON'Ts

1. **Don't hardcode agent names** — Use daemon config (agents.json) as source of truth
2. **Don't block on network** — If daemon response slow (>5s), timeout and retry
3. **Don't implement business logic** — All logic lives in daemon/agents, CLI is thin client
4. **Don't invent new output formats** — Use JSON for structured, plaintext for messages
5. **Don't ignore errors silently** — Log all network/parse errors to `~/.brain/logs/`
6. **Don't forget help text** — `brain [command] --help` should be useful (not empty)

---

## Common Mistakes

| Mistake                    | Why Bad                        | Fix                                    |
| -------------------------- | ------------------------------ | -------------------------------------- |
| **Implement logic in CLI** | Duplicates daemon code         | Move to daemon, CLI calls it           |
| **Hardcode timeouts**      | Breaks on slow networks        | Use configurable timeout               |
| **Ignore parse errors**    | Silent failures confuse users  | Log + display error clearly            |
| **Command name is noun**   | Unclear what it does           | Use verb: `/validate` not `/validator` |
| **No help text**           | Users don't know how to use it | Add `--help` with examples             |

---

## Template Checklist

- [ ] Name is verb-based: `/action`
- [ ] Input documented (accepts what?)
- [ ] Output documented (returns what?)
- [ ] Delegates to agent or daemon operation
- [ ] Error handling: shows helpful message if daemon down
- [ ] Help text: `--help` is useful
- [ ] Tested: response parsing verified

---

## Examples to Reference

- `/plan` — Delegates to planner-agent
- `/review` — Delegates to reviewer-agent
- `/research` — Delegates to researcher-agent

Location: `docs/templates/functional/commands/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
