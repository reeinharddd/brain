#!/bin/bash
# doctor.sh - integrity and runtime checks for the brain repo

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
FIX=0
VERBOSE=0
JSON_OUTPUT=0
PASS=0
FAIL=0
WARN=0
RESULTS=()

for arg in "$@"; do
  case "$arg" in
    --fix) FIX=1 ;;
    --verbose) VERBOSE=1 ;;
    --json) JSON_OUTPUT=1 ;;
    *)
      echo "ERROR: unknown option: $arg" >&2
      exit 1
      ;;
  esac
done

json_escape() {
  python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))'
}

record_result() {
  local name="$1"
  local status="$2"
  local severity="${3:-}"
  local detail="${4:-}"
  local entry

  if [ -n "$severity" ]; then
    entry=$(printf '{"name":%s,"status":"%s","severity":"%s","detail":%s}' \
      "$(printf '%s' "$name" | json_escape)" \
      "$status" \
      "$severity" \
      "$(printf '%s' "$detail" | json_escape)")
  else
    entry=$(printf '{"name":%s,"status":"%s","detail":%s}' \
      "$(printf '%s' "$name" | json_escape)" \
      "$status" \
      "$(printf '%s' "$detail" | json_escape)")
  fi
  RESULTS+=("$entry")
}

print_line() {
  local prefix="$1"
  local name="$2"
  local detail="${3:-}"
  if [ "$JSON_OUTPUT" -eq 0 ]; then
    if [ -n "$detail" ]; then
      echo "  $prefix $name - $detail"
    else
      echo "  $prefix $name"
    fi
  fi
}

assert_check() {
  local name="$1"
  local command="$2"
  local severity="${3:-error}"
  local detail="${4:-}"

  if eval "$command" >/dev/null 2>&1; then
    PASS=$((PASS + 1))
    record_result "$name" "pass" "" "$detail"
    [ "$VERBOSE" -eq 1 ] && print_line "PASS" "$name" "$detail"
  else
    if [ "$severity" = "warning" ]; then
      WARN=$((WARN + 1))
      record_result "$name" "warn" "warning" "$detail"
      print_line "WARN" "$name" "$detail"
    else
      FAIL=$((FAIL + 1))
      record_result "$name" "fail" "error" "$detail"
      print_line "FAIL" "$name" "$detail"
    fi
  fi
}

if [ "$FIX" -eq 1 ]; then
  if [ -x "$BRAIN_DIR/adapters/generate.sh" ]; then
    bash "$BRAIN_DIR/adapters/generate.sh" >/dev/null 2>&1 || true
  fi
fi

assert_check "brain_dir_exists" "[ -d '$BRAIN_DIR' ]"
assert_check "git_repo_exists" "[ -d '$BRAIN_DIR/.git' ]"
assert_check "canonical_exists" "[ -s '$BRAIN_DIR/rules/canonical.md' ]"
assert_check "modules_dir_exists" "[ -d '$BRAIN_DIR/rules/modules' ]"
assert_check "compiled_manifest_exists" "[ -f '$BRAIN_DIR/rules/compiled/manifest.md' ]"
assert_check "build_rules_executable" "[ -x '$BRAIN_DIR/scripts/build-rules.sh' ]"

assert_check "orchestrator_exists" "[ -f '$BRAIN_DIR/agents/orchestrator.md' ]"
assert_check "researcher_exists" "[ -f '$BRAIN_DIR/agents/researcher.md' ]"
assert_check "planner_exists" "[ -f '$BRAIN_DIR/agents/planner.md' ]"
assert_check "architect_exists" "[ -f '$BRAIN_DIR/agents/architect.md' ]"
assert_check "implementer_exists" "[ -f '$BRAIN_DIR/agents/implementer.md' ]"
assert_check "reviewer_exists" "[ -f '$BRAIN_DIR/agents/reviewer.md' ]"
assert_check "debugger_exists" "[ -f '$BRAIN_DIR/agents/debugger.md' ]"

assert_check "generate_sh_exists" "[ -x '$BRAIN_DIR/adapters/generate.sh' ]"
assert_check "detect_stack_exists" "[ -x '$BRAIN_DIR/scripts/detect-stack.sh' ]"
assert_check "render_skill_context_exists" "[ -x '$BRAIN_DIR/scripts/render-skill-context.sh' ]"
assert_check "memory_namespace_exists" "[ -x '$BRAIN_DIR/scripts/memory-namespace.sh' ]"
assert_check "contextualize_exists" "[ -x '$BRAIN_DIR/skills/codebase-contextualizer/contextualize.sh' ]"
assert_check "handover_exists" "[ -f '$BRAIN_DIR/commands/handover.md' ]"
assert_check "guardian_runner_exists" "[ -x '$BRAIN_DIR/guardian/run.sh' ]"
assert_check "guardian_wrapper_exists" "[ -x '$BRAIN_DIR/scripts/guardian.sh' ]"
assert_check "install_hooks_exists" "[ -x '$BRAIN_DIR/scripts/install-hooks.sh' ]"
assert_check "guardian_env_example_exists" "[ -f '$BRAIN_DIR/providers/guardian.env.example' ]"
assert_check "pre_commit_exists" "[ -f '$BRAIN_DIR/hooks/pre-commit.sh' ]"

