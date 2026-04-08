## Module: Memory Types and Classification

Every memory entry stored in the MCP knowledge graph MUST be classified with one of the following
entityType values. Unclassified memories cannot be retrieved reliably and degrade retrieval quality.

---

### Entity Types

| entityType | What it stores | Retention |
| :--- | :--- | :--- |
| `Decision` | Architecture or technology choices and their rationale | Permanent |
| `Preference` | How the user prefers things done (style, tooling, workflow) | Permanent |
| `Learning` | Mistakes made, bugs found, lessons from past sessions | Long (90 days) |
| `ProjectState` | Current status of an active project (what is done, what is next) | Medium (30 days) |
| `SessionSummary` | End-of-session summary including decisions and open items | Medium (30 days) |
| `RuleCandidate` | Pattern observed repeatedly that might belong in canonical.md | Medium (60 days) |
| `DeferredIdea` | Things to explore later, ideas not acted on immediately | Short (14 days) |
| `Constraint` | Hard limits that must not be violated (security, compliance, budget) | Permanent |
| `ExternalFact` | Facts about external systems, APIs, or libraries | Short (7 days) |

---

### How to create a typed memory entry

Always pass entityType explicitly:

```
create_entities([{
  name: "AuthStrategyDecision",
  entityType: "Decision",
  observations: [
    "Chose JWT RS256 over sessions for stateless horizontal scaling",
    "Rejected sessions because they require shared Redis in multi-instance deploy",
    "Decided: 2026-03-20, project: api-gateway"
  ]
}])
```

Naming convention: `[Subject][Type]` in PascalCase.
Examples: `PostgresDecision`, `ErrorHandlingPreference`, `AuthBugLearning`, `ApiGatewayState`

---

### Relations between entities

Use `create_relations` to connect related memories:

```
create_relations([{
  from: "AuthStrategyDecision",
  to: "ApiGatewayState",
  relationType: "influences"
}])
```

Common relation types: `influences`, `requires`, `contradicts`, `supersedes`, `learned_from`

---

### Retrieval protocol (tiered)

To minimize token usage, always retrieve in layers:

1. `search_nodes(query)` - keyword scan, returns entity names and first observation
2. `open_nodes([name1, name2])` - full detail for shortlisted entities only
3. Stop when you have enough context - do not pull the full graph

Do NOT call `read_graph` to get all memories. The graph grows unboundedly.

---

### Memory lifecycle rules

- `Decision` and `Constraint` entries are permanent - never delete without user confirmation
- `SessionSummary` older than 30 days should be archived (moved to a `Archived` relation)
- `ExternalFact` entries expire fastest - library versions and API shapes change
- Run `scripts/consolidate-memory.sh` monthly to detect contradictions and promote RuleCandidates
- Run `scripts/consolidate-memory.sh --dry-run` to preview without writing

---

### Anti-patterns

- Do NOT store the same fact under multiple entity names without a `supersedes` relation
- Do NOT use generic names like `Note1` or `Idea` - they cannot be retrieved reliably
- Do NOT skip entityType - untyped entities will be flagged by consolidation as orphans
- Do NOT store project-specific state in the global brain memory without a namespace prefix
