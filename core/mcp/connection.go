package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// jsonRPCRequest represents a JSON-RPC 2.0 request.
type jsonRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// jsonRPCResponse represents a JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// initializeRequest represents the MCP initialize request params.
type initializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]string      `json:"clientInfo"`
}

// initializeResult represents the MCP initialize response.
type initializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]string      `json:"serverInfo"`
}

// toolsListResult represents the result of tools/list.
type toolsListResult struct {
	Tools []struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		InputSchema map[string]interface{} `json:"inputSchema"`
	} `json:"tools"`
}

// toolCallResult represents the result of a tools/call response.
type toolCallResult struct {
	Content []json.RawMessage `json:"content"`
	IsError bool              `json:"isError,omitempty"`
}

// contentBlock represents a content block in the tool response.
type contentBlock struct {
	Type string                 `json:"type"`
	Text string                 `json:"text,omitempty"`
	Data string                 `json:"data,omitempty"`
	MIME string                 `json:"mimeType,omitempty"`
	Meta map[string]interface{} `json:"_meta,omitempty"`
}

// pendingRequest tracks in-flight JSON-RPC requests.
type pendingRequest struct {
	ch chan *jsonRPCResponse
}

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		tokens:     requestsPerSecond,
		maxTokens:  requestsPerSecond,
		refillRate: time.Second,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// Refill tokens based on elapsed time
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= rl.refillRate {
		tokensToAdd := int(elapsed / rl.refillRate)
		if tokensToAdd > 0 {
			rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
			rl.lastRefill = now
		}
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ConnectionManager manages MCP server connections.
type ConnectionManager struct {
	mu           sync.RWMutex
	instances    map[string]*MCPServer       // id -> running instance
	healthCheck  func(ctx context.Context, id string) error
	rateLimiter  map[string]*RateLimiter    // server ID -> rate limiter
	respHandlers map[string]map[int]chan *jsonRPCResponse // serverID -> requestID -> response channel
	nextID       map[string]int            // serverID -> next request ID
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		instances:    make(map[string]*MCPServer),
		rateLimiter:  make(map[string]*RateLimiter),
		respHandlers: make(map[string]map[int]chan *jsonRPCResponse),
		nextID:       make(map[string]int),
	}
}

// SetHealthCheck sets a custom health check function.
func (cm *ConnectionManager) SetHealthCheck(fn func(ctx context.Context, id string) error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.healthCheck = fn
}

