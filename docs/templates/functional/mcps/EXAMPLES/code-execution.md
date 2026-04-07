name: code-execution-mcp
version: 1.0.0
status: stable
protocol: MCP 1.0

<!-- markdownlint-disable-file -->

# MCP: Code Execution

## Summary

Tool integration allowing agents to execute Python/JavaScript code safely in sandbox.

## Tools Provided

### execute_python

```json
{
  "name": "execute_python",
  "description": "Execute Python code in sandbox",
  "input_schema": {
    "code": "string (Python code to execute)",
    "timeout": "number (seconds, default 30)"
  },
  "output": { "result": "stdout", "error": "stderr", "exit_code": 0 }
}
```

### execute_javascript

```json
{
  "name": "execute_javascript",
  "description": "Execute JavaScript code in Node.js sandbox",
  "input_schema": {
    "code": "string (JavaScript code)",
    "timeout": "number (seconds, default 30)"
  }
}
```

## Example

**Request**:

```python
execute_python(code="""
import json
data = {"users": [{"name": "Alice", "age": 30}]}
print(json.dumps(data, indent=2))
""")
```

**Response**:

```
{
  "result": "{\"users\": [{\"name\": \"Alice\", \"age\": 30}]}",
  "error": null,
  "exit_code": 0
}
```

## Constraints

- Timeout: 30 seconds max
- Memory: 512 MB limit
- Network: Blocked (no external requests)
- File system: Read-only
