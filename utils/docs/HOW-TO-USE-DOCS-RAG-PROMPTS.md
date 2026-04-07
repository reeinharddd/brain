---
type: guide
id: how-to-use-docs-rag-implementation-prompts
title: How to Use Docs-RAG MCP Implementation Prompts in New Session
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: workflow
---

# How to Use Docs-RAG MCP Implementation Prompts

**For**: Starting implementation in a fresh VS Code session  
**Time to Read**: 5 minutes  
**Time to Setup**: 5 minutes

---

## Quick Start (3 Steps)

### Step 1: Open New VS Code Session
```bash
code /home/reeinharrrd/.brain
```

### Step 2: Use This Prompt in Chat

Copy and paste into GitHub Copilot Chat (⌘+shift+i):

```
I need to implement a complete Docs-RAG MCP server for Brain.

Please read these two reference documents first:
1. /home/reeinharrrd/.brain/IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md (main guide - read this first)
2. /home/reeinharrrd/.brain/PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md (detailed SDD - reference for design)

Use the implementation prompt to guide your work. Follow the phases and task breakdown exactly.

Start with Phase 1: Core Indexing Engine (Tasks 1.1-1.9).

Stop at Phase 1 Checkpoint and ask for review before proceeding.

Key constraints:
- Language: Go only
- >80% test coverage required
- All errors logged (no silent failures)
- 100% English (code + comments + errors)
- Follow acceptance criteria from prompt

Let's begin. Are you ready to start Phase 1?
```

### Step 3: Let Agent Continue

Agent will:
1. Read both documents ✅
2. Start Phase 1 ✅
3. Ask questions if blocked ✅
4. Checkpoint at Phase 1 end ✅
5. Wait for your review ✅

---

## Document Organization

You have **3 Documents**:

### 1. IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md (YOU START HERE)
- **What**: Self-contained implementation guide
- **Contains**: Decisions, task breakdown, acceptance criteria
- **Use**: Main reference for agent implementer
- **Length**: ~400 lines, easy to read
- **Best For**: "Just do this" - step-by-step tasks

### 2. PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md (FOR DESIGN DETAILS)
- **What**: Complete Spec-Driven Development document
- **Contains**: Phases 1-8, detailed architecture, tool contracts, data flows
- **Use**: Reference when you need detailed design
- **Length**: ~3,500 lines, comprehensive
- **Best For**: "Why did we design it this way?" - deep understanding

### 3. docs-rag-technical-decisions.md (IN SESSION MEMORY)
- **What**: Session notes on decisions made
- **Contains**: Quick summary of 3 key decisions (Go, lazy-load, Qdrant native)
- **Use**: Quick reference for decision rationale
- **Best For**: "Why Go?" - quick answers

---

## Implementation Strategy

### Approach A: Strict Phase-by-Phase (Recommended)

```
Session 1 (Day 1): Phase 1 - Core Indexing
├─ Tasks 1.1-1.5 (setup, types, loader, chunking, Qdrant)
├─ Tests (1.10)
└─ Checkpoint: Can index docs/ folder

Session 2 (Day 2): Phase 1.6-4 + Phase 2 - MCP Server
├─ Tasks 1.6-1.9 (lazy-load, build, search, status)
├─ Tasks 2.1-2.4 (MCP stdio server)
└─ Checkpoint: MCP binary works

Session 3 (Day 3): Phase 3-4 - Integration + Docs
├─ Tasks 3.1-3.4 (registry, CLI, daemon, UI)
├─ Tasks 4.1-4.5 (tests, ADR-0006, docs)
└─ Checkpoint: All tests pass, >80% coverage

Session 4 (Day 4): Phase 5 - Build + Release
├─ Tasks 5.1-5.2 (build script, final validation)
└─ DONE: Production binary ready
```

### Approach B: Continuous Implementation (Faster)

```
Session 1 (1 long session): All Phases + Tests
├─ Don't stop at checkpoints
├─ Run all tasks continuously
├─ Tests at the end
└─ 10-12 hours nonstop
```