// Start launches an MCP server subprocess and initializes the JSON-RPC connection.
func (cm *ConnectionManager) Start(ctx context.Context, config MCPServerConfig) (*MCPServer, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp: context canceled while starting server %q: %w", config.ID, ctx.Err())
	default:
	}

	if err := config.Validate(); err != nil {
		return nil, &ServerError{
			ServerID: config.ID,
			Op:       "start",
			Err:      fmt.Errorf("%w: %v", ErrInvalidConfig, err),
		}
	}

	// Phase 1: Check for duplicates under lock
	cm.mu.Lock()
	if inst, exists := cm.instances[config.ID]; exists {
		status := inst.status()
		if status == StatusRunning {
			cm.mu.Unlock()
			return nil, &ServerError{
				ServerID: config.ID,
				Op:       "start",
				Err:      ErrServerAlreadyRunning,
			}
		}
		// Clean up old stopped/errored instance before restarting
		delete(cm.respHandlers, config.ID)
		delete(cm.nextID, config.ID)
		delete(cm.rateLimiter, config.ID)
		delete(cm.instances, config.ID)
		if inst.stdin != nil {
			_ = inst.stdin.Close()
		}
	}

	// Build the command (outside lock)
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)

	// Set environment variables
	env := os.Environ()
	for k, v := range config.Env {
		if v != "" {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	cmd.Env = env

	// Create stdin/stdout pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cm.mu.Unlock()
		return nil, &ServerError{
			ServerID: config.ID,
			Op:       "start",
			Err:      fmt.Errorf("failed to create stdin pipe: %w", err),
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cm.mu.Unlock()
		return nil, &ServerError{
			ServerID: config.ID,
			Op:       "start",
			Err:      fmt.Errorf("failed to create stdout pipe: %w", err),
		}
	}

	// Redirect stderr to our logs
	cmd.Stderr = os.Stderr

	// Create server instance
	instance := &MCPServer{
		Config:    config,
		Status:    StatusStarting,
		Tools:     []MCPTool{},
		Resources: []MCPResource{},
		cmd:       cmd,
		stdin:     stdin,
		stdout:    bufio.NewScanner(stdout),
		StartedAt: time.Now(),
	}

	// Register under lock
	cm.nextID[config.ID] = 1
	cm.respHandlers[config.ID] = make(map[int]chan *jsonRPCResponse)
	cm.instances[config.ID] = instance
	if config.RateLimit > 0 {
		cm.rateLimiter[config.ID] = NewRateLimiter(config.RateLimit)
	}
	cm.mu.Unlock()

	// Phase 2: Start the process (outside lock)
	if err := cmd.Start(); err != nil {
		instance.Status = StatusError
		instance.Error = fmt.Sprintf("failed to start process: %v", err)
		cm.cleanupInstance(config.ID)
		return nil, &ServerError{
			ServerID: config.ID,
			Op:       "start",
			Err:      fmt.Errorf("failed to start process: %w", err),
		}
	}

	// Start background goroutines
	go cm.readResponses(config.ID, instance)
	go cm.waitForExit(config.ID, instance)

	// Phase 3: Initialize the server (outside lock)
	initResult, err := cm.sendRequestAndWait(ctx, config.ID, "initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    "brain-daemon",
			"version": "0.1.0",
		},
	}, config.Timeout)
	if err != nil {
		_ = cmd.Process.Kill()
		cm.cleanupInstance(config.ID)
		return nil, &ServerError{
			ServerID: config.ID,
			Op:       "start",
			Err:      fmt.Errorf("initialize failed: %w", err),
		}
	}

	// Parse initialize result
	if initResult != nil {
		var initResp initializeResult
		if err := json.Unmarshal(*initResult, &initResp); err == nil {
			if initResp.ServerInfo != nil {
				if name, ok := initResp.ServerInfo["name"]; ok {
					instance.Config.Name = name
				}
				if version, ok := initResp.ServerInfo["version"]; ok {
					instance.Config.Version = version
				}
			}
		}
	}

	// Send initialized notification
	_ = cm.sendNotification(config.ID, "notifications/initialized", map[string]interface{}{})

	// Fetch tools list
	toolsResult, err := cm.sendRequestAndWait(ctx, config.ID, "tools/list", nil, config.Timeout)
	if err != nil {
		instance.Error = fmt.Sprintf("tools/list failed: %v", err)
	} else if toolsResult != nil {
		var toolsResp toolsListResult
		if err := json.Unmarshal(*toolsResult, &toolsResp); err == nil {
			for _, t := range toolsResp.Tools {
				instance.Tools = append(instance.Tools, MCPTool{
					Name:        t.Name,
					Description: t.Description,
					InputSchema: t.InputSchema,
				})
			}
		}
	}

	instance.Status = StatusRunning
	instance.LastChecked = time.Now()

	return instance, nil
}

// Stop stops a running MCP server instance.
func (cm *ConnectionManager) Stop(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("mcp: context canceled while stopping server %q: %w", id, ctx.Err())
	default:
	}

	cm.mu.Lock()
	instance, exists := cm.instances[id]
	if !exists {
		cm.mu.Unlock()
		return &ServerError{
			ServerID: id,
			Op:       "stop",
			Err:      ErrServerNotFound,
		}
	}

	if instance.Status != StatusRunning {
		cm.mu.Unlock()
		if instance.Status == StatusStopped {
			return &ServerError{
				ServerID: id,
				Op:       "stop",
				Err:      ErrServerAlreadyStopped,
			}
		}
		return &ServerError{
			ServerID: id,
			Op:       "stop",
			Err:      fmt.Errorf("server not running (status: %s)", instance.Status),
		}
	}

	// Mark as stopping before releasing lock so other callers see consistent state
	instance.Status = StatusStopped
	cm.mu.Unlock()

	// Perform runtime shutdown outside the main lock but keep the stopped
	// instance available for GetInstance/HealthCheck callers.
	cm.stopInstance(id)

	return nil
}

