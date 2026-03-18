#!/bin/bash
# Derive a stable namespace for project-scoped memory.

set -euo pipefail

PROJECT_ROOT="${1:-$PWD}"
START_TS="$(date +%s%3N 2>/dev/null || date +%s000)"

if [ ! -d "$PROJECT_ROOT" ]; then
  echo "ERROR: project root not found: $PROJECT_ROOT" >&2
  exit 1
fi

BASENAME="$(basename "$PROJECT_ROOT")"

if git -C "$PROJECT_ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  TOPLEVEL="$(git -C "$PROJECT_ROOT" rev-parse --show-toplevel)"
  REMOTE_URL="$(git -C "$TOPLEVEL" remote get-url origin 2>/dev/null || true)"
  if [ -n "$REMOTE_URL" ]; then
    REMOTE_SLUG="$(printf '%s' "$REMOTE_URL" | sed -E 's#.*[:/]([^/]+/[^/.]+)(\.git)?$#\1#')"
    NAMESPACE="$(printf '%s\n' "$REMOTE_SLUG" | tr '/:' '__')"
    printf '%s\n' "$NAMESPACE"
    if [ -x "${HOME}/.brain/scripts/telemetry.sh" ]; then
      END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
      DURATION_MS="$((END_TS - START_TS))"
      bash "${HOME}/.brain/scripts/telemetry.sh" record "memory-namespace" "ok" "$DURATION_MS" "$NAMESPACE" >/dev/null 2>&1 || true
    fi
    exit 0
  fi
  BASENAME="$(basename "$TOPLEVEL")"
fi

NAMESPACE="local__${BASENAME}"
printf '%s\n' "$NAMESPACE"
if [ -x "${HOME}/.brain/scripts/telemetry.sh" ]; then
  END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
  DURATION_MS="$((END_TS - START_TS))"
  bash "${HOME}/.brain/scripts/telemetry.sh" record "memory-namespace" "ok" "$DURATION_MS" "$NAMESPACE" >/dev/null 2>&1 || true
fi
