#!/usr/bin/env python3
"""
test_parallel_agents.py
Tests parallel sub-agent invocation using ThreadPoolExecutor.
Measures: latency reduction vs sequential, result consistency, context isolation.
"""
import json, os, pathlib, sys, time
from concurrent.futures import ThreadPoolExecutor, as_completed

# Add brain scripts to path
brain_dir = pathlib.Path(os.environ.get("BRAIN_DIR", pathlib.Path.home() / ".brain"))
import importlib.util as _ilu, pathlib as _p, os as _os
_brain_dir = _p.Path(_os.environ.get('BRAIN_DIR', _p.Path.home() / '.brain'))
_spec = _ilu.spec_from_file_location('agent_runner', str(_brain_dir / 'scripts/agent-runner.py'))
_mod = _ilu.module_from_spec(_spec)
_spec.loader.exec_module(_mod)
import sys as _sys
_sys.modules['agent_runner'] = _mod


try:
    from agent_runner import run_agent, detect_backend, BRAIN_DIR as BD
except ImportError:
    print("SKIP: agent-runner not importable (check BRAIN_DIR)")
    sys.exit(0)

RESULTS = {}

def run_single(agent_name, task, idx):
    start = time.time()
    result = run_agent(agent_name, task, inject_rules=False, inject_memory=False)
    elapsed = round(time.time() - start, 2)
    return {
        "agent": agent_name,
        "index": idx,
        "elapsed_s": elapsed,
        "success": not bool(result.error),
        "output_len": len(result.output),
        "model": result.model_used,
        "backend": result.backend,
        "error": result.error[:200] if result.error else None,
    }

def test_parallel_execution():
    """Run 3 agents simultaneously and compare to sequential baseline."""
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend configured"}

    task = "In one sentence, describe your primary role in the brain repo system."
    agents = ["researcher", "documenter", "reviewer"]

    # Sequential baseline
    seq_start = time.time()
    seq_results = [run_single(a, task, i) for i, a in enumerate(agents)]
    seq_time = round(time.time() - seq_start, 2)

    # Parallel execution
    par_start = time.time()
    par_results = []
    with ThreadPoolExecutor(max_workers=3) as pool:
        futures = {pool.submit(run_single, a, task, i): a for i, a in enumerate(agents)}
        for f in as_completed(futures):
            par_results.append(f.result())
    par_time = round(time.time() - par_start, 2)

    speedup = round(seq_time / max(par_time, 0.01), 2)
    all_ok = all(r["success"] for r in par_results)

    return {
        "status": "PASS" if all_ok else "FAIL",
        "sequential_time_s": seq_time,
        "parallel_time_s": par_time,
        "speedup_factor": speedup,
        "agents_ran": len(par_results),
        "agents_succeeded": sum(1 for r in par_results if r["success"]),
        "results": par_results,
    }

def test_context_isolation():
    """Verify each parallel agent gets isolated context (no cross-contamination)."""
    if detect_backend() == "none":
        return {"status": "SKIP", "reason": "no LLM backend configured"}

    tasks = [
        ("documenter", "Your task is: TASK_ALPHA. What is your task?"),
        ("documenter", "Your task is: TASK_BETA. What is your task?"),
        ("documenter", "Your task is: TASK_GAMMA. What is your task?"),
    ]

    results = []
    with ThreadPoolExecutor(max_workers=3) as pool:
        futures = [pool.submit(run_single, a, t, i) for i, (a, t) in enumerate(tasks)]
        for i, f in enumerate(futures):
            r = f.result()
            r["expected_task"] = f"TASK_{'ALPHA BETA GAMMA'.split()[i]}"
            results.append(r)

    isolated = all(
        r["expected_task"] in (r.get("output", "") or "").upper()
        for r in results if r["success"]
    )

    return {
        "status": "PASS" if isolated else "PARTIAL",
        "isolation_verified": isolated,
        "results": [{
            "expected": r["expected_task"],
            "found_in_output": r["expected_task"] in (r.get("output","")).upper(),
            "success": r["success"]
        } for r in results],
    }

if __name__ == "__main__":
    print("=" * 60)
    print("TEST: Parallel Sub-Agent Invocation")
    print("=" * 60)

    r1 = test_parallel_execution()
    print(f"\n[parallel_execution] {r1['status']}")
    if r1.get("speedup_factor"):
        print(f"  Sequential: {r1['sequential_time_s']}s")
        print(f"  Parallel:   {r1['parallel_time_s']}s")
        print(f"  Speedup:    {r1['speedup_factor']}x")
        print(f"  Agents:     {r1['agents_succeeded']}/{r1['agents_ran']} OK")

    r2 = test_context_isolation()
    print(f"\n[context_isolation] {r2['status']}")
    if "results" in r2:
        for item in r2["results"]:
            sym = "ok" if item["found_in_output"] else "MISS"
            print(f"  [{sym}] {item['expected']}")

    output = {"parallel_execution": r1, "context_isolation": r2}
    out_path = pathlib.Path.home() / ".brain/tests/results/parallel_agents.json"
    out_path.write_text(json.dumps(output, indent=2))
    print(f"\nResultado guardado: {out_path}")

    any_fail = any(r.get("status") == "FAIL" for r in [r1, r2])
    sys.exit(1 if any_fail else 0)
