package cost

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// OptimizationRecommendation suggests cost improvements
type OptimizationRecommendation struct {
	ID             string
	Type           string  // "model_switch", "cache_opportunity", "context_reduction", "batch_requests"
	Title          string
	Description    string
	CurrentCost    float64
	EstimatedSavings float64
	Confidence     float64 // 0.0-1.0
	Actionable     bool
}

// UsageRecord represents a cost usage event
type UsageRecord struct {
	ModelID      string
	InputTokens  int
	OutputTokens int
	Cost         float64
	Surface      string
	Timestamp    time.Time
	Cached       bool
}

// CostOptimizer generates optimization recommendations
type CostOptimizer struct {
	estimator *CostEstimator
}

// NewCostOptimizer creates a new CostOptimizer.
func NewCostOptimizer(estimator *CostEstimator) *CostOptimizer {
	return &CostOptimizer{
		estimator: estimator,
	}
}

// AnalyzeUsage generates optimization recommendations based on usage records.
func (co *CostOptimizer) AnalyzeUsage(ctx context.Context, usage []UsageRecord) []OptimizationRecommendation {
	if err := ctx.Err(); err != nil {
		return nil
	}
	if len(usage) == 0 {
		return nil
	}

	var recommendations []OptimizationRecommendation

	// Check for expensive model usage that could be switched
	recommendations = append(recommendations, co.analyzeModelSwitch(usage)...)

	// Check for cache opportunities (repeated queries)
	recommendations = append(recommendations, co.analyzeCacheOpportunities(usage)...)

	// Check for context reduction opportunities
	recommendations = append(recommendations, co.analyzeContextReduction(usage)...)

	// Check for batch opportunities
	recommendations = append(recommendations, co.analyzeBatchOpportunities(usage)...)

	return recommendations
}

// SuggestModelSwitch suggests a cheaper alternative model within the given budget.
func (co *CostOptimizer) SuggestModelSwitch(ctx context.Context, currentModelID string, budget float64) *OptimizationRecommendation {
	if err := ctx.Err(); err != nil {
		return nil
	}

	currentPricing, err := co.estimator.GetPricing(ctx, currentModelID)
	if err != nil {
		return nil
	}

	// Find the cheapest model that could work
	allPricing := co.estimator.ListPricing(ctx)
	var bestAlternative *ModelPricing
	var minCostPerToken float64 = math.MaxFloat64

	for i := range allPricing {
		p := &allPricing[i]
		if p.ModelID == currentModelID {
			continue
		}
		costPerToken := (p.CostPer1KInput + p.CostPer1KOutput) / 2000.0
		if costPerToken < minCostPerToken {
			minCostPerToken = costPerToken
			bestAlternative = p
		}
	}

	if bestAlternative == nil {
		return nil
	}

	currentCostPerToken := (currentPricing.CostPer1KInput + currentPricing.CostPer1KOutput) / 2000.0
	if minCostPerToken >= currentCostPerToken {
		return nil // No cheaper alternative
	}

	// Estimate savings per 1000 tokens
	savingsPer1K := currentPricing.CostPer1KInput + currentPricing.CostPer1KOutput - bestAlternative.CostPer1KInput - bestAlternative.CostPer1KOutput

	return &OptimizationRecommendation{
		ID:             uuid.New().String(),
		Type:           "model_switch",
		Title:          fmt.Sprintf("Switch from %s to %s", currentModelID, bestAlternative.ModelID),
		Description:    fmt.Sprintf("Switching from %s (%.4f/1K tokens) to %s (%.4f/1K tokens) could reduce costs significantly", currentModelID, currentCostPerToken*1000, bestAlternative.ModelID, minCostPerToken*1000),
		CurrentCost:    currentPricing.CostPer1KInput + currentPricing.CostPer1KOutput,
		EstimatedSavings: savingsPer1K,
		Confidence:     0.7,
		Actionable:     true,
	}
}

// EstimateCacheSavings estimates potential savings from caching repeated queries.
func (co *CostOptimizer) EstimateCacheSavings(ctx context.Context, repeatedQueries int, totalTokens int) *OptimizationRecommendation {
	if err := ctx.Err(); err != nil {
		return nil
	}
	if repeatedQueries <= 0 || totalTokens <= 0 {
		return nil
	}

	// Assume average cost per token using a mid-range model
	avgCostPer1K := (DefaultPricing["claude-sonnet"].CostPer1KInput + DefaultPricing["claude-sonnet"].CostPer1KOutput) / 2.0
	avgTokensPerQuery := totalTokens / repeatedQueries

	// Caching typically saves ~80% of input token cost for repeated queries
	cacheSavingsPercent := 0.8
	cacheCostPer1K := DefaultPricing["claude-sonnet"].CostPer1KCache
	if cacheCostPer1K == 0 {
		// If no cache pricing, estimate 10% of input cost
		cacheCostPer1K = DefaultPricing["claude-sonnet"].CostPer1KInput * 0.1
	}

	currentCost := float64(repeatedQueries) * float64(avgTokensPerQuery) / 1000.0 * avgCostPer1K
	cachedCost := float64(repeatedQueries) * float64(avgTokensPerQuery) / 1000.0 * cacheCostPer1K
	savings := currentCost - cachedCost

	return &OptimizationRecommendation{
		ID:             uuid.New().String(),
		Type:           "cache_opportunity",
		Title:          "Enable prompt caching for repeated queries",
		Description:    fmt.Sprintf("Found %d repeated queries with ~%d tokens each. Enabling caching could save %.2f%% of input costs", repeatedQueries, avgTokensPerQuery, cacheSavingsPercent*100),
		CurrentCost:    currentCost,
		EstimatedSavings: savings,
		Confidence:     0.85,
		Actionable:     true,
	}
}

