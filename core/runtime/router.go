package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// RouteRequest defines requirements for model routing
type RouteRequest struct {
	MinCapabilityTier CapabilityTier
	PreferredModels   []string // Ordered preference list
	FallbackModels    []string // Fallback if preferred unavailable
	MaxTokens         int
	RequiresTools     bool
	RequiresParallel  bool
	BudgetUSD         float64 // Maximum willing to spend
	PreferLocal       bool    // Prefer local models if capable
	LatencySLA        time.Duration
	InputTokens       int // For cost estimation
	OutputTokens      int // For cost estimation
}

// RouteResponse is the router's decision
type RouteResponse struct {
	SelectedModel     *ModelCapability
	Reason            string
	CostEstimate      float64
	FallbackChain     []string // Models to try if selected fails
	CompressionNeeded bool
}

// ScoredModel represents a model with its routing score
type ScoredModel struct {
	Model  *ModelCapability
	Score  float64
	Reason string
}

// Budget defines spending limits
type Budget struct {
	DailyUSD     float64
	MonthlyUSD   float64
	CurrentSpend float64
	ResetAt      time.Time
}

// ModelRouter routes tasks to appropriate models
type ModelRouter struct {
	mu       sync.RWMutex
	registry *ModelRegistry
	budgets  map[string]*Budget // user/workspace -> budget
}

// NewModelRouter creates a new ModelRouter
func NewModelRouter(registry *ModelRegistry) *ModelRouter {
	return &ModelRouter{
		registry: registry,
		budgets:  make(map[string]*Budget),
	}
}

var (
	ErrNoModelsAvailable   = errors.New("no models available matching requirements")
	ErrBudgetExceeded      = errors.New("budget exceeded")
	ErrBudgetNotFound      = errors.New("budget not found")
	ErrModelNotInRegistry = errors.New("model not in registry")
)

// Route selects the best model for the given request
func (r *ModelRouter) Route(ctx context.Context, req RouteRequest) (*RouteResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 1. Get all models from registry
	allModels := r.registry.List(ctx)
	if len(allModels) == 0 {
		return nil, ErrNoModelsAvailable
	}

	// 2. Filter by minimum capability tier
	var candidates []*ModelCapability
	for i := range allModels {
		m := &allModels[i]
		if m.Tier < req.MinCapabilityTier {
			continue
		}

		// 3. Filter by tools requirement
		if req.RequiresTools && !m.SupportsTools {
			continue
		}

		// 4. Filter by parallel requirement
		if req.RequiresParallel && !m.SupportsParallel {
			continue
		}

		// 5. Filter by latency SLA
		if req.LatencySLA > 0 && m.LatencyP50 > req.LatencySLA {
			continue
		}

		candidates = append(candidates, m)
	}

	if len(candidates) == 0 {
		return nil, ErrNoModelsAvailable
	}

	// 6. Score remaining candidates
	scored := r.scoreModels(candidates, req)
	if len(scored) == 0 {
		return nil, ErrNoModelsAvailable
	}

	// 7. Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// 8. Select best model that fits budget
	var selected *ScoredModel
	for _, s := range scored {
		cost := r.estimateCost(s.Model, req.InputTokens, req.OutputTokens)
		if req.BudgetUSD > 0 && cost > req.BudgetUSD {
			continue
		}
		selected = &s
		break
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: all candidates exceed budget", ErrBudgetExceeded)
	}

	// 9. Build fallback chain
	fallbackChain := r.buildFallbackChain(scored, *selected)

	// 10. Estimate cost and check compression needed
	costEstimate := r.estimateCost(selected.Model, req.InputTokens, req.OutputTokens)
	compressionNeeded := req.InputTokens > selected.Model.MaxContextTokens

	reason := selected.Reason
	if compressionNeeded {
		reason += "; context compression needed"
	}

	return &RouteResponse{
		SelectedModel:     selected.Model,
		Reason:            reason,
		CostEstimate:      costEstimate,
		FallbackChain:     fallbackChain,
		CompressionNeeded: compressionNeeded,
	}, nil
}

