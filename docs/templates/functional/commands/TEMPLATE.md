## <!-- markdownlint-disable-file -->

# COMMAND TEMPLATE - Special operations (/plan, /review, etc)

id: "[command-id]"
name: "[/command-name]"
version: "1.0.0"
type: "command"
description: "[One-line description]"
status: "stable"

# INVOCATION

invocation: "/[command-name] [arg1] [arg2]"
aliases: ["[alt-name]"]

# ROUTING

routed_to: "[daemon|cli|ui|all]"
requires_auth: false
requires_context: "[Optional context needed]"

# INTEGRATION

integrates_with:
agents: ["[Agent using this]"]
mcps: ["[MCP needed]"]

---

# Command: /[name]

## Usage

```
/[command-name] [arg1] [arg2]

Example:
/plan "Build user authentication"
```

## What It Does

[Clear description of what this command does]

## Arguments

| Arg    | Type   | Required | Description  |
| ------ | ------ | -------- | ------------ |
| [arg1] | string | yes      | [What is it] |
| [arg2] | list   | no       | [What is it] |

## Output

[What the command returns/displays]

## Example

Input:

```
/plan "Add 2FA to login"
```

Output:

```
[SDD breakdown with phases]
```

## When to Use

Use when:

- [Scenario 1]
- [Scenario 2]

Don't use for:

- [Scenario 3]

## Integration

**Daemon Endpoint**: `POST /api/commands/{command-id}`

**CLI Usage**: `brain [command-id] [args]`

**UI Button**: Yes/No + location

---

**Last Updated**: 2026-04-03
