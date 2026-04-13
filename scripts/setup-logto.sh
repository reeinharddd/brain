#!/usr/bin/env bash
# setup-logto.sh
# Initializes Logto with a test OIDC application and test user via direct DB access.
# Idempotent: safe to run multiple times.
#
# Usage:
#   docker compose -f docker-compose.logto.yml up -d
#   ./scripts/setup-logto.sh
#
# After setup:
#   source ./scripts/oidc-env.sh
#   cd apps/daemon && BRAIN_AUTH_MODE=oidc CGO_ENABLED=0 go run ./cmd/braind

set -euo pipefail

LOGTO_ADMIN_URL="${LOGTO_ADMIN_URL:-http://127.0.0.1:3001}"
LOGTO_ISSUER_URL="${LOGTO_ISSUER_URL:-http://127.0.0.1:3002/oidc}"
LOGTO_CONTAINER="${LOGTO_CONTAINER:-brain-logto-postgres}"
LOGTO_DB_USER="${LOGTO_DB_USER:-logto}"
LOGTO_DB_NAME="${LOGTO_DB_NAME:-logto}"

APP_ID="brain_daemon"
APP_NAME="brain-daemon"
APP_REDIRECT_URI="http://127.0.0.1:9090/api/auth/oidc/callback"
APP_POST_LOGOUT_URI="http://127.0.0.1:9090/"

TEST_USER_ID="test_brain"
TEST_EMAIL="test@brain.local"
TEST_PASSWORD="Test123456!"
TEST_NAME="Brain Test User"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${CYAN}[setup-logto]${NC} $*"; }
ok()    { echo -e "${GREEN}[setup-logto]${NC} $*"; }
warn()  { echo -e "${YELLOW}[setup-logto]${NC} $*"; }
error() { echo -e "${RED}[setup-logto]${NC} $*"; }

# Check Logto is running
check_logto() {
    if ! docker ps --format '{{.Names}}' | grep -q "brain-logto"; then
        error "Logto container not running. Start with: docker compose -f docker-compose.logto.yml up -d"
        exit 1
    fi
    ok "Logto container found"
}

# Generate random secret
gen_secret() {
    openssl rand -hex 32
}

# Generate Argon2i password hash
gen_password_hash() {
    python3 -c "
from argon2 import PasswordHasher, Type
ph = PasswordHasher(type=Type.I)
print(ph.hash('$1'))
" 2>/dev/null
}

main() {
    info "=== Logto Setup Script ==="
    info "Issuer URL:   ${LOGTO_ISSUER_URL}"
    info "App ID:       ${APP_ID}"
    info "Redirect URI: ${APP_REDIRECT_URI}"
    echo ""

    check_logto

    # Generate app secret
    APP_SECRET=$(gen_secret)

    # Create OIDC application
    info "Creating/updating OIDC application: ${APP_NAME}..."
    docker exec ${LOGTO_CONTAINER} psql -U ${LOGTO_DB_USER} -d ${LOGTO_DB_NAME} -c "
INSERT INTO applications (
    tenant_id, id, name, secret, type, description,
    oidc_client_metadata, custom_client_metadata,
    custom_data, is_third_party
) VALUES (
    'default',
    '${APP_ID}',
    '${APP_NAME}',
    '${APP_SECRET}',
    'Traditional',
    'Brain Daemon OIDC Client',
    '{\"redirectUris\": [\"${APP_REDIRECT_URI}\"], \"postLogoutRedirectUris\": [\"${APP_POST_LOGOUT_URI}\"]}',
    '{}',
    '{}',
    false
) ON CONFLICT (id) DO UPDATE SET
    oidc_client_metadata = EXCLUDED.oidc_client_metadata,
    secret = EXCLUDED.secret;
" > /dev/null 2>&1

    ok "OIDC application ready"

    # Generate password hash
    USER_HASH=$(gen_password_hash)
    if [ -z "$USER_HASH" ]; then
        # Fallback: use a known-good Argon2i hash for Test123456!
        USER_HASH='$argon2i$v=19$m=65536,t=3,p=4$EQHyinRN2C1LjyBTjlN0Ug$EThh7amfT/HVtX5YBCExVRToilWssxSPx4RgUxxdVno'
        warn "Using fallback password hash (python argon2-cffi not available)"
    fi

    # Create test user
    info "Creating test user: ${TEST_EMAIL}..."
    docker exec ${LOGTO_CONTAINER} psql -U ${LOGTO_DB_USER} -d ${LOGTO_DB_NAME} -c "
INSERT INTO users (
    tenant_id, id, username, primary_email, name,
    password_encrypted, password_encryption_method,
    profile, identities, custom_data, logto_config, mfa_verifications, is_suspended
) VALUES (
    'default',
    '${TEST_USER_ID}',
    '${TEST_EMAIL}',
    '${TEST_EMAIL}',
    '${TEST_NAME}',
    '${USER_HASH}',
    'Argon2i',
    '{}',
    '{}',
    '{}',
    '{}',
    '[]',
    false
) ON CONFLICT (id) DO NOTHING;
" > /dev/null 2>&1

    ok "Test user ready: ${TEST_EMAIL} / ${TEST_PASSWORD}"

    # Verify OIDC discovery
    info "Verifying OIDC discovery..."
    if curl -sS "${LOGTO_ISSUER_URL}/.well-known/openid-configuration" | grep -q '"issuer"'; then
        ok "OIDC discovery working"
    else
        error "OIDC discovery failed"
        exit 1
    fi

    echo ""
    ok "=== Setup Complete ==="
    echo ""
    echo -e "${GREEN}To start daemon with OIDC:${NC}"
    echo -e "  source ./scripts/oidc-env.sh"
    echo -e "  cd apps/daemon && CGO_ENABLED=0 go run ./cmd/braind"
    echo ""
    echo -e "${GREEN}Environment Variables:${NC}"
    echo -e "  BRAIN_OIDC_ISSUER_URL=${LOGTO_ISSUER_URL}"
    echo -e "  BRAIN_OIDC_CLIENT_ID=${APP_ID}"
    echo -e "  BRAIN_OIDC_CLIENT_SECRET=${APP_SECRET}"
    echo -e "  BRAIN_OIDC_REDIRECT_URL=${APP_REDIRECT_URI}"
    echo -e "  BRAIN_OIDC_PROVIDER=logto"
    echo -e "  BRAIN_OIDC_ENABLED=true"
    echo -e "  BRAIN_AUTH_MODE=oidc"
    echo -e "  BRAIN_AUTH_REQUIRED=true"
    echo ""
    echo -e "${GREEN}Test User:${NC}"
    echo -e "  Email:    ${TEST_EMAIL}"
    echo -e "  Password: ${TEST_PASSWORD}"
    echo ""
    echo -e "${GREEN}Admin Console:${NC}"
    echo -e "  URL:      ${LOGTO_ADMIN_URL}"
    echo -e "  Email:    admin@brain.local"
    echo -e "  Password: Admin123456!"
    echo ""
}

main "$@"
