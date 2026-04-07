# 🧪 Brain Testing System - Industry-Backed Implementation Guide

**Document Date**: April 3, 2026  
**Based On**: Google Bazel, Jest, Go Testing, Playwright best practices  
**Status**: READY FOR IMPLEMENTATION

---

## Executive Summary

This guide consolidates industry best practices (Google Bazel, Meta Jest, Go stdlib) into a centralized testing system for Brain that:

✅ **Eliminates unnecessary test runs** - Smart test filtering (only run affected tests)  
✅ **Supports dev-to-prod equivalence** - Same tests run locally, CI, and production  
✅ **Maintains security** - Isolation, secret management, sandboxing  
✅ **Keeps code clean** - Deterministic, hermetic, reentrant tests  
✅ **Provides clear feedback** - Structured logging, real-time reporting

---

## Part 1: Smart Test Filtering (The Game Changer)

### Problem: Running All Tests on Every Change is Wasteful

```bash
# Current problem:
$ git commit -m "fix typo in SkillsList.tsx"
$ brain test  # Runs 156 total tests (including slow E2E)
# ❌ 9 minutes of waiting for test feedback

# Desired outcome:
$ git commit -m "fix typo in SkillsList.tsx"
$ brain test  # Runs only 8 affected tests
# ✅ 15 seconds of feedback
```

### Solution: Build Dependency Graph

**How It Works** (from Bazel + Jest patterns):

#### Track file dependencies

```text
SkillsList.tsx
├─ depends on: api.ts, useSkills hook, types.ts
├─ has tests: SkillsList.test.tsx
└─ consumed by: desktop/src/App.tsx, dashboard.spec.ts (E2E)
```

#### On file change, find affected tests

```text
# File changed: desktop/src/components/SkillsList.tsx
# Affected tests:
#   - desktop/src/components/SkillsList.test.tsx (direct)
#   - desktop/e2e/dashboard.spec.ts (E2E - depends on SkillsList)
#   - integration tests that touch UI (optional)
```

#### Run only affected

```bash
brain test --onlyChanged    # Jest pattern
# OR
brain test --findRelatedTests desktop/src/components/SkillsList.tsx  # Explicit
```

### Implementation in Brain

**Layer: Smart Test Selection Engine** (new file)

**Location**: `daemon/cmd/testor/dependency_graph.go`

```go
type DependencyGraph struct {
    // Map: source file -> affected tests
    // Example: "desktop/src/utils/api.ts" -> ["api.test.ts", "integration/e2e.spec.ts"]
    FileToTests map[string][]string

    // Map: test file -> dependencies
    // Example: "daemon_unit" -> ["daemon/internal/*", "daemon/cmd/*"]
    TestToDependencies map[string][]string
}

// scanProjectDependencies reads package.json, go.mod, tsconfig.json
// builds graph of import relationships
func (g *DependencyGraph) ScanProjectDependencies(rootPath string) error {
    // For Go: parse imports, use go mod graph
    // For TS: parse tsconfig.json, analyze imports via ts-morph
    // Output: graph.json
}

// affectedTests returns test files affected by changed source files
func (g *DependencyGraph) AffectedTests(changedFiles []string) []string {
    var affected []string
    for _, file := range changedFiles {
        affected = append(affected, g.FileToTests[file]...)
    }
    return deduplicateAndSort(affected)
}

// isHermetic checks if test depends ONLY on declared inputs
func (g *DependencyGraph) Verify() error {
    // Prevent hidden dependencies
    // Example: test should not silently depend on global state
}
```

**Test Running Logic**:

```go
// In orchestrator:
func (o *Orchestrator) RunTests(ctx context.Context, watch bool) error {
    var targetTests []string

    if watch && hasChangedFiles() {
        changedFiles := detectChangedFiles()  // git diff, or --onlyChanged flag
        graph := LoadDependencyGraph()
        targetTests = graph.AffectedTests(changedFiles)
    } else {
        targetTests = allTestsInConfig()
    }

    // Run only targetTests
    return o.ExecuteRunners(ctx, targetTests)
}
```

**Config Example**:

