#!/bin/bash
# check_no_non_ascii.sh
# Checks for non-ASCII characters in source files.
# Intentionally SKIPS Python/JS/TS files because Unicode string literals
# are valid and common in those languages (e.g. u"\u2713" in Python).
# Only flags shell scripts and markdown where non-ASCII is almost always
# a mistake (stray emoji, copy-paste corruption, encoding error).

set -euo pipefail

. "${GUARDIAN_LIB}"

# File types where non-ASCII is a legitimate mistake
STRICT_EXTENSIONS=("sh" "bash" "zsh" "fish")

while IFS= read -r file; do
  [ -n "$file" ] || continue
  guardian_is_source_file "$file" || continue

  # Get file extension
  ext="${file##*.}"

  # Skip file types where Unicode is intentional
  case "$ext" in
    py|js|ts|tsx|jsx|go|rs|rb|java|kt|swift|md|json|yml|yaml|toml)
      continue
      ;;
  esac

  # For shell scripts: flag non-ASCII as likely a mistake
  strict=0
  for strict_ext in "${STRICT_EXTENSIONS[@]}"; do
    [ "$ext" = "$strict_ext" ] && strict=1 && break
  done
  # Also check files with no extension that look like shell (shebang)
  if [ "$strict" -eq 0 ]; then
    first_line="$(head -1 "$file" 2>/dev/null || true)"
    case "$first_line" in
      "#!/bin/bash"*|"#!/bin/sh"*|"#!/usr/bin/env bash"*) strict=1 ;;
    esac
  fi

  [ "$strict" -eq 1 ] || continue

  if guardian_added_lines "$file" | LC_ALL=C grep -P '[^\x00-\x7F]' >/dev/null; then
    guardian_report "medium" "plain-text-only" "$file" "Non-ASCII characters in shell script - use ASCII equivalents"
  fi
done < "${GUARDIAN_FILES_FILE}"
