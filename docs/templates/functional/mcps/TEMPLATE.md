## <!-- markdownlint-disable-file -->

# MCP SERVER TEMPLATE - Model Context Protocol servers

id: "[mcp-id]"
name: "[MCP Name]"
version: "1.0.0"
type: "mcp"
description: "[What this MCP provides]"
status: "stable"
protocol_version: "sse" # or sse|http|stdio

# IMPLEMENTATION

implementation:
language: "[go|rust|typescript]"
host: "localhost"
port: "9090"
startup_cmd: "[command to start this MCP]"

# RESOURCES (what it exposes)

resources:

- name: "[resource-1]"
  type: "[type]"
  description: "[description]"

# TOOLS (what agents can call)

tools:

- name: "[tool-1]"
  description: "[What it does]"
  input_schema:
  type: "object"
  properties:
  param1: { type: "string", description: "[param desc]" }
  required: ["param1"]

# RAG

keywords: ["[domain]", "[use-case]"]
applies_to: "_.go, _.ts"

---

# MCP: [Name]

## Overview

[Description of what this MCP provides]

## Tools Available

### Tool 1: [name]

**Purpose**: [What it does]

**Signature**:

```
tool_1(param1: string, param2: number) → result
```

**Example**:

```json
{
  "name": "[tool-1]",
  "arguments": {
    "param1": "value",
    "param2": 42
  }
}
```

**Output**:

```json
{
  "success": true,
  "data": [...]
}
```

---

### Tool 2: [name]

[Similar structure]

---

## When to Use

Use this MCP when:

- [Scenario 1]
- [Scenario 2]

Don't use for:

- [Scenario 3]

## Integration

**Config Location**: `config/mcps.json`

**Invocation**:

```bash
# Daemon launches via
go run apps/daemon/cmd/braind/main.go

# Agents call via
http://localhost:9090/api/tools
```

## Error Handling

| Error        | Meaning   | Recovery         |
| ------------ | --------- | ---------------- |
| [error-code] | [Meaning] | [How to recover] |

---

**Last Updated**: 2026-04-03
