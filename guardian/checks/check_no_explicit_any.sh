#!/bin/bash

set -euo pipefail

. "${GUARDIAN_LIB}"

while IFS= read -r file; do
  [ -n "$file" ] || continue

  case "$file" in
    *.ts|*.tsx)
      if guardian_added_lines "$file" | grep -Eq '(^|\W)any(\W|$)'; then
        guardian_report "critical" "no-explicit-any" "$file" "Explicit any detected in added lines"
      fi
      ;;
  esac
done < "${GUARDIAN_FILES_FILE}"
