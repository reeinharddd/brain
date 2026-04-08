#!/usr/bin/env python3

import json
import sys
import copy
from pathlib import Path

import jsonschema


ROOT = Path(__file__).resolve().parents[2]
SCHEMAS_DIR = ROOT / "artifacts" / "schemas"

SCHEMA_BY_KIND = {
    "agent": SCHEMAS_DIR / "agent-artifact.schema.json",
    "command": SCHEMAS_DIR / "command-artifact.schema.json",
    "rule": SCHEMAS_DIR / "rule-artifact.schema.json",
    "mcp": SCHEMAS_DIR / "mcp-artifact.schema.json",
    "skill": SCHEMAS_DIR / "skill-artifact.schema.json",
    "adapter": SCHEMAS_DIR / "adapter-artifact.schema.json",
    "memory": SCHEMAS_DIR / "memory-artifact.schema.json",
    "ai": SCHEMAS_DIR / "ai-artifact.schema.json",
    "identity": SCHEMAS_DIR / "identity-artifact.schema.json",
    "policy": SCHEMAS_DIR / "policy-artifact.schema.json",
}

BASE_SCHEMAS = [
    SCHEMAS_DIR / "artifact-envelope.schema.json",
]

MANIFEST_GLOBS = [
    ROOT / "artifacts" / "agents" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "commands" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "rules" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "adapters" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "mcps" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "skills" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "providers" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "memory" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "ai" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "identity" / "manifests" / "*.artifact.json",
    ROOT / "artifacts" / "policy" / "manifests" / "*.artifact.json",
]


def load_json(path: Path):
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def resolve_refs(node, schema_store):
    if isinstance(node, dict):
        if "$ref" in node and isinstance(node["$ref"], str):
            ref = node["$ref"]
            if ref not in schema_store:
                raise KeyError(f"missing schema ref: {ref}")
            return resolve_refs(copy.deepcopy(schema_store[ref]), schema_store)
        return {key: resolve_refs(value, schema_store) for key, value in node.items()}
    if isinstance(node, list):
        return [resolve_refs(item, schema_store) for item in node]
    return node


def main() -> int:
    errors = []
    schema_store = {}

    for schema_path in list(SCHEMA_BY_KIND.values()) + BASE_SCHEMAS:
        if not schema_path.exists():
            continue
        schema = load_json(schema_path)
        schema_store[schema_path.resolve().as_uri()] = schema
        schema_id = schema.get("$id")
        if schema_id:
            schema_store[schema_id] = schema

    manifests = []
    for pattern in MANIFEST_GLOBS:
        manifests.extend(sorted(pattern.parent.glob(pattern.name)))

    if not manifests:
        print("[artifact-manifests] no manifests found")
        return 1

    for manifest_path in manifests:
        try:
            payload = load_json(manifest_path)
        except Exception as exc:
            errors.append(f"{manifest_path}: invalid JSON: {exc}")
            continue

        kind = payload.get("kind")
        schema_path = SCHEMA_BY_KIND.get(kind)
        if not schema_path or not schema_path.exists():
            errors.append(f"{manifest_path}: no schema for kind '{kind}'")
            continue

        try:
            schema = load_json(schema_path)
            resolved_schema = resolve_refs(schema, schema_store)
            jsonschema.Draft202012Validator(resolved_schema).validate(payload)
        except Exception as exc:
            errors.append(f"{manifest_path}: schema validation failed: {exc}")

    if errors:
        print("[artifact-manifests] failed")
        for err in errors:
            print(f"  - {err}")
        return 1

    print(f"[artifact-manifests] ok ({len(manifests)} manifests)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
