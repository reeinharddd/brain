# Brain Orchestration Architecture

## Centralized Service Management for Multi-IDE/LLM Environments

> **Last Updated**: 2026-03-31  
> **Version**: 1.0.0  
> **Status**: Implemented

---

## Problem Statement

Previously, the Brain system had these issues:

1. **Uncoordinated Autostart** - `init.sh` executed globally at boot (wrong location)
2. **Service Conflicts** - Multiple IDEs launching MCPs without synchronization
3. **No Opt-In Control** - Services started automatically even when not in use
4. **Inconsistent Configuration** - Different IDEs read from different paths
5. **Poor Observability** - No central health checks or logging
6. **MCP Port Conflicts** - Stdio servers competing for ports 8001-8005

---

## New Architecture

### 1. **Central CLI: `brain` Command**

All service management flows through a single, unified interface:

```bash
brain start          # Start all services (Docker + MCP)
brain stop           # Stop all services
brain restart        # Restart services
brain status         # View status dashboard
brain health         # Run full health checks
brain dashboard      # Interactive control panel
brain config         # Show/edit configuration
brain logs [N]       # View logs
brain autostart-enable    # Enable boot startup (systemd)
brain autostart-disable   # Disable boot startup
brain reset          # Clear state and cache
```

### 2. **Configuration-Based Control**

All behavior controlled via `~/.brain/.brain.config`:

```bash
# Global autostart control
BRAIN_AUTOSTART_ENABLED=false

# Service-specific autostart
BRAIN_AUTOSTART_DOCKER=false        # Qdrant
BRAIN_AUTOSTART_MCP=false           # MCP servers

# MCP Port Configuration
BRAIN_MCP_MEMORY_PORT=8001
BRAIN_MCP_FILESYSTEM_PORT=8002
BRAIN_MCP_CONTEXT7_PORT=8003
BRAIN_MCP_SEQUENTIAL_THINKING_PORT=8004
BRAIN_MCP_GITHUB_PORT=8005

# Memory Backend
BRAIN_MEMORY_BACKEND=qdrant         # qdrant|mem0|git|upstash

# Logging
BRAIN_LOG_LEVEL=info                # debug|info|warn|error
BRAIN_HEALTH_CHECK_INTERVAL=60      # seconds
```

**Edit configuration**:

```bash
brain config           # View current config
nano ~/.brain/.brain.config
brain restart          # Apply changes
```

### 3. **Systemd Integration (Optional)**

Autostart only if explicitly enabled:

```bash
brain autostart-enable    # Creates ~/.config/systemd/user/brain.service
# or
brain autostart-disable   # Removes autostart
```

Service file:

```ini
[Unit]
Description=Brain - Centralized AI Engineering Platform
After=network.target

[Service]
Type=oneshot
ExecStart=/home/reeinharrrd/.brain/bin/brain autostart
RemainAfterExit=yes

[Install]
WantedBy=default.target
```

---

## Architecture Diagram

```text
┌─────────────────────────────────────────────────────────────┐
│                    USER INTERFACE LAYER                      │
│─────────────────────────────────────────────────────────────│
│  $ brain start           $ brain dashboard      $ brain status │
│         │                       │                      │       │
└─────────┼───────────────────────┼──────────────────────┼───────┘
          │                       │                      │
          └───────────┬───────────┴──────────┬──────────┘
                      │                      │
          ┌───────────▼────────┐   ┌────────▼─────────┐
          │     brain          │   │  .brain.config   │
          │  (Main Orchestrator)   │ (State & Control) │
          └─────┬──────┬───────┘   └──────────────────┘
                │      │
          ┌─────▼──┐  ┌▼──────────┐
          │ Docker │  │ MCP Stdio  │
          │ Compose│  │ Servers    │
          └────┬───┘  └┬───────────┘
               │      │
          ┌────▼──────▼────────────┐
          │   SERVICE LAYER         │
          │─────────────────────────│
          │ Qdrant (Memory Vector)  │
          │ Memory MCP              │
          │ Filesystem MCP          │
          │ Context7 MCP            │
          │ Sequential MCP          │
          │ GitHub MCP              │
          └─────────────────────────┘
```

---

## Service Initialization Flow

### Scenario 1: Manual Start