// scoreModels scores candidate models based on the request requirements
func (r *ModelRouter) scoreModels(models []*ModelCapability, req RouteRequest) []ScoredModel {
	// Calculate average cost for normalization
	var totalCost float64
	var count int
	for _, m := range models {
		avgInputCost := m.CostPer1KInput
		avgOutputCost := m.CostPer1KOutput
		if avgInputCost+avgOutputCost > 0 {
			totalCost += avgInputCost + avgOutputCost
			count++
		}
	}
	avgCost := 0.0
	if count > 0 {
		avgCost = totalCost / float64(count)
	}

	// Build preferred set for quick lookup
	preferredSet := make(map[string]bool)
	for _, p := range req.PreferredModels {
		preferredSet[p] = true
	}

	var scored []ScoredModel
	for _, m := range models {
		var score float64
		var reasons []string

		// Capability tier match: +tier * 10
		tierScore := float64(m.Tier) * 10
		score += tierScore

		// Lower cost: +(1/avg_cost) * 5
		modelAvgCost := m.CostPer1KInput + m.CostPer1KOutput
		if avgCost > 0 && modelAvgCost > 0 {
			costScore := (1.0 / modelAvgCost) * 5
			score += costScore
		}

		// Local preference: +20 if PreferLocal and is local
		if req.PreferLocal && m.IsLocal {
			score += 20
			reasons = append(reasons, "local model preferred")
		}

		// Latency SLA: +10 if P50 < SLA
		if req.LatencySLA > 0 && m.LatencyP50 < req.LatencySLA {
			score += 10
			reasons = append(reasons, "within latency SLA")
		}

		// Policy preference: +15 if in preferred list
		if preferredSet[m.ModelID] {
			score += 15
			reasons = append(reasons, "preferred model")
		}

		reason := "tier " + fmt.Sprintf("%d", m.Tier)
		if len(reasons) > 0 {
			reason += "; " + reasons[0]
			for _, r := range reasons[1:] {
				reason += "; " + r
			}
		}

		scored = append(scored, ScoredModel{
			Model:  m,
			Score:  math.Round(score*100) / 100,
			Reason: reason,
		})
	}

	return scored
}

// buildFallbackChain builds a list of fallback models (next 2 best after selected)
func (r *ModelRouter) buildFallbackChain(scored []ScoredModel, selected ScoredModel) []string {
	var fallback []string
	count := 0
	for _, s := range scored {
		if s.Model.ModelID == selected.Model.ModelID {
			continue
		}
		fallback = append(fallback, s.Model.ModelID)
		count++
		if count >= 2 {
			break
		}
	}
	return fallback
}

// estimateCost calculates the estimated cost for a request
func (r *ModelRouter) estimateCost(model *ModelCapability, inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) / 1000.0 * model.CostPer1KInput
	outputCost := float64(outputTokens) / 1000.0 * model.CostPer1KOutput
	return math.Round((inputCost+outputCost)*10000) / 10000
}

// SetBudget sets a budget for a user/workspace
func (r *ModelRouter) SetBudget(ctx context.Context, id string, budget Budget) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	b := budget
	r.budgets[id] = &b
	return nil
}

// GetBudget retrieves a budget for a user/workspace
func (r *ModelRouter) GetBudget(ctx context.Context, id string) (*Budget, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.budgets[id]
	if !ok {
		return nil, ErrBudgetNotFound
	}

	cpy := *b
	return &cpy, nil
}

// RecordSpend records spending against a user/workspace budget
func (r *ModelRouter) RecordSpend(ctx context.Context, id string, amount float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.budgets[id]
	if !ok {
		return ErrBudgetNotFound
	}

	b.CurrentSpend += amount
	return nil
}

// CheckBudget checks if an estimated cost fits within the budget
func (r *ModelRouter) CheckBudget(ctx context.Context, id string, estimatedCost float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.budgets[id]
	if !ok {
		return ErrBudgetNotFound
	}

	remaining := b.MonthlyUSD - b.CurrentSpend
	if b.DailyUSD > 0 && b.DailyUSD < b.MonthlyUSD {
		// Use daily as a sub-limit
		remaining = b.DailyUSD - b.CurrentSpend
	}

	if estimatedCost > remaining {
		return fmt.Errorf("%w: estimated %.4f exceeds remaining %.4f", ErrBudgetExceeded, estimatedCost, remaining)
	}

	return nil
}
