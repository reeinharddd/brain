#!/bin/bash

set -euo pipefail

RESULTS_FILE="$1"
FAIL_COUNT="$2"

if [ ! -s "$RESULTS_FILE" ]; then
  echo "Guardian verdict: PASS"
  exit 0
fi

echo "Guardian findings:"
while IFS=$'\t' read -r severity check_id file message; do
  printf -- '- [%s] %s :: %s :: %s\n' "$severity" "$check_id" "$file" "$message"
done < "$RESULTS_FILE"

if [ "$FAIL_COUNT" -gt 0 ]; then
  echo "Guardian verdict: BLOCK"
else
  echo "Guardian verdict: WARN"
fi
