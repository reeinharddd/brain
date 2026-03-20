#!/bin/bash
# Guardian Check: Detect hardcoded secrets in source files
# This check scans for common secret patterns that should never be committed

set -euo pipefail

. "${GUARDIAN_LIB}"

# Define secret patterns to detect
# Format: SEVERITY|PATTERN_NAME|REGEX_PATTERN
declare -a SECRET_PATTERNS=(
    # AWS Keys
    "CRITICAL|aws-access-key|AKIA[0-9A-Z]{16}"
    "CRITICAL|aws-secret-key|[A-Za-z0-9/+=]{40}"

    # GitHub Tokens
    "CRITICAL|github-token|(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}"
    "CRITICAL|github-classic-token|ghp_[0-9a-zA-Z]{36}"

    # Generic API Keys with values
    "CRITICAL|api-key-with-value|(?:api[_-]?key|apikey)\\s*[:=]\\s*[\"'][a-zA-Z0-9_\\-]{16,}[\"']"
    "CRITICAL|secret-with-value|(?:secret|private[_-]?key)\\s*[:=]\\s*[\"'][a-zA-Z0-9_\\-]{16,}[\"']"
    "CRITICAL|token-with-value|(?:token|access[_-]?token|auth[_-]?token)\\s*[:=]\\s*[\"'][a-zA-Z0-9_\\-]{16,}[\"']"
    "CRITICAL|password-with-value|(?:password|passwd|pwd)\\s*[:=]\\s*[\"'][^\"']{8,}[\"']"

    # Bearer tokens
    "CRITICAL|bearer-token|Bearer\\s+[a-zA-Z0-9_\\-\\.]{20,}"

    # JWT tokens (base64url encoded)
    "HIGH|jwt-token|eyJ[A-Za-z0-9_\\-]*\\.eyJ[A-Za-z0-9_\\-]*\\.[A-Za-z0-9_\\-]*"

    # Database connection strings with passwords
    "CRITICAL|db-connection-with-password|(?:mongodb|postgres|mysql)://[^:]*:[^@]*@"

    # Private keys
    "CRITICAL|private-key|-----BEGIN\\s+(?:RSA|EC|DSA|OPENSSH)\\s+PRIVATE\\s+KEY-----"

    # Certificate files
    "HIGH|certificate-file|-----BEGIN\\s+CERTIFICATE-----"

    # Slack tokens
    "CRITICAL|slack-token|xox[baprs]-[0-9]{10,13}-[0-9]{10,13}-[a-zA-Z0-9]{24,}"

    # Stripe keys
    "CRITICAL|stripe-key|sk_(live|test)_[0-9a-zA-Z]{24,}"
    "CRITICAL|stripe-publishable|pk_(live|test)_[0-9a-zA-Z]{24,}"

    # Generic high-entropy strings that look like secrets
    "HIGH|high-entropy-secret|(?:key|token|secret)\\s*[:=]\\s*[\"'][A-Za-z0-9/+=]{32,}[\"']"

    # Base64 encoded strings that might be secrets
    "MEDIUM|base64-secret|b[\"']?[A-Za-z0-9+/]{40,}[\"']?={0,2}"
)

# File patterns to skip (test files, examples, etc.)
declare -a SKIP_PATTERNS=(
    "*.test.*"
    "*.spec.*"
    "*example*"
    "*sample*"
    "*mock*"
    "*.md"
    "*.example"
    "*/test/*"
    "*/tests/*"
    "__snapshots__"
)

should_skip_file() {
    local file="$1"
    for pattern in "${SKIP_PATTERNS[@]}"; do
        if [[ "$file" == *${pattern}* ]]; then
            return 0
        fi
    done
    return 1
}

# Main check logic
found_secrets=0

while IFS= read -r file; do
    [ -n "$file" ] || continue
    guardian_is_source_file "$file" || continue
    should_skip_file "$file" && continue

    # Get added lines
    added_lines=$(guardian_added_lines "$file" 2>/dev/null || true)
    [ -n "$added_lines" ] || continue

    # Check each pattern
    for pattern_def in "${SECRET_PATTERNS[@]}"; do
        IFS='|' read -r severity pattern_name regex <<< "$pattern_def"

        if echo "$added_lines" | grep -Eiq "$regex" 2>/dev/null; then
            # Get the matching line for context
            match_line=$(echo "$added_lines" | grep -Ei "$regex" | head -1 || true)

            # Truncate match for display
            match_display="${match_line:0:80}"
            if [ ${#match_line} -gt 80 ]; then
                match_display="${match_display}..."
            fi

            guardian_report "$severity" "hardcoded-secret:$pattern_name" "$file" \
                "Found potential $pattern_name. Match: $match_display"
            found_secrets=$((found_secrets + 1))
        fi
    done
done < "${GUARDIAN_FILES_FILE}"

if [ $found_secrets -gt 0 ]; then
    exit 1
fi

exit 0
