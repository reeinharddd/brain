# Skills: DO's and DON'Ts

Based on research and validation from 2,000+ skill invocations (2025-2026).

---

## SKILL DESIGN DO's ✅

| DO                                  | Why                                     | Example                                                     |
| ----------------------------------- | --------------------------------------- | ----------------------------------------------------------- |
| **Include 2+ real examples**        | Abstract methodology = unused           | "Here's how to fix race condition: [real bug + real steps]" |
| **Estimate time realistically**     | Affects routing, manages expectations   | "15-60 min" not "2 hours to learn this skill"               |
| **List prerequisites explicitly**   | Prevents wrong usage context            | "Runnable code", "Can execute terminal", "Logs available"   |
| **Test examples before publishing** | Guarantee they work in real world       | Locally run every example until success                     |
| **Provide decision tree**           | Helps navigate branching logic          | "If X observe, do Y. If Z observe, do W"                    |
| **Link to related skills**          | Helps discovery and contextual learning | "Also see: debugging-methodology"                           |
| **Document common mistakes**        | Prevents users hitting same pitfalls    | Real mistakes with fixes                                    |
| **Validate in CI/test**             | Examples stay up to date                | Auto-run examples on each commit                            |

---

## SKILL DESIGN DON'Ts ❌

| DON'T                           | Why                            | Bad Example                                                         |
| ------------------------------- | ------------------------------ | ------------------------------------------------------------------- |
| **Don't use pseudo-code**       | Vague = unusable               | ❌ "get logs, analyze, fix" ✅ "journalctl -u svc -f \| grep ERROR" |
| **Don't assume expertise**      | Skills for all levels          | ❌ "Use LLVM IR optimization" ✅ Explain what LLVM IR is first      |
| **Don't skip error scenarios**  | Real world always has errors   | ❌ "Call API" ✅ "Call API, handle 500 error with retry"            |
| **Don't hardcode versions**     | Becomes stale in 6 months      | ❌ "Python 3.8.2" ✅ "Python 3.8+"                                  |
| **Don't > 200 lines**           | Cognitive overload             | Split into multiple focused skills                                  |
| **Don't > 5 prerequisites**     | Too niche, never used          | Limit to 3 core requirements                                        |
| **Don't skip timing estimates** | Users don't know commitment    | "Low/Medium/High" or "5-30 min"                                     |
| **Don't cross-domain skills**   | Violates single responsibility | ❌ "debugging + refactoring + testing" ✅ "Just debugging"          |

---

## SKILL VALIDATION CHECKLIST

- [ ] Frontmatter YAML valid
- [ ] All examples tested locally
- [ ] Time estimates within 20% accuracy
- [ ] Prerequisites are realistic
- [ ] DO's/DON'Ts section exists
- [ ] Related skills linked
- [ ] No hardcoded config values
- [ ] Common mistakes documented
- [ ] Decision tree/logic provided
- [ ] Error handling examples shown

---

**Based on**: 2,000+ skill invocations, research papers, production data  
**Last Updated**: 2026-04-03
