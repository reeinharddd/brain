# Brain - End-to-End Testing Guide

## Pre-flight Checklist

```bash
# Terminal 1: Start daemon
~/.local/bin/braind

# Terminal 2: Start Desktop UI
cd /mnt/main1tb/work/Personal/brain/apps/desktop && npm run dev

# Terminal 3: Run tests
# Open browser → http://localhost:5173
```

Verify all services:
```bash
curl -s http://localhost:9090/health          # Daemon health
curl -s http://localhost:9090/api/status      # Daemon status
curl -s http://localhost:5173                  # Desktop UI (HTML)
curl -s http://localhost:9090/metrics          # Prometheus metrics
```

---

## Phase 1: CLI Functional Testing

### 1.1 Environment & Root
```bash
# Verify environment detection
brain env
brain root

# Expected:
#   Brain environment: production (or development)
#   BRAIN_ROOT: /mnt/main1tb/work/Personal/brain
```

### 1.2 Daemon Lifecycle
```bash
# Check daemon status
brain status

# List managed processes
brain ps

# Start a test process
brain start my-test echo "hello from brain"

# Verify it's running
brain ps

# Stop it
brain stop my-test
```

### 1.3 Skills — Full CRUD + Search

```bash
# List ALL 35 Go skills
brain skills

# Expected output: 35 skills listed with IDs, names, descriptions

# Search skills
brain skills search "concurrency"
brain skills search "testing"
brain skills search "samber"

# Expected: Filtered results matching the search term

# Create a new skill via API
curl -X POST http://localhost:9090/api/skills \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-skill-e2e",
    "name": "E2E Test Skill",
    "kind": "skill",
    "scope": "global",
    "version": "0.1.0",
    "description": "Created during end-to-end testing",
    "tags": ["test", "e2e", "temporary"],
    "category": "testing",
    "maintained": true,
    "source": "manual-test"
  }'

# Verify it appears in the list
brain skills | grep "test-skill-e2e"

# Get single skill details
curl -s http://localhost:9090/api/skills/test-skill-e2e | python3 -m json.tool

# Update the skill
curl -X PUT http://localhost:9090/api/skills/test-skill-e2e \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-skill-e2e",
    "name": "E2E Test Skill (Updated)",
    "kind": "skill",
    "scope": "global",
    "version": "0.2.0",
    "description": "Updated during testing",
    "tags": ["test", "e2e", "updated"],
    "maintained": true
  }'

# Verify update
curl -s http://localhost:9090/api/skills/test-skill-e2e | python3 -m json.tool

# Delete the skill
curl -X DELETE http://localhost:9090/api/skills/test-skill-e2e

# Verify deletion (should return 404)
curl -s http://localhost:9090/api/skills/test-skill-e2e
```

### 1.4 Skills Security Scan
```bash
# Scan a skill directory for security issues
curl -X POST http://localhost:9090/api/skills/scan \
  -H "Content-Type: application/json" \
  -d '{"path": "/mnt/main1tb/work/Personal/brain/.github/skills/golang-testing"}' | python3 -m json.tool

# Expected: 8-point scan results, all "pass"
```

### 1.5 Skills IDE Compatibility
```bash
# Get skills compatible with a specific IDE
curl -s "http://localhost:9090/api/skills/compatible?surface=vscode" | python3 -m json.tool | head -30
curl -s "http://localhost:9090/api/skills/compatible?surface=claude-code" | python3 -m json.tool | head -30

# Expected: All 35 skills returned with compatible=true
```

### 1.6 MCP Servers — Full CRUD

```bash
# List all MCP servers
brain mcps

# Via API — list with status
curl -s http://localhost:9090/api/mcp/servers | python3 -m json.tool

# Expected: List of registered MCP servers with status, tool count

# Get single MCP details
curl -s http://localhost:9090/api/mcps/filesystem | python3 -m json.tool

# Search MCPs
curl -X POST "http://localhost:9090/api/mcps/search?q=git" | python3 -m json.tool

# List tools for a specific server
curl -s http://localhost:9090/api/mcp/tools/brain-filesystem | python3 -m json.tool

# Expected: read_file, write_file, list_directory, search_files, edit_file
```