### Approach C: Delegated Implementation

```
Session 1: Ask `implementer` agent to do it all
├─ Copy IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md
├─ Paste "Implement all tasks in order, show progress at each checkpoint"
├─ Agent codes everything
├─ You review PRs incrementally
└─ 40-50 hours distributed across multiple responses
```

**Recommendation**: **Approach A** (strict phases) for better code quality and understanding.

---

## What to Do in Each Session

### Before Each Session Starts

1. **Verify directory exists**:
   ```bash
   ls -la /home/reeinharrrd/.brain/mcp/docs-rag-mcp/
   ```

2. **Verify Qdrant is running** (required for testing):
   ```bash
   docker ps | grep qdrant
   # If not running:
   cd ~/.brain && docker compose up -d
   ```

3. **Review checkpoint from last session**:
   ```bash
   cat /home/reeinharrrd/.brain/mcp/docs-rag-mcp/PROGRESS.md
   ```

### In Each Session

1. **Create new planning note**:
   ```bash
   # At start of session
   cat > mcp/docs-rag-mcp/PROGRESS.md << 'EOF'
   ## Session [#]: [Date]
   
   Tasks to complete: 1.1 - 1.5
   Checkpoint: Index builds successfully
   
   Progress:
   - [ ] T1.1: Project setup
   - [ ] T1.2: Types + manifest
   - [ ] T1.3: Document loader
   - [ ] T1.4: Chunking
   - [ ] T1.5: Qdrant client
   - [ ] T1.10: Tests
   
   Issues:
   (none yet)
   EOF
   ```

2. **Start agent with phase tasks**:
   ```
   Read IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md and execute Phase 1 Tasks 1.1-1.5.
   Show code changes for each task.
   Run tests after each task (go test ./...)
   Stop at checkpoint and show test output.
   ```

3. **After phase completes**:
   - Review code + tests
   - Check coverage (`go test ./... -cover`)
   - Commit progress
   - Update PROGRESS.md
   - Plan next phase

---

## How to Use the Documents

### When Implementing Task 1.3 (Document Loader)

**Step 1**: Read the task description from IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md
```
Task 1.3: Document Loader
- Implement LoadDocument(path string) (*Document, error)
- Load markdown file, parse frontmatter, extract body
- Handle errors: file not found, invalid YAML, missing fields
- Estimated: 1.5 hours
- Success: Can load any doc from docs/ folder, errors logged
```

**Step 2**: If you need more detail, check PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md section 4.3:
```
"Key Data Structures":
type Document struct {
    ID               string            // From filename
    Domain           string            // adr, architecture, etc.
    Title            string            // From frontmatter
    Status           string            // active, archived, deprecated
    ...
}
```

**Step 3**: If still unclear, reference existing Brain code:
```bash
# Look at how skills.go parses manifest
grep -A 20 "type Manager struct" daemon/internal/manager/skills.go
```

---

## Checkpoints (When to Stop & Review)

### Phase 1 Checkpoint
**After Tasks 1.1-1.9** (before MCP server):
```bash
# Verify indexer works standalone
go test ./internal/indexer/... -cover -v
go test ./internal/search/... -cover -v

# Verify can index docs/ folder
# (Create simple test that loads mcp/docs-rag-mcp/test-docs/)
```

**Ask for code review** before continuing.

### Phase 2 Checkpoint
**After Tasks 2.1-2.4** (before Brain integration):
```bash
# Verify MCP binary works
./docs-rag-mcp < test-tool-call.json

# Verify all 3 tools callable
# (Create test inputs for docs_search, docs_status, docs_rebuild)
```

**Ask for code review** before continuing.

### Phase 3 Checkpoint
**After Tasks 3.1-3.4** (before docs):
```bash
# Verify CLI command works
brain docs-rag search "daemon" --limit 3

# Verify registry entry correct
grep -A 10 "docs-rag:" mcp/registry.yml
```

