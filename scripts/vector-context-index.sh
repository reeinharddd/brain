#!/bin/bash
# Build a lightweight codebase context index for later semantic ingestion.

set -euo pipefail

PROJECT_ROOT="${1:-$PWD}"
OUTPUT_PATH="${2:-$PROJECT_ROOT/.brain/codebase-context.ndjson}"

mkdir -p "$(dirname "$OUTPUT_PATH")"
> "$OUTPUT_PATH"

find "$PROJECT_ROOT" \
  -type f \
  \( -name 'README.md' -o -path '*/docs/*' -o -name 'CLAUDE.md' -o -name 'GEMINI.md' -o -name '*.md' \) \
  ! -path '*/.git/*' \
  ! -path "$PROJECT_ROOT/.brain/*" \
  ! -path '*/node_modules/*' \
  | sort \
  | while IFS= read -r file; do
      python3 - "$file" "$OUTPUT_PATH" <<'PY'
import hashlib
import json
import pathlib
import sys

file_path = pathlib.Path(sys.argv[1])
output_path = pathlib.Path(sys.argv[2])
content = file_path.read_text(encoding="utf-8", errors="ignore")
payload = {
    "id": hashlib.sha1(str(file_path).encode()).hexdigest(),
    "path": str(file_path),
    "title": file_path.name,
    "content": content[:4000],
}
with output_path.open("a", encoding="utf-8") as fh:
    fh.write(json.dumps(payload, ensure_ascii=True) + "\n")
PY
    done

echo "$OUTPUT_PATH"
