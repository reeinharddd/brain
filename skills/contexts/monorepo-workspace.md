# Skill Context: Monorepo Workspace

- Make ownership boundaries between apps and packages explicit.
- Share libraries intentionally; avoid leaking app-specific concerns into common packages.
- Keep task execution incremental and cache-friendly.
- Validate cross-package API changes with focused tests before broad builds.
- Prefer root-level conventions with package-level escape hatches only when justified.
