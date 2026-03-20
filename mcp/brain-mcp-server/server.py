#!/usr/bin/env python3
"""
brain-mcp-server/server.py
Custom MCP server that exposes brain repo internals as queryable tools.

Tools exposed:
  brain_get_rules          - Get relevant rules from canonical.md by topic
  brain_get_agent          - Get agent definition by name
  brain_list_agents        - List all available agents
  brain_get_command        - Get slash command definition
  brain_route_task         - Route a task to the appropriate agent/model tier
  brain_search_rules       - Full-text search across all rules and modules
  brain_get_provider       - Get model for a given task type

Protocol: MCP stdio (JSON-RPC 2.0)
Usage: python3 server.py [--brain-dir PATH]
"""

import argparse
import json
import pathlib
import re
import sys
import os


# ── Brain dir resolution ──────────────────────────────────────────────────────

def get_brain_dir(override: str | None = None) -> pathlib.Path:
    if override:
        return pathlib.Path(override)
    env = os.environ.get("BRAIN_DIR")
    if env:
        return pathlib.Path(env)
    return pathlib.Path.home() / ".brain"


# ── Tool implementations ──────────────────────────────────────────────────────

def brain_get_rules(brain_dir: pathlib.Path, topic: str, max_chars: int = 2000) -> str:
    """Return sections of canonical.md and modules relevant to topic."""
    results = []
    all_sources = [brain_dir / "rules" / "canonical.md"] + sorted((brain_dir / "rules" / "modules").glob("*.md"))

    for path in all_sources:
        if not path.exists():
            continue
        content = path.read_text(encoding="utf-8")
        # Split into sections and filter by relevance
        sections = re.split(r'\n(?=#{1,3} )', content)
        for section in sections:
            if topic.lower() in section.lower():
                results.append(f"[{path.stem}]\n{section.strip()}")

    if not results:
        return f"No rules found for topic: {topic}"

    combined = "\n\n---\n\n".join(results)
    return combined[:max_chars] if len(combined) > max_chars else combined


def brain_get_agent(brain_dir: pathlib.Path, name: str) -> str:
    path = brain_dir / "agents" / f"{name}.md"
    if not path.exists():
        available = [p.stem for p in (brain_dir / "agents").glob("*.md")]
        return f"Agent '{name}' not found. Available: {', '.join(sorted(available))}"
    return path.read_text(encoding="utf-8")


def brain_list_agents(brain_dir: pathlib.Path) -> str:
    agents_dir = brain_dir / "agents"
    if not agents_dir.exists():
        return "agents/ directory not found"
    agents = []
    for path in sorted(agents_dir.glob("*.md")):
        content = path.read_text(encoding="utf-8")
        desc = ""
        if content.startswith("---"):
            try:
                end = content.index("---", 3)
                for line in content[3:end].splitlines():
                    if line.startswith("description:"):
                        desc = line.split(":", 1)[1].strip()
            except ValueError:
                pass
        agents.append(f"- {path.stem}: {desc}" if desc else f"- {path.stem}")
    return "\n".join(agents)


def brain_get_command(brain_dir: pathlib.Path, name: str) -> str:
    path = brain_dir / "commands" / f"{name}.md"
    if not path.exists():
        available = [p.stem for p in (brain_dir / "commands").glob("*.md")]
        return f"Command '{name}' not found. Available: {', '.join(sorted(available))}"
    return path.read_text(encoding="utf-8")


def brain_route_task(brain_dir: pathlib.Path, task_description: str) -> str:
    """Suggest the right agent and model tier for a given task."""
    td = task_description.lower()

    routing = [
        (["plan", "architect", "design", "spec", "roadmap"],         "planner",      "powerful"),
        (["research", "investigate", "find", "look up"],              "researcher",   "standard"),
        (["implement", "build", "code", "write", "create"],          "implementer",  "standard"),
        (["debug", "fix", "error", "bug", "broken"],                  "debugger",     "standard"),
        (["review", "check", "audit", "approve"],                     "reviewer",     "standard"),
        (["refactor", "improve", "clean", "restructure"],             "refactor",     "standard"),
        (["document", "readme", "comment", "adr"],                    "documenter",   "fast"),
        (["security", "vulnerability", "secret", "unsafe"],          "guardian",     "standard"),
        (["ui", "design", "component", "style", "visual"],           "designer",     "powerful"),
        (["orchestrate", "coordinate", "delegate", "complex"],       "orchestrator", "powerful"),
    ]

    for keywords, agent, tier in routing:
        if any(kw in td for kw in keywords):
            return json.dumps({"suggested_agent": agent, "model_tier": tier,
                               "reasoning": f"Task contains keywords matching {agent} profile"})

    return json.dumps({"suggested_agent": "orchestrator", "model_tier": "standard",
                       "reasoning": "No specific pattern matched; defaulting to orchestrator"})


def brain_search_rules(brain_dir: pathlib.Path, query: str) -> str:
    """Full-text search across canonical.md and all modules."""
    results = []
    query_terms = query.lower().split()
    all_sources = [brain_dir / "rules" / "canonical.md"] + sorted((brain_dir / "rules" / "modules").glob("*.md"))

    for path in all_sources:
        if not path.exists():
            continue
        content = path.read_text(encoding="utf-8")
        lines = content.splitlines()
        for i, line in enumerate(lines):
            if all(term in line.lower() for term in query_terms):
                context_start = max(0, i - 1)
                context_end = min(len(lines), i + 3)
                snippet = "\n".join(lines[context_start:context_end])
                results.append(f"[{path.stem}:{i+1}] {snippet}")

    if not results:
        return f"No matches for: {query}"
    return "\n\n".join(results[:10])  # Top 10 matches


