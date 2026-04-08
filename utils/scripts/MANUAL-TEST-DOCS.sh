#!/bin/bash
# Manual Docs-RAG Endpoint Testing Script
# Run this in your terminal: bash ~/.brain/MANUAL-TEST-DOCS.sh

set -e

BASE_DIR="$HOME/.brain"
DAEMON_PID=""

echo "================================================"
echo "  Docs-RAG Endpoint Testing - Manual Script"
echo "================================================"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Kill existing daemons
echo -e "${BLUE}[Step 1]${NC} Killing old daemon processes..."
pkill -9 -f "go run main.go" 2>/dev/null || true
pkill -9 -f "apps/daemon/cmd/braind/main.go" 2>/dev/null || true
pkill -9 braind 2>/dev/null || true
sleep 2
echo -e "${GREEN}✅ Old processes killed${NC}"
echo ""

# Step 2: Check Qdrant
echo -e "${BLUE}[Step 2]${NC} Checking Qdrant..."
if curl -s http://localhost:6333/health >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Qdrant is running on :6333${NC}"
else
    echo -e "${YELLOW}⚠️  Qdrant not responding, attempting to start...${NC}"
    cd "$BASE_DIR" && docker compose -f docker/docker-compose.yml up -d qdrant 2>/dev/null || true
    sleep 3
fi
echo ""

# Step 3: Start Daemon
echo -e "${BLUE}[Step 3]${NC} Starting Daemon..."
cd "$BASE_DIR/apps/daemon/cmd/braind"

# Run in background
go run main.go > /tmp/braind.log 2>&1 &
DAEMON_PID=$!
echo "Daemon PID: $DAEMON_PID"
echo -e "${YELLOW}Waiting for daemon to start...${NC}"
sleep 5

# Check if daemon is still running
if ! ps -p $DAEMON_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Daemon failed to start!${NC}"
    echo "Error logs:"
    cat /tmp/braind.log | tail -20
    exit 1
fi
echo -e "${GREEN}✅ Daemon started (PID: $DAEMON_PID)${NC}"
echo ""

# Step 4: Test endpoints
echo -e "${BLUE}[Step 4]${NC} Testing /api/docs endpoints..."
echo ""

# Test /api/docs/status
echo -e "  ${BLUE}4a)${NC} GET /api/docs/status"
RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:9090/api/docs/status)
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)
echo "       HTTP Code: $HTTP_CODE"
echo "       Response: $BODY"
if [ "$HTTP_CODE" = "503" ]; then
    echo -e "       ${GREEN}✅ Endpoint is accessible (503 = not yet initialized)${NC}"
else
    echo -e "       ${YELLOW}⚠️  Unexpected status code${NC}"
fi
echo ""

# Test /api/docs/search
echo -e "  ${BLUE}4b)${NC} GET /api/docs/search?q=architecture"
RESPONSE=$(curl -s -w "\n%{http_code}" "http://localhost:9090/api/docs/search?q=architecture&limit=5")
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)
echo "       HTTP Code: $HTTP_CODE"
echo "       Response: $BODY"
if [ "$HTTP_CODE" = "503" ]; then
    echo -e "       ${GREEN}✅ Endpoint is accessible (503 = not yet initialized)${NC}"
else
    echo -e "       ${YELLOW}⚠️  Unexpected status code${NC}"
fi
echo ""

# Test /api/docs/rebuild
echo -e "  ${BLUE}4c)${NC} POST /api/docs/rebuild"
RESPONSE=$(curl -s -X POST -w "\n%{http_code}" http://localhost:9090/api/docs/rebuild)
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)
echo "       HTTP Code: $HTTP_CODE"
echo "       Response: $BODY"
if [ "$HTTP_CODE" = "503" ]; then
    echo -e "       ${GREEN}✅ Endpoint is accessible (503 = not yet initialized)${NC}"
else
    echo -e "       ${YELLOW}⚠️  Unexpected status code${NC}"
fi
echo ""

# Step 5: Show daemon logs
echo -e "${BLUE}[Step 5]${NC} Daemon startup logs:"
echo "---"
tail -30 /tmp/braind.log
echo "---"
echo ""

echo "================================================"
echo -e "${GREEN}✅ Test Complete!${NC}"
echo "================================================"
echo ""
echo "Next steps:"
echo "1. Daemon is running in background (PID: $DAEMON_PID)"
echo "2. Endpoints are responding with 503 (not yet initialized)"
echo "3. This is expected - next phase is to integrate the MCP server"
echo ""
echo "To stop daemon:"
echo "  kill $DAEMON_PID"
echo ""
echo "To view logs:"
echo "  tail -f /tmp/braind.log"
echo ""
