# Brain E2E Test Framework

End-to-end tests that exercise the full authentication flow against a **real** running daemon process. No mocks, no Docker, no external services.

## Quick Start

```bash
# Run all E2E tests (~24 seconds, 40 tests)
go test ./test/e2e/... -v

# Run with verbose output and race detection
go test ./test/e2e/... -v -race

# Skip slow tests (tests that start/stop the daemon)
go test ./test/e2e/... -short

# Run a specific test
go test ./test/e2e/... -run TestFullLoginFlow -v

# Run only auth flow tests
go test ./test/e2e/... -run TestAuth -v

# Run only daemon E2E tests
go test ./test/e2e/... -run TestDaemon -v
```

## Requirements

- **Go toolchain** (1.24.4+)
- **CGO_ENABLED=1** (for SQLite driver compilation)
- **No external services required** -- no Docker, Logto, PostgreSQL, or Redis

## What It Tests

### Daemon E2E Tests (`daemon_e2e_test.go`)

| #   | Test                                        | What it verifies                             |
| --- | ------------------------------------------- | -------------------------------------------- |
| 1   | `TestHealthEndpoint`                        | `/health` returns 200                        |
| 2   | `TestBootstrapLogin`                        | Bootstrap login with real daemon             |
| 3   | `TestAuthenticatedRequestAgents`            | Authenticated request to `/api/agents` works |
| 4   | `TestUnauthenticatedRequestAgents`          | Unauthenticated request returns 401          |
| 5   | `TestLogoutRevokesSession`                  | Logout revokes the session                   |
| 6   | `TestRevokedSessionReturns401`              | Revoked session returns 401                  |
| 7   | `TestSessionPersistsAcrossRestart`          | Session survives daemon restart              |
| 8   | `TestRefreshEndpointIssuesNewToken`         | Refresh endpoint issues new token            |
| 9   | `TestRateLimitingBlocksAfterFailedAttempts` | Rate limiting after 5 failed logins          |
| 10  | `TestProtectedRoutesBlockWithoutToken`      | Protected routes block without token         |

### Auth Flow E2E Tests (`auth_flow_e2e_test.go`)

| #   | Test                  | What it verifies                                          |
| --- | --------------------- | --------------------------------------------------------- |
| 1   | `TestFullLoginFlow`   | Login -> get token -> use token -> 200                    |
| 2   | `TestStatusFlow`      | Auth status with token returns user info + capabilities   |
| 3   | `TestLogoutFlow`      | Login -> logout -> 401 on next request                    |
| 4   | `TestPersistenceFlow` | Login -> kill daemon -> restart -> same token works       |
| 5   | `TestRefreshFlow`     | Login -> refresh -> new token works, old token revoked    |
| 6   | `TestInviteFlow`      | Admin creates invite -> consume invite -> new user exists |

## Architecture

```
Test Runner
    |
    +-- TestMain (builds daemon binary once)
    |
    +-- Each test:
        |
        +-- newTestEnv() -> isolated temp dir + free port
        |
        +-- startDaemon() -> exec.Command(daemon)
        |
        +-- waitForDaemonReady() -> poll /health
        |
        +-- HTTP requests -> real daemon on localhost
        |
        +-- stopDaemon() -> signal + kill
        |
        +-- t.TempDir() cleanup -> removes all files
```

### Key Design Decisions

1. **Real daemon binary**: Tests compile and run the actual `braind` binary, not an `httptest.Server`.
2. **Process isolation**: Each test gets its own daemon process, port, and temp directory.
3. **SQLite persistence**: Sessions are stored in SQLite, so they survive daemon restarts.
4. **No external dependencies**: The `local` profile uses SQLite + in-memory sessions only.
5. **Shared binary**: The daemon is built once in `TestMain` and reused by all tests.

## Test Helpers

The `suite_test.go` file provides these helper functions:

| Function                                            | Purpose                                                  |
| --------------------------------------------------- | -------------------------------------------------------- |
| `newTestEnv(t)`                                     | Creates isolated test environment (temp dir + free port) |
| `env.startDaemon()`                                 | Starts daemon process and waits for ready                |
| `env.stopDaemon()`                                  | Stops daemon gracefully (then kills if needed)           |
| `env.restartDaemon()`                               | Stops and restarts, preserving the store                 |
| `env.login(email, password)`                        | Performs bootstrap login, returns token + refresh token  |
| `env.doJSON(method, path, body, headers, respBody)` | Makes HTTP request with JSON encoding/decoding           |
| `authHeader(token)`                                 | Returns `Authorization: Bearer <token>` header map       |
| `requireStatusCode(t, resp, expected)`              | Asserts HTTP status code                                 |
| `skipIfSlow(t)`                                     | Skips test if `-short` flag is set                       |
| `waitForDaemonReady(url, timeout)`                  | Polls `/health` until daemon is ready                    |

## Troubleshooting

### Daemon binary fails to build

Ensure CGO is enabled:

```bash
CGO_ENABLED=1 go test ./test/e2e/... -v
```

### Port already in use

Tests use random free ports via `net.Listen("tcp", "127.0.0.1:0)`. If a port conflict occurs, it is likely a stray daemon process:

```bash
# Find stray braind processes
ps aux | grep braind

# Kill them
pkill -f braind
```

### Health endpoint does not respond

Check if the daemon process started successfully:

```bash
go test ./test/e2e/... -v -run TestHealthEndpoint 2>&1 | head -50
```

### Tests are slow

- Each test starts a new daemon process (takes ~1-2 seconds).
- Use `-short` to skip E2E tests in CI: `go test ./... -short`
- Run specific tests: `go test ./test/e2e/... -run TestHealth -v`

### Session does not persist after restart

Verify that the `BRAIN_AUTH_STORE_PATH` environment variable points to a persistent SQLite file, not a temp directory that gets cleaned up between restarts. The test framework handles this automatically.

## File Structure

```
test/
  go.mod                          # Test module dependencies
  README.md                       # This file
  e2e/
    suite_test.go                 # Test infrastructure, helpers, types
    daemon_e2e_test.go            # Daemon HTTP E2E tests
    auth_flow_e2e_test.go         # Full auth flow E2E tests
    testdata/
      test-config.yml             # Test configuration reference
```
