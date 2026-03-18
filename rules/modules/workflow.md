\n## Module: Workflow

\n### The AI Engineering Loop

\n## The AI Engineering Loop

Every task follows this cycle, no matter the size:

```text
Understand -> Plan -> Delegate -> Review -> Integrate -> Document
```text

\n### Spec-Driven Development (SDD) Phases
For complex features, the loop expands into explicit phases:
1. **Explore**: Analyze codebase, identify constraints, and feasibility
2. **Propose**: Draft internal proposal/RFC with options
3. **Spec**: Write formal technical specification
4. **Design**: System architecture, data flow, and UI/UX design
5. **Tasks**: Break down into atomic, manageable tasks (Kanban)
6. **Apply**: Implement code changes task by task
7. **Verify**: Run tests, linting, and manual validation
8. **Archive**: Close task, update documentation, and sync memory

1. **Understand**: What is the real problem? Who is it for? What does success look like?
2. **Plan**: Break it down. Use `/plan` for anything >30 min of estimated work.
   - Consider **Token Economics**: Prefer lean context over excessive file reading.
3. **Delegate**: Assign to the right agent with full context and constraints
4. **Review**: Never accept AI output without reading it. Use `/review` before merging
5. **Integrate**: Apply changes, run tests, verify behavior
6. **Document**: Update README, add comments, create ADR if architectural decision was made

\n### Project kickoff checklist

Before writing any code for a new project:


- [ ] Problem clearly defined in 1-2 sentences
- [ ] User/stakeholder identified
- [ ] Success criteria defined (how do we know when it's done?)
- [ ] Tech stack chosen with justification
- [ ] Risks identified (at least top 3)
- [ ] Git repo initialized with README and .gitignore
- [ ] `.env.example` created if secrets will be needed
- [ ] Basic project structure scaffolded

\n### Session management

**Start of session**:
1. Read context from memory (Engram) for the relevant project
2. Check what was left pending (use `/standup` or review last `/handover`)
3. Set clear goal for this session: "By the end of this session I will have X done"

**During session**:
- Commit frequently (every distinct working unit)
- Save discoveries to memory as you make them
- If you go down a rabbit hole, note it and come back to the main path

**End of session**:
1. Run `/handover` to generate context for next session
2. Commit all pending changes
3. Update the project's README or docs if anything changed
4. Save session learnings to memory

\n### Decision making

When facing a decision with multiple valid options:
1. State the options clearly
2. Identify the evaluation criteria (speed, cost, maintainability, complexity)
3. Make a decision
4. Document it - even informally in a comment or ADR

Use the planner agent for architectural decisions.
Use the researcher agent for "what's the best library for X" questions.
Don't spend more than 15 minutes deciding - pick the best option with current information.

\n### Task sizing

- **< 30 min**: Just do it. No formal plan needed.
- **30 min - 2 hrs**: Write a brief plan comment at the top of the session
- **> 2 hrs**: Use `/plan` to create a structured breakdown before starting
- **Multi-day**: Create a proper spec doc in `docs/`, track progress there

\n### Debugging workflow

1. Reproduce the bug reliably first
2. Identify the smallest code path that triggers it  
3. Add logging/breakpoints at the boundary of "working" vs. "broken"
4. Hypothesize cause (1-3 candidates)
5. Test hypotheses one at a time
6. Fix root cause, not symptom
7. Write a test that would have caught this bug
8. Commit with a descriptive message explaining what was wrong

Never fix a bug without understanding why it happened.
