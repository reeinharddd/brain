---
type: tool
id: tool-SLUG
title: [Tool Name]
version: 1.0.0
status: active
date_created: YYYY-MM-DD
language: en
category: documentation
---

## Tool Details

## Purpose

[What does this tool do?]

## Invocation

### OpenAI Format

```json
{
  "type": "function",
  "function": {
    "name": "[tool_name]",
    "description": "[What it does]",
    "parameters": {
      "type": "object",
      "properties": {
        "param1": {
          "type": "string",
          "description": "[What param1 does]"
        }
      },
      "required": ["param1"]
    }
  }
}
```

### Anthropic Format

```json
{
  "name": "[tool_name]",
  "description": "[What it does]",
  "input_schema": {
    "type": "object",
    "properties": {
      "param1": {
        "type": "string",
        "description": "[What param1 does]"
      }
    },
    "required": ["param1"]
  }
}
```

## Response Format

```json
{
  "status": "success|error",
  "data": {},
  "error_message": "[if status is error]"
}
```

## Error Codes

| Code          | Meaning                 | Retry? |
| ------------- | ----------------------- | ------ |
| INVALID_PARAM | Invalid parameter value | No     |
| TIMEOUT       | Operation timed out     | Yes    |
| NOT_FOUND     | Resource not found      | No     |

## Constraints

- Maximum input size: [X bytes]
- Rate limit: [Requests/minute]
- Timeout: [Seconds]

## Examples

```text
Input: {param1: "value1"}
Output: {status: "success", data: {...}}
```

---

**Status**: Active  
**Last updated**: [Date]