// analyzeModelSwitch checks if expensive models are being used where cheaper ones would work.
func (co *CostOptimizer) analyzeModelSwitch(usage []UsageRecord) []OptimizationRecommendation {
	// Group usage by model and calculate total cost
	modelCost := make(map[string]float64)
	for _, u := range usage {
		modelCost[u.ModelID] += u.Cost
	}

	var recommendations []OptimizationRecommendation

	// Define expensive models and their cheaper alternatives
	expensiveThreshold := 0.01 // per 1K tokens
	for modelID, totalCost := range modelCost {
		pricing, err := co.estimator.GetPricing(context.Background(), modelID)
		if err != nil {
			continue
		}

		costPer1K := pricing.CostPer1KInput + pricing.CostPer1KOutput
		if costPer1K > expensiveThreshold {
			recommendations = append(recommendations, OptimizationRecommendation{
				ID:             uuid.New().String(),
				Type:           "model_switch",
				Title:          fmt.Sprintf("Consider cheaper alternative for %s", modelID),
				Description:    fmt.Sprintf("Model %s is expensive (%.4f/1K tokens). Consider using claude-sonnet or claude-haiku for less demanding tasks.", modelID, costPer1K),
				CurrentCost:    totalCost,
				EstimatedSavings: totalCost * 0.5, // Estimate 50% savings
				Confidence:     0.6,
				Actionable:     true,
			})
		}
	}

	return recommendations
}

// analyzeCacheOpportunities checks for repeated queries that could benefit from caching.
func (co *CostOptimizer) analyzeCacheOpportunities(usage []UsageRecord) []OptimizationRecommendation {
	// Group by model+surface to find repeated patterns
	patternCount := make(map[string]int)
	patternCost := make(map[string]float64)
	patternTokens := make(map[string]int)

	for _, u := range usage {
		key := u.ModelID + ":" + u.Surface
		patternCount[key]++
		patternCost[key] += u.Cost
		patternTokens[key] += u.InputTokens
	}

	var recommendations []OptimizationRecommendation

	for key, count := range patternCount {
		if count < 3 { // Only suggest if there are enough repetitions
			continue
		}
		tokens := patternTokens[key]
		savings := co.EstimateCacheSavings(context.Background(), count, tokens)
		if savings != nil {
			savings.Title = fmt.Sprintf("Cache opportunity detected for %s", key)
			recommendations = append(recommendations, *savings)
		}
	}

	return recommendations
}

// analyzeContextReduction checks for excessive context token usage.
func (co *CostOptimizer) analyzeContextReduction(usage []UsageRecord) []OptimizationRecommendation {
	// Group by model and check for high input/output ratios
	modelInputTokens := make(map[string]int)
	modelOutputTokens := make(map[string]int)

	for _, u := range usage {
		modelInputTokens[u.ModelID] += u.InputTokens
		modelOutputTokens[u.ModelID] += u.OutputTokens
	}

	var recommendations []OptimizationRecommendation

	for modelID, inputTokens := range modelInputTokens {
		outputTokens := modelOutputTokens[modelID]
		if outputTokens == 0 {
			continue
		}

		ratio := float64(inputTokens) / float64(outputTokens)
		if ratio > 10 { // Input is 10x output - possible context bloat
			pricing, err := co.estimator.GetPricing(context.Background(), modelID)
			if err != nil {
				continue
			}

			inputCost := float64(inputTokens) / 1000.0 * pricing.CostPer1KInput
			estimatedSavings := inputCost * 0.3 // Assume 30% could be saved

			recommendations = append(recommendations, OptimizationRecommendation{
				ID:             uuid.New().String(),
				Type:           "context_reduction",
				Title:          fmt.Sprintf("Reduce context for %s", modelID),
				Description:    fmt.Sprintf("High input/output ratio (%.1f) for %s. Consider reducing system prompts or context window.", ratio, modelID),
				CurrentCost:    inputCost,
				EstimatedSavings: estimatedSavings,
				Confidence:     0.65,
				Actionable:     true,
			})
		}
	}

	return recommendations
}

// analyzeBatchOpportunities checks for many small requests that could be batched.
func (co *CostOptimizer) analyzeBatchOpportunities(usage []UsageRecord) []OptimizationRecommendation {
	// Count requests per surface
	surfaceCount := make(map[string]int)
	for _, u := range usage {
		surfaceCount[u.Surface]++
	}

	var recommendations []OptimizationRecommendation

	for surface, count := range surfaceCount {
		if count < 10 { // Need at least 10 requests to consider batching
			continue
		}

		// Calculate total cost for this surface
		totalCost := 0.0
		for _, u := range usage {
			if u.Surface == surface {
				totalCost += u.Cost
			}
		}

		recommendations = append(recommendations, OptimizationRecommendation{
			ID:             uuid.New().String(),
			Type:           "batch_requests",
			Title:          fmt.Sprintf("Batch requests for %s", surface),
			Description:    fmt.Sprintf("Detected %d requests to %s. Batching could reduce API call overhead and costs by ~20%%.", count, surface),
			CurrentCost:    totalCost,
			EstimatedSavings: totalCost * 0.2,
			Confidence:     0.7,
			Actionable:     true,
		})
	}

	// Sort by estimated savings descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].EstimatedSavings > recommendations[j].EstimatedSavings
	})

	return recommendations
}
