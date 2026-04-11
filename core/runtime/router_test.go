package runtime

import (
	"context"
	"sync"
	"testing"
	"time"
)

func setupRouter(t *testing.T) (*ModelRouter, context.Context) {
	t.Helper()
	ctx := context.Background()
	reg := NewModelRegistry()
	router := NewModelRouter(reg)

	// Register a variety of models
	models := []ModelCapability{
		{
			ModelID:          "tiny-model",
			Provider:         "test",
			Tier:             Tier1Constrained,
			MaxContextTokens: 2048,
			MaxOutputTokens:  512,
			SupportsTools:    false,
			SupportsParallel: false,
			CostPer1KInput:   0.001,
			CostPer1KOutput:  0.002,
			LatencyP50:       50 * time.Millisecond,
			LatencyP99:       200 * time.Millisecond,
			IsLocal:          true,
		},
		{
			ModelID:          "small-model",
			Provider:         "test",
			Tier:             Tier1Constrained,
			MaxContextTokens: 4096,
			MaxOutputTokens:  1024,
			SupportsTools:    true,
			SupportsParallel: false,
			CostPer1KInput:   0.002,
			CostPer1KOutput:  0.003,
			LatencyP50:       100 * time.Millisecond,
			LatencyP99:       400 * time.Millisecond,
			IsLocal:          false,
		},
		{
			ModelID:          "mid-model",
			Provider:         "test",
			Tier:             Tier2Standard,
			MaxContextTokens: 8192,
			MaxOutputTokens:  2048,
			SupportsTools:    true,
			SupportsParallel: true,
			CostPer1KInput:   0.01,
			CostPer1KOutput:  0.02,
			LatencyP50:       200 * time.Millisecond,
			LatencyP99:       800 * time.Millisecond,
			IsLocal:          false,
		},
		{
			ModelID:          "std-local",
			Provider:         "test",
			Tier:             Tier2Standard,
			MaxContextTokens: 8192,
			MaxOutputTokens:  2048,
			SupportsTools:    true,
			SupportsParallel: true,
			CostPer1KInput:   0.01,
			CostPer1KOutput:  0.02,
			LatencyP50:       150 * time.Millisecond,
			LatencyP99:       600 * time.Millisecond,
			IsLocal:          true,
		},
		{
			ModelID:          "advanced-model",
			Provider:         "test",
			Tier:             Tier3Advanced,
			MaxContextTokens: 32768,
			MaxOutputTokens:  8192,
			SupportsTools:    true,
			SupportsParallel: true,
			CostPer1KInput:   0.05,
			CostPer1KOutput:  0.10,
			LatencyP50:       500 * time.Millisecond,
			LatencyP99:       2000 * time.Millisecond,
			IsLocal:          false,
		},
	}

	for _, m := range models {
		if err := reg.Register(ctx, m); err != nil {
			t.Fatalf("failed to register model %s: %v", m.ModelID, err)
		}
	}

	return router, ctx
}

func TestRouter_RouteToTier1(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SelectedModel.Tier < Tier1Constrained {
		t.Fatalf("expected at least tier 1, got tier %d", resp.SelectedModel.Tier)
	}
}

func TestRouter_RouteToTier2(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier2Standard,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SelectedModel.Tier < Tier2Standard {
		t.Fatalf("expected at least tier 2, got tier %d", resp.SelectedModel.Tier)
	}
}

func TestRouter_RouteToTier3(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier3Advanced,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SelectedModel.Tier != Tier3Advanced {
		t.Fatalf("expected tier 3, got tier %d", resp.SelectedModel.Tier)
	}
	if resp.SelectedModel.ModelID != "advanced-model" {
		t.Fatalf("expected advanced-model, got %s", resp.SelectedModel.ModelID)
	}
}

func TestRouter_PreferredModelSelection(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier3Advanced,
		PreferredModels:   []string{"advanced-model"},
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SelectedModel.ModelID != "advanced-model" {
		t.Fatalf("expected advanced-model (preferred), got %s", resp.SelectedModel.ModelID)
	}
}

