#!/bin/bash
# ai-local/scripts/check-ollama.sh

if curl -s http://localhost:11435/api/tags >/dev/null; then
  echo "Ollama is running and responding."
  exit 0
else
  echo "Ollama is NOT running. Start it with: cd ~/.brain/ai-local && docker compose up -d" >&2
  exit 1
fi