```yaml
test_suites:
  daemon_unit:
    runner: go_test
    path: ./daemon/...
    timeout: 60s
    dependencies:
      - "daemon/**/*.go"

  ui_unit:
    runner: typescript_test
    path: ./desktop
    timeout: 45s
    dependencies:
      - "desktop/src/**/*.tsx"
      - "desktop/src/**/*.ts"
```

### Result: What Users See

```bash
$ brain test --watch
# Automatic mode: scans git diff

[10:15:22] Detected change: desktop/src/components/SkillsList.tsx

[10:15:22] ✅ Dependency graph loaded
[10:15:22] Affected tests:
  • desktop/src/components/SkillsList.test.tsx
  • desktop/e2e/SkillsList.spec.ts
[10:15:22] Total: 2 affected tests (out of 156 in suite)

[10:15:23] Running affected tests...
[10:15:25] ✅ PASS SkillsList.test.tsx (1.2s)
[10:15:35] ✅ PASS SkillsList.spec.ts (Playwright) (9.8s)

[10:15:35] ✅ All affected tests passed (11s total)
```

---

## Part 2: Test Isolation & Security

### Problem: Tests Interfering With Each Other

```go
// ❌ BAD: Global state pollution
var testCounter int = 0
func TestIncrement(t *testing.T) {
    testCounter++  // Affects other tests!
}

// ✅ GOOD: Isolated per test
func TestIncrement(t *testing.T) {
    tempDir := t.TempDir()  // Isolated temp directory
    t.Setenv("BRAIN_TEST", "true")  // Reverts after test
    // Test code here
    // All state cleaned up automatically
}
```

### Patterns for Brain

**Go Tests (Daemon/CLI)**:

```go
// Pattern 1: Use t.TempDir() for filesystem tests
func TestCreateSkillWithFile(t *testing.T) {
    tmpDir := t.TempDir()  // Auto-cleaned after test

    // Test code writes to tmpDir
    err := createSkill(tmpDir, skillData)
    assert.NoError(t, err)
}

// Pattern 2: Use t.Setenv() for environment isolation
func TestWithCustomConfig(t *testing.T) {
    t.Setenv("BRAIN_HOME", "/tmp/test-brain")
    t.Setenv("LOG_LEVEL", "debug")
    // Environment reverts automatically after test
    // No side effects on other tests
}

// Pattern 3: Use subtests for hierarchical cleanup
func TestSkillManager(t *testing.T) {
    manager := NewSkillManager(t.TempDir())

    t.Run("create", func(t *testing.T) {
        // Setup
        skill := &Skill{ID: "test"}
        err := manager.Create(skill)
        assert.NoError(t, err)

        // Cleanup happens after this subtest
    })

    t.Run("delete", func(t *testing.T) {
        // Independent test; no state from "create"
    })
}

// Pattern 4: Mark helpers to skip stack trace
func createTestSkill(t *testing.T, id string) *Skill {
    t.Helper()  // Stack trace points to caller, not helper
    return &Skill{ID: id}
}
```

**TypeScript Tests (React)**:

```typescript
// Pattern 1: Isolated test environment (jsdom)
describe('SkillsList', () => {
    let apiMock: jest.Mock;

    beforeEach(() => {
        apiMock = jest.fn();
        jest.resetModules();  // Clear module cache
    });

    afterEach(() => {
        jest.clearAllMocks();  // Explicit cleanup
    });

    test('renders skills', () => {
        // Test isolated from others
    });
});

// Pattern 2: Component isolation with mocks
test('SkillForm submits data', () => {
    const mockOnSubmit = jest.fn();
    render(
        <SkillForm onSubmit={mockOnSubmit} />
    );
    // Only testing SkillForm, not dependencies
});
```

**Playwright E2E (Browser)**:

```typescript
// Pattern 1: Test isolation via separate browser contexts
test("create skill", async ({ browser }) => {
  const context = await browser.newContext(); // Isolated session
  const page = await context.newPage();

  // Test code

  await context.close(); // All state cleaned up
});

// Pattern 2: API mocking to prevent side effects
test("shows error on API failure", async ({ page }) => {
  await page.route("**/api/skills", (route) => {
    route.abort("failed"); // Mock failure
  });

  // Test error handling without hitting real API
});
```

