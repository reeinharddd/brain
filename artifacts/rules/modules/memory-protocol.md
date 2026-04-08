# Memory Protocol - Engram

This module extends the base memory rules with an explicit operating protocol.

## Progressive disclosure

Always retrieve memory in layers:

1. `mem_search`
2. `mem_timeline`
3. `mem_get_observation`

Do not jump to full payload retrieval unless the summary and timeline were not
enough for the current task.

## Stable topic keys

Prefer semantic topic keys that survive multiple sessions.

Format:

`{project}:{domain}:{concept}`

Examples:

- `brain:architecture:sdd-dag`
- `brain:decisions:guardian-local-mode`
- `brain:patterns:skill-context-injection`

## Session closure

For substantial work:

1. save or update the relevant topic key
2. write a `mem_session_summary`
3. include namespace, decision, validation result, unresolved risk, and next step

## Agent guidance

- `orchestrator`: check prior memory before planning and at archive time
- `researcher`: search memory before exploration
- `architect`: search prior decisions before proposing or designing
- `implementer`: avoid broad memory retrieval; use only narrow, relevant context
- `debugger`: search similar bugs before root-cause analysis

## Namespace

When a project root is known, derive the namespace with:

`~/.brain/scripts/memory-namespace.sh [project_root]`

The namespace isolates project memory while global rules remain shared.