// stopInstance shuts down a server process without removing the instance record.
// Caller must NOT hold cm.mu.
func (cm *ConnectionManager) stopInstance(id string) {
	cm.mu.Lock()
	instance := cm.instances[id]
	for reqID, ch := range cm.respHandlers[id] {
		close(ch)
		delete(cm.respHandlers[id], reqID)
	}
	delete(cm.respHandlers, id)
	delete(cm.nextID, id)
	delete(cm.rateLimiter, id)
	cm.mu.Unlock()

	if instance == nil {
		return
	}

	if instance.cmd != nil && instance.cmd.Process != nil {
		_ = instance.cmd.Process.Kill()
		_ = instance.cmd.Wait()
	}

	if instance.stdin != nil {
		_ = instance.stdin.Close()
	}
}
// cleanupInstance releases all resources for a server instance.
// Caller must NOT hold cm.mu (it acquires its own lock on the instance).
func (cm *ConnectionManager) cleanupInstance(id string) {
	cm.mu.Lock()
	instance := cm.instances[id]
	// Close all pending response channels
	for reqID, ch := range cm.respHandlers[id] {
		close(ch)
		delete(cm.respHandlers[id], reqID)
	}
	delete(cm.respHandlers, id)
	delete(cm.nextID, id)
	delete(cm.rateLimiter, id)
	delete(cm.instances, id)
	cm.mu.Unlock()

	if instance == nil {
		return
	}

	// Kill the process if still running
	if instance.cmd != nil && instance.cmd.Process != nil {
		_ = instance.cmd.Process.Kill()
		_ = instance.cmd.Wait()
	}

	// Close stdin
	if instance.stdin != nil {
		_ = instance.stdin.Close()
	}
}

// readResponses runs in a goroutine, reading JSON-RPC responses from stdout and routing them.
func (cm *ConnectionManager) readResponses(serverID string, instance *MCPServer) {
	for instance.stdout != nil && instance.stdout.Scan() {
		line := instance.stdout.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Not a valid JSON-RPC response, skip
			continue
		}

		// Route to the waiting request handler
		if resp.ID != nil {
			cm.mu.RLock()
			handlers, ok := cm.respHandlers[serverID]
			cm.mu.RUnlock()

			if ok {
				reqID := *resp.ID
				cm.mu.Lock()
				if ch, exists := handlers[reqID]; exists {
					ch <- &resp
					delete(handlers, reqID)
				}
				cm.mu.Unlock()
			}
		}
	}

	// Scanner hit EOF or error -- process likely exited
	if instance.stdout.Err() != nil {
		instance.mu.Lock()
		instance.Status = StatusError
		instance.Error = fmt.Sprintf("stdout reader error: %v", instance.stdout.Err())
		instance.mu.Unlock()
	}
}

// waitForExit waits for the subprocess to exit and updates status.
func (cm *ConnectionManager) waitForExit(serverID string, instance *MCPServer) {
	if instance.cmd == nil {
		return
	}
	err := instance.cmd.Wait()

	instance.mu.Lock()
	defer instance.mu.Unlock()

	// If we're already stopped (user called Stop()), don't override
	if instance.Status == StatusStopped {
		return
	}

	if err != nil {
		// Process exited with error
		instance.Status = StatusError
		instance.Error = fmt.Sprintf("process exited: %v", err)
	} else {
		// Clean exit
		instance.Status = StatusStopped
	}

	// Clean up channels
	cm.mu.Lock()
	for reqID, ch := range cm.respHandlers[serverID] {
		close(ch)
		delete(cm.respHandlers[serverID], reqID)
	}
	delete(cm.respHandlers, serverID)
	delete(cm.nextID, serverID)
	cm.mu.Unlock()
}

// sendRequestAndWait sends a JSON-RPC request and waits for the response.
func (cm *ConnectionManager) sendRequestAndWait(ctx context.Context, serverID string, method string, params map[string]interface{}, timeout time.Duration) (*json.RawMessage, error) {
	cm.mu.Lock()
	instance, exists := cm.instances[serverID]
	if !exists {
		cm.mu.Unlock()
		return nil, ErrServerNotFound
	}
	status := instance.status()
	if status != StatusRunning && status != StatusStarting {
		cm.mu.Unlock()
		return nil, fmt.Errorf("server %q is not ready (status: %s)", serverID, status)
	}

	// Get next request ID
	reqID := cm.nextID[serverID]
	cm.nextID[serverID]++

	// Create response channel
	respCh := make(chan *jsonRPCResponse, 1)
	cm.respHandlers[serverID][reqID] = respCh
	stdin := instance.stdin
	cm.mu.Unlock()

	// Build the request
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		// Clean up
		cm.mu.Lock()
		delete(cm.respHandlers[serverID], reqID)
		cm.mu.Unlock()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write to stdin
	if _, err := stdin.Write(append(data, '\n')); err != nil {
		cm.mu.Lock()
		delete(cm.respHandlers[serverID], reqID)
		cm.mu.Unlock()
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response with timeout
	select {
	case <-ctx.Done():
		cm.mu.Lock()
		delete(cm.respHandlers[serverID], reqID)
		cm.mu.Unlock()
		return nil, fmt.Errorf("context canceled: %w", ctx.Err())
	case <-time.After(timeout):
		cm.mu.Lock()
		delete(cm.respHandlers[serverID], reqID)
		cm.mu.Unlock()
		return nil, fmt.Errorf("request timed out after %v", timeout)
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, fmt.Errorf("JSON-RPC error: code=%d message=%s", resp.Error.Code, resp.Error.Message)
		}
		return &resp.Result, nil
	}
}

