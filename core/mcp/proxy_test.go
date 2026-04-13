package mcp

import (
	"context"
	"testing"
	"time"
)

func setupProxy(t *testing.T) (*MCPProxy, *ConnectionManager, *MCPRegistry) {
	t.Helper()
	ctx := context.Background()
	cm := NewConnectionManager()
	reg := NewMCPRegistry()

	cfg := MCPServerConfig{
		ID:        "proxy-test",
		Name:      "Proxy Test Server",
		Version:   "1.0.0",
		Category:  "test",
		Transport: TransportStdIO,
		Command:   "brain-mcp-filesystem",
		Args:      []string{},
		Timeout:   30 * time.Second,
	}

	if err := reg.Register(ctx, cfg); err != nil {
		t.Fatalf("failed to register server: %v", err)
	}

	inst, err := cm.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	inst.Tools = []MCPTool{
		{Name: "tool1", Description: "Test tool 1", InputSchema: map[string]interface{}{"type": "object"}},
		{Name: "tool2", Description: "Test tool 2", InputSchema: map[string]interface{}{"type": "object"}},
	}

	proxy := NewMCPProxy(cm, reg)
	return proxy, cm, reg
}

func TestMCPProxy_CallTool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode: requires real MCP server binaries")
	}
	ctx := context.Background()

	t.Run("successful tool call", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "proxy-test",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{"key": "value"},
		}

		result, err := proxy.CallTool(ctx, call)
		if err != nil {
			t.Fatalf("CallTool() error = %v", err)
		}
		if !result.Success {
			t.Errorf("CallTool() Success = false, want true")
		}
		if len(result.Content) == 0 {
			t.Error("CallTool() Content is empty")
		}
		if result.Duration <= 0 {
			t.Error("CallTool() Duration should be positive")
		}
	})

	t.Run("unknown server", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "unknown-server",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		result, err := proxy.CallTool(ctx, call)
		if err == nil {
			t.Fatal("CallTool() expected error, got nil")
		}
		if result.Success {
			t.Error("CallTool() Success = true, want false for unknown server")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("CallTool() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerNotFound {
			t.Errorf("CallTool() error = %v, want %v", serverErr.Err, ErrServerNotFound)
		}
	})

	t.Run("tool not found", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "proxy-test",
			ToolName:  "nonexistent-tool",
			Arguments: map[string]interface{}{},
		}

		result, err := proxy.CallTool(ctx, call)
		if err == nil {
			t.Fatal("CallTool() expected error, got nil")
		}
		if result.Success {
			t.Error("CallTool() Success = true, want false for unknown tool")
		}
	})

	t.Run("server not running", func(t *testing.T) {
		proxy, cm, reg := setupProxy(t)

		// Stop the server
		if err := cm.Stop(ctx, "proxy-test"); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}

		// Re-register with same ID to keep in registry
		_ = reg

		call := ToolCall{
			ServerID:  "proxy-test",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		result, err := proxy.CallTool(ctx, call)
		if err == nil {
			t.Fatal("CallTool() expected error, got nil")
		}
		if result.Success {
			t.Error("CallTool() Success = true, want false for stopped server")
		}
	})
}

func TestMCPProxy_ListTools(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode: requires real MCP server binaries")
	}
	ctx := context.Background()

	t.Run("aggregate tools from multiple servers", func(t *testing.T) {
		cm := NewConnectionManager()
		reg := NewMCPRegistry()

		// Register two servers
		cfg1 := MCPServerConfig{
			ID:        "list-tools-1",
			Name:      "Server 1",
			Version:   "1.0.0",
			Category:  "test",
			Transport: TransportStdIO,
			Command:   "test1",
			Timeout:   30 * time.Second,
		}
		cfg2 := MCPServerConfig{
			ID:        "list-tools-2",
			Name:      "Server 2",
			Version:   "1.0.0",
			Category:  "test",
			Transport: TransportStdIO,
			Command:   "test2",
			Timeout:   30 * time.Second,
		}

		_ = reg.Register(ctx, cfg1)
		_ = reg.Register(ctx, cfg2)

		inst1, _ := cm.Start(ctx, cfg1)
		inst1.Tools = []MCPTool{
			{Name: "tool_a", Description: "Tool A", InputSchema: map[string]interface{}{}},
			{Name: "tool_b", Description: "Tool B", InputSchema: map[string]interface{}{}},
		}

		inst2, _ := cm.Start(ctx, cfg2)
		inst2.Tools = []MCPTool{
			{Name: "tool_b", Description: "Tool B", InputSchema: map[string]interface{}{}}, // Duplicate
			{Name: "tool_c", Description: "Tool C", InputSchema: map[string]interface{}{}},
		}

		proxy := NewMCPProxy(cm, reg)

		tools := proxy.ListTools(ctx)
		// Should deduplicate tool_b
		if len(tools) != 3 {
			t.Errorf("ListTools() count = %v, want 3", len(tools))
		}
	})

	t.Run("empty proxy", func(t *testing.T) {
		cm := NewConnectionManager()
		reg := NewMCPRegistry()
		proxy := NewMCPProxy(cm, reg)

		tools := proxy.ListTools(ctx)
		if len(tools) != 0 {
			t.Errorf("ListTools() count = %v, want 0", len(tools))
		}
	})
}

