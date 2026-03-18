# Orchestrator Enhancement: Delegación y Orquestación

## Patrón
- El agente principal lee las reglas globales (`BRAIN_GLOBAL_RULES`) al inicio de cada sesión.
- Delegación automática a subagentes según el tipo de tarea (plan, review, debug, research, etc.).
- Inyección de reglas base a cada subagente invocado.
- Uso de MCPs, skills, y memoria de sesiones previas.
- Orquestación completa del flujo, siguiendo los ejemplos:
  - https://github.com/Gentleman-Programming/agent-teams-lite (protocolo de orquestación)
  - https://github.com/SuperClaude-Org/SuperClaude_Framework (delegación y reglas)
  - https://github.com/vijaythecoder/awesome-claude-agents (team-configurator)

## Ejemplo de integración (pseudocódigo)

# ...existing code...
# Al iniciar sesión:
source ~/.brain/hooks/pre-tool-use/inject-global-rules.sh

# En cada delegación:
subagent_call() {
  local agent="$1"
  local task="$2"
  local rules="$BRAIN_GLOBAL_RULES"
  # Inyecta reglas y contexto
  echo "Delegando a $agent con reglas base..."
  # ...lógica de llamada al subagente...
}

# ...existing code...

## Documentación
- Este patrón asegura que todo LLM, agente o subagente reciba las reglas base y el contexto deseado.
- El orchestrator es responsable de la delegación, inyección y orquestación.
- Referencias a los patrones open-source incluidos.
