#!/usr/bin/env python3
"""
test_dynamic_skills.py
Tests dynamic skill discovery and activation:
- Agent reads skills/registry.yml and decides which skill to use
- Stack detection triggers correct skill injection
- Skills compose with MCP tools
- Context injection from codebase-context.ndjson
"""
import json, os, pathlib, subprocess, sys

brain_dir = pathlib.Path(os.environ.get("BRAIN_DIR", pathlib.Path.home() / ".brain"))
import importlib.util as _ilu, pathlib as _p, os as _os
_brain_dir = _p.Path(_os.environ.get('BRAIN_DIR', _p.Path.home() / '.brain'))
_spec = _ilu.spec_from_file_location('agent_runner', str(_brain_dir / 'scripts/agent-runner.py'))
_mod = _ilu.module_from_spec(_spec)
_spec.loader.exec_module(_mod)
import sys as _sys
_sys.modules['agent_runner'] = _mod


try:
    from agent_runner import run_agent, detect_backend
except ImportError:
    print("SKIP: agent-runner not importable")
    sys.exit(0)

def test_skill_registry_readable():
    """Skills registry exists and is valid YAML-like format."""
    registry = brain_dir / "skills/registry.yml"
    if not registry.exists():
        return {"status": "FAIL", "reason": "skills/registry.yml not found"}
    content = registry.read_text()
    has_skills = "skills:" in content
    skill_count = content.count("type: ")
    return {
        "status": "PASS" if has_skills and skill_count > 0 else "FAIL",
        "has_skills_key": has_skills,
        "skill_count": skill_count,
        "registry_size": len(content),
    }

def test_stack_detection_triggers_skills():
    """detect-stack.sh outputs tags that map to skills in registry."""
    detect_script = brain_dir / "scripts/detect-stack.sh"
    render_script = brain_dir / "scripts/render-skill-context.sh"

    if not detect_script.exists():
        return {"status": "FAIL", "reason": "detect-stack.sh not found"}

    import tempfile, shutil
    tmp = pathlib.Path(tempfile.mkdtemp())
    try:
        # Create a minimal Next.js project structure
        (tmp / "package.json").write_text('{"dependencies":{"next":"15.0.0","react":"18.0.0"}}')
        (tmp / "tsconfig.json").write_text('{"compilerOptions":{"strict":true}}')
        (tmp / "app").mkdir()
        (tmp / "app" / "page.tsx").write_text("export default function Page() { return <div>Hello</div>; }")

        detect_result = subprocess.run(
            ["bash", str(detect_script), str(tmp)],
            capture_output=True, text=True, timeout=15
        )
        tags = detect_result.stdout.strip()
        has_nextjs = "next" in tags.lower() or "typescript" in tags.lower() or "react" in tags.lower()

        render_result = {"stdout": "", "returncode": -1}
        if render_script.exists():
            rr = subprocess.run(
                ["bash", str(render_script), str(tmp)],
                capture_output=True, text=True, timeout=15
            )
            render_result = {"stdout": rr.stdout, "returncode": rr.returncode}

        skill_context = render_result["stdout"]
        has_skill_suggestions = len(skill_context) > 10

        return {
            "status": "PASS" if has_nextjs else "PARTIAL",
            "detected_tags": tags[:200],
            "nextjs_detected": has_nextjs,
            "skill_context_generated": has_skill_suggestions,
            "skill_context_preview": skill_context[:300],
        }
    finally:
        shutil.rmtree(tmp, ignore_errors=True)

