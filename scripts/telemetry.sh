#!/bin/bash
# Append and summarize lightweight telemetry events.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
LOG_DIR="$BRAIN_DIR/logs"
LOG_FILE="$LOG_DIR/telemetry.ndjson"

mkdir -p "$LOG_DIR"

record_event() {
  local event_name="$1"
  local status="${2:-ok}"
  local duration_ms="${3:-0}"
  local details="${4:-}"

  python3 - "$LOG_FILE" "$event_name" "$status" "$duration_ms" "$details" <<'PY'
import json
import sys
from datetime import datetime, timezone

path, event_name, status, duration_ms, details = sys.argv[1:6]
payload = {
    "ts": datetime.now(timezone.utc).isoformat(),
    "event": event_name,
    "status": status,
    "duration_ms": int(duration_ms or 0),
    "details": details,
}
with open(path, "a", encoding="utf-8") as fh:
    fh.write(json.dumps(payload, ensure_ascii=True) + "\n")
PY
}

print_summary() {
  python3 - "$LOG_FILE" <<'PY'
import json
import os
import sys
from collections import Counter

path = sys.argv[1]
if not os.path.exists(path):
    print("No telemetry recorded yet.")
    raise SystemExit(0)

events = []
with open(path, "r", encoding="utf-8") as fh:
    for line in fh:
        line = line.strip()
        if line:
            events.append(json.loads(line))

print(f"Events: {len(events)}")
by_event = Counter(event["event"] for event in events)
for name, count in sorted(by_event.items()):
    print(f"- {name}: {count}")
PY
}

case "${1:-}" in
  record)
    shift
    record_event "${1:-unknown}" "${2:-ok}" "${3:-0}" "${4:-}"
    ;;
  summary)
    print_summary
    ;;
  *)
    cat <<'EOF'
Usage:
  bash ~/.brain/scripts/telemetry.sh record <event> [status] [duration_ms] [details]
  bash ~/.brain/scripts/telemetry.sh summary
EOF
    ;;
esac
