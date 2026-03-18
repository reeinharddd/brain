#!/bin/bash

set -euo pipefail

guardian_is_source_file() {
  case "$1" in
    *.js|*.jsx|*.mjs|*.cjs|*.ts|*.tsx|*.py|*.go|*.rs|*.java|*.kt|*.sh|*.bash|*.zsh|*.md|*.yml|*.yaml|*.json|*.toml)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

guardian_added_lines() {
  local file="$1"

  if [ "${GUARDIAN_MODE:-staged}" = "staged" ]; then
    git -C "$GUARDIAN_REPO_ROOT" diff --cached --unified=0 -- "$file" \
      | grep -E '^\+[^+]' || true
  else
    git -C "$GUARDIAN_REPO_ROOT" diff "${GUARDIAN_DIFF_RANGE:-HEAD}" --unified=0 -- "$file" \
      | grep -E '^\+[^+]' || true
  fi
}

guardian_report() {
  local severity="$1"
  local check_id="$2"
  local file="$3"
  local message="$4"
  printf '%s\t%s\t%s\t%s\n' "$severity" "$check_id" "$file" "$message"
}
