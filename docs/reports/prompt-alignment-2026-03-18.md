# Prompt Alignment Report

Date: 2026-03-18
Repo: `~/.brain`

## Goal

Combine the two implementation prompts without duplicating systems or losing the
repo's Bash-and-Markdown-first identity.

## Adopted from Prompt 1

- single source of truth in `rules/canonical.md` plus `rules/modules/*.md`
- delegate-first SDD as the primary operating model
- dynamic stack detection and skill-context injection
- shift-left local validation with Guardian
- portable "second brain" framing across IDEs and adapters
- runtime resolution for `npx`/`npx-nvm` and `HOME`-based paths

## Adopted from Prompt 2

- explicit `architect` and `implementer` agent roles
- explicit `memory-protocol.md` module
- compatibility wrapper `scripts/guardian.sh`
- hook installer `scripts/install-hooks.sh`
- provider example `providers/guardian.env.example`
- stronger handover expectations around memory summary and DAG state

## Adopted with adaptation

- Artifact-driven SDD:
  use Markdown artifacts when phase boundaries matter, but do not force a full
  `.specs/` tree for every small task.
- Guardian provider model:
  keep deterministic local checks as the default enforcement path; treat provider
  configuration as optional future enrichment rather than mandatory runtime.
- Memory protocol:
  keep progressive disclosure and namespaces, but avoid adding tool-specific
  complexity where the current environment already validates memory behavior.

## Deferred or rejected

- hard stop on every batch until human approval
- mandatory LLM review in Guardian
- exact enterprise file explosion for every stack and phase
- replacing working scripts only to match prompt names
- forcing all work through `.specs/STATUS.md` even for tiny tasks

## Resulting repo stance

- canonical rules remain primary
- prompt-specific files are added only when they improve interoperability
- wrappers and docs may match prompt names, but operational behavior stays rooted
  in existing scripts and checks