```text
$ brain start
  -> Reads .brain.config
  -> Checks Docker
  -> Starts docker compose -f docker-compose.yml (Qdrant)
  -> Waits for Qdrant health check
  -> Starts MCP servers (stdio)
  -> Verifies port connectivity
  -> Reports status
```

### Scenario 2: Systemd Autostart (if enabled)

```text
Boot/Login
  -> systemd user service triggered
  -> ExecStart: brain autostart
  -> Reads BRAIN_AUTOSTART_DOCKER and BRAIN_AUTOSTART_MCP flags
  -> Starts services only if flags = true
  -> Logs boot sequence to telemetry.ndjson
  -> IDE launcher picks it up
```

### Scenario 3: IDE Opens Project

```text
$ code /project
  -> Reads generated MCP config under mcp/global-config-stdio.json
  -> Checks if MCPs are already running (first IDE to start them)
  -> If not running: spawns MCPs via brain
  -> If running: reuses existing MCP gateway
  -> Connects to Qdrant (if BRAIN_MEMORY_BACKEND=qdrant)
```

---

## Key Design Decisions

### 1. **Socket vs. Service Model**

- **Before**: Each IDE spawned its own MCPs (N processes for N IDEs)
- **After**: Single MCP gateway process, multiple IDEs connect to it
- **Benefit**: Reduced resource usage, prevents port conflicts

### 2. **Config-First Architecture**

- **Before**: Hardcoded autostart behavior
- **After**: All behavior driven by `.brain.config`
- **Benefit**: Users can toggle features without code changes

### 3. **Stdio + Docker Split**

- **Stdio MCPs** (memory, filesystem, context7, github, sequential-thinking)
  - No containers needed
  - Direct process management
  - Minimal overhead
- **Docker Services** (Qdrant vector DB)
  - Containerized for consistency
  - Persistent volumes
  - Health checks built-in

### 4. **Per-Project Initialization**

- **Before**: `init.sh` ran globally at boot (wrong)
- **After**: `init.sh` runs per-project when you enter folder
- **Implementation**: IDE rules (`.cursorrules`, etc.) trigger it

---

## IDE-Specific Behavior

### VS Code

```json
// .vscode/settings.json
{
  "github.copilot.advanced": {
    "debug.overrideChatSystemPrompt": false
  }
}
```

Reads config from: `~/.brain/mcp/global-config-stdio.json`

### Cursor / Windsurf

Reads config from: `.cursorrules` / `.windsurfrules` (symlinked to project)

### Claude Desktop

Reads config from: `~/.config/Claude Desktop/claude_desktop_config.json`

All point to same MCP servers on ports 8001-8005.

---

## Control Flow Examples

### Example 1: Start services for current session

```bash
$ brain start
[INFO] Starting Brain services...
[INFO] Starting Docker services (qdrant)...
[SUCCESS] Qdrant is healthy
[INFO] Starting MCP servers (stdio)...
[SUCCESS] MCP servers started (PID: 12345)
[SUCCESS] Brain is ready. Check status with: brain status
```

### Example 2: Enable autostart on boot

```bash
$ brain autostart-enable
[INFO] Enabling global autostart...
[SUCCESS] Autostart enabled (systemd)
[INFO] Services will start on next boot

# Next time you login:
# systemd automatically runs: brain autostart
```

### Example 3: Interactive dashboard

```bash
$ brain dashboard

╔════════════════════════════════════════╗
║         BRAIN STATUS DASHBOARD         ║
╚════════════════════════════════════════╝

Configuration:
  Autostart:       false
  Memory backend:  qdrant
  Config file:     /home/user/.brain/.brain.config

=== Docker Services ===
CONTAINER ID   STATUS
qdrant         healthy

=== MCP Servers ===
✓ MCP Gateway running (PID: 12345)
✓ Port 8001 is open (memory)
✓ Port 8002 is open (filesystem)
✓ Port 8003 is open (context7)
✓ Port 8004 is open (sequential-thinking)
✓ Port 8005 is open (github)

Commands:
  1) Start       - Start all services
  2) Stop        - Stop all services
  3) Restart     - Restart services
  ...
```

### Example 4: Check health

