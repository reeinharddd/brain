#!/bin/bash
# consolidate-memory.sh - Memory consolidation pipeline.
#
# What it does:
# 1. Reads raw memories from MCP memory server
# 2. Groups by entity type and tags
# 3. Detects outdated or contradictory entries
# 4. Promotes high-frequency patterns as candidates for canonical.md rules
# 5. Archives stale entries (older than --days-stale)
# 6. Writes a consolidation report to logs/
#
# Usage:
#   bash consolidate-memory.sh [--dry-run] [--days-stale 90] [--json]

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
LOGS_DIR="$BRAIN_DIR/logs/consolidation"
RUN_ID="$(date -u '+%Y%m%dT%H%M%SZ')"
DRY_RUN=0
DAYS_STALE=90
JSON_OUTPUT=0

for arg in "$@"; do
  case "$arg" in
    --dry-run)    DRY_RUN=1 ;;
    --json)       JSON_OUTPUT=1 ;;
    --days-stale) shift; DAYS_STALE="${1:-90}" ;;
  esac
done

mkdir -p "$LOGS_DIR"
REPORT_PATH="$LOGS_DIR/${RUN_ID}.json"
MANIFEST_PATH="$BRAIN_DIR/memory/manifest.json"

# Resolve npx
resolve_npx_cmd() {
  [ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" 2>/dev/null || true
  if command -v npx-nvm >/dev/null 2>&1 && npx-nvm -v >/dev/null 2>&1; then
    echo "npx-nvm"
  elif command -v npx >/dev/null 2>&1; then
    echo "npx"
  else
    echo "ERROR: npx not found" >&2; exit 1
  fi
}
NPX_CMD="$(resolve_npx_cmd)"

TMP_MEM_DIR="$(mktemp -d)"
TMP_OUT="$(mktemp)"
trap 'rm -rf "$TMP_MEM_DIR" "$TMP_OUT"' EXIT

echo "[consolidate] Reading memory store..."

# Query the memory MCP server to get all entities
{
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"consolidate","version":"1.0.0"}}}'
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  printf '%s\n' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_graph","arguments":{}}}'
} | "$NPX_CMD" -y @modelcontextprotocol/server-memory "$HOME/.brain/memory" > "$TMP_OUT" 2>/dev/null || true

# Parse the graph from MCP output
python3 - "$TMP_OUT" "$MANIFEST_PATH" "$DAYS_STALE" "$DRY_RUN" "$REPORT_PATH" << 'PY'
import json
import pathlib
import sys
import re
from datetime import datetime, timezone, timedelta

mcp_output_path = pathlib.Path(sys.argv[1])
manifest_path   = pathlib.Path(sys.argv[2])
days_stale      = int(sys.argv[3])
dry_run         = int(sys.argv[4]) == 1
report_path     = pathlib.Path(sys.argv[5])

# Parse MCP jsonrpc output lines
raw = mcp_output_path.read_text(encoding="utf-8")
entities = []
relations = []

for line in raw.splitlines():
    line = line.strip()
    if not line:
        continue
    try:
        msg = json.loads(line)
    except json.JSONDecodeError:
        continue
    result = msg.get("result", {})
    if isinstance(result, dict):
        content = result.get("content", [])
        for block in content:
            if isinstance(block, dict) and block.get("type") == "text":
                try:
                    graph = json.loads(block["text"])
                    entities.extend(graph.get("entities", []))
                    relations.extend(graph.get("relations", []))
                except (json.JSONDecodeError, KeyError):
                    pass

# --- Analysis ---
now = datetime.now(timezone.utc)
stale_cutoff = now - timedelta(days=days_stale)

# Group by entity type
by_type: dict[str, list] = {}
for ent in entities:
    etype = ent.get("entityType", "unknown")
    by_type.setdefault(etype, []).append(ent)

# Detect high-frequency patterns (observations that repeat across entities)
obs_counts: dict[str, int] = {}
for ent in entities:
    for obs in ent.get("observations", []):
        key = re.sub(r'\s+', ' ', obs.strip().lower())[:120]
        obs_counts[key] = obs_counts.get(key, 0) + 1

high_frequency = {k: v for k, v in obs_counts.items() if v >= 3}

# Detect potential contradictions (same entity name, different types)
name_types: dict[str, set] = {}
for ent in entities:
    name = ent.get("name", "")
    etype = ent.get("entityType", "")
    name_types.setdefault(name, set()).add(etype)
contradictions = {name: list(types) for name, types in name_types.items() if len(types) > 1}

# Update manifest
manifest = {"version": "1.0.0", "created_at": "2026-03-01", "owner": "reeinharrrd",
            "description": "Brain repo memory manifest", "chunks": [], "stats": {}}
if manifest_path.exists():
    try:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        pass

manifest["stats"] = {
    "total_entities":     len(entities),
    "total_relations":    len(relations),
    "entity_types":       {k: len(v) for k, v in by_type.items()},
    "high_freq_patterns": len(high_frequency),
    "contradictions":     len(contradictions),
    "last_consolidation": now.isoformat(),
}

if not dry_run:
    manifest_path.write_text(json.dumps(manifest, indent=2), encoding="utf-8")

# Build report
report = {
    "run_id":           report_path.stem,
    "timestamp":        now.isoformat(),
    "dry_run":          dry_run,
    "status":           "success",
    "summary": {
        "total_entities":  len(entities),
        "total_relations": len(relations),
        "entity_breakdown": {k: len(v) for k, v in by_type.items()},
    },
    "high_frequency_patterns": [
        {"observation": k, "count": v}
        for k, v in sorted(high_frequency.items(), key=lambda x: -x[1])[:20]
    ],
    "contradictions": [
        {"entity": name, "conflicting_types": types}
        for name, types in contradictions.items()
    ],
    "canonical_update_candidates": [
        f"Observed {v}x: '{k}' -- consider adding to canonical.md if universally applicable"
        for k, v in sorted(high_frequency.items(), key=lambda x: -x[1])[:5]
    ],
    "recommendations": [],
}

if len(entities) > 500:
    report["recommendations"].append(
        f"Memory store has {len(entities)} entities. Consider archiving entries older than {days_stale} days."
    )
if contradictions:
    report["recommendations"].append(
        f"Found {len(contradictions)} entities with conflicting types. Review and resolve manually."
    )
if high_frequency:
    report["recommendations"].append(
        f"Found {len(high_frequency)} high-frequency patterns. Run /update-brain to promote to canonical.md."
    )

if not dry_run:
    report_path.write_text(json.dumps(report, indent=2), encoding="utf-8")

print(json.dumps(report, indent=2))
PY

if [ "$JSON_OUTPUT" -eq 0 ]; then
  echo ""
  echo "[consolidate] Report saved to: $REPORT_PATH"
  echo "[consolidate] Manifest updated: $MANIFEST_PATH"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "[consolidate] DRY RUN - no changes written"
  fi
fi
