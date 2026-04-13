package mcp

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestConnectionManager_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode: requires real MCP server binaries")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	t.Run("start valid server", func(t *testing.T) {
		cfg := testConfig("start1", "Start Test", "official")
		inst, err := cm.Start(ctx, cfg)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if inst.Status != StatusRunning {
			t.Errorf("Start() Status = %v, want %v", inst.Status, StatusRunning)
		}
		if inst.Config.ID != "start1" {
			t.Errorf("Start() Config.ID = %v, want start1", inst.Config.ID)
		}
		if inst.StartedAt.IsZero() {
			t.Error("Start() StartedAt should be set")
		}
	})

	t.Run("start duplicate server", func(t *testing.T) {
		cfg := testConfig("start-dup", "Start Dup", "official")
		_, err := cm.Start(ctx, cfg)
		if err != nil {
			t.Fatalf("first Start() error = %v", err)
		}

		_, err = cm.Start(ctx, cfg)
		if err == nil {
			t.Fatal("second Start() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("Start() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerAlreadyRunning {
			t.Errorf("Start() error = %v, want %v", serverErr.Err, ErrServerAlreadyRunning)
		}
	})

	t.Run("start with invalid config", func(t *testing.T) {
		invalidCfg := MCPServerConfig{
			ID:   "",
			Name: "No ID",
		}
		_, err := cm.Start(ctx, invalidCfg)
		if err == nil {
			t.Fatal("Start() expected error, got nil")
		}
	})
}

func TestConnectionManager_Stop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	cfg := testConfig("stop1", "Stop Test", "official")
	_, err := cm.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	t.Run("stop running server", func(t *testing.T) {
		if err := cm.Stop(ctx, "stop1"); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		inst, err := cm.GetInstance(ctx, "stop1")
		if err != nil {
			t.Fatalf("GetInstance() error = %v", err)
		}
		if inst.Status != StatusStopped {
			t.Errorf("Stop() Status = %v, want %v", inst.Status, StatusStopped)
		}
	})

	t.Run("stop already stopped server", func(t *testing.T) {
		err := cm.Stop(ctx, "stop1")
		if err == nil {
			t.Fatal("Stop() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("Stop() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerAlreadyStopped {
			t.Errorf("Stop() error = %v, want %v", serverErr.Err, ErrServerAlreadyStopped)
		}
	})

	t.Run("stop non-existent server", func(t *testing.T) {
		err := cm.Stop(ctx, "nonexistent")
		if err == nil {
			t.Fatal("Stop() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("Stop() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerNotFound {
			t.Errorf("Stop() error = %v, want %v", serverErr.Err, ErrServerNotFound)
		}
	})
}

func TestConnectionManager_GetInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	cfg := testConfig("get-inst", "Get Instance Test", "official")
	_, err := cm.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	t.Run("get existing instance", func(t *testing.T) {
		inst, err := cm.GetInstance(ctx, "get-inst")
		if err != nil {
			t.Fatalf("GetInstance() error = %v", err)
		}
		if inst.Config.ID != "get-inst" {
			t.Errorf("GetInstance() Config.ID = %v, want get-inst", inst.Config.ID)
		}
	})

	t.Run("get non-existent instance", func(t *testing.T) {
		_, err := cm.GetInstance(ctx, "nonexistent")
		if err == nil {
			t.Fatal("GetInstance() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("GetInstance() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerNotFound {
			t.Errorf("GetInstance() error = %v, want %v", serverErr.Err, ErrServerNotFound)
		}
	})
}

func TestConnectionManager_ListInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	configs := []MCPServerConfig{
		testConfig("list-inst1", "List 1", "official"),
		testConfig("list-inst2", "List 2", "community"),
	}

	for _, cfg := range configs {
		if _, err := cm.Start(ctx, cfg); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	}

	t.Run("list all instances", func(t *testing.T) {
		instances := cm.ListInstances(ctx)
		if len(instances) != 2 {
			t.Fatalf("ListInstances() count = %v, want 2", len(instances))
		}
	})

	t.Run("empty manager", func(t *testing.T) {
		emptyCM := NewConnectionManager()
		instances := emptyCM.ListInstances(ctx)
		if len(instances) != 0 {
			t.Errorf("ListInstances() count = %v, want 0", len(instances))
		}
	})
}

func TestConnectionManager_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	cfg := testConfig("hc1", "Health Check Test", "official")
	_, err := cm.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	t.Run("healthy server", func(t *testing.T) {
		hc, err := cm.HealthCheck(ctx, "hc1")
		if err != nil {
			t.Fatalf("HealthCheck() error = %v", err)
		}
		if !hc.Healthy {
			t.Error("HealthCheck() Healthy = false, want true for running server")
		}
	})

	t.Run("non-existent server", func(t *testing.T) {
		_, err := cm.HealthCheck(ctx, "nonexistent")
		if err == nil {
			t.Fatal("HealthCheck() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("HealthCheck() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerNotFound {
			t.Errorf("HealthCheck() error = %v, want %v", serverErr.Err, ErrServerNotFound)
		}
	})

	t.Run("stopped server health", func(t *testing.T) {
		stopCfg := testConfig("hc-stopped", "Health Stopped", "official")
		_, _ = cm.Start(ctx, stopCfg)
		_ = cm.Stop(ctx, "hc-stopped")

		hc, err := cm.HealthCheck(ctx, "hc-stopped")
		if err != nil {
			t.Fatalf("HealthCheck() error = %v", err)
		}
		if hc.Healthy {
			t.Error("HealthCheck() Healthy = true, want false for stopped server")
		}
	})
}

func TestConnectionManager_HealthCheckAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	configs := []MCPServerConfig{
		testConfig("hca1", "HCA 1", "official"),
		testConfig("hca2", "HCA 2", "official"),
	}

	for _, cfg := range configs {
		if _, err := cm.Start(ctx, cfg); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	}

	results := cm.HealthCheckAll(ctx)
	if len(results) != 2 {
		t.Fatalf("HealthCheckAll() count = %v, want 2", len(results))
	}

	for id, hc := range results {
		if !hc.Healthy {
			t.Errorf("HealthCheckAll()[%s] Healthy = false, want true", id)
		}
	}
}

func TestRateLimiter(t *testing.T) {
	t.Run("allow requests within limit", func(t *testing.T) {
		rl := NewRateLimiter(10)
		for i := 0; i < 10; i++ {
			if !rl.Allow() {
				t.Fatalf("Allow() = false at request %d, expected true", i+1)
			}
		}
	})

	t.Run("rate limit enforcement", func(t *testing.T) {
		rl := NewRateLimiter(5)
		// Use all tokens
		for i := 0; i < 5; i++ {
			rl.Allow()
		}
		// Next request should be denied
		if rl.Allow() {
			t.Error("Allow() = true after exhausting tokens, want false")
		}
	})

	t.Run("token refill after time passes", func(t *testing.T) {
		rl := NewRateLimiter(5)
		// Use all tokens
		for i := 0; i < 5; i++ {
			rl.Allow()
		}

		// Simulate time passing
		rl.mu.Lock()
		rl.lastRefill = time.Now().Add(-2 * time.Second)
		rl.mu.Unlock()

		// Should allow again
		if !rl.Allow() {
			t.Error("Allow() = false after refill time, want true")
		}
	})
}

func TestConnectionManager_AcquireToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	cfg := testConfig("rate1", "Rate Test", "official")
	cfg.RateLimit = 3
	_, err := cm.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	t.Run("acquire within limit", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			if err := cm.AcquireToken("rate1"); err != nil {
				t.Fatalf("AcquireToken() error = %v at request %d", err, i+1)
			}
		}
	})

	t.Run("exceed rate limit", func(t *testing.T) {
		err := cm.AcquireToken("rate1")
		if err == nil {
			t.Fatal("AcquireToken() expected error, got nil")
		}
		if err != ErrRateLimitExceeded {
			t.Errorf("AcquireToken() error = %v, want %v", err, ErrRateLimitExceeded)
		}
	})

	t.Run("no rate limiter configured", func(t *testing.T) {
		// Server without rate limit should always allow
		err := cm.AcquireToken("nonexistent-limiter")
		if err != nil {
			t.Errorf("AcquireToken() for unconfigured server = %v, want nil", err)
		}
	})
}

func TestConnectionManager_ConcurrentConnections(t *testing.T) {
	ctx := context.Background()
	cm := NewConnectionManager()

	var wg sync.WaitGroup

	// Concurrent starts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := testConfig(
				"conc-start"+string(rune('0'+id)),
				"Concurrent Start "+string(rune('0'+id)),
				"test",
			)
			_, _ = cm.Start(ctx, cfg)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cm.ListInstances(ctx)
			_, _ = cm.GetInstance(ctx, "any")
		}()
	}

	// Concurrent health checks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cm.HealthCheckAll(ctx)
		}()
	}

	wg.Wait()
	// Ensure no race condition panic
}

