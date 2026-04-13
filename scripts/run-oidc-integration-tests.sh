#!/usr/bin/env bash
# run-oidc-integration-tests.sh
# Full E2E OIDC integration test suite.
#
# Starts a local Logto instance via Docker Compose, initializes it,
# starts the Brain daemon, and runs the full OIDC test suite.
#
# Usage:
#   ./scripts/run-oidc-integration-tests.sh          # Run all tests
#   ./scripts/run-oidc-integration-tests.sh --clean   # Only clean up resources
#   SKIP_OIDC_TESTS=true ./scripts/run-oidc-integration-tests.sh  # Skip tests
#
# Environment overrides:
#   LOGTO_PG_PASSWORD  - PostgreSQL password (default: logto_password_dev)
#   LOGTO_ADMIN_EMAIL  - Admin email (default: admin@brain.local)
#   LOGTO_ADMIN_PASSWORD - Admin password (default: Admin123456!)
#   DAEMON_PORT        - Daemon HTTP port (default: 9090)
#   CLEANUP_ON_EXIT    - Whether to clean up after tests (default: true)

set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_FILE="${PROJECT_DIR}/docker-compose.logto.yml"
SETUP_SCRIPT="${SCRIPT_DIR}/setup-logto.sh"

LOGTO_PG_PASSWORD="${LOGTO_PG_PASSWORD:-logto_password_dev}"
LOGTO_ADMIN_EMAIL="${LOGTO_ADMIN_EMAIL:-admin@brain.local}"
LOGTO_ADMIN_PASSWORD="${LOGTO_ADMIN_PASSWORD:-Admin123456!}"
DAEMON_PORT="${DAEMON_PORT:-9090}"
CLEANUP_ON_EXIT="${CLEANUP_ON_EXIT:-true}"

LOGTO_ADMIN_URL="http://127.0.0.1:3001"
LOGTO_ISSUER_URL="http://127.0.0.1:3002/oidc"

DAEMON_BINARY="${PROJECT_DIR}/apps/daemon/cmd/braind/braind"
DAEMON_PID=""
TESTS_PASSED=0
TESTS_FAILED=0

# ── Colors ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${CYAN}[oidc-e2e]${NC} $*"; }
ok()      { echo -e "${GREEN}[oidc-e2e]${NC} $*"; }
warn()    { echo -e "${YELLOW}[oidc-e2e]${NC} $*"; }
error()   { echo -e "${RED}[oidc-e2e]${NC} $*"; }
section() { echo -e "\n${BOLD}${CYAN}=== $* ===${NC}\n"; }

# ── Cleanup ───────────────────────────────────────────────────────────────────
cleanup() {
    local exit_code=${1:-0}

    if [ "$CLEANUP_ON_EXIT" != "true" ]; then
        warn "Cleanup skipped (CLEANUP_ON_EXIT=false)"
        exit "$exit_code"
    fi

    section "Cleanup"

    # Kill daemon if running
    if [ -n "$DAEMON_PID" ] && kill -0 "$DAEMON_PID" 2>/dev/null; then
        info "Stopping daemon (PID: ${DAEMON_PID})..."
        kill "$DAEMON_PID" 2>/dev/null || true
        wait "$DAEMON_PID" 2>/dev/null || true
        ok "Daemon stopped"
    fi

    # Kill any remaining daemon on our port
    local daemon_pids
    daemon_pids=$(lsof -ti :${DAEMON_PORT} 2>/dev/null) || true
    if [ -n "$daemon_pids" ]; then
        info "Killing remaining processes on port ${DAEMON_PORT}..."
        echo "$daemon_pids" | xargs kill -9 2>/dev/null || true
    fi

    # Stop and remove Logto containers
    info "Stopping Logto containers..."
    docker compose -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true

    # Remove volumes for clean state
    info "Removing Logto volumes..."
    docker compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true

    # Clean up temp files
    rm -f /tmp/logto-cookies.txt
    rm -f /tmp/logto-env.sh
    rm -f /tmp/brain-auth.sqlite

    # Unset env vars we may have set
    unset BRAIN_OIDC_ISSUER_URL
    unset BRAIN_OIDC_CLIENT_ID
    unset BRAIN_OIDC_CLIENT_SECRET
    unset BRAIN_OIDC_REDIRECT_URL
    unset BRAIN_OIDC_PROVIDER
    unset BRAIN_OIDC_ENABLED
    unset BRAIN_AUTH_MODE
    unset BRAIN_ENV

    ok "Cleanup complete"
}

trap 'cleanup $?' EXIT

