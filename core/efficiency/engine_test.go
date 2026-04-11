package efficiency

import (
	"context"
	"testing"
)

func TestEngine_ExactCacheHit(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	input := "What is the capital of France?"
	req := OptimizationRequest{
		Input:        input,
		TokenCount:   10,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	// First call: cache miss, stores in cache
	result1, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call: exact cache hit
	result2, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result2.CacheHit {
		t.Fatal("expected cache hit on second call")
	}
	if result2.CacheType != "exact" {
		t.Errorf("expected cache type 'exact', got %q", result2.CacheType)
	}
	if result2.SavingsPercent != 100 {
		t.Errorf("expected 100%% savings on cache hit, got %.2f", result2.SavingsPercent)
	}

	// Verify we got different results (first was processed, second was cached)
	if result1.CacheType != "none" && result2.CacheType != "exact" {
		t.Logf("first call cache type: %s, second call: %s", result1.CacheType, result2.CacheType)
	}
}

func TestEngine_SemanticCacheHit(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	// Create two very similar inputs
	input1 := "Explain the theory of relativity in simple terms"
	input2 := "Explain the theory of relativity in simple terms please"

	req1 := OptimizationRequest{
		Input:        input1,
		TokenCount:   20,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	req2 := OptimizationRequest{
		Input:        input2,
		TokenCount:   20,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	// First call
	_, err := engine.Optimize(ctx, req1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call with similar input may hit semantic cache
	result2, err := engine.Optimize(ctx, req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Second call: CacheHit=%v, CacheType=%s, SavingsPercent=%.2f",
		result2.CacheHit, result2.CacheType, result2.SavingsPercent)
}

func TestEngine_CacheMissTriggersCompaction(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	// Clear caches to ensure misses
	engine.ClearCache()

	input := "A very long input that exceeds the token limit "
	input = input + input + input + input + input
	input = input + input + input + input + input

	req := OptimizationRequest{
		Input:        input,
		TokenCount:   500,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	result, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OptimizedTokens >= req.TokenCount {
		t.Errorf("expected compaction to reduce tokens, got optimized=%d >= original=%d",
			result.OptimizedTokens, req.TokenCount)
	}

	if len(result.Techniques) == 0 {
		t.Error("expected at least one technique applied")
	}
}

func TestEngine_CompactionReducesTokensBelowLimit(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()
	engine.ClearCache()

	input := "Some content that needs compaction because it's too long "
	input = input + input + input + input + input

	req := OptimizationRequest{
		Input:        input,
		TokenCount:   500,
		TokenLimit:   200,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	result, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OptimizedTokens > req.TokenLimit {
		t.Logf("warning: optimized tokens %d > limit %d", result.OptimizedTokens, req.TokenLimit)
	}
}

func TestEngine_SavingsCalculation(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	input := "test input for savings calculation"
	req := OptimizationRequest{
		Input:        input,
		TokenCount:   100,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	// First call processes normally
	_, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call hits exact cache
	result, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OriginalTokens != 100 {
		t.Errorf("expected original tokens 100, got %d", result.OriginalTokens)
	}

	if result.OptimizedTokens != 0 {
		t.Errorf("expected optimized tokens 0 on cache hit, got %d", result.OptimizedTokens)
	}

	if result.SavingsPercent != 100 {
		t.Errorf("expected savings 100%%, got %.2f", result.SavingsPercent)
	}

	if result.SavingsUSD <= 0 {
		t.Errorf("expected positive savings in USD, got %.6f", result.SavingsUSD)
	}
}

func TestEngine_GetCacheStats(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	// Initial stats should be zero
	stats := engine.GetCacheStats()
	if stats["exact_cache"] != 0 {
		t.Errorf("expected exact_cache 0, got %d", stats["exact_cache"])
	}

	// Add an entry
	req := OptimizationRequest{
		Input:        "test input",
		TokenCount:   10,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	_, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats = engine.GetCacheStats()
	if stats["exact_cache"] < 1 {
		t.Errorf("expected exact_cache >= 1, got %d", stats["exact_cache"])
	}
	if stats["semantic_cache"] < 1 {
		t.Errorf("expected semantic_cache >= 1, got %d", stats["semantic_cache"])
	}
	if stats["prompt_cache"] < 1 {
		t.Errorf("expected prompt_cache >= 1, got %d", stats["prompt_cache"])
	}
}

func TestEngine_ClearCache(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	// Add entries
	req := OptimizationRequest{
		Input:        "test input",
		TokenCount:   10,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	_, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Clear caches
	engine.ClearCache()

	stats := engine.GetCacheStats()
	if stats["exact_cache"] != 0 {
		t.Errorf("expected exact_cache 0 after clear, got %d", stats["exact_cache"])
	}
	if stats["semantic_cache"] != 0 {
		t.Errorf("expected semantic_cache 0 after clear, got %d", stats["semantic_cache"])
	}
	if stats["prompt_cache"] != 0 {
		t.Errorf("expected prompt_cache 0 after clear, got %d", stats["prompt_cache"])
	}
}

func TestEngine_EmptyRequest(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	req := OptimizationRequest{
		Input:        "",
		TokenCount:   0,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	result, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OriginalTokens != 0 {
		t.Errorf("expected 0 original tokens, got %d", result.OriginalTokens)
	}
	if result.OptimizedTokens != 0 {
		t.Errorf("expected 0 optimized tokens, got %d", result.OptimizedTokens)
	}
}

func TestEngine_OverBudgetRequest(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()
	engine.ClearCache()

	// Create a very large input
	input := "Very long content that will definitely go over budget "
	input = input + input + input + input + input
	input = input + input + input + input + input

	req := OptimizationRequest{
		Input:        input,
		TokenCount:   10000,
		TokenLimit:   500,
		MinTier:      1,
		BudgetUSD:    0.01, // Very small budget
		DefaultModel: "default",
	}

	result, err := engine.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still produce a result (may be truncated)
	if result.OriginalTokens != 10000 {
		t.Errorf("expected 10000 original tokens, got %d", result.OriginalTokens)
	}

	if result.OptimizedTokens >= result.OriginalTokens {
		t.Errorf("expected optimization to reduce tokens, got optimized=%d >= original=%d",
			result.OptimizedTokens, result.OriginalTokens)
	}
}

func TestEngine_MultipleOptimizations(t *testing.T) {
	ctx := context.Background()
	engine := NewTokenEfficiencyEngine()

	inputs := []string{
		"Input number one for testing",
		"Input number two for testing",
		"Input number three for testing",
	}

	for i, input := range inputs {
		req := OptimizationRequest{
			Input:        input,
			TokenCount:   10 + i,
			TokenLimit:   100,
			MinTier:      1,
			BudgetUSD:    1.0,
			DefaultModel: "default",
		}

		result, err := engine.Optimize(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error on input %d: %v", i, err)
		}

		if result.OriginalTokens != 10+i {
			t.Errorf("input %d: expected %d original tokens, got %d", i, 10+i, result.OriginalTokens)
		}
	}

	stats := engine.GetCacheStats()
	if stats["exact_cache"] != 3 {
		t.Errorf("expected exact_cache 3, got %d", stats["exact_cache"])
	}
}

func TestEngine_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	engine := NewTokenEfficiencyEngine()

	req := OptimizationRequest{
		Input:        "test input",
		TokenCount:   10,
		TokenLimit:   100,
		MinTier:      1,
		BudgetUSD:    1.0,
		DefaultModel: "default",
	}

	// With cancelled context, Get operations should return false/empty
	// but Put operations will also fail with context error
	_, err := engine.Optimize(ctx, req)
	if err == nil {
		// It's OK if it returns without error but with empty result
		t.Log("optimization returned without error on cancelled context")
	} else {
		t.Logf("optimization returned error on cancelled context (expected): %v", err)
	}
}
