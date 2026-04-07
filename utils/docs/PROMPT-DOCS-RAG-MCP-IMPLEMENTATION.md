---
type: specification
id: SDD-docs-rag-mcp
title: Spec-Driven Development - Docs RAG MCP for Brain
version: 2.0.0
status: approved-for-implementation
date_created: 2026-04-03
language: en
category: architecture
related: ["ADR-0003-central-daemon-orchestration", "ADR-0005-strict-development-and-production-boundary"]
keywords: ["mcp", "rag", "documentation", "qdrant", "semantic-search", "go"]
rag_priority: critical
chunk_strategy: section
---

**Status**: READY FOR IMPLEMENTATION  
**Planning Date**: 2026-04-03  
**Decisions Made**: Go + Lazy-Load + Qdrant Native  
**Estimated Effort**: 40-50 hours  
**Next Phase**: Delegate to agent `implementer`

---

## PHASE 2: PROPOSE

### 2.1 Recommended Approach: Standalone Go MCP

**Architecture Decision**: Build `docs-rag` as a **stdio-based MCP server** (separate process).

```text
Brain Daemon (braind)
  ├─ Orchestrates services
  └─ MCPs launched globally (once)
     │
     ├─ memory MCP (stdio) ← built-in
     ├─ filesystem MCP (stdio) ← built-in
     ├─ github MCP (stdio) ← built-in
     ├─ brain-rules MCP (stdio) ← reads Brain repo
     │
     └─ docs-rag MCP (stdio) ← YOUR NEW MCP
        ├─ Language: Go
        ├─ Indexing: Lazy-load on first search
        ├─ Vector DB: Qdrant native embeddings
        ├─ Source files: reads docs/ + manifest + changelog
        └─ Tools: docs_search, docs_status, docs_rebuild
```

### 2.2 Why MCP Standalone (Not Daemon-Integrated)

| Factor | MCP Standalone | Daemon-Integrated |
|--------|---|---|
| **Architectural Pattern** | Follows brain-rules precedent ✓ | New approach ✗ |
| **Loose Coupling** | Yes ✓ | Tight ✗ |
| **Reusable in other projects** | Yes ✓ | No ✗ |
| **Version independent** | Yes ✓ | No ✗ |
| **Process overhead** | +1 process ✗ | 0 ✗ |
| **Composability** | Multiple IDEs, agents ✓ | Single daemon only ✗ |
| **Maintenance** | Simpler, focused ✓ | More daemon complexity ✗ |

**Verdict**: MCP standalone aligns with Brain's philosophy (composable, portable, focused).

---

### 2.3 Go Language Choice

**Why Go over Node.js:**
- Canonical rule: "Go-only for production internals of Brain"
- Single compiled binary (no runtime dependency)
- Fast startup (critical for stdio MCP)
- Aligns with daemon/CLI/skills manager patterns
- Can use existing daemon patterns (managers, io protocols)

**Implication**: Go binary in new folder `mcp/docs-rag-mcp/`.

---

### 2.4 Lazy-Load Indexing Strategy

**Why Lazy-Load over Nightly Pre-Build:**
- No rebuild jobs needed
- CPU efficient when no searches happening (dev workflow)
- Index builds on first search request (~2-5 seconds, then cached)
- All subsequent searches <200ms
- Qdrant holds index in memory + persists to volume

**Implication**: MCP startup is fast; first search is slow; thereafter instant.

---

### 2.5 Qdrant Native Embeddings

**Why Qdrant native over external APIs:**
- Qdrant has built-in FastEmbed model (all-small-MiniLM-L6-v2 or similar)
- Embeddings computed locally, never leave the container
- Privacy-first (no API keys, no cloud)
- Portable (works offline, works in any environment)
- Part of Qdrant container already running

**Implication**: No external dependencies; all indexing local.

---

## PHASE 3: SPECIFICATION

### 3.1 Acceptance Criteria

#### **Functional Requirements**

