---
type: design-doc
id: brain-os-visual-system
title: Brain OS Visual System
version: 1.0.0
status: active
date_created: 2026-04-12
language: en
category: architecture
---

## Overview

This document defines the canonical visual system for the Brain desktop surface.
The goal is to make the desktop feel like a technical operating console: precise,
minimal, high-contrast, and traceable.

## Source of Truth

The imported sample bundle lives in `apps/desktop/src/assets/brain-os-samples/`.
The design reference is the provided `brain_kernel/DESIGN.md` artifact and the
associated dark, light, and mobile sample screens.

## Design Principles

- Brutalist editorial presentation
- One primary accent color: violet
- Monospace for technical data, logs, commands, and metadata
- Hard borders instead of shadows or soft glow
- 4pt spacing system only
- Clear hierarchy: title, status, action, content
- One primary action per surface
- Dark mode is the default presentation

## Token Contract

### Colors

- Background primary: `#0A0A0A`
- Background secondary: `#111111`
- Background tertiary: `#171717`
- Border subtle: `#232323`
- Border strong: `#2E2E2E`
- Text primary: `#F5F5F3`
- Text secondary: `#A1A1A1`
- Text muted: `#6B6B6B`
- Accent: `#7C3AED`
- Accent hover: `#8B5CF6`
- Accent active: `#6D28D9`
- Success: `#10B981`
- Warning: `#F59E0B`
- Error: `#EF4444`

### Typography

- Display titles: Sans, semibold, compact tracking
- Technical copy: Monospace
- H1: 28px
- H2: 22px
- H3: 18px
- Body: 14px
- Small: 12px
- Micro: 11px

### Layout

- Desktop sidebar width: 240px
- Main content max width: 1280px
- Top bar height: 64px
- Mobile behavior: sidebar collapses into a drawer

### Shapes and motion

- Default radius: 6px
- Small radius: 4px
- Large radius: 8px
- Motion should be snap-fast or reduced when accessibility requests it

## Screen Taxonomy

The desktop should project the following canonical sections:

1. Runtime
   - System health, execution surface, providers, and command bar
2. Agents
   - Active agent pool, roles, status, and capabilities
3. Memory
   - Timeline or recall view for persistent context objects
4. Rules
   - Canonical rule file editor with validation and versioning
5. MCP Tools
   - Tool servers, connection state, and usage metrics
6. Logs
   - Monospace event stream and live activity
7. Evals
   - Success rate, latency, cost, and session history
8. Samples
   - Design reference gallery for dark, light, and mobile screens
9. Reference
   - Docs search, skills, and support surfaces

## Component Contract

### Shell

- Left sidebar for section navigation
- Top bar for identity, status, and theme toggle
- Main content area with one active section at a time
- Sticky command bar at the bottom of content

### Cards and panels

- Use hard borders
- Prefer dense but readable spacing
- Avoid gradients, blur, and decorative shadows

### Badges and status

- Status badges should be short, uppercase, and easy to scan
- Use color only when it carries meaning
- Pair status color with text whenever possible

### Tables

- Use tables for dense server and evaluation data
- Header rows should be clearly separated
- Hover should be subtle and not playful

### Editors and logs

- Use monospace and line-numbered layouts for rules and logs
- Preserve the feeling of a technical console
- Validation and metadata should sit beside the primary content, not above it

## Accessibility

- All controls must be keyboard reachable
- Contrast must remain WCAG AA or better
- Focus rings must be visible and consistent
- Reduced motion must be respected
- Mobile layouts must collapse cleanly to a single column

## Implementation Guidance

- Use shared design-system primitives in `apps/desktop/src/design-system/`
- Avoid inline styling for new desktop UI work
- Avoid ad hoc palette values when a token exists
- Keep the daemon as the data source; the desktop only renders projections
- Treat the imported sample bundle as reference material, not as throwaway assets

## Related Docs

- `docs/architecture/stack-and-implementation-baseline.md`
- `docs/architecture/capability-control-plane-roadmap.md`
- `docs/INDEX.md`
