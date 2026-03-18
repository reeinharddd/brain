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

1. **Captures the active SDD state** when the work used explicit phases
2. **Summarizes what was accomplished** in this session
3. **Lists what is in progress** (partially done tasks)
4. **Lists what is pending** (next steps, blockers)
5. **Captures key decisions** made during the session
6. **Saves everything to Engram** with the project namespace or tag
7. **Produces a handover document** ready for the next agent/session

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
- Use the project namespace when available
- Save or update the relevant session topic
- Call `mem_session_summary` for substantial work

### Step 4: Record the DAG state
If the session used explicit SDD artifacts, capture:
- current phase
- completed artifacts
- pending artifacts
- verification status

### Step 5: Commit if needed
If there are uncommitted changes: stage, commit, and push before the session ends when appropriate.

## When NOT to skip this

- Never end a session with uncommitted changes and no handover if the work will continue
- Never assume the next session will remember what happened in this one
- Never hand off without specifying the exact next step
- Never treat `/handover` as complete if the memory summary and next step are missing
