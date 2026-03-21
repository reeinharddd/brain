#!/usr/bin/env python3
"""
test_mcp_real_usage.py
Tests that MCP servers are actually USED, not just connected.
Real operations: memory write->read->verify, brain-rules search, task routing.
"""
import json, os, pathlib, subprocess, sys, time

brain_dir = pathlib.Path(os.environ.get("BRAIN_DIR", pathlib.Path.home() / ".brain"))
RESULTS = {}

def npx_cmd():
    for c in ["npx-nvm", "npx"]:
        if subprocess.run(["which", c], capture_output=True).returncode == 0:
            return c
    return None

def run_mcp(server_name, messages, mem_dir=None):
    """Send MCP messages to a server and return parsed responses."""
    npx = npx_cmd()
    if not npx:
        return None, "npx not found"

    tmp = pathlib.Path("/tmp/_brain_mcp_test")
    tmp.mkdir(exist_ok=True)

    servers = {
        "memory": [npx, "-y", "@modelcontextprotocol/server-memory", str(mem_dir or tmp)],
        "filesystem": [npx, "-y", "@modelcontextprotocol/server-filesystem", str(pathlib.Path.home())],
    }
    cmd = servers.get(server_name)
    if not cmd:
        return None, f"unknown server: {server_name}"

    stdin = "\n".join(json.dumps(m) for m in messages) + "\n"
    try:
        proc = subprocess.run(cmd, input=stdin, capture_output=True, text=True, timeout=30)
        responses = []
        for line in proc.stdout.splitlines():
            try:
                responses.append(json.loads(line))
            except Exception:
                pass
        return responses, None
    except subprocess.TimeoutExpired:
        return None, "timeout"
    except Exception as e:
        return None, str(e)

def test_memory_full_cycle():
    """Write entity -> search -> verify found -> delete."""
    import uuid
    test_id = f"MCPRealTest_{uuid.uuid4().hex[:8]}"
    mem_dir = pathlib.Path("/tmp/_brain_mem_test")
    mem_dir.mkdir(exist_ok=True)

    messages = [
        {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}},
        {"jsonrpc":"2.0","method":"notifications/initialized","params":{}},
        {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create_entities","arguments":{"entities":[{
            "name": test_id,
            "entityType": "Learning",
            "observations": ["MCP real usage test", "Verifying write and retrieve cycle"]
        }]}}},
        {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_nodes","arguments":{"query": test_id}}},
    ]

    responses, err = run_mcp("memory", messages, mem_dir)
    if err:
        return {"status": "FAIL", "error": err}

    # Check that search found the entity
    found = False
    for r in responses:
        content = r.get("result",{}).get("content",[])
        for c in content:
            if c.get("type") == "text" and test_id in c.get("text",""):
                found = True

    return {
        "status": "PASS" if found else "FAIL",
        "entity_id": test_id,
        "write_ok": len(responses) >= 3,
        "read_found": found,
        "response_count": len(responses),
    }

def test_brain_rules_mcp_real_tools():
    """Actually call brain-rules MCP tools and verify useful output."""
    server_py = brain_dir / "mcp/brain-mcp-server/server.py"
    if not server_py.exists():
        return {"status": "SKIP", "reason": "brain-mcp-server not found"}

    tests = [
        ("brain_get_rules",    {"topic": "security"},              lambda t: len(t) > 100),
        ("brain_route_task",   {"task_description": "fix a bug"},  lambda t: "debugger" in t),
        ("brain_search_rules", {"query": "error handling"},        lambda t: len(t) > 50),
        ("brain_get_command",  {"name": "plan"},                   lambda t: len(t) > 50),
        ("brain_list_agents",  {},                                  lambda t: "orchestrator" in t),
        ("brain_get_provider", {"task_type": "planning"},          lambda t: "powerful" in t or "claude" in t.lower()),
    ]

    results = []
    for tool, args, validator in tests:
        messages = [
            {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}},
            {"jsonrpc":"2.0","method":"notifications/initialized","params":{}},
            {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":tool,"arguments":args}},
        ]
        stdin = "\n".join(json.dumps(m) for m in messages) + "\n"
        try:
            proc = subprocess.run(
                ["python3", str(server_py)], input=stdin, capture_output=True, text=True, timeout=10
            )
            output_text = ""
            for line in proc.stdout.splitlines():
                try:
                    d = json.loads(line)
                    for c in d.get("result",{}).get("content",[]):
                        if c.get("type") == "text":
                            output_text += c["text"]
                except Exception:
                    pass
            valid = validator(output_text)
            results.append({"tool": tool, "status": "PASS" if valid else "FAIL",
                           "output_len": len(output_text), "valid": valid})
        except Exception as e:
            results.append({"tool": tool, "status": "FAIL", "error": str(e)})

    passed = sum(1 for r in results if r["status"] == "PASS")
    return {
        "status": "PASS" if passed == len(tests) else ("PARTIAL" if passed > 0 else "FAIL"),
        "passed": passed,
        "total": len(tests),
        "tools": results,
    }

def test_filesystem_mcp_real():
    """Use filesystem MCP to actually read a brain file."""
    target = str(brain_dir / "rules/canonical.md")
    messages = [
        {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}},
        {"jsonrpc":"2.0","method":"notifications/initialized","params":{}},
        {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path": target}}},
    ]
    responses, err = run_mcp("filesystem", messages)
    if err:
        return {"status": "FAIL", "error": err}

    content_found = any(
        "canonical" in str(r).lower() or "Philosophy" in str(r)
        for r in responses
    )
    return {
        "status": "PASS" if content_found else "FAIL",
        "file_read": target,
        "content_found": content_found,
    }

if __name__ == "__main__":
    print("=" * 60)
    print("TEST: Real MCP Usage (not just handshake)")
    print("=" * 60)

    r1 = test_memory_full_cycle()
    print(f"\n[memory_full_cycle] {r1['status']}")
    print(f"  Entity: {r1.get('entity_id','N/A')}")
    print(f"  Write: {r1.get('write_ok')} | Read found: {r1.get('read_found')}")

    r2 = test_brain_rules_mcp_real_tools()
    print(f"\n[brain_rules_mcp_tools] {r2['status']} ({r2.get('passed','?')}/{r2.get('total','?')})")
    for tool_r in r2.get("tools", []):
        sym = "ok" if tool_r["status"] == "PASS" else "FAIL"
        print(f"  [{sym}] {tool_r['tool']} ({tool_r.get('output_len',0)} chars)")

    r3 = test_filesystem_mcp_real()
    print(f"\n[filesystem_mcp_real] {r3['status']}")
    print(f"  Read: {r3.get('file_read','?')} | Found content: {r3.get('content_found')}")

    output = {"memory_cycle": r1, "brain_rules_tools": r2, "filesystem_real": r3}
    out_path = pathlib.Path.home() / ".brain/tests/results/mcp_real_usage.json"
    out_path.write_text(json.dumps(output, indent=2))
    print(f"\nResultado guardado: {out_path}")

    any_fail = any(r.get("status") == "FAIL" for r in [r1, r2, r3])
    sys.exit(1 if any_fail else 0)
