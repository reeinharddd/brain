#!/bin/bash
# Quick auth verification script
# Usage: ./scripts/verify-auth.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Brain Auth Verification ===${NC}\n"

# Test 1: Build verification
echo -e "1️⃣  Build verification..."
cd "$(dirname "$0")/.."

echo -n "   Core identity tests... "
if CGO_ENABLED=0 go test ./core/identity -count=1 >/dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

echo -n "   Daemon build... "
if CGO_ENABLED=0 go build ./apps/daemon/cmd/braind >/dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

echo -n "   CLI build... "
if CGO_ENABLED=0 go build ./apps/cli/cmd/brain >/dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

echo -n "   Desktop build... "
if (cd apps/desktop && npm run build >/dev/null 2>&1); then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

# Test 2: Check auth endpoints exist
echo -e "\n2️⃣  Auth endpoint presence check..."
ENDPOINTS=(
    "POST /api/auth/login"
    "POST /api/auth/logout"
    "POST /api/auth/refresh"
    "GET /api/auth/status"
    "GET /api/auth/oidc/start"
    "GET /api/auth/oidc/callback"
    "GET /api/auth/oidc/poll"
    "GET /api/users"
    "PATCH /api/users/{id}/role"
    "GET /api/invites"
    "POST /api/invites"
    "POST /api/invites/consume"
)

for endpoint in "${ENDPOINTS[@]}"; do
    path=$(echo "$endpoint" | awk '{print $2}' | sed 's/{id}/USER_ID/g')
    method=$(echo "$endpoint" | awk '{print $1}')
    
    # Check if the route exists in the code
    if grep -q "\"$path\"" apps/daemon/cmd/braind/main.go 2>/dev/null || \
       grep -q "$path" apps/daemon/cmd/braind/*.go 2>/dev/null; then
        echo -e "   ${GREEN}✓${NC} $endpoint"
    else
        echo -e "   ${YELLOW}○${NC} $endpoint (may be handled elsewhere)"
    fi
done

# Test 3: Check secure storage dependencies
echo -e "\n3️⃣  Secure storage setup..."
echo -n "   CLI keyring dependency... "
if grep -q "go-keyring" apps/cli/go.mod 2>/dev/null; then
    echo -e "${GREEN}✓ Found${NC}"
else
    echo -e "${RED}✗ Missing${NC}"
fi

echo -n "   Desktop Tauri keyring dependency... "
if grep -q "keyring" apps/desktop/src-tauri/Cargo.toml 2>/dev/null; then
    echo -e "${GREEN}✓ Found${NC}"
else
    echo -e "${YELLOW}○ Not configured (will use localStorage fallback)${NC}"
fi

# Test 4: Check OIDC dependencies
echo -e "\n4️⃣  OIDC setup..."
echo -n "   go-oidc dependency... "
if grep -q "go-oidc" core/go.mod 2>/dev/null; then
    echo -e "${GREEN}✓ Found${NC}"
else
    echo -e "${RED}✗ Missing${NC}"
fi

echo -n "   oauth2 dependency... "
if grep -q "oauth2" core/go.mod 2>/dev/null; then
    echo -e "${GREEN}✓ Found${NC}"
else
    echo -e "${RED}✗ Missing${NC}"
fi

echo -n "   SQLite dependency... "
if grep -q "sqlite" core/go.mod 2>/dev/null; then
    echo -e "${GREEN}✓ Found${NC}"
else
    echo -e "${RED}✗ Missing${NC}"
fi

# Test 5: Check env vars documentation
echo -e "\n5️⃣  Environment variables..."
ENV_VARS=(
    "BRAIN_AUTH_REQUIRED"
    "BRAIN_AUTH_MODE"
    "BRAIN_AUTH_BOOTSTRAP_EMAIL"
    "BRAIN_AUTH_BOOTSTRAP_PASSWORD"
    "BRAIN_AUTH_OIDC_ISSUER"
    "BRAIN_AUTH_OIDC_CLIENT_ID"
)

for var in "${ENV_VARS[@]}"; do
    if grep -q "$var" brain.env.example 2>/dev/null; then
        echo -e "   ${GREEN}✓${NC} $var"
    else
        echo -e "   ${YELLOW}○${NC} $var (not documented)"
    fi
done

# Test 6: Check test coverage
echo -e "\n6️⃣  Test files..."
TEST_FILES=(
    "core/identity/auth_test.go"
    "apps/daemon/cmd/braind/auth_test.go"
    "apps/desktop/src/api/auth.test.ts"
    "apps/desktop/src/App.test.tsx"
)

for test_file in "${TEST_FILES[@]}"; do
    if [ -f "$test_file" ]; then
        echo -e "   ${GREEN}✓${NC} $test_file"
    else
        echo -e "   ${RED}✗${NC} $test_file (missing)"
    fi
done

echo -e "\n${YELLOW}=== Verification Complete ===${NC}"
echo -e "\nFor manual testing, see: AUTH_IMPLEMENTATION_STATUS.md"
echo -e "Start daemon with: cd apps/daemon && CGO_ENABLED=0 go run ./cmd/braind"
