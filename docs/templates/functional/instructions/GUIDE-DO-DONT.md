<!-- markdownlint-disable-file -->

---

artifact_type: instructions
version: 2.0.0

---

# Guide: Instruction Artifacts (DO's & DON'Ts)

## What is an Instruction?

**Instructions** are editor/IDE guidance automatically loaded based on `applyTo` patterns.

Examples: Python debugging tips, Angular component patterns, REST API best practices

---

## ✅ DO's

1. **Define applyTo pattern** — Precisely which files trigger this guidance (`**/*.py`, `**/*.component.ts`)
2. **Teach through examples** — Show DO's + DON'Ts with real code, not just prose
3. **Keep focused** — One domain per instruction (Python debugging, not "Python guide")
4. **Link to related tools** — "Use Pylance for type checking" → reference docs/tools
5. **Update regularly** — Check monthly if guidance is still accurate (ESLint rules change)
6. **Provide escape hatch** — How to disable if guidance is wrong for edge case

---

## ❌ DON'Ts

1. **Don't make instruction too broad** — `applyTo: **/*.ts` loads for tests too? Too wide.
2. **Don't assume tools exist** — Instruction references eslint but project doesn't use it?
3. **Don't make it preachy** — Focus on "how to do it right", not "why people are wrong"
4. **Don't duplicate instructions** — If Python debugging shared, reference it don't repeat
5. **Don't forget context** — New dev doesn't know project structure? Add links to onboarding
6. **Don't ignore IDE limitations** — Does VS Code load this? Cursor? Check `applyTo` syntax

---

## Common Mistakes

| Mistake                  | Why Bad                                              | Fix                                       |
| ------------------------ | ---------------------------------------------------- | ----------------------------------------- |
| **applyTo too broad**    | Instruction shows Python tips when editing .go       | Use specific pattern: `**/*.py`           |
| **Tool not installed**   | Instruction says "Use tool X" but it's not installed | Check prerequisites or link to install    |
| **No examples**          | Dev doesn't know what "right" looks like             | Show 2-3 code examples (DO's + DON'Ts)    |
| **Instruction outdated** | "Use require() in Node" but project uses ES modules  | Update quarterly, date-check instructions |
| **Too long**             | Dev skips 50-line instruction                        | Keep to 1-2 screens max (20 lines)        |

---

## Template Checklist

- [ ] `applyTo` pattern is specific (not too broad)
- [ ] DO's section with code examples
- [ ] DON'Ts section with common mistakes
- [ ] Tools are linked/documented
- [ ] Edge cases mentioned (when instruction doesn't apply)
- [ ] Related guidance linked (don't duplicate)
- [ ] Last updated date visible

---

## Examples to Reference

- python-debugging — How to debug Python in VS Code
- angular-components — Angular component patterns
- rest-api-design — REST API guidelines

Location: `docs/templates/functional/instructions/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
