# OIDC Setup Guide

This guide explains how to set up and test OIDC authentication for the Brain project using a local Logto instance.

## Overview

Brain uses [Logto](https://logto.io) as its default identity provider. OIDC Authorization Code + PKCE is the transport contract for authentication. This setup provides a fully local OIDC provider for development and testing.

### Architecture

```
Browser/CLI  -->  Brain Daemon (:9090)  -->  Logto (:3002 OIDC / :3001 Admin)
                                             |
                                      PostgreSQL (:5432)
```

## Quick Start

Run the full end-to-end test suite (starts Logto, configures it, runs tests, cleans up):

```bash
./scripts/run-oidc-integration-tests.sh
```

To skip tests and only set up Logto:

```bash
SKIP_OIDC_TESTS=true ./scripts/run-oidc-integration-tests.sh
```

## Manual Setup

### Step 1: Start Logto

```bash
docker compose -f docker-compose.logto.yml up -d
```

This starts:
- **PostgreSQL** on port 5432 (Logto's database)
- **Logto Admin Console** on port 3001
- **Logto OIDC Endpoints** on port 3002

Wait ~30 seconds for Logto to initialize. Check status:

```bash
docker compose -f docker-compose.logto.yml ps
curl -sf http://127.0.0.1:3002/oidc/.well-known/openid-configuration | head -5
```

### Step 2: Initialize the OIDC Application

```bash
./scripts/setup-logto.sh
```

This script:
1. Waits for Logto to be ready
2. Authenticates as the admin user
3. Creates a Machine-to-Machine OIDC application named `brain-daemon`
4. Creates a test user (`test@brain.local` / `Test123456!`)
5. Outputs the required environment variables

**Output example:**

```
OIDC Configuration:
  BRAIN_OIDC_ISSUER_URL=http://127.0.0.1:3002/oidc
  BRAIN_OIDC_CLIENT_ID=<generated-id>
  BRAIN_OIDC_CLIENT_SECRET=<generated-secret>
  BRAIN_OIDC_REDIRECT_URL=http://127.0.0.1:9090/api/auth/oidc/callback
  BRAIN_OIDC_PROVIDER=logto
  BRAIN_OIDC_ENABLED=true

Test User Credentials:
  Email:    test@brain.local
  Password: Test123456!

Admin Console:
  URL:      http://127.0.0.1:3001
  Email:    admin@brain.local
  Password: Admin123456!
```

### Step 3: Configure the Daemon

Export the environment variables from Step 2:

```bash
export BRAIN_OIDC_ISSUER_URL="http://127.0.0.1:3002/oidc"
export BRAIN_OIDC_CLIENT_ID="<from-setup-output>"
export BRAIN_OIDC_CLIENT_SECRET="<from-setup-output>"
export BRAIN_OIDC_REDIRECT_URL="http://127.0.0.1:9090/api/auth/oidc/callback"
export BRAIN_OIDC_PROVIDER="logto"
export BRAIN_OIDC_ENABLED="true"
export BRAIN_AUTH_MODE="oidc"
export BRAIN_ENV="development"
```

### Step 4: Run the Daemon

```bash
cd apps/daemon
go run ./cmd/braind/ --http-addr ":9090"
```

### Step 5: Test OIDC Login

Start an OIDC login flow:

```bash
curl http://127.0.0.1:9090/api/auth/oidc/start?login_hint=test@brain.local
```

Response:

```json
{
  "success": true,
  "provider": "logto",
  "state": "<generated-state>",
  "authorization_url": "http://127.0.0.1:3002/oidc/auth?response_type=code&client_id=...",
  "expires_at": "2026-04-12T12:00:00Z"
}
```

Open the `authorization_url` in a browser to complete the login flow with Logto.

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `BRAIN_OIDC_ISSUER_URL` | OIDC provider issuer URL | (required for OIDC mode) |
| `BRAIN_OIDC_CLIENT_ID` | OIDC application client ID | (required for OIDC mode) |
| `BRAIN_OIDC_CLIENT_SECRET` | OIDC application client secret | (required for OIDC mode) |
| `BRAIN_OIDC_REDIRECT_URL` | Callback URL after OIDC login | `http://127.0.0.1:9090/api/auth/oidc/callback` |
| `BRAIN_OIDC_PROVIDER` | Provider name | `logto` |
| `BRAIN_OIDC_ENABLED` | Enable OIDC authentication | `false` |
| `BRAIN_OIDC_SCOPES` | Requested OIDC scopes (comma-separated) | `openid,profile,email` |
| `BRAIN_OIDC_LOGIN_HINT` | Pre-fill login hint | (empty) |
| `BRAIN_OIDC_PROMPT` | OIDC prompt parameter | (empty) |
| `BRAIN_OIDC_TRANSACTION_TTL_MINUTES` | Login transaction TTL | `10` |
| `BRAIN_AUTH_MODE` | Auth mode (`bootstrap` or `oidc`) | `bootstrap` |

## Docker Compose Configuration

The Logto docker-compose file at `docker-compose.logto.yml` includes:

- **PostgreSQL 15** with persistent volume
- **Logto** with admin seeding on first boot
- Health checks for both services
- Minimal resource usage (development mode)

### Customizing Logto

Override defaults via environment variables:

```bash
LOGTO_PG_PASSWORD=my-secret \
LOGTO_ADMIN_EMAIL=admin@mycompany.com \
LOGTO_ADMIN_PASSWORD=MyAdmin123! \
LOGTO_ENDPOINT=http://logto.mycompany.com:3002 \
LOGTO_ADMIN_ENDPOINT=http://logto.mycompany.com:3001 \
docker compose -f docker-compose.logto.yml up -d
```

## Testing

### Run All OIDC Tests

```bash
# Full E2E (starts Logto, runs tests, cleans up)
./scripts/run-oidc-integration-tests.sh

# Unit tests only (mock provider, no Logto needed)
cd apps/daemon
go test -v -run TestOIDC -timeout 60s ./cmd/braind/
```

### Skip OIDC Tests in CI

```bash
SKIP_OIDC_TESTS=true go test -v ./...
```

### Test Categories

1. **PKCE Tests** - Verify code verifier/challenge generation
2. **SSOManager Tests** - Test `StartLogin`, `CompleteLogin`, session management
3. **Handler Tests** - Test HTTP endpoint responses (mock provider)
4. **Integration Tests** - Test against a real Logto instance (skippable)

### Integration Test Requirements

To run integration tests against a real Logto instance, set:

```bash
export BRAIN_OIDC_ISSUER_URL="http://127.0.0.1:3002/oidc"
export BRAIN_OIDC_CLIENT_ID="<your-client-id>"
export BRAIN_OIDC_CLIENT_SECRET="<your-client-secret>"
export SKIP_OIDC_TESTS=false
```

Then run:

```bash
cd apps/daemon
go test -v -run TestOIDCIntegration -timeout 120s ./cmd/braind/
```

## Logto Admin Console

Access the admin console at [http://127.0.0.1:3001](http://127.0.0.1:3001).

**Default admin credentials:**
- Email: `admin@brain.local`
- Password: `Admin123456!`

Use the admin console to:
- View and manage OIDC applications
- Create and manage test users
- View audit logs
- Configure sign-in experiences

## Cleanup

### Remove Logto

```bash
# Stop and remove containers (keeps data volumes)
docker compose -f docker-compose.logto.yml down

# Stop and remove containers + data volumes (full reset)
docker compose -f docker-compose.logto.yml down -v

# Or use the integration test script's cleanup
./scripts/run-oidc-integration-tests.sh --clean
```

### Remove OIDC Environment Variables

```bash
unset BRAIN_OIDC_ISSUER_URL
unset BRAIN_OIDC_CLIENT_ID
unset BRAIN_OIDC_CLIENT_SECRET
unset BRAIN_OIDC_REDIRECT_URL
unset BRAIN_OIDC_PROVIDER
unset BRAIN_OIDC_ENABLED
```

## Troubleshooting

### Logto not starting

```bash
# Check container logs
docker compose -f docker-compose.logto.yml logs logto
docker compose -f docker-compose.logto.yml logs postgres
```

### "connection refused" errors

- Ensure Docker is running
- Check ports are not in use: `lsof -i :3001 -i :3002 -i :5432`
- Wait longer for initialization (Logto can take 30-60s)

### Setup script fails to create application

- Verify Logto is healthy: `curl http://127.0.0.1:3001/api/status`
- Check admin credentials match what was used during seeding
- The script is idempotent - running it again will skip existing resources

### Tests fail with "provider unavailable"

- Ensure `BRAIN_OIDC_ISSUER_URL` points to a running Logto instance
- Verify network connectivity: `curl $BRAIN_OIDC_ISSUER_URL/.well-known/openid-configuration`

### OIDC callback returns 503

- The daemon must be started with OIDC env vars set
- Check `BRAIN_OIDC_ENABLED=true` is set
- Verify the daemon logs show OIDC initialization

## File Reference

| File | Purpose |
|---|---|
| `docker-compose.logto.yml` | Logto + PostgreSQL Docker setup |
| `scripts/setup-logto.sh` | Initializes Logto app and test user |
| `scripts/run-oidc-integration-tests.sh` | Full E2E test orchestration |
| `apps/daemon/cmd/braind/oidc_integration_test.go` | Go OIDC integration tests |
| `core/enterprise/sso.go` | OIDC client implementation (do not modify) |
| `apps/daemon/cmd/braind/oidc.go` | OIDC HTTP handlers (do not modify) |
