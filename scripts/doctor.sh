#!/bin/bash
# doctor.sh - Full integrity, runtime, and functional checks for the brain repo.
# Upgraded: adds schema validation, embedding backend, brain-mcp-server, provider-proxy,
#           agent-runner, consolidation script, and new module checks.

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
    --fix)     FIX=1     ;;
    --verbose) VERBOSE=1 ;;
    --json)    JSON_OUTPUT=1 ;;
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
  local name="$1" status="$2" severity="${3:-}" detail="${4:-}" entry
  if [ -n "$severity" ]; then
    entry=$(printf '{"name":%s,"status":"%s","severity":"%s","detail":%s}' \
      "$(printf '%s' "$name" | json_escape)" "$status" "$severity" "$(printf '%s' "$detail" | json_escape)")
  else
    entry=$(printf '{"name":%s,"status":"%s","detail":%s}' \
      "$(printf '%s' "$name" | json_escape)" "$status" "$(printf '%s' "$detail" | json_escape)")
  fi
  RESULTS+=("$entry")
}

print_line() {
  local prefix="$1" name="$2" detail="${3:-}"
  if [ "$JSON_OUTPUT" -eq 0 ]; then
    [ -n "$detail" ] && echo "  $prefix $name - $detail" || echo "  $prefix $name"
  fi
}

assert_check() {
  local name="$1" command="$2" severity="${3:-error}" detail="${4:-}"
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
  return 0
}

# Auto-fix: regenerate adapters if requested
if [ "$FIX" -eq 1 ]; then
  [ -x "$BRAIN_DIR/adapters/generate.sh" ] && bash "$BRAIN_DIR/adapters/generate.sh" >/dev/null 2>&1 || true
fi

# ── Core structure ─────────────────────────────────────────────────────────────
assert_check "brain_dir_exists"           "[ -d '$BRAIN_DIR' ]"
assert_check "git_repo_exists"            "[ -d '$BRAIN_DIR/.git' ]"
assert_check "canonical_exists"           "[ -s '$BRAIN_DIR/rules/canonical.md' ]"
assert_check "modules_dir_exists"         "[ -d '$BRAIN_DIR/rules/modules' ]"
assert_check "compiled_manifest_exists"   "[ -f '$BRAIN_DIR/rules/compiled/manifest.md' ]"
assert_check "build_rules_executable"     "[ -x '$BRAIN_DIR/scripts/build-rules.sh' ]"

# ── Required modules ────────────────────────────────────────────────────────────
assert_check "module_code_style"          "[ -f '$BRAIN_DIR/rules/modules/code-style.md' ]"
assert_check "module_security"            "[ -f '$BRAIN_DIR/rules/modules/security.md' ]"
assert_check "module_git"                 "[ -f '$BRAIN_DIR/rules/modules/git.md' ]"
assert_check "module_workflow"            "[ -f '$BRAIN_DIR/rules/modules/workflow.md' ]"
assert_check "module_memory"              "[ -f '$BRAIN_DIR/rules/modules/memory.md' ]"
assert_check "module_memory_types"        "[ -f '$BRAIN_DIR/rules/modules/memory-types.md' ]" "warning" "run: add memory-types.md from brain upgrades"
assert_check "module_memory_protocol"     "[ -f '$BRAIN_DIR/rules/modules/memory-protocol.md' ]"

# ── Canonical schema validation (new) ─────────────────────────────────────────
assert_check "canonical_schema_valid"     "python3 '$BRAIN_DIR/scripts/validate-schema.py'" "warning" "run: python3 ~/.brain/scripts/validate-schema.py"

# ── Agents ─────────────────────────────────────────────────────────────────────
for agent in orchestrator researcher planner architect implementer reviewer debugger designer documenter guardian; do
  assert_check "agent_${agent}_exists" "[ -f '$BRAIN_DIR/agents/${agent}.md' ]"
done

