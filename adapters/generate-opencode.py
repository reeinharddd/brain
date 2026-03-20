#!/usr/bin/env python3
"""
generate-opencode.py - Generates opencode.json with proper JSON encoding.
Called by adapters/generate.sh.
"""
import json
import os
import pathlib
import subprocess
import sys

brain_dir = pathlib.Path(os.environ.get("BRAIN_DIR", str(pathlib.Path.home() / ".brain")))

result = subprocess.run(
    ["bash", str(brain_dir / "scripts" / "build-rules.sh"), "--stdout"],
    capture_output=True, text=True,
    env=dict(os.environ, BRAIN_DIR=str(brain_dir))
)
full_rules = result.stdout

opencode = {
    "//": "AUTO-GENERATED - DO NOT EDIT DIRECTLY. Source: ~/.brain/rules/",
    "model": "anthropic/claude-sonnet-4-6",
    "systemPrompt": full_rules,
    "mcpServers": {
        "memory":              {"command": "bash", "args": ["-lc", "npx -y @modelcontextprotocol/server-memory \"$HOME/.brain/memory\""]},
        "filesystem":          {"command": "bash", "args": ["-lc", "npx -y @modelcontextprotocol/server-filesystem \"$HOME\""]},
        "sequential-thinking": {"command": "bash", "args": ["-lc", "npx -y @modelcontextprotocol/server-sequential-thinking"]},
        "context7":            {"command": "bash", "args": ["-lc", "npx -y @upstash/context7-mcp@latest"]},
        "github":              {"command": "bash", "args": ["-lc", "npx -y @modelcontextprotocol/server-github"]},
        "brain-rules":         {"command": "python3", "args": [str(brain_dir / "mcp" / "brain-mcp-server" / "server.py")]},
        "skill-ninja":         {"command": "docker", "args": ["run", "-i", "--rm", "aktsmm/skill-ninja-mcp-server:latest"]},
        "duckduckgo":          {"command": "docker", "args": ["run", "-i", "--rm", "mcp/duckduckgo:latest"]},
        "crawl4ai":            {"command": "docker", "args": ["run", "-i", "--rm", "unclecode/crawl4ai:latest"]},
        "context-awesome":     {"command": "docker", "args": ["run", "-i", "--rm", "bh-rat/context-awesome:latest"]},
    }
}

out_path = brain_dir / "adapters" / "opencode" / "opencode.json"
out_path.parent.mkdir(parents=True, exist_ok=True)
out_path.write_text(json.dumps(opencode, indent=2, ensure_ascii=False), encoding="utf-8")

# Self-validate
json.loads(out_path.read_text())
