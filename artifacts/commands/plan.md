---
name: plan
description: Create a structured plan for any task > 30 minutes. Invokes the planner agent with the right context.
---

# /plan — Structured Task Planning

Use this command before starting any task that will take more than 30 minutes.

## How to invoke

```
/plan [brief description of what you want to accomplish]
```

## What this command does

1. **Invokes the Planner agent** with your goal
2. **Checks Engram** for any relevant past context (related decisions, prior work on the same topic)
3. **Produces a structured plan** with tasks, estimates, and decision log
4. **Asks for confirmation** before you start executing

## Step by step

### Step 1: Gather context
Before planning, check:
- Is there any relevant context in memory? (search Engram for the project/topic)
- Are there any existing docs, ADRs, or READMEs that define constraints?
- Is there a similar feature already implemented in the codebase?

### Step 2: Define the problem
Write a crisp problem statement:
> "I need to [achieve X] because [reason]. Success means [measurable outcome]. Constraints: [list key limits]."

### Step 3: Create the plan
Use the **Planner agent** to produce:
- Task list with estimates
- Dependencies between tasks
- Risks and mitigations
- Definition of done

### Step 4: Review and confirm
Present the plan. Ask the user:
> "Does this plan look right? Any missing tasks, wrong estimates, or constraints I missed?"

Do NOT start executing until you get explicit confirmation.

### Step 5: Save the plan
Once confirmed, save it to Engram with:
- Project name
- Plan date
- Key decisions

## Output format

```markdown
## Plan: [Goal]

**Estimated total**: [X hours/days]
**Priority**: [what to tackle first]

### Tasks
| # | Task | Agent | Est. | Depends on |
|---|------|-------|------|-----------|
| 1 | ... | implementer | 2h | — |
| 2 | ... | reviewer | 30m | 1 |

### Risks
1. [Risk] → [Mitigation]

### Done criteria
- [ ] [Specific, verifiable outcome]
- [ ] [...]

---
Confirm with: "looks good, let's start" or suggest changes.
```
