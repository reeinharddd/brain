package delegation

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewBudgetTracker(t *testing.T) {
	t.Run("creates tracker", func(t *testing.T) {
		budget := DelegationBudget{
			MaxTokens:   1000,
			MaxCostUSD:  10.0,
			MaxDuration: 5 * time.Minute,
			MaxRetries:  3,
		}
		bt := NewBudgetTracker(budget)
		if bt == nil {
			t.Fatal("expected non-nil budget tracker")
		}
	})
}

func TestRecordTokenUsage(t *testing.T) {
	tests := []struct {
		name      string
		maxTokens int
		usage     []int
		wantErr   bool
		errIs     error
	}{
		{
			name:      "within budget",
			maxTokens: 1000,
			usage:     []int{100, 200, 300},
			wantErr:   false,
		},
		{
			name:      "exceeds budget",
			maxTokens: 100,
			usage:     []int{50, 60},
			wantErr:   true,
			errIs:     ErrTokenLimitExceeded,
		},
		{
			name:      "unlimited budget",
			maxTokens: 0,
			usage:     []int{1000000},
			wantErr:   false,
		},
		{
			name:      "negative usage",
			maxTokens: 1000,
			usage:     []int{-100},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := DelegationBudget{MaxTokens: tt.maxTokens}
			bt := NewBudgetTracker(budget)

			var lastErr error
			for _, u := range tt.usage {
				lastErr = bt.RecordTokenUsage(u)
			}

			if tt.wantErr {
				if lastErr == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(lastErr, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, lastErr)
				}
			} else {
				if lastErr != nil {
					t.Fatalf("unexpected error: %v", lastErr)
				}
			}
		})
	}
}

func TestRecordCost(t *testing.T) {
	tests := []struct {
		name       string
		maxCost    float64
		costs      []float64
		wantErr    bool
		errIs      error
	}{
		{
			name:    "within budget",
			maxCost: 10.0,
			costs:   []float64{1.0, 2.0, 3.0},
			wantErr: false,
		},
		{
			name:    "exceeds budget",
			maxCost: 5.0,
			costs:   []float64{3.0, 3.0},
			wantErr: true,
			errIs:   ErrCostLimitExceeded,
		},
		{
			name:    "unlimited budget",
			maxCost: 0,
			costs:   []float64{1000000.0},
			wantErr: false,
		},
		{
			name:    "negative cost",
			maxCost: 10.0,
			costs:   []float64{-5.0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := DelegationBudget{MaxCostUSD: tt.maxCost}
			bt := NewBudgetTracker(budget)

			var lastErr error
			for _, c := range tt.costs {
				lastErr = bt.RecordCost(c)
			}

			if tt.wantErr {
				if lastErr == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(lastErr, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, lastErr)
				}
			} else {
				if lastErr != nil {
					t.Fatalf("unexpected error: %v", lastErr)
				}
			}
		})
	}
}

func TestRecordRetry(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		retries    int
		wantErr    bool
		errIs      error
	}{
		{
			name:       "within budget",
			maxRetries: 3,
			retries:    2,
			wantErr:    false,
		},
		{
			name:       "exceeds budget",
			maxRetries: 2,
			retries:    3,
			wantErr:    true,
			errIs:      ErrRetryLimitExceeded,
		},
		{
			name:       "unlimited retries",
			maxRetries: 0,
			retries:    1000,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := DelegationBudget{MaxRetries: tt.maxRetries}
			bt := NewBudgetTracker(budget)

			var lastErr error
			for i := 0; i < tt.retries; i++ {
				lastErr = bt.RecordRetry()
			}

			if tt.wantErr {
				if lastErr == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(lastErr, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, lastErr)
				}
			} else {
				if lastErr != nil {
					t.Fatalf("unexpected error: %v", lastErr)
				}
			}
		})
	}
}

