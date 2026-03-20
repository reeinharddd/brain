# brain-mcp-server

Custom MCP server that exposes brain repo internals as queryable tools for any MCP-compatible client.

## Tools

| Tool | Description |
| :--- | :--- |
| `brain_get_rules` | Get relevant rules from canonical.md by topic |
| `brain_get_agent` | Get full agent definition by name |
| `brain_list_agents` | List all agents with descriptions |
| `brain_get_command` | Get slash command definition |
| `brain_route_task` | Route a task to the right agent + model tier |
| `brain_search_rules` | Full-text search across all rules and modules |
| `brain_get_provider` | Get recommended model for a task type |

## Usage (stdio)

```bash
python3 ~/.brain/mcp/brain-mcp-server/server.py
```

## Add to Claude Code settings

```json
{
  "mcpServers": {
    "brain": {
      "command": "python3",
      "args": ["${HOME}/.brain/mcp/brain-mcp-server/server.py"]
    }
  }
}
```

## Add to generate.sh

The brain MCP server is automatically included when `generate.sh` is run.
It is registered in `mcp/registry.yml` under the `brain-rules` key.
