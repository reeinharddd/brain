---
type: agent
id: agent-SLUG
title: [Agent Name]
version: 1.0.0
status: active
date_created: YYYY-MM-DD
language: en
category: documentation
---

<!-- markdownlint-disable-file -->

# [Agent Name]

## Role & Purpose

[What is this agent's primary responsibility?]

[When and why would I invoke this agent?]

## When Invoked

- **Trigger**: [What causes this agent to run?]
- **Context**: [What information is available?]
- **Expected outcome**: [What should the agent produce?]

## Protocol

### Input

```json
{
  "goal": "string - what the agent should accomplish",
  "context": "string - relevant background",
  "constraints": "array - rules to follow"
}
```

### Output

```json
{
  "result": "string - what was accomplished",
  "status": "success|error",
  "reasoning": "string - how the decision was made"
}
```

## State Machine

```
[IDLE] --invoke--> [PROCESSING] --success--> [COMPLETE]
                       |
                       +--error--> [FAILED] --retry--> [PROCESSING]
```

## Decision Logic

1. **Receive goal**: What needs to be done?
2. **Analyze context**: What information do I have?
3. **Evaluate options**: What are possible approaches?
4. **Choose action**: What's the best course?
5. **Execute**: Do the work
6. **Report**: Communicate results

## Constraints & Rules

- [Rule 1]: Must always...
- [Rule 2]: Cannot...
- [Rule 3]: When X happens, always Y

## Examples

### Example 1: Normal Operation

**Input**: {goal: "Create ADR", context: "Need to document decision"}  
**Output**: {result: "ADR-007 created", status: "success"}

### Example 2: Error Handling

**Input**: {goal: "Invalid request"}  
**Output**: {result: "Invalid input", status: "error"}

## Composition

This agent can compose with:

- [Agent A]: For subtask X
- [Agent B]: For subtask Y

---

**Maintained by**: [Team]
**Last updated**: [Date]