# ── Scripts ───────────────────────────────────────────────────────────────────
assert_check "generate_sh_executable"     "[ -x '$BRAIN_DIR/adapters/generate.sh' ]"
assert_check "detect_stack_executable"    "[ -x '$BRAIN_DIR/scripts/detect-stack.sh' ]"
assert_check "render_skill_context"       "[ -x '$BRAIN_DIR/scripts/render-skill-context.sh' ]"
assert_check "memory_namespace"           "[ -x '$BRAIN_DIR/scripts/memory-namespace.sh' ]"
assert_check "contextualize_sh"           "[ -x '$BRAIN_DIR/skills/codebase-contextualizer/contextualize.sh' ]"
assert_check "guardian_runner"            "[ -x '$BRAIN_DIR/guardian/run.sh' ]"
assert_check "guardian_wrapper"           "[ -x '$BRAIN_DIR/scripts/guardian.sh' ]"
assert_check "install_hooks"              "[ -x '$BRAIN_DIR/scripts/install-hooks.sh' ]"

# ── New upgraded scripts (warn if missing, encourage adoption) ─────────────────
assert_check "embed_py_exists"            "[ -f '$BRAIN_DIR/scripts/embed.py' ]"            "warning" "semantic embedding backend missing - vector search will use hash fallback"
assert_check "agent_runner_exists"        "[ -f '$BRAIN_DIR/scripts/agent-runner.py' ]"     "warning" "agent runner missing - agents cannot be executed programmatically"
assert_check "consolidate_memory_exists"  "[ -x '$BRAIN_DIR/scripts/consolidate-memory.sh' ]" "warning" "memory consolidation missing - stale memories will not be cleaned"
assert_check "provider_proxy_exists"      "[ -x '$BRAIN_DIR/scripts/provider-proxy.sh' ]"   "warning" "provider proxy missing - model routing not runtime-enforced"
assert_check "validate_schema_exists"     "[ -f '$BRAIN_DIR/scripts/validate-schema.py' ]"  "warning" "schema validator missing - canonical.md corruption undetectable"
assert_check "cron_setup_exists"          "[ -f '$BRAIN_DIR/scripts/cron-setup.sh' ]"       "warning" "cron setup missing - no automated maintenance"

# ── Brain MCP server (new) ─────────────────────────────────────────────────────
assert_check "brain_mcp_server_exists"    "[ -f '$BRAIN_DIR/mcp/brain-mcp-server/server.py' ]" "warning" "brain MCP server missing - agents cannot query rules at runtime"
assert_check "brain_mcp_server_runs"      "echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\",\"capabilities\":{},\"clientInfo\":{\"name\":\"doctor\",\"version\":\"1.0.0\"}}}' | python3 '$BRAIN_DIR/mcp/brain-mcp-server/server.py' --brain-dir '$BRAIN_DIR' | grep -q 'brain-mcp-server'" "warning"

# ── Adapters ───────────────────────────────────────────────────────────────────
assert_check "claude_adapter_exists"      "[ -f '$BRAIN_DIR/adapters/claude-code/CLAUDE.md' ]"
assert_check "gemini_adapter_exists"      "[ -f '$BRAIN_DIR/adapters/gemini/GEMINI.md' ]"
assert_check "cursor_adapter_exists"      "[ -f '$BRAIN_DIR/adapters/cursor/.cursorrules' ]"
assert_check "opencode_adapter_exists"    "[ -f '$BRAIN_DIR/adapters/opencode/opencode.json' ]"    "warning"
assert_check "opencode_json_valid"        "python3 -c \"import json; json.load(open('$BRAIN_DIR/adapters/opencode/opencode.json'))\"" "warning" "opencode.json is invalid JSON"
assert_check "mcp_global_json_valid"      "python3 -c \"import json; json.load(open('$BRAIN_DIR/mcp/global-config.json'))\"" "warning" "mcp/global-config.json is invalid JSON"

# ── Commands ───────────────────────────────────────────────────────────────────
for cmd in plan review research handover update-brain standup memory-search consolidate; do
  assert_check "command_${cmd}_exists" "[ -f '$BRAIN_DIR/commands/${cmd}.md' ]" "warning" "command /${cmd} not defined"
done

# ── Hooks ──────────────────────────────────────────────────────────────────────
assert_check "rules_hook_injects_modules"  "grep -q 'rules/modules' '$BRAIN_DIR/hooks/pre-tool-use/inject-global-rules.sh'"
assert_check "guardian_has_head_fallback"  "grep -q 'AUTO_FALLBACK_TO_HEAD' '$BRAIN_DIR/guardian/run.sh'"
assert_check "pre_commit_exists"           "[ -f '$BRAIN_DIR/hooks/pre-commit.sh' ]"

