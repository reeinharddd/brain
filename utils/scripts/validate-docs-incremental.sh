#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

MODE="staged"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --staged)
      MODE="staged"
      shift
      ;;
    --all)
      MODE="all"
      shift
      ;;
    *)
      echo "Usage: $0 [--staged|--all]" >&2
      exit 2
      ;;
  esac
done

python3 - "$MODE" <<'PY'
from __future__ import annotations

import json
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

mode = sys.argv[1]
root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
docs_root = root / "docs"
manifest_path = docs_root / "metadata" / "docs-manifest.json"

emoji_re = re.compile(r"[\U0001F300-\U0001FAFF\U00002600-\U000027BF]")
disallowed_tokens = [
    "✅",
    "❌",
    "⚠️",
    "⚠",
    "→",
    "←",
    "↑",
    "↓",
    "➜",
    "➤",
    "✔",
    "✖",
    "☑",
    "📝",
    "🎯",
    "🧪",
    "🤖",
    "💾",
    "📌",
    "🔗",
    "🚀",
    "✨",
    "💡",
    "🔒",
    "🛑",
    "🔥",
]

@dataclass(frozen=True)
class DomainRules:
    name: str
    files: set[str]
    reference_files: set[str]
    strict_files: set[str]
    frontmatter_required: bool
    required_sections: list[str]
    has_strict_validation: bool


def git(*args: str) -> list[str]:
    result = subprocess.check_output(["git", *args], text=True)
    return [line for line in result.splitlines() if line.strip()]


def load_manifest() -> dict:
    try:
        return json.loads(manifest_path.read_text(encoding="utf-8"))
    except Exception as exc:  # pragma: no cover - surfaced to user
        print(f"[docs] failed to parse manifest: {exc}", file=sys.stderr)
        raise SystemExit(1)


def normalize_domain_paths(domain: str, entries: list[str]) -> set[str]:
    if domain == "templates":
        return {f"docs/templates/{entry}" for entry in entries}
    return {f"docs/{domain}/{entry}" for entry in entries}


def build_rules(manifest: dict) -> tuple[set[str], dict[str, DomainRules]]:
    expected: set[str] = set()
    domains: dict[str, DomainRules] = {}

    for root_name in manifest["root_files"]:
        expected.add(f"docs/{root_name}")

    for domain, cfg in manifest["domains"].items():
        domain_files = normalize_domain_paths(domain, cfg.get("files", []))
        expected |= domain_files

        reference_files = normalize_domain_paths(domain, cfg.get("reference_files", []))
        navigation_files = normalize_domain_paths(domain, cfg.get("navigation_files", []))
        reference_files |= navigation_files

        if cfg.get("validation_level") == "reference":
            reference_files |= domain_files

        strict_files = domain_files - reference_files
        domains[domain] = DomainRules(
            name=domain,
            files=domain_files,
            reference_files=reference_files,
            strict_files=strict_files,
            frontmatter_required=bool(cfg.get("rules", {}).get("frontmatter_required", False)),
            required_sections=list(cfg.get("rules", {}).get("min_sections", [])),
            has_strict_validation=cfg.get("validation_level") == "strict",
        )

    return expected, domains


def parse_frontmatter(text: str) -> dict[str, str] | None:
    match = re.match(r"^---\n(.*?)\n---\n", text, re.S)
    if not match:
        return None

    data: dict[str, str] = {}
    for raw_line in match.group(1).splitlines():
        if ":" not in raw_line:
            continue
        key, value = raw_line.split(":", 1)
        key = key.strip()
        value = value.strip()
        if key and key not in data:
            data[key] = value
    return data


def collect_candidates() -> list[tuple[str, str]]:
    if mode == "all":
        raw = git("ls-files", "docs")
    else:
        raw = git("diff", "--cached", "--name-status", "--diff-filter=ACMRD", "--", "docs")

    candidates: list[tuple[str, str]] = []
    for line in raw:
        parts = line.split("\t")
        if mode == "all":
            path = parts[0]
            candidates.append(("M", path))
            continue

        status = parts[0]
        if status.startswith("R") and len(parts) >= 3:
            candidates.append((status, parts[2]))
        elif len(parts) >= 2:
            candidates.append((status, parts[1]))
    return candidates


def is_disallowed_symbol_present(text: str) -> list[str]:
    hits: list[str] = []
    if emoji_re.search(text):
        hits.append("unicode-emoji-range")
    for token in disallowed_tokens:
        if token in text:
            hits.append(token)
    return sorted(set(hits))


manifest = load_manifest()
expected_paths, domain_rules = build_rules(manifest)
actual_canonical = {
    f"docs/{path.relative_to(docs_root).as_posix()}"
    for path in docs_root.rglob("*")
    if path.is_file() and path.relative_to(docs_root).as_posix() not in {
        "metadata/docs-manifest.json",
        "metadata/docs-changelog.jsonl",
    }
}

