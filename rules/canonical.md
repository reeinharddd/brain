# Canonical Rules - reeinharrrd's Brain Repo

> Version: 1.1.0 | Last updated: 2026-03-17
> This is the single source of truth. All agent-specific rule files are generated from here.
> DO NOT edit adapter files directly. Edit this file, then run: `~/.brain/adapters/generate.sh`

---

## Philosophy

I am an AI engineer. My value is not in knowing every syntax of every language -
it is in knowing how to **model problems clearly**, **delegate intelligently**, and **produce
high-quality outcomes using AI as a collaborator**, not just a code generator.

A programming language is a tool. The real craft is in:

- Understanding what needs to be built and why
- Decomposing problems into small, verifiable units
- Knowing when to build vs. when to reuse
- Making good architectural decisions that last

### Plain Text Only

To ensure longevity and compatibility across all tools, agents, and shells, all documentation, rules, and code comments MUST use only plain text (ASCII/UTF-8).

- **NO Emojis**: Never use emojis (e.g., 😀, 🚀). Use descriptive words instead.
- **NO Decorative Symbols**: Avoid checkmarks (✓), crossmarks (❌), or arrows (→).
- **Standard ASCII Symbols**: Use `-`, `*`, `->`, `+`, `|`, and standard brackets.

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

Use AI to accelerate well-reasoned decisions - not to skip the reasoning.
Always verify AI-generated code before using it in production contexts.
AI can be wrong. You are responsible for what you ship.

### 5. Everything is documented or it doesn't exist

If a decision was made and it isn't written down, it will be forgotten and re-debated.
Document the WHY, not just the WHAT.

### 6. Fail loudly, recover gracefully

Prefer explicit errors over silent failures. Log meaningful messages.
Always handle edge cases - null values, network failures, empty states.

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

### Definition of Done (DoD)

A task is NOT finished until:

1. Code follows the naming and style conventions in `rules/modules/code-style.md`.
2. All unit tests pass and new behavior is verified.
3. No new lint or security warnings are introduced.
4. Documentation (README, ADRs, comments) is updated.
5. All relevant context is saved to memory.

### Error Handling Patterns

1. **Explicit Return**: In languages like Go or Rust, handle every error locally.
2. **Result Object**: Prefer returning a `{ data, error }` object over throwing exceptions for business logic.
3. **No Silent Fails**: Never leave an empty catch block or log-only error if it blocks the flow.
4. **Contextual Errors**: Wrap errors with context (e.g., "Failed to load user: [Original Error]").

---

## Project Structure Conventions

- **README.md** always exists and explains: what, why, how to run, how to contribute
- **`.env.example`** exists for every project with secrets - never commit `.env`
- **`docs/`** folder for design decisions, ADRs, and architecture notes
- Monorepos use `packages/` or `apps/` - flat structures for small projects
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
| :---------------------- | :------------- |
| Universal coding principles | Project-specific conventions |
| Agent definitions and prompts | Project-specific prompts |
| MCP configurations | Project-specific integrations |
| Memory (cross-project) | Local README / docs |
| Command templates | Project-specific commands |

Never pollute the brain repo with project-specific knowledge.
Never put global reasoning rules inside a project.
