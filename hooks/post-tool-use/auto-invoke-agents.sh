#!/bin/bash
# post-tool-use/auto-invoke-agents.sh
# Hook for intelligent agent suggestion based on the tool used.

TOOL_NAME="$CLAUDE_TOOL_NAME"

case "$TOOL_NAME" in
  "write_file"|"edit_file"|"replace_file_content"|"multi_replace_file_content")
    echo "[SUGGEST] Task involves file writing. Consider consulting @reviewer."
    ;;
  "run_command"|"send_command_input")
    echo "[SUGGEST] Shell execution detected. Consider consulting @guardian for security."
    ;;
  "grep_search"|"find_by_name"|"search_web")
    echo "[SUGGEST] Research activity detected. @researcher might have more context."
    ;;
  "task_boundary")
    echo "[SUGGEST] Planning in progress. Ensure @planner has defined the roadmap."
    ;;
  *)
    # No specific suggestion
    ;;
esac
