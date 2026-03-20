#!/bin/bash
# evals/benchmarks/guardian-coverage.sh
# Tests guardian checks against known-bad and known-good code samples.
# Measures: true positive rate (catches bad code) and false positive rate.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
JSON_OUTPUT=0
[ "${1:-}" = "--json" ] && JSON_OUTPUT=1

TMP_REPO="$(mktemp -d)"
trap 'rm -rf "$TMP_REPO"' EXIT

git -C "$TMP_REPO" init -q
git -C "$TMP_REPO" config user.email "eval@example.com"
git -C "$TMP_REPO" config user.name "Eval Bot"
touch "$TMP_REPO/README.md"
git -C "$TMP_REPO" add README.md
git -C "$TMP_REPO" commit -m "initial commit" -q

PASS=0; FAIL=0
declare -a RESULTS=()

run_guardian() {
  local label="$1" file="$2" content="$3" expect_verdict="$4"
  echo "$content" > "$TMP_REPO/$file"
  git -C "$TMP_REPO" add "$file" 2>/dev/null
  
  OUTPUT="$(GUARDIAN_REPO_ROOT="$TMP_REPO" bash "$BRAIN_DIR/guardian/run.sh" --staged --output json || true)"
  VERDICT=$(echo "$OUTPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('verdict','pass'))" 2>/dev/null || echo "pass")
  
  if [ "$VERDICT" = "$expect_verdict" ]; then
    PASS=$((PASS+1))
    RESULTS+=("PASS | $label | expected=$expect_verdict got=$VERDICT")
  else
    FAIL=$((FAIL+1))
    RESULTS+=("FAIL | $label | expected=$expect_verdict got=$VERDICT")
  fi
  git -C "$TMP_REPO" reset HEAD "$file" 2>/dev/null || true
  rm -f "$TMP_REPO/$file"
}

# Known-bad: hardcoded secret
run_guardian "hardcoded_api_key"       "bad_secret.ts"  'const API_KEY = "sk-1234567890abcdef";'           "block"
run_guardian "hardcoded_password"      "bad_pass.js"    'const password = "SuperSecret123!";'               "block"
run_guardian "explicit_any_typescript" "bad_any.ts"     'function process(data: any): any { return data; }' "block"

# Known-good: clean code
run_guardian "env_var_usage"           "good_env.ts"    'const API_KEY = process.env.API_KEY;'              "pass"
run_guardian "typed_typescript"        "good_typed.ts"  'function process(data: string): number { return data.length; }' "pass"
run_guardian "normal_constant"         "good_const.ts"  'const MAX_RETRIES = 3;'                            "pass"

TOTAL=$((PASS+FAIL))
TPR=$(python3 -c "print(round($PASS / max($TOTAL, 1) * 100, 1))")
STATUS="PASS"
[ "$PASS" -lt 4 ] && STATUS="FAIL"

if [ "$JSON_OUTPUT" -eq 1 ]; then
  python3 -c "
import json
items = []
for line in '''$(printf '%s\n' "${RESULTS[@]}")'''.splitlines():
    if not line.strip(): continue
    parts = line.split(' | ')
    items.append({'status': parts[0].strip(), 'label': parts[1].strip() if len(parts)>1 else '', 'detail': parts[2].strip() if len(parts)>2 else ''})
print(json.dumps({'benchmark': 'guardian-coverage', 'total': $TOTAL, 'passed': $PASS, 'accuracy_pct': $TPR, 'status': '$STATUS', 'cases': items}, indent=2))
"
else
  echo ""
  echo "Guardian Coverage Benchmark"
  echo "  Total   : $TOTAL"
  echo "  Accurate: $PASS"
  echo "  Accuracy: ${TPR}%"
  echo "  Status  : $STATUS"
  echo ""
  for line in "${RESULTS[@]}"; do
    echo "  $line"
  done
fi

[ "$STATUS" = "PASS" ]
