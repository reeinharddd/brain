package agentpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrDuplicatePool = errors.New("pool already registered")
	ErrUnknownPool   = errors.New("unknown pool")
)

// PoolManager manages multiple agent pools
type PoolManager struct {
	mu      sync.RWMutex
	pools   map[AgentRole]*AgentPool
	defs    map[AgentRole]AgentDefinition
	configs map[AgentRole]PoolConfig
}

// NewPoolManager creates a new pool manager.
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools:   make(map[AgentRole]*AgentPool),
		defs:    make(map[AgentRole]AgentDefinition),
		configs: make(map[AgentRole]PoolConfig),
	}
}

// RegisterPool registers a new agent pool.
func (m *PoolManager) RegisterPool(ctx context.Context, role AgentRole, def AgentDefinition, config PoolConfig) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("register pool cancelled: %w", ctx.Err())
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[role]; exists {
		return fmt.Errorf("cannot register pool: %w for role %s", ErrDuplicatePool, role)
	}

	pool := NewAgentPool(def, config)
	m.pools[role] = pool
	m.defs[role] = def
	m.configs[role] = config
	return nil
}

// SubmitTask submits a task to a specific pool.
func (m *PoolManager) SubmitTask(ctx context.Context, role AgentRole, task *AgentTask) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("submit task cancelled: %w", ctx.Err())
	default:
	}

	m.mu.RLock()
	pool, ok := m.pools[role]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cannot submit task: %w for role %s", ErrUnknownPool, role)
	}

	return pool.SubmitTask(ctx, task)
}

// GetPoolStatus returns the status of a specific pool.
func (m *PoolManager) GetPoolStatus(ctx context.Context, role AgentRole) (PoolStatus, error) {
	select {
	case <-ctx.Done():
		return PoolStatus{}, fmt.Errorf("get pool status cancelled: %w", ctx.Err())
	default:
	}

	m.mu.RLock()
	pool, ok := m.pools[role]
	m.mu.RUnlock()

	if !ok {
		return PoolStatus{}, fmt.Errorf("cannot get pool status: %w for role %s", ErrUnknownPool, role)
	}

	return pool.GetStatus(ctx), nil
}

// GetAllStatuses returns the status of all registered pools.
func (m *PoolManager) GetAllStatuses(ctx context.Context) map[AgentRole]PoolStatus {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[AgentRole]PoolStatus, len(m.pools))
	for role, pool := range m.pools {
		result[role] = pool.GetStatus(ctx)
	}
	return result
}

// GetAvailableAgent returns an available agent from the specified pool.
func (m *PoolManager) GetAvailableAgent(ctx context.Context, role AgentRole) (*Agent, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("get available agent cancelled: %w", ctx.Err())
	default:
	}

	m.mu.RLock()
	pool, ok := m.pools[role]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("cannot get available agent: %w for role %s", ErrUnknownPool, role)
	}

	return pool.GetAvailableInstance(ctx)
}

// ScalePool scales a specific pool up or down.
func (m *PoolManager) ScalePool(ctx context.Context, role AgentRole, count int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("scale pool cancelled: %w", ctx.Err())
	default:
	}

	m.mu.RLock()
	pool, ok := m.pools[role]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cannot scale pool: %w for role %s", ErrUnknownPool, role)
	}

	if count > 0 {
		_, err := pool.ScaleUp(ctx, count)
		return err
	} else if count < 0 {
		return pool.ScaleDown(ctx, -count)
	}
	return nil
}

// ListPools returns all registered pool roles.
func (m *PoolManager) ListPools(ctx context.Context) []AgentRole {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	roles := make([]AgentRole, 0, len(m.pools))
	for role := range m.pools {
		roles = append(roles, role)
	}
	return roles
}

// PoolCount returns the number of registered pools.
func (m *PoolManager) PoolCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pools)
}