### 1.7 Agents
```bash
# List all agents
brain agents

# Via API
curl -s http://localhost:9090/api/agents | python3 -m json.tool

# Expected: 12 agents (architect, debugger, implementer, etc.)

# Get single agent
curl -s http://localhost:9090/api/agents/architect | python3 -m json.tool
```

### 1.8 Workflows

```bash
# List available workflows
brain workflows list

# Expected: feature-dev, bug-fix, refactor, code-review, migration, full-release

# Get workflow DAG structure
curl -s http://localhost:9090/api/workflows/feature-dev/dag | python3 -m json.tool

# Expected: DAG with nodes (architect, builder, tester, reviewer, documenter) and dependencies

# Execute a workflow
curl -X POST http://localhost:9090/api/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{"workflow": "code-review"}' | python3 -m json.tool

# Expected: execution_id returned
```

### 1.9 Delegation

```bash
# List active executions
brain delegation executions

# Create a delegation graph and execute it
cat > /tmp/test-delegation.json << 'EOF'
{
  "root_agent": "architect",
  "mode": "fork",
  "nodes": {
    "node-1": {
      "agent_id": "architect",
      "role": "architect",
      "input": {"description": "Design the feature"},
      "timeout": "30s"
    },
    "node-2": {
      "agent_id": "builder",
      "role": "builder",
      "input": {"description": "Implement the design"},
      "timeout": "30s"
    }
  },
  "edges": {"node-1": ["node-2"]},
  "budget": {"max_tokens": 10000, "max_cost_usd": 1.0, "max_duration": "5m"},
  "fallback": {"steps": []}
}
EOF

brain delegation execute /tmp/test-delegation.json

# Check execution status
brain delegation status <execution_id_from_output>
```

### 1.10 AutoEvolve

```bash
# Enable the engine
curl -X POST http://localhost:9090/api/autoevolve/enable

# Generate telemetry data (simulate activity)
for i in $(seq 1 20); do
  curl -s http://localhost:9090/api/skills > /dev/null
done

# Run analysis
curl -X POST http://localhost:9090/api/autoevolve/run | python3 -m json.tool

# Check recommendations
brain autoevolve recommendations

# Check status
brain autoevolve status

# Disable engine
curl -X POST http://localhost:9090/api/autoevolve/disable
```

### 1.11 Docs RAG (Real Indexer)

```bash
# Search documentation
brain docs-rag "agent delegation"

# Via API with different queries
curl -s "http://localhost:9090/api/docs/search?q=artifact+registry" | python3 -m json.tool
curl -s "http://localhost:9090/api/docs/search?q=policy+hierarchy" | python3 -m json.tool
curl -s "http://localhost:9090/api/docs/search?q=mcp+server" | python3 -m json.tool
curl -s "http://localhost:9090/api/docs/search?q=self-improvement" | python3 -m json.tool

# Expected: Real search results with scores, snippets, paths — NOT stub data

# Check index status
curl -s http://localhost:9090/api/docs/status | python3 -m json.tool

# Expected: document_count > 0, chunk_count > 0
```

### 1.12 Review & Token Waste

```bash
# Show recommendations
brain review list

# Show token waste analysis
brain review waste

# Approve a recommendation (if any)
brain review approve <id>

# Apply approved
brain review apply
```

### 1.13 Context & Memory

```bash
# Context bundle
brain context bundle

# Curator analysis
brain context curator run
brain context curator report

# Memory status
brain memory status
```

### 1.14 Cost Tracking

```bash
brain cost budget
brain cost report
```

### 1.15 Sync System

```bash
# Check sync status
brain sync status

# Trigger dry-run sync
brain sync --dry-run

# Trigger real sync
brain sync

# Monitor logs during sync
brain logs
```

---

## Phase 2: Desktop UI Testing

Open `http://localhost:5173` in browser. Test each tab:

### 2.1 Status Tab
- [ ] Shows 16 subsystems all green (OK)
- [ ] Daemon status card shows "Ready" or actual status
- [ ] Processes count matches `brain ps` output
- [ ] Auto-refreshes every 10 seconds

