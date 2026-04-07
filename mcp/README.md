# MCP layout

This folder contains the Brain MCP source-of-truth files, runtime launchers, and
compatibility configs used by the editor integrations.

## Canonical files

- `registry.yml` - primary catalog of MCP servers, profiles, and required env
  variables.
- `brain-mcp-server/` - custom Brain MCP server that exposes repo internals.
- `mcp-startup.sh` - runtime launcher used by `brain start`.
- `ports.yml` - port map for the shared MCP gateway.

## Generated or synced configs

- `global-config.json` - IDE-facing MCP config in the legacy command format.
- `global-config-stdio.json` - IDE-facing MCP config that prefers stdio.
- `gateway/mcp-registry.json` - gateway registry used by the shared launcher.
- `profiles/*.json` - ready-to-use MCP profiles for different IDE setups.

## Support files

- `docker_registry.yml` - reference list of Docker-based MCP images.
- `troubleshooting.md` - recovery and diagnostic notes.

## Current runtime behavior

- `brain start` launches the shared MCP ports defined in `mcp-startup.sh`.
- GitHub MCP starts only when `GITHUB_TOKEN` is available.
- `brain-rules` is included in the standard and full MCP profiles.

## Maintenance

- Keep `registry.yml`, `mcp-startup.sh`, and `profiles/*.json` aligned.
- Prefer `brain sync-mcp` for IDE config synchronization.
- Prefer `brain generate` for adapter refreshes and `brain status` for runtime
  verification.
