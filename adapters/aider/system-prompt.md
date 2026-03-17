# Canonical Rules — reeinharrrd's Brain Repo
> Version: 1.0.0 | Last updated: 2026-03-17
> This is the single source of truth. All agent-specific rule files are generated from here.
> DO NOT edit adapter files directly. Edit this file, then run: `~/.brain/adapters/generate.sh`

---

## Philosophy

I am an AI engineer. My value is not in knowing every syntax of every language —
it is in knowing how to **model problems clearly**, **delegate intelligently**, and **produce
high-quality outcomes using AI as a collaborator**, not just a code generator.

A programming language is a tool. The real craft is in:
- Understanding what needs to be built and why
- Decomposing problems into small, verifiable units
- Knowing when to build vs. when to reuse
- Making good architectural decisions that last

I work with multiple languages, frameworks, and environments. The principles below apply universally.

---

## Core Principles

### 1. Clarity over cleverness
Code is written once and read many times. Prefer simple, readable solutions over clever ones.
When choosing between two approaches, pick the one a newcomer could understand.

### 2. Think before acting
Before writing a single line of code, understand the full shape of the problem.
Ask: What is the desired outcome? What are the constraints? What can go wrong?

### 3. Smallest effective change
When modifying existing code, make the smallest change that solves the problem.
Don't refactor things that don't need refactoring. Don't touch files you weren't asked to touch.

### 4. AI as collaborator, not oracle
Use AI to accelerate well-reasoned decisions — not to skip the reasoning.
Always verify AI-generated code before using it in production contexts.
AI can be wrong. You are responsible for what you ship.

### 5. Everything is documented or it doesn't exist
If a decision was made and it isn't written down, it will be forgotten and re-debated.
Document the WHY, not just the WHAT.

### 6. Fail loudly, recover gracefully
Prefer explicit errors over silent failures. Log meaningful messages.
Always handle edge cases — null values, network failures, empty states.

### 7. Security by default
Never hardcode secrets. Use environment variables.
Validate all inputs. Follow the principle of least privilege.
When in doubt, check the OWASP Top 10.

### 8. Ship iteratively
A working v1 today beats a perfect v2 never.
Every change should leave the codebase in a better state than it was found.

---

## How I Work with AI Agents

- I treat AI agents as senior engineers: I give context, constraints, and expected outcomes
- I review every suggestion critically before applying it
- I prefer agents that explain their reasoning
- When stuck, I use the `/plan` command to structure my thinking before asking for code
- I use the `/review` command before merging anything significant
- I track decisions using `/handover` so context is never lost between sessions

---

## Language-Agnostic Code Standards

These apply regardless of stack:

- **Naming**: Use descriptive names. `user_id` is better than `uid`, `calculate_total_price()` is better than `calc()`
- **Functions**: One function = one responsibility. If it needs a long comment to explain what it does, split it
- **Comments**: Explain WHY, not WHAT. The code explains what — comments explain intent
- **Error handling**: Every external call (API, DB, file system) must handle failure explicitly
- **Tests**: Write tests for the behavior you care about, not the implementation details
- **Configuration**: All environment-specific values go in config files or env vars, never hardcoded
- **Dependencies**: Every new dependency is a liability. Justify it. Prefer established, maintained libraries

---

## Project Structure Conventions

- **README.md** always exists and explains: what, why, how to run, how to contribute
- **`.env.example`** exists for every project with secrets — never commit `.env`
- **`docs/`** folder for design decisions, ADRs, and architecture notes
- Monorepos use `packages/` or `apps/` — flat structures for small projects
- Infrastructure as Code when possible (Docker, Compose, Terraform)

---

## Communication Style (with AI agents)

When assigning tasks to AI:
1. State the goal clearly: "I need X that does Y"
2. Give constraints: "Must use library Z, must be compatible with Node 18"
3. Specify what success looks like: "Tests pass, no TypeScript errors, runs locally"
4. Tell the agent what NOT to do if relevant: "Don't change the database schema"

---

## What belongs in this repo (brain) vs. a project

| In `~/.brain` (global) | In the project |
|---|---|
| Universal coding principles | Project-specific conventions |
| Agent definitions and prompts | Project-specific prompts |
| MCP configurations | Project-specific integrations |
| Memory (cross-project) | Local README / docs |
| Command templates | Project-specific commands |

Never pollute the brain repo with project-specific knowledge.
Never put global reasoning rules inside a project.

## Module: Code Style

### Universal (applies to every language)

**Naming conventions**
- Variables and functions: descriptive, intent-revealing names
- Boolean variables: prefix with `is_`, `has_`, `can_`, `should_`
- Constants: ALL_CAPS_WITH_UNDERSCORES (or SCREAMING_SNAKE_CASE)
- Private members: prefix with `_` where convention allows
- Avoid abbreviations unless universally understood (e.g., `id`, `url`, `api`)

