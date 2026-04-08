<!-- markdownlint-disable-file -->

---

artifact_type: decisions
version: 2.0.0
internal: true

---

# Guide: Decision Artifacts (DO's & DON'Ts)

## What is a Decision?

**Decisions** document architectural choices with rationale, options considered, and consequences.

Examples: Artifact system architecture, async validation approach, Go-only orchestration

---

## ✅ DO's

1. **Document the problem** — What constraint or problem forced this decision?
2. **Show options considered** — Pro/con for EACH alternative (not just winner)
3. **Explain rationale** — WHY this option over others? (not just "it's faster")
4. **List consequences** — Positive AND negative (no decision is perfect)
5. **Include validation** — How did you verify this decision works? (tests, metrics)
6. **Set review date** — When should this decision be revisited? (3-6 months)

---

## ❌ DON'Ts

1. **Don't hide bad options** — Show all serious alternatives, even rejected ones
2. **Don't lack rationale** — "We chose Go because it's cool" = not acceptable
3. **Don't ignore consequences** — Decision has tradeoffs, document them
4. **Don't make decision in code comments** — Major decisions need their own doc (in docs/)
5. **Don't forget approval** — Who approved this? Document it with date & name
6. **Don't set review date to "never"** — Decisions should be revisited periodically

---

## Common Mistakes

| Mistake                            | Why Bad                                       | Fix                                   |
| ---------------------------------- | --------------------------------------------- | ------------------------------------- |
| **"We chose X because it's best"** | No rationale, not repeatable                  | Show comparison matrix with metrics   |
| **Ignore negative consequences**   | Team discovers problems later (too late)      | List tradeoffs upfront                |
| **No approval documented**         | Later, someone questions the decision         | Add "Approved by: @name (2026-04-03)" |
| **Options section missing**        | Reader doesn't know what alternatives existed | Always show 2-3 serious options       |
| **No review date**                 | Decision becomes stale (world changes)        | Set: "Revisit: 2026-07-03"            |
| **Decision in code comment**       | Hard to find, not indexed, forgotten          | Move to docs/decisions/               |

---

## Template Checklist

- [ ] Problem clearly stated
- [ ] Options considered (minimum 2)
- [ ] Pro/con for each option
- [ ] Selected approach + rationale
- [ ] Consequences (positive + negative)
- [ ] Validation method (tests, metrics)
- [ ] Approval documented (who, when)
- [ ] Review date set (3-6 months from now)

---

## Examples to Reference

- artifact-system-contract — Unified artifact architecture
- go-only-orchestration — Why shell scripts banned

Location: `docs/templates/internal/decisions/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
