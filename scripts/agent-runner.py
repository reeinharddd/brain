#!/usr/bin/env python3
"""
agent-runner.py - Lightweight agent execution runtime for brain repo.

Transforms the static markdown agent definitions into actual executable agents
that can call models, inject memory, and optionally pipe output to other agents.

Architecture:
  load_agent(name) -> AgentDef
  run_agent(name, task, context) -> AgentResult
  pipe(source_agent, target_agent, task) -> AgentResult

Usage:
  python3 agent-runner.py --agent researcher --task "What is LangGraph?"
  python3 agent-runner.py --agent planner --task "Build auth system" --memory
  python3 agent-runner.py --pipeline "orchestrator->planner->implementer" --task "..."
  python3 agent-runner.py --list
"""

import argparse
import json
import os
import pathlib
import re
import subprocess
import sys
import urllib.error
import urllib.request
from dataclasses import dataclass, field


BRAIN_DIR = pathlib.Path(os.environ.get("BRAIN_DIR", str(pathlib.Path.home() / ".brain")))
AGENTS_DIR = BRAIN_DIR / "agents"
PROVIDERS_PATH = BRAIN_DIR / "providers" / "providers.yml"


# ── Data structures ─────────────────────────────────────────────────────────

@dataclass
class AgentDef:
    name: str
    description: str
    system_prompt: str
    model_tier: str = "standard"  # fast | standard | powerful


@dataclass
class AgentResult:
    agent: str
    task: str
    output: str
    model_used: str
    tokens_used: int = 0
    error: str = ""
    metadata: dict = field(default_factory=dict)


# ── Provider resolution ──────────────────────────────────────────────────────

def load_providers() -> dict:
    """Parse providers.yml for model routing without a YAML library."""
    if not PROVIDERS_PATH.exists():
        return {}
    raw = PROVIDERS_PATH.read_text(encoding="utf-8")
    # Simple extraction: find fast/standard/powerful lines under 'claude:'
    providers: dict = {}
    current_provider = None
    for line in raw.splitlines():
        if re.match(r"^\w+:", line) and not line.startswith(" "):
            current_provider = line.strip().rstrip(":")
        if current_provider == "claude" and "fast:" in line:
            providers["fast"] = line.split(":")[1].strip()
        if current_provider == "claude" and "standard:" in line:
            providers["standard"] = line.split(":")[1].strip()
        if current_provider == "claude" and "powerful:" in line:
            providers["powerful"] = line.split(":")[1].strip()
    return providers


PROVIDER_MAP = {
    "planning":        "powerful",
    "architecture":    "powerful",
    "design":          "powerful",
    "review":          "standard",
    "implementation":  "standard",
    "debugging":       "standard",
    "documentation":   "fast",
    "summarization":   "fast",
    "research":        "standard",
    "security":        "standard",
    "refactor":        "standard",
}

AGENT_TIERS = {
    "orchestrator": "powerful",
    "planner":      "powerful",
    "architect":    "powerful",
    "researcher":   "standard",
    "implementer":  "standard",
    "reviewer":     "standard",
    "debugger":     "standard",
    "refactor":     "standard",
    "designer":     "powerful",
    "documenter":   "fast",
    "guardian":     "standard",
    "configurator": "fast",
}


def resolve_model(agent_name: str, providers: dict) -> str:
    tier = AGENT_TIERS.get(agent_name, "standard")
    model = providers.get(tier)
    if model:
        return model
    # Sensible defaults if providers.yml not parseable
    defaults = {
        "fast": "claude-haiku-4-5-20251001",
        "standard": "claude-sonnet-4-6",
        "powerful": "claude-opus-4-6",
    }
    return defaults.get(tier, defaults["standard"])


# ── Agent loading ─────────────────────────────────────────────────────────────

