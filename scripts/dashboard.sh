#!/bin/bash
# Lightweight terminal dashboard for brain telemetry and health.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
TELEMETRY_SCRIPT="$BRAIN_DIR/scripts/telemetry.sh"

echo "Brain Dashboard"
echo ""
echo "Health snapshot:"
bash "$BRAIN_DIR/scripts/doctor.sh" --quick || true
echo ""
echo "Telemetry summary:"
bash "$TELEMETRY_SCRIPT" summary
