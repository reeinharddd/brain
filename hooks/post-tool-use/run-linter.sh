#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/hooks/post-tool-use/run-linter.sh
#  Post-tool-use hook for Claude Code.
#  Runs the appropriate linter after a file is written.
#  Does not block — only notifies of issues.
#
#  Environment variable TOOL_OUTPUT_PATH contains the written file path.
# ═══════════════════════════════════════════════════════════

FILE="${TOOL_OUTPUT_PATH:-}"

# If no file path provided, exit silently
[ -z "$FILE" ] && exit 0
[ ! -f "$FILE" ] && exit 0

# Get file extension
EXT="${FILE##*.}"

# ── Run linter based on file type ────────────────────────────

run_eslint() {
  if command -v eslint &>/dev/null; then
    eslint --no-eslintrc --rule '{"no-unused-vars": "warn"}' "$FILE" 2>/dev/null \
      && echo "  ✓ ESLint: no issues" \
      || echo "  ⚠ ESLint: issues found in $FILE — run: eslint $FILE"
  fi
}

run_biome() {
  if command -v biome &>/dev/null; then
    biome lint "$FILE" 2>/dev/null \
      && echo "  ✓ Biome: no issues" \
      || echo "  ⚠ Biome: issues in $FILE"
  fi
}

run_ruff() {
  if command -v ruff &>/dev/null; then
    ruff check "$FILE" 2>/dev/null \
      && echo "  ✓ Ruff: no issues" \
      || echo "  ⚠ Ruff: issues found in $FILE — run: ruff check $FILE"
  fi
}

run_shellcheck() {
  if command -v shellcheck &>/dev/null; then
    shellcheck "$FILE" 2>/dev/null \
      && echo "  ✓ ShellCheck: no issues" \
      || echo "  ⚠ ShellCheck: issues found in $FILE — run: shellcheck $FILE"
  fi
}

case "$EXT" in
  js|jsx|mjs|cjs)
    # Prefer Biome if available, fall back to ESLint
    command -v biome &>/dev/null && run_biome || run_eslint
    ;;
  ts|tsx)
    command -v biome &>/dev/null && run_biome || run_eslint
    ;;
  py)
    run_ruff
    ;;
  sh|bash)
    run_shellcheck
    ;;
  *)
    # Unknown file type — skip silently
    ;;
esac

# Always exit 0 (post-hook should not block)
exit 0
