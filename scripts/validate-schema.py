#!/usr/bin/env python3
"""
validate-schema.py - Validates canonical.md and all rule modules
against expected structure. Fails loudly on schema violations.
"""

import re
import sys
import pathlib
import json

BRAIN_DIR = pathlib.Path.home() / ".brain"
if not BRAIN_DIR.exists():
    BRAIN_DIR = pathlib.Path(__file__).parent.parent

CANONICAL_PATH = BRAIN_DIR / "rules" / "canonical.md"
MODULES_DIR = BRAIN_DIR / "rules" / "modules"

REQUIRED_CANONICAL_SECTIONS = [
    "## Philosophy",
    "## Core Principles",
    "## How I Work with AI Agents",
    "## Language-Agnostic Code Standards",
    "## Project Structure Conventions",
    "## Communication Style",
]

REQUIRED_VERSION_HEADER = re.compile(r">\s*Version:\s*\d+\.\d+\.\d+")

REQUIRED_MODULE_FILES = [
    "code-style.md",
    "communication.md",
    "git.md",
    "security.md",
    "workflow.md",
    "memory.md",
]

errors = []
warnings = []
passed = 0

def check(condition, label, message="", severity="error"):
    global passed
    if condition:
        passed += 1
        return True
    if severity == "error":
        errors.append(f"FAIL  [{label}] {message}")
    else:
        warnings.append(f"WARN  [{label}] {message}")
    return False

# --- Check canonical.md exists and is non-empty ---
if not CANONICAL_PATH.exists():
    errors.append("FAIL  [canonical_exists] rules/canonical.md not found")
    print_results()
    sys.exit(1)

content = CANONICAL_PATH.read_text(encoding="utf-8")
lines = content.splitlines()

check(len(lines) > 20, "canonical_length", f"canonical.md is too short ({len(lines)} lines) - likely empty or stub")
check(REQUIRED_VERSION_HEADER.search(content), "canonical_version_header", "Missing version header: '> Version: X.X.X'")

# Check all required sections present
for section in REQUIRED_CANONICAL_SECTIONS:
    check(section in content, "canonical_section", f"Missing required section: {section}")

# Check for ASCII-only compliance (as per the rules themselves)
non_ascii = [(i+1, line) for i, line in enumerate(lines) if not all(ord(c) < 128 for c in line)]
if non_ascii:
    for lineno, line in non_ascii[:5]:
        warnings.append(f"WARN  [ascii_compliance] Line {lineno} contains non-ASCII characters: {repr(line[:80])}")

# Check no empty headings
empty_headings = [(i+1, line) for i, line in enumerate(lines) if re.match(r'^#{1,6}\s*$', line)]
for lineno, line in empty_headings:
    errors.append(f"FAIL  [empty_heading] Line {lineno}: empty heading found")

# Check no section longer than 5000 chars (bloat detector)
sections = re.split(r'\n(?=## )', content)
for section in sections:
    title_match = re.match(r'## (.+)', section)
    title = title_match.group(1) if title_match else "unknown"
    if len(section) > 5000:
        warnings.append(f"WARN  [section_bloat] Section '{title}' is {len(section)} chars - consider splitting into a module")

# --- Check modules ---
if not MODULES_DIR.exists():
    errors.append("FAIL  [modules_dir] rules/modules/ directory not found")
else:
    for module_file in REQUIRED_MODULE_FILES:
        module_path = MODULES_DIR / module_file
        check(module_path.exists(), "module_exists", f"Required module missing: rules/modules/{module_file}")
        if module_path.exists():
            module_content = module_path.read_text(encoding="utf-8")
            check(len(module_content.strip()) > 0, "module_non_empty", f"Module is empty: {module_file}")
            check("## " in module_content or "# " in module_content, "module_has_headings", f"Module has no headings: {module_file}", severity="warning")

# --- Output ---
output = {
    "schema_validation": {
        "passed": passed,
        "errors": len(errors),
        "warnings": len(warnings),
        "status": "FAIL" if errors else ("WARN" if warnings else "PASS"),
        "details": errors + warnings,
    }
}

json_mode = "--json" in sys.argv
if json_mode:
    print(json.dumps(output, indent=2))
else:
    print("\nCanonical Schema Validation")
    print(f"  Passed : {passed}")
    print(f"  Errors : {len(errors)}")
    print(f"  Warnings: {len(warnings)}")
    print()
    for line in errors:
        print(f"  {line}")
    for line in warnings:
        print(f"  {line}")
    print()
    status = output["schema_validation"]["status"]
    print(f"  Status: {status}")

sys.exit(1 if errors else 0)