// sendNotification sends a JSON-RPC notification (no response expected).
func (cm *ConnectionManager) sendNotification(serverID string, method string, params map[string]interface{}) error {
	cm.mu.RLock()
	instance, exists := cm.instances[serverID]
	var stdin io.WriteCloser
	if exists && instance != nil {
		stdin = instance.stdin
	}
	cm.mu.RUnlock()

	if !exists || stdin == nil {
		return ErrServerNotFound
	}

	// Notifications have no "id" field
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	if _, err := stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// CallTool invokes a tool on a running MCP server via JSON-RPC tools/call.
func (cm *ConnectionManager) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*ToolResult, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp: context canceled while calling tool %q on server %q: %w", toolName, serverID, ctx.Err())
	default:
	}

	// Check rate limit
	if err := cm.AcquireToken(serverID); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("rate limit exceeded for server %q", serverID),
		}, ErrRateLimitExceeded
	}

	// Get the instance
	cm.mu.RLock()
	instance, exists := cm.instances[serverID]
	if !exists {
		cm.mu.RUnlock()
		return nil, &ServerError{
			ServerID: serverID,
			Op:       "call_tool",
			Err:      ErrServerNotFound,
		}
	}
	if instance.status() != StatusRunning {
		cm.mu.RUnlock()
		return nil, &ServerError{
			ServerID: serverID,
			Op:       "call_tool",
			Err:      fmt.Errorf("server not running (status: %s)", instance.status()),
		}
	}
	cfg := instance.Config
	cm.mu.RUnlock()

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	start := time.Now()

	// Send tools/call request
	result, err := cm.sendRequestAndWait(ctx, serverID, "tools/call", map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}, timeout)

	duration := time.Since(start)

	if err != nil {
		return &ToolResult{
			Success:  false,
			Error:    fmt.Sprintf("tools/call failed: %v", err),
			Duration: duration,
		}, &ServerError{
			ServerID: serverID,
			Op:       "call_tool",
			Err:      fmt.Errorf("%w: %v", ErrToolCallFailed, err),
		}
	}

	if result == nil {
		return &ToolResult{
			Success:  false,
			Error:    "tools/call returned nil result",
			Duration: duration,
		}, &ServerError{
			ServerID: serverID,
			Op:       "call_tool",
			Err:      ErrToolCallFailed,
		}
	}

	// Parse the tool call result
	var toolResp toolCallResult
	if err := json.Unmarshal(*result, &toolResp); err != nil {
		return &ToolResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to parse tool result: %v", err),
			Duration: duration,
		}, &ServerError{
			ServerID: serverID,
			Op:       "call_tool",
			Err:      fmt.Errorf("%w: %v", ErrToolCallFailed, err),
		}
	}

	if toolResp.IsError {
		// Extract text content for error message
		var errMsg string
		for _, c := range toolResp.Content {
			var block contentBlock
			if err := json.Unmarshal(c, &block); err == nil && block.Text != "" {
				errMsg = block.Text
				break
			}
		}
		if errMsg == "" {
			errMsg = string(*result)
		}
		return &ToolResult{
			Success:  false,
			Error:    errMsg,
			Content:  *result,
			Duration: duration,
		}, &ServerError{
			ServerID: serverID,
			Op:       "call_tool",
			Err:      fmt.Errorf("%w: %s", ErrToolCallFailed, errMsg),
		}
	}

	// Build content bytes from all content blocks
	var contentParts []string
	for _, c := range toolResp.Content {
		var block contentBlock
		if err := json.Unmarshal(c, &block); err == nil && block.Text != "" {
			contentParts = append(contentParts, block.Text)
		}
	}

	contentBytes := *result
	if len(contentParts) > 0 {
		// Return a simpler text representation
		contentBytes = []byte(contentParts[0])
	}

	return &ToolResult{
		Success:  true,
		Content:  contentBytes,
		Duration: duration,
	}, nil
}

