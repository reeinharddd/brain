# Brain

Brain is a daemon-centered control plane for AI engineering environments.

It is being aligned around five permanent ideas:

- multiple client surfaces such as `cli`, `desktop`, and future `tui`
- one daemon as the runtime and orchestration source of truth
- one unified artifact system for rules, skills, agents, commands, MCPs, providers, and AI runtimes
- one hierarchical policy and security model across organization, user, workspace, and project scopes
- one projection and sync model across local and cloud state

## Canonical Documentation

- Architecture: [brain-v2-target-architecture.md](/home/reeinharrrd/Work/Personal/brain/docs/architecture/brain-v2-target-architecture.md)
- Documentation index: [INDEX.md](/home/reeinharrrd/Work/Personal/brain/docs/INDEX.md)
- ADR index: [README.md](/home/reeinharrrd/Work/Personal/brain/docs/adr/README.md)

## Current Direction

The repository is moving toward a clean structure built around:

- `apps/`
- `core/`
- `artifacts/`
- `internal/`
- `deploy/`

The implementation is still transitional, but the current source of truth for product direction is `docs/`.
