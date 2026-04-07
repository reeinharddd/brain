---
id: VERIFICATION-TESTING-PLAN
title: Complete Verification Testing - Phase 4
status: in-progress
date_created: 2026-04-03
language: en
type: testing-guide
category: verification
version: 1.0.0
---

# Complete Verification Testing Plan

**Objective**: Verify all 4 Phase 4 features work in real development environment  
**Date**: April 3, 2026  
**Components to Test**: React UI + Daemon API + MCP + Caching

---

## Prerequisites Check

```bash
# 1. Verify Go environment
go version  # Should be 1.24+

# 2. Verify Node environment
bun --version  # Should be 1.3.10+

# 3. Verify Docker
docker compose version  # Should have compose command (space, not hyphen)

# 4. Navigate to project root
cd ~/.brain
```

---

## Step 1: Start Infrastructure (Qdrant)

```bash
# Start Qdrant in background
docker compose up -d qdrant

# Verify Qdrant is running
sleep 5
curl -s http://localhost:6333/health | jq .

# Expected output: { "status": "ok" }
```

---

## Step 2: Test MCP Server (Phase 1-3)

### Option A: Via CLI Wrapper

```bash
# Navigate to MCP directory
cd ~/.brain/mcp/docs-rag-mcp

# Build the MCP binary
go build -o ../../bin/docs-rag-mcp main.go

# If build fails, check issues:
# - Missing redis dependency? Add: go get github.com/redis/go-redis/v9
# - Other errors? Check: go vet ./...

# Test MCP via CLI
cd ~/.brain
./bin/docs-rag-mcp search "daemon architecture" --limit 5

# Expected: JSON response with search results
```

### Option B: Direct MCP Protocol Test

```bash
# Test via JSON-RPC stdio
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | \
  ./bin/docs-rag-mcp

# Expected: List of 3 tools (docs_search, docs_status, docs_rebuild)
```

---

## Step 3: Start Daemon + API (Phase 4.2)

```bash
# In a NEW terminal:
cd ~/.brain/daemon/cmd/braind

# Build and run daemon
go run main.go

# Expected output:
# INFO: Daemon starting on :9090
# INFO: MCP server initialized
# INFO: API handlers registered
```

---

## Step 4: Test Daemon API Endpoints (Phase 4.2)

### 4A: Test Status Endpoint

```bash
# In another terminal
curl -s http://localhost:9090/api/docs/status | jq .

# Expected response:
# {
#   "status": "ready",
#   "total_indexed": 78,
#   "index_location": "/path/to/.brain/qdrant_data",
#   "last_rebuild": "2026-04-03T15:30:00Z",
#   "cache_status": "operational"
# }
```

### 4B: Test Search Endpoint

```bash
# First search (should be slower, 400-500ms)
time curl -s "http://localhost:9090/api/docs/search?q=architecture&limit=5" | jq '.results[0:2]'

# Expected: Array of search results with scores

# Second search same query (should be faster from cache, <50ms)
time curl -s "http://localhost:9090/api/docs/search?q=architecture&limit=5" | jq '.results[0:2]'

# Compare timing: 2nd should be ~8x faster
```

### 4C: Test with Domain Filter

```bash
# Search in specific domain
curl -s "http://localhost:9090/api/docs/search?q=testing&domain=testing&limit=3" | jq .

# Expected: Results filtered to testing domain only
```

### 4D: Test Rebuild Endpoint (Dev-only)

```bash
# Set to development (allows rebuild)
export BRAIN_ENV=development

# Restart daemon
# (kill previous, run again with BRAIN_ENV set)

# Test rebuild
curl -X POST http://localhost:9090/api/docs/rebuild

# Expected: Return with {"status": "success", "duration_ms": 2500}

# Test production guard (should fail)
export BRAIN_ENV=production
curl -X POST http://localhost:9090/api/docs/rebuild

# Expected: 403 Forbidden - rebuild not allowed in production
```

---

## Step 5: Test Incremental Indexing (Phase 4.3)

```bash
# Make a small documentation change
echo "" >> ~/.brain/docs/adr/ADR-0006-docs-rag-mcp.md

# Add entry to changelog.jsonl
echo '{"timestamp":"2026-04-03T15:35:00Z","commit":"test","action":"modify","domain":"adr","file":"docs/adr/ADR-0006-docs-rag-mcp.md","checksum":"abc123"}' \
  >> ~/.brain/docs-changelog.jsonl

# Trigger rebuild (should be FAST: 100-500ms instead of 2-5s)
time curl -X POST http://localhost:9090/api/docs/rebuild

# Expected timing: <1000ms (vs 2-5s full rebuild)
```

