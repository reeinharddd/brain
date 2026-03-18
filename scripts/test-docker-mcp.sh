#!/bin/bash
# test-docker-mcp.sh - Verify persistent Docker helper services

[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" # Load nvm

echo "Testing Docker helper services..."

# 1. Daemon Check
if ! docker info > /dev/null 2>&1; then
    echo "✗ Docker daemon not found or unreachable. Skipping Docker tests."
    exit 1
fi

# 2. Existing Containers Status
CONTAINERS=(brain-qdrant)
for container in "${CONTAINERS[@]}"; do
    if [ "$(docker inspect -f '{{.State.Running}}' "$container" 2>/dev/null || true)" = "true" ]; then
        echo "[ok] Container '$container' is running."
    else
        echo "[fail] Container '$container' is NOT running."
        exit 1
    fi
done

# 3. Registry file presence
if [ -f "mcp/docker_registry.yml" ]; then
    echo "[ok] Docker MCP Registry (mcp/docker_registry.yml) found and reachable."
else
    echo "[fail] Docker MCP Registry missing!"
    exit 1
fi

if curl -sf --max-time 5 --retry 5 --retry-delay 1 http://localhost:6333/collections >/dev/null 2>&1; then
    echo "[ok] Qdrant HTTP endpoint reachable."
else
    echo "[fail] Qdrant HTTP endpoint unreachable."
    exit 1
fi

echo "[ok] Docker helper implementation healthy and accessible."
exit 0
