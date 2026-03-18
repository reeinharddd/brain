## Module: Memory Protocol

### Topic keys and upserts

When storing persistent memory, prefer stable topic keys over ad-hoc duplicates.

1. First call `mem_suggest_topic_key` or the equivalent topic-key selection step
2. Reuse an existing topic key when the new information extends the same decision or concept
3. Only create a new topic key when the subject is materially different

### Progressive disclosure

To minimize token usage, memory retrieval must happen in layers:

1. `mem_search` for a high-level summary of relevant memories
2. `mem_timeline` for chronological context on the shortlisted topic
3. `mem_get_observation` only for the exact observation that needs full detail

Do not pull full memory payloads before the summary and timeline indicate they are relevant.

### Session closure

At the end of any significant task or session:

1. Save a concise `mem_session_summary`
2. Include the final decision, validation result, unresolved risks, and next step
3. Attach the project namespace when available

### Multi-project namespace convention

When a project root is known, derive a stable namespace with:

`~/.brain/scripts/memory-namespace.sh [project_root]`

Use that namespace for reads and writes so project memory stays isolated while global rules remain shared.