# ── Prerequisite Checks ───────────────────────────────────────────────────────
check_prerequisites() {
    section "Checking Prerequisites"

    local missing=0

    if ! command -v docker &>/dev/null; then
        error "docker is not installed"
        missing=1
    fi

    if ! docker compose version &>/dev/null 2>&1; then
        error "docker compose plugin is not installed"
        missing=1
    fi

    if ! command -v curl &>/dev/null; then
        error "curl is not installed"
        missing=1
    fi

    if ! command -v go &>/dev/null; then
        error "go is not installed (required for running tests)"
        missing=1
    fi

    if [ ! -f "$COMPOSE_FILE" ]; then
        error "docker-compose.logto.yml not found at ${COMPOSE_FILE}"
        missing=1
    fi

    if [ ! -f "$SETUP_SCRIPT" ]; then
        error "setup-logto.sh not found at ${SETUP_SCRIPT}"
        missing=1
    fi

    if [ $missing -ne 0 ]; then
        error "Prerequisites not met. Install missing tools and try again."
        exit 1
    fi

    ok "All prerequisites met"
}

# ── Start Logto ───────────────────────────────────────────────────────────────
start_logto() {
    section "Starting Logto"

    info "Starting Logto containers via docker compose..."
    export LOGTO_PG_PASSWORD
    export LOGTO_ADMIN_EMAIL
    export LOGTO_ADMIN_PASSWORD

    docker compose -f "$COMPOSE_FILE" up -d --wait --wait-timeout 180

    ok "Logto containers started"
}

