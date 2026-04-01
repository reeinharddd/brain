# Brain Repository Refactoring Plan

**Date**: March 31, 2026  
**Status**: Planning Phase  
**Goal**: Clean, consolidate, and reorganize the Brain codebase for maintainability and clarity

---

## Problem Statement

The Brain repository has accumulated technical debt:

1. **Script Proliferation**: 33 shell scripts across `scripts/` with overlapping responsibilities
2. **Documentation Duplication**: 5 overlapping markdown files in root (README, IMPLEMENTATION_COMPLETE, MIGRATION_GUIDE, STATUS, QUICKSTART)
3. **Mixed Concerns**: Setup, testing, maintenance, and operational scripts not clearly organized
4. **Legacy Code**: Several scripts appear unmaintained or obsoleted by newer alternatives
5. **Code Quality**: Inconsistent error handling, logging, and argument parsing across scripts

---

## Audit: Scripts Classification

### TIER 1: Core (Active, Essential)

These are actively maintained and essential to Brain operation:

- ✅ `brain-cli.sh` (562 lines) - **Orchestrator** for all services
- ✅ `brain-setup.sh` (75 lines) - **Setup wizard**, calls brain-cli.sh
- ✅ `init.sh` (138 lines) - **Per-project initialization**
- ✅ `doctor.sh` (200 lines) - **Health check** and diagnostics
- ✅ `consolidate-memory.sh` (205 lines) - **Memory maintenance**
- ✅ `guardiansh` (wrapper) - **Security checks**

**Action**: Keep, audit for quality, consolidate shared functions

### TIER 2: Infrastructure (Maintained, Specialized)

Configuration and adapter generation:

- ✅ `adapters/generate.sh` - **Compile rules to all IDEs**
- ✅ `cron-setup.sh` (94 lines) - **Maintenance scheduling**
- ✅ `install.sh` (343 lines) - **Full environment setup**
- ✅ `deploy.sh` (409 lines) - **Deployment automation**

**Action**: Consolidate overlapping init logic into single `install.sh`

### TIER 3: Utilities (Used Periodically)

Helper scripts for specific tasks:

- ⚠️ `detect-stack.sh` (97 lines) - **Stack detection** (used by agents)
- ⚠️ `render-skill-context.sh` (102 lines) - **Context generation**
- ⚠️ `memory-namespace.sh` (in commands) - **Memory utilities**
- ⚠️ `provider-proxy.sh` (232 lines) - **Model routing**
- ⚠️ `consolidate-memory.sh` (205 lines) - **Memory consolidation**

**Action**: Group into `lib/` directory, create unified interface

### TIER 4: Experimental/Test (Rarely Used)

Tests, benchmarks, and experimental code:

- ❓ `benchmark-brain.sh` - Output timing metrics
- ❓ `smoke-real-env.sh` - Environment validation
- ❓ `test-docker-mcp.sh` - Docker MCP testing
- ❓ `test-stdio-mcp.sh` - Stdio MCP testing
- ❓ `test-memory.sh` - Memory system testing
- ❓ `vector-context-index.sh` - Vector indexing

**Action**: Move to `tests/` directory, consolidate into pytest-based framework

### TIER 5: Obsolete/Confusing

Scripts that duplicate or are subsumed by newer tooling:

- ❌ `autostart-setup.sh` - Subsumed by `brain-cli.sh autostart-*`
- ❌ `setup-persistent.sh` - Overlaps with `install.sh` and `brain-setup.sh`
- ❌ `analyze-sources.sh` - Unclear purpose
- ❌ `sync-ai-local.sh` - Docker handles this now
- ❌ `install-hooks.sh` - Subsumed by `init.sh`

**Action**: Delete or consolidate into higher-tier scripts

### TIER 6: Incomplete/Legacy

Files that reference deprecated patterns:

- ❓ `agent-runner.py` - Still needed? (Check if agents are run via Python or CLI)
- ❓ `embed.py` - Vector embedding backend
- ❓ `validate-schema.py` - Schema validation for canonical.md

**Action**: Verify usage, consolidate with brain-cli.sh or separate into `lib/python/`

---

## Documentation Audit

### Root-Level Markdown Files (5 files, ~36KB total)

| File | Size | Purpose | Action |
|------|------|---------|--------|
| `README.md` | 8.9K | Architecture overview, component list | **Keep as main** - consolidate others here |
| `IMPLEMENTATION_COMPLETE.md` | 11K | Implementation report (duplicate of README) | **Delete** - move critical content to README |
| `MIGRATION_GUIDE.md` | 9.3K | Migration instructions (should be in QUICKSTART) | **Merge** into README 'Getting Started' section |
| `STATUS.md` | 5.1K | Status dashboard (should be from README) | **Delete** - auto-generate from `brain status` |
| `QUICKSTART.md` | 2.2K | 5-minute quickstart (good, but small) | **Expand** and fold into README 'Quick Start' |

**New Structure**:
- `README.md` (comprehensive) with sections:
  - Architecture & Design
  - Quick Start (5 min)
  - Installation & Setup
  - Commands Reference (replace BRAIN_CLI_REFERENCE.txt)
  - FAQ & Troubleshooting
- `docs/ADRs/` - Architecture decision records
- `docs/guides/` - Detailed guides (cloud sync, memory, security, etc.)

---

## Code Quality Audit

### Script Issues to Fix

1. **Inconsistent error handling**: Some scripts use `set -e`, others don't
2. **Logging inconsistency**: Mix of `echo`, custom log functions, `logger`
3. **Argument parsing**: Some scripts use getopts, others manual parsing
4. **Naming collisions**: `init.sh`, `doctor.sh` not namespaced
5. **Code duplication**: Common functions (docker_is_running, log_*, check_*) repeated across scripts

