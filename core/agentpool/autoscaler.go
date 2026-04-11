package agentpool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutoScaler manages pool scaling based on load
type AutoScaler struct {
	mu       sync.RWMutex
	manager  *PoolManager
	enabled  bool
	interval time.Duration
}

// NewAutoScaler creates a new AutoScaler for the given pool manager.
func NewAutoScaler(manager *PoolManager, interval time.Duration) *AutoScaler {
	return &AutoScaler{
		manager:  manager,
		interval: interval,
		enabled:  false,
	}
}

// Enable enables the autoscaler.
func (as *AutoScaler) Enable() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.enabled = true
}

// Disable disables the autoscaler.
func (as *AutoScaler) Disable() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.enabled = false
}

// IsEnabled returns whether the autoscaler is enabled.
func (as *AutoScaler) IsEnabled() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.enabled
}

// Tick evaluates all pools and performs scaling decisions.
func (as *AutoScaler) Tick(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("tick cancelled: %w", ctx.Err())
	default:
	}

	as.mu.RLock()
	enabled := as.enabled
	as.mu.RUnlock()

	if !enabled {
		return nil
	}

	roles := as.manager.ListPools(ctx)
	for _, role := range roles {
		if err := as.evaluatePool(ctx, role); err != nil {
			// Log error but continue with other pools
			// In production you'd want proper logging here
		}
	}
	return nil
}

// evaluatePool evaluates a single pool and performs scaling if needed.
func (as *AutoScaler) evaluatePool(ctx context.Context, role AgentRole) error {
	status, err := as.manager.GetPoolStatus(ctx, role)
	if err != nil {
		return fmt.Errorf("failed to get pool status for %s: %w", role, err)
	}

	// Scale up if load exceeds threshold and we haven't hit max
	if status.Load > as.getScaleUpThreshold(role) && status.TotalInstances < as.getMaxInstances(role) {
		// Check cooldown
		if as.isScaleUpCooldownExpired(role) {
			if err := as.manager.ScalePool(ctx, role, 1); err != nil {
				return fmt.Errorf("failed to scale up pool %s: %w", role, err)
			}
		}
	}

	// Scale down if all instances have been idle for too long and we're above min
	if status.AvailableInstances == status.TotalInstances && status.TotalInstances > as.getMinInstances(role) {
		if as.isScaleDownCooldownExpired(role) {
			if err := as.manager.ScalePool(ctx, role, -1); err != nil {
				return fmt.Errorf("failed to scale down pool %s: %w", role, err)
			}
		}
	}

	return nil
}

// getScaleUpThreshold returns the scale up threshold for a role.
func (as *AutoScaler) getScaleUpThreshold(role AgentRole) float64 {
	as.manager.mu.RLock()
	config, ok := as.manager.configs[role]
	as.manager.mu.RUnlock()
	if !ok {
		return 0.7 // default
	}
	return config.ScaleUpThreshold
}

// getMaxInstances returns the max instances for a role.
func (as *AutoScaler) getMaxInstances(role AgentRole) int {
	as.manager.mu.RLock()
	config, ok := as.manager.configs[role]
	as.manager.mu.RUnlock()
	if !ok {
		return 0
	}
	return config.MaxInstances
}

// getMinInstances returns the min instances for a role.
func (as *AutoScaler) getMinInstances(role AgentRole) int {
	as.manager.mu.RLock()
	config, ok := as.manager.configs[role]
	as.manager.mu.RUnlock()
	if !ok {
		return 0
	}
	return config.MinInstances
}

// isScaleUpCooldownExpired checks if enough time has passed since last scale up.
func (as *AutoScaler) isScaleUpCooldownExpired(role AgentRole) bool {
	as.manager.mu.RLock()
	pool, ok := as.manager.pools[role]
	as.manager.mu.RUnlock()
	if !ok {
		return false
	}

	pool.mu.RLock()
	lastScaleUp := pool.lastScaleUp
	config := pool.config
	pool.mu.RUnlock()

	if lastScaleUp.IsZero() {
		return true
	}

	cooldown := config.ScaleDownTimeout // reuse the same timeout for simplicity
	if config.ScaleDownTimeout > 0 {
		cooldown = config.ScaleDownTimeout
	}
	return time.Since(lastScaleUp) > cooldown
}

// isScaleDownCooldownExpired checks if enough time has passed since last scale down or scale up.
func (as *AutoScaler) isScaleDownCooldownExpired(role AgentRole) bool {
	as.manager.mu.RLock()
	pool, ok := as.manager.pools[role]
	as.manager.mu.RUnlock()
	if !ok {
		return false
	}

	pool.mu.RLock()
	lastScaleDown := pool.lastScaleDown
	lastScaleUp := pool.lastScaleUp
	config := pool.config
	pool.mu.RUnlock()

	// If we recently scaled up, don't scale down yet
	if !lastScaleUp.IsZero() && time.Since(lastScaleUp) <= config.ScaleDownTimeout {
		return false
	}

	if lastScaleDown.IsZero() {
		return true
	}

	return time.Since(lastScaleDown) > config.ScaleDownTimeout
}
