---
# RULE TEMPLATE - Copy to create new rule

id: "[rule-id-in-kebab-case]"
version: "[1.0.0]"
category: "[security|architecture|development|deployment]"
severity: "critical|high|medium|low"
status: "stable|deprecated|draft"
date_added: "[YYYY-MM-DD]"
date_updated: "[YYYY-MM-DD]"

# ENFORCEMENT
enforcement:
  pre_commit_hook: true # Blocks commits if violated
  linter: "[guardian|eslint|gofmt]"
  runtime_check: true # Validates at execution
  audit_log: true # Logs violations

# RELATIONSHIPS
supersedes: ["[old-rule-id]"] # If rule was updated/renamed
superseded_by: null # If rule is deprecated
related_rules:
  - "[related-rule-1]"
  - "[related-rule-2]"

# RAG OPTIMIZATION
keywords:
  - "[domain]"
  - "[search-term-1]"
  - "[search-term-2]"
chunk_strategy: "by-section"
---

<!-- markdownlint-disable-file -->

# [Rule Name]

## Definition

Brief, clear definition of what this rule requires:

> [One sentence definition of the rule]

## Why This Rule Exists

### The Problem

[Explain the problem that led to this rule]

### Real Impact

Provide concrete examples of damage if this rule is violated:

- **Business Impact**: [e.g., "$2M lost due to security breach"]
- **Technical Impact**: [e.g., "96-hour incident response required"]
- **Team Impact**: [e.g., "Reduced productivity by 40%"]

### Supporting Evidence

Cite specific examples or research:

- [CVE example]: [Impact]
- [Production incident]: [What happened]
- [Research paper]: [Finding]

---

## How to Comply

### Pattern 1: [First Way]

**Context**: [When to use this approach]

```
[Code example showing WRONG]

[Code example showing RIGHT]
```

**Explanation**: [Why this works]

---

### Pattern 2: [Second Way]

**Context**: [When to use this approach]

```
[Code example WRONG]

[Code example RIGHT]
```

---

### Pattern 3: [Third Way]

[Similar structure]

---

## Validation

### Automated Check

```bash
# Tool: pre-commit hook, linter, or daemon
Rule: [What's being checked]
Violation: [What triggers failure]
Message: "[User-friendly error message]"
Fix: "[How to fix the violation]"
```

**Example Output**:

```
ERROR: Possible hardcoded secret at line 42: DATABASE_PASS
Action: Remove secret, use environment variable instead
```

---

### Code Review

**Reviewer Checklist**:

- [ ] [Check 1]
- [ ] [Check 2]
- [ ] [Check 3]

**Sample Review Comment**:

```
Violation of "[rule-name]": [issue]
Requested Change: [what to change]
Justification: [why this matters]
```

---

### Runtime Check

**Daemon Startup Validation**:

```
IF rule is violated THEN:
  Log FATAL message
  Refuse to start service
  Exit with status code 1
```

**Example**:

```
SECURITY: Hardcoded secrets detected in config.
Fix before running: Use environment variables instead.
Exit code 1.
```

---

## Common Mistakes

### Mistake 1: [Title]

**Anti-Pattern**:

```
[Wrong way]
```

**Why It's Wrong**: [Explanation]

**Correct Way**:

```
[Right way]
```

---

### Mistake 2: [Title]

[Similar structure]

---

### Mistake 3: [Title]

[Similar structure]

---

## DO's ✅

When following this rule,ALWAYS:

| DO         | Why      | Example   |
| ---------- | -------- | --------- |
| **[DO 1]** | [Reason] | [Example] |
| **[DO 2]** | [Reason] | [Example] |
| **[DO 3]** | [Reason] | [Example] |

---

## DON'Ts ❌

When following this rule, NEVER:

| DON'T         | Why      | Example       |
| ------------- | -------- | ------------- |
| **[DON'T 1]** | [Reason] | [Bad example] |
| **[DON'T 2]** | [Reason] | [Bad example] |
| **[DON'T 3]** | [Reason] | [Bad example] |

---

## Exceptions & Edge Cases

**Are there any valid exceptions to this rule?**

> [Answer: Yes/No]

If YES:

- **Exception 1**: [Description]
  - Condition: [When it applies]
  - Justification: [Why it's allowed]
  - Who approves: [Role/person]

---

## Testing & Verification

### How to Verify Compliance

Test case for automated check:

```
Input: [Violating code/config]
Expected Output: [Tool should reject]
Actual Output: [Verify tool rejects it]
```

### Local Validation

```bash
# Before committing: validate locally
$ rules/validate-[rule-id].sh .

# Should output:
# ✅ All files compliant
# or
# ❌ N violations found
```

---

## Escalation & Questions

**Q: What if I need an exception?**
A: [Process for exception requests]

**Q: Who enforces this rule?**
A: [Role/tooling]

**Q: What if the tool reports false positive?**
A: [How to report]

---

## Migration (If Rule is New)

For existing codebases:

**Phase 1: Discovery** (Run linter, report violations)
**Phase 2: Fix** (Teams have N days to fix)
**Phase 3: Enforce** (Pre-commit hook blocks new violations)

---

## Version History

| Version               | Date       | Changes         |
| --------------------- | ---------- | --------------- |
| 1.0.0                 | 2026-XX-XX | Initial release |
| [2.0.0 if applicable] | [Date]     | [What changed]  |

---

## Related Documentation

- [Related Rule]: [Link]
- [Skill needed to comply]: [Link]
- [Security doc]: [Link]

---

**Last Updated**: 2026-04-03  
**Severity**: [critical|high|medium|low]  
**Enforced By**: [tools/roles]  
**Related Links**: [GUIDE-ENFORCEMENT.md](./GUIDE-ENFORCEMENT.md)
