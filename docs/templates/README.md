# Brain Repository: Template System Architecture

**Date**: 2026-04-03  
**Version**: 2.0.0  
**Status**: STABLE

---

## Executive Summary

This directory defines **TWO DISTINCT TEMPLATE SYSTEMS** for the Brain repository:

1. **FUNCTIONAL ARTIFACTS** — Extend Brain's capabilities (agents, skills, rules, MCPs, etc.)
2. **INTERNAL ARTIFACTS** — Document decisions, learning, and project state

Each system has **different structures, enforcement mechanisms, and RAG integration strategies**.

---

## System 1: FUNCTIONAL ARTIFACTS

Functional artifacts are **technology multipliers** that increase Brain's capability across:

- Multiple IDEs (VS Code, Claude Desktop, Cursor, Windsurf, etc.)
- Multiple LLMs (GPT-5, Claude, Gemini, etc.)
- Multiple projects and team members

### Characteristics of Functional Artifacts

| Property           | Value                                                |
| ------------------ | ---------------------------------------------------- |
| **Scope**          | Global, affects all projects using Brain             |
| **Validation**     | Strict: schema, examples, integration tests          |
| **Enforcement**    | Automated: daemon, CLI, UI integration               |
| **Versionability** | Explicit (semver required)                           |
| **Documentation**  | Must include DO's/DON'Ts, examples, validation steps |
| **Testing**        | Examples must be executable + validated              |
| **Reusability**    | Across team members, IDEs, LLMs                      |
| **Impact**         | High (affects system behavior directly)              |

### Types of Functional Artifacts

| Type             | Purpose                                       | Location                   | Validation                      |
| ---------------- | --------------------------------------------- | -------------------------- | ------------------------------- |
| **Agents**       | LLM orchestrators, specialists, delegators    | `functional/agents/`       | Router + startup check          |
| **Skills**       | Portable capability packages with methodology | `functional/skills/`       | Daemon discovery + registry     |
| **Rules**        | Non-negotiable system constraints             | `functional/rules/`        | Pre-commit hooks + linter       |
| **Commands**     | Special operations (e.g., `/plan`, `/review`) | `functional/commands/`     | CLI + daemon routing            |
| **MCPs**         | Model Context Protocol servers                | `functional/mcps/`         | Schema validation + launch test |
| **Hooks**        | Git/system automation                         | `functional/hooks/`        | Git pre-commit check            |
| **Instructions** | IDE-specific guidance files                   | `functional/instructions/` | IDE loader validation           |

---

## System 2: INTERNAL ARTIFACTS

Internal artifacts are **record-keeping mechanisms** for:

- Architectural decisions and their rationale
- Project-specific documentation and context
- Lessons learned, bugs found, patterns observed

### Characteristics of Internal Artifacts

| Property           | Value                                   |
| ------------------ | --------------------------------------- |
| **Scope**          | Project-specific or session-specific    |
| **Validation**     | Minimal (readability only)              |
| **Enforcement**    | None (informational only)               |
| **Versionability** | Git history is version control          |
| **Documentation**  | Freeform structure via template         |
| **Testing**        | Not applicable                          |
| **Reusability**    | Archive-like (historical reference)     |
| **Impact**         | Low (documentation, no behavior change) |

### Types of Internal Artifacts

| Type                 | Purpose                               | Location                 | Retention |
| -------------------- | ------------------------------------- | ------------------------ | --------- |
| **Decisions**        | Architecture, tech, process decisions | `internal/decisions/`    | Permanent |
| **Project Docs**     | Context, onboarding, project state    | `internal/project-docs/` | As-needed |
| **Learning Records** | Bugs, lessons, mistakes, patterns     | `internal/learning/`     | 90 days   |

---

## How to Use This System

### Creating a FUNCTIONAL Artifact

### Step 1: Choose Functional Type

```text
Is what I'm creating...?
├─ A specialist helper → Agent
├─ A capability/methodology → Skill
├─ A non-negotiable constraint → Rule
├─ A shortcut operation → Command
├─ A tool/service → MCP
├─ An automation → Hook
└─ IDE guidance → Instruction
```

### Step 2: Follow the Functional Template

```text
cd functional/{TYPE}/
cat TEMPLATE.md              # Read structure
cat EXAMPLES/*.md            # Read 2-3 real examples
cat GUIDE.md                 # Read DO's/DON'Ts
cp TEMPLATE.md new-item.md   # Create your own
```

