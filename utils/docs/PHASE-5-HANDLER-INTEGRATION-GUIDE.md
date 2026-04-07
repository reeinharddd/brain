# Phase 5: Complete Handler Integration

**Goal**: Wire DocsHandler to daemon so `/api/docs/*` endpoints return real responses

**Blocks**: Cannot test search, status, or rebuild without this

---

## The Missing Link

Currently:
```
HTTP Request → handleDocs() → Returns 503 "not initialized"
```

Needed:
```
HTTP Request → handleDocs() → DocsHandler.Search()/Status()/Rebuild()
```

---

## What Needs to Happen

### 1. Create Indexer Interface

The handlers expect an indexer with these methods:
```go
type Indexer interface {
    Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, int, error)
    GetStatus() (IndexHealth, error)
    EnsureIndexBuilt(ctx context.Context) error
}
```

**Options:**
- Option A: Create stub indexer (returns empty results)
- Option B: Use MCP stdio interface (more complex)
- Option C: Create in-memory indexer (for testing)

### 2. Initialize Handler in NewBrainDaemon()

**Current** (in daemon/cmd/braind/main.go around line 2020):
```go
func NewBrainDaemon() *BrainDaemon {
    d := &BrainDaemon{
        status:      "Running",
        clients:     make(map[*websocket.Conn]bool),
        procManager: manager.NewProcessManager(logCh),
        docker:      manager.NewDockerManager(...),
        // ... other managers ...
    }
    // No docsHandler initialization
    return d
}
```

**Add**:
```go
// Create handler with stub indexer (for now)
d.docsHandler = handlers.NewDocsHandler(
    &StubIndexer{},
    d.environment,
)
```

### 3. Update handleDocs() to Use Handler

**Current** (temporary):
```go
func (d *BrainDaemon) handleDocs(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    if strings.HasPrefix(r.URL.Path, "/api/docs/search") && r.Method == "GET" {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{"error": "docs search not yet initialized"})
        return
    }
    // ... etc
}
```

**New** (call actual handler):
```go
func (d *BrainDaemon) handleDocs(w http.ResponseWriter, r *http.Request) {
    if d.docsHandler == nil {
        http.Error(w, "docs handler not initialized", http.StatusInternalServerError)
        return
    }
    
    if strings.HasPrefix(r.URL.Path, "/api/docs/search") && r.Method == "GET" {
        d.docsHandler.Search(w, r)
        return
    }
    
    if strings.HasPrefix(r.URL.Path, "/api/docs/status") && r.Method == "GET" {
        d.docsHandler.Status(w, r)
        return
    }
    
    if strings.HasPrefix(r.URL.Path, "/api/docs/rebuild") && r.Method == "POST" {
        d.docsHandler.Rebuild(w, r)
        return
    }
    
    http.NotFound(w, r)
}
```

---

## Immediate Action: Create Stub Indexer

For testing, create a **stub indexer** in a new file:

**File**: `daemon/cmd/braind/stub_indexer.go`

```go
package main

import (
	"context"
	"fmt"
)

type StubIndexer struct{}

// SearchResult from handlers
type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Domain    string  `json:"domain"`
	Score     float64 `json:"score"`
	Path      string  `json:"path"`
	ChunkID   string  `json:"chunk_id"`
}

// IndexHealth from handlers
type IndexHealth struct {
	Status     string `json:"status"`
	DocCount   int    `json:"doc_count"`
	ChunkCount int    `json:"chunk_count"`
	LastRebuild string `json:"last_rebuild"`
	RebuildTime int    `json:"rebuild_time_ms"`
	Healthy    bool   `json:"healthy"`
}

func (s *StubIndexer) Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, int, error) {
	// Return mock results for testing
	return []SearchResult{
		{
			ID:      "mock-1",
			Title:   fmt.Sprintf("Mock Result for: %s", query),
			Content: "This is a stub indexer response for testing.",
			Domain:  domain,
			Score:   0.95,
			Path:    "docs/testing.md",
			ChunkID: "chunk-1",
		},
	}, 1, nil
}

func (s *StubIndexer) GetStatus() (IndexHealth, error) {
	return IndexHealth{
		Status:      "ready",
		DocCount:    78,
		ChunkCount:  512,
		LastRebuild: "2026-04-03T23:50:00Z",
		RebuildTime: 2400,
		Healthy:     true,
	}, nil
}

func (s *StubIndexer) EnsureIndexBuilt(ctx context.Context) error {
	return nil
}
```

