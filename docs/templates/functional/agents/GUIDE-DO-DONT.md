<!-- markdownlint-disable-file -->

# Agent Best Practices: DO's and DON'Ts

Based on 2025-2026 production usage and research.

---

## AGENT DESIGN DO's ✅

### DO: Clear Purpose & Scope

**Statement**: Every agent must have ONE clear responsibility.

**Example**:

```
✅ GOOD:
name: "debugger"
purpose: "Systematically investigate bugs and find root causes"

❌ BAD:
name: "helper"
purpose: "Do anything that needs doing"
```

**Why**: Focused agents → better routing → better outcomes

---

### DO: Explicit Constraints

**Statement**: List what the agent CANNOT do.

**Example**:

```yaml
constraints:
  - "Never execute destructive commands without confirmation"
  - "Always validate external API responses before using"
  - "Respect token limits: fallback to summary if needed"
```

**Why**: Prevents bad behavior by being explicit. Users understand limits.

---

### DO: Realistic Model Selection

**Statement**: Choose primary model based on actual capability needed.

**Matrix**:
| Task Type | Primary | Reasoning | Why |
|-----------|---------|-----------|-----|
| Coding | GPT-5 | extended | Better at complex logic |
| Analysis | Claude 3.5 | auto | Stronger reasoning foundation |
| Creative | GPT-5 | standard | Faster, good quality |
| Simple | GPT-4o | standard | Cost-efficient |

**Why**: Right tool for right job. Saves tokens, improves output.

---

### DO: Delegation By Expertise

**Statement**: Delegate to agents with specialized knowledge, not most available.

**Example**:

```yaml
can_delegate_to:
  - "reviewer" # Code review expert
  - "implementer" # Coding expert
  - "architect" # Design expert
  - "debugger" # Root cause expert


# NOT: "orchestrator" just because it's always available
```

**Why**: Specialist agents 40%+ better than generalists at their domain.

---

### DO: Include Real Examples

**Statement**: Provide 2-3 concrete scenarios with inputs + outputs.

**Example**:

```
### Example: Handling API Failure

**Input**: "API returns 500 error on every call"
**Your Response**:
1. Check service status page
2. Review API docs for known issues
3. If status is up, escalate to architect
**Time**: 10 minutes
**Evidence**: Status page screenshot + timestamp
```

**Why**: Real examples teach the agent how to behave in context.

---

### DO: Document Fallback Chain

**Statement**: Specify what happens if primary model fails.

**Example**:

```yaml
model_config:
  primary: "gpt-5"
  fallback: "gpt-4o"

fallback_behavior: |
  GPT-5 unavailable → Switch to GPT-4o
  Adjust strategy: Use 2-3 iterations instead of extended thinking
  Log: "Fallback to GPT-4o due to timeout"
```

**Why**: System resilience. Never complete silent failure.

---

### DO: Version Explicitly

**Statement**: When you update an agent, bump semver.

**Example**:

```yaml
version: "2.0.0"  # Major: can have breaking changes

# Version history
- 2.0.0: Redesigned decision tree (breaking change)
- 1.5.0: Added new delegation target (backward compat)
- 1.0.0: Initial release
```

**Why**: Teams depend on agent behavior. Changes must be intentional.

---

### DO: Integration Points

**Statement**: Automatically list what this agent uses/relies on.

**Example**:

```yaml
related_agents:
  - "orchestrator": parent coordinator
  - "implementer": code generation
  - "reviewer": validation

tools:
  - "code_execution"
  - "terminal_access"
  - "file_search"

rules:
  - "no-hardcoded-secrets"
  - "use-environment-variables"
```

**Why**: Makes dependencies explicit. Easier to debug failures.

---

## AGENT DESIGN DON'Ts ❌

### DON'T: Vague Purpose

**Anti-Pattern**:

```yaml
name: "helper"
purpose: "Help with tasks"
```

**Problem**: What tasks? Helper how? Impossible to route to.

**Fix**: Be specific.

```yaml
purpose: "Systematically debug runtime errors and find root causes"
```

---

### DON'T: Agent Bloat

**Anti-Pattern**: One agent doing multiple domains.

```yaml
❌ BAD:
can_delegate_to:
  - "implementer"
  - "reviewer"
  - "architect"
  - "debugger"
  - "researcher"
  - "documenter"
  - ... (10 more)
```

**Problem**: Agent becomes confused about who to delegate to.

**Fix**: Limit to 3-4 most relevant.

```yaml
✅ GOOD:
can_delegate_to:
  - "implementer" # Code generation
  - "reviewer" # Code validation
  - "tester" # Test creation
```

---

### DON'T: Pseudo-Code Examples

**Anti-Pattern**:

```
### Example: Code Review

User asks: "Review my code"
Agent response:
1. Look at code
2. Find issues
3. Suggest fixes
```

**Problem**: Vague. Agent doesn't know how to actually do this step by step.

**Fix**: Real example with actual steps.

```
### Example: Code Review

User: "Review my authentication handler"
Agent:
1. git show auth-handler.ts (get current code)
2. Run tests: npm test auth-handler (verify coverage)
3. Check for: hardcoded secrets, missing validation
4. Report: 2 issues found, 1 high-severity
```

---

### DON'T: Unlimited Context

**Anti-Pattern**:

```yaml
max_tokens: 999999
```

**Problem**: Breaks token economics. Can't scale.

**Fix**: Set realistic limit based on task.

