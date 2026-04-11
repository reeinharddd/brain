package cost

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCostEstimator_SetPricing(t *testing.T) {
	ce := NewCostEstimator()
	ctx := context.Background()

	tests := []struct {
		name      string
		pricing   ModelPricing
		wantError bool
	}{
		{
			name:      "set new pricing",
			pricing:   ModelPricing{ModelID: "custom-model", CostPer1KInput: 0.01, CostPer1KOutput: 0.02},
			wantError: false,
		},
		{
			name:      "empty model ID",
			pricing:   ModelPricing{ModelID: "", CostPer1KInput: 0.01},
			wantError: true,
		},
		{
			name:      "update existing pricing",
			pricing:   ModelPricing{ModelID: "gpt-4", CostPer1KInput: 0.05, CostPer1KOutput: 0.10},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ce.SetPricing(ctx, tt.pricing)
			if (err != nil) != tt.wantError {
				t.Errorf("SetPricing() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				p, err := ce.GetPricing(ctx, tt.pricing.ModelID)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if p.ModelID != tt.pricing.ModelID {
					t.Errorf("got ModelID %q, want %q", p.ModelID, tt.pricing.ModelID)
				}
			}
		})
	}
}

func TestCostEstimator_GetPricing(t *testing.T) {
	ce := NewCostEstimator()
	ctx := context.Background()

	tests := []struct {
		name      string
		modelID   string
		wantError bool
	}{
		{
			name:      "existing default model",
			modelID:   "gpt-4",
			wantError: false,
		},
		{
			name:      "unknown model",
			modelID:   "unknown-model",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ce.GetPricing(ctx, tt.modelID)
			if (err != nil) != tt.wantError {
				t.Errorf("GetPricing() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if p == nil {
					t.Fatal("expected non-nil pricing")
				}
				if p.ModelID != tt.modelID {
					t.Errorf("got modelID %q, want %q", p.ModelID, tt.modelID)
				}
			}
		})
	}
}

func TestCostEstimator_Estimate(t *testing.T) {
	ce := NewCostEstimator()
	ctx := context.Background()

	tests := []struct {
		name        string
		modelID     string
		inputTokens int
		outputTokens int
		wantCost    float64
		wantError   bool
	}{
		{
			name:        "gpt-4 cost calculation",
			modelID:     "gpt-4",
			inputTokens: 1000,
			outputTokens: 1000,
			wantCost:    0.09, // 0.03 + 0.06
			wantError:   false,
		},
		{
			name:        "local model is free",
			modelID:     "local-model",
			inputTokens: 1000,
			outputTokens: 1000,
			wantCost:    0,
			wantError:   false,
		},
		{
			name:        "zero tokens",
			modelID:     "gpt-4",
			inputTokens: 0,
			outputTokens: 0,
			wantCost:    0,
			wantError:   false,
		},
		{
			name:        "unknown model",
			modelID:     "unknown",
			inputTokens: 1000,
			outputTokens: 1000,
			wantCost:    0,
			wantError:   true,
		},
		{
			name:        "claude sonnet cost",
			modelID:     "claude-sonnet",
			inputTokens: 2000,
			outputTokens: 3000,
			wantCost:    0.051, // 2*0.003 + 3*0.015
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ce.Estimate(ctx, tt.modelID, tt.inputTokens, tt.outputTokens)
			if (err != nil) != tt.wantError {
				t.Errorf("Estimate() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && got != tt.wantCost {
				t.Errorf("Estimate() = %v, want %v", got, tt.wantCost)
			}
		})
	}
}

func TestCostEstimator_ListPricing(t *testing.T) {
	ce := NewCostEstimator()
	ctx := context.Background()

	pricing := ce.ListPricing(ctx)
	if len(pricing) != len(DefaultPricing) {
		t.Errorf("ListPricing() got %d items, want %d", len(pricing), len(DefaultPricing))
	}

	// Verify all default models are present
	modelIDs := make(map[string]bool)
	for _, p := range pricing {
		modelIDs[p.ModelID] = true
	}
	for modelID := range DefaultPricing {
		if !modelIDs[modelID] {
			t.Errorf("ListPricing() missing default model %q", modelID)
		}
	}
}

func TestCostEstimator_DefaultPricing(t *testing.T) {
	ce := NewCostEstimator()
	ctx := context.Background()

	// Verify default pricing is loaded correctly
	p, err := ce.GetPricing(ctx, "gpt-4")
	if err != nil {
		t.Fatalf("unexpected error getting default pricing: %v", err)
	}
	if p.CostPer1KInput != 0.03 {
		t.Errorf("gpt-4 CostPer1KInput = %v, want 0.03", p.CostPer1KInput)
	}
	if p.CostPer1KOutput != 0.06 {
		t.Errorf("gpt-4 CostPer1KOutput = %v, want 0.06", p.CostPer1KOutput)
	}
}

func TestCostEstimator_ConcurrentAccess(t *testing.T) {
	ce := NewCostEstimator()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(i int) {
			defer wg.Done()
			_ = ce.SetPricing(ctx, ModelPricing{
				ModelID:        "concurrent-model",
				CostPer1KInput: float64(i) * 0.001,
				CostPer1KOutput: float64(i) * 0.002,
			})
		}(i)

		go func() {
			defer wg.Done()
			_, _ = ce.GetPricing(ctx, "gpt-4")
		}()

		go func() {
			defer wg.Done()
			_, _ = ce.Estimate(ctx, "gpt-4", 100, 200)
		}()
	}
	wg.Wait()
}

func TestCostEstimator_ContextCancellation(t *testing.T) {
	ce := NewCostEstimator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ce.GetPricing(ctx, "gpt-4"); err == nil {
		t.Error("GetPricing() expected error on cancelled context, got nil")
	}
	if _, err := ce.Estimate(ctx, "gpt-4", 100, 100); err == nil {
		t.Error("Estimate() expected error on cancelled context, got nil")
	}
	if err := ce.SetPricing(ctx, ModelPricing{ModelID: "test", CostPer1KInput: 0.01}); err == nil {
		t.Error("SetPricing() expected error on cancelled context, got nil")
	}
}

func TestCostEstimator_ContextTimeout(t *testing.T) {
	ce := NewCostEstimator()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // ensure timeout fires

	if _, err := ce.Estimate(ctx, "gpt-4", 100, 100); err == nil {
		t.Error("Estimate() expected error on timed-out context, got nil")
	}
}
