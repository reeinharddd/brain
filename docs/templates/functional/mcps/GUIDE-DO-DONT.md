<!-- markdownlint-disable-file -->

---

artifact_type: mcps
version: 2.0.0

---

# Guide: MCP Artifacts (DO's & DON'Ts)

## What is an MCP?

**MCPs** (Model Context Protocol) are tool definitions that extend agent capabilities.

Examples: code-execution, web-fetch, file-system-read

---

## ✅ DO's

1. **Define clear input schema** — JSON Schema with descriptions of every field
2. **Document output format** — Show what the tool returns (success + error cases)
3. **Set realistic limits** — Timeout, payload size, rate limits (be explicit)
4. **Handle partial failures** — If 1 of 5 files fails, return partial result + error list
5. **Test with agents** — Verify an actual agent can use your tool successfully
6. **Log tool invocations** — Every call logged to `~/.brain/logs/mcp-usage.log`

---

## ❌ DON'Ts

1. **Don't expose secrets** — No API keys, tokens in MCP outputs
2. **Don't make tools too permissive** — Sandbox/limit what agents can do
3. **Don't ignore timeouts** — Hanging requests block agents indefinitely
4. **Don't return raw errors** — Wrap errors with context agent can understand
5. **Don't duplicate existing tools** — Check `config/mcps.json` before creating new one
6. **Don't skip error cases** — Document BOTH success and error responses

---

## Common Mistakes

| Mistake                    | Why Bad                            | Fix                                                |
| -------------------------- | ---------------------------------- | -------------------------------------------------- |
| **No timeout**             | Agent waits forever if tool hangs  | Always set timeout (30s default)                   |
| **Unclear schema**         | Agent doesn't know how to use tool | Use JSON Schema with descriptions                  |
| **Only show success case** | Agent doesn't handle errors        | Document error responses too                       |
| **Tool is too powerful**   | Agent deletes wrong files          | Apply security limits (read-only when possible)    |
| **No error wrapping**      | Agent sees raw cryptic error       | Wrap with context: "Failed to fetch URL: [reason]" |

---

## Template Checklist

- [ ] Input schema defined (JSON Schema)
- [ ] Output format documented (success + error)
- [ ] Timeout specified (default 30s)
- [ ] Security constraints documented
- [ ] Tested with actual agent invocation
- [ ] Errors are wrapped with context
- [ ] Logging enabled for all invocations

---

## Examples to Reference

- code-execution — Python/JS execution in sandbox
- web-fetch — HTTP client with retry logic
- file-read — Read-only file access

Location: `docs/templates/functional/mcps/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
