package cost

import (
	"context"
	"testing"
	"time"
)

func TestCostOptimizer_AnalyzeUsage(t *testing.T) {
	ce := NewCostEstimator()
	co := NewCostOptimizer(ce)
	ctx := context.Background()

	now := time.Now()

	tests := []struct {
		name             string
		usage            []UsageRecord
		minRecommendations int
		checkTypes       []string
	}{
		{
			name: "expensive model triggers model_switch",
			usage: []UsageRecord{
				{
					ModelID:      "claude-opus",
					InputTokens:  5000,
					OutputTokens: 500,
					Cost:         0.1125,
					Surface:      "search",
					Timestamp:    now,
				},
			},
			minRecommendations: 1,
			checkTypes:         []string{"model_switch"},
		},
		{
			name: "repeated queries trigger cache_opportunity",
			usage: []UsageRecord{
				{
					ModelID:      "gpt-4",
					InputTokens:  1000,
					OutputTokens: 100,
					Cost:         0.036,
					Surface:      "chat",
					Timestamp:    now,
				},
				{
					ModelID:      "gpt-4",
					InputTokens:  1000,
					OutputTokens: 100,
					Cost:         0.036,
					Surface:      "chat",
					Timestamp:    now.Add(1 * time.Minute),
				},
				{
					ModelID:      "gpt-4",
					InputTokens:  1000,
					OutputTokens: 100,
					Cost:         0.036,
					Surface:      "chat",
					Timestamp:    now.Add(2 * time.Minute),
				},
			},
			minRecommendations: 1,
			checkTypes:         []string{"cache_opportunity"},
		},
		{
			name: "high input/output ratio triggers context_reduction",
			usage: []UsageRecord{
				{
					ModelID:      "gpt-4",
					InputTokens:  50000,
					OutputTokens: 100,
					Cost:         1.56,
					Surface:      "agent",
					Timestamp:    now,
				},
			},
			minRecommendations: 1,
			checkTypes:         []string{"context_reduction"},
		},
		{
			name: "many small requests trigger batch_requests",
			usage: func() []UsageRecord {
				var records []UsageRecord
				for i := 0; i < 15; i++ {
					records = append(records, UsageRecord{
						ModelID:      "gpt-3.5-turbo",
						InputTokens:  100,
						OutputTokens: 50,
						Cost:         0.00025,
						Surface:      "search",
						Timestamp:    now.Add(time.Duration(i) * time.Minute),
					})
				}
				return records
			}(),
			minRecommendations: 1,
			checkTypes:         []string{"batch_requests"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recs := co.AnalyzeUsage(ctx, tt.usage)
			if len(recs) < tt.minRecommendations {
				t.Errorf("got %d recommendations, want at least %d", len(recs), tt.minRecommendations)
			}

			// Check that expected types are present
			typeSet := make(map[string]bool)
			for _, r := range recs {
				typeSet[r.Type] = true
			}
			for _, expectedType := range tt.checkTypes {
				if !typeSet[expectedType] {
					t.Errorf("missing recommendation type %q", expectedType)
				}
			}
		})
	}
}

func TestCostOptimizer_SuggestModelSwitch(t *testing.T) {
	ce := NewCostEstimator()
	co := NewCostOptimizer(ce)
	ctx := context.Background()

	tests := []struct {
		name          string
		currentModel  string
		budget        float64
		wantRec       bool
		wantSwitchTo  string
	}{
		{
			name:         "expensive model has cheaper alternative",
			currentModel: "claude-opus",
			budget:       10.0,
			wantRec:      true,
			wantSwitchTo: "local-model",
		},
		{
			name:         "already cheapest model",
			currentModel: "local-model",
			budget:       10.0,
			wantRec:      false,
		},
		{
			name:         "unknown model",
			currentModel: "unknown-model",
			budget:       10.0,
			wantRec:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := co.SuggestModelSwitch(ctx, tt.currentModel, tt.budget)
			if (rec != nil) != tt.wantRec {
				t.Errorf("SuggestModelSwitch() got rec = %v, want %v", rec != nil, tt.wantRec)
			}
			if rec != nil {
				if rec.Type != "model_switch" {
					t.Errorf("got type %q, want %q", rec.Type, "model_switch")
				}
				if rec.EstimatedSavings <= 0 {
					t.Errorf("expected positive savings, got %v", rec.EstimatedSavings)
				}
				if !rec.Actionable {
					t.Error("expected actionable recommendation")
				}
			}
		})
	}
}

