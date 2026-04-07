---
type: state
id: state-SLUG
title: [System State Name]
version: 1.0.0
status: active
date_created: YYYY-MM-DD
language: en
category: documentation
---

## State Snapshot

## Current State

```json
{
  "field1": "value1",
  "field2": "value2",
  "status": "active|inactive|error"
}
```

## State Transitions

```text
[State A] --trigger--> [State B] --trigger--> [State C]
  |                      |
  +--error--> [Error State]
```

## Field Descriptions

| Field  | Type   | Current Value | Updatable? | Constraints   |
| ------ | ------ | ------------- | ---------- | ------------- |
| field1 | string | [value]       | Yes        | Max 100 chars |
| field2 | number | [value]       | No         | Read-only     |

## Last Updated

- **Date**: [YYYY-MM-DD HH:MM:SS]
- **By**: [System/User]
- **Reason**: [Why it was updated]

## Related Objects

- Related object 1 - [Relationship]
- Related object 2 - [Relationship]

## Lifecycle

- **Created**: [Date]
- **Last modified**: [Date]
- **Expires**: [Date, if applicable]

---

**Auto-generated**: [Yes/No]  
**Next sync**: [Date/Time]
