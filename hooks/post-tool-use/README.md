# post-tool-use/README.md

Estos hooks permiten:
- Auto-actualizar skills y agentes tras cada tool use (`run-auto-update.sh`).
- Autoinvocar agentes/subagentes/skills según el tipo de tarea detectada (`auto-invoke-agents.sh`).

Integrar ambos hooks en el flujo post-tool-use para que el sistema sea proactivo y se mantenga sincronizado.

Patrones tomados de:
- https://github.com/Gentleman-Programming/agent-teams-lite (protocolo de orquestación)
- https://github.com/dyoshikawa/rulesync (auto-generación de adapters)
- https://github.com/SuperClaude-Org/SuperClaude_Framework (auto-trigger de agentes)

Para ampliar, conectar con scripts de cada agente y skill según el caso.