func TestCostOptimizer_EstimateCacheSavings(t *testing.T) {
	ce := NewCostEstimator()
	co := NewCostOptimizer(ce)
	ctx := context.Background()

	tests := []struct {
		name           string
		repeatedQueries int
		totalTokens    int
		wantRec        bool
	}{
		{
			name:           "valid repeated queries",
			repeatedQueries: 10,
			totalTokens:    50000,
			wantRec:        true,
		},
		{
			name:           "zero repeated queries",
			repeatedQueries: 0,
			totalTokens:    50000,
			wantRec:        false,
		},
		{
			name:           "zero tokens",
			repeatedQueries: 10,
			totalTokens:    0,
			wantRec:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := co.EstimateCacheSavings(ctx, tt.repeatedQueries, tt.totalTokens)
			if (rec != nil) != tt.wantRec {
				t.Errorf("EstimateCacheSavings() got rec = %v, want %v", rec != nil, tt.wantRec)
			}
			if rec != nil {
				if rec.Type != "cache_opportunity" {
					t.Errorf("got type %q, want %q", rec.Type, "cache_opportunity")
				}
				if rec.EstimatedSavings <= 0 {
					t.Errorf("expected positive savings, got %v", rec.EstimatedSavings)
				}
			}
		})
	}
}

func TestCostOptimizer_NoRecommendationsWhenOptimal(t *testing.T) {
	ce := NewCostEstimator()
	co := NewCostOptimizer(ce)
	ctx := context.Background()

	// Cheap model, reasonable I/O ratio, few requests, no repetition
	usage := []UsageRecord{
		{
			ModelID:      "claude-haiku",
			InputTokens:  500,
			OutputTokens: 300,
			Cost:         0.0005,
			Surface:      "chat",
			Timestamp:    time.Now(),
		},
	}

	recs := co.AnalyzeUsage(ctx, usage)
	if len(recs) != 0 {
		t.Errorf("got %d recommendations, want 0 for optimal usage", len(recs))
		for _, r := range recs {
			t.Logf("unexpected recommendation: %s - %s", r.Type, r.Title)
		}
	}
}

func TestCostOptimizer_EmptyUsage(t *testing.T) {
	ce := NewCostEstimator()
	co := NewCostOptimizer(ce)
	ctx := context.Background()

	recs := co.AnalyzeUsage(ctx, nil)
	if recs != nil {
		t.Errorf("got %d recommendations, want nil for empty usage", len(recs))
	}

	recs = co.AnalyzeUsage(ctx, []UsageRecord{})
	if recs != nil {
		t.Errorf("got %d recommendations, want nil for empty slice", len(recs))
	}
}

func TestCostOptimizer_ContextCancellation(t *testing.T) {
	ce := NewCostEstimator()
	co := NewCostOptimizer(ce)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := co.SuggestModelSwitch(ctx, "gpt-4", 10.0)
	if rec != nil {
		t.Error("SuggestModelSwitch() expected nil on cancelled context")
	}

	rec = co.EstimateCacheSavings(ctx, 10, 5000)
	if rec != nil {
		t.Error("EstimateCacheSavings() expected nil on cancelled context")
	}

	recs := co.AnalyzeUsage(ctx, []UsageRecord{
		{ModelID: "gpt-4", InputTokens: 1000, Cost: 0.03, Timestamp: time.Now()},
	})
	if recs != nil {
		t.Error("AnalyzeUsage() expected nil on cancelled context")
	}
}
