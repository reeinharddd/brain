package efficiency

import (
	"context"
	"fmt"
	"time"
)

// OptimizationRequest describes what needs to be optimized
type OptimizationRequest struct {
	Input        string
	TokenCount   int
	TokenLimit   int
	MinTier      int     // minimum capability tier
	BudgetUSD    float64
	DefaultModel string
}

// AppliedTechnique describes one optimization applied
type AppliedTechnique struct {
	Name        string
	Description string
	TokensSaved int
	CostSaved   float64
	RiskLevel   string
}

// OptimizationResult is the result of optimization
type OptimizationResult struct {
	OriginalTokens  int
	OptimizedTokens int
	SavingsPercent  float64
	SavingsUSD      float64
	Techniques      []AppliedTechnique
	CacheHit        bool
	CacheType       string // exact, semantic, none
}

// TokenEfficiencyEngine orchestrates all token optimization
type TokenEfficiencyEngine struct {
	exactCache    *ExactCache
	semanticCache *SemanticCache
	promptCache   *PromptCache
	compactor     *Compactor
	costPerToken  float64 // USD per token for default model
}

// NewTokenEfficiencyEngine creates a new token efficiency engine with default settings
func NewTokenEfficiencyEngine() *TokenEfficiencyEngine {
	return &TokenEfficiencyEngine{
		exactCache:    NewExactCache(1000, 1*time.Hour),
		semanticCache: NewSemanticCache(500, 0.85, 30*time.Minute),
		promptCache:   NewPromptCache(200, 2*time.Hour),
		compactor:     NewCompactor(4096, "moderate"),
		costPerToken:  0.00001, // Default: $0.01 per 1K tokens
	}
}

// Optimize runs the optimization pipeline for the given request
func (e *TokenEfficiencyEngine) Optimize(ctx context.Context, req OptimizationRequest) (*OptimizationResult, error) {
	if req.Input == "" {
		return &OptimizationResult{
			OriginalTokens:  0,
			OptimizedTokens: 0,
			SavingsPercent:  0,
			SavingsUSD:      0,
			Techniques:      []AppliedTechnique{},
			CacheHit:        false,
			CacheType:       "none",
		}, nil
	}

	result := &OptimizationResult{
		OriginalTokens:  req.TokenCount,
		OptimizedTokens: req.TokenCount,
		Techniques:      []AppliedTechnique{},
		CacheType:       "none",
	}

	// Step 1: Check exact cache
	if entry, found := e.exactCache.Get(ctx, req.Input); found {
		result.OptimizedTokens = 0
		result.SavingsPercent = 100
		result.SavingsUSD = entry.CostUSD
		result.CacheHit = true
		result.CacheType = "exact"
		result.Techniques = append(result.Techniques, AppliedTechnique{
			Name:        "exact_cache_hit",
			Description: "Identical request found in exact cache",
			TokensSaved: req.TokenCount,
			CostSaved:   entry.CostUSD,
			RiskLevel:   "none",
		})
		return result, nil
	}

	// Step 2: Check semantic cache
	if entry, similarity := e.semanticCache.Lookup(ctx, hashString(req.Input)); similarity > 0 {
		result.OptimizedTokens = 0
		result.SavingsPercent = 100
		result.SavingsUSD = entry.CostUSD
		result.CacheHit = true
		result.CacheType = "semantic"
		result.Techniques = append(result.Techniques, AppliedTechnique{
			Name:        "semantic_cache_hit",
			Description: fmt.Sprintf("Similar request found in semantic cache (similarity: %.2f)", similarity),
			TokensSaved: req.TokenCount,
			CostSaved:   entry.CostUSD,
			RiskLevel:   "none",
		})
		return result, nil
	}

	// Step 3: Check prompt cache
	var promptCacheHit bool
	if _, found := e.promptCache.Get(ctx, extractPrefix(req.Input, 50)); found {
		promptCacheHit = true
	}

	// Step 4: Apply compaction if over token limit
	currentTokens := req.TokenCount
	currentInput := req.Input

	if currentTokens > req.TokenLimit {
		// Update compactor's maxTokens to match the request limit
		e.compactor.maxTokens = req.TokenLimit

		compactionResult, compacted := e.compactor.Compact(ctx, currentInput, currentTokens)

		if compactionResult.Method != "none" {
			tokensSaved := compactionResult.OriginalTokens - compactionResult.CompactedTokens
			costSaved := float64(tokensSaved) * e.costPerToken

			result.Techniques = append(result.Techniques, AppliedTechnique{
				Name:        compactionResult.Method,
				Description: fmt.Sprintf("Applied %s compaction strategy", compactionResult.Method),
				TokensSaved: tokensSaved,
				CostSaved:   costSaved,
				RiskLevel:   compactionResult.RiskLevel,
			})

			currentTokens = compactionResult.CompactedTokens
			currentInput = compacted
		}
	}

	// Calculate savings
	tokensSaved := req.TokenCount - currentTokens
	result.OptimizedTokens = currentTokens

	if req.TokenCount > 0 {
		result.SavingsPercent = float64(tokensSaved) / float64(req.TokenCount) * 100
	}
	result.SavingsUSD = float64(tokensSaved) * e.costPerToken

	if promptCacheHit {
		result.CacheHit = true
		result.CacheType = "prompt"
		result.Techniques = append(result.Techniques, AppliedTechnique{
			Name:        "prompt_cache_hit",
			Description: "Matching prefix found in prompt cache",
			TokensSaved: 0,
			CostSaved:   0,
			RiskLevel:   "none",
		})
	}

	// Store in caches for future requests
	entry := &CacheEntry{
		Content:    currentInput,
		TokenCount: currentTokens,
		CostUSD:    float64(currentTokens) * e.costPerToken,
		CreatedAt:  time.Now(),
	}

	// Store in exact cache
	if err := e.exactCache.Put(ctx, req.Input, entry); err != nil {
		return result, fmt.Errorf("failed to store in exact cache: %w", err)
	}

	// Store in semantic cache
	semanticHash := hashString(req.Input)
	if err := e.semanticCache.Put(ctx, semanticHash, req.Input, entry); err != nil {
		return result, fmt.Errorf("failed to store in semantic cache: %w", err)
	}

	// Store prefix in prompt cache
	prefix := extractPrefix(req.Input, 50)
	if err := e.promptCache.Put(ctx, prefix, entry); err != nil {
		return result, fmt.Errorf("failed to store in prompt cache: %w", err)
	}

	return result, nil
}

// GetCacheStats returns the current counts for all caches
func (e *TokenEfficiencyEngine) GetCacheStats() map[string]int {
	return map[string]int{
		"exact_cache":    e.exactCache.Count(),
		"semantic_cache": e.semanticCache.Count(),
		"prompt_cache":   e.promptCache.Count(),
	}
}

// ClearCache resets all caches
func (e *TokenEfficiencyEngine) ClearCache() {
	e.exactCache = NewExactCache(1000, 1*time.Hour)
	e.semanticCache = NewSemanticCache(500, 0.85, 30*time.Minute)
	e.promptCache = NewPromptCache(200, 2*time.Hour)
}

// hashString produces a simple hash for the given string
func hashString(s string) string {
	h := uint64(0)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint64(s[i])
	}
	return fmt.Sprintf("%016x", h)
}

// extractPrefix extracts the first n characters of a string as a prefix
func extractPrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
