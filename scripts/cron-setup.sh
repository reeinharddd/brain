#!/bin/bash
# cron-setup.sh - Install recurring maintenance tasks for the brain repo.
#
# Tasks installed:
#   Daily   (02:00) - schema validation + adapter consistency check
#   Weekly  (Sun 03:00) - memory consolidation + benchmark run
#   Monthly (1st 04:00) - vector index rebuild from latest codebase context
#
# Usage:
#   bash cron-setup.sh [--remove] [--dry-run] [--list]

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
PYTHON3="$(command -v python3)"
BASH="$(command -v bash)"
LOG_DIR="$BRAIN_DIR/logs/cron"
mkdir -p "$LOG_DIR"

DRY_RUN=0
REMOVE=0
LIST=0

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --remove)  REMOVE=1  ;;
    --list)    LIST=1    ;;
  esac
done

# Define cron jobs
DAILY_VALIDATE="0 2 * * * $BASH $BRAIN_DIR/scripts/validate-schema.py --json >> $LOG_DIR/validate-\$(date +\%Y\%m\%d).json 2>&1"
DAILY_DOCTOR="30 2 * * * $BASH $BRAIN_DIR/scripts/doctor.sh --json >> $LOG_DIR/doctor-\$(date +\%Y\%m\%d).json 2>&1"
WEEKLY_CONSOLIDATE="0 3 * * 0 $BASH $BRAIN_DIR/scripts/consolidate-memory.sh --json >> $LOG_DIR/consolidate-\$(date +\%Y\%m\%d).json 2>&1"
WEEKLY_EVALS="30 3 * * 0 $BASH $BRAIN_DIR/evals/run.sh --json >> $LOG_DIR/evals-\$(date +\%Y\%m\%d).json 2>&1"
MONTHLY_VECTOR="0 4 1 * * $BASH $BRAIN_DIR/scripts/vector-context-index.sh $BRAIN_DIR && $BASH $BRAIN_DIR/scripts/vector-sync-qdrant.sh $BRAIN_DIR/.brain/codebase-context.ndjson >> $LOG_DIR/vector-\$(date +\%Y\%m\%d).json 2>&1"

BRAIN_MARKER="# brain-repo-managed"

JOBS=(
  "$DAILY_VALIDATE    $BRAIN_MARKER daily-validate"
  "$DAILY_DOCTOR      $BRAIN_MARKER daily-doctor"
  "$WEEKLY_CONSOLIDATE $BRAIN_MARKER weekly-consolidate"
  "$WEEKLY_EVALS       $BRAIN_MARKER weekly-evals"
  "$MONTHLY_VECTOR     $BRAIN_MARKER monthly-vector"
)

if [ "$LIST" -eq 1 ]; then
  echo "Current brain cron jobs:"
  crontab -l 2>/dev/null | grep "$BRAIN_MARKER" || echo "  (none installed)"
  exit 0
fi

if [ "$REMOVE" -eq 1 ]; then
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "[dry-run] Would remove all brain cron jobs"
    crontab -l 2>/dev/null | grep "$BRAIN_MARKER" || true
    exit 0
  fi
  CURRENT="$(crontab -l 2>/dev/null | grep -v "$BRAIN_MARKER" || true)"
  echo "$CURRENT" | crontab -
  echo "[cron] All brain cron jobs removed"
  exit 0
fi

# Install
EXISTING="$(crontab -l 2>/dev/null | grep -v "$BRAIN_MARKER" || true)"
NEW_BLOCK="$(printf '%s\n' "${JOBS[@]}")"

if [ "$DRY_RUN" -eq 1 ]; then
  echo "[dry-run] Would install these cron jobs:"
  printf '%s\n' "${JOBS[@]}"
  exit 0
fi

{
  echo "$EXISTING"
  echo ""
  echo "# === brain-repo maintenance (installed by cron-setup.sh) ==="
  printf '%s\n' "${JOBS[@]}"
} | crontab -

echo "[cron] Brain maintenance tasks installed:"
echo "  Daily  02:00 - schema validation"
echo "  Daily  02:30 - doctor check"
echo "  Weekly Sun 03:00 - memory consolidation"
echo "  Weekly Sun 03:30 - eval suite"
echo "  Monthly 1st 04:00 - vector index rebuild"
echo ""
echo "  Logs: $LOG_DIR/"
echo ""
echo "  To remove: bash $BRAIN_DIR/scripts/cron-setup.sh --remove"
echo "  To list:   bash $BRAIN_DIR/scripts/cron-setup.sh --list"