**Ask for code review** before continuing.

### Phase 4 Checkpoint
**After Tasks 4.1-4.5** (before build):
```bash
# Verify tests pass + coverage
go test ./... -cover

# Verify ADR written
cat docs/adr/ADR-0006-docs-rag-mcp-architecture.md
```

**Ready to ship** if >80% coverage + ADR written.

### Phase 5: Release
**After Tasks 5.1-5.2**:
```bash
# Build binary
make build

# Test binary works
./bin/docs-rag-mcp < test-tool-call.json

# Commit everything
git add -A && git commit -m "feat(mcp): docs-rag semantic search implementation"
```

**Done!** 🎉

---

## Troubleshooting

### Issue: "Qdrant connection refused"
```bash
# Start Qdrant
cd ~/.brain && docker compose up -d qdrant
# Wait 10 seconds for startup
sleep 10
# Verify
curl http://localhost:6333/collections
```

### Issue: "docs-manifest.json not found"
```bash
# Verify path from mcp/docs-rag-mcp/ context
ls -la ../../docs-manifest.json
# Use relative path in code
```

### Issue: "Tests failing with encoding errors"
```bash
# File encoding issue - check UTF-8
file docs/adr/*.md
# Re-run with UTF-8 locale
LC_ALL=en_US.UTF-8 go test ./...
```

### Issue: "go mod tidy failing"
```bash
# Dependencies issue - check module name
cat go.mod | head -5
# Ensure it matches: module github.com/reeinharrrd/brain/mcp/docs-rag-mcp
```

### Issue: "MCP stdin/stdout not working"
```bash
# Debug JSON-RPC protocol
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./docs-rag-mcp
# Should return tools list
```

---

## Commits to Make

### After Phase 1:
```bash
git add -A
git commit -m "feat(mcp-docs-rag): core indexing engine and lazy-load logic"
```

### After Phase 2:
```bash
git commit -m "feat(mcp-docs-rag): stdio MCP server with 3 tools"
```

### After Phase 3:
```bash
git commit -m "feat(mcp-docs-rag): brain integration - registry, cli, daemon api"
```

### After Phase 4:
```bash
git commit -m "docs: ADR-0006, architecture docs, >80% test coverage"
```

### After Phase 5:
```bash
git commit -m "build: production binary, release ready"
```

---

## Success Criteria Checklist

Print this and check off:

```
FUNCTIONAL:
☐ brain docs-rag search "daemon" returns ranked results
☐ Empty results handled gracefully
☐ Domain filtering works (--domain adr)
☐ Lazy-load: first search ~2-5s, subsequent <200ms
☐ Manifest validation: invalid docs skipped
☐ RAG priority: "critical" > "high" > "medium" > "low"
☐ Chunk strategy: both "section" and "sentence" work

NON-FUNCTIONAL:
☐ Language: Go (single binary)
☐ Tests: >80% coverage, all passing
☐ Memory: <200MB for index
☐ Logging: All errors logged (grep "ERROR" logs)
☐ English: No Spanish / mixed languages

INTEGRATION:
☐ BRAIN_ENV=development works
☐ BRAIN_ENV=production works
☐ Qdrant unavailable → graceful error
☐ Registry entry correct
☐ CLI command works
☐ ADR-0006 written

WHEN ALL CHECKED: Feature ships!
```

---

## Questions? Next Steps?

1. **Ready to start?** → Copy the quick-start prompt and paste into Copilot Chat
2. **Need clarification?** → Read the relevant section in PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md
3. **Something wrong?** → Check troubleshooting above
4. **Blocked?** → Ask in chat with context: "Task X.Y is blocked because..."

---

## Related Documents

- **Implementation Guide**: IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md (the main one!)
- **Detailed SDD**: PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md
- **Technical Decisions**: /memories/repo/docs-rag-mcp-approved-plan.md
- **Architecture**: docs/architecture/daemon-orchestration.md
- **Manifest Contract**: docs-manifest.json

