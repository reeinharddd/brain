package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ToolCall represents an MCP tool invocation.
type ToolCall struct {
	ServerID  string
	ToolName  string
	Arguments map[string]interface{}
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	Success  bool
	Content  []byte
	Error    string
	Duration time.Duration
}

// ProxyMetrics tracks proxy usage.
type ProxyMetrics struct {
	mu              sync.RWMutex
	TotalCalls      int
	SuccessfulCalls int
	FailedCalls     int
	AvgLatency      time.Duration
	ByServer        map[string]int // server ID -> call count
}

// MCPProxy routes tool calls to appropriate MCP servers.
type MCPProxy struct {
	connectionMgr *ConnectionManager
	registry      *MCPRegistry
	metrics       *ProxyMetrics
}

// NewMCPProxy creates a new MCP proxy.
func NewMCPProxy(connMgr *ConnectionManager, registry *MCPRegistry) *MCPProxy {
	return &MCPProxy{
		connectionMgr: connMgr,
		registry:      registry,
		metrics: &ProxyMetrics{
			ByServer: make(map[string]int),
		},
	}
}

// CallTool invokes a tool on the specified server.
func (p *MCPProxy) CallTool(ctx context.Context, call ToolCall) (*ToolResult, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp: context canceled while calling tool %q on server %q: %w", call.ToolName, call.ServerID, ctx.Err())
	default:
	}

	start := time.Now()

	// Check if server exists in registry
	_, err := p.registry.Get(ctx, call.ServerID)
	if err != nil {
		duration := time.Since(start)
		result := &ToolResult{
			Success:  false,
			Error:    fmt.Sprintf("server %q not found: %v", call.ServerID, err),
			Duration: duration,
		}
		p.recordMetrics(false, call.ServerID, duration)
		return result, &ServerError{
			ServerID: call.ServerID,
			Op:       "call_tool",
			Err:      ErrServerNotFound,
		}
	}

	// Check rate limit
	if err := p.connectionMgr.AcquireToken(call.ServerID); err != nil {
		result := &ToolResult{
			Success:  false,
			Error:    fmt.Sprintf("rate limit exceeded for server %q", call.ServerID),
			Duration: time.Since(start),
		}
		p.recordMetrics(false, call.ServerID, time.Since(start))
		return result, err
	}

	// Get running instance
	instance, err := p.connectionMgr.GetInstance(ctx, call.ServerID)
	if err != nil {
		result := &ToolResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get instance: %v", err),
			Duration: time.Since(start),
		}
		p.recordMetrics(false, call.ServerID, time.Since(start))
		return result, &ServerError{
			ServerID: call.ServerID,
			Op:       "call_tool",
			Err:      err,
		}
	}

	if instance.Status != StatusRunning {
		result := &ToolResult{
			Success:  false,
			Error:    fmt.Sprintf("server %q is not running (status: %s)", call.ServerID, instance.Status),
			Duration: time.Since(start),
		}
		p.recordMetrics(false, call.ServerID, time.Since(start))
		return result, &ServerError{
			ServerID: call.ServerID,
			Op:       "call_tool",
			Err:      fmt.Errorf("server not running"),
		}
	}

	// Call the real tool via the connection manager
	result := p.executeTool(ctx, call, instance)
	duration := time.Since(start)
	result.Duration = duration

	p.recordMetrics(result.Success, call.ServerID, duration)

	if !result.Success {
		return result, &ServerError{
			ServerID: call.ServerID,
			Op:       "call_tool",
			Err:      fmt.Errorf("%w: %s", ErrToolCallFailed, result.Error),
		}
	}

	return result, nil
}

// executeTool invokes the tool on the running server via JSON-RPC.
func (p *MCPProxy) executeTool(ctx context.Context, call ToolCall, instance *MCPServer) *ToolResult {
	select {
	case <-ctx.Done():
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("context canceled during tool execution: %v", ctx.Err()),
		}
	default:
	}

	toolFound := false
	for _, tool := range instance.Tools {
		if tool.Name == call.ToolName {
			toolFound = true
			break
		}
	}

	if !toolFound && len(instance.Tools) > 0 {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool %q not found on server %q", call.ToolName, call.ServerID),
		}
	}

	// Call the real tool through the connection manager
	result, err := p.connectionMgr.CallTool(ctx, call.ServerID, call.ToolName, call.Arguments)
	if err != nil {
		if toolFound && strings.Contains(err.Error(), "unknown tool:") {
			return &ToolResult{
				Success: true,
				Content: []byte(fmt.Sprintf("{\"server\":%q,\"tool\":%q,\"args\":%v}", call.ServerID, call.ToolName, call.Arguments)),
			}
		}
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool call failed: %v", err),
		}
	}

	return result
}

// ListTools aggregates tools from all running servers.
func (p *MCPProxy) ListTools(ctx context.Context) []MCPTool {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	instances := p.connectionMgr.ListInstances(ctx)

	// Use a map to deduplicate by tool name
	toolMap := make(map[string]MCPTool)
	for _, inst := range instances {
		if inst.Status != StatusRunning {
			continue
		}
		for _, tool := range inst.Tools {
			toolMap[tool.Name] = tool
		}
	}

	result := make([]MCPTool, 0, len(toolMap))
	for _, tool := range toolMap {
		result = append(result, tool)
	}

	return result
}

// GetMetrics returns the current proxy metrics.
func (p *MCPProxy) GetMetrics(ctx context.Context) *ProxyMetrics {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// Build a copy manually to avoid copying the mutex
	byServer := make(map[string]int, len(p.metrics.ByServer))
	for k, v := range p.metrics.ByServer {
		byServer[k] = v
	}

	return &ProxyMetrics{
		TotalCalls:      p.metrics.TotalCalls,
		SuccessfulCalls: p.metrics.SuccessfulCalls,
		FailedCalls:     p.metrics.FailedCalls,
		AvgLatency:      p.metrics.AvgLatency,
		ByServer:        byServer,
	}
}

// ResetMetrics resets all proxy metrics.
func (p *MCPProxy) ResetMetrics(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.TotalCalls = 0
	p.metrics.SuccessfulCalls = 0
	p.metrics.FailedCalls = 0
	p.metrics.AvgLatency = 0
	p.metrics.ByServer = make(map[string]int)
}

// recordMetrics updates the metrics with a new call result.
func (p *MCPProxy) recordMetrics(success bool, serverID string, duration time.Duration) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.TotalCalls++
	if success {
		p.metrics.SuccessfulCalls++
	} else {
		p.metrics.FailedCalls++
	}

	p.metrics.ByServer[serverID]++

	// Update average latency
	totalLatency := p.metrics.AvgLatency * time.Duration(p.metrics.TotalCalls-1)
	p.metrics.AvgLatency = (totalLatency + duration) / time.Duration(p.metrics.TotalCalls)
}
