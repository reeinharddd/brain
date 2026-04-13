package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
		name:    "brain-terminal",
		version: "1.0.0",
	}
	s.tools = s.defineTools()
	return s
}

func (s *Server) defineTools() []ToolDef {
	return []ToolDef{
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
					"working_dir": map[string]interface{}{"type": "string", "description": "Working directory for the command"},
					"timeout_ms":  map[string]interface{}{"type": "integer", "description": "Timeout in milliseconds"},
				},
				"required": []string{"script_path"},
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
	case "run_command":
		return s.callRunCommand(call.Arguments)
	case "run_script":
		return s.callRunScript(call.Arguments)
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

func (s *Server) callRunCommand(args map[string]interface{}) (interface{}, error) {
	command := getStr(args, "command")
	if command == "" {
		return nil, &RPCError{Code: -32602, Message: "command is required"}
	}

	workingDir := getStr(args, "working_dir")
	timeoutMs := getInt(args, "timeout_ms", 30000)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := buildCommandResult(stdout.String(), stderr.String(), err)
	return textResult(result), nil
}

func (s *Server) callRunScript(args map[string]interface{}) (interface{}, error) {
	scriptPath := getStr(args, "script_path")
	if scriptPath == "" {
		return nil, &RPCError{Code: -32602, Message: "script_path is required"}
	}

	interpreter := getStr(args, "interpreter")
	if interpreter == "" {
		interpreter = "bash"
	}

	workingDir := getStr(args, "working_dir")
	timeoutMs := getInt(args, "timeout_ms", 30000)

	// Resolve to absolute path
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to resolve script path: %v", err)), nil
	}

	// Check script exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errorResult(fmt.Sprintf("script file not found: %s", absPath)), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, interpreter, absPath)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	result := buildCommandResult(stdout.String(), stderr.String(), err)
	return textResult(result), nil
}

func buildCommandResult(stdoutStr, stderrStr string, cmdErr error) string {
	var b strings.Builder

	if stdoutStr != "" {
		b.WriteString("--- stdout ---\n")
		b.WriteString(stdoutStr)
		if !strings.HasSuffix(stdoutStr, "\n") {
			b.WriteString("\n")
		}
	}

	if stderrStr != "" {
		b.WriteString("--- stderr ---\n")
		b.WriteString(stderrStr)
		if !strings.HasSuffix(stderrStr, "\n") {
			b.WriteString("\n")
		}
	}

	if cmdErr != nil {
		if ctxErr := context.DeadlineExceeded; cmdErr == ctxErr || strings.Contains(cmdErr.Error(), "signal: killed") {
			b.WriteString("--- error ---\n")
			b.WriteString("Command timed out\n")
		} else {
			b.WriteString("--- error ---\n")
			b.WriteString(fmt.Sprintf("Exit error: %v\n", cmdErr))
		}
	} else {
		b.WriteString("--- status ---\n")
		b.WriteString("Command completed successfully\n")
	}

	return b.String()
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
