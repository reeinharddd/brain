# Skills Quick Reference

## Add a Skill

1. Update the registry through the Brain workflow.
2. Create the skill directory and `SKILL.md` definition.
3. Run `brain skills validate`.
4. Commit only after the daemon reports the registry and filesystem are in sync.

## Remove a Skill

1. Remove the registry entry first.
2. Remove the skill directory second.
3. Run `brain skills validate`.
4. Commit only after the daemon reports no orphans.

## Validate Skills

- Use `brain skills validate` for the current sync state.
- Use the UI if you want a live view without a terminal.
- The daemon remains the source of truth for validation.

## Rules of Thumb

- Registry first, filesystem second.
- No orphan skills.
- Production-facing automation stays in Go.
- Shell helpers are development-only if they exist at all.

## Common Mistakes

- Creating a directory before registering it.
- Removing a registry entry without deleting the matching directory.
- Shipping a helper as a shell script when a Go executable should exist instead.

## Where to Learn More

- `SKILLS-ORCHESTRATION-RULES.md`
- `SKILLS-ENFORCEMENT-STRATEGY.md`
- `SKILLS-ARCHITECTURE-GO-FIRST.md`
