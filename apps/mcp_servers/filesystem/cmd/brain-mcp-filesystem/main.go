package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		name:    "brain-filesystem",
		version: "1.0.0",
	}
	s.tools = s.defineTools()
	return s
}

func (s *Server) defineTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "read_file",
			Description: "Read the contents of a file at the specified path",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":   map[string]interface{}{"type": "string", "description": "Absolute path to the file"},
					"limit":  map[string]interface{}{"type": "integer", "description": "Maximum number of lines to read"},
					"offset": map[string]interface{}{"type": "integer", "description": "Starting line number (0-based)"},
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
			Description: "Search for files matching a glob pattern",
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
	case "read_file":
		return s.callReadFile(call.Arguments)
	case "write_file":
		return s.callWriteFile(call.Arguments)
	case "list_directory":
		return s.callListDirectory(call.Arguments)
	case "search_files":
		return s.callSearchFiles(call.Arguments)
	case "edit_file":
		return s.callEditFile(call.Arguments)
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

func (s *Server) callReadFile(args map[string]interface{}) (interface{}, error) {
	path := getStr(args, "path")
	if path == "" {
		return nil, &RPCError{Code: -32602, Message: "path is required"}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	lines := strings.Split(string(data), "\n")
	limit := getInt(args, "limit", 0)
	offset := getInt(args, "offset", 0)

	if offset > 0 && offset < len(lines) {
		lines = lines[offset:]
	}
	if limit > 0 && len(lines) > limit {
		lines = lines[:limit]
	}

	return textResult(strings.Join(lines, "\n")), nil
}

func (s *Server) callWriteFile(args map[string]interface{}) (interface{}, error) {
	path := getStr(args, "path")
	content := getStr(args, "content")
	if path == "" {
		return nil, &RPCError{Code: -32602, Message: "path is required"}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errorResult(fmt.Sprintf("failed to create parent directories: %v", err)), nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return errorResult(fmt.Sprintf("failed to write file: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Successfully wrote to %s", path)), nil
}

func (s *Server) callListDirectory(args map[string]interface{}) (interface{}, error) {
	path := getStr(args, "path")
	if path == "" {
		return nil, &RPCError{Code: -32602, Message: "path is required"}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list directory: %v", err)), nil
	}

	var items []string
	for _, e := range entries {
		if e.IsDir() {
			items = append(items, fmt.Sprintf("[DIR]  %s", e.Name()))
		} else {
			info, err := e.Info()
			if err == nil {
				items = append(items, fmt.Sprintf("[FILE] %s (%d bytes)", e.Name(), info.Size()))
			} else {
				items = append(items, fmt.Sprintf("[FILE] %s", e.Name()))
			}
		}
	}

	if len(items) == 0 {
		return textResult("(empty directory)"), nil
	}

	return textResult(strings.Join(items, "\n")), nil
}

func (s *Server) callSearchFiles(args map[string]interface{}) (interface{}, error) {
	root := getStr(args, "path")
	pattern := getStr(args, "pattern")
	if root == "" || pattern == "" {
		return nil, &RPCError{Code: -32602, Message: "path and pattern are required"}
	}

	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		matched, merr := filepath.Match(pattern, info.Name())
		if merr != nil {
			return nil
		}
		if matched && !info.IsDir() {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return errorResult(fmt.Sprintf("failed to search files: %v", err)), nil
	}

	if len(matches) == 0 {
		return textResult("No files matched the pattern"), nil
	}

	return textResult(strings.Join(matches, "\n")), nil
}

func (s *Server) callEditFile(args map[string]interface{}) (interface{}, error) {
	path := getStr(args, "path")
	oldText := getStr(args, "old_text")
	newText := getStr(args, "new_text")
	if path == "" || oldText == "" {
		return nil, &RPCError{Code: -32602, Message: "path and old_text are required"}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	content := string(data)
	if !strings.Contains(content, oldText) {
		return errorResult("old_text not found in file"), nil
	}

	content = strings.Replace(content, oldText, newText, 1)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return errorResult(fmt.Sprintf("failed to write file: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Successfully edited %s", path)), nil
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
