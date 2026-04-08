## Module: Autostart

### Principle
The brain environment must be ready without manual steps after login. If the environment requires a command to activate, that command must be registered as an autostart service.

### Components
1. **Startup Script**: `scripts/init.sh` is the primary entry point for environment initialization.
2. **Registration**: `scripts/autostart-setup.sh` handles the registration of the startup script with the OS (systemd, shell profiles).

### Initialization Steps
At system startup (or shell login), the following must be executed:
- **MCP Check**: Verify all required MCP servers are running.
- **Rules Refresh**: Ensure adapter files match the latest `canonical.md`.
- **Memory Sync**: Perform a background sync of the knowledge graph if cloud backup is enabled.
- **Health Check**: Run a silent `doctor.sh` and log results to `~/.brain/logs/boot.log`.

### Maintenance
- Weekly: Automatically run `update.sh` to pull latest brain repo changes (user confirmed).
- Daily: Run a validation check on all rule schemas.

### Failure Handling
If autostart fails:
1. Log the failure to `~/.brain/logs/autostart-error.log`.
2. Set `BRAIN_READY=0`.
3. Notify the user in the next shell session with: `[BOOT-FAIL] Brain environment failed to initialize. Run 'brain doctor' to diagnose.`
