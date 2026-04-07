# Handover Protocol

## Version: 1.0.0
## Objective: Ensure zero-context-loss during agent switches or session ends.

The `/handover` command (or any agent-to-agent delegator) MUST produce a summary following this structure:

### 1. Goal & Context
What are we building/fixing and why? Mention relevant issue/ticket IDs.

### 2. Current State
What is the latest working revision? What files were modified?
- **Status**: (Working / Broken / In-progress)
- **Last Commit**: Short description.

### 3. Decisions Made
Key architectural or technical choices (with links to ADRs if applicable).

### 4. Blockers & Risks
What is stopping us? What could break in the next phase?

### 5. Next Steps (Task by Task)
Bullet list of atomic tasks for the incoming agent.
- `[ ] Task 1 (SDD Phase X)`
- `[ ] Task 2 (Verify with Y)`

## Handover Anti-Patterns
- **The Dump**: Sending raw logs without a summary.
- **The Mystery**: "I fixed most of it, you finish it" (No list of remaining tasks).
- **The Stale Checkpoint**: Referring to files that have already been modified.
