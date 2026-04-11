package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

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
	mu          sync.RWMutex
	instances   map[string]*MCPServer               // id -> running instance
	healthCheck func(ctx context.Context, id string) error
	rateLimiter map[string]*RateLimiter             // server ID -> rate limiter
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		instances:   make(map[string]*MCPServer),
		rateLimiter: make(map[string]*RateLimiter),
	}
}

// SetHealthCheck sets a custom health check function.
func (cm *ConnectionManager) SetHealthCheck(fn func(ctx context.Context, id string) error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.healthCheck = fn
}

// Start simulates starting an MCP server instance.
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

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if inst, exists := cm.instances[config.ID]; exists {
		if inst.Status == StatusRunning {
			return nil, &ServerError{
				ServerID: config.ID,
				Op:       "start",
				Err:      ErrServerAlreadyRunning,
			}
		}
	}

	// Create server instance
	instance := &MCPServer{
		Config:      config,
		Status:      StatusStarting,
		Tools:       []MCPTool{},
		Resources:   []MCPResource{},
		ClientCount: 0,
		StartedAt:   time.Now(),
	}

	cm.instances[config.ID] = instance

	// Set up rate limiter if configured
	if config.RateLimit > 0 {
		cm.rateLimiter[config.ID] = NewRateLimiter(config.RateLimit)
	}

	// Simulate server startup
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
	defer cm.mu.Unlock()

	instance, exists := cm.instances[id]
	if !exists {
		return &ServerError{
			ServerID: id,
			Op:       "stop",
			Err:      ErrServerNotFound,
		}
	}

	if instance.Status != StatusRunning {
		return &ServerError{
			ServerID: id,
			Op:       "stop",
			Err:      ErrServerAlreadyStopped,
		}
	}

	instance.Status = StatusStopped
	delete(cm.rateLimiter, id)

	return nil
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

	// Return a copy
	instanceCopy := *instance
	return &instanceCopy, nil
}

// ListInstances returns all running MCP server instances.
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
		instCopy := *inst
		result = append(result, &instCopy)
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

	if instance.Status == StatusRunning && healthErr == nil {
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
	if instance.Status != StatusRunning {
		errMsg = fmt.Sprintf("server status: %s", instance.Status)
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
