package cost

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBudgetManager_CreateBudget(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	tests := []struct {
		name      string
		budget    Budget
		wantError bool
	}{
		{
			name: "create daily budget",
			budget: Budget{
				ID:       "daily-1",
				Period:   BudgetDaily,
				LimitUSD: 10.0,
				Alerts: []BudgetAlert{
					{ThresholdPercent: 0.8},
				},
			},
			wantError: false,
		},
		{
			name: "create monthly budget",
			budget: Budget{
				ID:       "monthly-1",
				Period:   BudgetMonthly,
				LimitUSD: 100.0,
			},
			wantError: false,
		},
		{
			name: "empty budget ID",
			budget: Budget{
				ID:       "",
				Period:   BudgetDaily,
				LimitUSD: 10.0,
			},
			wantError: true,
		},
		{
			name: "zero limit",
			budget: Budget{
				ID:       "zero-limit",
				Period:   BudgetDaily,
				LimitUSD: 0,
			},
			wantError: true,
		},
		{
			name: "negative limit",
			budget: Budget{
				ID:       "neg-limit",
				Period:   BudgetDaily,
				LimitUSD: -5.0,
			},
			wantError: true,
		},
		{
			name: "invalid period",
			budget: Budget{
				ID:       "invalid-period",
				Period:   BudgetPeriod("weekly"),
				LimitUSD: 10.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bm.CreateBudget(ctx, tt.budget)
			if (err != nil) != tt.wantError {
				t.Errorf("CreateBudget() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				b, err := bm.GetBudget(ctx, tt.budget.ID)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if b.LimitUSD != tt.budget.LimitUSD {
					t.Errorf("got LimitUSD %v, want %v", b.LimitUSD, tt.budget.LimitUSD)
				}
				if b.CurrentSpend != 0 {
					t.Errorf("got CurrentSpend %v, want 0", b.CurrentSpend)
				}
			}
		})
	}
}