### Security Checklist

- ✅ **No hardcoded secrets**: Use `t.Setenv()` or mock providers
- ✅ **Isolated temp dirs**: `t.TempDir()` auto-cleaned
- ✅ **Module reset**: `jest.resetModules()` prevents pollution
- ✅ **Environment reversion**: `t.Setenv()` reverts automatically
- ✅ **Mock external services**: Don't hit real APIs in tests
- ✅ **Deterministic**: Same input → same output (no random state)
- ✅ **Hermetic**: Only depends on declared inputs

---

## Part 3: Dev-to-Prod Pipeline

### Unified Test Execution Across Environments

**Same test code, different configurations**:

```bash
# LOCAL DEVELOPMENT
$ brain test --watch
# - Only affected tests run
# - Interactive feedback
# - File watchers enabled
# - Fast feedback loop

# PRE-COMMIT
$ brain test --check
# - All tests in affected suite run
# - Exit code 0 = safe to commit
# - Exit code 1 = fix before committing

# CI/CD (GitHub Actions)
$ brain test --ci-mode
# - Full suite runs
# - Structured JSON output
# - Artifacts uploaded
# - Failure comments on PR

# PRODUCTION SMOKE TESTS
$ brain test --smoke
# - Quick validation tests only
# - Tags filter: `+smoke` tests
# - Checks critical paths
# - ~1 minute execution
```

### Example: Full Pipeline

**File changes**:

```bash
git commit -m "fix(skills): handle empty description"
```

**Local development**:

```bash
$ brain test --watch
# Runs: SkillsList.test.tsx + integration tests
# Time: 12 seconds
# Result: ✅ All passed
```

**Pre-commit hook** (`.git/hooks/pre-commit`):

```bash
#!/bin/bash
brain test --findRelatedTests HEAD  # Only changed tests
exit_code=$?
[[ $exit_code -eq 0 ]] && git commit || echo "Fix tests before committing"
```

**CI/CD** (`.github/workflows/test.yml`):

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: brain test --ci-mode --report json
      - uses: actions/upload-artifact@v4
        with:
          name: test-results
          path: .logs/
```

**Production deployment**:

```bash
# On deployment to production:
$ brain test --smoke --tag=production
# Runs only critical tests
# Validates: API connectivity, data migrations, security checks
# Time: ~1 minute
# Fails fast if critical path broken
```

**Result**: Same test code, optimized for context

---

## Part 4: Test Phases with Industry Patterns

### Phase 1: Foundation (Go Patterns)

**File**: `daemon/cmd/testor/main.go`

Key features from Go testing best practices:

```go
// Parallel test execution
func TestMultipleSkills(t *testing.T) {
    t.Parallel()  // Run alongside other parallel tests
    // Go automatically manages concurrency
}

// Subtests for organization
func TestSkillOperations(t *testing.T) {
    tests := []struct {
        name string
        id   string
    }{
        {"create", "skill-1"},
        {"update", "skill-2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Each subtest can run in parallel
            // Test code
        })
    }
}

// Context integration for timeouts
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
    defer cancel()

    // AutoTest respects context deadline
    runTestWithDeadline(ctx)
}

// Cleanup in LIFO order
func TestWithCleanup(t *testing.T) {
    setupA()
    t.Cleanup(func() { teardownA() })

    setupB()
    t.Cleanup(func() { teardownB() })

    // Cleanup order: B, A (LIFO)
}
```

### Phase 2: Multi-Runner (Bazel Parallelism)

From Bazel test encyclopedia:

````yaml
# testconfig.yml - Size-based resource allocation

test_suites:
  daemon_unit:
    size: small # Runs in parallel, <60s
    timeout: 60s
    resources: {} # Can share resources

  integration:
    size: medium # Runs sequentially if needed, <300s
    timeout: 300s
    resources:
      cpu: 2 # Requires 2 CPU cores

  browser_e2e:
    size: large # Long-running, <900s
    timeout: 900s
    resources:
      cpu: 4
      memory: 2Gi
```text

