# Brain CLI Integration for Adapters

## How to Call `brain` from IDE Rules and Adapters

This document explains how `.cursorrules`, `.windsurfrules`, and AI instructions can integrate with the centralized `brain` CLI.

---

## Quick Reference

```bash
# These can be called from IDE rules, adapters, or scripts:

brain start              # Start all services
brain stop               # Stop all services
brain restart            # Restart services
brain status             # Show compact status
brain health             # Full health check
brain dashboard          # Interactive menu (for development)
brain config             # Show current configuration
brain logs [N]           # Show N last log lines
```

---

## Integration Points

### 1. In `.cursorrules` or `.windsurfrules`

These files are sourced by the IDE when the project opens. You can add:

```markdown
# Ensure Brain services are ready

# Call before complex tasks that need memory/context:

# $ brain health

## Guidelines

- Always run `brain status` before multi-IDE sessions
- Use `brain dashboard` to manage services interactively
- If you see port conflicts, use `brain config` to adjust ports
```

### 2. In AI Instructions (copilot-instructions.md)

```markdown
## Brain Service Dependencies

I work with a centralized brain infrastructure. However, services are not automatically started.

### Before Complex Tasks

- Check service status: `brain status`
- Start services if needed: `brain start`
- Verify health: `brain health`

### If You See Errors

- "Connection refused" → Run `brain start`
- "Port already in use" → Run `brain stop` first, then `brain start`
- "Qdrant not responding" → Run `brain health` to diagnose
```

### 3. In Pre-Commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Ensure Brain services are healthy before committing
if ! brain health &>/dev/null; then
    echo "Warning: Brain services not healthy"
    echo "Run: brain start"
    exit 1
fi

# Run guardian checks
bash ~/.brain/guardian/run.sh --staged --threshold critical
```

### 4. In build automation

```text
setup:
  brain daemon-stop
  brain sync

start:
  brain daemon-start

stop:
  brain daemon-stop

dev: setup start
  brain status
```

---

## Environment Integration

### Share Brain State with IDEs

Add to your project's `.env.example`:

```bash
# Brain Service Configuration
BRAIN_ROOT=~/.brain
BRAIN_CONFIG=~/.brain/.brain.config
BRAIN_LOGS=~/.brain/logs

# MCP Server Ports (must match brain config)
BRAIN_MCP_MEMORY_PORT=8001
BRAIN_MCP_FILESYSTEM_PORT=8002
BRAIN_MCP_CONTEXT7_PORT=8003
BRAIN_MCP_SEQUENTIAL_THINKING_PORT=8004
BRAIN_MCP_GITHUB_PORT=8005

# Memory Backend
BRAIN_MEMORY_BACKEND=qdrant
```

Then in your adapter/instruction files:

```bash
# Source brain configuration
if [[ -f ~/.brain/.brain.config ]]; then
    source ~/.brain/.brain.config
fi

# Now you can use BRAIN_MCP_MEMORY_PORT, etc.
```

---

## Examples

### Example 1: IDE startup automation

In a custom shell function or script:

```bash
# ~/.config/zsh/functions/brain-dev

brain-dev() {
    local project="${1:-.}"

    cd "$project"

    # Refresh project with Brain rules
    brain sync

    # Start services
    echo "Starting Brain services..."
    brain daemon-start

    # Show status
    brain status

    echo ""
    echo "Ready to work! Use:"
    echo "  brain status      - Check service status"
    echo "  brain health      - Run health checks"
    echo "  brain stop        - Stop services"
}
```

Usage:

```bash
brain-dev /path/to/project
```

### Example 2: GitHub Actions workflow

```yaml
# .github/workflows/brain-tests.yml
name: Brain Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Start Brain services
        run: |
          brain sync
          brain daemon-start

      - name: Verify health
        run: brain status

      - name: Run tests
        run: npm test

      - name: Stop services
        run: brain stop
```

### Example 3: Multi-project development

```bash
#!/bin/bash
# ~/bin/work-on

project="$1"

if [[ ! -d "$project" ]]; then
    echo "Project not found: $project"
    exit 1
fi

# Stop previous project services
brain stop 2>/dev/null

# Switch to new project
cd "$project"

# Initialize if needed
if [[ ! -f .cursorrules ]]; then
  brain sync
fi

# Start fresh
brain restart
brain status

# Open IDE
code .
```

Usage:

```bash
work-on ~/projects/my-app
```

---

## Troubleshooting in Adapters

### Problem: "command brain not found"

**Solution**: Add to your adapter:

```bash
export BRAIN_ROOT="${BRAIN_ROOT:-$HOME/.brain}"
export PATH="${BRAIN_ROOT}/bin:${PATH}"
```

### Problem: IDE doesn't have permission to run brain

**Solution**: Ensure scripts are executable:

```bash
chmod +x ~/.brain/bin/brain
chmod +x ~/.brain/bin/brain
```

### Problem: brain command hangs or times out

**Solution**: Check if services are stuck:

```bash
# Kill stuck processes
pkill -f "brain-cli"
pkill -f "mcp-gateway"

# Reset state
brain reset

# Try again
brain start
```

---

## Best Practices

1. **Always check before starting**

   ```bash
   brain status   # See if already running
   brain start    # Only if needed
   ```

2. **Use health checks in critical paths**

   ```bash
   if ! brain health >/dev/null 2>&1; then
       echo "Starting services..."
       brain start
   fi
   ```

3. **Clean up after yourself**

   ```bash
   # End of script
   brain stop
   ```

4. **Log service actions**

   ```bash
   echo "Brain logs: $(brain logs 5)"
   ```

5. **Don't hardcode ports**

   ```bash
   # Read from config instead
   source ~/.brain/.brain.config
   curl "http://localhost:${BRAIN_MCP_MEMORY_PORT}/..."
   ```

---

## IDE-Specific Integration

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Start Brain",
      "command": "bash",
      "args": ["-c", "brain start"],
      "problemMatcher": [],
      "presentation": {
        "echo": true,
        "reveal": "silent"
      }
    },
    {
      "label": "Brain Status",
      "command": "bash",
      "args": ["-c", "brain status"],
      "problemMatcher": []
    }
  ]
}
```

### Cursor / Windsurf

Add to `.cursorrules`:

```markdown
## Development Environment

Start development environment:
$ brain start

Check status anytime:
$ brain status

Stop when done:
$ brain stop
```

---

## Summary

The `brain` CLI is designed to be called from anywhere:

- IDE rules files (`.cursorrules`, `.windsurfrules`)
- AI instructions (`copilot-instructions.md`)
- Git hooks
- Build scripts
- GitHub Actions
- Custom shell functions
- Makefiles

This ensures services are coordinated across all your tools while remaining simple and opt-in.
