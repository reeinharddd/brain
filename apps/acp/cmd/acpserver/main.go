package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ACPRequest represents an incoming ACP request
type ACPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// ACPResponse represents an ACP response
type ACPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ACPError   `json:"error,omitempty"`
}

// ACPError represents an ACP error
type ACPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server handles ACP protocol requests
type Server struct {
	daemonURL string
	scanner   *bufio.Scanner
}

func NewServer(daemonURL string) *Server {
	return &Server{
		daemonURL: daemonURL,
		scanner:   bufio.NewScanner(os.Stdin),
	}
}

func (s *Server) Run() error {
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}

		var req ACPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(0, -32700, "Parse error")
			continue
		}

		result, err := s.handleMethod(req.Method, req.Params)
		if err != nil {
			s.sendError(req.ID, -32603, err.Error())
		} else {
			s.sendResult(req.ID, result)
		}
	}
	return s.scanner.Err()
}

func (s *Server) handleMethod(method string, params map[string]interface{}) (interface{}, error) {
	switch method {
	case "initialize":
		return map[string]interface{}{
			"serverInfo": map[string]string{
				"name":    "brain-acp-server",
				"version": "0.1.0",
			},
			"capabilities": map[string]bool{
				"tools":   true,
				"context": true,
				"policy":  true,
				"memory":  true,
			},
		}, nil
	case "tools/list":
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "brain_resolve_artifact",
					"description": "Resolve artifact via Brain daemon",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"kind": map[string]string{"type": "string"},
							"id":   map[string]string{"type": "string"},
						},
						"required": []string{"kind", "id"},
					},
				},
				{
					"name":        "brain_get_context",
					"description": "Get compiled context bundle",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"scope_chain": map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
						},
					},
				},
				{
					"name":        "brain_check_policy",
					"description": "Check policy for an action",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"action": map[string]string{"type": "string"},
							"scope":  map[string]string{"type": "string"},
						},
						"required": []string{"action"},
					},
				},
			},
		}, nil
	case "tools/call":
		toolName, _ := params["name"].(string)
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("Called tool: %s", toolName)},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func (s *Server) sendResult(id int, result interface{}) {
	resp := ACPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeResponse(resp)
}

func (s *Server) sendError(id int, code int, message string) {
	resp := ACPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ACPError{
			Code:    code,
			Message: message,
		},
	}
	s.writeResponse(resp)
}

func (s *Server) writeResponse(resp ACPResponse) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", string(data))
}

func main() {
	daemonURL := "http://localhost:8080"
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--daemon-url=") {
			daemonURL = strings.TrimPrefix(arg, "--daemon-url=")
		}
	}

	server := NewServer(daemonURL)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
