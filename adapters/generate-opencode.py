#!/usr/bin/env python3
"""
generate-opencode.py - Generates opencode.json with proper JSON encoding.
Called by adapters/generate.sh.
"""
import json
import os
import pathlib
import subprocess
import subprocess

brain_dir = pathlib.Path(os.environ.get("BRAIN_DIR", str(pathlib.Path.home() / ".brain")))

result = subprocess.run(
    ["bash", str(brain_dir / "scripts" / "build-rules.sh"), "--stdout"],
    capture_output=True, text=True,
    env=dict(os.environ, BRAIN_DIR=str(brain_dir))
)
full_rules = result.stdout

opencode = {
    "model": "anthropic/claude-sonnet-4-6",
    "instructions": [full_rules],
    "mcp": {
        "memory":              {"type": "local", "command": ["bash", "-lc", "npx -y @modelcontextprotocol/server-memory \"$HOME/.brain/memory\""]},
        "filesystem":          {"type": "local", "command": ["bash", "-lc", "npx -y @modelcontextprotocol/server-filesystem \"$HOME\""]},
        "sequential-thinking": {"type": "local", "command": ["bash", "-lc", "npx -y @modelcontextprotocol/server-sequential-thinking"]},
        "context7":            {"type": "local", "command": ["bash", "-lc", "npx -y @upstash/context7-mcp@latest"]},
        "github":              {"type": "local", "command": ["bash", "-lc", "npx -y @modelcontextprotocol/server-github"]},
        "brain-rules":         {"type": "local", "command": ["python3", str(brain_dir / "mcp" / "brain-mcp-server" / "server.py")]},
        "skill-ninja":         {"type": "local", "command": ["docker", "run", "-i", "--rm", "aktsmm/skill-ninja-mcp-server:latest"]},
        "duckduckgo":          {"type": "local", "command": ["docker", "run", "-i", "--rm", "mcp/duckduckgo:latest"]},
        "crawl4ai":            {"type": "local", "command": ["docker", "run", "-i", "--rm", "unclecode/crawl4ai:latest"]},
        "context-awesome":     {"type": "local", "command": ["docker", "run", "-i", "--rm", "bh-rat/context-awesome:latest"]},
    }
}

out_path = brain_dir / "adapters" / "opencode" / "opencode.json"
out_path.parent.mkdir(parents=True, exist_ok=True)
out_path.write_text(json.dumps(opencode, indent=2, ensure_ascii=False), encoding="utf-8")

# Self-validate
json.loads(out_path.read_text())
