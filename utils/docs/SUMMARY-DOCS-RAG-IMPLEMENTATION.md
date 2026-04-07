---
type: summary
id: docs-rag-implementation-summary
title: Docs-RAG MCP Implementation - Complete Summary
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: reference
---

# Docs-RAG MCP Implementation: Complete Summary

**Date**: April 3, 2026  
**Status**: Ready for Implementation  
**Effort**: 40-50 hours  

---

## What You're Building

A **Docs-RAG MCP Server** that provides semantic search over Brain's documentation using Qdrant.

```
User Query: "How does daemon initialization work?"
         ↓
    MCP Tool Call
         ↓
    Indexer + Qdrant
         ↓
    Ranked Results (with RAG priority)
         ↓
    Agent gets better context
```

---

## 4 THINGS YOU NEED TO KNOW

### 1️⃣ Architecture Decision
- **Approach**: MCP standalone (stdio, not daemon-integrated)
- **Pattern**: Like `brain-rules` MCP (self-referential but clean)
- **Language**: Go (canonical rule, single binary)
- **Deployment**: Single process, all IDEs share it

### 2️⃣ Indexing Strategy
- **Lazy**: Builds on first search (~2-5 seconds), then cached
- **Vector DB**: Qdrant + native FastEmbed embeddings
- **Source**: Indexes from `docs/` folder
- **Smart**: Respects RAG priority (critical > high > medium > low)

### 3️⃣ Storage Contract
- **Git-tracked**: Manifest + changelog (source of truth)
- **Ephemeral**: Indices rebuild on startup (safe for Docker cleanup)
- **No data loss**: Pull docs from git, rebuild indices, always consistent

### 4️⃣ Integration Points
- Register in `mcp/registry.yml`
- Implement CLI: `brain docs-rag search <query>`
- Optional daemon API + UI components
- Testing: >80% coverage required

---

## FILES YOU'LL NEED

### In New Session, Use These (in `/home/reeinharrrd/.brain/`)

| File | Size | Use | Read First? |
|------|------|-----|-----------|
| **QUICK-START-COPY-PASTE.md** | 50 lines | Copy prompt into chat | ⭐ YES |
| **IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md** | 400 lines | Main guide for agent | YES |
| **HOW-TO-USE-DOCS-RAG-PROMPTS.md** | 300 lines | Workflow guidance | If unsure |
| **PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md** | 3500 lines | Design details | For deep dives |

---

## THE FASTEST WAY TO START

### Step 1: Copy the Prompt (30 seconds)
Open: `/home/reeinharrrd/.brain/QUICK-START-COPY-PASTE.md`
Copy the prompt under "THE PROMPT (Copy-Paste Below)"

### Step 2: Paste into New VS Code Session (30 seconds)
```
1. Open VS Code in ~/.brain
2. Press ⌘+shift+i (Copilot Chat)
3. Paste the prompt
4. Hit Enter
```

### Step 3: Let Agent Implement (40-50 hours)
Agent will:
- Read the implementation guide
- Execute 23 tasks in 5 phases
- Stop at checkpoints for your review
- Ask questions if blocked

---

## PHASE BREAKDOWN

```
PHASE 1: Core Indexing Engine (14-18h)
├─ Project setup
├─ Types + manifest parsing
├─ Document loader
├─ Chunking strategy
├─ Qdrant integration
├─ Lazy-load implementation
├─ Full index builder
├─ Search with ranking
└─ Status endpoints + Tests
   Checkpoint: Index builds & searches work

PHASE 2: MCP Server (4-5h)
├─ Stdio server scaffolding
├─ docs_search tool
├─ docs_status & docs_rebuild tools
└─ Protocol tests
   Checkpoint: MCP binary works

PHASE 3: Brain Integration (4-5h)
├─ registry.yml entry
├─ CLI command: brain docs-rag search
├─ Daemon API (optional)
└─ UI component (optional)
   Checkpoint: CLI works

PHASE 4: Docs + Tests (8-10h)
├─ Comprehensive testing (>80% coverage)
├─ ADR-0006 (architecture decision record)
├─ Architecture docs
├─ MCP README
└─ Code quality
   Checkpoint: All tests pass, ADR written

PHASE 5: Build + Release (1-2h)
├─ Build script
└─ Final validation
   Done: Production binary ready
```

---

## HOW THE IMPLEMENTATION WILL WORK

### Session 1: Tell Agent to Start

```
I'm implementing Docs-RAG MCP for Brain.

Read: /home/reeinharrrd/.brain/IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md

Execute Phase 1 (Tasks 1.1-1.9) to build the core indexing engine.

Show code for each task.
Run tests after each (go test ./...)
Stop at Phase 1 checkpoint.
```

### Agent Will:
✅ Read the implementation guide  
✅ Create project structure  
✅ Implement each task  
✅ Write tests  
✅ Show progress at checkpoints  

### You Will:
✅ Review code at checkpoints  
✅ Ask questions if unclear  
✅ Approve to continue to next phase  

---

## ACCEPTANCE CRITERIA (What "Done" Means)

### Functional ✅
- [ ] `brain docs-rag search "authentication"` returns results
- [ ] Results ranked by RAG priority (critical > high > medium > low)
- [ ] First search ~2-5s, subsequent searches <200ms
- [ ] Domain filtering works
- [ ] Lazy-load triggers on startup

