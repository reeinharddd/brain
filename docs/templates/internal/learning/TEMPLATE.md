## <!-- markdownlint-disable-file -->

# LEARNING RECORD TEMPLATE - Document bugs, lessons, patterns

title: "[What Happened]"
date: "[YYYY-MM-DD]"
author: "[Your Name]"
type: "bug-lesson|pattern-discovered|mistake-made|incident"
severity: "critical|high|medium|low"
status: "archived|reference"

---

# Learning: [Title]

**Date**: [YYYY-MM-DD]  
**Author**: [Your Name]  
**Severity**: [Critical/High/Medium/Low]  
**Type**: [Bug Lesson / Pattern Discovered / Mistake Made / Incident]

---

## What Happened

### Situation

[Describe the situation when this problem occurred]

- **Component**: [What system/code]
- **Context**: [What was happening]
- **Impact**: [Who/what was affected]

### Initial Symptom

[What was the first thing you noticed was wrong?]

---

## Root Cause

### What We Found

[The actual root cause, discovered through investigation]

**Contributing Factors**:

1. [Factor 1]
2. [Factor 2]
3. [Factor 3]

### Why It Happened

[Explanation of the chain of events]

---

## Timeline

When did each thing happen?

| Time    | Event                      |
| ------- | -------------------------- |
| [T-0]   | [Initial symptom observed] |
| [T+5m]  | [Investigation starts]     |
| [T+30m] | [Root cause identified]    |
| [T+60m] | [Fix applied]              |
| [T+90m] | [Validated]                |

---

## The Fix

### What We Changed

[How we fixed the root cause]

```
[Code changes / config changes / etc]
```

### Why This Works

[Explanation of how the fix addresses the root cause]

---

## How to Prevent

### System Level

What should we change to prevent this category of bug?

- **Idea 1**: [Process/automation/test that would catch this]
- **Idea 2**: [Another prevention mechanism]

### Recommendation

> [Single strongest recommendation to prevent this in future]

### Who Should Do This

- [Owner]: [Action needed]

---

## Lessons Learned

### What We Learned

1. **Learning 1**: [Key takeaway]
   - Why it matters: [Explanation]
2. **Learning 2**: [Key takeaway]
   - Why it matters: [Explanation]

3. **Learning 3**: [Key takeaway]
   - Why it matters: [Explanation]

### What To Do Differently Next Time

- ✅ Next time, [action] to prevent this
- ✅ Next time, [action] to catch this earlier
- ✅ Next time, [action] to minimize impact

---

## Pattern Discovered (If Applicable)

Is this an instance of a broader pattern we should know about?

### Pattern: [Name]

**Description**: [What the pattern is]

**Frequency**: [How often does this happen?]

**Severity if repeated**: [Impact on system]

**Better approach**: [How to avoid the pattern]

---

## Improvement Opportunities

### Quick Wins

Things we could do immediately:

- [ ] [Action]
- [ ] [Action]

### Medium Term

Things worth doing soon:

- [ ] [Add test to catch this]
- [ ] [Improve tooling]
- [ ] [Update documentation]

### Long Term

Structural improvements:

- [ ] [Architecture change]
- [ ] [Process improvement]
- [ ] [Tooling investment]

---

## Real Numbers

If applicable, include metrics:

- **Time to detect**: [How long before anyone noticed?]
- **Time to fix**: [How long from detection to resolution?]
- **Impact**: [Users affected / data loss / revenue impact / etc]
- **Cost**: [Estimation of remediation cost]

---

## Related Resources

### Similar Incidents

Have we seen this before?

- [Incident 1]: [Link] — [How it relates]
- [Incident 2]: [Link] — [How it relates]

### Recommended Reading

- [Skill]: [Link] — Why relevant
- [Rule]: [Link] — Relevant constraint
- [Guide]: [Link] — How to do it right

---

## Follow-Up

### Action Items

- [ ] [Action]: [Owner] by [Date]
- [ ] [Action]: [Owner] by [Date]
- [ ] [Action]: [Owner] by [Date]

### Success Criteria

How will we know the prevention was successful?

- [Metric 1]: Target value
- [Metric 2]: Target value

---

## Appendix: Investigation Notes

[Keep brief discovery notes here for reference]

**Note 1**: [Finding during investigation]

**Note 2**: [Finding]

---

**Created By**: [Name]  
**Reviewed By**: [Name]  
**Status**: LEARNING DOCUMENTED

---

## When to Use This Template

Use this template to document:

- ✅ Bugs that teach you something about the system
- ✅ Mistakes that resulted in outages or data loss
- ✅ Patterns you discovered that repeat
- ✅ Incidents that have systemic lessons

Do NOT use for:

- ❌ Minor typo fixes
- ❌ Routine performance tuning
- ❌ Things nobody needs to learn from
