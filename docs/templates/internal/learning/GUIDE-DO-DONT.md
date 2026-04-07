<!-- markdownlint-disable-file -->

---

artifact_type: learning
version: 2.0.0
internal: true

---

# Guide: Learning Artifacts (DO's & DON'Ts)

## What is a Learning?

**Learning** documents what we learned from incidents, bugs, or experiments.

Examples: "Race condition in Goroutine sync", "Why we switched to Go", "Database deadlock case study"

---

## ✅ DO's

1. **Start with the symptom** — What went wrong from user perspective?
2. **Show timeline** — When discovered? Time to root cause? Time to fix?
3. **Document root cause** — The ACTUAL cause, not the symptom
4. **Show code examples** — BEFORE (buggy) and AFTER (fixed) code
5. **Extract prevention** — What will prevent this in the future?
6. **Link to related skills/rules** — "See Debugging Methodology skill" or "See Go Concurrency Rule"

---

## ❌ DON'Ts

1. **Don't blame individuals** — "Bob wrote a bug" is not acceptable. Focus on process.
2. **Don't just show the fix** — WHY was it wrong? What was the root cause?
3. **Don't forgot metrics** — How many incidents? How long? What was the cost?
4. **Don't have no follow-up** — Learning document with no prevention = wasted opportunity
5. **Don't duplicate learnings** — Check if similar issue already documented (reference it)
6. **Don't make it theoretical** — Real incident + real fix + real impact

---

## Common Mistakes

| Mistake                         | Why Bad                                   | Fix                                    |
| ------------------------------- | ----------------------------------------- | -------------------------------------- |
| **"A bug was filed and fixed"** | No root cause analysis, will happen again | Show root cause + why it wasn't caught |
| **"Always use X pattern"**      | Vague, not actionable                     | Show specific code pattern + test      |
| **Blame developer**             | Demoralizes team, doesn't fix process     | Focus on system improvements           |
| **No prevention listed**        | Same issue happens again 3 months later   | Add test case + verification           |
| **No timeline**                 | Reader doesn't understand urgency         | Show: discovered → root cause → fixed  |

---

## Template Checklist

- [ ] Incident described (symptom, timeline, impact)
- [ ] Root cause identified + explained
- [ ] Code examples (BEFORE + AFTER)
- [ ] Prevention strategy documented
- [ ] Test added (to catch in future)
- [ ] Links to related skills/rules
- [ ] Metrics/evidence (success rate, time saved)
- [ ] Not blaming individuals (focus on system)

---

## Examples to Reference

- skills-race-condition — Goroutine synchronization bug
- database-deadlock — Query ordering issue
- performance-regression — N+1 query discovery

Location: `docs/templates/internal/learning/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
