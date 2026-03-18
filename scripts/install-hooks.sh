#!/usr/bin/env bash
# install-hooks.sh - install the Brain Guardian pre-commit hook

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
GUARDIAN_SCRIPT="$BRAIN_DIR/scripts/guardian.sh"
GLOBAL=0
UNINSTALL=0

for arg in "$@"; do
  case "$arg" in
    --global) GLOBAL=1 ;;
    --uninstall) UNINSTALL=1 ;;
    *)
      echo "ERROR: unknown option: $arg" >&2
      exit 1
      ;;
  esac
done

if [ ! -x "$GUARDIAN_SCRIPT" ]; then
  echo "ERROR: guardian wrapper not found at $GUARDIAN_SCRIPT" >&2
  exit 1
fi

write_hook() {
  local hook_file="$1"
  cat > "$hook_file" <<HOOK
#!/usr/bin/env bash
# Brain Guardian pre-commit hook
export BRAIN_DIR="$BRAIN_DIR"
exec "$GUARDIAN_SCRIPT" --staged
HOOK
  chmod +x "$hook_file"
}

if [ "$GLOBAL" -eq 1 ]; then
  TEMPLATE_DIR="${GIT_TEMPLATE_DIR:-$HOME/.git-templates}"
  mkdir -p "$TEMPLATE_DIR/hooks"
  write_hook "$TEMPLATE_DIR/hooks/pre-commit"
  git config --global init.templateDir "$TEMPLATE_DIR"
  echo "[ok] Guardian hook installed globally via git template"
  exit 0
fi

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "ERROR: not inside a git repository" >&2
  exit 1
fi

HOOK_FILE="$(git rev-parse --git-dir)/hooks/pre-commit"

if [ "$UNINSTALL" -eq 1 ]; then
  if [ -f "$HOOK_FILE" ] && grep -q "Brain Guardian pre-commit hook" "$HOOK_FILE"; then
    rm -f "$HOOK_FILE"
    echo "[ok] Guardian hook uninstalled from current repository"
  else
    echo "[warn] No Brain Guardian hook found in current repository"
  fi
  exit 0
fi

if [ -f "$HOOK_FILE" ] && ! grep -q "Brain Guardian pre-commit hook" "$HOOK_FILE"; then
  cp "$HOOK_FILE" "${HOOK_FILE}.backup"
  echo "[ok] Existing hook backed up to ${HOOK_FILE}.backup"
fi

write_hook "$HOOK_FILE"
echo "[ok] Guardian hook installed in $(git rev-parse --show-toplevel)"
