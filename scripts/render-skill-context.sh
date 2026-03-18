#!/bin/bash
# Render the relevant skill context bundle for the current project.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
PROJECT_ROOT="${1:-$PWD}"
REGISTRY="$BRAIN_DIR/skills/dynamic-registry.tsv"
DETECT_SCRIPT="$BRAIN_DIR/scripts/detect-stack.sh"
WRITE_MODE=0
START_TS="$(date +%s%3N 2>/dev/null || date +%s000)"

if [ "${1:-}" = "--write" ]; then
  WRITE_MODE=1
  PROJECT_ROOT="${2:-$PWD}"
fi

if [ ! -x "$DETECT_SCRIPT" ]; then
  echo "ERROR: missing detect-stack script: $DETECT_SCRIPT" >&2
  exit 1
fi

if [ ! -f "$REGISTRY" ]; then
  echo "ERROR: missing registry: $REGISTRY" >&2
  exit 1
fi

readarray -t DETECTED_TAGS < <("$DETECT_SCRIPT" "$PROJECT_ROOT")

has_detected_tag() {
  local wanted="$1"
  local tag
  for tag in "${DETECTED_TAGS[@]:-}"; do
    if [ "$tag" = "$wanted" ]; then
      return 0
    fi
  done
  return 1
}

render_bundle() {
  local skill_id title detect_tags context_path summary match tag old_ifs

  echo "# Project Skill Context"
  echo
  echo "- Project root: $PROJECT_ROOT"
  if [ "${#DETECTED_TAGS[@]}" -gt 0 ]; then
    echo "- Detected tags: ${DETECTED_TAGS[*]}"
  else
    echo "- Detected tags: none"
  fi
  echo
  echo "Load only the contexts below in addition to the global brain rules."

  while IFS=$'\t' read -r skill_id title detect_tags context_path summary; do
    [ -n "$skill_id" ] || continue
    case "$skill_id" in
      \#*) continue ;;
    esac

    match=0
    old_ifs="$IFS"
    IFS=','
    for tag in $detect_tags; do
      if has_detected_tag "$tag"; then
        match=1
        break
      fi
    done
    IFS="$old_ifs"

    if [ "$match" -eq 1 ]; then
      echo
      echo "## $title"
      echo
      echo "- Skill ID: $skill_id"
      echo "- Match tags: $detect_tags"
      echo "- Summary: $summary"
      echo
      cat "$BRAIN_DIR/$context_path"
    fi
  done < "$REGISTRY"
}

if [ "$WRITE_MODE" -eq 1 ]; then
  mkdir -p "$PROJECT_ROOT/.brain"
  printf '%s\n' "${DETECTED_TAGS[@]:-}" > "$PROJECT_ROOT/.brain/stack-tags.txt"
  render_bundle > "$PROJECT_ROOT/.brain/skill-context.md"
  if [ -x "${HOME}/.brain/scripts/telemetry.sh" ]; then
    END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
    DURATION_MS="$((END_TS - START_TS))"
    bash "${HOME}/.brain/scripts/telemetry.sh" record "render-skill-context" "ok" "$DURATION_MS" "$PROJECT_ROOT" >/dev/null 2>&1 || true
  fi
  echo "$PROJECT_ROOT/.brain/skill-context.md"
else
  render_bundle
  if [ -x "${HOME}/.brain/scripts/telemetry.sh" ]; then
    END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
    DURATION_MS="$((END_TS - START_TS))"
    bash "${HOME}/.brain/scripts/telemetry.sh" record "render-skill-context" "ok" "$DURATION_MS" "$PROJECT_ROOT" >/dev/null 2>&1 || true
  fi
fi
