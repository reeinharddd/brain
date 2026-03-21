#!/usr/bin/env python3
"""
test_context_management.py
Tests context window management:
- Token budget tracking across agent chain
- Context compression at boundaries
- Memory injection without bloat
- Small model compatibility (reduced context)
- Handoff artifacts between agents
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
    from agent_runner import run_agent, detect_backend, load_agent
except ImportError:
    print("SKIP: agent-runner not importable")
    sys.exit(0)

def estimate_tokens(text):
    """Rough token estimate: ~4 chars per token for English."""
    return len(text) // 4

def test_context_budget_chain():
    """
    Tests that as context passes through agents, it doesn't exceed budget.
    Simulates: planner output -> implementer input -> reviewer input.
    Max budget: 8000 tokens total (safe for small models).
    """
    MAX_TOKENS = 8000

    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend"}

    task = "Add input validation to a user registration form. Username 3-20 chars, email format, password 8+ chars."

    # Stage 1: planner produces spec
    plan_result = run_agent("planner", task, inject_rules=False, inject_memory=False)
    if plan_result.error:
        return {"status": "FAIL", "stage": "planner", "error": plan_result.error}

    plan_tokens = estimate_tokens(plan_result.output)

    # Stage 2: implementer receives plan (context = planner output, truncated if needed)
    max_context = 2000  # chars, ~500 tokens for context
    plan_context = plan_result.output[:max_context]
    compressed = len(plan_result.output) > max_context

    impl_result = run_agent("documenter",
                            "Document the implementation plan in 3 bullet points",
                            context=f"Plan from planner:\n{plan_context}",
                            inject_rules=False, inject_memory=False)
    if impl_result.error:
        return {"status": "FAIL", "stage": "documenter", "error": impl_result.error}

    impl_tokens = estimate_tokens(impl_result.output)

    # Stage 3: reviewer receives compressed summary
    summary_context = f"Plan: {plan_result.output[:500]}\n\nDocs: {impl_result.output[:500]}"
    review_result = run_agent("reviewer",
                              "Review the plan and documentation quality",
                              context=summary_context,
                              inject_rules=False, inject_memory=False)

    total_tokens = plan_tokens + impl_tokens + estimate_tokens(review_result.output or "")

    return {
        "status": "PASS" if total_tokens <= MAX_TOKENS else "WARN",
        "stages_completed": 3,
        "token_budget": MAX_TOKENS,
        "tokens_used": total_tokens,
        "within_budget": total_tokens <= MAX_TOKENS,
        "context_compressed": compressed,
        "plan_tokens": plan_tokens,
        "impl_tokens": impl_tokens,
        "review_ok": not bool(review_result.error),
    }

def test_small_model_compatibility():
    """
    Tests that agents work with minimal context injection
    (simulating small/fast models with limited context windows).
    """
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend"}

    # Use documenter (fast tier = smallest model)
    # Give it ONLY the task, no canonical rules, no memory
    result = run_agent(
        "documenter",
        "Write a 2-sentence summary of what JWT authentication is.",
        inject_rules=False,
        inject_memory=False,
    )

    output_tokens = estimate_tokens(result.output or "")

    return {
        "status": "PASS" if not result.error and output_tokens > 10 else "FAIL",
        "model_used": result.model_used,
        "backend": result.backend,
        "output_tokens": output_tokens,
        "error": result.error,
        "note": "Fast-tier model ran without full rule injection",
    }

def test_rules_injection_overhead():
    """
    Measures how much token overhead canonical.md adds vs no injection.
    Helps understand cost of full context vs selective injection.
    """
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend"}

    task = "In exactly 10 words, describe what a REST API is."

    # Without rules
    r_no_rules = run_agent("documenter", task, inject_rules=False, inject_memory=False)
    # With rules
    r_with_rules = run_agent("documenter", task, inject_rules=True, inject_memory=False)

    canonical_size = estimate_tokens(
        (brain_dir / "rules/canonical.md").read_text() if
        (brain_dir / "rules/canonical.md").exists() else ""
    )

    return {
        "status": "PASS",
        "canonical_tokens": canonical_size,
        "output_without_rules": r_no_rules.output[:200] if not r_no_rules.error else "ERROR",
        "output_with_rules": r_with_rules.output[:200] if not r_with_rules.error else "ERROR",
        "rules_affect_output": r_no_rules.output != r_with_rules.output,
        "recommendation": (
            "Inject rules only for planning/design agents (powerful tier). "
            "Skip for fast-tier repetitive tasks."
        ),
    }

def test_handoff_artifact_format():
    """
    Tests that agents produce structured handoff artifacts
    matching the agent-contracts.md spec (summary, artifact, risks, handoff).
    """
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend"}

    contract_prompt = """
You must respond with a structured handoff artifact in this exact format:
SUMMARY: [one sentence of what you did]
ARTIFACT: [your main output]
RISKS: [any concerns]
HANDOFF: [what the next agent needs]

Task: Analyze the risk of using MD5 for password hashing.
"""
    result = run_agent("reviewer", contract_prompt, inject_rules=False, inject_memory=False)

    if result.error:
        return {"status": "FAIL", "error": result.error}

    output = result.output or ""
    has_summary  = "SUMMARY:" in output
    has_artifact = "ARTIFACT:" in output
    has_risks    = "RISKS:" in output
    has_handoff  = "HANDOFF:" in output
    score = sum([has_summary, has_artifact, has_risks, has_handoff])

    return {
        "status": "PASS" if score >= 3 else ("PARTIAL" if score >= 2 else "FAIL"),
        "contract_fields_present": score,
        "has_summary":  has_summary,
        "has_artifact": has_artifact,
        "has_risks":    has_risks,
        "has_handoff":  has_handoff,
        "output_preview": output[:400],
    }

if __name__ == "__main__":
    print("=" * 60)
    print("TEST: Context Management & Delegation")
    print("=" * 60)

    r1 = test_context_budget_chain()
    print(f"\n[context_budget_chain] {r1['status']}")
    if "tokens_used" in r1:
        print(f"  Budget: {r1['token_budget']} | Used: {r1['tokens_used']} | OK: {r1['within_budget']}")
        print(f"  Stages: {r1.get('stages_completed')} | Compressed: {r1.get('context_compressed')}")

    r2 = test_small_model_compatibility()
    print(f"\n[small_model_compat] {r2['status']}")
    if "model_used" in r2:
        print(f"  Model: {r2['model_used']} | Tokens out: {r2['output_tokens']}")

    r3 = test_rules_injection_overhead()
    print(f"\n[rules_injection_overhead] {r3['status']}")
    if "canonical_tokens" in r3:
        print(f"  canonical.md tokens: {r3['canonical_tokens']}")
        print(f"  Rules affect output: {r3['rules_affect_output']}")

    r4 = test_handoff_artifact_format()
    print(f"\n[handoff_artifact_format] {r4['status']}")
    if "contract_fields_present" in r4:
        print(f"  Contract fields: {r4['contract_fields_present']}/4")

    output = {
        "context_budget_chain": r1,
        "small_model_compat": r2,
        "rules_injection_overhead": r3,
        "handoff_artifact_format": r4,
    }
    out_path = pathlib.Path.home() / ".brain/tests/results/context_management.json"
    out_path.write_text(json.dumps(output, indent=2))
    print(f"\nResultado guardado: {out_path}")

    any_fail = any(r.get("status") == "FAIL" for r in [r1, r2, r3, r4])
    sys.exit(1 if any_fail else 0)