func TestMCPProxy_Metrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()

	t.Run("track successful calls", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "proxy-test",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		// Make 3 successful calls
		for i := 0; i < 3; i++ {
			_, err := proxy.CallTool(ctx, call)
			if err != nil {
				t.Fatalf("CallTool() error = %v", err)
			}
		}

		metrics := proxy.GetMetrics(ctx)
		if metrics.TotalCalls != 3 {
			t.Errorf("GetMetrics() TotalCalls = %v, want 3", metrics.TotalCalls)
		}
		if metrics.SuccessfulCalls != 3 {
			t.Errorf("GetMetrics() SuccessfulCalls = %v, want 3", metrics.SuccessfulCalls)
		}
		if metrics.FailedCalls != 0 {
			t.Errorf("GetMetrics() FailedCalls = %v, want 0", metrics.FailedCalls)
		}
		if metrics.ByServer["proxy-test"] != 3 {
			t.Errorf("GetMetrics() ByServer[proxy-test] = %v, want 3", metrics.ByServer["proxy-test"])
		}
	})

	t.Run("track failed calls", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "unknown",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		// Make 2 failed calls
		for i := 0; i < 2; i++ {
			_, _ = proxy.CallTool(ctx, call)
		}

		metrics := proxy.GetMetrics(ctx)
		if metrics.TotalCalls != 2 {
			t.Errorf("GetMetrics() TotalCalls = %v, want 2", metrics.TotalCalls)
		}
		if metrics.FailedCalls != 2 {
			t.Errorf("GetMetrics() FailedCalls = %v, want 2", metrics.FailedCalls)
		}
	})

	t.Run("reset metrics", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "proxy-test",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		// Make some calls
		_, _ = proxy.CallTool(ctx, call)

		// Reset
		proxy.ResetMetrics(ctx)

		metrics := proxy.GetMetrics(ctx)
		if metrics.TotalCalls != 0 {
			t.Errorf("GetMetrics() after reset TotalCalls = %v, want 0", metrics.TotalCalls)
		}
		if metrics.SuccessfulCalls != 0 {
			t.Errorf("GetMetrics() after reset SuccessfulCalls = %v, want 0", metrics.SuccessfulCalls)
		}
		if metrics.FailedCalls != 0 {
			t.Errorf("GetMetrics() after reset FailedCalls = %v, want 0", metrics.FailedCalls)
		}
		if len(metrics.ByServer) != 0 {
			t.Errorf("GetMetrics() after reset ByServer = %v, want empty", metrics.ByServer)
		}
	})

	t.Run("average latency calculation", func(t *testing.T) {
		proxy, _, _ := setupProxy(t)

		call := ToolCall{
			ServerID:  "proxy-test",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		// Make multiple calls
		for i := 0; i < 5; i++ {
			_, _ = proxy.CallTool(ctx, call)
		}

		metrics := proxy.GetMetrics(ctx)
		if metrics.AvgLatency <= 0 {
			t.Error("GetMetrics() AvgLatency should be positive after calls")
		}
	})
}

func TestMCPProxy_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cm := NewConnectionManager()
	reg := NewMCPRegistry()
	proxy := NewMCPProxy(cm, reg)

	t.Run("CallTool with canceled context", func(t *testing.T) {
		call := ToolCall{
			ServerID:  "any",
			ToolName:  "tool1",
			Arguments: map[string]interface{}{},
		}

		_, err := proxy.CallTool(ctx, call)
		if err == nil {
			t.Fatal("CallTool() expected error with canceled context")
		}
	})

	t.Run("ListTools with canceled context", func(t *testing.T) {
		tools := proxy.ListTools(ctx)
		if tools != nil {
			t.Errorf("ListTools() = %v, want nil with canceled context", tools)
		}
	})

	t.Run("GetMetrics with canceled context", func(t *testing.T) {
		metrics := proxy.GetMetrics(ctx)
		if metrics != nil {
			t.Errorf("GetMetrics() = %v, want nil with canceled context", metrics)
		}
	})

	t.Run("ResetMetrics with canceled context", func(t *testing.T) {
		// Should not panic
		proxy.ResetMetrics(ctx)
	})
}