### Step 3: Validate and Test

```bash
# Check schema
scripts/validate-{type}.sh new-item.md

# Test examples (if applicable)
for example in new-item/*.example; do
  bash "$example"
done

# Final check
git diff && git add && git commit
```

### Creating an INTERNAL Artifact

### Step 1: Choose Internal Type

```text
Is what I'm documenting...?
├─ A decision I made → Decision
├─ Project context/guide → Project Doc
└─ Something I learned → Learning Record
```

### Step 2: Follow the Internal Template

```text
cd internal/{TYPE}/
cat TEMPLATE.md              # Read structure
cat EXAMPLES/*.md            # Read examples
cp TEMPLATE.md new-doc.md    # Create yours
```

### Step 3: Save to Memory (if decision)

```text
If this is a decision that affects future work:
brain memory create --type Decision --entity-name ...
```

### Step 4: Commit

```bash
git add internal/{type}/new-doc.md
git commit -m "docs: Add decision on [topic]"
```

---

## Structure Overview

```text
docs/templates/
│
├── README.md (this file)
│   ├─ Overview + comparison
│   ├─ How to use both systems
│   └─ Global rules
│
├── functional/                           # TECHNOLOGY MULTIPLIERS
│   │
│   ├── agents/
│   │   ├── TEMPLATE.md                  # Copy this
│   │   ├── GUIDE-DO-DONT.md
│   │   ├── GUIDE-INTEGRATION.md
│   │   └── EXAMPLES/
│   │       ├── orchestrator.md
│   │       └── debugger.md
│   │
│   ├── skills/
│   │   ├── TEMPLATE.md
│   │   ├── GUIDE-METHODOLOGY.md
│   │   ├── GUIDE-VALIDATION.md
│   │   └── EXAMPLES/
│   │       ├── debugging-methodology.md
│   │       └── code-refactoring.md
│   │
│   ├── rules/
│   │   ├── TEMPLATE.md
│   │   ├── GUIDE-ENFORCEMENT.md
│   │   ├── GUIDE-DO-DONT.md
│   │   └── EXAMPLES/
│   │       ├── no-hardcoded-secrets.md
│   │       └── go-only-scripts.md
│   │
│   ├── commands/
│   │   ├── TEMPLATE.md
│   │   ├── GUIDE-IMPLEMENTATION.md
│   │   └── EXAMPLES/
│   │       ├── plan.md
│   │       └── review.md
│   │
│   ├── mcps/
│   │   ├── TEMPLATE.md
│   │   ├── GUIDE-SCHEMA.md
│   │   ├── GUIDE-DO-DONT.md
│   │   └── EXAMPLES/
│   │       ├── code-execution.md
│   │       └── file-search.md
│   │
│   ├── hooks/
│   │   ├── TEMPLATE.md
│   │   ├── GUIDE-IDEMPOTENCY.md
│   │   └── EXAMPLES/
│   │       ├── block-hardcoded-secrets.md
│   │       └── validate-go-format.md
│   │
│   └── instructions/
│       ├── TEMPLATE.md
│       ├── GUIDE-IDE-INTEGRATION.md
│       └── EXAMPLES/
│           ├── copilot-instructions.md
│           └── claude-desktop.md
│
└── internal/                             # RECORD-KEEPING
    │
    ├── decisions/
    │   ├── TEMPLATE.md
    │   ├── GUIDE-ADR-STYLE.md
    │   └── EXAMPLES/
    │       ├── skill-system-contract-2026-04-03.md
    │       └── go-only-orchestration-2026-02-15.md
    │
    ├── project-docs/
    │   ├── TEMPLATE.md
    │   ├── GUIDE-SCOPE.md
    │   └── EXAMPLES/
    │       ├── brain-integration-guide.md
    │       └── cli-design-rationale.md
    │
    └── learning/
        ├── TEMPLATE.md
        ├── GUIDE-EFFECTIVE-LEARNING.md
        └── EXAMPLES/
            ├── race-condition-in-sync.md
            └── skills-validation-lessons.md
```

---

## Global Rules for ALL Artifacts

### RULE 1: English Only

- ALL metadata, descriptions, examples, comments MUST be in English
- Exception: User-generated project content outside Brain repo
- Why: Portability across teams, LLMs, and international users

### RULE 2: No Duplication