- [ ] **MCP Registration**: Registered in `mcp/registry.yml` with correct syntax
- [ ] **Tool: docs_search**: Accepts query string, returns ranked documents with snippets
- [ ] **Tool: docs_status**: Returns index state (docs indexed, last rebuild, health)
- [ ] **Tool: docs_rebuild**: Force rebuild of index (dev-only tool)
- [ ] **Lazy Loading**: First search triggers index build (~2-5 seconds)
- [ ] **Caching**: Subsequent searches use cached index (<200ms)
- [ ] **Manifest Compliance**: Only indexes docs matching `docs-manifest.json` structure
- [ ] **RAG Priority Boost**: Respects `rag_priority` field (critical>high>medium>low) in scoring
- [ ] **Chunk Strategy**: Respects `chunk_strategy` field (section or sentence-based splits)
- [ ] **Changelog Watching**: Detects changes via `docs-changelog.jsonl` (on rebuild)
- [ ] **Error Handling**: No silent failures; all errors logged with context

#### **Non-Functional**

- [ ] **Language**: Go (single binary, no runtime)
- [ ] **Startup Time**: <500ms (stdio server launch)
- [ ] **First Search**: <5 seconds (lazy index build)
- [ ] **Subsequent Searches**: <200ms (P95 latency)
- [ ] **Memory Usage**: <200MB (for 78 docs index in memory)
- [ ] **Test Coverage**: >80% of indexer + search logic
- [ ] **English Only**: All code, comments, error messages in English

#### **Integration Requirements**

- [ ] Works with BRAIN_ENV=development ✓
- [ ] Works with BRAIN_ENV=production ✓
- [ ] Qdrant available: normal operation ✓
- [ ] Qdrant unavailable: graceful error (retry, suggest docker restart)
- [ ] Manifest validation failures: logged, clear error messages
- [ ] Works with all 3 IDE surfaces (if integrated: daemon + CLI + UI)

---

### 3.2 MCP Tool Contracts

#### **Tool 1: docs_search**

**Input Schema:**
```json
{
  "query": "string (required, max 500 chars)",
  "limit": "number (optional, default 5, min 1, max 20)",
  "domain": "string (optional: adr, architecture, skills, testing, archive)"
}
```

**Output Schema:**
```json
{
  "type": "object",
  "properties": {
    "results": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": "string (document ID from manifest)",
          "title": "string",
          "domain": "string",
          "file_path": "string (relative to docs/)",
          "score": "number (0.0-1.0)",
          "snippet": "string (200 chars max, HTML escaped)",
          "rag_priority": "string (critical|high|medium|low)",
          "url": "string (GitHub raw URL or local file path)"
        }
      }
    },
    "metadata": {
      "type": "object",
      "properties": {
        "total_indexed": "number",
        "query_time_ms": "number",
        "index_rebuilt_at": "string (ISO timestamp or null)",
        "index_status": "string (ready|building|error)"
      }
    }
  }
}
```

**Error Response:**
```json
{
  "error": "string (human-readable)",
  "code": "string (e.g., invalid_query, qdrant_unavailable, no_results)"
}
```

**Examples:**

**Request:**
```json
{
  "query": "authentication and authorization",
  "limit": 3
}
```

**Success Response:**
```json
{
  "results": [
    {
      "id": "ADR-0005-strict-development-and-production-boundary",
      "title": "Strict Development and Production Boundary",
      "domain": "adr",
      "file_path": "docs/adr/ADR-0005-strict-development-and-production-boundary.md",
      "score": 0.87,
      "snippet": "The runtime must deny-by-default all services in production. Only BRAIN_ENV=development enables...",
      "rag_priority": "high",
      "url": "https://github.com/.../docs/adr/ADR-0005-..."
    }
  ],
  "metadata": {
    "total_indexed": 78,
    "query_time_ms": 145,
    "index_rebuilt_at": "2026-04-03T10:30:00Z",
    "index_status": "ready"
  }
}
```

---

#### **Tool 2: docs_status**

**Input:** None (no parameters)

