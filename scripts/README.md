# Scripts Directory

This directory contains all executable scripts for the Brain repo. Scripts are organized by purpose.

## Critical Scripts (use directly)

| Script | Purpose | Usage |
|--------|---------|-------|
| `brain-cli.sh` | Central service orchestrator | `brain-cli.sh [start\|stop\|status\|health]` |
| `init.sh` | Per-project initialization | `bash init.sh` (in project directory) |
| `doctor.sh` | Full integrity & functional checks | `doctor.sh [--fix\|--verbose\|--json]` |
| `install.sh` | Complete setup (5 phases) | `bash install.sh [--bootstrap\|--persistent\|--cli]` |
| `validate.sh` | Validate rules and configuration | `validate.sh [--ci]` |
| `mcp-sync.sh` | Sync MCP servers across IDEs | `mcp-sync.sh` |

## Utility Scripts

Located in `scripts/utils/`:

| Script | Purpose |
|--------|---------|
| `consolidate-memory.sh` | Consolidate knowledge graph monthly |
| `guardian.sh` | Security policy enforcement (@pre-commit) |
| `deploy.sh` | Deployment helper |
| `cron-setup.sh` | Register cron jobs |
| `parallel-analysis.sh` | Parallel file analysis |
| `analyze-sources.sh` | Code source analysis |
| `embed.py` | Vector embedding (Python) |

## Library Modules

Located in `scripts/lib/`:

| Module | Purpose |
|--------|---------|
| `common.sh` | Shared utilities (retry, wait_for, confirm, logging) |
| `colors.sh` | ANSI color definitions |
| `logging.sh` | Consistent logging functions |
| `docker.sh` | Docker/Compose operations |
| `assert.sh` | Assertion and validation helpers |

## Quick Reference

### Initial Setup
```bash
# Full setup (packages + IDE config + CLI + autostart)
bash ~/.brain/scripts/install.sh

# Or bootstrap only (OS packages)
bash ~/.brain/scripts/install.sh --bootstrap
```

### Daily Usage
```bash
# Check health/status
brain status
brain health

# Initialize a project
bash ~/.brain/scripts/init.sh  # (in project directory)

# Full validation
bash ~/.brain/scripts/doctor.sh --fix
```

### Maintenance
```bash
# Sync MCP servers to all IDEs
bash ~/.brain/scripts/mcp-sync.sh

# Consolidate memory (monthly)
bash ~/.brain/scripts/consolidate-memory.sh

# Deploy changes
bash ~/.brain/scripts/deploy.sh
```

## Adding New Scripts

1. **Name clearly**: `{action}-{subject}.sh` (e.g., `sync-mcp.sh`)
2. **Use library modules**: Source `lib/common.sh` for shared functions
3. **Add documentation**: Include usage comment at top
4. **Test before**: Verify script works before committing
5. **Update this README**: Add to appropriate section

## See Also

- [docs/](../docs/) - Architecture and decision records
- [rules/canonical.md](../rules/canonical.md) - Core rules and principles
- [../README.md](../README.md) - Main documentation