```bash
$ brain health

[INFO] Running health checks...

=== Docker Health ===
✓ Docker services healthy

=== MCP Health ===
✓ MCP Gateway running (PID: 12345)
✓ Port 8001 is open
✓ Port 8002 is open
✓ Port 8003 is open
✓ Port 8004 is open
✓ Port 8005 is open

=== Memory Backend Health ===
✓ Qdrant responding

=== Environment ===
✓ brain.env loaded

[SUCCESS] All health checks passed
```

---

## Debugging Commands

### View recent logs

```bash
brain logs 100         # Last 100 lines
cat ~/.brain/logs/brain-cli.log
cat ~/.brain/logs/telemetry.ndjson
```

### Check config

```bash
brain config
cat ~/.brain/.brain.config
```

### Reset state (clears pids, cache)

```bash
brain reset
```

### View Docker status

```bash
docker compose -f ~/.brain/docker/docker-compose.yml ps
docker logs qdrant
```

### Check MCP ports

```bash
netstat -tlnp | grep 800[1-5]
lsof -i :8001
```

---

## Migration Guide

### From Old Autostart to New System

**Old behavior** (automatic, uncontrolled):

```bash
# init.sh ran at boot via systemd
# MCPs started automatically
# Hard to disable or control
```

**New behavior** (explicit, configurable):

```bash
# Autostart disabled by default
# User explicitly enables: brain autostart-enable
# Change behavior in ~/.brain/.brain.config
# Use: brain start / brain stop for manual control
```

### Migration Steps

- Disable old systemd service

  ```bash
  systemctl --user disable brain.service
  systemctl --user daemon-reload
  ```

- Install new brain CLI

  ```bash
  chmod +x ~/.brain/bin/brain
  ln -sf ~/.brain/bin/brain ~/.local/bin/brain
  ```

- Test manually

  ```bash
  brain start
  brain status
  brain stop
  ```

- Optional: Enable autostart

  ```bash
  brain autostart-enable
  ```

- Update project initialization

  ```text
  /init
  brain sync
  ```

---

## Troubleshooting

### Problem: MCP servers not starting

```bash
# 1. Check if ports are in use
lsof -i :8001

# 2. Use different port in config
nano ~/.brain/.brain.config
# Change: BRAIN_MCP_MEMORY_PORT=8011
brain restart

# 3. Check logs
brain logs
cat ~/.brain/logs/brain-cli.log
```

### Problem: Qdrant not responding

```bash
# 1. Check Docker status
docker ps | grep qdrant

# 2. Restart Qdrant
docker compose -f ~/.brain/docker/docker-compose.yml restart qdrant

# 3. Check Qdrant logs
docker logs qdrant

# 4. Verify volume
docker volume ls | grep brain
```

### Problem: Autostart not working

```bash
# 1. Verify systemd service
systemctl --user status brain.service

# 2. Check logs
journalctl --user -u brain.service -n 50

# 3. Re-enable
brain autostart-disable
brain autostart-enable

# 4. Manual test
brain autostart
```

---

## Future Enhancements

- [ ] Web dashboard (port 8000)
- [ ] Health check notifications
- [ ] Multi-cloud memory sync (Mem0, Upstash)
- [ ] Provider fallback monitoring
- [ ] Performance metrics dashboard
- [ ] Automatic MCP restart on failure
- [ ] TLS/auth for remote MCPs

---

## Files Modified/Created

| File                   | Purpose                                        |
| ---------------------- | ---------------------------------------------- |
| `cli/cmd/brain/main.go`| Main orchestrator                              |
| `daemon/cmd/braind/`   | Runtime daemon and sync engine                 |
| `brain sync`           | Project-local configuration refresh            |
| `.brain.config`        | Central configuration (generated on first run) |
| `.brain-state/pids/`   | Process tracking                               |
| `logs/brain-cli.log`   | Detailed logs                                  |

---

## Summary

The new system provides:

✓ **One command for everything** - `brain start|stop|status|dashboard`  
✓ **Opt-in autostart** - Disabled by default, user controls via config  
✓ **Multi-IDE support** - Single MCP gateway, multiple IDE clients  
✓ **Central control** - All configuration in `~/.brain/.brain.config`  
✓ **Better debugging** - Health checks, logs, status dashboard  
✓ **No conflicts** - Port management, process tracking  
✓ **Agnóstic** - Works with VS Code, Cursor, Windsurf, Claude Desktop

This enables truly shared knowledge that works consistently across all IDEs and LLM providers.