**Function design**
- Max 30 lines per function as a soft limit — if longer, consider extracting
- Single responsibility: one function does one thing
- Pure functions preferred when possible (no side effects, easier to test)
- Functions that can fail should communicate failure explicitly (return error/Result, throw exception — document which)

**File organization**
- Imports/dependencies at the top, grouped: stdlib → external → internal
- Constants and types/interfaces before functions
- Helper functions after the main function that uses them, or in a separate helpers file
- Max 300 lines per file as a soft limit

**Code formatting**
- Always use the project's configured formatter (Prettier, Black, gofmt, etc.)
- If no formatter is configured, ask before assuming a style
- 2 or 4 spaces for indentation (follow existing project convention, never mix)
- Trailing newline at end of every file

**Complexity**
- Cyclomatic complexity per function: aim for ≤ 10
- Nesting depth: ≤ 3 levels. Use early returns to reduce nesting
- Ternary operators: only for simple, readable cases. Never nested ternaries

**Dead code**
- Never leave commented-out code in final commits
- Remove unused imports, unused variables, unused functions
- If something is "for later", open a TODO issue instead of leaving dead code

### Language-specific hints (AI guidance)

When writing code in any language:
1. Follow the idiomatic style of THAT language (e.g., error handling in Go vs. Python vs. Rust)
2. Use the ecosystem's standard tools (npm/pnpm for Node, pip/uv for Python, cargo for Rust)
3. Don't impose patterns from one language into another
4. Ask which pattern to follow if you see multiple valid options in the existing codebase

## Module: Communication

### How I communicate with AI agents

**Context first**: Always open with context before the request.
Bad: "Fix the login bug"
Good: "The login endpoint at POST /auth/login returns 500 when the email contains uppercase letters. Fix it without changing the response structure."

**Be specific about scope**: Clearly state what's in and out of scope.
Bad: "Improve this code"  
Good: "Improve the readability of this function. Don't change the logic or the function signature."

**Ask for reasoning when uncertain**: If I don't understand why an agent chose an approach, I ask.
"Why did you choose approach X over approach Y?"

**Acknowledge mistakes openly**: If I gave wrong context, I correct it immediately.
Don't waste tokens on retries with the same bad context.

### How AI agents should communicate with me

- **Be direct**: Skip preamble. Don't start responses with "Sure!" or "Great question!"
- **Show your work briefly**: When making significant decisions, explain the tradeoff in 1-2 sentences
- **Use lists for steps**: Sequential tasks should be numbered. Options should be bulleted
- **Flag uncertainty**: If you're not sure, say so. Don't hallucinate confidence
- **Ask before assuming**: If a requirement is ambiguous, ask ONE clarifying question before proceeding
- **Format code correctly**: Use proper code blocks with language tags
- **Cite sources when relevant**: If referencing a library or pattern, mention where it's documented

### Response length guidelines
- Simple questions → 1-3 sentences
- Code tasks → Code + brief explanation only
- Architecture/planning → Structured with headers, as long as needed
- Never pad responses. Quality > quantity.

## Module: Git

### Commit messages

Use Conventional Commits format:
```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

**Types**: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert

**Rules**:
- Subject line: max 72 characters, imperative mood ("add feature" not "added feature")
- Body when needed: explain WHY, not WHAT (the diff shows what changed)
- Break lines at 80 characters in the body
- Reference issues: `Closes #123`, `Fixes #456`

**Good examples**:
```
feat(auth): add JWT refresh token rotation

Prevents token replay attacks by invalidating the old refresh token
when a new one is issued. The old token becomes invalid immediately.

Closes #89
```

```
fix(api): handle null user_id in payment endpoint

Caused 500 errors when unauthenticated requests reached the payment
handler. Added early return with 401 response.
```

### Branch strategy

- `main` / `master`: always deployable, protected
- `develop`: integration branch (if using GitFlow)
- Feature branches: `feat/<short-description>` or `feat/<issue-number>-<description>`
- Fix branches: `fix/<issue-number>-<description>`
- Hotfixes: `hotfix/<description>`

### Workflow rules

1. **Never force-push to main/master** — use revert commits instead
2. **Never commit secrets** — use pre-commit hooks or `.gitignore`
3. **Keep commits atomic** — each commit should be one logical change
4. **Review your diff before committing** — `git diff --staged`
5. **Pull before pushing** — always fetch/pull to avoid diverged history
6. **Sign commits** when working on security-sensitive projects

### PR / MR conventions

- Title: same format as commit message
- Description: What changed, why, how to test
- Link to issue/ticket
- Assign reviewers explicitly
- Don't merge your own PRs without review (unless solo project)
- Keep PRs small: aim for < 400 lines changed per PR

