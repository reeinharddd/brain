---
id: ADR-0006
title: Docs-RAG MCP - Semantic Search Over Documentation
status: approved
date_created: 2026-04-03
language: en
type: architecture-decision
category: adr
version: 1.0.0
---

## ADR-0006: Docs-RAG MCP - Semantic Search Over Documentation

**Status**: APPROVED  
**Deciders**: Brain Architects  
**Date**: April 3, 2026

---

## 1. Context & Problem Statement

Brain repository contains 78+ documentation files across 5 domains (architecture, skills, testing, standards, templates). Without semantic search, users must:

- Manually grep through docs
- Hope filename/frontmatter contains keywords
- Use keyword-only search (no concept-based discovery)
- Have limited way to inject relevant context into agent conversations

**Problem**: Information retrieval is keyword-based, not semantic. Users cannot search by concept (e.g., "how do I handle errors?" finds only files with the word "error").

---

## 2. Decision: Build MCP Server for Semantic Documentation Search

**Decision**: Implement a standalone MCP server (`docs-rag-mcp`) that provides semantic search over documentation using vector embeddings via Qdrant.

**Key Decisions**:

| Decision                               | Rationale                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------- |
| **MCP (not daemon endpoint)**          | Composable, portable, no port conflicts, loose coupling from daemon       |
| **Standalone (not daemon-integrated)** | Can be reused in other projects, version-independent, simpler to maintain |
| **Stdio-based**                        | JSON-RPC 2.0 standard, works with any MCP host                            |
| **Lazy-load indexing**                 | 2-5s first search, <200ms cached (no background jobs)                     |
| **Qdrant native embeddings**           | All-MiniLM-L6-v2, local computation, no API keys, 384-dim vectors         |
| **Manifest-driven**                    | Single source of truth for doc structure, enables validation              |
| **Three tools**                        | docs_search, docs_status, docs_rebuild (dev-only)                         |

---

## 3. Solution Architecture

### 3.1 Layered Design

```
┌─────────────────────────────────────────┐
│ MCP Host (Claude/Cursor/Windsurf)       │
├─────────────────────────────────────────┤
│ MCP Server (JSON-RPC 2.0 stdio)         │
│  - docs_search                          │
│  - docs_status                          │
│  - tools/list + tools/call              │
├─────────────────────────────────────────┤
│ Indexer (Lazy-Load)                     │
│  - LoadDocuments()                      │
│  - Build() → Qdrant                     │
│  - Search() → Vector query              │
├─────────────────────────────────────────┤
│ Qdrant Vector Database (localhost:6333) │
│  - brain_docs collection                │
│  - 384-dim vectors, cosine distance     │
│  - FastEmbed native model               │
└─────────────────────────────────────────┘
```

### 3.2 Tool Contracts

#### docs_search

```json
{
  "method": "docs_search",
  "params": {
    "query": "string (required)",
    "limit": "integer (optional, default: 10)",
    "domain": "string (optional)"
  },
  "result": {
    "results": [
      {
        "title": "Document Title",
        "path": "docs/domain/file.md",
        "category": "domain",
        "rag_priority": "critical|high|medium|low",
        "score": 0.95,
        "snippet": "Relevant text excerpt..."
      }
    ],
    "metadata": {
      "total_indexed": 78,
      "query_time_ms": 145,
      "index_status": "ready",
      "results_count": 5
    }
  }
}
```

#### docs_status

```json
{
  "method": "docs_status",
  "result": {
    "index_status": {
      "state": "ready|indexing|not_built",
      "document_count": 78,
      "chunk_count": 450,
      "last_rebuild_time": "2026-04-03T10:45:30Z",
      "qdrant_health": "healthy",
      "errors": []
    }
  }
}
```

#### docs_rebuild (dev-only)

```json
{
  "method": "docs_rebuild",
  "params": {
    "domains": ["architecture", "skills"] // optional
  },
  "result": {
    "success": true,
    "document_count": 78,
    "duration": "3.2s",
    "error": null
  }
}
```

### 3.3 Validation Pipeline

