#!/bin/bash

set -euo pipefail

. "${GUARDIAN_LIB}"

while IFS= read -r file; do
  [ -n "$file" ] || continue

  case "$file" in
    .env|.env.*|*/.env|*/.env.*)
      case "$file" in
        *.example) ;;
        *) guardian_report "critical" "tracked-env-file" "$file" "Tracked environment file detected in diff" ;;
      esac
      ;;
  esac
done < "${GUARDIAN_FILES_FILE}"
