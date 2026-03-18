#!/bin/bash

set -euo pipefail

. "${GUARDIAN_LIB}"

while IFS= read -r file; do
  [ -n "$file" ] || continue

  case "$file" in
    .env|.env.*|*/.env|*/.env.*)
      guardian_report "critical" "tracked-env-file" "$file" "Tracked environment file detected in diff"
      ;;
  esac
done < "${GUARDIAN_FILES_FILE}"
