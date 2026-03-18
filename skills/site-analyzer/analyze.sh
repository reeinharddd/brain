#!/bin/bash

# analyze.sh - Core logic for the site-analyzer skill

BRAIN_DIR="/home/reeinharrrd/.brain"
SOURCES_FILE="$BRAIN_DIR/docs/sources.md"
TMP_TARGETS="/tmp/crawl_targets.txt"

echo "Parsing URLs from $SOURCES_FILE..."

# Extract Markdown links using grep/sed
grep -oP '\[.*?\]\(\Khttps?://.*?(?=\))' "$SOURCES_FILE" | sort -u > "$TMP_TARGETS"

COUNT=$(wc -l < "$TMP_TARGETS")
echo "Found $COUNT unique target URLs."

if [ "$COUNT" -eq 0 ]; then
    echo "No URLs found to analyze."
    exit 0
fi

echo "Targets identified:"
cat "$TMP_TARGETS"

echo "---"
echo "Recommended Action: Run an agentic crawl on the first 3 targets to identify new assets."
echo "Priority targets (Awesome lists):"
grep "awesome" "$TMP_TARGETS" || true
