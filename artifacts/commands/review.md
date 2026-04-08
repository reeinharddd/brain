---
name: review
description: Perform a thorough review of code, a PR, or a feature before merging or shipping. Invokes the Reviewer and optionally the Guardian agent.
---

# /review — Pre-merge Code Review

Use this before merging any significant change.

## How to invoke

```
/review [optional: specific area to focus on]
```

Examples:
- `/review` — full review of current changes
- `/review security` — focus on security issues
- `/review the auth module` — scope to a specific area

## What this command does

1. **Diffs current changes** (or specified file/PR)
2. **Invokes Reviewer agent** for code quality, correctness, tests
3. **Invokes Guardian agent** if changes touch: auth, payments, user data, file system, external APIs
4. **Produces structured findings** with severity levels
5. **Gives a final verdict**: APPROVE / REQUEST CHANGES / NEEDS DISCUSSION

## Step by step

### Step 1: Identify scope
What needs to be reviewed?
- If in a git repo: `git diff main...HEAD` or staged changes
- If a specific file: read that file
- If a PR: the PR description + diff

### Step 2: Run Reviewer agent
Check all categories:
- Security (if in scope)
- Correctness and edge cases
- Code quality and naming
- Error handling
- Test coverage
- Documentation

### Step 3: Escalate to Guardian if needed
If the change touches any security-sensitive code:
- Authentication or authorization
- Any external API or payment integration
- File system operations
- User data or PII

### Step 4: Synthesize findings
Produce the review output with:
- Summary verdict
- All findings sorted by severity (BLOCKER first)
- Specific, actionable fix suggestions

### Step 5: Decide
- APPROVE: no blockers, ready to merge
- REQUEST CHANGES: has blockers or majors that must be addressed
- NEEDS DISCUSSION: design or architectural concerns that need a conversation first

## Output format

```markdown
## Review: [Feature/PR Name]

**Files reviewed**: [list]
**Security check**: [yes/no — Guardian invoked]

### Findings

🔴 BLOCKER: [title]
...

🟡 MAJOR: [title]
...

✅ GOOD: [title]
...

---
### Verdict: REQUEST CHANGES
**Blockers to fix**: [N]
**Suggested changes**: [N]
```
