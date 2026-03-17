---
name: update-brain
description: Propose improvements to the global brain repo based on learnings from this session. The self-improvement loop.
---

# /update-brain — Brain Repo Self-Improvement

Use when something was learned in this session that should become global knowledge — applicable to all future projects and sessions.

## How to invoke

```
/update-brain
```

Or trigger explicitly: "this should go into the brain repo"

## The Self-Improvement Loop

```
Session learning → /update-brain → Proposal → User confirms → Applied → commit "brain: ..."
```

This is how the brain repo improves over time from real work, not just manual editing.

## Step by step

### Step 1: Identify what to capture
What type of update is this?

| Type | When to use |
|------|------------|
| `rule` | A principle that should apply universally to all projects |
| `agent` | A new agent behavior or improvement to an existing agent |
| `command` | A new slash command that would be reusable |
| `mcp` | A new MCP discovered that would be useful globally |
| `provider` | A new model or provider to add to providers.yml |
| `correction` | A mistake in existing rules that needs fixing |

### Step 2: Write the proposal
Use this exact format:

```markdown
## Brain Repo Update Proposal

**Type**: [rule / agent / command / mcp / provider / correction]
**File to modify**: ~/.brain/[relative path]
**Change summary**: [1 sentence]

**Proposed content**:
[the actual content to add or change — be specific]

**Reason**: [why this should be global, not project-specific]

**Risk**: [could this rule conflict with anything? cause issues in other projects?]
```

### Step 3: Ask for confirmation
Present the proposal and ask:
> "Should I apply this change to the brain repo? (yes / no / modify)"

**NEVER apply without explicit confirmation.**

### Step 4: Apply (if confirmed)
1. Apply the change to the appropriate file in `~/.brain/`
2. If it modified `rules/canonical.md` or a module: run `adapters/generate.sh`
3. Commit: `git -C ~/.brain commit -am "brain: [summary]"`
4. Optionally push: `git -C ~/.brain push`

### Step 5: If declined
If the user says no:
- Ask: should I save this as a deferred idea?
- If yes: save to Engram with tag `brain-improvement-deferred`

## What qualifies as a brain update

**✅ YES — belongs in brain:**
- A coding principle I realize I always follow but hadn't written down
- A debugging technique that proved effective
- A new agent behavior that would help in any project
- A model or MCP that should always be available

**❌ NO — belongs in the project:**
- A rule specific to this framework/language only
- A business rule for this client/product
- A project-specific agent or prompt
- Any configuration referencing a specific repo or environment

## Anti-patterns

- NEVER add project-specific rules to the brain repo
- NEVER apply changes without user confirmation
- NEVER update brain during a critical session — do it at the end
- NEVER make vague proposals — be specific about what changes
