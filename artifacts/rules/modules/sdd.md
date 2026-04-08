## Module: Spec-Driven Development

### Batch 1 foundation rules

The brain repo uses a two-layer context model:

1. Global rules come only from `rules/canonical.md` and `rules/modules/*.md`
2. Project context is injected dynamically from the current repository

Do not hardcode project-specific framework guidance inside global adapters.
Project-specific guidance must be generated on demand from the active repo.

### Dynamic skill injection protocol

Before planning or implementation in any project:

1. Detect the stack with `~/.brain/scripts/detect-stack.sh [project_root]`
2. Render only matching skill contexts with `~/.brain/scripts/render-skill-context.sh [project_root]`
3. Load only the generated skill context for the current project
4. State which stack tags were detected when they materially affect decisions

If no stack-specific skill matches, continue with the global rules only.

### SDD DAG

For substantial work, follow this DAG in order:

1. Explore
2. Propose
3. Spec
4. Design
5. Tasks
6. Implement
7. Verify
8. Archive

Each phase must produce an artifact or explicit handoff note.
Do not skip directly from vague intent to implementation.

### Phase contracts

- Explore -> inputs: user goal, repo state; outputs: constraints, assumptions, detected stack
- Propose -> inputs: exploration notes; outputs: candidate approaches and tradeoffs
- Spec -> inputs: chosen proposal; outputs: acceptance criteria and boundaries
- Design -> inputs: spec; outputs: architecture, flow, interfaces, UX if relevant
- Tasks -> inputs: design; outputs: atomic executable work items
- Implement -> inputs: tasks; outputs: smallest effective code or doc changes
- Verify -> inputs: implementation; outputs: test and validation evidence
- Archive -> inputs: verification results; outputs: docs, handover, memory summary

### Delegate-first behavior

The orchestrator coordinates the DAG and specialist routing.
Specialists should receive:

- goal
- constraints
- relevant files
- phase name
- expected output artifact

Avoid mixing artifacts from different phases in one response unless the task is tiny.