### Brain repo specific

When updating `~/.brain/`:
- Commit prefix: `brain: ` (e.g., `brain: add debugging agent`)
- Always run `adapters/generate.sh` after modifying `rules/`
- Commit the generated artifacts alongside the source change
- Never commit environment-specific state (no hardcoded paths, no secrets)

## Module: Security

### Non-negotiable rules (apply always, everywhere)

1. **No hardcoded secrets** — API keys, passwords, tokens, private URLs must always come from environment variables or a secrets manager
2. **`.env` is always in `.gitignore`** — always. No exceptions.
3. **`.env.example` always exists** — with placeholder values, committed to the repo
4. **Input validation** — validate and sanitize ALL inputs from external sources (users, APIs, files, env vars)
5. **Least privilege** — every component, service, and user should have only the permissions it needs

### Secrets management

- Use environment variables for local development
- Use a secrets manager (Vault, AWS Secrets Manager, 1Password Secrets Automation) for production
- Never log secrets — redact before logging
- Rotate secrets regularly, especially after personnel changes
- Use short-lived tokens when possible (JWT with expiry, OAuth refresh tokens)

### Dependency security

- Audit dependencies before adding them: check stars, maintenance status, known vulnerabilities
- Run `npm audit` / `pip audit` / `cargo audit` regularly
- Pin dependency versions in production
- Update dependencies in a dedicated branch, run tests, then merge
- Use Dependabot or Renovate for automated updates

### API security basics

- Always use HTTPS in production
- Rate limit public endpoints
- Authenticate before authorizing
- Return 401 (unauthorized) not 403 (forbidden) when the user isn't logged in
- Never expose internal error messages to end users — log internally, return generic message

### AI-specific security

- Never send real production data to an AI API without scrubbing PII first
- Be aware of prompt injection in user-facing AI features
- Log AI requests and responses for audit purposes (with appropriate retention policy)
- Don't trust AI-generated code blindly — review for security issues before deploying
- Be careful with AI-generated SQL/shell commands — verify before execution

### What to do when you find a vulnerability

1. Document what you found (description, severity, affected component)
2. Don't push a fix directly to main — use a private branch
3. Fix it before disclosing publicly
4. Add a test that would have caught it
5. Update the `SECURITY.md` if the project has one

### OWASP Top 10 awareness

Always keep in mind: Injection, Broken Auth, Sensitive Data Exposure, XXE, Broken Access Control,
Security Misconfiguration, XSS, Insecure Deserialization, Known Vulnerable Components, Insufficient Logging.

When building web-facing features, check each one is addressed.

## Module: Workflow

### The AI Engineering Loop

Every task follows this cycle, no matter the size:

```
Understand → Plan → Delegate → Review → Integrate → Document
```

1. **Understand**: What is the real problem? Who is it for? What does success look like?
2. **Plan**: Break it down. Use `/plan` for anything >30 min of estimated work
3. **Delegate**: Assign to the right agent with full context and constraints
4. **Review**: Never accept AI output without reading it. Use `/review` before merging
5. **Integrate**: Apply changes, run tests, verify behavior
6. **Document**: Update README, add comments, create ADR if architectural decision was made

### Project kickoff checklist

Before writing any code for a new project:
- [ ] Problem clearly defined in 1-2 sentences
- [ ] User/stakeholder identified
- [ ] Success criteria defined (how do we know when it's done?)
- [ ] Tech stack chosen with justification
- [ ] Risks identified (at least top 3)
- [ ] Git repo initialized with README and .gitignore
- [ ] `.env.example` created if secrets will be needed
- [ ] Basic project structure scaffolded

### Session management

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

### Decision making

When facing a decision with multiple valid options:
1. State the options clearly
2. Identify the evaluation criteria (speed, cost, maintainability, complexity)
3. Make a decision
4. Document it — even informally in a comment or ADR

Use the planner agent for architectural decisions.
Use the researcher agent for "what's the best library for X" questions.
Don't spend more than 15 minutes deciding — pick the best option with current information.

### Task sizing

- **< 30 min**: Just do it. No formal plan needed.
- **30 min – 2 hrs**: Write a brief plan comment at the top of the session
- **> 2 hrs**: Use `/plan` to create a structured breakdown before starting
- **Multi-day**: Create a proper spec doc in `docs/`, track progress there

### Debugging workflow

1. Reproduce the bug reliably first
2. Identify the smallest code path that triggers it  
3. Add logging/breakpoints at the boundary of "working" vs. "broken"
4. Hypothesize cause (1-3 candidates)
5. Test hypotheses one at a time
6. Fix root cause, not symptom
7. Write a test that would have caught this bug
8. Commit with a descriptive message explaining what was wrong

Never fix a bug without understanding why it happened.