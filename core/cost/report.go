package cost

import (
	"context"
	"sort"
	"time"
)

// ModelCost breaks down cost by model
type ModelCost struct {
	ModelID      string
	TotalCost    float64
	TotalTokens  int
	CallCount    int
	AvgCostPerCall float64
}

// SurfaceCost breaks down cost by surface
type SurfaceCost struct {
	Surface   string
	TotalCost float64
	CallCount int
}

// Expense represents a significant expense
type Expense struct {
	Description string
	Cost        float64
	ModelID     string
	Surface     string
	Timestamp   time.Time
}

// TrendPoint represents a cost trend
type TrendPoint struct {
	Timestamp time.Time
	Cost      float64
	Tokens    int
}

// CostReport generates cost analytics
type CostReport struct {
	Period      string // "daily", "weekly", "monthly"
	TotalCost   float64
	TotalTokens int
	ByModel     map[string]ModelCost
	BySurface   map[string]SurfaceCost
	TopExpenses []Expense
	Trends      []TrendPoint
}

// CostReporter generates reports
type CostReporter struct {
	transactions []Transaction
}

// NewCostReporter creates a new CostReporter.
func NewCostReporter(transactions []Transaction) *CostReporter {
	return &CostReporter{
		transactions: transactions,
	}
}

// GenerateReport generates cost analytics for a given period.
func (cr *CostReporter) GenerateReport(ctx context.Context, period string) *CostReport {
	if err := ctx.Err(); err != nil {
		return nil
	}

	report := &CostReport{
		Period:    period,
		ByModel:   make(map[string]ModelCost),
		BySurface: make(map[string]SurfaceCost),
	}

	// Filter transactions by period
	txs := cr.filterByPeriod(period)

	for _, tx := range txs {
		report.TotalCost += tx.Amount
		report.TotalTokens += 0 // Tokens not stored in Transaction; cost is direct

		// By model
		mc := report.ByModel[tx.ModelID]
		mc.ModelID = tx.ModelID
		mc.TotalCost += tx.Amount
		mc.CallCount++
		mc.TotalTokens += 0
		mc.AvgCostPerCall = mc.TotalCost / float64(mc.CallCount)
		report.ByModel[tx.ModelID] = mc

		// By surface
		if tx.Surface != "" {
			sc := report.BySurface[tx.Surface]
			sc.Surface = tx.Surface
			sc.TotalCost += tx.Amount
			sc.CallCount++
			report.BySurface[tx.Surface] = sc
		}
	}

	// Get top expenses
	report.TopExpenses = cr.getTopExpenses(txs, 10)

	// Generate trends
	report.Trends = cr.generateTrends(txs, period)

	return report
}

// GetTopExpenses returns the top N expenses.
func (cr *CostReporter) GetTopExpenses(ctx context.Context, n int) []Expense {
	if err := ctx.Err(); err != nil {
		return nil
	}
	return cr.getTopExpenses(cr.transactions, n)
}

// GetTrend returns the cost trend.
func (cr *CostReporter) GetTrend(ctx context.Context) []TrendPoint {
	if err := ctx.Err(); err != nil {
		return nil
	}
	return cr.generateTrends(cr.transactions, "all")
}

// filterByPeriod filters transactions based on the period string.
func (cr *CostReporter) filterByPeriod(period string) []Transaction {
	if len(cr.transactions) == 0 {
		return nil
	}

	now := time.Now()
	var cutoff time.Time

	switch period {
	case "daily":
		cutoff = now.Add(-24 * time.Hour)
	case "weekly":
		cutoff = now.Add(-7 * 24 * time.Hour)
	case "monthly":
		cutoff = now.AddDate(0, -1, 0)
	default:
		// Return all transactions
		result := make([]Transaction, len(cr.transactions))
		copy(result, cr.transactions)
		return result
	}

	var result []Transaction
	for _, tx := range cr.transactions {
		if tx.Timestamp.After(cutoff) {
			result = append(result, tx)
		}
	}
	return result
}

// getTopExpenses returns the top N expenses from transactions.
func (cr *CostReporter) getTopExpenses(txs []Transaction, n int) []Expense {
	if len(txs) == 0 {
		return nil
	}

	expenses := make([]Expense, len(txs))
	for i, tx := range txs {
		expenses[i] = Expense{
			Description: tx.Description,
			Cost:        tx.Amount,
			ModelID:     tx.ModelID,
			Surface:     tx.Surface,
			Timestamp:   tx.Timestamp,
		}
	}

	// Sort by cost descending
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].Cost > expenses[j].Cost
	})

	if n > len(expenses) {
		n = len(expenses)
	}
	return expenses[:n]
}

// generateTrends generates trend points from transactions.
func (cr *CostReporter) generateTrends(txs []Transaction, period string) []TrendPoint {
	if len(txs) == 0 {
		return nil
	}

	// Sort transactions by timestamp
	sorted := make([]Transaction, len(txs))
	copy(sorted, txs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Group by hour for the trend
	type hourBucket struct {
		hour time.Time
		cost float64
	}

	buckets := make(map[time.Time]*hourBucket)
	for _, tx := range sorted {
		hour := tx.Timestamp.Truncate(time.Hour)
		b, ok := buckets[hour]
		if !ok {
			b = &hourBucket{hour: hour}
			buckets[hour] = b
		}
		b.cost += tx.Amount
	}

	// Convert to sorted trend points
	trends := make([]TrendPoint, 0, len(buckets))
	for _, b := range buckets {
		trends = append(trends, TrendPoint{
			Timestamp: b.hour,
			Cost:      b.cost,
		})
	}

	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Timestamp.Before(trends[j].Timestamp)
	})

	return trends
}
