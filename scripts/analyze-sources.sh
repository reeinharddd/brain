#!/bin/bash

# analyze-sources.sh - Triggers deep crawling of research sources

set -e

BRAIN_DIR="/home/reeinharrrd/.brain"
SOURCES_FILE="$BRAIN_DIR/docs/sources.md"
REPORTS_DIR="$BRAIN_DIR/docs/reports/scrapes"

mkdir -p "$REPORTS_DIR"

echo "Starting deep analysis of sources defined in $SOURCES_FILE..."

# Check if crawl4ai is available (Docker)
if ! docker images | grep -q "unclecode/crawl4ai"; then
    echo "Crawl4AI image not found. Pulling latest..."
    docker pull unclecode/crawl4ai:latest
fi

# Logic for iterating sources would go here
# For now, it initializes the report structure

TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
REPORT_FILE="$REPORTS_DIR/analysis_$TIMESTAMP.md"

cat <<EOF > "$REPORT_FILE"
# Deep Analysis Report - $TIMESTAMP

## Status: Initialized
Targets extracted from sources.md.

## Next Step
Run 'mcp call crawl4ai' on detected URLs.
EOF

echo "Analysis initialized. Report created at $REPORT_FILE"