**Execution strategy**:

```text
Parallel:
├─ daemon_unit (small, no resources)
├─ cli_unit (small)
└─ ui_unit (medium)

Sequential:
└─ integration (depends on unit, medium)
└─ browser_e2e (depends on integration, large)
````

### Phase 3-4: Test Structure

**Go Integration Tests** (from stdlib patterns):

```go
// File: daemon/tests/integration_test.go

func TestIntegration(t *testing.T) {
    // Setup:shared resources
    tmpDir := t.TempDir()
    daemon, err := startDaemon(tmpDir)
    require.NoError(t, err)
    defer daemon.Stop()

    // Subtests for different workflows
    t.Run("CLI creates skill via API", func(t *testing.T) {
        // Test: CLI can create
    })

    t.Run("Daemon persists to disk", func(t *testing.T) {
        // Test: files written
    })

    t.Run("Sync detects changes", func(t *testing.T) {
        // Test: validation works
    })
}
```

**TypeScript Jest Config** (from Jest patterns):

```javascript
// jest.config.js
module.exports = {
  testEnvironment: "jsdom", // Browser-like environment
  testTimeout: 5000, // 5s timeout
  maxWorkers: "50%", // Use half available cores
  collectCoverageFrom: [
    "src/**/*.{ts,tsx}",
    "!src/index.ts", // Skip entry files
  ],
  displayName: "Frontend Tests", // Clear output
  verbose: true, // Detailed output
};
```

**Playwright Config** (from framework patterns):

```typescript
// playwright.config.ts
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  workers: 4, // Run 4 tests in parallel
  timeout: 30000, // 30s per test
  retries: 1, // Retry flakies once
  use: {
    baseURL: "http://localhost:5173",
    screenshot: "only-on-failure", // Save on error
    video: "retain-on-failure", // Record on error
  },
});
```

### Phase 5: Reporting

**NDJSON Output** (structured, parseable):

```json
{"ts":"2026-04-03T10:15:22Z","event":"test_start","suite":"daemon_unit"}
{"ts":"2026-04-03T10:15:23Z","event":"test_pass","test":"TestCreate","duration_ms":234}
{"ts":"2026-04-03T10:15:45Z","event":"suite_complete","passed":34,"failed":0}
```

**Aggregation for CI**:

```json
{
  "summary": {
    "total": 156,
    "passed": 154,
    "failed": 2,
    "skipped": 0
  },
  "suites": [
    {
      "name": "daemon_unit",
      "passed": 34,
      "failed": 0
    }
  ],
  "duration_ms": 45000,
  "exit_code": 1
}
```

---

## Part 5: Test Maintenance & Flakiness

### 🚩 Flaky Tests Are a Crime

From Bazel test encyclopedia:

```go
// ❌ FLAKY: Depends on timing
func TestWithoutWait(t *testing.T) {
    go doAsync()
    time.Sleep(10 * time.Millisecond)  // Might not be enough
    assert.True(t, isDone())
}

// ✅ DETERMINISTIC: Waits for condition
func TestWithWait(t *testing.T) {
    go doAsync()
    waitFor(t, func() bool { return isDone() }, 5*time.Second)
}

// Helper function
func waitFor(t *testing.T, fn func() bool, timeout time.Duration) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    for {
        if fn() {
            return
        }
        select {
        case <-ctx.Done():
            t.Fatalf("timeout waiting for condition")
        case <-time.After(10 * time.Millisecond):
        }
    }
}
```

### Playwright Reliability

```typescript
// ❌ FLAKY: Assumes element exists immediately
test("click button", async ({ page }) => {
  await page.click("button"); // May not exist yet
});

// ✅ RELIABLE: Waits for element
test("click button", async ({ page }) => {
  await page.locator("button").waitFor({ state: "visible" });
  await page.locator("button").click();
});

