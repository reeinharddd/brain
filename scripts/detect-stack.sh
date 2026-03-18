#!/bin/bash
# Detect technology tags for the current project.

set -euo pipefail

PROJECT_ROOT="${1:-$PWD}"
START_TS="$(date +%s%3N 2>/dev/null || date +%s000)"

if [ ! -d "$PROJECT_ROOT" ]; then
  echo "ERROR: project root not found: $PROJECT_ROOT" >&2
  exit 1
fi

declare -a TAGS=()

add_tag() {
  local tag="$1"
  local existing
  for existing in "${TAGS[@]:-}"; do
    if [ "$existing" = "$tag" ]; then
      return
    fi
  done
  TAGS+=("$tag")
}

has_file() {
  [ -f "$PROJECT_ROOT/$1" ]
}

has_dir() {
  [ -d "$PROJECT_ROOT/$1" ]
}

file_contains() {
  local path="$1"
  local pattern="$2"
  [ -f "$PROJECT_ROOT/$path" ] && grep -Eq "$pattern" "$PROJECT_ROOT/$path"
}

if find "$PROJECT_ROOT" -maxdepth 2 -type f \( -name '*.sh' -o -name '.bashrc' -o -name '.zshrc' \) | grep -q .; then
  add_tag "bash"
  add_tag "shell"
fi

if find "$PROJECT_ROOT" -maxdepth 2 -type f \( -name '*.md' -o -name '*.mdx' \) | grep -q .; then
  add_tag "markdown"
  add_tag "docs"
fi

if has_file "package.json"; then
  add_tag "nodejs"
fi

if has_file "tsconfig.json" || find "$PROJECT_ROOT" -maxdepth 3 -type f \( -name '*.ts' -o -name '*.tsx' \) | grep -q .; then
  add_tag "typescript"
fi

if file_contains "package.json" '"react"|react-dom' || has_dir "src/components"; then
  add_tag "react"
fi

if file_contains "package.json" '"next"' || has_dir "app" || has_dir "pages"; then
  add_tag "nextjs"
fi

if has_file "pnpm-workspace.yaml" || has_file "turbo.json" || has_file "nx.json" || has_file "lerna.json"; then
  add_tag "monorepo"
fi

if has_file "pyproject.toml" || has_file "requirements.txt" || has_file "uv.lock"; then
  add_tag "python"
fi

if has_file "go.mod"; then
  add_tag "go"
fi

if has_file "Cargo.toml"; then
  add_tag "rust"
fi

if has_file "Dockerfile" || has_file "docker-compose.yml" || has_file "docker-compose.yaml" || has_dir "docker"; then
  add_tag "docker"
fi

if has_dir "domain" || has_dir "application" || has_dir "usecases" || has_dir "ports" || has_dir "adapters" || has_dir "internal/domain"; then
  add_tag "clean-architecture"
fi

printf '%s\n' "${TAGS[@]}" | sort -u

if [ -x "${HOME}/.brain/scripts/telemetry.sh" ]; then
  END_TS="$(date +%s%3N 2>/dev/null || date +%s000)"
  DURATION_MS="$((END_TS - START_TS))"
  bash "${HOME}/.brain/scripts/telemetry.sh" record "detect-stack" "ok" "$DURATION_MS" "$PROJECT_ROOT" >/dev/null 2>&1 || true
fi
