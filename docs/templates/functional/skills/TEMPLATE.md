---
# SKILL TEMPLATE - Copy to create new skill

id: "[skill-id]"
name: "[Skill Name]"
version: "[1.0.0]"
type: "skill"
category: "[engineering|debugging|architecture|documentation]"
scope: "global"
status: "stable"

# ACTIVATION
activation:
  keywords:
    - "[primary-keyword]"
    - "[secondary-keyword]"
    - "[search-keyword]"
  triggers:
    - "explicit_user_request"
    - "[other_trigger]"

# PREREQUISITES
prerequisites:
  knowledge: ["[Required knowledge]", "[Required knowledge]"]
  tools: ["[Tool needed]", "[Tool needed]"]
  context: "[What context must exist? e.g., 'runnable code']"

# METHODS
methods:
  - name: "[method-1]"
    description: "[What this method does]"
    complexity: "low|medium|high"
    time: "[time-range]"

  - name: "[method-2]"
    description: "[Description]"
    complexity: "medium"
    time: "[time-range]"

# VALIDATED EXAMPLES
examples:
  - title: "[Real example title]"
    context: "[Context/scenario]"
    method: "[which methods used]"
    result: "[What was the outcome]"
    time: "[Actual time taken]"
    difficulty: "[easy|medium|hard]"
    evidence: "[How to verify success]"

# RAG & SEARCH
keywords:
  - "[domain]"
  - "[subdomain]"
  - "[search-term-1]"
chunk_strategy: "by-section"
priority: "high"

# INTEGRATION
integrates_with:
  skills: ["[Related skill]", "[Related skill]"]
  agents: ["[Uses agent]", "[Uses agent]"]
  tools: ["[Tool needed]"]
  mcps: ["[MCP needed]"]
---

<!-- markdownlint-disable-file -->

# [Skill Name]

## Overview

[2-3 sentence description of what this skill is and why it matters]

## When to Use This Skill

Use this skill when:

- [Condition 1]
- [Condition 2]
- [Condition 3]

Do NOT use when:

- [When it's not applicable]
- [When another skill is better]

## What You'll Need

### Knowledge Prerequisites

- [Understanding of concept X]
- [Familiarity with tool Y]

### Tools & Environment

- [Tool 1]: [what for]
- [Tool 2]: [what for]

### Context Requirements

- [What must be true about the codebase/system]

## The Methodology

### Phase 1: [Phase Name]

**Goal**: [What you're trying to achieve]

**Steps**:

1. [Step 1]
2. [Step 2]
3. [Step 3]

**Output**: [What you should have after this phase]

**Time Estimate**: [realistic range]

---

### Phase 2: [Phase Name]

**Goal**: [Goal]

**Steps**:

1. [Step 1]
2. [Step 2]

**Output**: [Result]

**Time Estimate**: [range]

---

## Real-World Examples

### Example 1: [Title]

**Scenario**: [What was the situation]

**Applied Methodology**:

1. Phase 1: [What was done]
   - Result: [Outcome]
   - Time: [Actual]

2. Phase 2: [What was done]
   - Result: [Outcome]
   - Time: [Actual]

3. Result: [Final outcome]

**Metrics**:

- [Metric 1]: [Value]
- [Metric 2]: [Value]

**Why This Worked**:

- [Key factor 1]
- [Key factor 2]

---

### Example 2: [Title]

[Similar structure to Example 1]

---

### Example 3: [Edge Case or Complex Scenario]

[Similar structure]

---

## Common Mistakes

| Mistake     | Fix                | Why      |
| ----------- | ------------------ | -------- |
| [Mistake 1] | [Correct approach] | [Reason] |
| [Mistake 2] | [Correct approach] | [Reason] |
| [Mistake 3] | [Correct approach] | [Reason] |

---

## DO's ✅

| DO                                | Why                       | Example                               |
| --------------------------------- | ------------------------- | ------------------------------------- |
| **Include concrete examples**     | Abstract = not useful     | Show real code + real scenario        |
| **Estimate time realistically**   | Affects routing decisions | "15-60 min" not "5-10 hours"          |
| **List prerequisites explicitly** | Prevents wrong usage      | "Runnable code", "Access to logs"     |
| **Validate examples first**       | Guarantee they work       | Test locally before publishing        |
| **Provide decision trees**        | Helps navigate complexity | "If X, do Y" not just "do everything" |
| **Link to related skills**        | Helps discovery           | "Also see: debugging-methodology"     |

---

## DON'Ts ❌

| DON'T                            | Why                            | Bad Example                            |
| -------------------------------- | ------------------------------ | -------------------------------------- |
| **Don't use pseudo-code**        | Real code is clearer           | ❌ "get logs" ✅ "journalctl -u svc"   |
| **Don't assume expertise**       | Skills must work for beginners | ❌ "Use LLVM IR" ✅ Explain what it is |
| **Don't have > 5 prerequisites** | Too niche if too many          | Limit to 3 core requirements           |
| **Don't skip error examples**    | Real world has errors          | Show what to do if API fails           |
| **Don't hardcode versions**      | Becomes stale                  | ❌ "Python 3.8.2" ✅ "Python 3.8+"     |
| **Don't > 200 lines**            | Cognitive overload             | Split into focused sub-skills          |

---

## Escalation Points

When to escalate / ask for help:

- **If you've spent > 2 hours** on Phase 1: Escalate to architect
- **If you've tested > 5 hypotheses** without finding root cause: Escalate to research phase
- **If error is in external system**: Contact that system's owner
- **If fix is outside your scope**: Escalate to appropriate agent

---

## Integration

This skill integrates with:

**Other Skills**:

- [Skill]: [How they relate]
- [Skill]: [How they relate]

**Agents**:

- [Agent]: [When they use this skill]
- [Agent]: [When they use this skill]

**Tools** (required):

- [Tool]: [Used for what]

**MCPs** (if applicable):

- [MCP]: [What it provides]

---

## Troubleshooting

**Q: How do I know if I'm doing this right?**
A: [Validation method]

**Q: What if [common issue]?**
A: [Solution]

**Q: Can I skip [phase]?**
A: [Answer + rationale]

---

## Version History

| Version | Date       | Changes         |
| ------- | ---------- | --------------- |
| 1.0.0   | 2026-XX-XX | Initial release |

---

**Last Updated**: 2026-04-03  
**Validated By**: [Your name]  
**Related Links**: [GUIDE-METHODOLOGY.md](./GUIDE-METHODOLOGY.md)
