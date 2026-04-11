package mcp

import "time"

// OfficialServers returns the 5 official Brain MCP server definitions.
func OfficialServers() []MCPServerConfig {
	return []MCPServerConfig{
		BrainFilesystemServer(),
		BrainGitServer(),
		BrainGithubServer(),
		BrainTerminalServer(),
		BrainKnowledgeServer(),
	}
}

// BrainFilesystemServer returns the filesystem MCP config.
func BrainFilesystemServer() MCPServerConfig {
	return MCPServerConfig{
		ID:          "brain-filesystem",
		Name:        "Brain Filesystem",
		Version:     "1.0.0",
		Description: "Official Brain MCP server for filesystem operations: read, write, list, search files and directories",
		Category:    "official",
		Transport:   TransportStdIO,
		Command:     "brain-mcp-filesystem",
		Args:        []string{"--mode", "stdio"},
		Env:         map[string]string{},
		Timeout:     30 * time.Second,
		RateLimit:   100,
	}
}

// BrainGitServer returns the git MCP config.
func BrainGitServer() MCPServerConfig {
	return MCPServerConfig{
		ID:          "brain-git",
		Name:        "Brain Git",
		Version:     "1.0.0",
		Description: "Official Brain MCP server for Git operations: status, diff, log, commit, branch, merge",
		Category:    "official",
		Transport:   TransportStdIO,
		Command:     "brain-mcp-git",
		Args:        []string{"--mode", "stdio"},
		Env:         map[string]string{},
		Timeout:     60 * time.Second,
		RateLimit:   50,
	}
}

// BrainGithubServer returns the GitHub MCP config.
func BrainGithubServer() MCPServerConfig {
	return MCPServerConfig{
		ID:          "brain-github",
		Name:        "Brain GitHub",
		Version:     "1.0.0",
		Description: "Official Brain MCP server for GitHub operations: PRs, issues, reviews, actions, releases",
		Category:    "official",
		Transport:   TransportHTTP,
		Command:     "brain-mcp-github",
		Args:        []string{"--mode", "stdio"},
		Env: map[string]string{
			"GITHUB_TOKEN": "",
		},
		Timeout:   30 * time.Second,
		RateLimit: 30,
	}
}

// BrainTerminalServer returns the terminal MCP config.
func BrainTerminalServer() MCPServerConfig {
	return MCPServerConfig{
		ID:          "brain-terminal",
		Name:        "Brain Terminal",
		Version:     "1.0.0",
		Description: "Official Brain MCP server for terminal/shell command execution with sandbox support",
		Category:    "official",
		Transport:   TransportStdIO,
		Command:     "brain-mcp-terminal",
		Args:        []string{"--mode", "stdio"},
		Env:         map[string]string{},
		Timeout:     120 * time.Second,
		RateLimit:   20,
	}
}

// BrainKnowledgeServer returns the knowledge (Qdrant) MCP config.
func BrainKnowledgeServer() MCPServerConfig {
	return MCPServerConfig{
		ID:          "brain-knowledge",
		Name:        "Brain Knowledge",
		Version:     "1.0.0",
		Description: "Official Brain MCP server for knowledge base operations using Qdrant vector database",
		Category:    "official",
		Transport:   TransportHTTP,
		Command:     "brain-mcp-knowledge",
		Args:        []string{"--mode", "stdio"},
		Env: map[string]string{
			"QDRANT_URL": "",
		},
		Timeout:   30 * time.Second,
		RateLimit: 60,
	}
}

// BrainContextServer returns the context bundle MCP config.
func BrainContextServer() MCPServerConfig {
	return MCPServerConfig{
		ID:          "brain-context",
		Name:        "Brain Context",
		Version:     "1.0.0",
		Description: "Official Brain MCP server for context management: session history, project context, memory",
		Category:    "official",
		Transport:   TransportStdIO,
		Command:     "brain-mcp-context",
		Args:        []string{"--mode", "stdio"},
		Env:         map[string]string{},
		Timeout:     30 * time.Second,
		RateLimit:   100,
	}
}

// OfficialFilesystemTools returns the tools available on the Brain Fileserver MCP.
func OfficialFilesystemTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "read_file",
			Description: "Read the contents of a file at the specified path",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":  map[string]interface{}{"type": "string", "description": "Absolute path to the file"},
					"limit": map[string]interface{}{"type": "integer", "description": "Maximum number of lines to read"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write contents to a file at the specified path",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Absolute path to the file"},
					"content": map[string]interface{}{"type": "string", "description": "Content to write"},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "list_directory",
			Description: "List files and directories at the specified path",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Absolute path to the directory"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "search_files",
			Description: "Search for files matching a pattern",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Root path to search"},
					"pattern": map[string]interface{}{"type": "string", "description": "Glob pattern to match"},
				},
				"required": []string{"path", "pattern"},
			},
		},
		{
			Name:        "edit_file",
			Description: "Edit a file by replacing a specific section of text",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":     map[string]interface{}{"type": "string", "description": "Absolute path to the file"},
					"old_text": map[string]interface{}{"type": "string", "description": "Text to replace"},
					"new_text": map[string]interface{}{"type": "string", "description": "Replacement text"},
				},
				"required": []string{"path", "old_text", "new_text"},
			},
		},
	}
}

