# Rules: DO's and DON'Ts

Based on production enforcement data and security research (2025-2026).

---

## RULE DESIGN DO's ✅

| DO                                     | Why                                       | Example                                                                       |
| -------------------------------------- | ----------------------------------------- | ----------------------------------------------------------------------------- |
| **Define enforcement mechanism first** | Rule without enforcement = recommendation | "Pre-commit hook blocks commits" not "We should probably not do this"         |
| **Provide working code examples**      | WRONG vs. RIGHT format                    | ❌ `SECRET="key"` ✅ `SECRET=process.env.API_KEY`                             |
| **Document real impact**               | Prevents "why does this exist?" questions | "Exposed secrets = $6K stolen in 6 hours (AWS CVE-2024-XXXX)"                 |
| **Include common mistakes**            | Prevents users hitting traps              | "❌ Mistake: .env.example with real values ✅ Fix: Example with placeholders" |
| **Specify who enforces it**            | Clear ownership                           | "Pre-commit hook + code review + runtime daemon"                              |
| **Provide migration path**             | Teams don't know how to become compliant  | "Week 1: Run linter report, Week 2: Fix, Week 3: Enforce"                     |
| **Link related rules**                 | Shows dependencies                        | "Works together with: use-environment-variables"                              |
| **Test validation tooling**            | Enforcement must actually work            | Pre-commit hook tested, linter rule verified                                  |

---

## RULE DESIGN DON'Ts ❌

| DON'T                                      | Why                                  | Bad Example                                                                      |
| ------------------------------------------ | ------------------------------------ | -------------------------------------------------------------------------------- |
| **Don't create rules without enforcement** | Unenforceable = ignored              | ❌ "We should use" ✅ "Pre-commit hook blocks"                                   |
| **Don't have vague violation messages**    | Users don't know what to fix         | ❌ "Bad error" ✅ "Line 42: Hardcoded secret detected. Use: process.env.VARNAME" |
| **Don't conflict with other rules**        | Causes confusion                     | ❌ Rule A: Use logging ❌ Rule B: No logs in prod (conflicting)                  |
| **Don't skip examples**                    | Users guess at compliance            | Always show WRONG + RIGHT                                                        |
| **Don't ignore edge cases**                | Real world isn't simple              | "Exception: [scenario], approved by [role]"                                      |
| **Don't skip explaining WHY**              | Rules feel arbitrary without context | Always explain business/security impact                                          |
| **Don't make exceptions easy**             | Rules get hollowed out               | Make both compliance and exceptions painful (rare)                               |
| **Don't version rules vaguely**            | Breaking changes unannounced         | Always: "v1.0 → v2.0: Breaking change in X"                                      |

---

## RULE ENFORCEMENT DO's ✅

| DO                              | Why                                         | Example                                                     |
| ------------------------------- | ------------------------------------------- | ----------------------------------------------------------- |
| **Layer enforcement**           | Different layers catch different violations | Pre-commit (local) + linter (CI) + runtime (daemon)         |
| **Fail loudly**                 | Never silent failures                       | Log FATAL, exit with error code 1                           |
| **Provide clear fix message**   | Users must know what to do                  | "Add this line: ..." not just "violation detected"          |
| **Test enforcement tooling**    | Hook/linter must actually work              | Run pre-commit on test violation, verify it blocks          |
| **Allow documented exceptions** | Real world needs flexibility                | "Exception approved by [role] with reason: [justification]" |
| **Log all violations**          | Audit trail                                 | Even if exception granted, log it                           |
| **Review enforcement annually** | Tech changes, rules might become outdated   | Check: "Does this rule still make sense?"                   |

---

## RULE ENFORCEMENT DON'Ts ❌

| DON'T                                      | Why                                    | Bad Example                                         |
| ------------------------------------------ | -------------------------------------- | --------------------------------------------------- |
| **Don't rely on single enforcement layer** | If one fails, rule becomes ineffective | ❌ Only pre-commit ✅ Pre-commit + linter + runtime |
| **Don't allow silent violations**          | Silent = never fixed                   | ❌ Warning in logs ✅ Blocks merge, notifies user   |
| **Don't make exceptions default**          | Rule loses all power                   | ❌ "Approved by anyone" ✅ "Approved by CISO"       |
| **Don't allow old/stale rules**            | Creates confusion                      | Remove rules older than 12 months without review    |
| **Don't have false positives**             | Tool becomes distrusted                | If > 10% false positives, must fix tool             |

---

## COMMON MISTAKES IN RULE CREATION

| Mistake                   | Example                         | Fix                                                      |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| **Vague definition**      | "Never do bad things"           | "Never hardcode secrets (API keys, passwords, tokens)"   |
| **No enforcement method** | "We should probably..."         | "Pre-commit hook blocks commits with hardcoded API keys" |
| **No examples**           | Only describes, doesn't show    | Add ❌ WRONG and ✅ RIGHT code samples                   |
| **Impossible to follow**  | "No external dependencies"      | Allow core dependencies, prohibit only unnecessary ones  |
| **Too many exceptions**   | "Except in cases A, B, C, D..." | If > 3 exceptions, rule is ineffective                   |
| **No migration path**     | "Starts today" without warning  | 2-week notice, then enforce                              |
| **Silent failures**       | Tool silently fails to detect   | Always log and fail loudly                               |

---

## TESTING A RULE DEFINITION

**Checklist**:

1. [ ] Write test case that violates rule
2. [ ] Run enforcement tool on test case
3. [ ] Verify tool detects violation
4. [ ] Verify error message is clear
5. [ ] Verify fix instructions are actionable
6. [ ] Test with 3+ real-world examples
7. [ ] Verify no false positives on legitimate code

---

## RULE LIFECYCLE

### Creation Phase

- Define purpose and rationale
- Examples (WRONG vs. RIGHT)
- Enforcement method
- Request review

### Rollout Phase

- 1-week notice to teams
- Run linter in report mode (don't fail yet)
- Accept exceptions with justification
- Provide tooling/guidance

### Enforcement Phase

- Pre-commit hook now blocks
- CI linter now fails
- Runtime daemon now checks
- Linter/hook/daemon all agree

### Maintenance Phase

- Monitor for false positives
- Quarterly review for relevance
- Update if tech changes
- Archive if becomes obsolete

---

**Based on**: Production rule enforcement data, security audit results  
**Last Updated**: 2026-04-03  
**Link**: [TEMPLATE.md](./TEMPLATE.md)
