---
name: desktop-visual-system
version: "1.0.0"
status: stable
ide: copilot
applyTo: "apps/desktop/**/*"
description: "Use when editing Brain desktop UI files: apply the canonical visual system, shared design-system primitives, and no inline styling."
deprecated: false
keywords:
  - desktop
  - visual system
  - design system
  - runtime
  - Tauri
applies_to_languages:
  - typescript
  - tsx
  - css
---

# Brain Desktop Visual System

## When this applies

Use this guidance for any file under `apps/desktop/`.

## Required behaviors

- Use `docs/architecture/brain-os-visual-system.md` as the source of truth.
- Prefer shared primitives from `apps/desktop/src/design-system/`.
- Keep the daemon as the source of truth for data.
- Use the canonical 4pt spacing system and the defined color tokens.
- Keep the desktop feel technical, high-contrast, and restrained.

## Do

- Build with cards, panels, badges, tables, code blocks, and command bars.
- Keep section titles short and status text easy to scan.
- Use monospace for logs, commands, metadata, and editor-like surfaces.
- Make keyboard focus visible.
- Support dark, light, and mobile variants.

## Do not

- Do not add inline `style=` blocks for new desktop UI work.
- Do not introduce arbitrary colors, spacing, or shadows when a token exists.
- Do not make the UI look decorative or glossy.
- Do not duplicate visual rules in every component.
- Do not move canonical state into the desktop client.

## Escape hatch

If a one-off exception is unavoidable, keep it local, document the reason in a comment, and prefer adding a design-system primitive instead of repeating the exception.
