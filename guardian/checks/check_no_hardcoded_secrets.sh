#!/bin/bash

set -euo pipefail

. "${GUARDIAN_LIB}"

while IFS= read -r file; do
  [ -n "$file" ] || continue
  guardian_is_source_file "$file" || continue

  if guardian_added_lines "$file" | grep -Eiq '(api[_-]?key|secret|token|password|auth[_-]?key).{0,40}["'"'"'][A-Za-z0-9_/\+=-]{12,}["'"'"']'; then
    guardian_report "critical" "hardcoded-secrets" "$file" "Potential hardcoded secret found in added lines"
  fi
done < "${GUARDIAN_FILES_FILE}"
