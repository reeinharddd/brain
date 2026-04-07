---
id: PHASE-4-PLAN
title: Docs-RAG MCP Phase 4 - Optimization & UI
status: in_progress
date_created: 2026-04-03
language: en
type: implementation-plan
category: planning
version: 1.0.0
---

## Phase 4 Implementation Plan: Optimization & UI

**Status**: IN PROGRESS  
**Date**: April 3, 2026  
**Scope**: React UI + Daemon API + Incremental Indexing + Caching

---

## Overview

Phase 4 implements 4 optional enhancements to Phase 1-3 foundation:

1. **React Dashboard UI** (Task 4.1-4.3) - Search interface + results
2. **Daemon API Wrapper** (Task 4.4-4.6) - HTTP endpoints for MCP tools
3. **Incremental Indexing** (Task 4.7-4.9) - Rebuild only changed docs
4. **Caching Layer** (Task 4.10-4.12) - Redis cache for frequent queries

**Priority Order**:

1. React UI first (user-visible, high impact)
2. Daemon API (enables integration)
3. Incremental indexing (performance)
4. Caching (optimization)

---

## Task Breakdown

### Phase 4.1: React Dashboard UI (12 hours)

**Task 4.1.1**: Create DocsSearch component (3h)

- Search input with autocomplete
- Query suggestions from manifest domains
- UX: Clear, responsive, real-time feedback

**Task 4.1.2**: Results display component (3h)

- Result card list
- Score display + ranking
- Snippet preview
- Link to full document

**Task 4.1.3**: Status panel (2h)

- Index status (building/ready/error)
- Document count + chunk count
- Last rebuild time
- Qdrant health indicator

**Task 4.1.4**: Integration with daemon API (4h)

- Fetch via /api/docs/search
- Polling for status updates (30s interval)
- Error states + retry logic
- Loading states

**Deliverables**:

- `desktop/src/components/DocsSearch.tsx`
- `desktop/src/components/DocsResults.tsx`
- `desktop/src/components/DocsStatus.tsx`
- `desktop/src/hooks/useDocsSearch.ts`
- `desktop/src/api/docsApi.ts`

---

### Phase 4.2: Daemon API Wrapper (8 hours)

**Task 4.2.1**: API handler structure (2h)

- REST routes: `/api/docs/search`, `/api/docs/status`, `/api/docs/rebuild`
- Request/response types
- Error handling middleware

**Task 4.2.2**: Search endpoint (2h)

- POST /api/docs/search
- Calls MCP docs_search tool
- Returns formatted JSON

**Task 4.2.3**: Status endpoint (2h)

- GET /api/docs/status
- Calls MCP docs_status tool
- Caches response (5s)

**Task 4.2.4**: Rebuild endpoint (2h)

- POST /api/docs/rebuild (dev-only)
- Blocked in production
- Returns rebuild status

**Deliverables**:

- `daemon/internal/api/handlers/docs.go`
- Integration with daemon main
- Error responses
- Auth/security layer

---

### Phase 4.3: Incremental Indexing (8 hours)

**Task 4.3.1**: Changelog watcher (2h)

- Monitor `docs-changelog.jsonl`
- Parse daily entries
- Extract changed files list

**Task 4.3.2**: Delta detection (2h)

- Compare current docs vs changelog
- Identify added/modified/deleted
- Filter by domain if needed

**Task 4.3.3**: Incremental build (3h)

- Load only changed documents
- Re-chunk changed docs
- Upsert to Qdrant (update vectors)
- Remove deleted docs from index

**Task 4.3.4**: Validation (1h)

- Verify rebuild integrity
- Check document count matches manifest
- Log discrepancies

**Deliverables**:

- `internal/indexer/changelog.go`
- `internal/indexer/delta.go`
- `internal/indexer/incremental.go`
- Tests (5+ test cases)

---

### Phase 4.4: Caching Layer (6 hours)

**Task 4.4.1**: Redis client setup (1h)

- Connect to Redis (optional, defaults to no-cache)
- Graceful degradation if Redis unavailable
- Connection pool management

**Task 4.4.2**: Search query cache (2h)

- Cache search results by query
- TTL: 1 hour for frequent queries
- Invalidate on rebuild

**Task 4.4.3**: Status cache (1h)

- Cache status response (5s TTL)
- Reduce Qdrant queries
- Invalidate on rebuild

**Task 4.4.4**: Cache invalidation (2h)