def load_agent(name: str) -> AgentDef:
    agent_path = AGENTS_DIR / f"{name}.md"
    if not agent_path.exists():
        available = [p.stem for p in AGENTS_DIR.glob("*.md")]
        raise FileNotFoundError(
            f"Agent '{name}' not found. Available: {', '.join(sorted(available))}"
        )

    raw = agent_path.read_text(encoding="utf-8")

    # Parse YAML frontmatter
    description = ""
    if raw.startswith("---"):
        end = raw.index("---", 3)
        frontmatter = raw[3:end]
        for line in frontmatter.splitlines():
            if line.startswith("description:"):
                description = line.split(":", 1)[1].strip()
        system_prompt = raw[end + 3:].strip()
    else:
        system_prompt = raw.strip()

    providers = load_providers()
    model_tier = AGENT_TIERS.get(name, "standard")

    return AgentDef(
        name=name,
        description=description,
        system_prompt=system_prompt,
        model_tier=model_tier,
    )


# ── Memory injection ──────────────────────────────────────────────────────────

def fetch_memory_context(query: str) -> str:
    """Query MCP memory server for relevant context via stdio."""
    try:
        npx = _resolve_npx()
        mem_dir = str(BRAIN_DIR / "memory")
        proc = subprocess.run(
            [npx, "-y", "@modelcontextprotocol/server-memory", mem_dir],
            input="\n".join([
                '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"agent-runner","version":"1.0.0"}}}',
                '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}',
                json.dumps({"jsonrpc": "2.0", "id": 2, "method": "tools/call",
                            "params": {"name": "search_nodes", "arguments": {"query": query}}}),
            ]) + "\n",
            capture_output=True,
            text=True,
            timeout=15,
        )
        # Parse memory results
        results = []
        for line in proc.stdout.splitlines():
            try:
                msg = json.loads(line)
                content = msg.get("result", {}).get("content", [])
                for block in content:
                    if block.get("type") == "text":
                        data = json.loads(block["text"])
                        for ent in data.get("entities", []):
                            name = ent.get("name", "")
                            etype = ent.get("entityType", "")
                            obs = ent.get("observations", [])
                            results.append(f"[{etype}] {name}: " + " | ".join(obs[:2]))
            except (json.JSONDecodeError, KeyError, TypeError):
                continue

        if results:
            return "## Relevant Memory Context\n" + "\n".join(f"- {r}" for r in results[:10])
        return ""
    except Exception as e:
        return f"[memory unavailable: {e}]"


def _resolve_npx() -> str:
    for cmd in ["npx-nvm", "npx"]:
        result = subprocess.run(["which", cmd], capture_output=True, text=True)
        if result.returncode == 0:
            return cmd
    raise RuntimeError("npx not found")


# ── Model call ────────────────────────────────────────────────────────────────

