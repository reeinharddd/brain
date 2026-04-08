---
name: standup
description: Quick session kickoff. Check what's pending from last session, set today's goal, and get context from memory.
---

# /standup — Session Kickoff

Use at the beginning of any new session to get oriented quickly.

## How to invoke

```
/standup [optional: project name or focus area]
```

Examples:
- `/standup` — general standup for whatever I was working on
- `/standup mnemos` — standup scoped to the mnemos project

## What this command does

1. **Checks Engram** for the last handover or session notes for the relevant project
2. **Reviews git status** (if in a repo) — uncommitted changes, current branch
3. **Shows today's context** — date, active project, last known state
4. **Helps set today's goal** — what will be done by end of this session
5. **Checks for blockers** — anything that might prevent progress

## Step by step

### Step 1: Load context
Check Engram for:
- Most recent handover document for this project
- Any pending tasks, blockers, or deferred items
- Key decisions that affect today's work

### Step 2: Check repo state
If in a git repo:
```bash
git status
git log --oneline -5
```

Show: current branch, any uncommitted changes, last 5 commits.

### Step 3: Orient
Show a brief summary:
- What was last done
- What is currently in progress
- What is the next priority

### Step 4: Set today's goal
Ask (or propose based on context):
> "Based on the last session, the next step is [X]. Is that the goal for today, or do you want to focus on something else?"

Once confirmed: state the goal clearly.
> "Today's goal: [specific outcome]. Let's start with [first task]."

## Output format

```markdown
## Standup: [Project] — [Date]

### Last session summary
[1-3 bullet points of what was done]

### Currently in progress
- [task]: [last state]

### Pending / next
- [task 1] — priority: high
- [task 2] — priority: medium

### Blockers
- [blocker if any]

### Git state
- Branch: [name]
- Uncommitted: [yes/no]
- Last commit: [message]

---
**Today's goal**: [specific outcome]
**Starting with**: [first concrete action]
```

## Tips

- If Engram has no context for this project: ask the user for a 1-sentence update on where things stand
- If there's an uncommitted change from last session: surface it and ask if it should be committed or discarded
- Keep standup short: this is orientation, not planning. Use `/plan` if the session needs a full plan
