package runtime

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CapabilityTier represents model capability levels
type CapabilityTier int

const (
	Tier1Constrained CapabilityTier = 1 // Small context, limited tools, no complex reasoning
	Tier2Standard   CapabilityTier = 2 // Standard context, tool use, moderate reasoning
	Tier3Advanced   CapabilityTier = 3 // Full context, advanced reasoning, complex orchestration
)

// ModelCapability describes a model's capabilities
type ModelCapability struct {
	ModelID          string
	Provider         string
	Tier             CapabilityTier
	MaxContextTokens int
	MaxOutputTokens  int
	SupportsTools    bool
	SupportsParallel bool
	CostPer1KInput   float64
	CostPer1KOutput  float64
	LatencyP50       time.Duration
	LatencyP99       time.Duration
	IsLocal          bool
}

// ModelRegistry manages registered models
type ModelRegistry struct {
	mu     sync.RWMutex
	models map[string]*ModelCapability // modelID -> capability
}

// NewModelRegistry creates a new empty ModelRegistry
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*ModelCapability),
	}
}

var (
	ErrModelNotFound      = errors.New("model not found")
	ErrModelAlreadyExists = errors.New("model already exists")
)

// Register adds a model to the registry
func (r *ModelRegistry) Register(ctx context.Context, model ModelCapability) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.models[model.ModelID]; exists {
		return ErrModelAlreadyExists
	}

	cpy := model
	r.models[model.ModelID] = &cpy
	return nil
}

// Get retrieves a model by ID
func (r *ModelRegistry) Get(ctx context.Context, modelID string) (*ModelCapability, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.models[modelID]
	if !ok {
		return nil, ErrModelNotFound
	}

	cpy := *m
	return &cpy, nil
}

// List returns all registered models
func (r *ModelRegistry) List(ctx context.Context) []ModelCapability {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ModelCapability, 0, len(r.models))
	for _, m := range r.models {
		cpy := *m
		result = append(result, cpy)
	}
	return result
}

// GetByTier returns all models matching the given capability tier
func (r *ModelRegistry) GetByTier(ctx context.Context, tier CapabilityTier) []ModelCapability {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ModelCapability
	for _, m := range r.models {
		if m.Tier == tier {
			cpy := *m
			result = append(result, cpy)
		}
	}
	return result
}

// Delete removes a model from the registry
func (r *ModelRegistry) Delete(ctx context.Context, modelID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.models[modelID]; !exists {
		return ErrModelNotFound
	}

	delete(r.models, modelID)
	return nil
}

// Count returns the number of registered models
func (r *ModelRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.models)
}
