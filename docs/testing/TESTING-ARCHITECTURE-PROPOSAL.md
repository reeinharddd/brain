# 🧪 Brain Testing System - Architectural Proposal

**Document Date**: April 3, 2026  
**Status**: PROPOSAL (Awaiting Approval)  
**Effort Estimate**: 56 hours (1-2 weeks)

---

## Executive Summary

This proposal outlines a **centralized, professional testing and CI/CD system** for Brain that:

✅ Eliminates shell scripts (no `.sh` files)  
✅ Provides unified test coordination via Go  
✅ Supports multiple test types: unit (Go), unit (TS), integration, browser E2E  
✅ Includes structured logging (NDJSON) and real-time dashboards  
✅ File watchers for continuous validation during development  
✅ Full GitHub Actions CI/CD integration

---

## Problem Statement

Current state:

- No centralized testing framework
- Test coverage scattered or missing
- Hard to know what's broken and why
- No visibility during test execution
- Shell scripts discouraged by repository policy

Desired state:

- Single command: `brain test` runs everything
- Real-time visibility (logs, dashboard, progress)
- All logging structured and parseable
- File watchers for rapid feedback loop
- Professional CI/CD integration

---

## Architecture Overview (8 Layers)

```text
┌─────────────────────────────────────────────────────────────┐
│  LAYER 1: Test Orchestrator (Go)                            │
│  - Central CLI: brain test <suite>                          │
│  - Runs phases sequence or parallel                         │
│  - Coordinates daemon startup/shutdown                      │
│  - Aggregates and reports results                           │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 2: Test Registry & Configuration                     │
│  - testconfig.yml: maps test suites to runners             │
│  - Defines groups, env vars, dependencies                   │
│  - No external frameworks                                   │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 3: Test Runners (4 Parallel Executors)               │
│  1. Go Unit Tests (daemon, CLI)                             │
│  2. TypeScript Unit Tests (React components)                │
│  3. Integration Tests (daemon + CLI + YAML)                 │
│  4. Browser E2E Tests (Playwright)                          │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 4: Structured Logging & Event Stream                 │
│  - NDJSON format (.logs/test-run.ndjson)                    │
│  - Event types: START, PASS, FAIL, SKIP, DURATION, SUMMARY  │
│  - Real-time to stdout + file                               │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 5: File Watchers & Change Detection                  │
│  - Go fsnotify-based watchers                               │
│  - Detects changes in daemon/, cli/, desktop/src/          │
│  - Triggers incremental rebuild/retest on change            │
│  - --watch mode for development                             │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 6: Real-Time Dashboard & Reporting                   │
│  - Web UI on localhost:9091 (during test runs)              │
│  - Live log streams, pass/fail rates, duration              │
│  - HTML report generation (.logs/test-report.html)          │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 7: Artifact Persistence                              │
│  - Stores: logs, reports, screenshots, failures             │
│  - Retention policy: last 30 runs                            │
│  - Enables debugging failed test runs from weeks prior      │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 8: GitHub Actions CI/CD Integration                  │
│  - test.yml, build.yml, deploy.yml workflows                │
│  - Orchestrator runs via GitHub environment                 │
│  - Reports published to PR/commit status                    │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Details

### Layer 1: Test Orchestrator

**Location**: `daemon/cmd/testor/main.go` (new)

**CLI Usage**:

```bash
brain test                    # Run all suites
brain test --suite daemon     # Run only daemon tests
brain test --watch           # Watch mode (rerun on changes)
brain test --filter TestCRUD  # Run specific tests by name
brain test --report html      # Generate HTML report
brain test --ci-mode          # GitHub Actions mode (no interactive features)
```

**Responsibilities**:

- Parse test configuration from YAML
- Coordinate daemon startup/shutdown
- Run test suites in parallel (units) or sequence (integration)
- Aggregate results and exit with proper code
- Stream logs to stdout and `.logs/` directory

### Layer 2: Test Configuration

**Location**: `testconfig.yml` (new, at repo root)

```yaml
test_suites:
  daemon_unit:
    runner: go_test
    path: ./daemon/...
    timeout: 60s
    env:
      BRAIN_TEST: "true"
    depends_on: []

  cli_unit:
    runner: go_test
    path: ./cli/...
    timeout: 30s

  ui_unit:
    runner: typescript_test
    path: ./desktop
    timeout: 45s

  integration:
    runner: integration
    timeout: 120s
    depends_on: [daemon_unit, ui_unit]
    setup:
      - daemon_start
      - seed_test_data

  browser_e2e:
    runner: playwright
    timeout: 180s
    depends_on: [integration]
    env:
      BROWSER: chromium
      HEADLESS: "true"
