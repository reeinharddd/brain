---
name: handover
description: Generate a context document for the next session. Captures what was done, what is pending, and what the next agent needs to know to continue without losing context.
---

# /handover — Session Context Transfer

Use at the end of a session, or before handing off work to another agent or model.

## How to invoke

```
/handover
```

Or at the end of a work session. Can also be triggered mid-session when switching contexts.

## What this command does

1. **Summarizes what was accomplished** in this session
2. **Lists what is in progress** (partially done tasks)
3. **Lists what is pending** (next steps, blockers)
4. **Captures key decisions** made during the session
5. **Saves everything to Engram** with the project tag
6. **Produces a handover document** ready for the next agent/session

## The Handover Document

```markdown
## Handover: [Project Name]
**Date**: [YYYY-MM-DD]
**Session**: [brief description of what this session was about]

### ✅ Done this session
- [What was completed, with specifics]
- [Include file names, function names, API endpoints changed]

### 🔄 In progress (incomplete)
- [Task X]: [what's been done, what remains]
- [Task Y]: [last state, next step to take]

### ⏳ Pending (not started)
- [Task Z]: [description, any relevant context]
- [Blocker if any]: [what's blocking this]

### 🧠 Key decisions made
- **[Topic]**: decided [X] because [reason]
- **[Topic]**: rejected [Y] because [reason]

### ⚠️ Known issues / gotchas
- [Issue]: [what it is, whether it's blocking, any workarounds]

### 🗺️ Where to start next session
1. [First thing to do — be specific]
2. [Second thing]

### Context for next agent
- Working branch: [branch name]
- Environment: [relevant env vars, services needed]
- Key files: [files most relevant to current work]
```

## Step by step

### Step 1: Review the session
Look back at what was done:
- What files were changed?
- What decisions were made?
- What was left incomplete?

### Step 2: Draft the handover
Fill in each section of the template above. Be specific — the next agent has no memory of what happened here.

### Step 3: Save to Engram
Tag with: `[project-name]`, `handover`, `[date]`

### Step 4: Commit if needed
If there are uncommitted changes: stage, commit, and push before the session ends.

## When NOT to skip this

- Never end a session with uncommitted changes and no handover if the work will continue
- Never assume the next session will remember what happened in this one
- Never hand off without specifying the exact next step