// OfficialGitTools returns the tools available on the Brain Git MCP.
func OfficialGitTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "git_status",
			Description: "Show the working tree status",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{"type": "string", "description": "Path to the git repository"},
				},
				"required": []string{"repo_path"},
			},
		},
		{
			Name:        "git_diff",
			Description: "Show changes between commits, working tree and index",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{"type": "string", "description": "Path to the git repository"},
					"files":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Specific files to diff"},
				},
				"required": []string{"repo_path"},
			},
		},
		{
			Name:        "git_log",
			Description: "Show commit logs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{"type": "string", "description": "Path to the git repository"},
					"max_count": map[string]interface{}{"type": "integer", "description": "Maximum number of commits to show"},
				},
				"required": []string{"repo_path"},
			},
		},
		{
			Name:        "git_branch",
			Description: "List, create, or delete branches",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path":   map[string]interface{}{"type": "string", "description": "Path to the git repository"},
					"branch_name": map[string]interface{}{"type": "string", "description": "Name of the branch"},
					"action":      map[string]interface{}{"type": "string", "enum": []string{"list", "create", "delete"}, "description": "Action to perform"},
				},
				"required": []string{"repo_path", "action"},
			},
		},
		{
			Name:        "git_commit",
			Description: "Record changes to the repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{"type": "string", "description": "Path to the git repository"},
					"message":   map[string]interface{}{"type": "string", "description": "Commit message"},
				},
				"required": []string{"repo_path", "message"},
			},
		},
	}
}

// OfficialGithubTools returns the tools available on the Brain GitHub MCP.
func OfficialGithubTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "github_create_pr",
			Description: "Create a pull request on GitHub",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo":  map[string]interface{}{"type": "string", "description": "Repository in owner/repo format"},
					"title": map[string]interface{}{"type": "string", "description": "PR title"},
					"body":  map[string]interface{}{"type": "string", "description": "PR description"},
					"head":  map[string]interface{}{"type": "string", "description": "Head branch"},
					"base":  map[string]interface{}{"type": "string", "description": "Base branch"},
				},
				"required": []string{"repo", "title", "head", "base"},
			},
		},
		{
			Name:        "github_list_prs",
			Description: "List pull requests on GitHub",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo":  map[string]interface{}{"type": "string", "description": "Repository in owner/repo format"},
					"state": map[string]interface{}{"type": "string", "enum": []string{"open", "closed", "all"}, "description": "PR state filter"},
				},
				"required": []string{"repo"},
			},
		},
		{
			Name:        "github_review_pr",
			Description: "Review a pull request on GitHub",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo":      map[string]interface{}{"type": "string", "description": "Repository in owner/repo format"},
					"pr_number": map[string]interface{}{"type": "integer", "description": "PR number"},
					"event":     map[string]interface{}{"type": "string", "enum": []string{"approve", "request_changes", "comment"}, "description": "Review event"},
					"body":      map[string]interface{}{"type": "string", "description": "Review body"},
				},
				"required": []string{"repo", "pr_number", "event"},
			},
		},
		{
			Name:        "github_list_issues",
			Description: "List issues on GitHub",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo":  map[string]interface{}{"type": "string", "description": "Repository in owner/repo format"},
					"state": map[string]interface{}{"type": "string", "enum": []string{"open", "closed", "all"}, "description": "Issue state filter"},
				},
				"required": []string{"repo"},
			},
		},
	}
}

// OfficialTerminalTools returns the tools available on the Brain Terminal MCP.
func OfficialTerminalTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "run_command",
			Description: "Execute a shell command in a sandbox terminal",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command":     map[string]interface{}{"type": "string", "description": "Command to execute"},
					"working_dir": map[string]interface{}{"type": "string", "description": "Working directory for the command"},
					"timeout_ms":  map[string]interface{}{"type": "integer", "description": "Timeout in milliseconds"},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "run_script",
			Description: "Execute a script file in a sandbox terminal",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"script_path": map[string]interface{}{"type": "string", "description": "Path to the script file"},
					"interpreter": map[string]interface{}{"type": "string", "description": "Script interpreter (bash, python, etc.)"},
				},
				"required": []string{"script_path"},
			},
		},
	}
}

// OfficialKnowledgeTools returns the tools available on the Brain Knowledge (Qdrant) MCP.
func OfficialKnowledgeTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "search_knowledge",
			Description: "Search the knowledge base using vector similarity",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":           map[string]interface{}{"type": "string", "description": "Search query"},
					"collection":      map[string]interface{}{"type": "string", "description": "Qdrant collection name"},
					"limit":           map[string]interface{}{"type": "integer", "description": "Number of results to return"},
					"score_threshold": map[string]interface{}{"type": "number", "description": "Minimum similarity score"},
				},
				"required": []string{"query", "collection"},
			},
		},
		{
			Name:        "add_knowledge",
			Description: "Add content to the knowledge base",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content":    map[string]interface{}{"type": "string", "description": "Content to add"},
					"collection": map[string]interface{}{"type": "string", "description": "Qdrant collection name"},
					"metadata":   map[string]interface{}{"type": "object", "description": "Additional metadata"},
				},
				"required": []string{"content", "collection"},
			},
		},
		{
			Name:        "list_collections",
			Description: "List available Qdrant collections",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// OfficialContextTools returns the tools available on the Brain Context MCP.
func OfficialContextTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "get_session_context",
			Description: "Retrieve the current session context and history",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_id": map[string]interface{}{"type": "string", "description": "Session identifier"},
				},
				"required": []string{"session_id"},
			},
		},
		{
			Name:        "update_context",
			Description: "Update the project context with new information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key":   map[string]interface{}{"type": "string", "description": "Context key"},
					"value": map[string]interface{}{"type": "string", "description": "Context value"},
				},
				"required": []string{"key", "value"},
			},
		},
		{
			Name:        "list_memory",
			Description: "List stored memory entries",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": map[string]interface{}{"type": "string", "enum": []string{"global", "project"}, "description": "Memory scope"},
					"limit": map[string]interface{}{"type": "integer", "description": "Maximum entries to return"},
				},
				"required": []string{"scope"},
			},
		},
	}
}
