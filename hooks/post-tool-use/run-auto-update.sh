#!/bin/bash
# post-tool-use/run-auto-update.sh
# Hook para auto-actualización de skills y agentes tras cada tool use

BRAIN_DIR="$HOME/.brain"
SKILLS_DIR="$BRAIN_DIR/skills"
AGENTS_DIR="$BRAIN_DIR/agents"
ADAPTERS_SCRIPT="$BRAIN_DIR/adapters/generate.sh"
SITE_ANALYZER_SCRIPT="$SKILLS_DIR/site-analyzer/analyze.sh"

# Detectar cambios en skills o agentes
if git diff --name-only | grep -E "skills/|agents/"; then
  echo "[HOOK] Cambios detectados en skills o agentes. Ejecutando actualización..."
  # Actualizar adapters
  bash "$ADAPTERS_SCRIPT"
  # Ejecutar análisis de skills
  bash "$SITE_ANALYZER_SCRIPT"
  echo "[HOOK] Actualización completa."
else
  echo "[HOOK] No hay cambios en skills o agentes."
fi
