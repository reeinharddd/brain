#!/bin/bash
# evals/benchmarks/adapter-schema.sh
# Validates that generate.sh produces structurally correct adapter outputs.
# This is the test that prevents silent corruption from canonical.md changes.

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
JSON_OUTPUT=0
[ "${1:-}" = "--json" ] && JSON_OUTPUT=1

PASS=0; FAIL=0
declare -a RESULTS=()

check() {
  local label="$1" condition="$2"
  if eval "$condition" >/dev/null 2>&1; then
    PASS=$((PASS+1))
    RESULTS+=("PASS | $label")
  else
    FAIL=$((FAIL+1))
    RESULTS+=("FAIL | $label")
  fi
}

# Check all adapter files exist and are non-empty
check "claude_adapter_non_empty"   "[ -s '$BRAIN_DIR/adapters/claude-code/CLAUDE.md' ]"
check "cursor_adapter_non_empty"   "[ -s '$BRAIN_DIR/adapters/cursor/.cursorrules' ]"
check "windsurf_adapter_non_empty" "[ -s '$BRAIN_DIR/adapters/windsurf/.windsurfrules' ]"
check "gemini_adapter_non_empty"   "[ -s '$BRAIN_DIR/adapters/gemini/GEMINI.md' ]"
check "aider_adapter_non_empty"    "[ -s '$BRAIN_DIR/adapters/aider/system-prompt.md' ]"
check "opencode_adapter_valid_json" "python3 -c \"import json; json.load(open('$BRAIN_DIR/adapters/opencode/opencode.json'))\" "
check "claude_settings_valid_json"  "python3 -c \"import json; json.load(open('$BRAIN_DIR/adapters/claude-code/settings.json'))\" "
check "mcp_global_valid_json"       "python3 -c \"import json; json.load(open('$BRAIN_DIR/mcp/global-config.json'))\" "

# Check content of adapters contains canonical content
check "claude_has_core_principles"  "grep -q 'Core Principles' '$BRAIN_DIR/adapters/claude-code/CLAUDE.md'"
check "cursor_has_core_principles"  "grep -q 'Core Principles' '$BRAIN_DIR/adapters/cursor/.cursorrules'"
check "aider_has_auto_generated"    "grep -q 'AUTO-GENERATED' '$BRAIN_DIR/adapters/aider/.aider.conf.yml'"
check "opencode_has_system_prompt"  "python3 -c \"import json; d=json.load(open('$BRAIN_DIR/adapters/opencode/opencode.json')); assert 'systemPrompt' in d\""

# Check opencode.json mcpServers has required entries
check "opencode_has_memory_mcp"     "python3 -c \"import json; d=json.load(open('$BRAIN_DIR/adapters/opencode/opencode.json')); assert 'memory' in d['mcpServers']\""
check "opencode_has_filesystem_mcp" "python3 -c \"import json; d=json.load(open('$BRAIN_DIR/adapters/opencode/opencode.json')); assert 'filesystem' in d['mcpServers']\""

# Check canonical.md schema
check "canonical_schema_valid"      "python3 '$BRAIN_DIR/scripts/validate-schema.py'"

TOTAL=$((PASS+FAIL))
SCORE=$(python3 -c "print(round($PASS / max($TOTAL, 1) * 100, 1))")
STATUS="PASS"
[ "$FAIL" -gt 0 ] && STATUS="FAIL"

if [ "$JSON_OUTPUT" -eq 1 ]; then
  python3 -c "
import json
items = []
for line in '''$(printf '%s\n' "${RESULTS[@]}")'''.splitlines():
    if not line.strip(): continue
    parts = line.split(' | ', 1)
    items.append({'status': parts[0].strip(), 'label': parts[1].strip() if len(parts)>1 else ''})
print(json.dumps({'benchmark': 'adapter-schema', 'total': $TOTAL, 'passed': $PASS, 'score_pct': $SCORE, 'status': '$STATUS', 'cases': items}, indent=2))
"
else
  echo ""
  echo "Adapter Schema Benchmark"
  echo "  Total   : $TOTAL"
  echo "  Passed  : $PASS"
  echo "  Score   : ${SCORE}%"
  echo "  Status  : $STATUS"
  echo ""
  for line in "${RESULTS[@]}"; do
    echo "  $line"
  done
fi

[ "$STATUS" = "PASS" ]
