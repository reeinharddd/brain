#!/usr/bin/env bash

# reeinharrrd's Brain Repo - Git Pre-commit Guardian
# Version: 1.0.0

echo "Running Pre-commit Guardian..."

# 1. No Secrets Check (Basic) - Looking for keys, tokens, etc.
# Note: In production, use a real tool like 'gitleaks'.
SECRETS=$(grep -rE "(token|api_key|password|secret|auth_key).*['\"][a-zA-Z0-9_-]{16,}['\"]" .)
if [[ -n "$SECRETS" ]]; then
  echo "ERROR: Potential secrets found in your commit!"
  echo "$SECRETS"
  exit 1
fi

# 2. No Emojis/Symbols Check (Rule: Plain Text Only)
# Searching for characters outside of ASCII 0-127
NON_ASCII=$(grep -rnP "[^\x00-\x7f]" .)
if [[ -n "$NON_ASCII" ]]; then
  echo "ERROR: Non-ASCII characters (emojis/symbols) found! Rule: Plain Text Only."
  echo "$NON_ASCII"
  exit 1
fi

# 3. Environment Health Check (doctor.sh)
if [[ -f "./scripts/doctor.sh" ]]; then
  ./scripts/doctor.sh --quick
  if [[ $? -ne 0 ]]; then
    echo "ERROR: Environment health check failed. Run ./scripts/doctor.sh for details."
    exit 1
  fi
fi

# 4. Generate Adapters if rules changed
if git diff --cached --name-only | grep -q "^rules/"; then
  echo "Rules changed. Re-generating adapters..."
  ./adapters/generate.sh
  git add adapters/
fi

echo "Guardian Passed."
exit 0