func TestCheckBudget(t *testing.T) {
	t.Run("all within budget", func(t *testing.T) {
		budget := DelegationBudget{
			MaxTokens:   1000,
			MaxCostUSD:  10.0,
			MaxDuration: 1 * time.Hour,
			MaxRetries:  5,
		}
		bt := NewBudgetTracker(budget)

		bt.RecordTokenUsage(100)
		bt.RecordCost(1.0)
		bt.RecordRetry()

		err := bt.CheckBudget()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("token limit exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxTokens: 100}
		bt := NewBudgetTracker(budget)
		bt.RecordTokenUsage(150)

		err := bt.CheckBudget()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrTokenLimitExceeded) {
			t.Errorf("expected ErrTokenLimitExceeded, got %v", err)
		}
	})

	t.Run("cost limit exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxCostUSD: 1.0}
		bt := NewBudgetTracker(budget)
		bt.RecordCost(2.0)

		err := bt.CheckBudget()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrCostLimitExceeded) {
			t.Errorf("expected ErrCostLimitExceeded, got %v", err)
		}
	})

	t.Run("duration limit exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxDuration: 1 * time.Millisecond}
		bt := NewBudgetTracker(budget)
		time.Sleep(10 * time.Millisecond)

		err := bt.CheckBudget()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrDurationLimitExceeded) {
			t.Errorf("expected ErrDurationLimitExceeded, got %v", err)
		}
	})

	t.Run("retry limit exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxRetries: 1}
		bt := NewBudgetTracker(budget)
		bt.RecordRetry()
		bt.RecordRetry()

		err := bt.CheckBudget()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrRetryLimitExceeded) {
			t.Errorf("expected ErrRetryLimitExceeded, got %v", err)
		}
	})
}

func TestGetRemaining(t *testing.T) {
	t.Run("returns remaining values", func(t *testing.T) {
		budget := DelegationBudget{
			MaxTokens:   1000,
			MaxCostUSD:  10.0,
			MaxDuration: 1 * time.Hour,
			MaxRetries:  5,
		}
		bt := NewBudgetTracker(budget)
		bt.RecordTokenUsage(200)
		bt.RecordCost(2.0)
		bt.RecordRetry()

		remaining := bt.GetRemaining()

		tokens, ok := remaining["tokens"].(int)
		if !ok || tokens != 800 {
			t.Errorf("expected 800 tokens remaining, got %v", remaining["tokens"])
		}

		cost, ok := remaining["cost"].(float64)
		if !ok || cost != 8.0 {
			t.Errorf("expected 8.0 cost remaining, got %v", remaining["cost"])
		}

		retries, ok := remaining["retries"].(int)
		if !ok || retries != 4 {
			t.Errorf("expected 4 retries remaining, got %v", remaining["retries"])
		}

		tokensUsed, ok := remaining["tokens_used"].(int)
		if !ok || tokensUsed != 200 {
			t.Errorf("expected 200 tokens used, got %v", remaining["tokens_used"])
		}
	})

	t.Run("unlimited budget shows -1", func(t *testing.T) {
		budget := DelegationBudget{}
		bt := NewBudgetTracker(budget)
		remaining := bt.GetRemaining()

		if remaining["tokens"] != -1 {
			t.Errorf("expected -1 for unlimited tokens, got %v", remaining["tokens"])
		}
		if remaining["cost"] != -1.0 {
			t.Errorf("expected -1.0 for unlimited cost, got %v", remaining["cost"])
		}
		if remaining["retries"] != -1 {
			t.Errorf("expected -1 for unlimited retries, got %v", remaining["retries"])
		}
	})
}

