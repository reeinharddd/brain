---
type: prompt
id: implementation-docs-rag-mcp
title: Implementation Prompt - Docs-RAG MCP for Brain
version: 2.0.0
status: active
date_created: 2026-04-03
language: en
category: development
---

# Implementation Prompt: Docs-RAG MCP for Brain

**Status**: Ready to Delegate to `implementer` Agent  
**Estimated Effort**: 40-50 hours  
**Date**: 2026-04-03

---

## CONTEXT: What Are You Building?

A **Docs-RAG MCP Server** that provides semantic search over Brain's documentation.

### Quick Facts

- **Type**: Model Context Protocol (MCP) server - standalone process
- **Language**: Go (binary)
- **Architecture**: Stdio-based, launches globally (once), all IDEs share
- **Indexing**: Lazy-load on first search (~2-5 seconds, then <200ms cached)
- **Vector DB**: Qdrant (native embeddings, local, no API keys)
- **Source**: Reads from `docs/` folder + `docs-manifest.json` + `docs-changelog.jsonl`

### Key Integration Points

1. Register in `mcp/registry.yml`
2. Implement CLI command: `brain docs-rag search <query>`
3. Expose 3 MCP tools: `docs_search`, `docs_status`, `docs_rebuild`
4. Respect Brain's doc structure (manifest, RAG priority, chunk strategy)

---

## DECISIONS (FINAL - Do Not Change)

| Decision | Choice | Why |
|----------|--------|-----|
| **Language** | Go | Canonical rule: Go-only for production internals |
| **Architecture** | MCP Standalone (stdio) | Follows brain-rules pattern; reusable, composable |
| **Indexing** | Lazy-load on first search | No rebuild jobs; CPU efficient; first search ~2-5s, rest <200ms |
| **Embeddings** | Qdrant native FastEmbed | Local, private, no API keys, portable |
| **Storage** | Ephemeral dev / Volatile prod | Rebuild from git-tracked docs/ on startup; no data loss |

---

## IMPLEMENTATION PHASES (DO THIS IN ORDER)

### Phase 1: Core Indexing Engine (Sprint 1-2, ~14-18 hours)

**Deliverable**: Standalone Go library that loads docs, validates manifest, builds Qdrant index, performs searches.

#### Task 1.1: Project Setup
- Create `mcp/docs-rag-mcp/` folder structure
- Initialize `go.mod` with dependencies (qdrant-go client, yaml, etc.)
- Create internal/indexer, internal/search, internal/store directories
- **Estimated**: 1 hour
- **Success**: `go mod tidy` works, project compiles

#### Task 1.2: Document & Manifest Types
- Define Go structs: `Document`, `Manifest`, `ManifestDomain`
- Implement YAML frontmatter parsing
- Validate required fields (type, id, title, status, date_created, language, category)
- **Estimated**: 1.5 hours
- **Success**: Can parse `docs-manifest.json` without errors, validates frontmatter

#### Task 1.3: Document Loader
- Implement `LoadDocument(path string) (*Document, error)`
- Load markdown file, parse frontmatter, extract body
- Handle errors: file not found, invalid YAML, missing fields
- **Estimated**: 1.5 hours
- **Success**: Can load any doc from docs/ folder, errors logged

