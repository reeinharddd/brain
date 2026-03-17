---
name: research
description: Deep-dive research on a technology, library, pattern, or decision. Returns actionable findings with a clear recommendation.
---

# /research — Focused Technical Research

Use when you need a well-researched answer before making a decision.

## How to invoke

```
/research [question or topic]
```

Examples:
- `/research best state management for React in 2026`
- `/research tradeoffs between PostgreSQL and MongoDB for this use case`
- `/research how to implement rate limiting in Node.js`
- `/research is this library still maintained?`

## What this command does

1. **Invokes Researcher agent** with full context
2. **Checks Context7** for library-specific documentation (if relevant)
3. **Searches for current information** (versions, maintenance status, known issues)
4. **Returns a structured report** with a clear recommendation
5. **Saves findings to Engram** so they don't need to be repeated

## Step by step

### Step 1: Clarify the question
Before researching, restate the question precisely:
- What specific problem are we solving?
- Are there constraints? (budget, existing stack, team size, performance requirements)
- What does "good enough" look like for this decision?

### Step 2: Research
Use available tools:
- **Context7 MCP**: for official library documentation
- **Web search**: for recent comparisons, benchmarks, community sentiment
- **GitHub**: for dependency health (stars, last commit, open issues, contributors)
- **npm/PyPI/crates.io**: for download trends and version history

### Step 3: Evaluate
For decisions between options, use a scoring matrix:
- Criteria relevant to this specific context
- Not generic pros/cons

### Step 4: Recommend
Always end with one clear recommendation.
If genuinely unclear, state: "I cannot recommend between X and Y without knowing [missing information]."

### Step 5: Save to memory
Save the research summary and recommendation to Engram with tags: `[project-name]`, `[technology]`, `research`.

## Output format

```markdown
## Research: [Topic]

**Context**: [why this is being researched]
**Question**: [precise question]

### Findings
[structured explanation — use tables for comparisons]

### Recommendation
**Use [X]** because [1 clear reason].

[If conditions apply]: If [condition], prefer [Y] instead.

### Sources
- [source 1]
- [source 2]

### Saved to memory: [yes/no]
```
