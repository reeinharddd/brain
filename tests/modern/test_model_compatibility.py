#!/usr/bin/env python3
"""
test_model_compatibility.py
Tests that the brain system works correctly across different model tiers and backends.
- Fast tier: minimal context, quick responses
- Standard tier: normal context, full delegation
- Powerful tier: full rules injection, complex reasoning
- Small model injection: testing capability injection for weaker models
"""
import json, os, pathlib, sys, time

brain_dir = pathlib.Path(os.environ.get("BRAIN_DIR", pathlib.Path.home() / ".brain"))
import importlib.util as _ilu, pathlib as _p, os as _os
_brain_dir = _p.Path(_os.environ.get('BRAIN_DIR', _p.Path.home() / '.brain'))
_spec = _ilu.spec_from_file_location('agent_runner', str(_brain_dir / 'scripts/agent-runner.py'))
_mod = _ilu.module_from_spec(_spec)
_spec.loader.exec_module(_mod)
import sys as _sys
_sys.modules['agent_runner'] = _mod


try:
    from agent_runner import run_agent, detect_backend, resolve_model, AGENT_TIERS
except ImportError:
    print("SKIP: agent-runner not importable")
    sys.exit(0)

def load_providers():
    """Load providers config if available."""
    p = brain_dir / "providers/providers.yml"
    return p.read_text() if p.exists() else ""

def test_tier_routing():
    """Verify each agent tier maps to the correct model."""
    tier_map = {
        "documenter": "fast",
        "researcher": "standard",
        "planner": "powerful",
        "orchestrator": "powerful",
    }
    backend = detect_backend()

    results = []
    for agent, expected_tier in tier_map.items():
        model = resolve_model(agent, backend)
        actual_tier = AGENT_TIERS.get(agent, "standard")
        results.append({
            "agent": agent,
            "expected_tier": expected_tier,
            "actual_tier": actual_tier,
            "model": model,
            "backend": backend,
            "tier_match": actual_tier == expected_tier,
        })

    all_match = all(r["tier_match"] for r in results)
    return {
        "status": "PASS" if all_match else "FAIL",
        "backend": backend,
        "routing": results,
    }

def test_capability_injection_for_small_models():
    """
    Tests that modern capabilities can be injected into agents
    that would otherwise be limited (e.g., small models via Ollama).
    Strategy: prepend capability context to system prompt.
    """
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no backend"}

    # Inject a structured reasoning capability via context
    capability_injection = """
You have been enhanced with structured reasoning capability.
When answering, follow this protocol:
THINK: [brief internal reasoning]
ANSWER: [your actual response]
CONFIDENCE: [high/medium/low]
"""

    result = run_agent(
        "documenter",
        "Is SHA-256 suitable for password hashing? Why or why not?",
        context=capability_injection,
        inject_rules=False,
        inject_memory=False,
    )

    if result.error:
        return {"status": "FAIL", "error": result.error}

    output = result.output or ""
    has_think      = "THINK:" in output
    has_answer     = "ANSWER:" in output
    has_confidence = "CONFIDENCE:" in output
    adopted_protocol = sum([has_think, has_answer, has_confidence])

    return {
        "status": "PASS" if adopted_protocol >= 2 else "PARTIAL",
        "protocol_fields_adopted": adopted_protocol,
        "has_think": has_think,
        "has_answer": has_answer,
        "has_confidence": has_confidence,
        "model_used": result.model_used,
        "output_preview": output[:400],
        "note": "Capability injection allows small models to mimic structured output",
    }

def test_chain_of_thought_injection():
    """Inject chain-of-thought prompting to improve small model reasoning."""
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no backend"}

    cot_prefix = "Let's think step by step before answering."
    task = f"{cot_prefix}\n\nA function runs in O(n^2). If n=1000, and we improve it to O(n log n), how much faster is it approximately?"

    result = run_agent("documenter", task, inject_rules=False, inject_memory=False)
    if result.error:
        return {"status": "FAIL", "error": result.error}

    output = result.output or ""
    shows_steps = any(w in output.lower() for w in ["step", "first", "then", "therefore", "so"])
    has_number  = any(c.isdigit() for c in output)

    return {
        "status": "PASS" if shows_steps and has_number else "PARTIAL",
        "shows_reasoning_steps": shows_steps,
        "includes_number": has_number,
        "model_used": result.model_used,
        "output_preview": output[:400],
    }

def test_provider_fallback_simulation():
    """
    Simulate what happens when primary model is unavailable.
    Tests fallback_chain in providers.yml is readable and ordered.
    """
    import re
    providers_path = brain_dir / "providers/providers.yml"
    if not providers_path.exists():
        return {"status": "FAIL", "reason": "providers.yml not found"}

    content = providers_path.read_text()
    fallback_match = re.search(r'fallback_chain:(.*?)(?=\n\w|\Z)', content, re.DOTALL)
    if not fallback_match:
        return {"status": "FAIL", "reason": "fallback_chain not found in providers.yml"}

    chain = [l.strip().lstrip("- ").strip() for l in fallback_match.group(1).splitlines()
             if l.strip() and not l.strip().startswith("#")]
    chain = [c for c in chain if c]

    return {
        "status": "PASS" if len(chain) >= 2 else "FAIL",
        "fallback_chain": chain,
        "chain_length": len(chain),
        "note": "System will try providers in order if primary fails",
    }

if __name__ == "__main__":
    print("=" * 60)
    print("TEST: Multi-Model Compatibility")
    print("=" * 60)

    r1 = test_tier_routing()
    print(f"\n[tier_routing] {r1['status']} | backend: {r1.get('backend')}")
    for row in r1.get("routing", []):
        print(f"  {row['agent']:<15} [{row['expected_tier']}] -> {row['model']}")

    r2 = test_capability_injection_for_small_models()
    print(f"\n[capability_injection] {r2['status']}")
    print(f"  Protocol fields adopted: {r2.get('protocol_fields_adopted','?')}/3")
    print(f"  Model: {r2.get('model_used','?')}")

    r3 = test_chain_of_thought_injection()
    print(f"\n[chain_of_thought] {r3['status']}")
    print(f"  Shows reasoning: {r3.get('shows_reasoning_steps')} | Has numbers: {r3.get('includes_number')}")

    r4 = test_provider_fallback_simulation()
    print(f"\n[fallback_chain] {r4['status']}")
    print(f"  Chain: {r4.get('fallback_chain','?')}")

    output = {
        "tier_routing": r1, "capability_injection": r2,
        "chain_of_thought": r3, "fallback_chain": r4,
    }
    out_path = pathlib.Path.home() / ".brain/tests/results/model_compatibility.json"
    out_path.write_text(json.dumps(output, indent=2))
    print(f"\nResultado guardado: {out_path}")

    any_fail = any(r.get("status") == "FAIL" for r in [r1, r2, r3, r4])
    sys.exit(1 if any_fail else 0)
