package cost

import (
	"context"
	"testing"
	"time"
)

func TestCostReporter_GenerateReport(t *testing.T) {
	now := time.Now()
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.03, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "tx2", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-3.5-turbo", Surface: "search", Timestamp: now.Add(-30 * time.Minute)},
		{ID: "tx3", BudgetID: "b2", Amount: 0.05, ModelID: "claude-sonnet", Surface: "agent", Timestamp: now.Add(-15 * time.Minute)},
	}

	cr := NewCostReporter(txs)
	ctx := context.Background()

	report := cr.GenerateReport(ctx, "daily")
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Period != "daily" {
		t.Errorf("got period %q, want %q", report.Period, "daily")
	}

	expectedTotal := 0.03 + 0.01 + 0.05
	if report.TotalCost != expectedTotal {
		t.Errorf("got TotalCost %v, want %v", report.TotalCost, expectedTotal)
	}
}

func TestCostReporter_ByModelBreakdown(t *testing.T) {
	now := time.Now()
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.03, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "tx2", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-3.5-turbo", Surface: "search", Timestamp: now.Add(-30 * time.Minute)},
		{ID: "tx3", BudgetID: "b2", Amount: 0.06, ModelID: "gpt-4", Surface: "agent", Timestamp: now.Add(-15 * time.Minute)},
	}

	cr := NewCostReporter(txs)
	ctx := context.Background()

	report := cr.GenerateReport(ctx, "daily")

	if len(report.ByModel) != 2 {
		t.Fatalf("got %d models, want 2", len(report.ByModel))
	}

	gpt4 := report.ByModel["gpt-4"]
	if gpt4.ModelID != "gpt-4" {
		t.Errorf("got model ID %q, want %q", gpt4.ModelID, "gpt-4")
	}
	if gpt4.TotalCost != 0.09 {
		t.Errorf("got gpt-4 TotalCost %v, want 0.09", gpt4.TotalCost)
	}
	if gpt4.CallCount != 2 {
		t.Errorf("got gpt-4 CallCount %d, want 2", gpt4.CallCount)
	}
	if gpt4.AvgCostPerCall != 0.045 {
		t.Errorf("got gpt-4 AvgCostPerCall %v, want 0.045", gpt4.AvgCostPerCall)
	}

	turbo := report.ByModel["gpt-3.5-turbo"]
	if turbo.TotalCost != 0.01 {
		t.Errorf("got gpt-3.5-turbo TotalCost %v, want 0.01", turbo.TotalCost)
	}
}

func TestCostReporter_BySurfaceBreakdown(t *testing.T) {
	now := time.Now()
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.03, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "tx2", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-3.5-turbo", Surface: "search", Timestamp: now.Add(-30 * time.Minute)},
		{ID: "tx3", BudgetID: "b2", Amount: 0.05, ModelID: "claude-sonnet", Surface: "agent", Timestamp: now.Add(-15 * time.Minute)},
	}

	cr := NewCostReporter(txs)
	ctx := context.Background()

	report := cr.GenerateReport(ctx, "daily")

	if len(report.BySurface) != 2 {
		t.Fatalf("got %d surfaces, want 2", len(report.BySurface))
	}

	search := report.BySurface["search"]
	if search.TotalCost != 0.04 {
		t.Errorf("got search TotalCost %v, want 0.04", search.TotalCost)
	}
	if search.CallCount != 2 {
		t.Errorf("got search CallCount %d, want 2", search.CallCount)
	}

	agent := report.BySurface["agent"]
	if agent.TotalCost != 0.05 {
		t.Errorf("got agent TotalCost %v, want 0.05", agent.TotalCost)
	}
}

func TestCostReporter_TopExpenses(t *testing.T) {
	now := time.Now()
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-3.5-turbo", Surface: "search", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "tx2", BudgetID: "b1", Amount: 0.10, ModelID: "gpt-4", Surface: "agent", Timestamp: now.Add(-30 * time.Minute)},
		{ID: "tx3", BudgetID: "b2", Amount: 0.05, ModelID: "claude-sonnet", Surface: "chat", Timestamp: now.Add(-15 * time.Minute)},
		{ID: "tx4", BudgetID: "b2", Amount: 0.02, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-10 * time.Minute)},
	}

	cr := NewCostReporter(txs)
	ctx := context.Background()

	t.Run("top 3 expenses", func(t *testing.T) {
		expenses := cr.GetTopExpenses(ctx, 3)
		if len(expenses) != 3 {
			t.Fatalf("got %d expenses, want 3", len(expenses))
		}
		// Should be sorted by cost descending
		if expenses[0].Cost != 0.10 {
			t.Errorf("got top expense cost %v, want 0.10", expenses[0].Cost)
		}
		if expenses[1].Cost != 0.05 {
			t.Errorf("got second expense cost %v, want 0.05", expenses[1].Cost)
		}
		if expenses[2].Cost != 0.02 {
			t.Errorf("got third expense cost %v, want 0.02", expenses[2].Cost)
		}
	})

	t.Run("top more than available", func(t *testing.T) {
		expenses := cr.GetTopExpenses(ctx, 100)
		if len(expenses) != 4 {
			t.Fatalf("got %d expenses, want 4", len(expenses))
		}
	})
}

