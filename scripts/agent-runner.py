#!/usr/bin/env python3
"""
agent-runner.py - LLM-agnostic executable agent runtime for brain repo.

Supports: Anthropic, OpenAI, Gemini, Ollama (auto-detected from env)
Reads: ~/.brain/brain.env for configuration
Reads: ~/.brain/providers/providers.yml for model routing

Usage:
  python3 agent-runner.py --list
  python3 agent-runner.py --agent researcher --task "Compare Qdrant vs Chroma"
  python3 agent-runner.py --agent planner --task "..." --memory
  python3 agent-runner.py --pipeline "orchestrator->planner->implementer" --task "..."
  python3 agent-runner.py --agent researcher --task "..." --backend openai
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
from dataclasses import dataclass


# ── Config loading ────────────────────────────────────────────────────────────

def get_brain_dir() -> pathlib.Path:
    env = os.environ.get("BRAIN_DIR")
    if env:
        return pathlib.Path(env)
    return pathlib.Path.home() / ".brain"


def load_brain_env(brain_dir: pathlib.Path) -> None:
    """Source brain.env into os.environ if it exists."""
    env_path = brain_dir / "brain.env"
    if not env_path.exists():
        return
    for line in env_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" in line:
            key, _, value = line.partition("=")
            key = key.strip()
            value = value.strip().strip('"').strip("'")
            if key and value and key not in os.environ:
                os.environ[key] = value


BRAIN_DIR = get_brain_dir()
load_brain_env(BRAIN_DIR)
AGENTS_DIR = BRAIN_DIR / "agents"


# ── Data types ────────────────────────────────────────────────────────────────

@dataclass
class AgentDef:
    name: str
    description: str
    system_prompt: str
    model_tier: str = "standard"


@dataclass
class AgentResult:
    agent: str
    task: str
    output: str
    model_used: str
    backend: str = ""
    tokens_used: int = 0
    error: str = ""


# ── Provider resolution ───────────────────────────────────────────────────────

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

DEFAULTS = {
    "anthropic": {"fast": "claude-haiku-4-5-20251001", "standard": "claude-sonnet-4-6",   "powerful": "claude-opus-4-6"},
    "openai":    {"fast": "gpt-4o-mini",                "standard": "gpt-4o",               "powerful": "gpt-4o"},
    "gemini":    {"fast": "gemini-2.5-flash",           "standard": "gemini-2.5-pro",       "powerful": "gemini-2.5-pro"},
    "ollama":    {"fast": "qwen2.5-coder:7b",           "standard": "qwen2.5-coder:32b",    "powerful": "deepseek-coder-v3:latest"},
}


def detect_backend(preferred: str = "auto") -> str:
    """Auto-detect available LLM backend from env vars."""
    if preferred != "auto":
        return preferred
    explicit = os.environ.get("BRAIN_LLM_BACKEND", "auto")
    if explicit != "auto":
        return explicit
    if os.environ.get("ANTHROPIC_API_KEY"):
        return "anthropic"
    if os.environ.get("OPENAI_API_KEY"):
        return "openai"
    if os.environ.get("GEMINI_API_KEY"):
        return "gemini"
    # Check Ollama
    try:
        urllib.request.urlopen(os.environ.get("OLLAMA_BASE_URL", "http://localhost:11434") + "/api/tags", timeout=2)
        return "ollama"
    except Exception:
        pass
    return "none"


def resolve_model(agent_name: str, backend: str) -> str:
    """Get the model string for an agent on a given backend."""
    tier = AGENT_TIERS.get(agent_name, "standard")
    providers_path = BRAIN_DIR / "providers" / "providers.yml"
    if providers_path.exists():
        content = providers_path.read_text(encoding="utf-8")
        # Find model for this backend+tier
        in_backend = False
        for line in content.splitlines():
            if re.match(rf"^\s*{re.escape(backend)}\s*:", line):
                in_backend = True
            if in_backend and f"{tier}:" in line:
                # Split on the tier key only to preserve model names with colons (e.g. "qwen2.5-coder:7b")
                candidate = line.split(f"{tier}:", 1)[-1].strip()
                if candidate and not candidate.startswith("#"):
                    return candidate
            if in_backend and re.match(r"^\w+:", line) and backend not in line:
                in_backend = False
    return DEFAULTS.get(backend, DEFAULTS["anthropic"]).get(tier, "claude-sonnet-4-6")


# ── LLM backends ─────────────────────────────────────────────────────────────

def _http_post(url: str, payload: dict, headers: dict, timeout: int = 120) -> dict:
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json", **headers}, method="POST")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode("utf-8"))


def call_anthropic(model: str, system: str, user: str) -> tuple[str, int]:
    api_key = os.environ.get("ANTHROPIC_API_KEY", "")
    if not api_key:
        raise RuntimeError("ANTHROPIC_API_KEY not set")
    data = _http_post(
        "https://api.anthropic.com/v1/messages",
        {"model": model, "max_tokens": 4096, "system": system, "messages": [{"role": "user", "content": user}]},
        {"x-api-key": api_key, "anthropic-version": "2023-06-01"},
    )
    text = "".join(b["text"] for b in data.get("content", []) if b.get("type") == "text")
    tokens = data.get("usage", {}).get("output_tokens", 0)
    return text, tokens


def call_openai(model: str, system: str, user: str) -> tuple[str, int]:
    api_key = os.environ.get("OPENAI_API_KEY", "")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY not set")
    base_url = os.environ.get("OPENAI_BASE_URL", "https://api.openai.com/v1")
    data = _http_post(
        f"{base_url}/chat/completions",
        {"model": model, "max_tokens": 4096, "messages": [{"role": "system", "content": system}, {"role": "user", "content": user}]},
        {"Authorization": f"Bearer {api_key}"},
    )
    text = data["choices"][0]["message"]["content"]
    tokens = data.get("usage", {}).get("completion_tokens", 0)
    return text, tokens


def call_gemini(model: str, system: str, user: str) -> tuple[str, int]:
    api_key = os.environ.get("GEMINI_API_KEY", "")
    if not api_key:
        raise RuntimeError("GEMINI_API_KEY not set")
    url = f"https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={api_key}"
    data = _http_post(url, {
        "system_instruction": {"parts": [{"text": system}]},
        "contents": [{"role": "user", "parts": [{"text": user}]}],
        "generationConfig": {"maxOutputTokens": 4096},
    }, {})
    text = data["candidates"][0]["content"]["parts"][0]["text"]
    tokens = data.get("usageMetadata", {}).get("candidatesTokenCount", 0)
    return text, tokens


def call_ollama(model: str, system: str, user: str) -> tuple[str, int]:
    base_url = os.environ.get("OLLAMA_BASE_URL", "http://localhost:11434")
    data = _http_post(
        f"{base_url}/api/chat",
        {"model": model, "stream": False, "messages": [{"role": "system", "content": system}, {"role": "user", "content": user}]},
        {},
    )
    text = data["message"]["content"]
    tokens = data.get("eval_count", 0)
    return text, tokens


BACKEND_CALLERS = {
    "anthropic": call_anthropic,
    "openai":    call_openai,
    "gemini":    call_gemini,
    "ollama":    call_ollama,
}


def call_model(backend: str, model: str, system: str, user: str) -> tuple[str, int]:
    caller = BACKEND_CALLERS.get(backend)
    if not caller:
        raise RuntimeError("No LLM backend available. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, or start Ollama.")
    return caller(model, system, user)


# ── Agent loading ─────────────────────────────────────────────────────────────

def load_agent(name: str) -> AgentDef:
    path = AGENTS_DIR / f"{name}.md"
    if not path.exists():
        available = sorted(p.stem for p in AGENTS_DIR.glob("*.md"))
        raise FileNotFoundError(f"Agent '{name}' not found. Available: {', '.join(available)}")
    raw = path.read_text(encoding="utf-8")
    description = ""
    if raw.startswith("---"):
        try:
            end = raw.index("---", 3)
            for line in raw[3:end].splitlines():
                if line.startswith("description:"):
                    description = line.split(":", 1)[1].strip()
            system_prompt = raw[end + 3:].strip()
        except ValueError:
            system_prompt = raw.strip()
    else:
        system_prompt = raw.strip()
    return AgentDef(name=name, description=description, system_prompt=system_prompt,
                    model_tier=AGENT_TIERS.get(name, "standard"))


# ── Memory injection ──────────────────────────────────────────────────────────

def fetch_memory_context(query: str) -> str:
    try:
        npx = next((c for c in ["npx-nvm", "npx"] if
                    subprocess.run(["which", c], capture_output=True).returncode == 0), None)
        if not npx:
            return ""
        mem_dir = str(BRAIN_DIR / "memory")
        proc = subprocess.run(
            [npx, "-y", "@modelcontextprotocol/server-memory", mem_dir],
            input="\n".join([
                '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"agent-runner","version":"1.0.0"}}}',
                '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}',
                json.dumps({"jsonrpc": "2.0", "id": 2, "method": "tools/call",
                            "params": {"name": "search_nodes", "arguments": {"query": query}}}),
            ]) + "\n",
            capture_output=True, text=True, timeout=15,
        )
        results = []
        for line in proc.stdout.splitlines():
            try:
                msg = json.loads(line)
                for block in msg.get("result", {}).get("content", []):
                    if block.get("type") == "text":
                        data = json.loads(block["text"])
                        for ent in data.get("entities", [])[:5]:
                            obs = " | ".join(ent.get("observations", [])[:2])
                            results.append(f"[{ent.get('entityType','?')}] {ent.get('name','')}: {obs}")
            except Exception:
                continue
        return ("## Memory Context\n" + "\n".join(f"- {r}" for r in results)) if results else ""
    except Exception:
        return ""


# ── Agent execution ───────────────────────────────────────────────────────────

def run_agent(name: str, task: str, context: str = "", inject_memory: bool = False,
              inject_rules: bool = True, backend_override: str = "auto") -> AgentResult:
    agent = load_agent(name)
    backend = detect_backend(backend_override)
    if backend == "none":
        return AgentResult(agent=name, task=task, output="", model_used="none", backend="none",
                           error="No LLM backend available. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, or start Ollama.")
    model = resolve_model(name, backend)

    system_parts = [agent.system_prompt]
    if inject_rules:
        canonical = BRAIN_DIR / "rules" / "canonical.md"
        if canonical.exists():
            system_parts.append("\n## Global Rules\n" + canonical.read_text(encoding="utf-8")[:2000])
    system = "\n\n".join(system_parts)

    user_parts = [f"## Task\n{task}"]
    if context:
        user_parts.insert(0, f"## Context\n{context}")
    if inject_memory:
        mem = fetch_memory_context(task[:200])
        if mem:
            user_parts.insert(0, mem)
    user = "\n\n".join(user_parts)

    try:
        output, tokens = call_model(backend, model, system, user)
        return AgentResult(agent=name, task=task, output=output, model_used=model, backend=backend, tokens_used=tokens)
    except Exception as e:
        return AgentResult(agent=name, task=task, output="", model_used=model, backend=backend, error=str(e))


def run_pipeline(pipeline: str, task: str, inject_memory: bool = False,
                 backend_override: str = "auto") -> list[AgentResult]:
    results = []
    context = ""
    for name in [a.strip() for a in pipeline.split("->")]:
        print(f"[pipeline] {name}...", file=sys.stderr)
        result = run_agent(name, task, context=context, inject_memory=inject_memory, backend_override=backend_override)
        results.append(result)
        if result.error:
            print(f"[pipeline] {name} failed: {result.error}", file=sys.stderr)
            break
        if context:
            context += f"\n\n---\n\nOutput from {name}:\n\n{result.output}"
        else:
            context = f"Output from {name}:\n\n{result.output}"
    return results


# ── CLI ───────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Brain repo agent runner (LLM-agnostic)")
    parser.add_argument("--agent",    help="Agent name to run")
    parser.add_argument("--pipeline", help="Pipeline: 'agent1->agent2->agent3'")
    parser.add_argument("--task",     help="Task description")
    parser.add_argument("--context",  help="Additional context")
    parser.add_argument("--memory",   action="store_true", help="Inject memory context")
    parser.add_argument("--no-rules", action="store_true", help="Skip canonical.md injection")
    parser.add_argument("--backend",  default="auto", choices=["auto","anthropic","openai","gemini","ollama"])
    parser.add_argument("--json",     action="store_true", help="JSON output")
    parser.add_argument("--list",     action="store_true", help="List available agents")
    parser.add_argument("--backends", action="store_true", help="Show available backends")
    args = parser.parse_args()

    if args.backends:
        active = detect_backend()
        info = {
            "active": active,
            "anthropic": bool(os.environ.get("ANTHROPIC_API_KEY")),
            "openai":    bool(os.environ.get("OPENAI_API_KEY")),
            "gemini":    bool(os.environ.get("GEMINI_API_KEY")),
            "ollama":    os.environ.get("OLLAMA_BASE_URL", "http://localhost:11434"),
        }
        print(json.dumps(info, indent=2))
        return

    if args.list:
        for p in sorted(AGENTS_DIR.glob("*.md")):
            try:
                a = load_agent(p.stem)
                tier = AGENT_TIERS.get(p.stem, "standard")
                print(f"  {p.stem:<24} [{tier}]  {a.description[:60]}")
            except Exception:
                print(f"  {p.stem}")
        return

    if not args.task:
        parser.error("--task is required")

    if args.pipeline:
        results = run_pipeline(args.pipeline, args.task, inject_memory=args.memory, backend_override=args.backend)
        if args.json:
            print(json.dumps([{"agent": r.agent, "output": r.output, "model": r.model_used,
                                "backend": r.backend, "tokens": r.tokens_used, "error": r.error}
                               for r in results], indent=2))
        else:
            for r in results:
                print(f"\n{'='*60}\nAgent: {r.agent} | Backend: {r.backend} | Model: {r.model_used} | Tokens: {r.tokens_used}\n{'='*60}")
                print(r.error if r.error else r.output)
        return

    if args.agent:
        result = run_agent(args.agent, args.task, context=args.context or "",
                           inject_memory=args.memory, inject_rules=not args.no_rules,
                           backend_override=args.backend)
        if args.json:
            print(json.dumps({"agent": result.agent, "output": result.output, "model": result.model_used,
                               "backend": result.backend, "tokens": result.tokens_used, "error": result.error}, indent=2))
        else:
            if result.error:
                print(f"ERROR [{result.backend}]: {result.error}", file=sys.stderr)
                sys.exit(1)
            print(result.output)
        return

    parser.print_help()


if __name__ == "__main__":
    main()
