<!-- markdownlint-disable-file -->

---

artifact_type: project-docs
version: 2.0.0
internal: true

---

# Guide: Project Docs Artifacts (DO's & DON'Ts)

## What is a Project Doc?

**Project Docs** are permanent project reference materials (onboarding, runbooks, architecture).

Examples: Onboarding guide, Development workflow, Troubleshooting runbook

---

## ✅ DO's

1. **Audience is clear** — "For new developers", "For operations", "For security audit"
2. **Structure is scannable** — Headers, bullet points, tables (not walls of prose)
3. **Include links** — Point to templates, tools, related docs (not self-contained)
4. **Keep updated** — If outdated, explicitly say "OUTDATED: See X instead" at top
5. **Add quick reference** — 5-10 min quick-start at the top, deep content below
6. **Link from README** — Discoverable from main project README

---

## ❌ DON'Ts

1. **Don't repeat documentation** — Link to existing docs instead of copying
2. **Don't assume prior knowledge** — New dev reads this; explain concepts from first principles
3. **Don't be too theoretical** — Include actual commands/examples user can run
4. **Don't write prose walls** — Use headers, bullet points, tables for scannability
5. **Don't hide crucial info** — "FYI the database is at..." buried in paragraph 3 = bad
6. **Don't assume tools are installed** — Link to install instructions or provide them

---

## Common Mistakes

| Mistake                                    | Why Bad                                           | Fix                                        |
| ------------------------------------------ | ------------------------------------------------- | ------------------------------------------ |
| **"For developers" but assumes knowledge** | New dev reads, can't follow                       | Add quick glossary or links                |
| **Instructions are 3 paragraphs**          | Dev skims, misses critical detail                 | Use sections + checklists                  |
| **Outdated info**                          | Dev follows old instructions, hits errors         | Date-check quarterly, mark OUTDATED at top |
| **No links**                               | Doc is self-contained, hard to navigate           | Link to related docs, templates, tools     |
| **No quick-start**                         | Dev spends 30 min reading before running anything | Add 10-line "Get started in 5 min" at top  |

---

## Template Checklist

- [ ] Audience identified at top
- [ ] Quick-start section (first 10-20 lines)
- [ ] Scannable structure (headers, bullets, tables)
- [ ] Real commands (not pseudocode)
- [ ] Links to related docs/templates
- [ ] Date of last update
- [ ] Links FROM README.md (discoverable)

---

## Examples to Reference

- onboarding — New developer guide
- troubleshooting — Common errors + solutions
- workflow — Development process walkthrough

Location: `docs/templates/internal/project-docs/EXAMPLES/`

---

**Created**: 2026-04-03  
**Status**: Stable