---

## Step 6: Test Caching Layer (Phase 4.4)

### 6A: Without Redis (NoopCache)

```bash
# Kill any Redis if running
docker ps | grep redis

# Search query (no cache)
time curl -s "http://localhost:9090/api/docs/search?q=architecture&limit=5" -w "\nTime: %{time_total}s\n"

# Result: ~400-500ms (Qdrant search time)

# Search again (still no cache, no Redis)
time curl -s "http://localhost:9090/api/docs/search?q=architecture&limit=5" -w "\nTime: %{time_total}s\n"

# Result: ~400-500ms (no improvement without Redis)
```

### 6B: With Redis (Redis Cache)

```bash
# Start Redis
docker run -d -p 6379:6379 redis:7

# Add REDIS_ADDR to environment
export REDIS_ADDR=localhost:6379

# Restart daemon with Redis
# (kill and restart daemon with REDIS_ADDR set)

# First search (cache miss)
time curl -s "http://localhost:9090/api/docs/search?q=daemon&limit=5" -w "\nTime: %{time_total}s\n"

# Result: ~400-500ms (Qdrant search)

# Second search (cache hit from Redis)
time curl -s "http://localhost:9090/api/docs/search?q=daemon&limit=5" -w "\nTime: %{time_total}s\n"

# Result: <100ms (Redis cache hit - 4-5x faster!)

# Status endpoint (5s TTL cache)
time curl -s http://localhost:9090/api/docs/status -w "\nTime: %{time_total}s\n"

# Repeat status (within 5s cache window)
time curl -s http://localhost:9090/api/docs/status -w "\nTime: %{time_total}s\n"

# Should be <10ms (cache hit)
```

### 6C: Test Cache Invalidation

```bash
# Verify cache populated
curl -s "http://localhost:9090/api/docs/search?q=test&limit=5" > /tmp/result1.json

# Trigger rebuild (should flush cache)
curl -X POST http://localhost:9090/api/docs/rebuild

# Search same query (should be slow = cache miss)
time curl -s "http://localhost:9090/api/docs/search?q=test&limit=5" > /tmp/result2.json

# Result: ~400-500ms (cache was invalidated and cleared)
```

---

## Step 7: Test React UI (Phase 4.1)

### Option A: Desktop App (Tauri)

```bash
# In terminal at ~/.brain/desktop
cd ~/.brain/desktop

# Install dependencies
bun install

# Build and run desktop app
bun tauri dev

# Open app -> Find "Docs" section
# Try searching for: "daemon", "qdrant", "cache"
# Verify:
# - ✅ Results appear
# - ✅ Relevance scores show
# - ✅ Dark mode works (click theme toggle)
# - ✅ Status panel shows index health on right side
```

### Option B: Web App (React Dev Server)

```bash
cd ~/.brain/desktop

# Run dev server only (no Tauri)
bun run dev

# Open http://localhost:5173
# Navigate to Docs section
# Test functionality:
# - ✅ Type in search box
# - ✅ Results load with spinner
# - ✅ Click domain filter
# - ✅ See error handling (try empty query)
# - ✅ Status panel auto-updates (30s polling)
```

---

## Step 8: Integration Test - Full Workflow

```bash
# Scenario: User searches, gets results, status updates

# Start everything:
# 1. Docker Qdrant: docker compose up -d qdrant
# 2. Redis (optional): docker run -d -p 6379:6379 redis:7
# 3. Daemon: cd daemon/cmd/braind && go run main.go
# 4. React UI: cd desktop && bun tauri dev

# Workflow:
# 1. Open Docs page in desktop app
# 2. Search: "authentication"
#    - Should show 5-10 results
#    - Scores should be >0.5
# 3. Check Status panel
#    - Should show: "Ready" status
#    - Total docs: 78
# 4. Rebuild index
#    - Click "Rebuild" button (if available)
#    - Should complete in <5s
#    - Results cache cleared
# 5. Search again
#    - Should show updated results
#    - If cached: <100ms, if not: ~400ms

print "✅ Full workflow test complete!"
```

---

## Verification Checklist

### Phase 4.1: React UI