### 2.2 Skills Tab (CRITICAL)
- [ ] Shows all 35 Go skills in a table
- [ ] Search works: type "concurrency" → filters to golang-concurrency
- [ ] Search by tag: type "testing" → shows golang-testing, golang-stretchr-testify
- [ ] Click "Details" on any skill → shows:
  - Full description
  - Compatible IDEs (15 green checkmarks)
  - Security scan (8/8 pass)
  - Tags list
- [ ] Shows skill count: "Skills Marketplace (35 skills)"

### 2.3 MCP Hub Tab (CRITICAL)
- [ ] Lists all registered MCP servers
- [ ] Shows status (running/stopped/error) for each
- [ ] Shows tool count per server
- [ ] Click a server → right panel shows its tools
- [ ] Tools display name and description
- [ ] Start button attempts to start server

### 2.4 Agents Tab
- [ ] Shows all 12 agents in card layout
- [ ] Each card shows role, tier, capabilities
- [ ] Grid layout (3 columns)

### 2.5 IDE Matrix Tab (CRITICAL)
- [ ] Shows 16 IDEs/CLIs in a table
- [ ] Tier badges (1 = blue, 2 = light blue)
- [ ] All show ✓ for Skills (35)
- [ ] All show ✓ for MCPs, Context, Policy, Memory
- [ ] Icons render correctly (⌨️, 🟦, 🔷, 🟣, etc.)

### 2.6 Events Tab
- [ ] WebSocket connects (or shows disconnect message gracefully)
- [ ] Events stream in real-time (colored by type)
- [ ] Log events = cyan, healthcheck = green, error = red
- [ ] Keeps last 100 events

---

## Phase 3: Cross-IDE/CLI Artifact Sync Testing

This tests the core promise: **one change propagates to all surfaces**.

### 3.1 Create Artifact → Verify Everywhere

```bash
# Step 1: Create a new skill via CLI
curl -X POST http://localhost:9090/api/skills \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cross-ide-test-skill",
    "name": "Cross-IDE Test",
    "kind": "skill",
    "scope": "global",
    "version": "1.0.0",
    "description": "This skill tests cross-surface synchronization",
    "tags": ["cross-ide", "sync-test", "shared"],
    "category": "testing",
    "maintained": true,
    "source": "cross-ide-test",
    "sync_to": ["cli", "vscode", "cursor", "claude-code", "qwen-code"]
  }'

# Step 2: Verify via CLI
brain skills | grep "cross-ide-test-skill"
# Expected: Found

# Step 3: Verify via API (simulating IDE request)
curl -s http://localhost:9090/api/skills/cross-ide-test-skill | python3 -m json.tool
# Expected: Full skill object with all fields

# Step 4: Verify via Desktop UI
# Open http://localhost:5173 → Skills tab → search "cross-ide"
# Expected: "Cross-IDE Test" skill appears in table

# Step 5: Verify search returns it
curl -s "http://localhost:9090/api/skills/search?q=cross-ide" | python3 -m json.tool
# Expected: Returns the skill in results

# Step 6: Verify compatibility check includes it
curl -s "http://localhost:9090/api/skills/compatible?surface=vscode" | python3 -m json.tool | grep "cross-ide"
# Expected: Found in compatible list
```

### 3.2 Update Artifact → Verify Propagation

```bash
# Step 1: Update the skill
curl -X PUT http://localhost:9090/api/skills/cross-ide-test-skill \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cross-ide-test-skill",
    "name": "Cross-IDE Test (v2)",
    "kind": "skill",
    "scope": "global",
    "version": "2.0.0",
    "description": "Updated to version 2 with new features",
    "tags": ["cross-ide", "sync-test", "shared", "v2"],
    "maintained": true
  }'

# Step 2: Verify update reflected everywhere
brain skills | grep "Cross-IDE Test (v2)"
curl -s http://localhost:9090/api/skills/cross-ide-test-skill | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Version: {d[\"version\"]}, Name: {d[\"name\"]}')"
# Expected: Version: 2.0.0, Name: Cross-IDE Test (v2)

# Step 3: Check Desktop UI
# Refresh http://localhost:5173 → Skills tab → search
# Expected: Shows "Cross-IDE Test (v2)" with version 2.0.0
```

### 3.3 Delete Artifact → Verify Removal