**Output Schema:**
```json
{
  "status": "string (ready|building|degraded|error)",
  "indexed_docs": "number",
  "index_rebuilt_at": "string (ISO timestamp)",
  "next_rebuild_trigger": "string (on-next-search|nightly|manual)",
  "qdrant_status": {
    "connected": "boolean",
    "latency_ms": "number"
  },
  "manifest_validation": {
    "domains_expected": "number",
    "domains_validated": "number",
    "validation_errors": "array[string]"
  }
}
```

**Example Response:**
```json
{
  "status": "ready",
  "indexed_docs": 78,
  "index_rebuilt_at": "2026-04-03T10:30:00Z",
  "next_rebuild_trigger": "on-next-search",
  "qdrant_status": {
    "connected": true,
    "latency_ms": 12
  },
  "manifest_validation": {
    "domains_expected": 5,
    "domains_validated": 5,
    "validation_errors": []
  }
}
```

---

#### **Tool 3: docs_rebuild** (Dev-Only)

**Input:**
```json
{
  "domains": "array[string] (optional, e.g., [adr, architecture])"
}
```

**Output:**
```json
{
  "status": "string (success|in_progress|error)",
  "rebuilt_docs": "number",
  "duration_ms": "number",
  "errors": "array[string] (if any)"
}
```

**Authorization**: This tool should only be callable in BRAIN_ENV=development. In production, return error: "docs_rebuild not available in production mode."

---

### 3.3 Manifest Compliance Rules

**MCP MUST:**

1. Load manifest from `docs-manifest.json` at startup (or first search)
2. Parse YAML frontmatter of each document
3. Validate all required fields exist (type, id, title, status, date_created, language, category)
4. **Skip archived/deprecated docs** (status != "active")
5. **Respect rag_priority** in vector scoring:
   - `critical`: 1.0x multiplier
   - `high`: 0.8x multiplier
   - `medium`: 0.6x multiplier
   - `low`: 0.4x multiplier