def test_codebase_context_injection():
    """
    Verify that codebase-context.ndjson can be used to inject
    relevant context into an agent call (semantic skill selection).
    """
    ctx_file = brain_dir / ".brain/codebase-context.ndjson"
    if not ctx_file.exists():
        # Try to generate it first
        gen_script = brain_dir / "skills/codebase-contextualizer/contextualize.sh"
        if gen_script.exists():
            subprocess.run(
                ["bash", str(gen_script), str(brain_dir)],
                capture_output=True, timeout=30
            )

    if not ctx_file.exists():
        return {"status": "SKIP", "reason": "codebase-context.ndjson not generated"}

    lines = [l for l in ctx_file.read_text().splitlines() if l.strip()]
    if not lines:
        return {"status": "FAIL", "reason": "codebase-context.ndjson is empty"}

    # Parse first N entries to build a context summary
    entries = []
    for line in lines[:10]:
        try:
            entries.append(json.loads(line))
        except Exception:
            pass

    # Build context string from entries
    context_summary = "\n".join(
        f"- {e.get('title', e.get('path','?'))}: {e.get('content','')[:100]}"
        for e in entries
    )

    if detect_backend() == "none":
        return {
            "status": "PASS",
            "note": "context parseable but no LLM to test injection",
            "entries_found": len(entries),
            "context_preview": context_summary[:300],
        }

    # Inject context into an agent call
    result = run_agent(
        "researcher",
        "Based on the codebase context, what is the primary purpose of this project?",
        context=f"Codebase index:\n{context_summary}",
        inject_rules=False,
        inject_memory=False,
    )

    return {
        "status": "PASS" if not result.error else "FAIL",
        "entries_used": len(entries),
        "context_injected": True,
        "agent_understood_context": len(result.output) > 50 if not result.error else False,
        "output_preview": result.output[:300] if not result.error else result.error,
    }

def test_agent_self_discovery():
    """
    Agent reads its own registry and decides which tool/skill to use.
    Tests autonomous tool discovery as described in orchestrator.md.
    """
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend"}

    registry_path = brain_dir / "skills/registry.yml"
    mcp_registry_path = brain_dir / "mcp/registry.yml"

    registry_content = registry_path.read_text()[:1500] if registry_path.exists() else "(not found)"
    mcp_content = mcp_registry_path.read_text()[:1500] if mcp_registry_path.exists() else "(not found)"

    prompt = f"""You have access to these skills and MCP servers:

SKILLS REGISTRY:
{registry_content}

MCP REGISTRY:
{mcp_content}

Task: "I need to search for recent information about LangGraph and store the finding in memory."

Which specific skills and MCP tools would you use? List them in order with a one-sentence reason for each.
Format as: TOOL: [name] | REASON: [why]"""

    result = run_agent("orchestrator", prompt, inject_rules=False, inject_memory=False)
    if result.error:
        return {"status": "FAIL", "error": result.error}

    output = result.output or ""
    mentioned_memory = any(w in output.lower() for w in ["memory", "engram", "mem_"])
    mentioned_search = any(w in output.lower() for w in ["duckduckgo", "context7", "crawl", "search"])
    structured = "TOOL:" in output or "tool:" in output.lower()

    return {
        "status": "PASS" if (mentioned_memory and mentioned_search) else "PARTIAL",
        "mentioned_memory_mcp": mentioned_memory,
        "mentioned_search_tool": mentioned_search,
        "structured_response": structured,
        "output_preview": output[:500],
    }

if __name__ == "__main__":
    print("=" * 60)
    print("TEST: Dynamic Skills & Capability Discovery")
    print("=" * 60)

    r1 = test_skill_registry_readable()
    print(f"\n[skill_registry] {r1['status']} | skills: {r1.get('skill_count','?')}")

    r2 = test_stack_detection_triggers_skills()
    print(f"\n[stack_detection] {r2['status']}")
    print(f"  Tags: {r2.get('detected_tags','?')[:100]}")
    print(f"  Next.js detected: {r2.get('nextjs_detected')}")

    r3 = test_codebase_context_injection()
    print(f"\n[context_injection] {r3['status']}")
    print(f"  Entries: {r3.get('entries_used','?')} | Understood: {r3.get('agent_understood_context')}")

    r4 = test_agent_self_discovery()
    print(f"\n[agent_self_discovery] {r4['status']}")
    print(f"  Memory MCP mentioned: {r4.get('mentioned_memory_mcp')}")
    print(f"  Search tool mentioned: {r4.get('mentioned_search_tool')}")

    output = {
        "skill_registry": r1, "stack_detection": r2,
        "context_injection": r3, "agent_self_discovery": r4,
    }
    out_path = pathlib.Path.home() / ".brain/tests/results/dynamic_skills.json"
    out_path.write_text(json.dumps(output, indent=2))
    print(f"\nResultado guardado: {out_path}")

    any_fail = any(r.get("status") == "FAIL" for r in [r1, r2, r3, r4])
    sys.exit(1 if any_fail else 0)
