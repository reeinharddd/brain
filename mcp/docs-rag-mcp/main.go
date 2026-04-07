package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/reeinharrrd/brain/mcp/docs-rag-mcp/internal/indexer"
	"github.com/reeinharrrd/brain/mcp/docs-rag-mcp/internal/tools"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPServer represents the MCP server instance.
type MCPServer struct {
	indexer   *indexer.Indexer
	brainRoot string
	brainEnv  string
	logger    *log.Logger
}

// NewMCPServer creates a new MCP server.
func NewMCPServer(brainRoot string) (*MCPServer, error) {
	if brainRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		brainRoot = homeDir + "/.brain"
	}

	idx, err := indexer.NewIndexer(brainRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	brainEnv := os.Getenv("BRAIN_ENV")
	if brainEnv == "" {
		brainEnv = "development"
	}

	return &MCPServer{
		indexer:   idx,
		brainRoot: brainRoot,
		brainEnv:  brainEnv,
		logger:    log.New(os.Stderr, "[docs-rag-mcp] ", log.LstdFlags),
	}, nil
}

// Start begins the MCP server listening on stdin/stdout.
func (s *MCPServer) Start() error {
	s.logger.Printf("Docs-RAG MCP Server starting (BRAIN_ENV=%s)\n", s.brainEnv)
	s.logger.Printf("Brain Root: %s\n", s.brainRoot)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Parse JSON-RPC request
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.respond(&JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
					Data:    err.Error(),
				},
			})
			continue
		}

		// Handle the request
		resp := s.handleRequest(&req)
		
		// Only respond if this is a request (has an id), not a notification
		if req.ID != nil {
			s.respond(resp)
		}
	}

	return scanner.Err()
}

// handleRequest processes a JSON-RPC request.
func (s *MCPServer) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = s.initialize()
		return resp

	case "initialized":
		// This is a notification from the client - don't respond
		return nil

	case "tools/list":
		resp.Result = s.listTools()
		return resp

	case "tools/call":
		result, err := s.callTool(ctx, req.Params)
		if err != nil {
			resp.Error = err
		} else {
			resp.Result = result
		}
		return resp

	default:
		resp.Error = &JSONRPCError{
			Code:    -32601,
			Message: "Method not found",
		}
		return resp
	}
}

// initialize performs the MCP handshake and returns server capabilities.
func (s *MCPServer) initialize() map[string]interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "brain-docs-rag",
			"version": "1.0.0",
		},
	}
}

// listTools returns available MCP tools.
func (s *MCPServer) listTools() map[string]interface{} {
	return map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "docs_search",
				"description": "Search Brain documentation using semantic query",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Max results (default 10)",
						},
						"domain": map[string]interface{}{
							"type":        "string",
							"description": "Optional domain filter",
						},
					},
					"required": []string{"query"},
				},
			},
			{
				"name":        "docs_status",
				"description": "Get documentation index status",
				"inputSchema": map[string]interface{}{
					"type": "object",
				},
			},
			{
				"name":        "docs_rebuild",
				"description": "Rebuild documentation index (dev-only)",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"domains": map[string]interface{}{
							"type":        "array",
							"items": map[string]interface{}{
								"type": "string",
							},
							"description": "Optional domain filters",
						},
					},
				},
			},
		},
	}
}

// callTool calls a tool by name.
func (s *MCPServer) callTool(ctx context.Context, params json.RawMessage) (interface{}, *JSONRPCError) {
	// Support both "params" and "arguments" field names for compatibility
	var toolCall struct {
		Name      string          `json:"name"`
		Params    json.RawMessage `json:"params"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(params, &toolCall); err != nil {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params"}
	}

	// Use arguments if params is empty (support both MCP spec variations)
	toolParams := toolCall.Params
	if len(toolParams) == 0 && len(toolCall.Arguments) > 0 {
		toolParams = toolCall.Arguments
	}

	switch toolCall.Name {
	case "docs_search":
		var req tools.SearchRequest
		if err := json.Unmarshal(toolParams, &req); err != nil {
			return nil, &JSONRPCError{Code: -32602, Message: "Invalid tool params"}
		}
		return tools.DocsSearch(ctx, s.indexer, req), nil

	case "docs_status":
		return tools.DocsStatus(ctx, s.indexer), nil

	case "docs_rebuild":
		return tools.DocsRebuild(ctx, s.indexer, s.brainEnv), nil

	default:
		return nil, &JSONRPCError{Code: -32601, Message: "Tool not found"}
	}
}

// respond writes a JSON-RPC response to stdout.
func (s *MCPServer) respond(resp *JSONRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		s.logger.Printf("Error marshaling response: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func main() {
	// Get brain root from environment or use default
	brainRoot := os.Getenv("BRAIN_ROOT")
	if brainRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		brainRoot = homeDir + "/.brain"
	}

	// Create and start server
	server, err := NewMCPServer(brainRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
