#!/usr/bin/env bash
# Source this file to get OIDC env vars for the daemon
# Usage: source scripts/oidc-env.sh

set -euo pipefail

# OIDC Configuration
export BRAIN_AUTH_REQUIRED="${BRAIN_AUTH_REQUIRED:-true}"
export BRAIN_AUTH_MODE="${BRAIN_AUTH_MODE:-oidc}"
export BRAIN_OIDC_ENABLED="${BRAIN_OIDC_ENABLED:-true}"
export BRAIN_OIDC_ISSUER_URL="${BRAIN_OIDC_ISSUER_URL:-http://127.0.0.1:3002/oidc}"
export BRAIN_OIDC_CLIENT_ID="${BRAIN_OIDC_CLIENT_ID:-brain_daemon}"
export BRAIN_OIDC_CLIENT_SECRET="${BRAIN_OIDC_CLIENT_SECRET:-a4a7ee1101f87747440ca83b7bd7dc63}"
export BRAIN_OIDC_REDIRECT_URL="${BRAIN_OIDC_REDIRECT_URL:-http://127.0.0.1:9090/api/auth/oidc/callback}"
export BRAIN_OIDC_PROVIDER="${BRAIN_OIDC_PROVIDER:-logto}"
export BRAIN_OIDC_SCOPES="${BRAIN_OIDC_SCOPES:-openid profile email}"

echo "OIDC env vars loaded for daemon:"
echo "  Issuer:       $BRAIN_OIDC_ISSUER_URL"
echo "  Client ID:    $BRAIN_OIDC_CLIENT_ID"
echo "  Redirect:     $BRAIN_OIDC_REDIRECT_URL"
echo "  Provider:     $BRAIN_OIDC_PROVIDER"
echo ""
echo "Test user: test@brain.local / Test123456!"
echo "Admin console: http://127.0.0.1:3001 (admin@brain.local / Admin123456!)"
