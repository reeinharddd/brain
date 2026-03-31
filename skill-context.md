# Project Skill Context

- Project root: .
- Detected tags: bash docs markdown shell

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
