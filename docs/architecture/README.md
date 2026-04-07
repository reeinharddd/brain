---
type: documentation
id: readme-architecture
title: Architecture Documentation
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: documentation
---

## Architecture Documentation

Welcome to the architecture documentation folder. This is where design decisions, system architecture, and technical decisions are documented.

## What's Here

This folder contains:

- **System Design Documents**: High-level architecture and design of major systems
- **Integration Guides**: How components interact
- **Decision Context**: Background and reasoning for architectural choices
- **API Design**: System interfaces and contracts
- **Deployment Strategies**: How systems are deployed and scaled

## How to Find What You Need

**Looking for a design decision?** → Check ADRs in `/adr/`

**Want to understand system architecture?** → Start here with design documents

**Need implementation details?** → See `/specifications/`

**Want examples?** → Go to `/examples/`

## Creating New Architecture Documents

1. **Use the template**: `templates/design-doc-template.md`
2. **Follow the structure**: Overview → Motivation → Design → Trade-offs → Risks
3. **Add examples**: Include code or diagrams
4. **Link related docs**: Use the `/related:` field in frontmatter
5. **Validate**: Run the validation checklist
6. **Keep production helpers in Go**: any operational script or helper that ships with the product must be implemented as a Go executable; shell/Python examples stay development-only

## Key Design Documents

- [daemon-orchestration.md](daemon-orchestration.md) — How the daemon coordinates all services
- [cli-integration.md](cli-integration.md) — CLI communication with daemon
- [github-operating-model.md](github-operating-model.md) — GitHub-based planning, execution, review, and release flow
- [development-methodology.md](development-methodology.md) — Planning, tracking, quality gates, and case-based working rules
- [environment-configuration.md](environment-configuration.md) — Dev-only and production-safe boundary
- [documentation-enforcement-system.md](documentation-enforcement-system.md) — Documentation validation and enforcement rules

## Related Areas

- Architecture Decisions: See `/adr/`
- Implementing: See `/specifications/`
- Examples: See `/examples/`

---

**Last updated**: 2026-04-03
