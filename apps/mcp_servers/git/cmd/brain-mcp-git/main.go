package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// JSON-RPC types

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return e.Message
}

type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Server

type Server struct {
	name    string
	version string
	tools   []ToolDef
}

func newServer() *Server {
	s := &Server{
		name:    "brain-git",
		version: "1.0.0",
	}
	s.tools = s.defineTools()
	return s
}

func (s *Server) defineTools() []ToolDef {
	return []ToolDef{
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

func (s *Server) handleInitialize(_ json.RawMessage) interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
		"serverInfo":      map[string]interface{}{"name": s.name, "version": s.version},
	}
}

func (s *Server) handleToolsList(_ json.RawMessage) interface{} {
	return map[string]interface{}{
		"tools": s.tools,
	}
}

func (s *Server) handleToolsCall(params json.RawMessage) (interface{}, error) {
	var call struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, &RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
	}

	switch call.Name {
	case "git_status":
		return s.callGitStatus(call.Arguments)
	case "git_diff":
		return s.callGitDiff(call.Arguments)
	case "git_log":
		return s.callGitLog(call.Arguments)
	case "git_branch":
		return s.callGitBranch(call.Arguments)
	case "git_commit":
		return s.callGitCommit(call.Arguments)
	default:
		return nil, &RPCError{Code: -32601, Message: fmt.Sprintf("unknown tool: %s", call.Name)}
	}
}

func getStr(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func getInt(args map[string]interface{}, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return defaultVal
}

func getStrSlice(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok {
		return nil
	}
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for _, item := range raw {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func textResult(text string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func errorResult(msg string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": "Error: " + msg},
		},
		"isError": true,
	}
}

func runGit(repoPath string, gitArgs ...string) (string, error) {
	args := []string{"-C", repoPath}
	args = append(args, gitArgs...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s (%v)", strings.Join(gitArgs, " "), string(output), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (s *Server) callGitStatus(args map[string]interface{}) (interface{}, error) {
	repoPath := getStr(args, "repo_path")
	if repoPath == "" {
		return nil, &RPCError{Code: -32602, Message: "repo_path is required"}
	}

	output, err := runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return errorResult(fmt.Sprintf("failed to get git status: %v", err)), nil
	}

	if output == "" {
		return textResult("Working tree is clean"), nil
	}

	return textResult(output), nil
}

func (s *Server) callGitDiff(args map[string]interface{}) (interface{}, error) {
	repoPath := getStr(args, "repo_path")
	if repoPath == "" {
		return nil, &RPCError{Code: -32602, Message: "repo_path is required"}
	}

	files := getStrSlice(args, "files")
	gitArgs := []string{"diff"}
	if len(files) > 0 {
		gitArgs = append(gitArgs, "--")
		gitArgs = append(gitArgs, files...)
	}

	output, err := runGit(repoPath, gitArgs...)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to get git diff: %v", err)), nil
	}

	if output == "" {
		return textResult("No changes to show"), nil
	}

	return textResult(output), nil
}

func (s *Server) callGitLog(args map[string]interface{}) (interface{}, error) {
	repoPath := getStr(args, "repo_path")
	if repoPath == "" {
		return nil, &RPCError{Code: -32602, Message: "repo_path is required"}
	}

	maxCount := getInt(args, "max_count", 10)
	output, err := runGit(repoPath, "log", "--oneline", "-n", strconv.Itoa(maxCount))
	if err != nil {
		return errorResult(fmt.Sprintf("failed to get git log: %v", err)), nil
	}

	if output == "" {
		return textResult("No commits found"), nil
	}

	return textResult(output), nil
}

func (s *Server) callGitBranch(args map[string]interface{}) (interface{}, error) {
	repoPath := getStr(args, "repo_path")
	if repoPath == "" {
		return nil, &RPCError{Code: -32602, Message: "repo_path is required"}
	}

	action := getStr(args, "action")
	if action == "" {
		return nil, &RPCError{Code: -32602, Message: "action is required"}
	}

	switch action {
	case "list":
		output, err := runGit(repoPath, "branch", "-a")
		if err != nil {
			return errorResult(fmt.Sprintf("failed to list branches: %v", err)), nil
		}
		if output == "" {
			return textResult("No branches found"), nil
		}
		return textResult(output), nil

	case "create":
		branchName := getStr(args, "branch_name")
		if branchName == "" {
			return nil, &RPCError{Code: -32602, Message: "branch_name is required for create action"}
		}
		output, err := runGit(repoPath, "branch", branchName)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to create branch: %v", err)), nil
		}
		return textResult(fmt.Sprintf("Created branch '%s'\n%s", branchName, output)), nil

	case "delete":
		branchName := getStr(args, "branch_name")
		if branchName == "" {
			return nil, &RPCError{Code: -32602, Message: "branch_name is required for delete action"}
		}
		output, err := runGit(repoPath, "branch", "-d", branchName)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to delete branch: %v", err)), nil
		}
		return textResult(fmt.Sprintf("Deleted branch '%s'\n%s", branchName, output)), nil

	default:
		return nil, &RPCError{Code: -32602, Message: fmt.Sprintf("unknown action: %s (expected list, create, or delete)", action)}
	}
}

func (s *Server) callGitCommit(args map[string]interface{}) (interface{}, error) {
	repoPath := getStr(args, "repo_path")
	message := getStr(args, "message")
	if repoPath == "" {
		return nil, &RPCError{Code: -32602, Message: "repo_path is required"}
	}
	if message == "" {
		return nil, &RPCError{Code: -32602, Message: "message is required"}
	}

	// Stage all changes
	_, err := runGit(repoPath, "add", "-A")
	if err != nil {
		return errorResult(fmt.Sprintf("failed to stage changes: %v", err)), nil
	}

	// Commit
	output, err := runGit(repoPath, "commit", "-m", message)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to commit: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Committed successfully\n%s", output)), nil
}

func (s *Server) handleRequest(line string) {
	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		resp := JSONRPCResponse{JSONRPC: "2.0", ID: nil, Error: &RPCError{Code: -32700, Message: "Parse error"}}
		sendResponse(resp)
		return
	}

	var result interface{}
	var callErr error

	switch req.Method {
	case "initialize":
		result = s.handleInitialize(req.Params)
	case "tools/list":
		result = s.handleToolsList(req.Params)
	case "tools/call":
		result, callErr = s.handleToolsCall(req.Params)
	default:
		callErr = &RPCError{Code: -32601, Message: fmt.Sprintf("unknown method: %s", req.Method)}
	}

	resp := JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
	if callErr != nil {
		resp.Error = callErr.(*RPCError)
	} else {
		resp.Result = result
	}
	sendResponse(resp)
}

func sendResponse(resp JSONRPCResponse) {
	data, _ := json.Marshal(resp)
	fmt.Fprintln(os.Stdout, string(data))
	os.Stdout.Sync()
}

func main() {
	server := newServer()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		server.handleRequest(line)
	}
}