# ── Setup Logto Application ───────────────────────────────────────────────────
setup_logto() {
    section "Setting up Logto"

    info "Running setup-logto.sh..."
    local env_file="/tmp/logto-env.sh"

    export LOGTO_ADMIN_URL
    export LOGTO_ISSUER_URL
    export LOGTO_ADMIN_EMAIL
    export LOGTO_ADMIN_PASSWORD

    # Capture the setup output and extract env vars
    local setup_output
    setup_output=$("$SETUP_SCRIPT" 2>&1) || {
        error "Logto setup failed!"
        echo "$setup_output"
        return 1
    }

    echo "$setup_output"

    # Extract client_id and client_secret from output
    local client_id client_secret
    client_id=$(echo "$setup_output" | grep "BRAIN_OIDC_CLIENT_ID=" | head -1 | cut -d'=' -f2)
    client_secret=$(echo "$setup_output" | grep "BRAIN_OIDC_CLIENT_SECRET=" | head -1 | cut -d'=' -f2)

    if [ -z "$client_id" ] || [ -z "$client_secret" ]; then
        warn "Could not extract credentials from setup output"
        warn "Attempting to parse from Logto API directly..."

        # Try to get credentials directly
        # This is a fallback - the setup script should have output them
        client_id=$(echo "$setup_output" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        client_secret=$(echo "$setup_output" | grep -o '"clientSecret":"[^"]*"' | head -1 | cut -d'"' -f4)
    fi

    if [ -n "$client_id" ]; then
        export BRAIN_OIDC_CLIENT_ID="$client_id"
        export BRAIN_OIDC_CLIENT_SECRET="$client_secret"
        export BRAIN_OIDC_ISSUER_URL="$LOGTO_ISSUER_URL"
        export BRAIN_OIDC_REDIRECT_URL="http://127.0.0.1:${DAEMON_PORT}/api/auth/oidc/callback"
        export BRAIN_OIDC_PROVIDER="logto"
        export BRAIN_OIDC_ENABLED="true"

        info "Exported environment variables:"
        info "  BRAIN_OIDC_ISSUER_URL=${BRAIN_OIDC_ISSUER_URL}"
        info "  BRAIN_OIDC_CLIENT_ID=${BRAIN_OIDC_CLIENT_ID}"
        info "  BRAIN_OIDC_REDIRECT_URL=${BRAIN_OIDC_REDIRECT_URL}"

        ok "Logto setup complete"
        return 0
    else
        error "Failed to extract OIDC credentials"
        return 1
    fi
}

# ── Build and Start Daemon ────────────────────────────────────────────────────
start_daemon() {
    section "Starting Brain Daemon"

    # Build the daemon
    info "Building daemon..."
    (cd "${PROJECT_DIR}/apps/daemon" && go build -o "${DAEMON_BINARY}" ./cmd/braind/) || {
        error "Failed to build daemon"
        return 1
    }

    # Set environment for daemon
    export BRAIN_ENV="development"
    export BRAIN_AUTH_MODE="oidc"
    export BRAIN_AUTH_REQUIRED="true"

    # Start daemon in background
    info "Starting daemon on port ${DAEMON_PORT}..."
    "${DAEMON_BINARY}" --http-addr ":${DAEMON_PORT}" &
    DAEMON_PID=$!

    # Wait for daemon to be ready
    local max_wait=30
    local elapsed=0
    info "Waiting for daemon to be ready..."
    while [ $elapsed -lt $max_wait ]; do
        if curl -sf "http://127.0.0.1:${DAEMON_PORT}/health" -o /dev/null 2>/dev/null; then
            ok "Daemon is ready (PID: ${DAEMON_PID})"
            return 0
        fi
        # Check if process is still running
        if ! kill -0 "$DAEMON_PID" 2>/dev/null; then
            error "Daemon process exited unexpectedly"
            return 1
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done

    error "Daemon did not become ready within ${max_wait}s"
    return 1
}

# ── Run Tests ─────────────────────────────────────────────────────────────────
run_tests() {
    section "Running OIDC Integration Tests"

    # Export OIDC env vars for Go tests
    export BRAIN_OIDC_ISSUER_URL="${BRAIN_OIDC_ISSUER_URL:-$LOGTO_ISSUER_URL}"
    export BRAIN_OIDC_CLIENT_ID="${BRAIN_OIDC_CLIENT_ID:-}"
    export BRAIN_OIDC_CLIENT_SECRET="${BRAIN_OIDC_CLIENT_SECRET:-}"
    export BRAIN_OIDC_REDIRECT_URL="${BRAIN_OIDC_REDIRECT_URL:-http://127.0.0.1:${DAEMON_PORT}/api/auth/oidc/callback}"

    info "Running Go OIDC integration tests..."
    local test_exit=0

    (cd "${PROJECT_DIR}/apps/daemon" && \
        BRAIN_OIDC_ISSUER_URL="$BRAIN_OIDC_ISSUER_URL" \
        BRAIN_OIDC_CLIENT_ID="$BRAIN_OIDC_CLIENT_ID" \
        BRAIN_OIDC_CLIENT_SECRET="$BRAIN_OIDC_CLIENT_SECRET" \
        BRAIN_OIDC_REDIRECT_URL="$BRAIN_OIDC_REDIRECT_URL" \
        SKIP_OIDC_TESTS="${SKIP_OIDC_TESTS:-false}" \
        go test -v -run TestOIDC -timeout 120s ./cmd/braind/ 2>&1) || test_exit=$?

    if [ $test_exit -eq 0 ]; then
        ok "All OIDC tests passed!"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        error "OIDC tests failed (exit code: ${test_exit})"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Test daemon OIDC endpoints directly
    section "Testing Daemon OIDC Endpoints"

    # Test /api/auth/oidc/start (should return 501 if not fully configured or 200 if configured)
    info "Testing OIDC start endpoint..."
    local start_response
    start_response=$(curl -sf -w "\n%{http_code}" "http://127.0.0.1:${DAEMON_PORT}/api/auth/oidc/start" 2>/dev/null) || true

    local start_code
    start_code=$(echo "$start_response" | tail -1)
    if [ "$start_code" = "200" ] || [ "$start_code" = "501" ] || [ "$start_code" = "503" ]; then
        ok "OIDC start endpoint responded (HTTP ${start_code})"
    else
        warn "OIDC start endpoint returned unexpected code: ${start_code:-no response}"
    fi

    # Test /api/auth/oidc/callback with missing params
    info "Testing OIDC callback with missing params..."
    local callback_response
    callback_response=$(curl -sf -w "\n%{http_code}" "http://127.0.0.1:${DAEMON_PORT}/api/auth/oidc/callback" 2>/dev/null) || true

    local callback_code
    callback_code=$(echo "$callback_response" | tail -1)
    if [ "$callback_code" = "400" ]; then
        ok "OIDC callback correctly returned 400 for missing params"
    else
        warn "OIDC callback returned unexpected code: ${callback_code:-no response}"
    fi

    # Test /api/auth/oidc/poll with state
    info "Testing OIDC poll endpoint..."
    local poll_response
    poll_response=$(curl -sf -w "\n%{http_code}" "http://127.0.0.1:${DAEMON_PORT}/api/auth/oidc/poll?state=test-state" 2>/dev/null) || true

    local poll_code
    poll_code=$(echo "$poll_response" | tail -1)
    if [ "$poll_code" = "202" ]; then
        ok "OIDC poll correctly returned 202 (waiting for login)"
    else
        warn "OIDC poll returned unexpected code: ${poll_code:-no response}"
    fi
}

# ── Summary ───────────────────────────────────────────────────────────────────
print_summary() {
    section "Test Summary"

    echo -e "  ${GREEN}Passed: ${TESTS_PASSED}${NC}"
    echo -e "  ${RED}Failed: ${TESTS_FAILED}${NC}"
    echo ""

    if [ $TESTS_FAILED -eq 0 ]; then
        ok "All tests completed successfully!"
        exit 0
    else
        error "Some tests failed. Check the output above for details."
        exit 1
    fi
}

# ── Main ──────────────────────────────────────────────────────────────────────
main() {
    # Handle --clean flag
    if [ "${1:-}" = "--clean" ]; then
        info "Running cleanup only..."
        CLEANUP_ON_EXIT=true
        exit 0
    fi

    if [ "${SKIP_OIDC_TESTS:-}" = "true" ]; then
        warn "OIDC tests are disabled (SKIP_OIDC_TESTS=true)"
        exit 0
    fi

    info "OIDC Integration Test Suite"
    info "Project: ${PROJECT_DIR}"
    echo ""

    check_prerequisites
    start_logto
    setup_logto
    start_daemon
    run_tests
    print_summary
}

main "$@"
