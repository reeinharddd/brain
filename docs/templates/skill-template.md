---
type: skill
id: skill-SLUG
title: [Skill Name]
version: 1.0.0
status: active
date_created: YYYY-MM-DD
language: en
category: documentation
---

## Skill Details

## What It Does

[One-sentence description of what this skill does]

[Detailed explanation of purpose and use cases]

## Input Contract

```json
{
  "type": "object",
  "properties": {
    "param1": {
      "type": "string",
      "description": "What this parameter does"
    },
    "param2": {
      "type": "number",
      "description": "What this parameter does"
    }
  },
  "required": ["param1"]
}
```

### Validation Rules

- param1: Must not be empty
- param2: Must be positive

## Output Contract

```json
{
  "type": "object",
  "properties": {
    "result": {
      "type": "string",
      "description": "Result of operation"
    },
    "status": {
      "type": "string",
      "enum": ["success", "error", "partial"]
    }
  },
  "required": ["result", "status"]
}
```

## Error Cases & Handling

| Error Code    | HTTP | Cause                   | Resolution                      |
| ------------- | ---- | ----------------------- | ------------------------------- |
| INVALID_INPUT | 400  | Missing required fields | Provide all required parameters |
| NOT_FOUND     | 404  | Resource doesn't exist  | Verify ID is correct            |
| TIMEOUT       | 504  | Operation took too long | Retry with longer timeout       |

## Examples

### Example 1: Basic Usage

**Input**:

```json
{
  "param1": "value1",
  "param2": 42
}
```

**Output**:

```json
{
  "result": "Operation completed",
  "status": "success"
}
```

### Example 2: Error Case

**Input**:

```json
{
  "param2": 42
}
```

**Output**:

```json
{
  "result": "missing param1",
  "status": "error"
}
```

## Anti-Patterns

❌ **Don't**: Assume param1 is always provided
✅ **Do**: Always validate required fields

❌ **Don't**: Return incomplete results on error
✅ **Do**: Return error object with status "error"

## Compatibility

- **Models**: Claude 3+, GPT-4+
- **Systems**: All platforms
- **Versions**: Compatible with v1.0+

## Cost & Performance

- **Latency**: <100ms typical
- **Cost**: Free
- **Rate limit**: 100 requests/minute

---

**Maintained by**: [Team]
**Last updated**: [Date]