```

### Layer 3: Test Runners

#### 3a: Go Unit Tests

- **Runner**: Standard `go test` via orchestrator
- **Output Format**: Parsed from `go test -json` (structured)
- **Coverage**: Via `go test -cover`
- **Example Tests**:
  - `daemon/internal/manager/skills_test.go`
  - `daemon/cmd/braind/handlers_test.go` (new)
  - `cli/cmd/brain/registry_test.go` (new)

#### 3b: TypeScript Unit Tests

- **Framework**: Vitest (lightweight, React-friendly)
- **Location**: `desktop/**/*.test.ts` and `desktop/**/*.test.tsx`
- **Example Tests**:
  - `desktop/src/components/SkillsList.test.tsx`
  - `desktop/src/utils/api.test.ts`
- **Coverage**: Via c8 (built-in)

#### 3c: Integration Tests

- **Type**: Go-based, orchestrates daemon + CLI + filesystem
- **Location**: `daemon/tests/integration_test.go` (new)
- **Examples**:
  - `TestFull_CreateSkillViaAPI_VerifyFilesAndSync()`
  - `TestCLI_CreateSkill_ViaAPI_VerifyOutputFormat()`
  - `TestConcurrentWrites_NoDirtyReads()`

#### 3d: Browser E2E Tests

- **Framework**: Playwright
- **Location**: `desktop/e2e/` (new)
- **Examples**:
  - `SkillsList_LoadsAndDisplaysAllItems.spec.ts`
  - `SkillsForm_CreateAndDeleteViaUI.spec.ts`
  - `DaemonRestart_UIReflectsChanges.spec.ts`

### Layer 4: Structured Logging (NDJSON)

**Format**: One JSON object per line, parseable

```json
{"ts":"2026-04-03T10:15:22Z","event":"test_start","suite":"daemon_unit","version":"1.0"}
{"ts":"2026-04-03T10:15:23Z","event":"test_pass","test":"TestCreateSkill","duration_ms":234,"level":"PASS"}
{"ts":"2026-04-03T10:15:45Z","event":"suite_complete","suite":"daemon_unit","passed":34,"failed":2,"skipped":0,"duration_ms":22000}
{"ts":"2026-04-03T10:16:30Z","event":"test_complete","total_passed":156,"total_failed":3,"exit_code":1}
```

**Log Destinations**:

- **Stdout**: Real-time stream (colorized)
- **File**: `.logs/test-run-{timestamp}.ndjson`
- **Symbolic Link**: `.logs/test-run-latest.ndjson`

### Layer 5: File Watchers

**Location**: `daemon/cmd/testor/watcher.go`

```bash
brain test --watch
```

**Watcher Strategy**:

- Daemon files (`.go` in `daemon/`) → rerun `daemon_unit + integration`
- CLI files → rerun `cli_unit + integration`
- UI files (`.tsx` in `desktop/src/`) → rerun `ui_unit + browser_e2e`
- YAML config changes → rerun all

**Smart Test Selection**: Only affected tests rerun (faster iteration)

### Layer 6: Real-Time Dashboard

**Location**: `daemon/cmd/testor/dashboard.go`

**Access**: `http://localhost:9091/tests` (while tests running)

