---
# DECISION TEMPLATE - Internal architectural/technology decisions

title: "[Decision Title]"
date: "[YYYY-MM-DD]"
author: "[Your Name]"
status: "accepted|rejected|pending|superseded"
category: "[architecture|technology|process|design]"

# For MEMORY integration (if this affects future work)

save_to_memory: true
memory_type: "Decision" # Will become entity in knowledge graph
memory_entity: "[DescriptiveTitle]Decision"
---

<!-- markdownlint-disable-file -->

# Decision: [Title]

**Date**: [YYYY-MM-DD]  
**Author**: [Your Name]  
**Status**: [Accepted/Rejected/Pending/Superseded]  
**Last Updated**: [YYYY-MM-DD]

---

## Context

### Problem Statement

[Clear description of the problem you're solving]

What situation led you to need to make a decision?

- [Factor 1]
- [Factor 2]
- [Factor 3]

### Constraints

What rules or limitations apply to this decision?

- [Constraint 1]: [Impact]
- [Constraint 2]: [Impact]
- [Constraint 3]: [Impact]

---

## Options Considered

### Option 1: [Name]

**Description**: [What is this option]

**Pros**:

- [Advantage 1]
- [Advantage 2]
- [Advantage 3]

**Cons**:

- [Disadvantage 1]
- [Disadvantage 2]
- [Disadvantage 3]

**Effort**: [How much work: low|medium|high]

**Risk**: [Risk level: low|medium|high] + [What could go wrong]

**Real Example**: [If you've seen this work before, cite it]

---

### Option 2: [Name]

[Same structure as Option 1]

---

### Option 3: [Name]

[Same structure]

---

## Decision

### What We Chose

> We chose **Option [N]: [Name]** for the following reasons.

### Rationale

Why is this the best option?

1. **Primary Reason**: [Strongest reason]
   - Evidence: [Data, experience, or research]
2. **Secondary Reason**: [Next most important]
   - Evidence: [Backing data]

3. **Trade-offs Accepted**:
   - We accept [disadvantage] because [reason]
   - We accept [disadvantage] because [reason]

### Rejected Alternatives

**Why NOT Option 1?**

- [Reason 1]
- [Reason 2]

**Why NOT Option 2?**

- [Reason 1]
- [Reason 2]

---

## Consequences

### Positive Impacts

What good things result from this decision?

- [Impact 1]: [How it helps]
- [Impact 2]: [How it helps]
- [Impact 3]: [How it helps]

### Risks & Mitigations

What could go wrong?

| Risk     | Mitigation              | Owner         |
| -------- | ----------------------- | ------------- |
| [Risk 1] | [How to prevent/handle] | [Who handles] |
| [Risk 2] | [How to prevent/handle] | [Who handles] |
| [Risk 3] | [How to prevent/handle] | [Who handles] |

### Implementation Changes

What has to change now that we've made this decision?

- [ ] [Change 1]
- [ ] [Change 2]
- [ ] [Change 3]

---

## Validation

### How Will We Know This Decision Was Right?

Define success criteria:

- [Metric 1]: Target value
- [Metric 2]: Target value
- [Metric 3]: Target value

### Review Schedule

- **1 week**: [What to check]
- **1 month**: [What to evaluate]
- **3 months**: [Long-term assessment]

---

## Real-World Example

If you've implemented or seen this decision work before:

**Context**: [What was the situation]

**How It Went**: [What actually happened]

**Results**:

- [Metric 1]: [Result]
- [Metric 2]: [Result]

**Lessons**:

- [What we learned]

---

## Related Decisions

**Prior Decisions That Affect This**:

- [Prior Decision 1]: [How it relates]
- [Prior Decision 2]: [How it relates]

**Future Decisions Affected**:

- [Future Decision 1]: [Dependency]
- [Future Decision 2]: [Dependency]

---

## Supersessions

**If this decision replaces an older one**:

- Supersedes: [Old Decision ID]
- Justification: [Why we changed our mind]
- Migration Path: [How to transition]

**If this decision was later overridden**:

- Superseded By: [New Decision ID]
- Reason: [Why we changed]

---

## Appendix

### Supporting Research

External references, papers, or authoritative sources:

- [Reference 1]: [Link + summary]
- [Reference 2]: [Link + summary]

### Team Feedback

Who was involved in making this decision?

- [Role/Person]: [Their input]
- [Role/Person]: [Their input]

### Code/Implementation

Pointers to where this decision is implemented:

- [File/Component]: [Link]
- [File/Component]: [Link]

---

**Decision ID**: [auto-generated based on date + slug]  
**Tags**: [decision, architecture, technology-choice]  
**Archive Date**: [N/A unless superseded]

---

**Reviewed By**: [Reviewer name, date]  
**Approved By**: [Approver name, date]