```bash
# Step 1: Delete the skill
curl -X DELETE http://localhost:9090/api/skills/cross-ide-test-skill

# Step 2: Verify removal from CLI
brain skills | grep "cross-ide-test-skill"
# Expected: NOT found (empty output)

# Step 3: Verify removal from API (should 404)
curl -s -o /dev/null -w "%{http_code}" http://localhost:9090/api/skills/cross-ide-test-skill
# Expected: 404

# Step 4: Verify removal from search
curl -s "http://localhost:9090/api/skills/search?q=cross-ide" | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Results: {len(d.get(\"results\", []))}')"
# Expected: Results: 0

# Step 5: Check Desktop UI
# Refresh http://localhost:5173 → Skills tab
# Expected: "Cross-IDE Test" no longer appears
```

### 3.4 MCP Server Sync Test

```bash
# Step 1: List MCP servers
curl -s http://localhost:9090/api/mcp/servers | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Servers: {d[\"total\"]}')"

# Step 2: Start a server
curl -X POST http://localhost:9090/api/mcp/servers/brain-filesystem/start

# Step 3: Verify status changed
curl -s http://localhost:9090/api/mcp/servers | python3 -c "import sys,json; [print(f'{s[\"id\"]}: {s[\"status\"]}') for s in json.load(sys.stdin)['servers']]"

# Step 4: List tools for that server
curl -s http://localhost:9090/api/mcp/tools/brain-filesystem | python3 -c "import sys,json; [print(f'  - {t[\"name\"]}: {t[\"description\"][:50]}') for t in json.load(sys.stdin).get('tools',[])]"

# Expected: read_file, write_file, list_directory, search_files, edit_file

# Step 5: Verify in Desktop UI
# Open http://localhost:5173 → MCP Hub tab
# Expected: brain-filesystem shows status, tool count, tools listed in right panel
```

### 3.5 Events Stream Test

```bash
# Step 1: Open Desktop UI Events tab at http://localhost:5173

# Step 2: Trigger various actions and watch events stream:
brain skills search "testing"        # Should generate log events
brain sync --dry-run                  # Should generate sync events
brain agents                          # Should generate agent load events

# Step 3: Verify WebSocket receives events
# Expected: Events appear in real-time with colored types:
#   - "log" events in cyan
#   - "healthcheck" events in green
#   - "status" events in white
```

---

## Phase 4: Registry File Persistence Test

```bash
# Step 1: Create a skill
curl -X POST http://localhost:9090/api/skills \
  -H "Content-Type: application/json" \
  -d '{
    "id": "persistence-test",
    "name": "Persistence Test",
    "kind": "skill",
    "scope": "global",
    "version": "1.0.0",
    "description": "Tests registry.yml persistence",
    "tags": ["persistence", "test"],
    "maintained": true
  }'

# Step 2: Check the actual YAML file
grep -A 5 "persistence-test" /mnt/main1tb/work/Personal/brain/artifacts/skills/registry.yml
# Expected: Skill entry exists in the YAML file

# Step 3: Restart daemon (simulates server restart)
pkill braind
sleep 2
~/.local/bin/braind &
sleep 3

# Step 4: Verify skill survived restart
curl -s http://localhost:9090/api/skills/persistence-test | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Survived restart: {d[\"name\"]}')"
# Expected: Survived restart: Persistence Test

# Cleanup
curl -X DELETE http://localhost:9090/api/skills/persistence-test
```

---

## Phase 5: MCP Tool Call Test

```bash
# Test actual MCP tool execution through the proxy
curl -X POST http://localhost:9090/api/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "server_id": "brain-filesystem",
    "tool_name": "list_directory",
    "arguments": {"path": "/tmp"}
  }' | python3 -m json.tool

# Expected: JSON response with directory listing or error message
# If server is running: actual /tmp contents
# If server not running: error about server status
```

---

## Phase 6: Full System Health

```bash
# Prometheus metrics
curl -s http://localhost:9090/metrics | head -30

# Expected: Prometheus-format metrics including:
#   brain_daemon_uptime_seconds
#   brain_active_sessions
#   brain_artifact_load_errors_total
#   http_requests_duration_seconds

# Health endpoint
curl -s http://localhost:9090/api/v1/health | python3 -m json.tool

# Expected: Detailed health check with component statuses
```

