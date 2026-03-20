#!/bin/bash
# ai-local/scripts/pull-models.sh

set -euo pipefail

MODELS=("qwen2.5-coder:7b" "qwen2.5-coder:32b" "deepseek-coder-v3:latest")

for model in "${MODELS[@]}"; do
  echo "Pulling $model..."
  if docker ps --format '{{.Names}}' | grep -q "^brain-ollama$"; then
    docker exec brain-ollama ollama pull "$model"
  elif command -v ollama >/dev/null; then
    ollama pull "$model"
  else
    echo "Ollama is not running. Please start it first."
    exit 1
  fi
done
echo "All models pulled successfully."