func TestBudgetManager_GetBudget(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()
	bm.CreateBudget(ctx, Budget{ID: "test-budget", Period: BudgetDaily, LimitUSD: 50.0})

	tests := []struct {
		name      string
		id        string
		wantError bool
	}{
		{
			name:      "existing budget",
			id:        "test-budget",
			wantError: false,
		},
		{
			name:      "unknown budget",
			id:        "unknown",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bm.GetBudget(ctx, tt.id)
			if (err != nil) != tt.wantError {
				t.Errorf("GetBudget() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBudgetManager_Spend(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()
	bm.CreateBudget(ctx, Budget{ID: "spend-budget", Period: BudgetDaily, LimitUSD: 100.0})

	t.Run("spend within limit", func(t *testing.T) {
		err := bm.Spend(ctx, "spend-budget", 25.0, "search", "gpt-4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, _ := bm.GetBudget(ctx, "spend-budget")
		if b.CurrentSpend != 25.0 {
			t.Errorf("got CurrentSpend %v, want 25.0", b.CurrentSpend)
		}
	})

	t.Run("spend at limit", func(t *testing.T) {
		err := bm.Spend(ctx, "spend-budget", 75.0, "search", "gpt-4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, _ := bm.GetBudget(ctx, "spend-budget")
		if b.CurrentSpend != 100.0 {
			t.Errorf("got CurrentSpend %v, want 100.0", b.CurrentSpend)
		}
	})

	t.Run("spend over limit rejected", func(t *testing.T) {
		err := bm.Spend(ctx, "spend-budget", 0.01, "search", "gpt-4")
		if err == nil {
			t.Fatal("expected error when spending over limit")
		}
	})

	t.Run("spend unknown budget", func(t *testing.T) {
		err := bm.Spend(ctx, "unknown", 10.0, "search", "gpt-4")
		if err == nil {
			t.Fatal("expected error when spending to unknown budget")
		}
	})

	t.Run("negative amount rejected", func(t *testing.T) {
		err := bm.Spend(ctx, "spend-budget", -1.0, "search", "gpt-4")
		if err == nil {
			t.Fatal("expected error when spending negative amount")
		}
	})
}

func TestBudgetManager_CheckBudget(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()
	bm.CreateBudget(ctx, Budget{ID: "check-budget", Period: BudgetDaily, LimitUSD: 50.0})
	bm.Spend(ctx, "check-budget", 30.0, "search", "gpt-4")

	tests := []struct {
		name            string
		budgetID        string
		requestedAmount float64
		wantError       bool
	}{
		{
			name:            "within budget",
			budgetID:        "check-budget",
			requestedAmount: 10.0,
			wantError:       false,
		},
		{
			name:            "exactly remaining",
			budgetID:        "check-budget",
			requestedAmount: 20.0,
			wantError:       false,
		},
		{
			name:            "over budget",
			budgetID:        "check-budget",
			requestedAmount: 21.0,
			wantError:       true,
		},
		{
			name:            "unknown budget",
			budgetID:        "unknown",
			requestedAmount: 1.0,
			wantError:       true,
		},
		{
			name:            "negative amount",
			budgetID:        "check-budget",
			requestedAmount: -1.0,
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bm.CheckBudget(ctx, tt.budgetID, tt.requestedAmount)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckBudget() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBudgetManager_AlertsTrigger(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	bm.CreateBudget(ctx, Budget{
		ID:       "alert-budget",
		Period:   BudgetDaily,
		LimitUSD: 100.0,
		Alerts: []BudgetAlert{
			{ThresholdPercent: 0.5},
			{ThresholdPercent: 0.8},
			{ThresholdPercent: 1.0},
		},
	})

	// Spend 50 - should trigger 50% alert
	err := bm.Spend(ctx, "alert-budget", 50.0, "search", "gpt-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := bm.GetBudget(ctx, "alert-budget")
	if !b.Alerts[0].Triggered {
		t.Error("50% alert should have triggered")
	}
	if b.Alerts[1].Triggered {
		t.Error("80% alert should not have triggered yet")
	}

	// Spend 30 more (total 80) - should trigger 80% alert
	err = bm.Spend(ctx, "alert-budget", 30.0, "search", "gpt-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ = bm.GetBudget(ctx, "alert-budget")
	if !b.Alerts[1].Triggered {
		t.Error("80% alert should have triggered")
	}
	if b.Alerts[2].Triggered {
		t.Error("100% alert should not have triggered yet")
	}
}

func TestBudgetManager_GetTransactions(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	bm.CreateBudget(ctx, Budget{ID: "tx-budget-1", Period: BudgetDaily, LimitUSD: 100.0})
	bm.CreateBudget(ctx, Budget{ID: "tx-budget-2", Period: BudgetDaily, LimitUSD: 100.0})

	bm.Spend(ctx, "tx-budget-1", 10.0, "search", "gpt-4")
	bm.Spend(ctx, "tx-budget-1", 20.0, "agent", "claude-sonnet")
	bm.Spend(ctx, "tx-budget-2", 5.0, "chat", "gpt-3.5-turbo")

	txs := bm.GetTransactions(ctx, "tx-budget-1")
	if len(txs) != 2 {
		t.Fatalf("got %d transactions, want 2", len(txs))
	}
	for _, tx := range txs {
		if tx.BudgetID != "tx-budget-1" {
			t.Errorf("unexpected budget ID %q", tx.BudgetID)
		}
	}

	txs2 := bm.GetTransactions(ctx, "tx-budget-2")
	if len(txs2) != 1 {
		t.Fatalf("got %d transactions, want 1", len(txs2))
	}
}

func TestBudgetManager_ListBudgets(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	bm.CreateBudget(ctx, Budget{ID: "list-1", Period: BudgetDaily, LimitUSD: 10.0})
	bm.CreateBudget(ctx, Budget{ID: "list-2", Period: BudgetMonthly, LimitUSD: 100.0})

	budgets := bm.ListBudgets(ctx)
	if len(budgets) != 2 {
		t.Fatalf("got %d budgets, want 2", len(budgets))
	}

	ids := make(map[string]bool)
	for _, b := range budgets {
		ids[b.ID] = true
	}
	if !ids["list-1"] || !ids["list-2"] {
		t.Error("missing expected budgets")
	}
}

func TestBudgetManager_DeleteBudget(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	bm.CreateBudget(ctx, Budget{ID: "delete-me", Period: BudgetDaily, LimitUSD: 10.0})

	t.Run("delete existing budget", func(t *testing.T) {
		err := bm.DeleteBudget(ctx, "delete-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = bm.GetBudget(ctx, "delete-me")
		if err == nil {
			t.Fatal("expected error after deletion")
		}
	})

	t.Run("delete unknown budget", func(t *testing.T) {
		err := bm.DeleteBudget(ctx, "unknown")
		if err == nil {
			t.Fatal("expected error when deleting unknown budget")
		}
	})
}

func TestBudgetManager_ResetAfterPeriod(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	// Create a budget with a reset time in the past to simulate expired budget
	bm.mu.Lock()
	bm.budgets["reset-budget"] = &Budget{
		ID:           "reset-budget",
		Period:       BudgetDaily,
		LimitUSD:     100.0,
		CurrentSpend: 80.0,
		ResetAt:      time.Now().Add(-1 * time.Second), // already past
		Alerts: []BudgetAlert{
			{ThresholdPercent: 0.5, Triggered: true, TriggeredAt: time.Now()},
		},
	}
	bm.mu.Unlock()

	// Spend should trigger a reset
	err := bm.Spend(ctx, "reset-budget", 10.0, "search", "gpt-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := bm.GetBudget(ctx, "reset-budget")
	if b.CurrentSpend != 10.0 {
		t.Errorf("got CurrentSpend %v, want 10.0 (after reset)", b.CurrentSpend)
	}
	if b.Alerts[0].Triggered {
		t.Error("alerts should have been reset")
	}
}

func TestBudgetManager_ConcurrentSpend(t *testing.T) {
	bm := NewBudgetManager()
	ctx := context.Background()

	bm.CreateBudget(ctx, Budget{ID: "concurrent-budget", Period: BudgetDaily, LimitUSD: 1000.0})

	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bm.Spend(ctx, "concurrent-budget", 1.0, "search", "gpt-4")
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	b, _ := bm.GetBudget(ctx, "concurrent-budget")
	if b.CurrentSpend != float64(successCount) {
		t.Errorf("got CurrentSpend %v, want %v", b.CurrentSpend, successCount)
	}
	if successCount != 100 {
		t.Errorf("got %d successful spends, want 100", successCount)
	}
}

func TestBudgetManager_ContextCancellation(t *testing.T) {
	bm := NewBudgetManager()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	bm.CreateBudget(context.Background(), Budget{ID: "cancel-budget", Period: BudgetDaily, LimitUSD: 10.0})

	if _, err := bm.GetBudget(ctx, "cancel-budget"); err == nil {
		t.Error("GetBudget() expected error on cancelled context")
	}
	if err := bm.Spend(ctx, "cancel-budget", 1.0, "search", "gpt-4"); err == nil {
		t.Error("Spend() expected error on cancelled context")
	}
	if err := bm.CheckBudget(ctx, "cancel-budget", 1.0); err == nil {
		t.Error("CheckBudget() expected error on cancelled context")
	}
	if err := bm.DeleteBudget(ctx, "cancel-budget"); err == nil {
		t.Error("DeleteBudget() expected error on cancelled context")
	}
}