- On rebuild: clear all caches
- On new docs: invalidate related
- Manual cache clear endpoint

**Deliverables**:

- `internal/cache/redis.go`
- `internal/cache/cache.go`
- Cache layer in tools
- Tests (4+ test cases)

---

## Implementation Order

### Priority 1: React UI (Weeks 1-2)

1. Setup Vite + React components
2. Build search interface
3. Build results display
4. Add status panel
5. Wire to daemon (once API ready)

### Priority 2: Daemon API (Weeks 2)

1. Create handlers
2. Wire /api/docs/\* routes
3. Test with curl/Postman
4. Connect to React UI

### Priority 3: Incremental Indexing (Weeks 3)

1. Changelog watcher
2. Delta detection
3. Incremental build
4. Validation

### Priority 4: Caching (Week 3)

1. Redis setup
2. Search query cache
3. Status cache
4. Cache invalidation

---

## Acceptance Criteria

### Phase 4.1: React UI Complete

- [ ] Search component renders
- [ ] Input validation (query required)
- [ ] Results display with scores
- [ ] Status panel shows metrics
- [ ] Loading states visible
- [ ] Error handling graceful
- [ ] Mobile responsive
- [ ] Keyboard accessible
- [ ] All components typed (TypeScript strict)
- [ ] Tests >80% coverage

### Phase 4.2: Daemon API Complete

- [ ] GET /api/docs/status returns index status
- [ ] POST /api/docs/search returns results
- [ ] POST /api/docs/rebuild (dev-only)
- [ ] Error responses are JSON-RPC format
- [ ] Production guard working
- [ ] Latency <1s for search
- [ ] Status cached (5s)
- [ ] Integration tests passing

### Phase 4.3: Incremental Indexing Complete

- [ ] Changelog watcher works
- [ ] Delta detection accurate
- [ ] Incremental rebuild faster
- [ ] Document count matches manifest
- [ ] No orphaned vectors in Qdrant
- [ ] Tests covering all scenarios
- [ ] Performance <2s for small changes

### Phase 4.4: Caching Complete

- [ ] Redis optional (graceful fallback)
- [ ] Search queries cached (1h TTL)
- [ ] Status cached (5s TTL)
- [ ] Cache invalidation on rebuild
- [ ] Metrics: hit/miss rates
- [ ] Tests for cache logic
- [ ] Memory efficient

---

## Technology Stack

### React UI

- **Framework**: React 18 + TypeScript
- **Build**: Vite (already in project)
- **Styling**: TailwindCSS (if available)
- **State**: React hooks (useDocsSearch, useDocsStatus)
- **HTTP**: fetch API

### Daemon API

- **Language**: Go (like rest of daemon)
- **HTTP Framework**: Standard net/http
- **Routing**: Custom handler registration
- **Docs**: OpenAPI/Swagger optional

### Incremental Indexing

- **Language**: Go
- **File Watching**: Standard os.Stat + polling
- **JSON**: encoding/json for changelog

### Caching

- **Database**: Redis (optional dependency)
- **Go Client**: redis/go-redis v9
- **Fallback**: In-memory map if Redis unavailable

---

## Files to Create/Modify

### React Components (New)

```
desktop/src/
├── components/
│   ├── DocsSearch.tsx          [150 lines]
│   ├── DocsResults.tsx         [200 lines]
│   └── DocsStatus.tsx          [100 lines]
├── hooks/
│   └── useDocsSearch.ts        [80 lines]
├── api/
│   └── docsApi.ts              [150 lines]
├── types/
│   └── docs.ts                 [50 lines]
└── pages/
    └── Docs.tsx               [200 lines]
```

### Daemon API (New)

```
daemon/internal/
├── api/handlers/
│   └── docs.go                [250 lines]
├── docs/
│   ├── client.go              [80 lines]
│   └── types.go               [40 lines]
└── cache/
    └── docs.go                [100 lines]
```

### Indexer Enhancement (Modify)

```
internal/indexer/
├── changelog.go               [150 lines] NEW
├── delta.go                   [120 lines] NEW
├── incremental.go             [200 lines] NEW
├── indexer.go                 [+50 lines] modify Build()
└── indexer_test.go            [+100 lines] new tests
```

### Caching (New)

```
internal/cache/
├── redis.go                   [150 lines] NEW
├── cache.go                   [200 lines] NEW
└── cache_test.go              [100 lines] NEW
```

