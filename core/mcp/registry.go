package mcp

import (
	"context"
	"fmt"
	"sync"
)

// MCPRegistry manages MCP server definitions and instances.
type MCPRegistry struct {
	mu           sync.RWMutex
	servers      map[string]*MCPServerConfig // id -> config
	instances    map[string]*MCPServer       // id -> running instance
	byCategory   map[string][]string         // category -> server IDs
}

// NewMCPRegistry creates a new MCP registry.
func NewMCPRegistry() *MCPRegistry {
	return &MCPRegistry{
		servers:      make(map[string]*MCPServerConfig),
		instances:    make(map[string]*MCPServer),
		byCategory:   make(map[string][]string),
	}
}

// Register adds a server configuration to the registry.
func (r *MCPRegistry) Register(ctx context.Context, config MCPServerConfig) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("mcp: context canceled while registering server %q: %w", config.ID, ctx.Err())
	default:
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[config.ID]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateServer, config.ID)
	}

	cfgCopy := config
	r.servers[config.ID] = &cfgCopy

	// Add to category index
	if config.Category != "" {
		r.byCategory[config.Category] = append(r.byCategory[config.Category], config.ID)
	}

	return nil
}

// Get retrieves a server configuration by ID.
func (r *MCPRegistry) Get(ctx context.Context, id string) (*MCPServerConfig, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp: context canceled while getting server %q: %w", id, ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, exists := r.servers[id]
	if !exists {
		return nil, &ServerError{
			ServerID: id,
			Op:       "get",
			Err:      ErrServerNotFound,
		}
	}

	cfgCopy := *cfg
	return &cfgCopy, nil
}

// List returns all server configurations.
func (r *MCPRegistry) List(ctx context.Context) []MCPServerConfig {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]MCPServerConfig, 0, len(r.servers))
	for _, cfg := range r.servers {
		cfgCopy := *cfg
		result = append(result, cfgCopy)
	}

	return result
}

// GetByCategory returns all server configurations in a category.
func (r *MCPRegistry) GetByCategory(ctx context.Context, category string) []MCPServerConfig {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, exists := r.byCategory[category]
	if !exists {
		return nil
	}

	result := make([]MCPServerConfig, 0, len(ids))
	for _, id := range ids {
		if cfg, ok := r.servers[id]; ok {
			cfgCopy := *cfg
			result = append(result, cfgCopy)
		}
	}

	return result
}

// Unregister removes a server configuration from the registry.
func (r *MCPRegistry) Unregister(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("mcp: context canceled while unregistering server %q: %w", id, ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, exists := r.servers[id]
	if !exists {
		return &ServerError{
			ServerID: id,
			Op:       "unregister",
			Err:      ErrServerNotFound,
		}
	}

	// Remove from category index
	if cfg.Category != "" {
		ids := r.byCategory[cfg.Category]
		for i, sid := range ids {
			if sid == id {
				r.byCategory[cfg.Category] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
		// Clean up empty category slice
		if len(r.byCategory[cfg.Category]) == 0 {
			delete(r.byCategory, cfg.Category)
		}
	}

	delete(r.servers, id)
	return nil
}

// Count returns the number of registered servers.
func (r *MCPRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.servers)
}
