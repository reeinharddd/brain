#!/bin/bash

# QUICK VERIFICATION TEST - Copy and paste these commands one by one in your terminal

echo "=== VERIFICATION TESTS ==="
echo ""
echo "TEST 1: Qdrant Health"
echo "Command: curl http://localhost:6333/health"
echo "Expected: Should return JSON with status"
echo ""

echo "TEST 2: Daemon Status Endpoint"  
echo "Command: curl http://localhost:9090/api/docs/status"
echo "Expected: JSON with index status"
echo ""

echo "TEST 3: Search Endpoint (First - no cache)"
echo "Command: time curl 'http://localhost:9090/api/docs/search?q=architecture&limit=5' | jq .results"
echo "Expected: Array of results, timing ~400-500ms"
echo ""

echo "TEST 4: Search Endpoint (Second - with cache)"
echo "Command: time curl 'http://localhost:9090/api/docs/search?q=architecture&limit=5' | jq .results"
echo "Expected: Same results, timing ~50-100ms (cached!)"
echo ""

echo "TEST 5: Domain Filter"
echo "Command: curl 'http://localhost:9090/api/docs/search?q=testing&domain=testing&limit=3' | jq ."
echo "Expected: Results filtered to testing domain"
echo ""

echo "TEST 6: Rebuild Endpoint (dev only)"
echo "Command: curl -X POST http://localhost:9090/api/docs/rebuild"
echo "Expected: JSON with success status"
echo ""

echo "=== HOW TO RUN TESTS ==="
echo "1. Open a NEW terminal"
echo "2. Make sure daemon is running (in another terminal)"
echo "3. Copy each TEST command and run it"
echo "4. Check the Expected output"
echo ""
