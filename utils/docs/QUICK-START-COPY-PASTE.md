---
type: prompt
id: quick-start-docs-rag-implementation
title: Quick Start Prompt for Docs-RAG MCP Implementation
version: 1.0.0
status: active
date_created: 2026-04-03
language: en
category: workflow
---

# Quick Start: Copy-Paste This Into Copilot Chat

⚠️ **Prerequisites**: Open VS Code in `/home/reeinharrrd/.brain`

---

## THE PROMPT (Copy-Paste Below)

```
I need to implement a complete Docs-RAG MCP server for Brain.

Here's what you need to know:

CONTEXT:
- Building an MCP (Model Context Protocol) server for semantic search over Brain docs
- Language: Go
- Storage: Qdrant vector DB (already running in docker-compose)
- Indexing: Lazy-load (builds on first search, then cached)
- Design: Standalone stdio process (follows brain-rules MCP pattern)

YOUR TASK:
1. Read: /home/reeinharrrd/.brain/IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md
2. Follow: The 5 phases, 23 tasks, task breakdown EXACTLY
3. Phase 1 (Priority): Core indexing engine (Tasks 1.1-1.9 + tests)
4. Stop at: Phase 1 checkpoint and wait for my review

KEY CONSTRAINTS:
- Go only (single binary)
- >80% test coverage required
- All errors logged (no silent failures)
- 100% English (code + comments)
- Follow acceptance criteria from prompt

BEFORE YOU START:
Verify Qdrant is running:
  docker ps | grep qdrant
  # If not: cd ~/.brain && docker compose up -d

Are you ready to read the implementation prompt and start Phase 1?
```

---

## THEN SAY:

After copying the prompt above to Copilot Chat and agent responds "ready":

```
Start Phase 1: Core Indexing Engine.

Execute tasks in order:
- T1.1: Project setup
- T1.2: Document & manifest types
- T1.3: Document loader
- T1.4: Document chunking
- T1.5: Qdrant client wrapper

For each task:
1. Show code changes
2. Explain what was done
3. Run: go test ./... -cover
4. Show test output

When all 5 tasks + tests complete, stop and wait for my review.
Don't proceed to T1.6-1.9 until I say to.
```

---

## THEN AFTER CHECKPOINT:

```
Review complete. Proceed to Phase 1 Tasks 1.6-1.9.

Execute:
- T1.6: Lazy-load indexer initialization
- T1.7: Full index build
- T1.8: Search implementation
- T1.9: Status & metadata

Same process: code → explanation → tests → stop at checkpoint.
```

---

## ALTERNATIVE: Skip the Steps (Do All at Once)

If you want faster implementation without step-by-step reviews:

```
Read /home/reeinharrrd/.brain/IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md

Implement everything in order. Show progress at each phase checkpoint:
- After Phase 1: Show test coverage report
- After Phase 2: Verify MCP tools work
- After Phase 3: Verify CLI command works
- After Phase 4: Verify >80% coverage + ADR-0006
- After Phase 5: Show final binary and confirm it works

When complete, ask if anything needs fixing.
```

---

## HELPFUL REFERENCES

If stuck on any task, these files have details:

- **IMPLEMENTATION-PROMPT** (main): /home/reeinharrrd/.brain/IMPLEMENTATION-PROMPT-DOCS-RAG-MCP.md (400 lines, read this first)
- **Full SDD** (detailed): /home/reeinharrrd/.brain/PROMPT-DOCS-RAG-MCP-IMPLEMENTATION.md (3500 lines, for design details)
- **How to Use**: /home/reeinharrrd/.brain/HOW-TO-USE-DOCS-RAG-PROMPTS.md (this explains structure)
- **Architecture**: /home/reeinharrrd/.brain/docs/architecture/daemon-orchestration.md (how MCPs work in Brain)

---

## VERIFY BEFORE STARTING

Run these in terminal to confirm setup:

```bash
# 1. Check Brain directory
ls -la ~/.brain/mcp/ | grep docs-rag

# 2. Check Qdrant is running
docker ps | grep qdrant

# 3. Check docs can be indexed
ls ~/.brain/docs/{adr,architecture,skills}/*.md | wc -l

# 4. Check manifest exists
cat ~/.brain/docs-manifest.json | head -10
```

All should succeed. If not:
```bash
# Start Qdrant
cd ~/.brain && docker compose up -d
```

---

## GO TIME! 🚀

Paste the prompt above into Copilot Chat and enjoy building.

Ask questions only if blocked.

When done, you'll have:
✅ Go MCP binary (standalone)
✅ CLI command: `brain docs-rag search <query>`
✅ 3 MCP tools: docs_search, docs_status, docs_rebuild
✅ Registered in mcp/registry.yml
✅ >80% test coverage
✅ Complete ADR-0006 + docs

Total time: 40-50 hours (can be split across sessions)
