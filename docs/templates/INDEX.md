# Brain Templates: Quick Index

---

## 🎯 Quick Navigation

| I want to create...       | Type        | Go to                                                                        |
| ------------------------- | ----------- | ---------------------------------------------------------------------------- |
| **An LLM helper**         | Agent       | [`functional/agents/TEMPLATE.md`](functional/agents/TEMPLATE.md)             |
| **A proven methodology**  | Skill       | [`functional/skills/TEMPLATE.md`](functional/skills/TEMPLATE.md)             |
| **A system constraint**   | Rule        | [`functional/rules/TEMPLATE.md`](functional/rules/TEMPLATE.md)               |
| **A shortcut command**    | Command     | [`functional/commands/TEMPLATE.md`](functional/commands/TEMPLATE.md)         |
| **A tool/service**        | MCP         | [`functional/mcps/TEMPLATE.md`](functional/mcps/TEMPLATE.md)                 |
| **An automation**         | Hook        | [`functional/hooks/TEMPLATE.md`](functional/hooks/TEMPLATE.md)               |
| **IDE guidance**          | Instruction | [`functional/instructions/TEMPLATE.md`](functional/instructions/TEMPLATE.md) |
| **A decision**            | Decision    | [`internal/decisions/TEMPLATE.md`](internal/decisions/TEMPLATE.md)           |
| **Project documentation** | Project Doc | [`internal/project-docs/TEMPLATE.md`](internal/project-docs/TEMPLATE.md)     |
| **A lesson learned**      | Learning    | [`internal/learning/TEMPLATE.md`](internal/learning/TEMPLATE.md)             |

---

## 📚 Full Documentation

### Main Reference

- **[README.md](README.md)** — Architecture overview + global rules

### FUNCTIONAL Artifacts (Extend Brain's Capability)

**Agents** — LLM orchestrators & specialists

- 📄 [TEMPLATE.md](functional/agents/TEMPLATE.md) — Copy this to create
- ✅ [GUIDE-DO-DONT.md](functional/agents/GUIDE-DO-DONT.md) — Best practices
- 💡 [EXAMPLES/](functional/agents/EXAMPLES/) — Real agent definitions

**Skills** — Portable methodologies & procedures

- 📄 [TEMPLATE.md](functional/skills/TEMPLATE.md) — Copy this
- ✅ [GUIDE-DO-DONT.md](functional/skills/GUIDE-DO-DONT.md) — DO's/DON'Ts
- 💡 [EXAMPLES/](functional/skills/EXAMPLES/) — Real skills

**Rules** — Non-negotiable constraints

- 📄 [TEMPLATE.md](functional/rules/TEMPLATE.md) — Copy this
- ✅ [GUIDE-DO-DONT.md](functional/rules/GUIDE-DO-DONT.md) — DO's/DON'Ts
- 💡 [EXAMPLES/](functional/rules/EXAMPLES/) — Real rules (security, code style, etc)

**Commands** — Special operations

- 📄 [TEMPLATE.md](functional/commands/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](functional/commands/EXAMPLES/) — `/plan`, `/review`, etc

**MCPs** — Model Context Protocol servers

- 📄 [TEMPLATE.md](functional/mcps/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](functional/mcps/EXAMPLES/) — Real MCPs

**Hooks** — Git & system automation

- 📄 [TEMPLATE.md](functional/hooks/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](functional/hooks/EXAMPLES/) — Real hooks

**Instructions** — IDE-specific guidance

- 📄 [TEMPLATE.md](functional/instructions/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](functional/instructions/EXAMPLES/) — Real instruction files

### INTERNAL Artifacts (Record-Keeping)

**Decisions** — Architecture & technology choices

- 📄 [TEMPLATE.md](internal/decisions/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](internal/decisions/EXAMPLES/) — Real decisions

**Project Docs** — Guides, context, onboarding

- 📄 [TEMPLATE.md](internal/project-docs/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](internal/project-docs/EXAMPLES/) — Real guides

**Learning Records** — Lessons, bugs, patterns

- 📄 [TEMPLATE.md](internal/learning/TEMPLATE.md) — Copy this
- 💡 [EXAMPLES/](internal/learning/EXAMPLES/) — Real learnings

---

## 🔍 By Use Case

### "I'm building something new"

1. Are you extending Brain's capability? → **FUNCTIONAL** artifact
2. Are you improving how Brain works? → Choose type (Agent/Skill/Rule/etc)
3. Copy appropriate template from above
4. Read the GUIDE-DO-DONT.md for that type
5. Create your artifact
6. Run validation script

### "I'm documenting a decision"

1. Go to [`internal/decisions/TEMPLATE.md`](internal/decisions/TEMPLATE.md)
2. Fill in: Problem → Options → Decision → Rationale
3. Save to memory (if affects future work)
4. Commit to git

### "I'm sharing what I learned"

1. Go to [`internal/learning/TEMPLATE.md`](internal/learning/TEMPLATE.md)
2. Fill in: What happened → Root cause → How to prevent
3. Link to related rules/skills if applicable
4. Commit to git

---

## 📋 Validation Checklist

### FUNCTIONAL Artifacts

- [ ] Frontmatter YAML valid
- [ ] All required fields filled
- [ ] Examples tested locally
- [ ] DO's/DON'Ts present
- [ ] Related artifacts linked
- [ ] Keywords/search terms included
- [ ] Time estimates realistic
- [ ] No hardcoded values

### INTERNAL Artifacts

- [ ] Title and date present
- [ ] Content is clear and readable
- [ ] Structure follows template
- [ ] Links to related docs (if applicable)

---

## 🚀 Getting Started

**New to this system?** Follow these 3 steps:

1. **Read** `README.md` (overview + global rules) — 10 min
2. **Browse** EXAMPLES in your artifact type — 5 min
3. **Copy** TEMPLATE.md and fill in placeholder — 5-30 min
4. **Validate** using checklist above — 5 min

**Total**: 25-50 minutes from start to ready-to-use artifact.

---

## ❓ Questions?

**Q: What's the difference between functional and internal?**
A: See [README.md → Comparison Matrix](README.md#comparison-matrix-functional-vs-internal)

**Q: How do I ensure my template is good?**
A: Use the GUIDE-DO-DONT.md for your artifact type

**Q: Can I modify templates?**
A: Yes, but changes should apply globally. If project-specific, create your own.

**Q: Where are example artifacts?**
A: In `EXAMPLES/` subdirectory within each type folder

**Q: How often should I review this?**
A: Check when creating new artifact type. Otherwise, just follow the template.

---

## 📊 System Statistics

| Artifact Type | Templates | Guides | Examples  |
| ------------- | --------- | ------ | --------- |
| Agents        | 1         | 1      | 2-3       |
| Skills        | 1         | 1      | 2-3       |
| Rules         | 1         | 1      | 2-3       |
| Commands      | 1         | -      | 2-3       |
| MCPs          | 1         | -      | 2-3       |
| Hooks         | 1         | -      | 2-3       |
| Instructions  | 1         | -      | 2-3       |
| Decisions     | 1         | -      | 2-3       |
| Project Docs  | 1         | -      | 2-3       |
| Learning      | 1         | -      | 2-3       |
| **TOTAL**     | **10**    | **3**  | **20-30** |

---

**Last Updated**: 2026-04-03  
**Version**: 2.0.0 (Functional/Internal Split)  
**Status**: STABLE & READY FOR USE

Start with [README.md](README.md) if you want the full architecture.  
Start here if you just want to create something.
