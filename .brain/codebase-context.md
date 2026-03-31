# Codebase Context Pack

- Generated at: 2026-03-20T22:17:06Z
- Project root: /home/reeinharrrd/.brain
- Namespace: reeinharddd_brain
- Context index: /home/reeinharrrd/.brain/.brain/codebase-context.ndjson
- Context documents: 73
- Vector status: unavailable
- Vector details:
  urllib.error.URLError: <urlopen error [Errno 111] Connection refused>

# Project Skill Context

- Project root: /home/reeinharrrd/.brain
- Detected tags: bash clean-architecture docker docs markdown shell

Load only the contexts below in addition to the global brain rules.

## Bash Platform

- Skill ID: bash-platform
- Match tags: bash,shell,docker
- Summary: Bash-first automation, portability, idempotent scripting

# Skill Context: Bash Platform

- Prefer POSIX-friendly shell where possible and isolate shell-specific behavior.
- Scripts must be idempotent and safe to re-run.
- Fail loudly with actionable error messages.
- Keep dependencies minimal and explicit.
- When editing config files, preserve user-owned data and comments.

## Markdown Specs

- Skill ID: markdown-specs
- Match tags: markdown,docs
- Summary: Specs, ADRs, and handoff artifacts in plain Markdown

# Skill Context: Markdown Specs

- Architecture and workflow artifacts live in Markdown first.
- Specs should separate goals, non-goals, constraints, decisions, and verification.
- Keep one artifact per phase when possible so handoffs stay easy to diff.
- Prefer simple tables and flat lists over decorative formatting.

## Clean Architecture

- Skill ID: clean-architecture
- Match tags: clean-architecture
- Summary: Ports/adapters, use cases, dependency direction

# Skill Context: Clean Architecture

- Domain rules must not depend on transport, persistence, or frameworks.
- Use cases orchestrate behavior and depend on ports, not adapters.
- Adapters translate external concerns into domain-friendly interfaces.
- Keep dependency direction inward and verify it in reviews.
- Favor thin controllers and repositories that do one job.