---

## Test Summary Checklist

| Test Area | What to Verify | Pass/Fail |
|-----------|---------------|-----------|
| **Daemon** | Starts, responds to health check | ☐ |
| **CLI: Env** | `brain env` and `brain root` work | ☐ |
| **CLI: Skills** | 35 skills listed, search works | ☐ |
| **CLI: Skills CRUD** | Create → Read → Update → Delete via API | ☐ |
| **CLI: MCPs** | List servers, search, get details | ☐ |
| **CLI: Agents** | 12 agents listed | ☐ |
| **CLI: Workflows** | 6 workflows listed, DAG structure | ☐ |
| **CLI: Delegation** | Execute graph, check status | ☐ |
| **CLI: AutoEvolve** | Enable, run analysis, recommendations | ☐ |
| **CLI: Docs RAG** | Real search results (not stubs) | ☐ |
| **CLI: Review** | List, approve, apply, waste | ☐ |
| **CLI: Context** | Bundle, curator run | ☐ |
| **CLI: Cost** | Budget, report | ☐ |
| **CLI: Sync** | Status, dry-run, real sync | ☐ |
| **UI: Status** | 16 subsystems, stat cards | ☐ |
| **UI: Skills** | 35 skills, search, details, IDE compat | ☐ |
| **UI: MCP Hub** | Servers list, tools browser, start/stop | ☐ |
| **UI: Agents** | Agent pool cards | ☐ |
| **UI: IDE Matrix** | 16 IDEs, all capabilities green | ☐ |
| **UI: Events** | WebSocket streaming | ☐ |
| **Cross-IDE: Create** | Skill appears in CLI + API + UI | ☐ |
| **Cross-IDE: Update** | Changes propagate everywhere | ☐ |
| **Cross-IDE: Delete** | Removed from all surfaces | ☐ |
| **Cross-IDE: Sync** | Registry.yml persisted, survives restart | ☐ |
| **MCP: Tools** | Tool list, tool call through proxy | ☐ |
| **Metrics** | Prometheus metrics scrapeable | ☐ |

---

## Quick Smoke Test (5 minutes)

```bash
# 1. Start daemon
~/.local/bin/braind &
sleep 3

# 2. Check health
curl -s http://localhost:9090/health && echo ""

# 3. List skills (should show 35)
curl -s http://localhost:9090/api/skills | python3 -c "import sys,json; print(f'Skills: {len(json.load(sys.stdin)[\"skills\"])}')"

# 4. Create, update, delete cycle
curl -s -X POST http://localhost:9090/api/skills -H "Content-Type: application/json" -d '{"id":"smoke","name":"Smoke","kind":"skill","scope":"global","version":"1.0.0","description":"test","maintained":true}' > /dev/null
curl -s http://localhost:9090/api/skills/smoke | python3 -c "import sys,json; print(f'Created: {json.load(sys.stdin)[\"name\"]}')"
curl -s -X PUT http://localhost:9090/api/skills/smoke -H "Content-Type: application/json" -d '{"id":"smoke","name":"Smoke v2","kind":"skill","scope":"global","version":"2.0.0","description":"updated","maintained":true}' > /dev/null
curl -s http://localhost:9090/api/skills/smoke | python3 -c "import sys,json; print(f'Updated: {json.load(sys.stdin)[\"version\"]}')"
curl -s -X DELETE http://localhost:9090/api/skills/smoke > /dev/null
code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9090/api/skills/smoke)
echo "Delete: HTTP $code (expected 404)"

# 5. Search docs (real indexer)
curl -s "http://localhost:9090/api/docs/search?q=architecture" | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Docs results: {d[\"metadata\"][\"results_count\"]}')"

# 6. Check IDE compatibility
curl -s "http://localhost:9090/api/skills/compatible?surface=vscode" | python3 -c "import sys,json; print(f'Compatible skills for VS Code: {len(json.load(sys.stdin)[\"skills\"])}')"

# Cleanup
pkill braind
```

Expected output:
```
{"status":"ok"}
Skills: 35
Created: Smoke
Updated: 2.0.0
Delete: HTTP 404 (expected 404)
Docs results: N  (N > 0, real results)
Compatible skills for VS Code: 35
```