func TestIsExceeded(t *testing.T) {
	t.Run("not exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxTokens: 1000, MaxCostUSD: 10.0}
		bt := NewBudgetTracker(budget)
		bt.RecordTokenUsage(100)
		bt.RecordCost(1.0)

		if bt.IsExceeded() {
			t.Error("expected budget not exceeded")
		}
	})

	t.Run("tokens exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxTokens: 100}
		bt := NewBudgetTracker(budget)
		bt.RecordTokenUsage(150)

		if !bt.IsExceeded() {
			t.Error("expected budget exceeded")
		}
	})

	t.Run("cost exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxCostUSD: 1.0}
		bt := NewBudgetTracker(budget)
		bt.RecordCost(2.0)

		if !bt.IsExceeded() {
			t.Error("expected budget exceeded")
		}
	})

	t.Run("duration exceeded", func(t *testing.T) {
		budget := DelegationBudget{MaxDuration: 1 * time.Millisecond}
		bt := NewBudgetTracker(budget)
		time.Sleep(10 * time.Millisecond)

		if !bt.IsExceeded() {
			t.Error("expected budget exceeded")
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("concurrent token recording", func(t *testing.T) {
		budget := DelegationBudget{MaxTokens: 1000000}
		bt := NewBudgetTracker(budget)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bt.RecordTokenUsage(1)
			}()
		}
		wg.Wait()

		if bt.GetTokensUsed() != 100 {
			t.Errorf("expected 100 tokens used, got %d", bt.GetTokensUsed())
		}
	})

	t.Run("concurrent cost recording", func(t *testing.T) {
		budget := DelegationBudget{MaxCostUSD: 1000000.0}
		bt := NewBudgetTracker(budget)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bt.RecordCost(1.0)
			}()
		}
		wg.Wait()

		if bt.GetCostUsed() != 100.0 {
			t.Errorf("expected 100.0 cost used, got %f", bt.GetCostUsed())
		}
	})

	t.Run("concurrent mixed operations", func(t *testing.T) {
		budget := DelegationBudget{
			MaxTokens:   1000000,
			MaxCostUSD:  1000000.0,
			MaxRetries:  1000000,
		}
		bt := NewBudgetTracker(budget)

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(3)
			go func() {
				defer wg.Done()
				bt.RecordTokenUsage(1)
			}()
			go func() {
				defer wg.Done()
				bt.RecordCost(1.0)
			}()
			go func() {
				defer wg.Done()
				bt.RecordRetry()
			}()
		}
		wg.Wait()

		if bt.GetTokensUsed() != 50 {
			t.Errorf("expected 50 tokens used, got %d", bt.GetTokensUsed())
		}
		if bt.GetCostUsed() != 50.0 {
			t.Errorf("expected 50.0 cost used, got %f", bt.GetCostUsed())
		}
		if bt.GetRetriesUsed() != 50 {
			t.Errorf("expected 50 retries used, got %d", bt.GetRetriesUsed())
		}
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		budget := DelegationBudget{MaxTokens: 1000000}
		bt := NewBudgetTracker(budget)

		var wg sync.WaitGroup
		// Writers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bt.RecordTokenUsage(1)
			}()
		}
		// Readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bt.GetRemaining()
				bt.IsExceeded()
				bt.CheckBudget()
			}()
		}
		wg.Wait()
	})
}

func TestGetters(t *testing.T) {
	t.Run("GetTokensUsed", func(t *testing.T) {
		budget := DelegationBudget{MaxTokens: 1000}
		bt := NewBudgetTracker(budget)
		bt.RecordTokenUsage(100)
		bt.RecordTokenUsage(50)

		if bt.GetTokensUsed() != 150 {
			t.Errorf("expected 150 tokens used, got %d", bt.GetTokensUsed())
		}
	})

	t.Run("GetCostUsed", func(t *testing.T) {
		budget := DelegationBudget{MaxCostUSD: 10.0}
		bt := NewBudgetTracker(budget)
		bt.RecordCost(1.5)
		bt.RecordCost(2.5)

		if bt.GetCostUsed() != 4.0 {
			t.Errorf("expected 4.0 cost used, got %f", bt.GetCostUsed())
		}
	})

	t.Run("GetRetriesUsed", func(t *testing.T) {
		budget := DelegationBudget{MaxRetries: 5}
		bt := NewBudgetTracker(budget)
		bt.RecordRetry()
		bt.RecordRetry()
		bt.RecordRetry()

		if bt.GetRetriesUsed() != 3 {
			t.Errorf("expected 3 retries used, got %d", bt.GetRetriesUsed())
		}
	})
}