#### Task 1.4: Document Chunking
- Implement `ChunkDocument(doc *Document) []string`
- Support "section" strategy (split by ## headers)
- Support "sentence" strategy (split by periods + newlines)
- **Estimated**: 1.5 hours
- **Success**: Can chunk all doc types, preserves context

#### Task 1.5: Qdrant Client Wrapper
- Implement connection to Qdrant @ http://localhost:6333
- Implement `UpsertVector(id, content, metadata)` method
- Implement `QueryVectors(query, limit)` method
- Handle connection failures gracefully (retry with backoff)
- **Estimated**: 2 hours
- **Success**: Can connect to Qdrant, upsert + query vectors work

#### Task 1.6: Lazy-Load Indexer Initialization
- Implement `NewIndexer(brainRoot string) *Indexer` struct
- Implement `(i *Indexer) EnsureIndexBuilt(ctx context.Context) error`
- Lazy-build on first call; check if already indexed in Qdrant
- Use sync.RWMutex for thread safety (multiple searches concurrently)
- **Estimated**: 2 hours
- **Success**: Index builds on first search, subsequent searches use cache

#### Task 1.7: Full Index Build
- Implement `(i *Indexer) Build(ctx context.Context) error`
- Load manifest from `docs-manifest.json`
- Scan docs/ recursively
- For each markdown: load, validate, chunk, embed, upsert
- Log progress + handle errors (skip invalid docs, continue with others)
- **Estimated**: 2.5 hours
- **Success**: Can index all 78 docs in <5 seconds, respects manifest

#### Task 1.8: Search Implementation
- Implement `(i *Indexer) Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, error)`
- Embed query via Qdrant
- Query Qdrant (hybrid BM25 + vector)
- Re-rank results by `rag_priority` (critical>high>medium>low)
- Extract snippet (200 chars context)
- **Estimated**: 2 hours
- **Success**: Search returns relevant docs ranked correctly

#### Task 1.9: Status & Metadata
- Implement `(i *Indexer) GetStatus() IndexStatus`
- Return: index state (ready/building), doc count, rebuild time, Qdrant health
- **Estimated**: 1 hour
- **Success**: Status endpoint returns all metadata

#### Task 1.10: Tests for Core Logic (~4 hours, run concurrently)
- Test manifest loading + validation
- Test document parsing + chunking
- Test vector embedding + ranking
- Test error handling (Qdrant down, invalid docs, etc.)
- **Target Coverage**: >85% of indexer/ logic
- **Success**: `go test ./... -cover` shows >85%

**Phase 1 Checkpoint**: Before proceeding, verify:
- [ ] All 9 core tasks working
- [ ] Tests passing
- [ ] Can index docs/ folder end-to-end
- [ ] Index builds in <5 seconds
- [ ] Searches return results <200ms after index built

---

### Phase 2: MCP Server & Tools (Sprint 3, ~4-5 hours)

**Deliverable**: Standalone stdio MCP server that exposes 3 tools (docs_search, docs_status, docs_rebuild).

#### Task 2.1: MCP Stdio Server Scaffolding
- Create `main.go` entry point
- Implement JSON-RPC 2.0 handler (read stdin, write stdout)
- Register tools (docs_search, docs_status, docs_rebuild)
- Error handling + logging to stderr
- **Estimated**: 2 hours
- **Success**: Can send tool call via stdin, get response via stdout

#### Task 2.2: docs_search Tool
- Input: `{query: string, limit?: int, domain?: string}`
- Call indexer.Search()
- Output: `{results: [...], metadata: {...}}`
- Validate inputs (query length, limit bounds)
- **Estimated**: 1.5 hours
- **Success**: MCP tool works, returns correctly formatted JSON

#### Task 2.3: docs_status & docs_rebuild Tools
- Implement `docs_status` (read-only status)
- Implement `docs_rebuild` (dev-only, check BRAIN_ENV)
- Block rebuild in production
- **Estimated**: 1.5 hours
- **Success**: Tools work, dev-only enforcement active

#### Task 2.4: MCP Protocol Tests (~1 hour)
- Test tool call serialization/deserialization
- Test error responses
- Test "tools/list" capability
- **Success**: `echo '...' | ./docs-rag-mcp` works

**Phase 2 Checkpoint**: Before proceeding, verify:
- [ ] MCP binary builds and runs
- [ ] Can call docs_search via stdin/stdout
- [ ] Can call docs_status
- [ ] docs_rebuild blocked in production
- [ ] Error handling is graceful

---

### Phase 3: Brain Integration (Sprint 5, ~4-5 hours)

**Deliverable**: Register MCP in Brain, implement CLI command, optional daemon API + UI.

#### Task 3.1: Registry Entry
- File: `mcp/registry.yml` (ADD entry)
- Define `docs-rag` MCP with:
  - `command`: path to Go binary
  - `type`: "stdio"
  - `visibility`: "dev-only" or "prod-safe"
  - `profile`: ["standard", "full"]
  - `purpose`: "Semantic search over Brain documentation"
- **Estimated**: 45 minutes
- **Success**: Entry properly formatted, no syntax errors

#### Task 3.2: CLI Command
- File: `cli/cmd/brain/docs_rag.go` (NEW)
- Implement `brain docs-rag search <query>` command
- Parse flags: `--limit`, `--domain`
- Call MCP tool via stdio or daemon
- Format + display results in CLI
- **Estimated**: 1.5 hours
- **Success**: `brain docs-rag search "authentication"` works

#### Task 3.3: Daemon API Integration (Optional)
- File: `daemon/internal/api/handlers_docs_rag.go` (NEW)
- Add HTTP endpoint: `POST /api/docs-rag/search`
- Daemon calls MCP tool, returns JSON
- **Estimated**: 1-1.5 hours
- **Success**: Agents can call daemon API to search

#### Task 3.4: React UI Component (Optional)
- File: `desktop/src/components/DocsSearch.tsx` (NEW)
- Search input + results display
- Call daemon or MCP
- **Estimated**: 1.5 hours
- **Success**: UI searches and displays results

**Phase 3 Checkpoint**: Before proceeding, verify:
- [ ] Registry entry correct
- [ ] `brain docs-rag search` works via CLI
- [ ] Results display correctly
- [ ] (Optional) Daemon API works
- [ ] (Optional) UI works

---

### Phase 4: Documentation & Testing (Sprint 4+6, ~8-10 hours)

**Deliverable**: Complete tests (>80% coverage), ADR-0006, updated docs.

#### Task 4.1: Comprehensive Testing
- Unit tests: manifest, document loading, chunking, search ranking
- Integration tests: end-to-end index build + search
- Protocol tests: MCP stdio tool contracts
- Error tests: Qdrant down, invalid manifest, malformed queries
- **Target Coverage**: >80%
- **Estimated**: 4-5 hours
- **Success**: `go test ./... -cover` shows >80%, all tests pass

#### Task 4.2: ADR-0006 (Architecture Decision Record)
- File: `docs/adr/ADR-0006-docs-rag-mcp-architecture.md` (NEW)
- Document decision: Why MCP standalone (not daemon-integrated)
- Explain lazy-load + Qdrant native choices
- Include: Context, Options, Decision, Rationale, Consequences
- **Estimated**: 1 hour
- **Success**: ADR follows Brain's template, decision clearly justified

#### Task 4.3: Architecture Documentation
- File: `docs/architecture/docs-rag-mcp.md` (NEW)
- Explain: How docs-RAG works, data flow, component architecture
- Include: Lazy-load flow, Qdrant schema, error handling
- **Estimated**: 1 hour
- **Success**: Doc explains implementation clearly

#### Task 4.4: MCP README
- File: `mcp/docs-rag-mcp/README.md` (NEW)
- Build instructions
- Configuration
- Development notes
- **Estimated**: 45 minutes
- **Success**: Developer can understand and build MCP from README

#### Task 4.5: Code Quality
- Run `go fmt`, `go vet`
- Remove dead code, unused dependencies
- Ensure all error messages are in English
- **Estimated**: 1 hour
- **Success**: No warnings, clean code

**Phase 4 Checkpoint**: Before shipping, verify:
- [ ] All tests passing
- [ ] >80% coverage
- [ ] No lint warnings
- [ ] ADR-0006 written
- [ ] Docs updated

---

### Phase 5: Build & Release (Sprint 7, ~1-2 hours)

**Deliverable**: Production-ready Go binary + packaging.

#### Task 5.1: Build Script
- Create `Makefile` or `scripts/build-docs-rag-mcp.sh`
- Compile Go binary
- Output to `bin/docs-rag-mcp` or `mcp/docs-rag-mcp/bin/`
- **Estimated**: 1 hour
- **Success**: `make build` produces working binary

#### Task 5.2: Final Validation
- Verify binary works as stdio MCP
- Verify all 3 tools callable
- Verify no secrets in binary
- Cleanup temp/test files
- **Estimated**: 1 hour
- **Success**: Production binary ready

---

## ACCEPTANCE CRITERIA (All Must Pass)

### Functional

- [ ] `brain docs-rag search "daemon"` returns ranked results (score >0.7)
- [ ] Empty results handled gracefully
- [ ] Domain filtering works
- [ ] Lazy-load triggers on first search (~2-5s)
- [ ] Subsequent searches <200ms
- [ ] Manifest validation: invalid docs skipped with warning
- [ ] RAG priority scoring: "critical" ranks higher than "low"
- [ ] Chunk strategy respected: "section" and "sentence" both work

### Non-Functional

- [ ] Language: Go (single binary, no runtime)
- [ ] Tests: >80% coverage, all passing
- [ ] Memory: <200MB for index
- [ ] Startup: <500ms MCP launch
- [ ] Logging: All errors logged with context (no silent failures)
- [ ] English: 100% (code, comments, errors)

### Integration

- [ ] Works with BRAIN_ENV=development
- [ ] Works with BRAIN_ENV=production
- [ ] Qdrant unavailable → graceful error + retry
- [ ] Manifest invalid → clear error message
- [ ] Registered in `mcp/registry.yml`
- [ ] CLI command works: `brain docs-rag search`

---

## REFERENCE DOCUMENTS

**See these for detailed design**:

1. **Full SDD**: `/home/reeinharrrd/.brain/PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md`
   - Phases 2-8 (Propose, Spec, Design, full task breakdown)
   - Component architecture diagrams
   - Data flow details
   - MCP tool contract JSON schemas

2. **Current Architecture**:
   - `mcp/registry.yml` - How to register MCP
   - `docs/architecture/daemon-orchestration.md` - MCP Launch pattern
   - `docs-manifest.json` - Documentation structure contract
   - `docs-changelog.jsonl` - Change tracking format

3. **Existing Patterns** (follow these):
   - `mcp/brain-mcp-server/README.md` - Self-referential MCP pattern
   - `daemon/internal/manager/skills.go` - Manager pattern (borrow structure)
   - `cli/cmd/brain/mcp.go` - CLI tool pattern

---

## GETTING STARTED

### 1. Create the project structure
```bash
mkdir -p mcp/docs-rag-mcp/internal/{indexer,search,store,tools}
cd mcp/docs-rag-mcp
go mod init github.com/reeinharrrd/brain/mcp/docs-rag-mcp
```

### 2. Understand what you're indexing
```bash
# Read the manifest to understand doc structure
cat docs-manifest.json | head -100

# Check changelog format
head -5 docs-changelog.jsonl

# Browse docs/ to see what you're indexing
ls -la docs/{adr,architecture,skills,testing}/
```

### 3. Start with Phase 1 (Core Indexing)
- Begin with Tasks 1.1-1.5 (setup, types, document loader, chunking, Qdrant)
- Write tests as you go
- Verify Phase 1 Checkpoint before moving to Phase 2

### 4. Test Early & Often
```bash
go test ./... -cover -v
```

### 5. Reference the full SDD if blocked
- Detailed design in `/home/reeinharrrd/.brain/PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md`
- Component architecture section 4.1-4.5
- Task breakdown section 5.1-5.6

---

## TIME EXPECTATIONS

| Phase | Tasks | Hours | Checkpoint |
|-------|-------|-------|-----------|
| 1 | Core indexing (9 tasks + tests) | 14-18 | Index builds & searches work |
| 2 | MCP server (3 tools + tests) | 4-5 | Stdio MCP binary works |
| 3 | Brain integration (4 tasks) | 4-5 | CLI command works |
| 4 | Tests + docs + ADR (5 tasks) | 8-10 | >80% coverage, ADR-0006 written |
| 5 | Build + release (2 tasks) | 1-2 | Production binary ready |
| **TOTAL** | **23 tasks** | **40-50** | **Feature complete** |

---

## QUESTIONS? KNOWN ISSUES?

If you get stuck:

1. **Check the full SDD**: `/home/reeinharrrd/.brain/PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md`
   - Section 4 has detailed component architecture
   - Section 3 has full tool contract specifications

2. **Review existing patterns**:
   - `brain-rules MCP` for self-referential pattern
   - `skills.go` for manager pattern
   - `daemon-orchestration.md` for MCP lifecycle

3. **Verify Qdrant is running**:
   ```bash
   docker ps | grep qdrant
   curl http://localhost:6333/collections
   ```

4. **Test with sample query**:
   ```bash
   brain docs-rag search "authentication" --limit 3
   ```

---

## SUCCESS = Ship Day

When all tasks complete and acceptance criteria pass:

1. MCP binary works standalone ✓
2. CLI command `brain docs-rag search` works ✓
3. Tests >80%, all passing ✓
4. No lint warnings ✓
5. ADR-0006 + docs complete ✓
6. Registered in `mcp/registry.yml` ✓

You're done. Feature ships.