### Tests ✅
- [ ] >80% code coverage
- [ ] Unit tests for indexer + search
- [ ] Integration tests end-to-end
- [ ] Protocol tests for MCP tools
- [ ] All passing

### Documentation ✅
- [ ] ADR-0006 written (why MCP standalone)
- [ ] Architecture docs (how it works)
- [ ] MCP README (building/running)
- [ ] Updated main README

### Integration ✅
- [ ] Registered in `mcp/registry.yml`
- [ ] CLI command works
- [ ] Tests passing
- [ ] No lint warnings
- [ ] 100% English (no Spanish)

---

## TIME EXPECTATIONS

| Approach | Duration | Sessions | Best For |
|----------|----------|----------|----------|
| Phase-by-phase | 40-50h | 4-5 days (one per phase) | Quality + understanding |
| Continuous | 10-12h | 1 long session | Speed (nonstop coding) |
| Delegated | 40-50h | Distributed (agent codes, you review) | Highest quality + async work |

**Recommended**: Phase-by-phase (one phase per day, checkpoints for review)

---

## WHAT YOU'LL HAVE AT THE END

### Code
- ✅ Go binary: `mcp/docs-rag-mcp/docs-rag-mcp` (standalone exe)
- ✅ Tests: >80% coverage, all passing
- ✅ CLI: Brain integration with search command

### Documentation
- ✅ ADR-0006: Decision record (why this approach)
- ✅ Architecture: How indexing + search works
- ✅ README: Build + development guide
- ✅ API docs: MCP tool contracts

### Integration
- ✅ Listed in `mcp/registry.yml`
- ✅ Callable via: `brain docs-rag search <query>`
- ✅ Used by agents via MCP tools

---

## VERIFICATION CHECKLIST (Before You Start)

Run these to confirm setup:

```bash
# 1. Qdrant running?
docker ps | grep qdrant

# 2. Docs exist?
ls ~/.brain/docs/{adr,architecture}/ | wc -l

# 3. Manifest exists?
cat ~/.brain/docs-manifest.json | head -5

# 4. Git ready?
cd ~/.brain && git status
```

All should succeed. If not:
```bash
# Start services
cd ~/.brain && docker compose up -d
```

---

## NEXT ACTION

### Option A: Do It Yourself (Recommended for Learning)
1. Open `/home/reeinharrrd/.brain/IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md`
2. Code each task manually
3. Learn the system in depth
4. ~50h of work

### Option B: Delegate to Agent (Fastest)
1. Open `/home/reeinharrrd/.brain/QUICK-START-COPY-PASTE.md`
2. Copy the prompt
3. Paste into Copilot Chat in new VS Code session
4. Agent handles all 23 tasks
5. You review at checkpoints

### Option C: Hybrid (Balanced)
1. Agent implements Phase 1-2 (indexing + MPC server)
2. You review + understand core logic
3. Agent implements Phase 3-5 (integration + docs)
4. You review final output
5. ~10h total time commitment

**My Recommendation**: Option C (get agent to do boring parts, review the interesting ones)

---

## FILES YOU'LL TOUCH

These will be created/modified:

```
CREATE:
├── mcp/docs-rag-mcp/                   Main project folder
│   ├── main.go                        Entry point
│   ├── go.mod                         Go modules
│   ├── internal/
│   │   ├── indexer/                   Core indexing logic
│   │   ├── search/                    Search ranking
│   │   ├── store/                     Qdrant wrapper
│   │   └── tools/                     MCP tool handlers
│   ├── docs-rag-mcp                   Final binary
│   └── README.md                      Development guide
├── cli/cmd/brain/docs_rag.go          CLI command
├── daemon/internal/api/handlers_docs_rag.go  (Optional)
├── desktop/src/components/DocsSearch.tsx     (Optional)
└── docs/adr/ADR-0006-docs-rag-mcp-architecture.md

MODIFY:
├── mcp/registry.yml                   Add docs-rag entry
└── docs/architecture/docs-rag-mcp.md  Add architecture doc

GIT:
├── .gitignore                         Update if needed
└── Commits after each phase
```

---

## GETTING HELP

### If Stuck During Implementation:

**In Agent Chat**:
```
I'm blocked on Task X because [specific issue].

Can you:
1. Explain what this is supposed to do
2. Show an example
3. Suggest how to debug it
```

**Check References**:
- `/home/reeinharrrd/.brain/PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md` (section 4 = architecture)
- `/home/reeinharrrd/.brain/docs/architecture/daemon-orchestration.md` (MCP patterns)
- `/home/reeinharrrd/.brain/mcp/brain-mcp-server/` (existing MCP example)

**Run Tests**:
```bash
go test ./... -v    # Verbose tests
go test ./... -cover  # Coverage report
```

---

## FINAL DECISION

**You've approved**:
- ✅ Language: Go
- ✅ Architecture: MCP standalone
- ✅ Indexing: Lazy-load
- ✅ Storage: Qdrant native embeddings

**Now**: Execute implementation in new session.

---

## TL;DR

1. **New Session**: Open VS Code in `~/.brain`
2. **Read**: `QUICK-START-COPY-PASTE.md` (1 min)
3. **Copy**: The prompt from that file
4. **Paste**: Into Copilot Chat (⌘+shift+i)
5. **Done**: Agent implements, you review at checkpoints

Estimated total time: 40-50 hours
Total human effort: 3-5 hours (checkpoints only)

🚀 Ready to ship in 4-5 days (one phase per day)

