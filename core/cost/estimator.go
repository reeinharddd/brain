package cost

import (
	"context"
	"fmt"
	"sync"
)

// ModelPricing holds pricing information for a model
type ModelPricing struct {
	ModelID         string
	Provider        string
	CostPer1KInput  float64 // USD per 1K input tokens
	CostPer1KOutput float64 // USD per 1K output tokens
	CostPer1KCache  float64 // USD per 1K cache read tokens (if applicable)
	Currency        string
}

// CostEstimator estimates costs for model API calls
type CostEstimator struct {
	mu      sync.RWMutex
	pricing map[string]*ModelPricing // modelID -> pricing
}

// Common model pricing
var DefaultPricing = map[string]ModelPricing{
	"gpt-4":         {ModelID: "gpt-4", CostPer1KInput: 0.03, CostPer1KOutput: 0.06, Currency: "USD"},
	"gpt-3.5-turbo": {ModelID: "gpt-3.5-turbo", CostPer1KInput: 0.0015, CostPer1KOutput: 0.002, Currency: "USD"},
	"claude-opus":   {ModelID: "claude-opus", CostPer1KInput: 0.015, CostPer1KOutput: 0.075, Currency: "USD"},
	"claude-sonnet": {ModelID: "claude-sonnet", CostPer1KInput: 0.003, CostPer1KOutput: 0.015, Currency: "USD"},
	"claude-haiku":  {ModelID: "claude-haiku", CostPer1KInput: 0.00025, CostPer1KOutput: 0.00125, Currency: "USD"},
	"llama-3-70b":   {ModelID: "llama-3-70b", CostPer1KInput: 0.0005, CostPer1KOutput: 0.001, Currency: "USD"},
	"local-model":   {ModelID: "local-model", CostPer1KInput: 0, CostPer1KOutput: 0, Currency: "USD"},
}

// NewCostEstimator creates a new CostEstimator with default pricing loaded.
func NewCostEstimator() *CostEstimator {
	pricing := make(map[string]*ModelPricing, len(DefaultPricing))
	for id, p := range DefaultPricing {
		p := p
		pricing[id] = &p
	}
	return &CostEstimator{
		pricing: pricing,
	}
}

// SetPricing sets or updates pricing for a model.
func (ce *CostEstimator) SetPricing(ctx context.Context, pricing ModelPricing) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cost estimator: set pricing: %w", err)
	}
	if pricing.ModelID == "" {
		return fmt.Errorf("cost estimator: model ID is required")
	}
	ce.mu.Lock()
	ce.pricing[pricing.ModelID] = &pricing
	ce.mu.Unlock()
	return nil
}

// GetPricing retrieves pricing for a model.
func (ce *CostEstimator) GetPricing(ctx context.Context, modelID string) (*ModelPricing, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("cost estimator: get pricing: %w", err)
	}
	ce.mu.RLock()
	p, ok := ce.pricing[modelID]
	ce.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("cost estimator: no pricing found for model %q", modelID)
	}
	return p, nil
}

// Estimate calculates the estimated cost for a model API call.
func (ce *CostEstimator) Estimate(ctx context.Context, modelID string, inputTokens, outputTokens int) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("cost estimator: estimate: %w", err)
	}
	p, err := ce.GetPricing(ctx, modelID)
	if err != nil {
		return 0, err
	}
	inputCost := float64(inputTokens) / 1000.0 * p.CostPer1KInput
	outputCost := float64(outputTokens) / 1000.0 * p.CostPer1KOutput
	return inputCost + outputCost, nil
}

// ListPricing returns all pricing configurations.
func (ce *CostEstimator) ListPricing(ctx context.Context) []ModelPricing {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	result := make([]ModelPricing, 0, len(ce.pricing))
	for _, p := range ce.pricing {
		result = append(result, *p)
	}
	return result
}
