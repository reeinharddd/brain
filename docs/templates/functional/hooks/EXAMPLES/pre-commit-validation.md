name: pre-commit-schema-validation
version: 1.0.0
status: stable
trigger: pre-commit
blocking: true

<!-- markdownlint-disable-file -->

# Hook: Pre-Commit Schema Validation

## Summary

Git pre-commit hook validating YAML/JSON schemas before commit.

## Behavior

**On commit attempt**:

1. Detects changed YAML/JSON files
2. Validates against schema (.schema.json)
3. If invalid: blocks commit, shows errors
4. If valid: allows commit

## Example

**File changed**: `docs/templates/functional/agents/orchestrator.md` (frontmatter)

**Validation**:

```
✅ YAML frontmatter valid
✅ Required fields present (name, version, status)
✅ model_config structure correct
✅ Commit allowed
```

**If error**:

```
❌ VALIDATION FAILED
❌ Field "status" must be one of: [stable, beta, deprecated]
❌ Current value: "unstable"
❌ Commit blocked. Fix and retry.
```

## Installation

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
git config core.hooksPath .githooks
```