// GetInstance retrieves a running MCP server instance.
func (cm *ConnectionManager) GetInstance(ctx context.Context, id string) (*MCPServer, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp: context canceled while getting instance %q: %w", id, ctx.Err())
	default:
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	instance, exists := cm.instances[id]
	if !exists {
		return nil, &ServerError{
			ServerID: id,
			Op:       "get_instance",
			Err:      ErrServerNotFound,
		}
	}

	// Return a copy (without runtime fields which are not safe to copy)
	instanceCopy := &MCPServer{
		Config:      instance.Config,
		Status:      instance.status(),
		Tools:       instance.Tools,
		Resources:   instance.Resources,
		Health:      instance.Health,
		LastChecked: instance.LastChecked,
		ClientCount: instance.ClientCount,
		StartedAt:   instance.StartedAt,
		Error:       instance.Error,
	}
	return instanceCopy, nil
}

// ListInstances returns all registered MCP server instances (running or not).
func (cm *ConnectionManager) ListInstances(ctx context.Context) []*MCPServer {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*MCPServer, 0, len(cm.instances))
	for _, inst := range cm.instances {
		instCopy := &MCPServer{
			Config:      inst.Config,
			Status:      inst.status(),
			Tools:       inst.Tools,
			Resources:   inst.Resources,
			Health:      inst.Health,
			LastChecked: inst.LastChecked,
			ClientCount: inst.ClientCount,
			StartedAt:   inst.StartedAt,
			Error:       inst.Error,
		}
		result = append(result, instCopy)
	}

	return result
}

// HealthCheck performs a health check on a specific server.
func (cm *ConnectionManager) HealthCheck(ctx context.Context, id string) (*HealthCheck, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp: context canceled during health check %q: %w", id, ctx.Err())
	default:
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	instance, exists := cm.instances[id]
	if !exists {
		return nil, &ServerError{
			ServerID: id,
			Op:       "healthcheck",
			Err:      ErrServerNotFound,
		}
	}

	start := time.Now()
	var healthErr error

	// Use custom health check if set
	if cm.healthCheck != nil {
		healthErr = cm.healthCheck(ctx, id)
	}

	latency := time.Since(start)

	status := instance.status()
	if status == StatusRunning && healthErr == nil {
		instance.Health = &HealthCheck{
			Healthy:   true,
			Latency:   latency,
			LastCheck: time.Now(),
		}
		instance.LastChecked = time.Now()
		return instance.Health, nil
	}

	errMsg := ""
	if healthErr != nil {
		errMsg = healthErr.Error()
	}
	if status != StatusRunning {
		errMsg = fmt.Sprintf("server status: %s", status)
	}

	instance.Health = &HealthCheck{
		Healthy:   false,
		Latency:   latency,
		LastCheck: time.Now(),
		Error:     errMsg,
	}
	instance.LastChecked = time.Now()

	return instance.Health, nil
}

// HealthCheckAll performs health checks on all servers.
func (cm *ConnectionManager) HealthCheckAll(ctx context.Context) map[string]*HealthCheck {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	results := make(map[string]*HealthCheck)

	// Get all server IDs
	cm.mu.RLock()
	ids := make([]string, 0, len(cm.instances))
	for id := range cm.instances {
		ids = append(ids, id)
	}
	cm.mu.RUnlock()

	// Check each server
	for _, id := range ids {
		hc, err := cm.HealthCheck(ctx, id)
		if err != nil {
			results[id] = &HealthCheck{
				Healthy:   false,
				LastCheck: time.Now(),
				Error:     err.Error(),
			}
		} else {
			results[id] = hc
		}
	}

	return results
}

// SetRateLimit configures rate limiting for a server.
func (cm *ConnectionManager) SetRateLimit(serverID string, requestsPerSecond int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.rateLimiter[serverID] = NewRateLimiter(requestsPerSecond)
}

// AcquireToken attempts to acquire a token for rate limiting.
// Returns ErrRateLimitExceeded if the server is rate limited.
func (cm *ConnectionManager) AcquireToken(serverID string) error {
	cm.mu.RLock()
	limiter, exists := cm.rateLimiter[serverID]
	cm.mu.RUnlock()

	if !exists {
		// No rate limiter configured, allow request
		return nil
	}

	if !limiter.Allow() {
		return ErrRateLimitExceeded
	}

	return nil
}

// status returns the current server status in a thread-safe way.
func (s *MCPServer) status() MCPServerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Status
}