- [ ] Components render without errors
- [ ] Search input accepts text
- [ ] Results display with scores
- [ ] Domain filter works
- [ ] Status panel shows index health
- [ ] Dark mode toggles
- [ ] Mobile responsive (resize browser)
- [ ] Loading state visible
- [ ] Error handling works (tryempty query)

### Phase 4.2: Daemon API

- [ ] Status endpoint returns IndexHealth
- [ ] Search endpoint returns results
- [ ] Results have scores (floats 0-1)
- [ ] Domain filter reduces results
- [ ] Rebuild endpoint works in dev
- [ ] Rebuild blocked in production
- [ ] Error responses are meaningful
- [ ] All endpoints return JSON

### Phase 4.3: Incremental Indexing

- [ ] Small change triggers incremental rebuild
- [ ] Incremental rebuild <1s (vs 2-5s full)
- [ ] Changelog parsed correctly
- [ ] Large change (>30%) triggers full rebuild
- [ ] Index integrity validated
- [ ] No orphaned documents

### Phase 4.4: Caching

- [ ] Without Redis: no caching (NoopCache)
- [ ] With Redis: search cached 1h
- [ ] Status cached 5s
- [ ] Cache hit 8-10x faster
- [ ] Cache invalidation on rebuild
- [ ] Graceful fallback if Redis unavailable
- [ ] Redis connection retries work
- [ ] Cache metrics available (via health endpoint)

---

## Common Issues & Fixes

### Issue: "Cannot connect to localhost:9090"

**Fix**: Ensure daemon is running and on correct port

```bash
ps aux | grep braind
```

### Issue: "Qdrant connection refused"

**Fix**: Start Qdrant first

```bash
docker compose up -d qdrant
docker compose ps
```

### Issue: "Redis connection timeout"

**Fix**: Either start Redis or allow NoopCache fallback

```bash
docker run -d -p 6379:6379 redis:7
# OR just let it use NoopCache fallback (slower but works)
```

### Issue: "Search returns no results"

**Fix**: Index might not be built. Trigger rebuild:

```bash
curl -X POST http://localhost:9090/api/docs/rebuild
# Wait 5-10 seconds for completion
```

### Issue: "React app won't start"

**Fix**: Install dependencies and check Node version

```bash
cd ~/.brain/desktop
bun install
bun run dev  # Should start on :5173
```

### Issue: Go build fails with "redis/go-redis/v9 not found"

**Fix**: Add dependency

```bash
cd ~/.brain/mcp/docs-rag-mcp
go get github.com/redis/go-redis/v9
go build -o ../../bin/docs-rag-mcp main.go
```

---

## Performance Benchmarks (Expected)

| Operation                   | Expected Time | Notes            |
| --------------------------- | ------------- | ---------------- |
| First search (no cache)     | 400-500ms     | Qdrant + network |
| Cached search               | 50-100ms      | Redis lookup     |
| Incremental rebuild (1 doc) | 100-200ms     | Fast delta       |
| Full rebuild (all docs)     | 2-5s          | Qdrant indexing  |
| Status check (no cache)     | 50ms          | Qdrant query     |
| Status check (cached)       | <5ms          | Redis hit        |
| React UI render             | <500ms        | All components   |
| Cache invalidation          | <50ms         | Redis flush      |

**If your numbers are significantly slower, something may be wrong!**

---

## Success Criteria

✅ **All tests pass**:

- React UI loads and is responsive
- Search returns relevant results
- API endpoints work (status, search, rebuild)
- Caching works (with/without Redis)
- Incremental indexing is faster
- No errors in console/logs

✅ **Performance meets expectations**:

- Search: 400-500ms uncached, <100ms cached
- Rebuild: <1s incremental, 2-5s full
- Status: 5-10ms (cached)

✅ **Production safety**:

- Rebuild blocked when BRAIN_ENV=production
- All errors logged properly
- No hardcoded secrets visible
- Graceful degradation without Redis

---

## Next Steps After Verification

If all tests pass:

1. Deploy daemon as systemd service (auto-start)
2. Integrate React UI into desktop app permanently
3. Add cron job for scheduled rebuilds
4. Set up monitoring/alerting
5. Deploy to production

If tests fail:

1. Check specific error messages above
2. Review checkpoint documents for implementation details
3. Use logs to debug issues
4. Verify all dependencies installed

---

**Ready to test? Start with Step 1! 🚀**
