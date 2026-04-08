---
name: memory-search
description: Search cross-session memory for relevant context before starting a task. Surfaces episodic and semantic memories from the MCP knowledge graph.
---

# /memory-search - Retrieve Relevant Memory

Use before starting any task where past decisions, learnings, or project context might be relevant.

## How to invoke

```text
/memory-search [query]
```

Examples:
```
/memory-search authentication patterns
/memory-search postgres migration decisions
/memory-search what did we decide about error handling
/memory-search last session summary
```

## What this does

1. Queries the MCP memory server (knowledge graph) for entities and observations matching the query
2. Queries Qdrant vector store (if available) for semantic similarity results
3. Merges and ranks results by relevance
4. Presents a structured context summary ready for injection into the current task

## Step by step

### Step 1: Query the knowledge graph

Call `search_nodes` with the query text:
```
search_nodes(query="[your query]")
```

This returns entities, their types, and related observations from the graph.

### Step 2: Query vector store (if Qdrant is running)

Check if Qdrant is reachable:
```bash
curl -sf http://localhost:6333/collections/brain-codebase-context
```

If reachable, perform a semantic search against the codebase context index.

### Step 3: Layer retrieval (progressive disclosure)

Follow the memory-protocol.md tiered approach:
1. `search_nodes` for summary
2. `open_nodes` for full detail on top-3 matches
3. Check `mem_timeline` for chronological context if decisions are relevant

### Step 4: Format output

Present the retrieved context in this format:

```
## Memory Context for: [query]

### From Knowledge Graph
[entity name] ([type]):
  - [observation 1]
  - [observation 2]

### From Codebase Index
[file/component]: [relevant excerpt]

### Last Session Summary
[summary if available]

### Relevant Decisions
[any ADRs or decisions retrieved]
```

### Step 5: Confirm relevance

Ask:
> "I found [N] relevant memories. Should I include this context in the current task?"

## Memory types to look for

| Type | entityType tag | When relevant |
| :--- | :--- | :--- |
| Decision | `Decision` | Architecture or tech choices |
| Preference | `Preference` | How the user likes things done |
| Learning | `Learning` | Mistakes or discoveries |
| Project State | `ProjectState` | Where a project currently stands |
| Session Summary | `SessionSummary` | What happened last time |
| Rule Candidate | `RuleCandidate` | Patterns not yet in canonical.md |
| Deferred Idea | `DeferredIdea` | Things to do but not now |

## Anti-patterns

- Do NOT load all memories -- use progressive disclosure
- Do NOT skip memory search for tasks > 30 min
- Do NOT inject irrelevant memories -- relevance-filter first
- Do NOT query memory in a loop -- one focused query is better than five scattered ones

## Relation to other commands

- Run `/memory-search` at session start alongside `/standup`
- After completing a task, use `/update-brain` or `mem_session_summary` to save new context
- Run `/handover` at end of session to persist full context