# ── Runtime dependencies ────────────────────────────────────────────────────────
assert_check "git_available"    "command -v git"
assert_check "bash_available"   "command -v bash"
assert_check "curl_available"   "command -v curl"
assert_check "python3_available" "command -v python3"
assert_check "node_available"   "command -v node"   "warning"
assert_check "docker_available" "command -v docker" "warning"
assert_check "ollama_running"   "bash '$BRAIN_DIR/ai-local/scripts/check-ollama.sh'" "warning" "run: docker compose up -d in ai-local"

# ── Functional tests ────────────────────────────────────────────────────────────
assert_check "stack_detection_runs"       "bash '$BRAIN_DIR/scripts/detect-stack.sh' '$BRAIN_DIR' >/dev/null"
assert_check "skill_context_write_runs"   "bash '$BRAIN_DIR/scripts/render-skill-context.sh' --write '$BRAIN_DIR' >/dev/null"
assert_check "embed_backend_runs"         "python3 '$BRAIN_DIR/scripts/embed.py' --info >/dev/null" "warning" "embed.py info check failed"
assert_check "agent_runner_lists"         "python3 '$BRAIN_DIR/scripts/agent-runner.py' --list >/dev/null" "warning" "agent-runner --list failed"

# ── MCP handshakes ─────────────────────────────────────────────────────────────
assert_check "memory_test_runs"           "bash '$BRAIN_DIR/scripts/test-memory.sh' >/dev/null"                        "warning"
assert_check "stdio_memory_handshake"     "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' memory >/dev/null"             "warning"
assert_check "stdio_filesystem_handshake" "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' filesystem >/dev/null"         "warning"
assert_check "stdio_sequential_handshake" "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' sequential >/dev/null"         "warning"
assert_check "stdio_context7_handshake"   "bash '$BRAIN_DIR/scripts/test-stdio-mcp.sh' context7 >/dev/null"           "warning"

# ── Vector / Qdrant ────────────────────────────────────────────────────────────
assert_check "vector_config_exists"       "[ -f '$BRAIN_DIR/memory/vector-config.json' ]"                             "warning"
assert_check "vector_sync_script_exists"  "[ -x '$BRAIN_DIR/scripts/vector-sync-qdrant.sh' ]"                         "warning"
assert_check "vector_http_reachable"      "curl -sf http://localhost:6333/collections >/dev/null"                      "warning" "qdrant not running - vector search unavailable"

# ── Evals ──────────────────────────────────────────────────────────────────────
assert_check "evals_run_executable"       "[ -x '$BRAIN_DIR/evals/run.sh' ]"                                          "warning"
assert_check "eval_adapter_schema"        "[ -x '$BRAIN_DIR/evals/benchmarks/adapter-schema.sh' ]"                    "warning"
assert_check "eval_guardian_coverage"     "[ -x '$BRAIN_DIR/evals/benchmarks/guardian-coverage.sh' ]"                 "warning"
assert_check "eval_memory_retrieval"      "[ -x '$BRAIN_DIR/evals/benchmarks/memory-retrieval.sh' ]"                  "warning"

# ── SDD integrity ──────────────────────────────────────────────────────────────
assert_check "sdd_flow_defined"           "grep -q 'Explore -> Propose -> Spec' '$BRAIN_DIR/docs/sdd/flow.md'"
assert_check "memory_rule_has_namespace"  "grep -q 'memory-namespace.sh' '$BRAIN_DIR/rules/modules/memory.md'"
assert_check "handover_mentions_engram"   "grep -Eiq 'engram|mem_session_summary' '$BRAIN_DIR/commands/handover.md'"

# ── Summary ────────────────────────────────────────────────────────────────────
TOTAL=$((PASS + FAIL + WARN))

if [ "$JSON_OUTPUT" -eq 1 ]; then
  printf '{"total":%d,"pass":%d,"fail":%d,"warn":%d,"results":[%s]}\n' \
    "$TOTAL" "$PASS" "$FAIL" "$WARN" "$(IFS=,; echo "${RESULTS[*]}")"
else
  echo ""
  echo "Brain Doctor v2"
  echo "Total: $TOTAL  Pass: $PASS  Fail: $FAIL  Warn: $WARN"
  if [ "$FAIL" -gt 0 ]; then
    echo "Status: FAIL"
  elif [ "$WARN" -gt 0 ]; then
    echo "Status: WARN (non-blocking)"
  else
    echo "Status: PASS"
  fi
fi

[ "$FAIL" -gt 0 ] && exit 1
exit 0
