#!/bin/bash

set -euo pipefail

RESULTS_FILE="$1"
FAIL_COUNT="$2"

python3 - "$RESULTS_FILE" "$FAIL_COUNT" <<'PY'
import json
import sys

results_path = sys.argv[1]
fail_count = int(sys.argv[2])
findings = []

with open(results_path, "r", encoding="utf-8") as fh:
    for line in fh:
        line = line.rstrip("\n")
        if not line:
            continue
        severity, check_id, file_path, message = line.split("\t", 3)
        findings.append(
            {
                "severity": severity,
                "check_id": check_id,
                "file": file_path,
                "message": message,
            }
        )

payload = {
    "verdict": "block" if fail_count > 0 else ("warn" if findings else "pass"),
    "findings": findings,
}

print(json.dumps(payload, indent=2))
PY
