---
name: consolidate
description: Run the memory consolidation pipeline. Detects contradictions, surfaces rule candidates, updates manifest. Run monthly or when memory feels noisy.
---

# /consolidate - Memory Consolidation

Use when cross-session memory has grown large, feels inconsistent, or after a long project ends.

## How to invoke

```text
/consolidate
/consolidate --dry-run
```

## What it does

1. Reads all entities from the MCP knowledge graph via `read_graph`
2. Groups entries by entityType
3. Detects contradictions (same entity, conflicting observations)
4. Identifies high-frequency patterns (seen 3+ times) as `RuleCandidate` promotions
5. Flags stale `ExternalFact` entries (> 7 days)
6. Updates `memory/manifest.json` with stats
7. Writes a consolidation report to `logs/consolidation/`

## Step by step

### Step 1: Run dry run first

```bash
bash ~/.brain/scripts/consolidate-memory.sh --dry-run
```

Review the report for anything surprising before committing changes.

### Step 2: Review candidates

The report will include `canonical_update_candidates` - patterns seen 3+ times across sessions.
For each candidate, decide:
- Does this belong in `canonical.md` as a universal rule? -> Use `/update-brain`
- Is it project-specific? -> Leave it in memory, note the project namespace
- Is it outdated? -> Delete the entities manually

### Step 3: Run for real

```bash
bash ~/.brain/scripts/consolidate-memory.sh
```

### Step 4: Act on contradictions

If the report shows contradictions (same entity, conflicting types), resolve manually:

```
# Delete the wrong entity
delete_entities(["WrongEntityName"])

# Recreate correctly
create_entities([{
  name: "CorrectEntityName",
  entityType: "Decision",
  observations: ["..."]
}])
```

### Step 5: Promote rule candidates (optional)

If the report surfaces patterns that should be global rules:

```text
/update-brain
```

Follow the update-brain protocol to propose and apply the change to canonical.md.

## When to run

| Trigger | Action |
| :--- | :--- |
| End of major project | `/consolidate` |
| Memory feels inconsistent | `/consolidate --dry-run` first |
| Monthly maintenance | Automated by cron (see `scripts/cron-setup.sh`) |
| Before onboarding a new machine | `/consolidate` then push brain repo |

## Output

The report is saved to:
```
~/.brain/logs/consolidation/[TIMESTAMP].json
```

Fields: `total_entities`, `entity_breakdown`, `high_frequency_patterns`, `contradictions`,
`canonical_update_candidates`, `recommendations`