---

## Testing Strategy

### React Components

```typescript
// DocsSearch.tsx tests
-test("renders search input") -
  test("validates query required") -
  test("calls api on submit") -
  test("shows loading state") -
  test("displays error on failure") -
  // DocsResults.tsx tests
  test("renders result cards") -
  test("displays scores and priority") -
  test("shows snippet preview") -
  test("links to full document") -
  // DocsStatus.tsx tests
  test("shows index status") -
  test("displays document count") -
  test("shows Qdrant health") -
  test("indicates last rebuild time");
```

### Daemon API

```go
// handlers/docs_test.go
- TestSearchEndpoint_ValidQuery()
- TestSearchEndpoint_EmptyQuery()
- TestSearchEndpoint_DomainFilter()
- TestStatusEndpoint_ReturnsStatus()
- TestRebuildEndpoint_DevAllowed()
- TestRebuildEndpoint_ProdBlocked()
```

### Incremental Indexing

```go
// incremental_test.go
- TestChangelogWatcher_DetectsChanges()
- TestDeltaDetection_AddsRemovesModifies()
- TestIncrementalBuild_UpdatesOnly()
- TestValidation_MatchesManifest()
```

### Caching

```go
// cache_test.go
- TestRedisCache_SetGet()
- TestCacheHitRate_Metrics()
- TestCacheInvalidation_OnRebuild()
- TestGracefulFallback_NoRedis()
```

---

## Definition of Done

Phase 4 is COMPLETE when:

✅ **React UI**

- [ ] All components render correctly
- [ ] Search works end-to-end
- [ ] Type-safe (TypeScript strict)
- [ ] > 80% test coverage
- [ ] Responsive design verified
- [ ] Error handling comprehensive

✅ **Daemon API**

- [ ] All endpoints functional
- [ ] Integration tests passing
- [ ] Production guard verified
- [ ] Performance acceptable (<1s)
- [ ] Error responses valid

✅ **Incremental Indexing**

- [ ] Detects changed documents
- [ ] Rebuilds are faster than full
- [ ] Document count always accurate
- [ ] Tests covering all scenarios

✅ **Caching**

- [ ] Works with/without Redis
- [ ] Performance improvement measurable
- [ ] Invalidation reliable
- [ ] Cache metrics available

---

## Success Metrics

| Metric                    | Target |
| ------------------------- | ------ |
| Search latency (cached)   | <100ms |
| Search latency (uncached) | <500ms |
| First rebuild             | ~2-5s  |
| Incremental rebuild       | <1s    |
| React UI load time        | <2s    |
| Tests coverage            | >80%   |
| Lint warnings             | 0      |

---

## Risk Mitigation

| Risk                      | Mitigation                         |
| ------------------------- | ---------------------------------- |
| React UI complexity       | Start with simple UI, iterate      |
| Daemon API conflicts      | Use `/api/docs/*` namespace        |
| Incremental indexing bugs | Comprehensive tests before deploy  |
| Redis unavailable         | Graceful fallback to no-cache      |
| Performance regression    | Benchmark before/after improvement |

---

## Timeline Estimate

| Phase           | Effort  | Timeline       |
| --------------- | ------- | -------------- |
| 4.1 React UI    | 12h     | Week 1-2       |
| 4.2 Daemon API  | 8h      | Week 2         |
| 4.3 Incremental | 8h      | Week 3         |
| 4.4 Caching     | 6h      | Week 3         |
| **Total**       | **34h** | **~2-3 weeks** |

---

## Next Steps

### Immediate

1. Setup React components structure
2. Create API client for daemon
3. Build search interface
4. Test with mock data

### Follow-up

1. Implement daemon API handlers
2. Wire React UI to real API
3. Add incremental indexing
4. Add caching layer

### Future (Phase 5+)

- Performance monitoring dashboard
- Metrics collection
- Automated doc validation in CI/CD
- User analytics

---

## Checkpoint Structure

Phase 4 will have checkpoints at:

- **4.1.4**: React UI complete + UI tests passing
- **4.2.4**: Daemon API complete + integration tests
- **4.3.4**: Incremental indexing complete + validation
- **4.4.4**: Caching complete + metrics

Each checkpoint requires all acceptance criteria met before proceeding.

---

**Ready to begin Phase 4.1: React UI Components**

Next: Start with `DocsSearch.tsx` component and API client setup.
