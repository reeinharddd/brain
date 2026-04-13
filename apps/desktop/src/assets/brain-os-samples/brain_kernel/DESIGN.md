# Design System: Technical Brutalism & Editorial Precision

## 1. Overview & Creative North Star
**Creative North Star: "The Infrastructure of Intelligence"**

This design system moves beyond the "friendly" AI tropes of soft bubbles and colorful gradients. Instead, it treats the AI development OS as a high-performance terminal—a piece of industrial equipment for the mind. It is a **Brutalist Editorial** system: brutal in its honesty, editorial in its clarity. 

The aesthetic is driven by the tension between "The Architect" (Geometric Sans titles) and "The Engineer" (Monospace technical data). We break the traditional UI template by using extreme high contrast, intentional asymmetry, and a rigorous 4pt grid that feels less like a website and more like a blueprint.

---

## 2. Colors & Surface Logic
We define depth not through light and shadow, but through **Tonal Isolation**. In this system, light does not exist; only information density.

### The Palette
*   **Surface Lowest (#0A0A0A):** The void. Used for the primary canvas and background.
*   **Surface Low (#111111):** Secondary containers, sidebar foundations.
*   **Surface High (#171717):** Elevated modules, active states, or "pop-over" terminal windows.
*   **Accent (#7C3AED):** The "Pulse." Used exclusively for primary actions, active code execution, and critical path highlights.
*   **On-Surface (#F5F5F3):** Pure data. High legibility for primary content.
*   **On-Surface Variant (#A1A1A1):** Metadata, comments, and inactive technical strings.

### The "No-Line" Rule
While the base spec allows for a 1px border, it must be used sparingly for **containment, not sectioning**. Large architectural regions (Sidebar vs. Main Editor) should be separated by a shift from `#111111` to `#0A0A0A`. Do not use lines to separate the header from the body; use the spatial 4pt system to create a clear "break" in the data flow.

### Surface Hierarchy & Nesting
Treat the UI as a series of nested machine parts.
1.  **Chassis:** `surface_container_lowest` (#0A0A0A)
2.  **Module:** `surface_container_low` (#111111)
3.  **Component:** `surface_container_high` (#201F1F)
Each inner container must be visually "heavier" or "lighter" than its parent to signify nesting without the need for shadows.

---

## 3. Typography
The system uses a dual-type approach to distinguish between "Human Intent" and "Machine Execution."

*   **Display & Headlines (Space Grotesk):** Set with tight tracking (-2%). These are the high-level commands and system states. Use `headline-lg` (2rem) for major OS modules.
*   **The Technical Core (Monospace):** Everything else—labels, buttons, logs, and inputs—is rendered in Monospace. This reinforces the "OS" feel.
*   **Hierarchy via Weight:** In the Monospace scale, use `Medium` weight for `label-md` to ensure high contrast against dark backgrounds. Avoid `Light` weights which break under the brutalist high-contrast requirement.

---

## 4. Elevation & Depth (The Layering Principle)
We reject "Soft Shadows" and "Glassmorphism." This is a solid-state system.

*   **Tonal Layering:** Hierarchy is achieved by "stacking." A card (`surface_container_high`) sitting on the main background (`surface_container_lowest`) provides all the visual elevation required. 
*   **Hard Borders:** Use `outline_variant` (#4A4455) at 100% opacity for borders. No blurs. The 6px radius (`md`) provides a slight mechanical "milling" to the edges, preventing the UI from feeling sharp and hostile while maintaining its precision.
*   **The "Active" Glow:** Instead of a shadow, an active element (like a focused input) uses a 1px solid `primary` (#7C3AED) border. This is the only "glow" permitted—the glow of a powered-on circuit.

---

## 5. Components

### Buttons
*   **Primary:** Background: `primary_container` (#7C3AED), Text: `on_primary` (#3F008E). Monospace Bold. 6px radius.
*   **Ghost (Secondary):** Background: Transparent, Border: 1px `outline` (#958DA1), Text: `on_surface`.
*   **Interaction:** On hover, the primary button should shift to a hard `primary` (#D2BBFF) with no transition time. Instant feedback is a core OS principle.

### Input Fields
*   **Structure:** Rectangular, `surface_container_low` background, 1px `outline_variant` border.
*   **Typography:** All user input must be Monospace. Labels sit *above* the field, never inside as placeholders. This is an OS, not a marketing landing page.

### Cards & Lists
*   **No Dividers:** Forbid the use of 1px lines between list items. Use 8px or 12px of vertical white space. To separate logical groups, use a background color shift (e.g., a `surface_container_highest` header row).
*   **Data Density:** Lists should be compact. Use `body-sm` (0.75rem) Monospace for list items to maximize information density.

### Chips (Tags)
*   **Technical Labels:** Small, `surface_container_highest` background, no border. Text: `on_surface_variant`. Use these for language tags (e.g., `PYTHON`, `RUST`).

---

## 6. Do's and Don'ts

### Do
*   **Embrace Monospace:** Use it for everything that isn't a major page title.
*   **High Contrast:** Ensure text is either `#F5F5F3` or `#A1A1A1`. Never use mid-greys that wash out against the black.
*   **Rigid Spacing:** Stick to the 4pt system. Every margin and padding should be a multiple of 4 (4, 8, 16, 24, 32...).
*   **Asymmetry:** Align titles to the left but allow technical data to be right-aligned or offset to create an "engineered" editorial look.

### Don't
*   **No Softness:** No blur, no shadows, no gradients. If a component needs to stand out, give it a border or a brighter background color.
*   **No Animation Tweens:** Use "Instant" or "Step" transitions (0ms or 50ms). Elements should "snap" into existence, mimicking terminal performance.
*   **No Center Alignment:** This is a technical tool. Content should be anchored to the grid, usually left-aligned, to facilitate fast scanning.