// Or use built-in waits:
test("verify content", async ({ page }) => {
  await expect(page.locator("h1")).toContainText("Skills");
});
```

---

## Part 6: Recommended Tool Stack

| Component             | Tool                    | Why                   | Industry Use             |
| --------------------- | ----------------------- | --------------------- | ------------------------ |
| **Testing Framework** | Go `testing` stdlib     | Built-in, no deps     | Google, Uber, CloudFlare |
| **Test Filtering**    | Custom dependency graph | Smart incremental     | Bazel model              |
| **TypeScript Tests**  | Vitest                  | Fast, zero config     | Next.js, Vite ecosystem  |
| **Browser E2E**       | Playwright              | Powerful, reliable    | Meta, Google, Microsoft  |
| **Logging**           | NDJSON custom           | Structured, parseable | CI/CD standard           |
| **Orchestration**     | Go CLI                  | Type-safe, fast       | Brain's architecture     |
| **CI/CD**             | GitHub Actions          | This is where code is | Already using            |

---

## Part 7: Implementation Roadmap

### Phase 1: Foundation & Smart Filtering (Week 1)

- [ ] Implement `daemon/cmd/testor/main.go`
- [ ] Build dependency graph scanner
- [ ] Add test filtering (`--onlyChanged`, `--findRelatedTests`)
- [ ] Create `testconfig.yml`
- [ ] Write 3-5 initial Go tests
- **Validation**: `brain test --watch` works, detects file changes

### Phase 2: Multi-Runner + Isolation (Week 1-2)

- [ ] Wire TypeScript + Playwright runners
- [ ] Implement test isolation patterns
- [ ] Add NDJSON logging
- [ ] Parallel execution with resource limits
- **Validation**: All test types run in parallel without interference

### Phase 3: Integration Tests (Week 2)

- [ ] Full-stack integration tests (Go)
- [ ] Setup/teardown for daemon lifecycle
- [ ] File system mocking
- **Validation**: Integration tests pass, cover daemon + CLI + UI flow

### Phase 4: Production-Ready Features (Week 3)

- [ ] Dashboard + real-time reporting
- [ ] Artifact persistence + cleanup
- [ ] Test retry for flakies
- [ ] Performance profiling
- **Validation**: Can debug failed tests from weeks ago

### Phase 5: CI/CD Integration (Week 3-4)

- [ ] GitHub Actions workflow
- [ ] PR comment bot with results
- [ ] Artifact upload
- [ ] Production smoke tests
- **Validation**: Tests run automatically on push, report to PR

---

## Part 8: Security Checklist

- ✅ **No secrets in code**: Use environment variables
- ✅ **Isolated execution**: `t.TempDir()`, mocked APIs
- ✅ **Deterministic tests**: No random state, fixed seeds
- ✅ **Module isolation**: `jest.resetModules()`, no global state
- ✅ **Cleanup in LIFO**: Resources released in reverse order
- ✅ **Timeout enforcement**: All tests have timeouts
- ✅ **Flakiness prevented**: Wait for conditions, not time
- ✅ **Coverage tracked**: Minimum thresholds enforced

---

## Success Criteria

When implemented, you'll have:

```text
✅ brain test                 # Runs all tests
✅ brain test --watch        # Watches files, reruns affected only
✅ brain test --check        # Pre-commit validation
✅ brain test --ci-mode      # CI/CD mode
✅ brain test --smoke        # Production smoke tests
✅ brain test --filter NAME  # Run specific test

Output:
✅ Real-time feedback (no more 9-minute waits)
✅ Structured NDJSON logging
✅ HTML reports + artifacts
✅ Dashboard on localhost:9091
✅ GitHub Actions integration
✅ No flaky tests (deterministic)
✅ Isolated test execution (no pollution)
✅ Clear debugging (failed test artifacts saved)
```

---

## Next Steps

1. **Review this guide** - Confirm all patterns align with your goals
2. **Start Phase 1** - Implement orchestrator + smart filtering
3. **Integrate with daemon** - Wire into existing Go CLI
4. **Add sample tests** - Convert existing tests to new system
5. **Document for team** - TESTING.md guide for contributors

**Ready to proceed?**

Document Status: ✅ READY FOR IMPLEMENTATION