func TestRouter_FallbackWhenPreferredUnavailable(t *testing.T) {
	router, ctx := setupRouter(t)

	// Request tools, so tiny-model is excluded (no tools support)
	// Preferred is nonexistent model, so it picks the next best
	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		PreferredModels:   []string{"nonexistent-model"},
		RequiresTools:     true,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SelectedModel.ModelID == "nonexistent-model" {
		t.Fatal("should not have selected nonexistent model")
	}
	// Fallback chain should be populated
	if len(resp.FallbackChain) == 0 {
		t.Fatal("expected fallback chain to be populated")
	}
}

func TestRouter_BudgetEnforcement(t *testing.T) {
	router, ctx := setupRouter(t)

	// Try to route with a very low budget that no model can fit
	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		BudgetUSD:         0.0000001, // impossibly low
		InputTokens:       1000000,
		OutputTokens:      1000000,
	})
	if err == nil {
		t.Fatalf("expected budget error, got model %s", resp.SelectedModel.ModelID)
	}
}

func TestRouter_LocalModelPreference(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		PreferLocal:       true,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.SelectedModel.IsLocal {
		t.Fatalf("expected local model, got %s (IsLocal=%v)", resp.SelectedModel.ModelID, resp.SelectedModel.IsLocal)
	}
}

func TestRouter_LatencySLAFiltering(t *testing.T) {
	router, ctx := setupRouter(t)

	// Set a very tight SLA that only tiny-model meets (50ms)
	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		LatencySLA:        75 * time.Millisecond,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SelectedModel.LatencyP50 > 75*time.Millisecond {
		t.Fatalf("expected model with P50 < 75ms, got %v", resp.SelectedModel.LatencyP50)
	}
}

func TestRouter_CostEstimation(t *testing.T) {
	ctx := context.Background()

	// Register a model with known costs
	reg := NewModelRegistry()
	model := ModelCapability{
		ModelID:          "cost-test",
		Provider:         "test",
		Tier:             Tier1Constrained,
		MaxContextTokens: 4096,
		MaxOutputTokens:  1024,
		SupportsTools:    true,
		SupportsParallel: true,
		CostPer1KInput:   0.01,
		CostPer1KOutput:  0.02,
		LatencyP50:       100 * time.Millisecond,
		LatencyP99:       500 * time.Millisecond,
		IsLocal:          false,
	}
	if err := reg.Register(ctx, model); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	r := NewModelRouter(reg)

	resp, err := r.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		InputTokens:       1000,
		OutputTokens:      1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := 0.03 // 1000/1000 * 0.01 + 1000/1000 * 0.02
	if resp.CostEstimate != expected {
		t.Fatalf("expected cost estimate %.4f, got %.4f", expected, resp.CostEstimate)
	}
}

func TestRouter_BuildFallbackChain(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		PreferredModels:   []string{"tiny-model"},
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fallback chain should have at most 2 entries
	if len(resp.FallbackChain) > 2 {
		t.Fatalf("expected at most 2 fallback models, got %d", len(resp.FallbackChain))
	}
	// Selected model should not be in fallback chain
	for _, fb := range resp.FallbackChain {
		if fb == resp.SelectedModel.ModelID {
			t.Fatal("selected model should not be in fallback chain")
		}
	}
}

func TestRouter_EmptyRegistry(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()
	router := NewModelRouter(reg)

	_, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
	})
	if err != ErrNoModelsAvailable {
		t.Fatalf("expected ErrNoModelsAvailable, got %v", err)
	}
}

