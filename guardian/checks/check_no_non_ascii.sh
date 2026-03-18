#!/bin/bash

set -euo pipefail

. "${GUARDIAN_LIB}"

while IFS= read -r file; do
  [ -n "$file" ] || continue
  guardian_is_source_file "$file" || continue
  case "$file" in
    *.md|*.json)
      continue
      ;;
  esac

  if guardian_added_lines "$file" | LC_ALL=C grep -P '[^\x00-\x7F]' >/dev/null; then
    guardian_report "medium" "plain-text-only" "$file" "Non-ASCII characters detected in added lines"
  fi
done < "${GUARDIAN_FILES_FILE}"
