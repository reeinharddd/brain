\n## Module: Git

\n### Commit messages

Use Conventional Commits format:
```text
<type>(<scope>): <short description>

[optional body]

[optional footer]
```text

**Types**: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert

**Rules**:
- Subject line: max 72 characters, imperative mood ("add feature" not "added feature")
- Body when needed: explain WHY, not WHAT (the diff shows what changed)
- Break lines at 80 characters in the body
- Reference issues: `Closes #123`, `Fixes #456`

**Good examples**:
```text
feat(auth): add JWT refresh token rotation

Prevents token replay attacks by invalidating the old refresh token
when a new one is issued. The old token becomes invalid immediately.

Closes #89
```text

```text
fix(api): handle null user_id in payment endpoint

Caused 500 errors when unauthenticated requests reached the payment
handler. Added early return with 401 response.
```text

\n## Branch strategy

- `main` / `master`: always deployable, protected
- `develop`: integration branch (if using GitFlow)
- Feature branches: `feat/<short-description>` or `feat/<issue-number>-<description>`
- Fix branches: `fix/<issue-number>-<description>`
- Hotfixes: `hotfix/<description>`

\n## Workflow rules

1. **Never force-push to main/master** - use revert commits instead
2. **Never commit secrets** - use pre-commit hooks or `.gitignore`
3. **Keep commits atomic** - each commit should be one logical change
4. **Review your diff before committing** - `git diff --staged`
5. **Pull before pushing** - always fetch/pull to avoid diverged history
6. **Sign commits** when working on security-sensitive projects

\n## PR / MR conventions

- Title: same format as commit message
- Description: What changed, why, how to test
- Link to issue/ticket
- Assign reviewers explicitly
- Don't merge your own PRs without review (unless solo project)
- Keep PRs small: aim for < 400 lines changed per PR

\n## Brain repo specific

When updating `~/.brain/`:


- Commit prefix: `brain: ` (e.g., `brain: add debugging agent`)
- Always run `adapters/generate.sh` after modifying `rules/`
- Commit the generated artifacts alongside the source change
- Never commit environment-specific state (no hardcoded paths, no secrets)