6. **Respect chunk_strategy**:
   - `section`: Split by H2 headers (##)
   - `sentence`: Split by periods + newlines
7. Log all validation errors (missing field, invalid status, wrong domain)
8. Continue indexing other docs even if one fails (best-effort)

---

### 3.4 Error Handling & Fallback Strategy

| Scenario | Detection | Response | Fallback |
|---|---|---|---|
| Qdrant unavailable at startup | Connection timeout | Return status="degraded" | Retry on next search with exponential backoff |
| Manifest file missing | File not found | Return error: "Manifest not found" | Load default manifest structure or fail gracefully |
| Invalid frontmatter in doc | YAML parse error | Log WARNING, skip doc | Continue indexing other docs |
| Changelog.jsonl corrupted | JSONL parse error | Log ERROR | Do full filesystem rescan of docs/ folder |
| Query timeout | Search takes >10 seconds | Return partial results + warning | Suggest query refinement |
| Out of memory during indexing | OOM error | Log FATAL, return error | Suggest reducing document set or increasing memory |
| Qdrant connection lost mid-search | I/O error | Reconnect with backoff | Return error with retry hint |

**Key Principle**: Every error is logged with sufficient context. No silent failures.

---

## PHASE 4: DESIGN

### 4.1 Component Architecture

```
mcp/docs-rag-mcp/
├── main.go                    (MCP stdio server entry point)
├── go.mod & go.sum           (dependencies)
├── internal/
│   ├── indexer/
│   │   ├── indexer.go         (main Indexer struct + Load/Build methods)
│   │   ├── manifest.go        (manifest loading + validation)
│   │   ├── document.go        (Document struct + parsing)
│   │   ├── chunker.go         (text splitting per chunk_strategy)
│   │   ├── embeddings.go      (Qdrant embedding API calls)
│   │   └── indexer_test.go
│   ├── search/
│   │   ├── searcher.go        (Query handling + vector search)
│   │   ├── ranking.go         (Score adjustment per rag_priority)
│   │   └── search_test.go
│   ├── store/
│   │   ├── qdrant.go          (Qdrant client wrapper)
│   │   └── store_test.go
│   └── tools/
│       ├── docs_search.go     (MCP tool implementation)
│       ├── docs_status.go     (Status tool)
│       ├── docs_rebuild.go    (Rebuild tool)
│       └── tools_test.go
└── README.md
```

### 4.2 Data Flow: Lazy-Load Index Build

```text
1. MCP process starts (stdio server ready)
   ├─ Connect to Qdrant
   └─ Index NOT loaded yet (lazy)

2. First docs_search request arrives
   ├─ Check if index exists in Qdrant
   ├─ If NOT:
   │  ├─ Load docs-manifest.json
   │  ├─ Load docs-changelog.jsonl (get list of files)
   │  ├─ Scan docs/ folder
   │  └─ For each doc:
   │     ├─ Load + parse markdown
   │     ├─ Extract + validate frontmatter
   │     ├─ Apply chunk_strategy (section or sentence)
   │     ├─ Embed chunks via Qdrant API
   │     └─ Upsert vectors to Qdrant
   │  └─ Index built (~2-5 seconds)
   │
   ├─ Embed query string
   ├─ Query Qdrant (vector similarity + BM25)
   ├─ Re-rank results by rag_priority
   ├─ Extract snippet context
   └─ Return results to agent

3. Subsequent searches (cached index)
   ├─ Query Qdrant directly
   └─ Return results <200ms
```

### 4.3 Key Data Structures

```golang
// Document loaded from markdown
type Document struct {
    ID               string            // From filename
    Domain           string            // adr, architecture, etc.
    Title            string            // From frontmatter
    Status           string            // active, archived, deprecated
    RagPriority      string            // critical, high, medium, low
    ChunkStrategy    string            // section, sentence
    Content          string            // Full markdown body
    Chunks           []string          // Chunks after splitting
    LastIndexed      *time.Time
    ValidationErrors []string
}

// Shared index state (in-memory cache)
type Indexer struct {
    mu              sync.RWMutex
    docs            map[string]*Document  // id → doc
    isBuilding      bool
    buildStartTime  *time.Time
    qdrantClient    *qdrant.Client
    manifest        *Manifest
}

// Search result returned to agent
type SearchResult struct {
    ID          string
    Title       string
    Domain      string
    FilePath    string
    Score       float32           // 0.0-1.0
    Snippet     string            // 200 chars
    RagPriority string
    URL         string
}

// Internal Qdrant vector
type Vector struct {
    ID       string            // doc.ID + "#" + chunk#
    Values   []float32         // Embedding from Qdrant model
    Metadata map[string]string // {domain, doc_id, priority, ...}
}
```

### 4.4 Qdrant Collection Schema

**Collection Name**: `brain_docs`

**Point Schema:**
```json
{
  "id": "unique-doc-id#chunk-0",
  "vector": [...300+ floats from FastEmbed...],
  "payload": {
    "doc_id": "ADR-0005-...",
    "domain": "adr",
    "title": "Strict Development...",
    "chunk_num": 0,
    "chunk_text": "The runtime must deny-by-default...",
    "rag_priority": "high",
    "status": "active",
    "file_path": "docs/adr/ADR-0005-...",
    "url": "https://github.com/..."
  }
}
```

**Index Config** (in Qdrant):
- Vector size: ~384 (FastEmbed small model)
- Distance metric: Cosine
- Hnsw parameter: M=16, EF=200
- Search: BM25 + vector hybrid

---

### 4.5 Stdio MCP Protocol

**MCP Server** (go binary) listens on stdin/stdout for JSON-RPC 2.0.

**Call to docs_search:**
```json
{ 
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "docs_search",
    "arguments": {
      "query": "authentication",
      "limit": 5
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"results\": [...], \"metadata\": {...}}"
      }
    ]
  }
}
```

**Go library**: Use `github.com/go-echarts/mcp-server` or hand-rolled JSON-RPC handler.

---

## PHASE 5: TASKS (Implementation Breakdown)

### Task Group 1: Core Indexing Engine (8-10 hours)

**T1: Project Setup**
- Create `mcp/docs-rag-mcp/` folder
- Initialize `go.mod` with dependencies (qdrant client, etc.)
- Set up project structure (internal/indexer, internal/search, etc.)
- **Estimated**: 1 hour

**T2: Document & Manifest Types**
- Define `Document`, `Manifest`, `ManifestDomain` structs
- Implement YAML frontmatter parsing
- Validate required fields (type, id, title, status, date_created, language, category)
- **Estimated**: 1.5 hours
- **Test**: manifest_test.go (valid/invalid manifests)

**T3: Document Loader**
- Implement `LoadDocument(path string) (*Document, error)`
- Load markdown file
- Parse frontmatter YAML
- Extract body content
- Handle encoding errors, missing files
- **Estimated**: 1.5 hours
- **Test**: Test loading from sample docs in docs/ folder

**T4: Chunking Strategy**
- Implement `ChunkDocument(doc *Document) []string`
- Support "section" (split by ##) and "sentence" (split by .)
- Preserve context (snippet extraction)
- Handle edge cases (empty sections, very long chunks)
- **Estimated**: 1.5 hours
- **Test**: chunker_test.go with multiple doc types

**T5: Qdrant Integration (Client Wrapper)**
- Implement `QdrantClient` wrapper
- Connect to Qdrant @ http://localhost:6333
- Error handling for unavailable Qdrant
- Implement upsert vector logic
- Implement query logic (hybrid search)
- **Estimated**: 2 hours
- **Test**: Mock Qdrant responses; test connection retry logic

---

### Task Group 2: Indexer Core Logic (6-8 hours)

**T6: Lazy-Load Indexer Initialization**
- Implement `NewIndexer(brainRoot string) *Indexer`
- Implement `(i *Indexer) EnsureIndexBuilt(ctx context.Context) error`
- Lazy-build on first call (check if already in Qdrant, if not, build)
- Use sync.Once pattern or sync.RWMutex for thread safety
- **Estimated**: 2 hours
- **Test**: Test concurrent access (multiple searches simultaneously)

**T7: Full Index Build**
- Implement `(i *Indexer) Build(ctx context.Context) error`
- Load manifest from `docs-manifest.json`
- Scan docs/ recursively
- For each markdown:
  - Load + validate
  - Chunk
  - Embed via Qdrant
  - Upsert to Qdrant
- Log progress + errors
- **Estimated**: 2.5 hours
- **Test**: Test with actual docs/ folder; test error recovery (skip invalid docs)

**T8: Search Implementation**
- Implement `(i *Indexer) Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, error)`
- Embed query
- Query Qdrant
- Re-rank by rag_priority
- Extract snippets
- Return results
- **Estimated**: 2 hours
- **Test**: search_test.go (various queries, domain filtering, ranking)

**T9: Status & Metadata**
- Implement `(i *Indexer) GetStatus() IndexStatus`
- Return index state (ready, building, error)
- Return doc count, rebuild time
- Return Qdrant health
- **Estimated**: 1 hour

---

### Task Group 3: MCP Tools (4-5 hours)

**T10: MCP Server Scaffolding**
- Implement `main.go` stdio MCP server
- JSON-RPC 2.0 handler
- Tool registration (docs_search, docs_status, docs_rebuild)
- Error handling + logging
- **Estimated**: 2 hours
- **Test**: Test basic MCP protocol (send a tool call, get response)

**T11: docs_search Tool**
- Implement MCP tool interface for `docs_search`
- Input validation (query length, limit bounds)
- Call indexer.Search()
- Format JSON response
- **Estimated**: 1.5 hours
- **Test**: Tool request/response serialization

**T12: docs_status & docs_rebuild Tools**
- Implement `docs_status` tool (read-only)
- Implement `docs_rebuild` tool (dev-only, check BRAIN_ENV)
- **Estimated**: 1.5 hours

---

### Task Group 4: Testing (6-8 hours)

**T13: Unit Tests - Indexer**
- Test document loading + parsing
- Test chunking strategies
- Test manifest validation
- **Estimated**: 2 hours
- **Coverage target**: >85% indexer/

**T14: Unit Tests - Search & Ranking**
- Test search query embedding
- Test vector ranking
- Test rag_priority scoring
- Test domain filtering
- **Estimated**: 1.5 hours

**T15: Integration Tests - Full Flow**
- Test lazy-load (first search triggers build)
- Test subsequent searches (cached)
- Test manifest compliance
- Test error scenarios (Qdrant down, invalid docs)
- **Estimated**: 2 hours

**T16: MCP Protocol Tests**
- Test docs_search request/response
- Test docs_status
- Test docs_rebuild (dev-only enforcement)
- Test error handling
- **Estimated**: 1.5 hours

---

### Task Group 5: Brain Integration (4-5 hours)

**T17: Registry Entry & Configuration**
- File: `mcp/registry.yml` (ADD entry)
- Define `docs-rag` MCP with:
  - command: path to binary
  - type: "stdio"
  - visibility: "dev-only" (or "prod-safe" if production-ready)
  - profile: ["standard", "full"]
- **Estimated**: 45 minutes

**T18: Test with Brain CLI/Daemon**
- File: `cli/cmd/brain/docs_rag.go` (NEW)
- Implement `brain docs-rag search <query>` command
- Call MCP tool via daemon socket
- Format + display results
- **Estimated**: 1.5 hours
- **Test**: CLI integration test

**T19: Daemon Integration (if not MCP-only)**
- If adding to daemon HTTP API:
  - File: `daemon/internal/api/handlers_docs_rag.go` (NEW)
  - POST /api/docs-rag/search → call MCP → return JSON
- If keeping MCP-only:
  - Skip (MCP is sufficient)
- **Estimated**: 1-1.5 hours (skip if MCP-only approach)

**T20: UI Integration (Optional)**
- File: `desktop/src/components/DocsSearch.tsx` (NEW)
- React component for docs search
- Calls daemon API or MCP directly
- **Estimated**: 1.5 hours (optional, can defer)

---

### Task Group 6: Documentation & ADR (2-3 hours)

**T21: ADR-0006 (Write Decision Record)**
- File: `docs/adr/ADR-0006-docs-rag-mcp-architecture.md` (NEW)
- Document why MCP standalone (not daemon-integrated)
- Explain lazy-load + Qdrant native decisions
- Include consequences + trade-offs
- **Estimated**: 1 hour

**T22: Architecture & Implementation Docs**
- File: `docs/architecture/docs-rag-mcp.md` (NEW) OR update existing
- Explain how docs-RAG works
- Data flow diagram
- Configuration options
- **Estimated**: 1 hour

**T23: Update MCP README**
- File: `mcp/docs-rag-mcp/README.md` (NEW)
- Build instructions
- Configuration
- Development notes
- **Estimated**: 45 minutes

---

### Task Group 7: Build & Packaging (1-2 hours)

**T24: Build Script & Distribution**
- Create `Makefile` or script to build Go binary
- Binary location: `bin/docs-rag-mcp` (or in mcp/docs-rag-mcp/)
- Ensure binary works as stdio server
- **Estimated**: 1 hour

**T25: Validation & Cleanup**
- Remove dead code, unused dependencies
- Run `go fmt`, `go vet`
- Ensure tests pass completely
- Document any known limitations
- **Estimated**: 1 hour

---

## PHASE 5.2: Task Execution Order (Critical Path)

### **Sprint 1: Core (8-10 hours)** 
1. T1 - Project setup
2. T2 - Document types + manifest
3. T3 - Document loader
4. T4 - Chunking
5. T5 - Qdrant client

**Output**: Indexer library ready to test

---

### **Sprint 2: Indexer Logic (6-8 hours)**
6. T6 - Lazy-load initialization (test concurrent access)
7. T7 - Full index build (test with real docs/)
8. T8 - Search implementation (test ranking, filtering)
9. T9 - Status methods

**Output**: Indexer fully functional

---

### **Sprint 3: MCP Server (4-5 hours)**
10. T10 - MCP scaffolding + JSON-RPC
11. T11 - docs_search tool
12. T12 - docs_status & docs_rebuild tools

**Output**: Standalone Go MCP server ready

---

### **Sprint 4: Testing (6-8 hours)** (Run in parallel with previous)
13. T13-16 - All unit + integration + protocol tests
14. Ensure >80% coverage
15. Test error scenarios

**Output**: All tests passing, ready for integration

---

### **Sprint 5: Integration (4-5 hours)**
16. T17 - Registry entry in mcp/registry.yml
17. T18 - CLI command (brain docs-rag search)
18. T19 - (Optional) Daemon API endpoints
19. T20 - (Optional) React UI

**Output**: Users can search docs via CLI and optionally UI

---

### **Sprint 6: Documentation (2-3 hours)**
20. T21 - ADR-0006
21. T22 - Architecture docs
22. T23 - README

**Output**: Completely documented feature

---

### **Sprint 7: Build & Release (1-2 hours)**
23. T24 - Build script + binary
24. T25 - Cleanup + validation

**Output**: Ready for production

---

## PHASE 6: IMPLEMENTATION READY

### What Needs to Happen

1. **Delegated to `implementer` agent** with:
   - This full SDD
   - Task breakdown (25 tasks organized into sprints)
   - Test fixtures (sample docs from docs/ folder)
   - Acceptance criteria per task

2. **Code Review Checkpoints**:
   - After Sprint 1 (library fundamentals)
   - After Sprint 2 (full indexer)
   - After Sprint 3 (MCP server)
   - After Sprints 4-6 (integration + docs + release)

3. **Success Criteria**:
   - All tasks completed
   - All tests passing (>80% coverage)
   - No lint warnings
   - Docs match implementation
   - ADR-0006 written
   - Registered in mcp/registry.yml

---

## PHASE 7: VERIFICATION CHECKLIST

### Functional Tests
- [ ] `brain docs-rag search "daemon"` returns relevant docs >0.7 score
- [ ] `brain docs-rag search` with no results returns empty gracefully
- [ ] Domain filtering works (`search "auth" --domain adr`)
- [ ] Lazy-load: first search takes 2-5s, subsequent <200ms
- [ ] Manifest validation: invalid docs are skipped with warning logged
- [ ] rag_priority scoring: "critical" docs rank higher than "low"

### Non-Functional Tests
- [ ] Qdrant unavailable → graceful error, retry on next search
- [ ] Memory usage <200MB for index
- [ ] No silent failures (all errors logged)
- [ ] 100% English (code, comments, error messages)

### Integration Tests
- [ ] Works in single IDE
- [ ] Works in multi-IDE (shared Qdrant)
- [ ] Dev environment (lazy-load)
- [ ] Production cleanup (docker compose down -v → rebuild works)

---

## PHASE 8: HANDOFF FOR IMPLEMENTATION

**Status**: ✅ READY TO DELEGATE

**Next Command to Implementer Agent:**
```
Task: Implement Docs-RAG MCP following SDD-docs-rag-mcp

Start with Sprint 1 (Project Setup → Qdrant Integration).
Follow task execution order.
Create one commit per task group.
Run tests after each sprint.

Acceptance: All tests passing, >80% coverage, no lint warnings, ADR-0006 written.
```

---

## See Also

- `docs-manifest.json` - Documentation structure contract
- `docs-changelog.jsonl` - Change tracking (append-only log)
- `mcp/registry.yml` - Where to register docs-rag MCP
- `docs/architecture/daemon-orchestration.md` - MCP launch architecture
- `docs/adr/ADR-0003-central-daemon-orchestration.md` - MCP stdio pattern