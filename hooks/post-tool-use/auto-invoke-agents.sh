#!/bin/bash
# post-tool-use/auto-invoke-agents.sh
# Hook para autoinvocación de agentes/subagentes/skills según tipo de tarea

BRAIN_DIR="$HOME/.brain"
AGENTS_DIR="$BRAIN_DIR/agents"
SKILLS_DIR="$BRAIN_DIR/skills"

# Detectar tipo de tarea desde commit o mensaje
TASK_TYPE=$(git log -1 --pretty=%B | grep -Eo 'type:\s*\w+' | awk '{print $2}')

case "$TASK_TYPE" in
  "plan")
    echo "[HOOK] Invocando planner..."
    # Aquí se podría llamar a un script o agente específico
    ;;
  "review")
    echo "[HOOK] Invocando reviewer..."
    ;;
  "debug")
    echo "[HOOK] Invocando debugger..."
    ;;
  "refactor")
    echo "[HOOK] Invocando refactor..."
    ;;
  "research")
    echo "[HOOK] Invocando researcher..."
    ;;
  *)
    echo "[HOOK] No se detectó tipo de tarea específico."
    ;;
esac