### Code Quality Standards to Enforce

- ✅ Use `set -euo pipefail` consistently
- ✅ Create `lib/common.sh` for shared functions
- ✅ Consistent log format: `[INFO] | [WARN] | [ERROR]`
- ✅ Argument parsing via getopts
- ✅ Consistent exit codes (0=success, 1=error, 2=invalid args)
- ✅ Shellcheck-compliant (no warnings)

---

## Refactoring Strategy

### Phase 1: Create Foundation (Safety Nets)
1. **Create `lib/` directory structure**:
   ```
   lib/
     common.sh          # Shared functions (logging, arg parsing, docker, etc.)
     assert.sh          # Assertion helpers
     colors.sh          # Color definitions
   ```
2. **Create test framework** in `tests/`:
   ```
   tests/
     test.sh            # Main test runner
     fixtures/          # Test data
     scripts/           # Script-specific tests
   ```
3. **Document current state** in REFACTOR_PLAN.md (this file)

### Phase 2: Consolidate & Extract
1. **Extract shared functions** from scripts into `lib/common.sh`
2. **Delete obsolete scripts** (TIER 5)
3. **Merge overlapping setup scripts**:
   - `install.sh` + `setup-persistent.sh` + `brain-setup.sh` → `scripts/install.sh` (single)
4. **Move test scripts** to `tests/scripts/`
5. **Move utility scripts** to `lib/` or `scripts/lib/`

### Phase 3: Consolidate Documentation
1. **Expand `README.md`** with all content
2. **Delete**: IMPLEMENTATION_COMPLETE.md, MIGRATION_GUIDE.md, STATUS.md, QUICKSTART.md
3. **Generate**: BRAIN_CLI_REFERENCE.txt from `brain help` output
4. **Migrate**: docs to `docs/guides/` where appropriate

### Phase 4: Improve Code Quality
1. **Lint all shell scripts** with shellcheck
2. **Add proper error handling**
3. **Improve logging consistency**
4. **Add usage/help text** to all scripts
5. **Add test coverage** for critical scripts

### Phase 5: Validation & Verification
1. **Run smoke tests** for all high-level commands
2. **Verify backward compatibility** (scripts callable from commands)
3. **Test on multiple shell versions** (bash, zsh)
4. **Update memory & documentation** with changes

---

## Deletion Candidates (Verify First)

Before deleting, verify these are truly obsolete:

1. `autostart-setup.sh` - Functionality moved to `brain-cli.sh autostart-*`
2. `setup-persistent.sh` - Overlaps significantly with `install.sh`
3. `install-hooks.sh` - Logic moved to `init.sh`
4. `sync-ai-local.sh` - Docker compose handles this
5. `analyze-sources.sh` - Purpose unclear, no references found

---

## New Structure (Target)

```
scripts/
  brain-cli.sh          # Main orchestrator (keep, improve)
  install.sh            # Installation (consolidated from 3 scripts)
  deploy.sh             # Deployment
  lib/
    common.sh           # Shared functions
    assert.sh           # Assertions
    colors.sh           # Color definitions
  utils/
    detect-stack.sh
    render-skill-context.sh
    provider-proxy.sh
    # ... other utilities

docs/
  ARCHITECTURE.md       # Moved from README
  guides/
    quickstart.md
    cloud-sync.md
    memory-system.md
    security-setup.md
  adr/
    adr-001-centralized-brain-cli.md
    adr-002-modular-refactor.md

tests/
  test.sh               # Main test runner
  fixtures/
  scripts/
    test-brain-cli.sh
    test-init.sh
    test-docker.sh

# Consolidated documentation
README.md               # Single source of truth for root

# Deleted
-IMPLEMENTATION_COMPLETE.md
-MIGRATION_GUIDE.md
-STATUS.md
-QUICKSTART.md
```

---

## Implementation Order

1. ✅ **Audit & Planning** (this document)
2. ⏳ **Create lib/ and shared functions**
3. ⏳ **Verify & lint existing scripts**
4. ⏳ **Delete obsolete scripts**
5. ⏳ **Consolidate setup scripts**
6. ⏳ **Move utilities to lib/**
7. ⏳ **Consolidate documentation**
8. ⏳ **Create test framework**
9. ⏳ **Verify all functionality**
10. ⏳ **Update CI/CD and documentation**

---

## Success Criteria

- [ ] All TIER 1 scripts pass shellcheck with 0 warnings
- [ ] Single `install.sh` handles all setup scenarios
- [ ] Documentation consolidated into README + docs/guides/
- [ ] Test coverage > 80% for critical paths
- [ ] Code duplication < 5% (currently ~15%)
- [ ] Setup time reduced (consolidation)
- [ ] Cognitive load reduced (clear file organization)

---

## Rollback Strategy

Given that this is a configuration/script refactor (not logic-heavy):

1. **Snapshot current state**: `git stash` before starting
2. **Commit per phase**: Each major consolidation is a separate commit
3. **Keep old scripts in archive**: Don't delete immediately, mv to `.archive/` first
4. **Tag before/after**: `git tag refactor-v1-start` and `refactor-v1-complete`
5. **Easy rollback**: `git reset --hard refactor-v1-start` if needed

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Break user setup flow | Medium | High | Test all install paths before cleanup |
| Break IDE integration | Low | Medium | Verify adapters still generate correctly |
| Lose important logic | Low | High | Archive scripts first, don't delete |
| Incomplete migration | Medium | Medium | Automated tests, phase-by-phase rollout |

