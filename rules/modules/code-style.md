## Module: Code Style

### Universal (applies to every language)

**Naming conventions**
- Variables and functions: descriptive, intent-revealing names
- Boolean variables: prefix with `is_`, `has_`, `can_`, `should_`
- Constants: ALL_CAPS_WITH_UNDERSCORES (or SCREAMING_SNAKE_CASE)
- Private members: prefix with `_` where convention allows
- Avoid abbreviations unless universally understood (e.g., `id`, `url`, `api`)

**Function design**
- Max 30 lines per function as a soft limit — if longer, consider extracting
- Single responsibility: one function does one thing
- Pure functions preferred when possible (no side effects, easier to test)
- Functions that can fail should communicate failure explicitly (return error/Result, throw exception — document which)

**File organization**
- Imports/dependencies at the top, grouped: stdlib → external → internal
- Constants and types/interfaces before functions
- Helper functions after the main function that uses them, or in a separate helpers file
- Max 300 lines per file as a soft limit

**Code formatting**
- Always use the project's configured formatter (Prettier, Black, gofmt, etc.)
- If no formatter is configured, ask before assuming a style
- 2 or 4 spaces for indentation (follow existing project convention, never mix)
- Trailing newline at end of every file

**Complexity**
- Cyclomatic complexity per function: aim for ≤ 10
- Nesting depth: ≤ 3 levels. Use early returns to reduce nesting
- Ternary operators: only for simple, readable cases. Never nested ternaries

**Dead code**
- Never leave commented-out code in final commits
- Remove unused imports, unused variables, unused functions
- If something is "for later", open a TODO issue instead of leaving dead code

### Language-specific hints (AI guidance)

When writing code in any language:
1. Follow the idiomatic style of THAT language (e.g., error handling in Go vs. Python vs. Rust)
2. Use the ecosystem's standard tools (npm/pnpm for Node, pip/uv for Python, cargo for Rust)
3. Don't impose patterns from one language into another
4. Ask which pattern to follow if you see multiple valid options in the existing codebase