func TestCostReporter_TrendGeneration(t *testing.T) {
	now := time.Now()
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-3 * time.Hour)},
		{ID: "tx2", BudgetID: "b1", Amount: 0.02, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-3*time.Hour + 30*time.Minute)},
		{ID: "tx3", BudgetID: "b2", Amount: 0.05, ModelID: "claude-sonnet", Surface: "agent", Timestamp: now.Add(-2 * time.Hour)},
		{ID: "tx4", BudgetID: "b2", Amount: 0.03, ModelID: "gpt-4", Surface: "chat", Timestamp: now.Add(-1 * time.Hour)},
	}

	cr := NewCostReporter(txs)
	ctx := context.Background()

	trends := cr.GetTrend(ctx)
	if len(trends) == 0 {
		t.Fatal("expected non-empty trends")
	}

	// Trends should be sorted by timestamp
	for i := 1; i < len(trends); i++ {
		if trends[i].Timestamp.Before(trends[i-1].Timestamp) {
			t.Errorf("trends not sorted: %v before %v", trends[i].Timestamp, trends[i-1].Timestamp)
		}
	}

	// Check that costs are aggregated properly
	totalTrendCost := 0.0
	for _, tp := range trends {
		totalTrendCost += tp.Cost
	}
	expectedTotal := 0.01 + 0.02 + 0.05 + 0.03
	if totalTrendCost != expectedTotal {
		t.Errorf("got total trend cost %v, want %v", totalTrendCost, expectedTotal)
	}
}

func TestCostReporter_EmptyTransactions(t *testing.T) {
	cr := NewCostReporter(nil)
	ctx := context.Background()

	t.Run("empty report", func(t *testing.T) {
		report := cr.GenerateReport(ctx, "daily")
		if report == nil {
			t.Fatal("expected non-nil report")
		}
		if report.TotalCost != 0 {
			t.Errorf("got TotalCost %v, want 0", report.TotalCost)
		}
		if len(report.ByModel) != 0 {
			t.Errorf("got %d models, want 0", len(report.ByModel))
		}
		if len(report.TopExpenses) != 0 {
			t.Errorf("got %d top expenses, want 0", len(report.TopExpenses))
		}
		if len(report.Trends) != 0 {
			t.Errorf("got %d trends, want 0", len(report.Trends))
		}
	})

	t.Run("empty top expenses", func(t *testing.T) {
		expenses := cr.GetTopExpenses(ctx, 5)
		if expenses != nil {
			t.Errorf("got expenses for empty reporter, want nil")
		}
	})

	t.Run("empty trend", func(t *testing.T) {
		trends := cr.GetTrend(ctx)
		if trends != nil {
			t.Errorf("got trends for empty reporter, want nil")
		}
	})
}

func TestCostReporter_FilterByPeriod(t *testing.T) {
	now := time.Now()
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-48 * time.Hour)}, // outside daily
		{ID: "tx2", BudgetID: "b1", Amount: 0.02, ModelID: "gpt-4", Surface: "search", Timestamp: now.Add(-12 * time.Hour)}, // inside daily
		{ID: "tx3", BudgetID: "b2", Amount: 0.03, ModelID: "claude-sonnet", Surface: "agent", Timestamp: now.Add(-2 * time.Hour)}, // inside daily
	}

	cr := NewCostReporter(txs)
	ctx := context.Background()

	dailyReport := cr.GenerateReport(ctx, "daily")
	if dailyReport.TotalCost != 0.05 {
		t.Errorf("got daily report TotalCost %v, want 0.05", dailyReport.TotalCost)
	}

	allReport := cr.GenerateReport(ctx, "all")
	if allReport.TotalCost != 0.06 {
		t.Errorf("got all report TotalCost %v, want 0.06", allReport.TotalCost)
	}
}

func TestCostReporter_ContextCancellation(t *testing.T) {
	txs := []Transaction{
		{ID: "tx1", BudgetID: "b1", Amount: 0.01, ModelID: "gpt-4", Surface: "search", Timestamp: time.Now()},
	}
	cr := NewCostReporter(txs)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	report := cr.GenerateReport(ctx, "daily")
	if report != nil {
		t.Error("GenerateReport() expected nil on cancelled context")
	}

	expenses := cr.GetTopExpenses(ctx, 5)
	if expenses != nil {
		t.Error("GetTopExpenses() expected nil on cancelled context")
	}

	trends := cr.GetTrend(ctx)
	if trends != nil {
		t.Error("GetTrend() expected nil on cancelled context")
	}
}