**Shows**:

- Live log stream (tail last 500 lines)
- Suite progress bars
- Pass/fail counts by layer
- Estimated time remaining
- Quick links to failed tests
- Performance metrics (slowest tests)

**HTML Report** (generated after):

- `.logs/test-report-{timestamp}.html`
- Pass/fail breakdown, timings, failure diffs, screenshots
- Can be linked from PR

### Layer 7: Artifact Persistence

**Location**: `.logs/` directory (in .gitignore)

```text
.logs/
├── test-run-2026-04-03T10-15-22Z.ndjson
├── test-run-2026-04-03T10-15-22Z.html
├── screenshots/
│   ├── SkillsList_LoadsAndDisplays.png
│   └── DaemonRestart_UIReflects.png
├── failures/
│   ├── TestCreateSkill_stderr.log
│   └── TestCreateSkill_stdout.log
└── retention.json
```

**Retention**: Automatically keep last 30 test runs

### Layer 8: GitHub Actions Integration

**Location**: `.github/workflows/test.yml` (new)

```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - uses: actions/setup-node@v4
      - run: go run ./daemon/cmd/testor/main.go --report json --ci-mode
      - name: Upload Results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: test-logs
          path: .logs/
      - name: Comment on PR
        if: failure()
        uses: actions/github-script@v7
```

---

## Technology Choices

| Component           | Technology                   | Rationale                                        |
| ------------------- | ---------------------------- | ------------------------------------------------ |
| **Orchestrator**    | Go std lib                   | No external deps; integrates with daemon/CLI     |
| **Unit Tests (Go)** | `testing` + `testify`        | Built-in, proven                                 |
| **Unit Tests (TS)** | Vitest                       | Zero config, fast, ESM-native                    |
| **Integration**     | Go only                      | Type-safe; same language as daemon               |
| **Browser E2E**     | Playwright                   | Already desired; supports debugging, screenshots |
| **Logging**         | NDJSON (custom)              | Structured, streaming-friendly                   |
| **Watchers**        | fsnotify (Go lib)            | Lightweight, cross-platform                      |
| **Dashboard**       | Go HTTP + Server-Sent Events | Minimal overhead; real-time                      |
| **CI/CD**           | GitHub Actions               | Already in use; no new services                  |

---

## Implementation Phases

### Phase 1: Foundation (8 hours)

- [ ] Create test orchestrator CLI (`daemon/cmd/testor/main.go`)
- [ ] Add testconfig.yml and loader
- [ ] Create 2-3 sample Go unit tests (daemon + CLI)
- [ ] Implement basic NDJSON logging
- [ ] **Validation**: `brain test daemon` works locally

### Phase 2: Parallel Execution (12 hours)

- [ ] Implement test runner abstraction (interface)
- [ ] Wire Go and TypeScript runners
- [ ] Add file watcher (`--watch` mode)
- [ ] Smart incremental test selection
- [ ] **Validation**: All test types run together, watch works

### Phase 3: Integration Tests (10 hours)

- [ ] Create integration test suite (Go)
- [ ] Add runner to orchestrator
- [ ] Add setup/teardown logic
- [ ] Write 3-5 representative integration tests
- [ ] **Validation**: Multi-layer testing works

### Phase 4: Browser E2E Tests (12 hours)

- [ ] Set up Playwright
- [ ] Create browser test runner
- [ ] Write 3-5 E2E tests
- [ ] Add screenshot capture
- [ ] **Validation**: Full UI stack testing works

### Phase 5: Dashboard & Reporting (8 hours)

- [ ] HTTP dashboard server
- [ ] Real-time log streaming
- [ ] HTML report generation
- [ ] Artifact retention logic
- [ ] **Validation**: Operators see real-time status

### Phase 6: CI/CD Integration (6 hours)

