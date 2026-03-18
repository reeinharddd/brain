#!/bin/bash
# Build a deterministic rules bundle from canonical.md + modules/.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
SOURCE="$BRAIN_DIR/rules/canonical.md"
MODULES_DIR="$BRAIN_DIR/rules/modules"
COMPILED_DIR="$BRAIN_DIR/rules/compiled"

usage() {
  cat <<'EOF'
Usage:
  bash ~/.brain/scripts/build-rules.sh
  bash ~/.brain/scripts/build-rules.sh --stdout
EOF
}

collect_modules() {
  if [ ! -d "$MODULES_DIR" ]; then
    return
  fi

  find "$MODULES_DIR" -maxdepth 1 -type f -name '*.md' | sort
}

build_rules() {
  if [ ! -f "$SOURCE" ]; then
    echo "ERROR: missing source file: $SOURCE" >&2
    exit 1
  fi

  cat "$SOURCE"

  while IFS= read -r module; do
    printf '\n\n'
    cat "$module"
  done < <(collect_modules)
}

write_artifacts() {
  local full_rules="$1"

  mkdir -p "$COMPILED_DIR"
  printf '%s\n' "$full_rules" > "$COMPILED_DIR/full.md"

  {
    echo "# Compiled Rules Manifest"
    echo
    echo "- Source: $SOURCE"
    echo "- Modules:"
    while IFS= read -r module; do
      echo "  - $module"
    done < <(collect_modules)
  } > "$COMPILED_DIR/manifest.md"
}

case "${1:-}" in
  --stdout)
    build_rules
    ;;
  "")
    FULL_RULES="$(build_rules)"
    write_artifacts "$FULL_RULES"
    printf '%s\n' "$FULL_RULES"
    ;;
  -h|--help)
    usage
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