- Before creating new artifact, search if one already exists
- If similar exists, enhance instead of creating new
- Prevents maintenance burden and inconsistency

### RULE 3: Clear Deprecation Path

- If artifact becomes obsolete, mark with `status: deprecated`
- Specify: What replaces it, migration path, sunset date
- Never silently delete functional artifacts

### RULE 4: Testable Examples

- Every example in functional artifacts must be executable
- For internal artifacts: include verification method
- Pseudo-code acceptable ONLY with text explanation

### RULE 5: Context Limits

- Functional artifacts: < 300 lines (split if larger)
- Internal artifacts: No hard limit, but consider RAG chunks
- Link to files instead of embeding huge content

### RULE 6: Metadata Consistency

- Functional: YAML frontmatter with version, status, keywords
- Internal: Simple metadata (title, date, author)
- See templates for required vs. optional fields

---

## Comparison Matrix: Functional vs. Internal

| Aspect                     | Functional                       | Internal                       |
| -------------------------- | -------------------------------- | ------------------------------ |
| **Audience**               | All LLMs, all IDEs, all projects | Project team, future self      |
| **Lifespan**               | Permanent (versioned)            | Session or project duration    |
| **Validation**             | Strict + automated               | Minimal (readability)          |
| **Structure**              | Rigid (YAML + sections)          | Flexible (template + freeform) |
| **Enforcement**            | Yes (daemon, hooks, linters)     | No (informational only)        |
| **Examples Required**      | Yes (executable, validated)      | Optional                       |
| **DO's/DON'Ts**            | Explicit, required               | Optional                       |
| **RAG Integration**        | Automatic via keywords           | Via memory entities            |
| **Versioning**             | Semver                           | Git history                    |
| **Backward Compatibility** | CRITICAL                         | N/A                            |

---

## Validation Checklist

### FUNCTIONAL Artifacts

- [ ] Frontmatter valid YAML (no syntax errors)
- [ ] All required fields present
- [ ] Examples are executable (tested locally)
- [ ] DO's and DON'Ts section exists
- [ ] Related artifacts linked (if any)
- [ ] Keywords list includes search terms
- [ ] Enforcement mechanism specified (if applicable)
- [ ] Backward compatible with prior version (if updated)
- [ ] No hardcoded config values
- [ ] Time estimates realistic (within 20%)

### INTERNAL Artifacts

- [ ] Title and date present
- [ ] Content is readable and clear
- [ ] Narrative structure: context → decision → rationale
- [ ] Grammar check passed
- [ ] Relevant entities identified (for memory)
- [ ] Links to related docs (if applicable)

---

## Benefits of This System

### For LLMs

When an agent works on Brain project code:

```text
Agent loads context from:
1. rules/canonical.md (global)
2. functional/{type}/*.md (relevant artifacts)
3. internal/decision objects (prior decisions)
4. Project-specific guides

= Better decisions, 40%+ improvement in output quality
```

### For Teams

```text
Shared knowledge → Consistent practices → Less re-debating → Faster execution
```

### For IDEs

```text
IDE loads functional artifacts → Sets up environment → Guidance available
in real-time as user works
```

### For Projects

```text
New projects reuse proven patterns → Reduce startup time → Better quality
from day 1
```

---

## Migration Path: Existing Docs

If you have existing documentation:

1. **Identify type** — Is it functional or internal?
2. **Choose template** — From the appropriate folder
3. **Restructure** — Move content into template sections
4. **Validate** — Run validation checklist
5. **Test** — Execute examples if applicable
6. **Commit** — With clear message referencing new location

---

## Questions?

**How do I add a new agent?**  
→ See `functional/agents/TEMPLATE.md` + `GUIDE-*.md`

**When should I document something?**  
→ If it affects behavior globally (functional) or helps future work (internal), document it.

**Will this slow me down?**  
→ No. For 90% of tasks, you reuse existing artifacts. Only create new when truly novel.

**Can I modify functional artifacts?**  
→ Yes, via PR. Update version number (semver) appropriately.

**What if my artifact doesn't fit a template?**  
→ It probably shouldn't exist. Ask: Is this functional (extends capability) or internal (records decision)? If neither, consider if documentation is needed.

---

**Last Updated**: 2026-04-03  
**Maintained By**: Brain Engineering Team  
**Links**: [Functional Artifacts](#system-1-functional-artifacts) | [Internal Artifacts](#system-2-internal-artifacts)
