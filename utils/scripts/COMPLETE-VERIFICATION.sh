#!/bin/bash
# Complete Phase 5 Integration Verification Script
# Ejecutar: bash ~/.brain/COMPLETE-VERIFICATION.sh

set +e

echo "=========================================="
echo "  Phase 5: Complete Handler Integration"
echo "         Verification Script"
echo "=========================================="
echo ""
echo "Date: $(date)"
echo ""

# Step 1: Check prerequisites
echo "[Step 1] Checking prerequisites..."
echo "  - Qdrant: $(curl -s http://localhost:6333/health | head -c 30)..."
echo "  - MCP binary: $(ls -lh ~/.brain/bin/docs-rag-mcp | awk '{print $5}')"
echo "  - Daemon binary: $(ls -lh ~/.brain/bin/braind | awk '{print $5}')"
echo ""

# Step 2: Kill old processes
echo "[Step 2] Cleaning up old processes..."
pkill -9 braind 2>/dev/null
pkill -9 -f "go run" 2>/dev/null
sleep 3
echo "  ✅ Old processes killed"
echo ""

# Step 3: Start new daemon
echo "[Step 3] Starting daemon (with integrated handlers)..."
~/.brain/bin/braind > /tmp/braind-verify.log 2>&1 &
DAEMON_PID=$!
echo "  PID: $DAEMON_PID"
sleep 6
echo ""

# Step 4: Test endpoints
echo "[Step 4] Testing /api/docs endpoints..."
echo ""

# Test /api/docs/status
echo "  4a) GET /api/docs/status"
HTTP_CODE=$(curl -s -o /tmp/status-response.json -w "%{http_code}" http://localhost:9090/api/docs/status)
BODY=$(cat /tmp/status-response.json)
echo "       HTTP Code: $HTTP_CODE"
echo "       Body: ${BODY:0:100}..."
if [ "$HTTP_CODE" = "200" ]; then
    echo "       ✅ Status endpoint working!"
else
    echo "       ❌ Got HTTP $HTTP_CODE instead of 200"
fi
echo ""

# Test /api/docs/search
echo "  4b) GET /api/docs/search?q=daemon"
HTTP_CODE=$(curl -s -o /tmp/search-response.json -w "%{http_code}" "http://localhost:9090/api/docs/search?q=daemon&limit=3")
BODY=$(cat /tmp/search-response.json)
echo "       HTTP Code: $HTTP_CODE"
echo "       Body: ${BODY:0:150}..."
if [ "$HTTP_CODE" = "200" ]; then
    echo "       ✅ Search endpoint working!"
else
    echo "       ❌ Got HTTP $HTTP_CODE instead of 200"
fi
echo ""

# Test /api/docs/rebuild
echo "  4c) POST /api/docs/rebuild"
HTTP_CODE=$(curl -s -X POST -o /tmp/rebuild-response.json -w "%{http_code}" http://localhost:9090/api/docs/rebuild)
BODY=$(cat /tmp/rebuild-response.json)
echo "       HTTP Code: $HTTP_CODE"
echo "       Body: ${BODY:0:100}..."
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "503" ]; then
    echo "       ✅ Rebuild endpoint accessible!"
else
    echo "       ❌ Got HTTP $HTTP_CODE"
fi
echo ""

# Step 5: Show daemon logs
echo "[Step 5] Daemon startup logs (last 30 lines):"
echo "---"
tail -30 /tmp/braind-verify.log
echo "---"
echo ""

# Step 6: Verify structure
echo "[Step 6] Code structure verification:"
echo "  - Handler file: $([ -f ~/.brain/daemon/internal/api/handlers/docs.go ] && echo '✅' || echo '❌') docs.go"
echo "  - Stub indexer: $([ -f ~/.brain/daemon/cmd/braind/stub_indexer.go ] && echo '✅' || echo '❌') stub_indexer.go"
echo "  - Daemon binary: $([ -f ~/.brain/bin/braind ] && echo '✅' || echo '❌') braind"
echo ""

echo "=========================================="
echo "✅ Verification Complete!"
echo "=========================================="
echo ""
echo "Results:"
echo "  - HTTP endpoints should return 200 OK"
echo "  - Search should return mock results"
echo "  - Status should return index health"
echo "  - Rebuild is protected by BRAIN_ENV"
echo ""
echo "Next steps:"
echo "  1. If tests passed: Ready for React UI integration"
echo "  2. If tests failed: Check logs above for errors"
echo "  3. Run: bash ~/.brain/QUICK-VERIFICATION.sh for quick test"
echo ""
echo "To keep daemon running:"
echo "  nohup ~/.brain/bin/braind > /tmp/braind.log 2>&1 &"
echo ""
echo "To view live logs:"
echo "  tail -f /tmp/braind.log"
echo ""