func TestConnectionManager_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cm := NewConnectionManager()

	t.Run("Start with canceled context", func(t *testing.T) {
		cfg := testConfig("cancel-start", "Cancel Start", "test")
		_, err := cm.Start(ctx, cfg)
		if err == nil {
			t.Fatal("Start() expected error with canceled context")
		}
	})

	t.Run("Stop with canceled context", func(t *testing.T) {
		err := cm.Stop(ctx, "any")
		if err == nil {
			t.Fatal("Stop() expected error with canceled context")
		}
	})

	t.Run("GetInstance with canceled context", func(t *testing.T) {
		_, err := cm.GetInstance(ctx, "any")
		if err == nil {
			t.Fatal("GetInstance() expected error with canceled context")
		}
	})

	t.Run("HealthCheck with canceled context", func(t *testing.T) {
		_, err := cm.HealthCheck(ctx, "any")
		if err == nil {
			t.Fatal("HealthCheck() expected error with canceled context")
		}
	})
}

func TestConnectionManager_SetRateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	cm := NewConnectionManager()

	cfg := testConfig("set-rate", "Set Rate Test", "official")
	_, err := cm.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Set rate limit after start
	cm.SetRateLimit("set-rate", 5)

	// Should allow 5 requests
	for i := 0; i < 5; i++ {
		if err := cm.AcquireToken("set-rate"); err != nil {
			t.Fatalf("AcquireToken() error = %v at request %d", err, i+1)
		}
	}

	// 6th should fail
	if err := cm.AcquireToken("set-rate"); err != ErrRateLimitExceeded {
		t.Errorf("AcquireToken() error = %v, want %v", err, ErrRateLimitExceeded)
	}
}