assert_check "git_available" "command -v git"
assert_check "bash_available" "command -v bash"
assert_check "curl_available" "command -v curl"
assert_check "python3_available" "command -v python3"
assert_check "node_available" "command -v node" "warning"
assert_check "docker_available" "command -v docker" "warning"

assert_check "claude_adapter_exists" "[ -f '$BRAIN_DIR/adapters/claude-code/CLAUDE.md' ]"
assert_check "gemini_adapter_exists" "[ -f '$BRAIN_DIR/adapters/gemini/GEMINI.md' ]"
assert_check "cursor_adapter_exists" "[ -f '$BRAIN_DIR/adapters/cursor/.cursorrules' ]"
assert_check "opencode_adapter_exists" "[ -f '$BRAIN_DIR/adapters/opencode/opencode.json' ]" "warning"

assert_check "rules_hook_injects_modules" "grep -q 'rules/modules' '$BRAIN_DIR/hooks/pre-tool-use/inject-global-rules.sh'"
assert_check "sdd_flow_defined" "grep -q 'Explore -> Propose -> Spec' '$BRAIN_DIR/docs/sdd/flow.md'"
assert_check "memory_rule_has_namespace" "grep -q 'memory-namespace.sh' '$BRAIN_DIR/rules/modules/memory.md'"
assert_check "memory_protocol_exists" "[ -f '$BRAIN_DIR/rules/modules/memory-protocol.md' ]"
assert_check "handover_mentions_engram" "grep -Eiq 'engram|mem_session_summary' '$BRAIN_DIR/commands/handover.md'"
assert_check "guardian_has_head_fallback" "grep -q 'AUTO_FALLBACK_TO_HEAD' '$BRAIN_DIR/guardian/run.sh'"

assert_check "stack_detection_runs" "bash '$BRAIN_DIR/scripts/detect-stack.sh' '$BRAIN_DIR' >/dev/null"
assert_check "skill_context_write_runs" "bash '$BRAIN_DIR/scripts/render-skill-context.sh' --write '$BRAIN_DIR' >/dev/null"
assert_check "context_pack_build_runs" "bash '$BRAIN_DIR/skills/codebase-contextualizer/contextualize.sh' '$BRAIN_DIR' >/dev/null"
assert_check "memory_test_runs" "bash '$BRAIN_DIR/scripts/test-memory.sh' >/dev/null" "warning"
assert_check "stdio_memory_handshake" "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' memory >/dev/null" "warning"
assert_check "stdio_filesystem_handshake" "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' filesystem >/dev/null" "warning"
assert_check "stdio_sequential_handshake" "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' sequential >/dev/null" "warning"
assert_check "stdio_context7_handshake" "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' context7 >/dev/null" "warning"

assert_check "vector_config_exists" "[ -f '$BRAIN_DIR/memory/vector-config.json' ]" "warning"
assert_check "vector_sync_script_exists" "[ -x '$BRAIN_DIR/scripts/vector-sync-qdrant.sh' ]" "warning"
assert_check "vector_context_index_exists" "[ -f '$BRAIN_DIR/.brain/codebase-context.ndjson' ]" "warning"
assert_check "vector_http_reachable" "curl -sf http://localhost:6333/collections >/dev/null" "warning" "qdrant live check"

TOTAL=$((PASS + FAIL + WARN))

if [ "$JSON_OUTPUT" -eq 1 ]; then
  printf '{"total":%d,"pass":%d,"fail":%d,"warn":%d,"results":[%s]}\n' \
    "$TOTAL" "$PASS" "$FAIL" "$WARN" "$(IFS=,; echo "${RESULTS[*]}")"
else
  echo ""
  echo "Brain Doctor"
  echo "Total: $TOTAL  Pass: $PASS  Fail: $FAIL  Warn: $WARN"
  if [ "$FAIL" -gt 0 ]; then
    echo "Status: FAIL"
  elif [ "$WARN" -gt 0 ]; then
    echo "Status: WARN"
  else
    echo "Status: PASS"
  fi
fi

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
exit 0
