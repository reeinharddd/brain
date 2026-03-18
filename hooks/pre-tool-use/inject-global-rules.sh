#!/bin/bash
# pre-tool-use/inject-global-rules.sh
# Inyecta reglas base (idioma, estilo, comentarios, commits, etc.) en cada sesión/tool use

BRAIN_DIR="$HOME/.brain"
RULES_FILE="$BRAIN_DIR/rules/canonical.md"
MODULES_DIR="$BRAIN_DIR/rules/modules"

# Ensambla reglas completas
FULL_RULES=$(cat "$RULES_FILE")
for module in "$MODULES_DIR"/*.md; do
  [ -f "$module" ] || continue
  FULL_RULES="${FULL_RULES}"$'\n\n'"$(cat "$module")"
done

# Exporta reglas como variable de entorno para inyección
export BRAIN_GLOBAL_RULES="$FULL_RULES"
echo "[HOOK] Reglas base inyectadas en la sesión/tool use."