func TestRouter_BudgetRecordingAndSpending(t *testing.T) {
	ctx := context.Background()
	router, _ := setupRouter(t)

	t.Run("set and get budget", func(t *testing.T) {
		budget := Budget{
			DailyUSD:   10.0,
			MonthlyUSD: 100.0,
			ResetAt:    time.Now().Add(24 * time.Hour),
		}
		if err := router.SetBudget(ctx, "user-1", budget); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := router.GetBudget(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.MonthlyUSD != 100.0 {
			t.Fatalf("expected monthly 100, got %.2f", got.MonthlyUSD)
		}
	})

	t.Run("record spend", func(t *testing.T) {
		budget := Budget{
			DailyUSD:   10.0,
			MonthlyUSD: 100.0,
			ResetAt:    time.Now().Add(24 * time.Hour),
		}
		if err := router.SetBudget(ctx, "user-2", budget); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := router.RecordSpend(ctx, "user-2", 5.0); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := router.GetBudget(ctx, "user-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.CurrentSpend != 5.0 {
			t.Fatalf("expected spend 5.0, got %.2f", got.CurrentSpend)
		}
	})

	t.Run("check budget ok", func(t *testing.T) {
		budget := Budget{
			DailyUSD:   10.0,
			MonthlyUSD: 100.0,
			ResetAt:    time.Now().Add(24 * time.Hour),
		}
		if err := router.SetBudget(ctx, "user-3", budget); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := router.CheckBudget(ctx, "user-3", 1.0)
		if err != nil {
			t.Fatalf("expected budget ok, got error: %v", err)
		}
	})

	t.Run("check budget exceeded", func(t *testing.T) {
		budget := Budget{
			DailyUSD:   1.0,
			MonthlyUSD: 100.0,
			ResetAt:    time.Now().Add(24 * time.Hour),
		}
		if err := router.SetBudget(ctx, "user-4", budget); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := router.RecordSpend(ctx, "user-4", 0.9); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := router.CheckBudget(ctx, "user-4", 0.5)
		if err == nil {
			t.Fatal("expected budget exceeded error")
		}
	})

	t.Run("budget not found", func(t *testing.T) {
		_, err := router.GetBudget(ctx, "nonexistent")
		if err != ErrBudgetNotFound {
			t.Fatalf("expected ErrBudgetNotFound, got %v", err)
		}

		err = router.RecordSpend(ctx, "nonexistent", 1.0)
		if err != ErrBudgetNotFound {
			t.Fatalf("expected ErrBudgetNotFound, got %v", err)
		}

		err = router.CheckBudget(ctx, "nonexistent", 1.0)
		if err != ErrBudgetNotFound {
			t.Fatalf("expected ErrBudgetNotFound, got %v", err)
		}
	})
}

func TestRouter_ConcurrentRouting(t *testing.T) {
	router, ctx := setupRouter(t)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = router.Route(ctx, RouteRequest{
				MinCapabilityTier: Tier1Constrained,
				PreferredModels:   []string{"tiny-model"},
				InputTokens:       500,
				OutputTokens:      200,
			})
		}()
	}

	wg.Wait()
	// Just ensure no panics or data races occurred
}

func TestRouter_RequiresTools(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		RequiresTools:     true,
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.SelectedModel.SupportsTools {
		t.Fatalf("expected model with tool support, got %s", resp.SelectedModel.ModelID)
	}
}

func TestRouter_RequiresParallel(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		RequiresTools:    true,
		RequiresParallel: true,
		InputTokens:      500,
		OutputTokens:     200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.SelectedModel.SupportsParallel {
		t.Fatalf("expected model with parallel support, got %s", resp.SelectedModel.ModelID)
	}
}

func TestRouter_CompressionNeeded(t *testing.T) {
	router, ctx := setupRouter(t)

	resp, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
		InputTokens:       10000, // Exceeds tiny-model's 2048
		OutputTokens:      200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.CompressionNeeded {
		t.Fatal("expected compression needed")
	}
}

func TestRouter_NoCandidatesAfterFiltering(t *testing.T) {
	router, ctx := setupRouter(t)

	// Request tier 3 + parallel + tools + very tight latency SLA
	_, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier3Advanced,
		RequiresTools:     true,
		RequiresParallel:  true,
		LatencySLA:        10 * time.Millisecond, // No model can meet this
		InputTokens:       500,
		OutputTokens:      200,
	})
	if err != ErrNoModelsAvailable {
		t.Fatalf("expected ErrNoModelsAvailable, got %v", err)
	}
}

func TestRouter_EstimateCost(t *testing.T) {
	router, _ := setupRouter(t)

	model := &ModelCapability{
		CostPer1KInput:  0.01,
		CostPer1KOutput: 0.02,
	}

	cases := []struct {
		name       string
		input      int
		output     int
		want       float64
	}{
		{"zero tokens", 0, 0, 0.0},
		{"1k input 1k output", 1000, 1000, 0.03},
		{"5k input 2k output", 5000, 2000, 0.09},
		{"100 input 50 output", 100, 50, 0.002},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := router.estimateCost(model, tc.input, tc.output)
			if got != tc.want {
				t.Fatalf("expected %.4f, got %.4f", tc.want, got)
			}
		})
	}
}

func TestRouter_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reg := NewModelRegistry()
	router := NewModelRouter(reg)

	_, err := router.Route(ctx, RouteRequest{
		MinCapabilityTier: Tier1Constrained,
	})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
