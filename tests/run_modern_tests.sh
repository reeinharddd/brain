#!/bin/bash
# run_modern_tests.sh
# Executes all modern agentic tests and generates a unified JSON report.

set -uo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
TESTS_DIR="$BRAIN_DIR/tests/modern"
RESULTS_DIR="$BRAIN_DIR/tests/results"
RUN_ID="$(date -u '+%Y%m%dT%H%M%SZ')"

mkdir -p "$RESULTS_DIR"

G='\033[0;32m'; Y='\033[1;33m'; R='\033[0;31m'; B='\033[1m'; RS='\033[0m'
ok()   { echo -e "  ${G}[PASS]${RS} $1"; }
fail() { echo -e "  ${R}[FAIL]${RS} $1"; }
warn() { echo -e "  ${Y}[WARN]${RS} $1"; }
skip() { echo -e "  ${Y}[SKIP]${RS} $1"; }

echo ""
echo -e "${B}Brain Modern Agentic Tests --- $RUN_ID${RS}"
echo ""

TOTAL=0; PASSED=0; FAILED=0; SKIPPED=0
declare -a SUMMARY=()

run_test() {
    local name="$1"
    local script="$2"
    TOTAL=$((TOTAL+1))

    if [ ! -f "$script" ]; then
        warn "$name --- script not found: $script"
        SKIPPED=$((SKIPPED+1))
        SUMMARY+=("{\"name\":\"$name\",\"status\":\"SKIP\",\"reason\":\"script not found\"}")
        return
    fi

    echo -e "\n${B}-- $name${RS}"
    python3 "$script" 2>&1
    EXIT=$?
    if [ $EXIT -eq 0 ]; then
        ok "$name"
        PASSED=$((PASSED+1))
        SUMMARY+=("{\"name\":\"$name\",\"status\":\"PASS\"}")
    else
        fail "$name (exit $EXIT)"
        FAILED=$((FAILED+1))
        SUMMARY+=("{\"name\":\"$name\",\"status\":\"FAIL\",\"exit\":$EXIT}")
    fi
}

run_test "Parallel Sub-Agents"     "$TESTS_DIR/test_parallel_agents.py"
run_test "Real MCP Usage"          "$TESTS_DIR/test_mcp_real_usage.py"
run_test "Context Management"      "$TESTS_DIR/test_context_management.py"
run_test "Dynamic Skills"          "$TESTS_DIR/test_dynamic_skills.py"
run_test "Model Compatibility"     "$TESTS_DIR/test_model_compatibility.py"

echo ""
echo "======================================="
echo -e "${B}Results: $PASSED/$TOTAL PASS | $FAILED FAIL | $SKIPPED SKIP${RS}"
echo "======================================="

# Consolidate individual JSON results
python3 - << PY
import json, pathlib, sys

results_dir = pathlib.Path("$RESULTS_DIR")
combined = {}
for f in sorted(results_dir.glob("*.json")):
    try:
        combined[f.stem] = json.loads(f.read_text())
    except Exception:
        pass

report = {
    "run_id": "$RUN_ID",
    "total": $TOTAL,
    "passed": $PASSED,
    "failed": $FAILED,
    "skipped": $SKIPPED,
    "results": combined,
}
out = results_dir / "REPORT_${RUN_ID}.json"
out.write_text(json.dumps(report, indent=2))
print(f"Report saved: {out}")
PY

[ $FAILED -eq 0 ]