errors: list[str] = []

# Validate the manifest against the current tree whenever docs validation runs.
missing_from_manifest = sorted(actual_canonical - expected_paths)
extra_in_manifest = sorted(expected_paths - actual_canonical)
if missing_from_manifest or extra_in_manifest:
    errors.append("[docs] manifest is out of sync with the canonical tree")
    if missing_from_manifest:
        errors.append("[docs] files present but missing from manifest:")
        errors.extend([f"  - {item}" for item in missing_from_manifest])
    if extra_in_manifest:
        errors.append("[docs] manifest entries missing from tree:")
        errors.extend([f"  - {item}" for item in extra_in_manifest])

for domain, cfg in manifest["domains"].items():
    files = set(cfg.get("files", []))
    navigation_files = set(cfg.get("navigation_files", []))
    reference_files = set(cfg.get("reference_files", []))
    if not navigation_files.issubset(files):
        errors.append(f"[docs] domain {domain} has navigation files outside its files list")
    if not reference_files.issubset(files):
        errors.append(f"[docs] domain {domain} has reference files outside its files list")

candidates = collect_candidates()
if not candidates:
    print("[docs] no candidate files to validate")
    sys.exit(0)

required_frontmatter = ["type", "id", "title", "version", "status", "date_created", "language", "category"]

for status, rel_path in candidates:
    if not rel_path.startswith("docs/"):
        continue

    path = root / rel_path
    if rel_path == "docs/metadata/docs-manifest.json":
        continue
    if rel_path == "docs/metadata/docs-changelog.jsonl":
        continue

    if not path.exists():
        if rel_path in expected_paths:
            errors.append(f"[docs] deleted canonical file without manifest update: {rel_path}")
        continue

    if rel_path not in expected_paths:
        errors.append(f"[docs] file is not declared in manifest: {rel_path}")
        continue

    if rel_path == "docs/DOCUMENTATION-UNIFIED-SCHEMA.md":
        # Root reference doc: frontmatter plus ASCII/plain-text rules.
        pass

    # Find the owning domain or root file.
    domain = None
    is_template_doc = rel_path.startswith("docs/templates/")
    if rel_path in {f"docs/{name}" for name in manifest["root_files"]}:
        domain = "root"
        frontmatter_required = True
        strict_sections: list[str] = []
    else:
        for candidate_domain, rule in domain_rules.items():
            if rel_path in rule.files:
                domain = candidate_domain
                frontmatter_required = rule.frontmatter_required
                strict_sections = rule.required_sections if rel_path in rule.strict_files else []
                break
        else:
            domain = "unknown"
            frontmatter_required = False
            strict_sections = []

    text = path.read_text(encoding="utf-8")
    fm = parse_frontmatter(text)

    if frontmatter_required and fm is None:
        errors.append(f"[docs] missing frontmatter: {rel_path}")
    elif fm is not None and frontmatter_required:
        missing_fields = [field for field in required_frontmatter if field not in fm or not fm[field].strip()]
        if missing_fields:
            errors.append(f"[docs] missing frontmatter fields in {rel_path}: {', '.join(missing_fields)}")
        if fm.get("language", "").lower() != "en":
            errors.append(f"[docs] language must be en in {rel_path}")

    if not is_template_doc:
        symbol_hits = is_disallowed_symbol_present(text)
        if symbol_hits:
            errors.append(f"[docs] disallowed symbols in {rel_path}: {', '.join(symbol_hits)}")

    if strict_sections and not is_template_doc:
        headings = {
            match.group(1).strip().lower()
            for match in re.finditer(r"^#{2,6}\s+(.+?)\s*$", text, re.M)
        }
        missing_sections = [section for section in strict_sections if section.lower() not in headings]
        if missing_sections:
            errors.append(
                f"[docs] missing required sections in {rel_path}: {', '.join(missing_sections)}"
            )

# Manifest counts should match the actual canonical tree counts.
if manifest.get("file_count_target"):
    file_count_target = manifest["file_count_target"]
    if file_count_target.get("current") != len(expected_paths):
        errors.append(
            f"[docs] manifest current file count is {file_count_target.get('current')} but expected {len(expected_paths)}"
        )
    per_domain = file_count_target.get("per_domain", {})
    for domain, rule in domain_rules.items():
        current_count = len(rule.files)
        if domain in per_domain and per_domain[domain].get("current") != current_count:
            errors.append(
                f"[docs] manifest current count for {domain} is {per_domain[domain].get('current')} but expected {current_count}"
            )

if errors:
    print("[docs] validation failed")
    for line in errors:
        print(line)
    sys.exit(1)

print("[docs] validation ok")
PY