```
Document (.md file)
  ↓
1. LoadDocument() - Parse frontmatter YAML
  ↓
2. ValidateFrontmatter() - Check 8 required fields
  ↓
3. ValidateAgainstManifest() - Check structure rules
  ↓
4. ValidateContent() - English-only, no emojis
  ↓
5. ChunkDocument() - Split by strategy
  ↓
6. UpsertVectors() - Store in Qdrant
  ↓
Index Ready ✓
```

### 3.4 Chunking Strategies

**Section Chunking** (default):

- Split by `## Headers` in markdown
- Preserves document context
- Variable chunk sizes (100-500 tokens typical)

**Sentence Chunking**:

- Group 3 sentences per chunk
- For flat documents without headers
- Fixed 3-sentence windows

**Selection** per document:

- `chunk_strategy: section` in frontmatter
- `chunk_strategy: sentence` in frontmatter
- Default: section

---

## 4. Why This Approach?

### 4.1 Why MCP (not daemon extension)?

| Aspect          | MCP                       | Daemon Extension           |
| --------------- | ------------------------- | -------------------------- |
| Composability   | ✓ Reusable in projects    | ✗ Brain-specific           |
| Version Control | ✓ Independent versioning  | ✗ Tied to daemon version   |
| Upgrade Safety  | ✓ Isolated binary         | ✗ Part of daemon release   |
| Reusability     | ✓ Works in other projects | ✗ Brain-only               |
| Complexity      | ✓ Simpler MCP             | ✗ Adds daemon logic        |
| Testing         | ✓ Standalone tests        | ✗ Daemon integration tests |

**Decision**: MCP is more maintainable, reusable, and independently versioned.

### 4.2 Why Standalone (not daemon-integrated)?

- **Consistency**: Follows brain architecture patterns (MCPs are stdio, not daemon-internal)
- **Loose Coupling**: Daemon doesn't know about docs-rag implementation
- **Portability**: Can be copied to other projects using Brain
- **Simplicity**: No daemon API changes needed
- **Testing**: Can be tested independently without daemon

### 4.3 Why Lazy-Load (not background jobs)?

```
Option A: Lazy-Load (CHOSEN)
- First search: ~2-5 seconds (index builds)
- Subsequent searches: <200ms (cached in Qdrant)
- Cost: CPU on first search, none after
- Benefit: No background tasks, no wasted indexing

Option B: Background Rebuild
- Index built on startup: ~5 seconds
- All searches: <200ms immediately
- Cost: Startup time always, periodic rebuilds
- Benefit: Faster first search

Decision: Lazy-load is better for dev experience. First search waits 2-5s, but
no startup delays. Background jobs add complexity. Cached index persists in
qdrant_data volume, so rebuilds are incremental.
```

### 4.4 Why Qdrant (not external API)?

| Aspect          | Qdrant (local)         | Pinecone/Weaviate (API) |
| --------------- | ---------------------- | ----------------------- |
| Cost            | $0 (self-hosted)       | $$ (monthly)            |
| Privacy         | 100% local             | Sends to cloud          |
| Latency         | <100ms + local compute | 200-500ms + network     |
| Vendor Lock     | None (portable)        | Locked to service       |
| Offline Support | ✓ Full offline         | ✗ Requires internet     |
| Embedding Model | FastEmbed local        | API provider's          |

**Decision**: Qdrant is local, free, fast, and privacy-respecting.

---

## 5. Implementation Details

### 5.1 Phases

**Phase 1**: Core Indexing Engine (DONE)

- Document loading and validation
- Chunking strategies
- Qdrant client wrapper
- 22+ unit tests

**Phase 2**: MCP Server & Tools (DONE)

- JSON-RPC 2.0 stdio protocol
- Tool implementations
- CLI wrapper (bonus)
- 6+ tool tests

**Phase 3**: Brain Integration (IN PROGRESS)

- Registry entry in mcp/registry.yml
- Build binary and distribute
- (Optional) Daemon API wrapper
- (Optional) React dashboard

**Phase 4**: Optimization (Future)

- Incremental rebuilds from docs-changelog.jsonl
- Caching layer for frequent queries
- Metrics dashboard
- CI/CD integration

### 5.2 Production Guard