---

## Integration Steps (5-10 minutes)

1. **Create stub_indexer.go** - Copy code above
2. **Update NewBrainDaemon()** - Add handler initialization
3. **Update handleDocs()** - Call actual handler methods
4. **Rebuild daemon** - `cd daemon/cmd/braind && go build -o ../../../bin/braind main.go`
5. **Test** - `bash ~/.brain/MANUAL-TEST-DOCS.sh`

---

## Expected Test Results After Integration

```bash
$ bash ~/.brain/MANUAL-TEST-DOCS.sh

[Step 4] Testing /api/docs endpoints...

  4a) GET /api/docs/status
       HTTP Code: 200
       Response: {"status":"ready","doc_count":78,"chunk_count":512,...}
       ✅ Endpoint is accessible (200 = OK)

  4b) GET /api/docs/search?q=architecture
       HTTP Code: 200
       Response: [{"id":"mock-1","title":"Mock Result for: architecture",...}]
       ✅ Endpoint is accessible (200 = OK)

  4c) POST /api/docs/rebuild
       HTTP Code: 200
       Response: {"status":"rebuild_complete"}
       ✅ Endpoint is accessible (200 = OK)
```

---

## Files to Modify

| File | Change | Complexity |
|------|--------|-----------|
| `daemon/cmd/braind/main.go` | Update handleDocs(), initialize handler in NewBrainDaemon() | Low |
| `daemon/cmd/braind/stub_indexer.go` | **Create new file** | Low |
| Compile | `go build` | Auto |

---

## Decision: Stub vs Real

### Stub Indexer (Quick)
- ✅ Fast to implement (5 min)
- ✅ Tests routing works
- ✅ Tests handler calls succeed
- ❌ Doesn't test actual search functionality
- ❌ Doesn't test caching
- ❌ Doesn't test Qdrant integration

### Real Indexer (Complete)
- ❌ Slower (30-60 min)
- ✅ Tests full pipeline
- ✅ Tests caching works
- ✅ Tests Qdrant integration
- ✅ Ready for React UI

### Recommendation

**Phase 5a** (Now): Use **Stub Indexer** to verify routing
**Phase 5b** (Later): Replace with **Real Indexer** connected to MCP/Qdrant

---

## Next After Integration

Once handleDocs() works:
1. **React UI Testing** - `bun tauri dev` (desktop/)
2. **Search Functionality** - Query architecture, security, etc.
3. **Caching Validation** - 2nd search <200ms
4. **Rebuild Endpoint** - Force reindex
5. **Full End-to-End** - UI → API → Qdrant

---

## Troubleshooting

If tests still fail after integration:

```bash
# Check daemon is running
ps aux | grep braind

# View logs
tail -f /tmp/braind.log

# Kill and restart
pkill -9 braind ; cd ~/.brain/daemon/cmd/braind && go run main.go

# Test with verbose curl
curl -v http://localhost:9090/api/docs/status
```

---

## Files Modified Summary

After this phase:
```
✅ daemon/cmd/braind/main.go
   ├─ NewBrainDaemon() - Initialize docsHandler
   └─ handleDocs() - Call handler methods

✅ daemon/cmd/braind/stub_indexer.go
   ├─ StubIndexer struct
   └─ Search, Status, EnsureIndexBuilt methods

✅ Compiled binary
   └─ ~/.brain/bin/braind (updated)
```

---

## Status After Completion

```
HTTP Request
    ↓
ServeHTTP() → /api/docs*
    ↓
handleDocs()
    ↓
    ├─ Search() → StubIndexer.Search() → 200 OK
    ├─ Status() → StubIndexer.Status() → 200 OK
    └─ Rebuild()→ StubIndexer.EnsureIndexBuilt() → 200 OK
```

**Ready for**: React UI integration
**Next**: Replace stub with real Qdrant indexer