def brain_get_provider(brain_dir: pathlib.Path, task_type: str) -> str:
    """Get the recommended model for a task type from providers.yml."""
    providers_path = brain_dir / "providers" / "providers.yml"
    if not providers_path.exists():
        return "providers.yml not found"

    content = providers_path.read_text(encoding="utf-8")

    # Extract task_routing section
    routing_match = re.search(r'task_routing:(.*?)(?=\n\w|\Z)', content, re.DOTALL)
    if routing_match:
        routing_text = routing_match.group(1)
        task_type_lower = task_type.lower()
        for line in routing_text.splitlines():
            if task_type_lower in line.lower() and ":" in line:
                tier = line.split(":")[-1].strip().split("#")[0].strip()
                # Find model for this tier
                model_match = re.search(rf'{tier}:\s*(\S+)', content)
                if model_match:
                    return json.dumps({"task_type": task_type, "tier": tier, "model": model_match.group(1)})

    return json.dumps({"task_type": task_type, "tier": "standard", "model": "claude-sonnet-4-6"})


# ── MCP Protocol handler ──────────────────────────────────────────────────────

TOOLS = [
    {
        "name": "brain_get_rules",
        "description": "Get relevant rules from canonical.md and modules by topic",
        "inputSchema": {
            "type": "object",
            "properties": {
                "topic": {"type": "string", "description": "Topic to search rules for (e.g. 'error handling', 'security', 'git')"},
                "max_chars": {"type": "integer", "description": "Max characters to return (default 2000)", "default": 2000},
            },
            "required": ["topic"],
        },
    },
    {
        "name": "brain_get_agent",
        "description": "Get the full definition of a brain repo agent by name",
        "inputSchema": {
            "type": "object",
            "properties": {
                "name": {"type": "string", "description": "Agent name (e.g. orchestrator, researcher, planner)"},
            },
            "required": ["name"],
        },
    },
    {
        "name": "brain_list_agents",
        "description": "List all available brain repo agents with their descriptions",
        "inputSchema": {"type": "object", "properties": {}},
    },
    {
        "name": "brain_get_command",
        "description": "Get the definition of a brain repo slash command",
        "inputSchema": {
            "type": "object",
            "properties": {
                "name": {"type": "string", "description": "Command name without slash (e.g. plan, review, handover)"},
            },
            "required": ["name"],
        },
    },
    {
        "name": "brain_route_task",
        "description": "Get the suggested agent and model tier for a task description",
        "inputSchema": {
            "type": "object",
            "properties": {
                "task_description": {"type": "string", "description": "Description of the task to route"},
            },
            "required": ["task_description"],
        },
    },
    {
        "name": "brain_search_rules",
        "description": "Full-text search across all brain repo rules and modules",
        "inputSchema": {
            "type": "object",
            "properties": {
                "query": {"type": "string", "description": "Search query"},
            },
            "required": ["query"],
        },
    },
    {
        "name": "brain_get_provider",
        "description": "Get recommended model for a specific task type based on providers.yml routing",
        "inputSchema": {
            "type": "object",
            "properties": {
                "task_type": {"type": "string", "description": "Task type (e.g. planning, implementation, debugging)"},
            },
            "required": ["task_type"],
        },
    },
]


def handle_request(request: dict, brain_dir: pathlib.Path) -> dict:
    method = request.get("method", "")
    req_id = request.get("id")
    params = request.get("params", {})

    def ok(result):
        return {"jsonrpc": "2.0", "id": req_id, "result": result}

    def err(code, message):
        return {"jsonrpc": "2.0", "id": req_id, "error": {"code": code, "message": message}}

    if method == "initialize":
        return ok({
            "protocolVersion": "2024-11-05",
            "capabilities": {"tools": {}},
            "serverInfo": {"name": "brain-mcp-server", "version": "1.0.0"},
        })

    if method == "notifications/initialized":
        return None  # No response needed for notifications

    if method == "tools/list":
        return ok({"tools": TOOLS})

    if method == "tools/call":
        tool_name = params.get("name", "")
        args = params.get("arguments", {})

        try:
            if tool_name == "brain_get_rules":
                result = brain_get_rules(brain_dir, args["topic"], args.get("max_chars", 2000))
            elif tool_name == "brain_get_agent":
                result = brain_get_agent(brain_dir, args["name"])
            elif tool_name == "brain_list_agents":
                result = brain_list_agents(brain_dir)
            elif tool_name == "brain_get_command":
                result = brain_get_command(brain_dir, args["name"])
            elif tool_name == "brain_route_task":
                result = brain_route_task(brain_dir, args["task_description"])
            elif tool_name == "brain_search_rules":
                result = brain_search_rules(brain_dir, args["query"])
            elif tool_name == "brain_get_provider":
                result = brain_get_provider(brain_dir, args["task_type"])
            else:
                return err(-32601, f"Unknown tool: {tool_name}")

            return ok({"content": [{"type": "text", "text": result}]})
        except Exception as e:
            return err(-32603, str(e))

    return err(-32601, f"Unknown method: {method}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--brain-dir", help="Path to brain directory")
    args = parser.parse_args()
    brain_dir = get_brain_dir(args.brain_dir)

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            request = json.loads(line)
        except json.JSONDecodeError:
            continue

        response = handle_request(request, brain_dir)
        if response is not None:
            sys.stdout.write(json.dumps(response) + "\n")
            sys.stdout.flush()


if __name__ == "__main__":
    main()