def call_model(model: str, system: str, user: str) -> tuple[str, int]:
    """Call Anthropic API directly."""
    api_key = os.environ.get("ANTHROPIC_API_KEY", "")
    if not api_key:
        raise RuntimeError(
            "ANTHROPIC_API_KEY not set. Export it in your shell profile."
        )

    payload = json.dumps({
        "model": model,
        "max_tokens": 4096,
        "system": system,
        "messages": [{"role": "user", "content": user}],
    }).encode("utf-8")

    req = urllib.request.Request(
        "https://api.anthropic.com/v1/messages",
        data=payload,
        headers={
            "x-api-key": api_key,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json",
        },
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = json.loads(resp.read().decode("utf-8"))

    text = "".join(block["text"] for block in data.get("content", []) if block.get("type") == "text")
    tokens = data.get("usage", {}).get("output_tokens", 0)
    return text, tokens


# ── Agent execution ────────────────────────────────────────────────────────────

def run_agent(
    name: str,
    task: str,
    context: str = "",
    inject_memory: bool = False,
    inject_rules: bool = True,
) -> AgentResult:
    agent = load_agent(name)
    providers = load_providers()
    model = resolve_model(name, providers)

    # Build system prompt
    system_parts = [agent.system_prompt]

    # Inject global rules (canonical.md excerpt - first 2000 chars to stay within budget)
    if inject_rules:
        canonical_path = BRAIN_DIR / "rules" / "canonical.md"
        if canonical_path.exists():
            rules_excerpt = canonical_path.read_text(encoding="utf-8")[:2000]
            system_parts.append(f"\n\n## Global Rules (from canonical.md)\n{rules_excerpt}")

    system = "\n\n".join(system_parts)

    # Build user message
    user_parts = [f"## Task\n{task}"]
    if context:
        user_parts.insert(0, f"## Context\n{context}")
    if inject_memory:
        mem_ctx = fetch_memory_context(task[:200])
        if mem_ctx:
            user_parts.insert(0, mem_ctx)

    user = "\n\n".join(user_parts)

    try:
        output, tokens = call_model(model, system, user)
        return AgentResult(
            agent=name,
            task=task,
            output=output,
            model_used=model,
            tokens_used=tokens,
        )
    except Exception as e:
        return AgentResult(
            agent=name,
            task=task,
            output="",
            model_used=model,
            error=str(e),
        )


# ── Pipeline execution ────────────────────────────────────────────────────────

def run_pipeline(pipeline: str, task: str, inject_memory: bool = False) -> list[AgentResult]:
    """Run a pipeline of agents where output of each feeds into the next."""
    agent_names = [a.strip() for a in pipeline.split("->")]
    results = []
    context = ""

    for agent_name in agent_names:
        print(f"[pipeline] Running agent: {agent_name}", file=sys.stderr)
        result = run_agent(agent_name, task, context=context, inject_memory=inject_memory)
        results.append(result)
        if result.error:
            print(f"[pipeline] Agent {agent_name} failed: {result.error}", file=sys.stderr)
            break
        # Pass output as context to next agent
        context = f"Output from {agent_name}:\n\n{result.output}"

    return results


# ── CLI ───────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Brain repo agent runner")
    parser.add_argument("--agent",    help="Agent name to run")
    parser.add_argument("--pipeline", help="Pipeline: 'agent1->agent2->agent3'")
    parser.add_argument("--task",     help="Task description")
    parser.add_argument("--context",  help="Additional context string")
    parser.add_argument("--memory",   action="store_true", help="Inject relevant memory context")
    parser.add_argument("--no-rules", action="store_true", help="Skip injecting canonical.md rules")
    parser.add_argument("--json",     action="store_true", help="Output JSON")
    parser.add_argument("--list",     action="store_true", help="List available agents")
    args = parser.parse_args()

    if args.list:
        available = sorted(p.stem for p in AGENTS_DIR.glob("*.md"))
        for name in available:
            try:
                agent = load_agent(name)
                tier = AGENT_TIERS.get(name, "standard")
                print(f"  {name:<20} [{tier}]  {agent.description[:60]}")
            except Exception:
                print(f"  {name}")
        return

    if not args.task:
        parser.error("--task is required")

    if args.pipeline:
        results = run_pipeline(args.pipeline, args.task, inject_memory=args.memory)
        if args.json:
            print(json.dumps([
                {"agent": r.agent, "output": r.output, "model": r.model_used,
                 "tokens": r.tokens_used, "error": r.error}
                for r in results
            ], indent=2))
        else:
            for r in results:
                print(f"\n{'='*60}")
                print(f"Agent: {r.agent} | Model: {r.model_used} | Tokens: {r.tokens_used}")
                print(f"{'='*60}")
                if r.error:
                    print(f"ERROR: {r.error}")
                else:
                    print(r.output)
        return

    if args.agent:
        result = run_agent(
            args.agent,
            args.task,
            context=args.context or "",
            inject_memory=args.memory,
            inject_rules=not args.no_rules,
        )
        if args.json:
            print(json.dumps({
                "agent":   result.agent,
                "output":  result.output,
                "model":   result.model_used,
                "tokens":  result.tokens_used,
                "error":   result.error,
            }, indent=2))
        else:
            if result.error:
                print(f"ERROR: {result.error}", file=sys.stderr)
                sys.exit(1)
            print(result.output)
        return

    parser.print_help()


if __name__ == "__main__":
    main()