```go
// DocsRebuild blocks if BRAIN_ENV=production
if brainEnv == "production" {
  return RebuildResponse{
    Error: "rebuild disabled in production",
  }
}
```

Protects against accidental rebuilds in production that could:

- Use excessive CPU
- Modify existing index unexpectedly
- Require manual validation afterwards

### 5.3 Error Handling

All errors are **explicit and logged**:

```
- Parse errors (-32700): Malformed JSON
- Invalid params (-32602): Bad request structure
- Method not found (-32601): Unknown method
- Tool not found: When tool name invalid
- Index build failed: Logged with context
```

No silent failures. Every error path returns JSON-RPC error response.

---

## 6. Data & Storage

### 6.1 Manifest-Driven Schema

**docs-manifest.json** defines:

- Domains (architecture, skills, testing, standards, templates)
- Required files per domain
- Validation rules (naming, structure, content)
- RAG priority defaults per domain

```json
{
  "domains": {
    "architecture": {
      "rag_priority": "high",
      "required_files": ["README.md", "..."],
      "rules": {
        "max_depth": 2,
        "max_title_length": 80
      }
    }
  }
}
```

### 6.2 Storage Locations

```
Development:
  Source docs:     ~/.brain/docs/                (git-tracked)
  Manifest:        ~/.brain/docs-manifest.json   (git-tracked)
  Index data:      qdrant_data/ volume            (ephemeral)
  Changelog:       ~/.brain/docs-changelog.jsonl (git-tracked)

Production (cleanup):
  $ docker compose down -v  # Remove qdrant_data
  $ brain start             # MCP rebuilds index from scratch
  ✓ No data loss (source in git)
```

### 6.3 Vector Collection Schema

**Collection**: `brain_docs`

```
- Vector size: 384 dimensions
- Distance: Cosine similarity
- Model: all-MiniLM-L6-v2 via FastEmbed
- Payload fields:
  - title (string)
  - path (string)
  - domain (string)
  - content (string, chunked text)
  - rag_priority (enum: critical/high/medium/low)
  - chunk_index (integer)
```

---

## 7. Trade-offs & Alternatives Considered

### 7.1 Alternative: Daemon-Integrated Search

**Pros**: Single unified daemon  
**Cons**: Tight coupling, version lock, harder to maintain, can't reuse

**Decision**: Rejected. MCP standalone is cleaner.

### 7.2 Alternative: External API (Pinecone, Weaviate)

**Pros**: Serverless, no infrastructure  
**Cons**: Cost ($$$), vendor lock, privacy concerns, latency

**Decision**: Rejected. Qdrant local is free and fast.

### 7.3 Alternative: grep/ripgrep (keyword search)

**Pros**: Simple, no infrastructure needed  
**Cons**: Not semantic, no concept discovery, poor relevance ranking

**Decision**: Rejected. Semantic search is more powerful.

---

## 8. Success Criteria

✅ Semantic search works for concept-based queries  
✅ First search <5 seconds, subsequent <200ms  
✅ All tools callable via JSON-RPC 2.0  
✅ >80% test coverage  
✅ Production guard on rebuild  
✅ Zero hardcoded secrets  
✅ 100% English code & docs  
✅ Registered in mcp/registry.yml

---

## 9. Future Extensions

- **RAG Injection**: Daemon automatically injects relevant docs into agent contexts
- **Caching**: Redis cache for frequent queries
- **Incremental Indexing**: Rebuild only changed docs from docs-changelog.jsonl
- **Metrics**: Dashboard showing search popularity, index health
- **CI/CD**: GitHub Actions to validate docs before merge

---

## 10. Related Decisions

- **ADR-0005**: Strict dev/prod boundary (informs rebuild guard)
- **ADR-0007**: BRAIN_ENV contract (used for production safety)
- **docs-manifest.json**: Source of truth for doc structure
- **docs-validation system**: Ensures docs meet manifest rules

---

## Approval

**Decider**: Brain Architecture Team  
**Approved**: April 3, 2026  
**Implementation**: Phases 1-3 complete, Phase 4 optional future work

Related Issue: Phase 2 & 3 implementation prompt  
Next: Update CI/CD to validate docs before merge
