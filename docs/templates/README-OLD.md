# Brain Repository: Template System Architecture

**Date**: 2026-04-03  
**Version**: 1.0.0  
**Status**: STABLE  

---

## Executive Summary

This directory defines **TWO DISTINCT TEMPLATE SYSTEMS** for the Brain repository:

1. **FUNCTIONAL ARTIFACTS** — Extend Brain's capabilities (agents, skills, rules, MCPs, etc.)
2. **INTERNAL ARTIFACTS** — Document decisions, learning, and project state

Each system has **different structures, enforcement mechanisms, and RAG integration strategies**.
   - When to use: Documenting persistent system state
   - Structure: State → Transitions → Fields → Lifecycle

### Educational Types

9. **example-template.md** — Examples & Tutorials
   - When to use: Teaching how to use a system/API
   - Structure: Scenario → Prerequisites → Steps → Output → Variations

10. **guidance-template.md** — Best Practices Guides
    - When to use: Documenting practices and patterns
    - Structure: Principles → DO/DON'T → When to use → Pitfalls

## How to Use a Template

### Quick Start

```bash
# Copy the template
cp templates/adr-template.md ../adr/ADR-NNN-slug.md

# Edit it
edit ../adr/ADR-NNN-slug.md

# Replace placeholders [like this]
# Fill in real content
# Save and validate
```

### Universal Workflow

1. **Copy**: Pick appropriate template, copy to destination
2. **Replace frontmatter**: Update type, id, title, date_created
3. **Fill content**: Replace all [placeholders] with real content
4. **Add examples**: Include 2-3 concrete examples
5. **Validate**: Run checklist from DOCUMENTATION-UNIFIED-SCHEMA.md
6. **Reference**: Update related: links in other docs

## Template Metadata

All templates include YAML frontmatter with:

```yaml
type: [document-type]
id: [unique-id]
title: [Human-readable]
version: 1.0.0
status: active|draft
date_created: YYYY-MM-DD
language: en
category: [domain]
```

## Customization

Templates are starting points. You can:

✅ **DO**:
- Remove unused sections
- Reorder sections for clarity
- Add custom sections as needed
- Include domain-specific content

❌ **DON'T**:
- Skip required fields in frontmatter
- Mix languages
- Remove examples
- Remove error documentation (for specs)

---

**Last updated**: 2026-04-03  
**Templates**: 10 copy-paste ready  
**Coverage**: All document types
