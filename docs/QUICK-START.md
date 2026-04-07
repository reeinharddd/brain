---
type: guidance
id: quick-start
title: Quick Start - 5 Minute Introduction
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: documentation
---

# Quick Start — 5 Minutes

## Welcome!

You're reading the **Brain repository** documentation. This guide gets you oriented in 5 minutes.

## What Is Brain?

Brain is a unified AI development environment with:
- Central daemon orchestrator (Go)
- CLI interface for commands
- React desktop UI
- Portable skills and agents
- All coordinated by a single source of truth

## Where Are Things?

| I want to... | Go to... |
|-------------|-----------|
| **Understand the system** | [architecture/](architecture/) |
| **See what decisions were made** | [adr/](adr/) |
| **Learn about capabilities** | [skills/](skills/) |
| **Write tests** | [testing/](testing/) |
| **Find examples** | [examples/](examples/) |
| **Check API specs** | [specifications/](specifications/) |
| **Use a template** | [templates/](templates/) |
| **Find old docs** | [archive/](archive/) |

## Most Common Questions

### "How do I propose a change?"

1. Create an ADR (Architecture Decision Record)
2. Use template: `templates/adr-template.md`
3. Put it in: `adr/ADR-NNN-my-decision.md`
4. Structure: Context → Options → Decision → Rationale

→ **See**: [adr/README.md](adr/)

### "How do I add a new skill?"

1. Define input/output contracts
2. Document error cases
3. Provide examples
4. Use template: `templates/skill-template.md`
5. Put it in: `skills/skill-name.md`

→ **See**: [skills/README.md](skills/)

### "How do I write a test?"

1. Use template: (see testing/)
2. Follow Arrange-Act-Assert pattern
3. Test both success and failure
4. Aim for 80%+ coverage

→ **See**: [testing/README.md](testing/)

### "How do I understand the architecture?"

Start here:
1. Read [architecture/daemon-orchestration.md](architecture/daemon-orchestration.md)
2. Read related ADRs in [adr/](adr/)
3. Check examples in [examples/](examples/)

→ **See**: [architecture/README.md](architecture/)

## Key Rules

✅ **ALWAYS**:
- Use YAML frontmatter (type, id, title, status at top)
- Write in English (no Spanish!)
- Include examples
- Check the validation checklist

❌ **NEVER**:
- Leave metadata in prose (use frontmatter)
- Create shell scripts (.sh) in /docs
- Duplicate content (link instead)
- Break existing links

## Document Templates

For every new document:

```bash
# 1. Choose type (adr, skill, design-doc, etc.)
# 2. Copy template
cp templates/[TYPE]-template.md [destination]

# 3. Fill in content
# 4. Validate against checklist
# 5. Done!
```

→ **See**: [templates/README.md](templates/)

## The Golden Rule

**Every document should answer**:
- **What**: What is this document about?
- **Why**: Why does this matter?
- **How**: How do I use this?
- **Examples**: Show me working examples
- **Errors**: What can go wrong?

## If You Get Lost

1. **Check the README** in the folder you're in
2. **Search with grep**: `grep -r "what you need" /docs`
3. **Ask**: File an issue or check related docs

## Next Steps

Pick one:

- **I need to make a decision** → [Create an ADR](adr/README.md)
- **I need to design a feature** → [Use design template](templates//design-doc-template.md)
- **I need to write tests** → [See testing guide](testing/README.md)
- **I need to understand something** → [Browse architecture/](architecture/)
- **I need an example** → [Check examples/](examples/)

---

**You're ready!** Pick one of the links above and start.

Still have questions? Check the full [README.md](README.md) or browse the domain folders.