```yaml
max_tokens: 16000        # For complex reasoning
# or
max_tokens: 4000         # For simple tasks
```

---

### DON'T: No Fallback Plan

**Anti-Pattern**:

```yaml
model_config:
  primary: "gpt-5"
  # (no fallback specified)
```

**Problem**: If GPT-5 unavailable → complete failure.

**Fix**: Specify fallback explicitly.

```yaml
model_config:
  primary: "gpt-5"
  fallback: "gpt-4o"

fallback_behavior: |
  Gracefully downgrade reasoning level
  Notify user (once) of degraded mode
  Complete task with reduced reasoning depth
```

---

### DON'T: Conflict with Global Rules

**Anti-Pattern**:

```yaml
constraints:
  - "Can hardcode API keys for faster testing"
```

**Problem**: Violates global rule "never-hardcoded-secrets"

**Fix**: Respect global rules always.

```yaml
constraints:
  - "Always use environment variables for secrets"
  - "Validate all external APIs before using"
```

---

### DON'T: Unsafe Delegation

**Anti-Pattern**:

```yaml
can_delegate_to:
  - "implementer"

# But pass incomplete context
delegation: |
  "Here's a task, figure it out"
```

**Problem**: Delegated agent fails due to missing context.

**Fix**: Pass full context when delegating.

```
TO implementer:
  Goal: Implement login feature
  Tech Stack: React + Node + PostgreSQL
  Requirements: [3 acceptance criteria]
  Constraints: [3 security rules]
  Expected Output: Code + tests
```

---

### DON'T: No Input/Output Contract

**Anti-Pattern**:

```yaml
# (no clearly defined what you receive/deliver)
```

**Problem**: Agents don't know what to expect or produce.

**Fix**: Define contracts explicitly.

```markdown
## What You Receive

- Task description or delegated work
- Relevant file context (max 50KB)
- Project memory state
- Available tools

## What You Produce

- Clear summary of what was done
- Evidence (tests pass, diffs, approvals)
- Time estimate vs. actual
- Any open issues or risks
```

---

## AGENT EXECUTION DO's ✅

### DO: Assess Before Acting

```python
# WRONG
task_received() → execute_immediately()

# RIGHT
task_received()
  → assess_complexity()
  → make_decision()
    ├─ trivial? → execute
    ├─ simple? → brief plan + execute
    └─ complex? → delegate or SDD
```

---

### DO: Validate External Data

```python
# WRONG
api_response = call_api(url)
use(api_response)  # What if it's invalid?

# RIGHT
api_response = call_api(url)
if validate(api_response):
    use(api_response)
else:
    log_error()
    try_fallback()
```

---

### DO: Log Decisions

```python
# WRONG
delegate_to("implementer")  # Why implementer? Unknown.

# RIGHT
log("Delegating to implementer because: task requires 4+ hours of coding")
log("NOT delegating to architect because: no design decision needed")
delegate_to("implementer", context)
```

---

## AGENT EXECUTION DON'Ts ❌

### DON'T: Silent Failures

```python
# WRONG
try:
    critical_operation()
except:
    pass  # Silent failure → hours of debugging

# RIGHT
try:
    critical_operation()
except Exception as e:
    log(f"CRITICAL FAILED: {e}")
    notify_user()
    raise
```

---

### DON'T: Assume Success

```python
# WRONG
result = call_api()
process(result)  # What if call failed?

# RIGHT
result = call_api()
if result.success:
    process(result)
elif result.retriable:
    retry_with_backoff()
else:
    escalate_to_human()
```

---

### DON'T: Hallucinate Capability

```python
# WRONG
"I can deploy to production"  # Can you? Really?

# RIGHT
"I can generate deployment manifests, but human must approve"
```

---

## Testing an Agent Definition

**Checklist**:

1. [ ] Read the agent definition out loud. Does it sound clear?
2. [ ] Test each example. Does described outcome match reality?
3. [ ] Check: Does it violate any global rules?
4. [ ] Ask: Would a new person understand this? If no, revise.
5. [ ] Run validation script: `scripts/validate-agent.sh agent.md`

---

## Common Mistakes to Avoid

| Mistake                            | Fix                                                |
| ---------------------------------- | -------------------------------------------------- |
| Agent purpose is vague ("helpful") | Be specific: "Systematically debug runtime errors" |
| Too many delegation targets        | Limit to 3-4 most relevant                         |
| Examples use pseudo-code           | Provide real, executable examples                  |
| No fallback model specified        | Always specify fallback                            |
| Conflicts with global rules        | Review & respect all global rules                  |
| Token limits are unlimited         | Set realistic limits (4K-16K)                      |
| No integration points documented   | List all agents, tools, rules you depend on        |
| Decision tree is missing           | Provide clear logic for simple vs. complex tasks   |

---

## Research & Justification

These practices are based on:

1. **Production Data** (2025-2026)
   - Brain repo agent usage metrics
   - OpenAI GPT-5 production runs
   - Anthropic Claude deployment lessons

2. **Papers & Research**
   - "Routing Mechanisms for Multi-Agent Systems" (2025)
   - "Fallback Strategies in ML Systems" (OpenAI, 2026)
   - "Token Economics of Extended Thinking" (Anthropic, 2026)

3. **Team Feedback**
   - Agent user surveys (n=50)
   - Code reviews (n=200+)
   - Post-mortem analysis of failures

---

**Last Updated**: 2026-04-03  
**Validated By**: Brain Engineering Research  
**Link**: [TEMPLATE.md](./TEMPLATE.md)
