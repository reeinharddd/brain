// Package mcp provides MCP (Model Context Protocol) server management,
// including server definitions, registry, connection management, and proxy routing.
package mcp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// MCPServerStatus represents server operational state.
type MCPServerStatus string

const (
	StatusRunning  MCPServerStatus = "running"
	StatusStopped  MCPServerStatus = "stopped"
	StatusError    MCPServerStatus = "error"
	StatusStarting MCPServerStatus = "starting"
)

// TransportType defines how the server communicates.
type TransportType string

const (
	TransportStdIO TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

// MCPTool represents a tool exposed by an MCP server.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"` // JSON schema
}

// MCPResource represents a resource exposed by an MCP server.
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// MCPServerConfig holds server configuration.
type MCPServerConfig struct {
	ID          string
	Name        string
	Version     string
	Description string
	Category    string // official, community, enterprise, private
	Transport   TransportType
	Command     string
	Args        []string
	Env         map[string]string
	Timeout     time.Duration
	RateLimit   int // requests per second
}

// Validate checks if the configuration is valid.
func (c MCPServerConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("mcp: server ID is required")
	}
	if c.Name == "" {
		return fmt.Errorf("mcp: server name is required")
	}
	if c.Command == "" {
		return fmt.Errorf("mcp: server command is required")
	}
	return nil
}

// MCPServer represents a running MCP server instance.
type MCPServer struct {
	Config      MCPServerConfig
	Status      MCPServerStatus
	Tools       []MCPTool
	Resources   []MCPResource
	Health      *HealthCheck
	LastChecked time.Time
	ClientCount int // number of connected clients
	StartedAt   time.Time
	Error       string
	// Runtime fields for stdio subprocess management
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	mu     sync.Mutex // protects concurrent stdio access
}

// HealthCheck represents server health status.
type HealthCheck struct {
	Healthy   bool
	Latency   time.Duration
	LastCheck time.Time
	Error     string
}

// Custom error types for domain errors.
var (
	ErrServerNotFound       = errors.New("mcp: server not found")
	ErrServerAlreadyRunning = errors.New("mcp: server is already running")
	ErrServerAlreadyStopped = errors.New("mcp: server is already stopped")
	ErrDuplicateServer      = errors.New("mcp: server with this ID already exists")
	ErrRateLimitExceeded    = errors.New("mcp: rate limit exceeded")
	ErrInvalidConfig        = errors.New("mcp: invalid server configuration")
	ErrToolCallFailed       = errors.New("mcp: tool call failed")
	ErrContextCanceled      = errors.New("mcp: context canceled")
)

// ServerError wraps an error with server context.
type ServerError struct {
	ServerID string
	Op       string // operation that failed
	Err      error
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("mcp: %s on server %q: %v", e.Op, e.ServerID, e.Err)
}

func (e *ServerError) Unwrap() error {
	return e.Err
}