- [ ] Create `.github/workflows/test.yml`
- [ ] Add PR comment bot
- [ ] Artifact upload
- [ ] Update documentation
- [ ] **Validation**: Tests run in GitHub Actions

**Total**: 56 hours (1-2 weeks, parallelizable)

---

## Integration Points

### With Daemon

- Orchestrator calls `daemon start` before integration tests
- Integration tests use daemon's HTTP endpoints
- Daemon logs optionally streamed to test logs

### With CLI

- Integration tests invoke `brain` binary directly
- Capture stdout/stderr and verify output format
- Tests verify CLI reflects daemon state changes

### With Desktop UI

- Browser tests launch UI via Tauri test APIs
- Verify UI reflects daemon state changes
- Screenshot captures on failures

### With GitHub Actions

- Guardian workflow runs in parallel (separate)
- Test workflow publishes results to PR

---

## Alternative Approaches (Why Not Used)

### Option A: Shell Scripts + Go Hybrid ❌

**Cons**: Violates "no shell scripts" principle; harder to debug; not portable

### Option B: Full TypeScript (Jest + Node) ❌

**Cons**: Can't invoke Go binaries easily; duplicates framework complexity; not Go-first

### Option C: Go Orchestrator + Standard Unit Tests ✅ **RECOMMENDED**

Uses Go for orchestration, standard frameworks for each language

---

## Pro/Con Analysis

| Aspect               | Pro                                | Con                                |
| -------------------- | ---------------------------------- | ---------------------------------- |
| **Centralization**   | Single command `brain test`        | Requires new CLI implementation    |
| **Logging**          | Structured NDJSON is tool-friendly | ~5KB per test run                  |
| **Watchers**         | Fast feedback loop                 | Complexity (smart selection logic) |
| **Dashboard**        | Real-time visibility               | Adds HTTP server                   |
| **Playwright E2E**   | Full UI stack testing              | Slower (180s per run)              |
| **No Shell Scripts** | Aligns with repo policy            | More Go code upfront               |
| **Extensibility**    | Easy to add new test types         | Need new runner interface impl     |
| **CI Integration**   | Seamless with GitHub Actions       | Requires secrets management        |

---

## Success Criteria

A test run is **successful** when:

```text
✅ All Go unit tests pass (daemon + CLI)
✅ All TypeScript unit tests pass (React)
✅ All integration tests pass (full-stack)
✅ All browser E2E tests pass (Playwright)
✅ Code coverage >= 70% (daemon), >= 60% (CLI), >= 50% (UI)
✅ No test takes > 30s (E2E can be > 2 min)
✅ Total execution < 10 minutes
✅ All logs are NDJSON structured
✅ HTML report generated and accessible
✅ File watcher mode works (retest on change)
✅ CI integration shows results in PR
```

---

## Effort Breakdown

| Phase     | Component       | Time    | Dependencies                    |
| --------- | --------------- | ------- | ------------------------------- |
| 1         | Orchestrator    | 8h      | —                               |
| 2         | Multi-runner    | 12h     | Phase 1                         |
| 3         | Integration     | 10h     | Phase 2                         |
| 4         | Browser E2E     | 12h     | Phase 2                         |
| 5         | Dashboard       | 8h      | Phase 1-3                       |
| 6         | CI/CD           | 6h      | Phase 5                         |
| **TOTAL** | **Full system** | **56h** | Sequential with parallelization |

**Timeline**: 1-2 weeks (7-10 working days)

---

## Next Steps

**If Approved**:

1. Start Phase 1: Test Orchestrator CLI
2. Migrate existing tests
3. Add sample tests (unit, integration, E2E)
4. Document in TESTING.md for contributors

**For Refinement**:

- Any technology changes?
- Add/remove layers?
- Adjust success criteria?
- Modify implementation phases?

---

**Decision Needed**: Proceed with Phase 1 or refine architecture first?

**Document Status**: AWAITING USER APPROVAL
