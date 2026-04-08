#!/usr/bin/env bash

set -euo pipefail

ROOT="${1:-.}"

cd "$ROOT"

echo "[markdown-boundary] validating markdown placement"

invalid_files=()

while IFS= read -r path; do
  case "$path" in
    docs/*)
      ;;
    artifacts/*)
      ;;
    README.md)
      ;;
    utils/docs/README.md)
      ;;
    *)
      invalid_files+=("$path")
      ;;
  esac
done < <(rg --files -g '*.md' -g '!docs/archive/**')

legacy_dirs=(agents commands rules adapters mcp guardian hooks)

legacy_regular_files=()
for dir in "${legacy_dirs[@]}"; do
  if [ -d "$dir" ]; then
    while IFS= read -r file; do
      if [ ! -L "$file" ]; then
        legacy_regular_files+=("$file")
      fi
    done < <(find "$dir" -type f -name '*.md' | sort)
  fi
done

if [ "${#invalid_files[@]}" -gt 0 ]; then
  echo
  echo "[markdown-boundary] markdown files found outside allowed locations:"
  printf '  - %s\n' "${invalid_files[@]}"
fi

if [ "${#legacy_regular_files[@]}" -gt 0 ]; then
  echo
  echo "[markdown-boundary] legacy directories still contain real markdown files."
  echo "[markdown-boundary] only compatibility symlinks are allowed in legacy locations:"
  printf '  - %s\n' "${legacy_regular_files[@]}"
fi

if [ "${#invalid_files[@]}" -gt 0 ] || [ "${#legacy_regular_files[@]}" -gt 0 ]; then
  echo
  echo "[markdown-boundary] failed"
  exit 1
fi

echo "[markdown-boundary] ok"
