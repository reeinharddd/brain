package cost

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BudgetPeriod defines the budget reset period
type BudgetPeriod string

const (
	BudgetDaily   BudgetPeriod = "daily"
	BudgetMonthly BudgetPeriod = "monthly"
)

// BudgetAlert defines a threshold alert
type BudgetAlert struct {
	ThresholdPercent float64   // 0.0-1.0
	Triggered        bool
	TriggeredAt      time.Time
}

// Budget defines spending limits
type Budget struct {
	ID           string
	Period       BudgetPeriod
	LimitUSD     float64
	CurrentSpend float64
	ResetAt      time.Time
	Alerts       []BudgetAlert // Threshold alerts
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Transaction represents a cost transaction
type Transaction struct {
	ID          string
	BudgetID    string
	Amount      float64
	Description string
	Timestamp   time.Time
	Surface     string
	ModelID     string
}

// BudgetManager manages spending budgets
type BudgetManager struct {
	mu      sync.RWMutex
	budgets map[string]*Budget    // budgetID -> budget
	txLog   []Transaction         // append-only transaction log
}

// NewBudgetManager creates a new BudgetManager.
func NewBudgetManager() *BudgetManager {
	return &BudgetManager{
		budgets: make(map[string]*Budget),
		txLog:   make([]Transaction, 0),
	}
}

// CreateBudget creates a new budget.
func (bm *BudgetManager) CreateBudget(ctx context.Context, budget Budget) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("budget manager: create budget: %w", err)
	}
	if budget.ID == "" {
		return fmt.Errorf("budget manager: budget ID is required")
	}
	if budget.LimitUSD <= 0 {
		return fmt.Errorf("budget manager: limit must be positive")
	}
	if budget.Period != BudgetDaily && budget.Period != BudgetMonthly {
		return fmt.Errorf("budget manager: invalid budget period %q", budget.Period)
	}

	now := time.Now()
	budget.CreatedAt = now
	budget.UpdatedAt = now
	budget.ResetAt = bm.computeResetAt(now, budget.Period)
	budget.CurrentSpend = 0

	bm.mu.Lock()
	bm.budgets[budget.ID] = &budget
	bm.mu.Unlock()
	return nil
}

// GetBudget retrieves a budget by ID.
func (bm *BudgetManager) GetBudget(ctx context.Context, id string) (*Budget, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("budget manager: get budget: %w", err)
	}
	bm.mu.RLock()
	b, ok := bm.budgets[id]
	bm.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("budget manager: budget %q not found", id)
	}
	return b, nil
}

// Spend records a spend against a budget.
func (bm *BudgetManager) Spend(ctx context.Context, budgetID string, amount float64, surface, modelID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("budget manager: spend: %w", err)
	}
	if amount < 0 {
		return fmt.Errorf("budget manager: amount must be non-negative")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	b, ok := bm.budgets[budgetID]
	if !ok {
		return fmt.Errorf("budget manager: budget %q not found", budgetID)
	}

	// Check if budget needs reset
	bm.resetIfNeeded(b)

	// Check if spend would exceed limit
	if b.CurrentSpend+amount > b.LimitUSD {
		return fmt.Errorf("budget manager: spend of %.4f would exceed budget limit of %.4f (current: %.4f)", amount, b.LimitUSD, b.CurrentSpend)
	}

	b.CurrentSpend += amount
	b.UpdatedAt = time.Now()

	// Check alerts
	for i := range b.Alerts {
		threshold := b.LimitUSD * b.Alerts[i].ThresholdPercent
		if b.CurrentSpend >= threshold && !b.Alerts[i].Triggered {
			b.Alerts[i].Triggered = true
			b.Alerts[i].TriggeredAt = time.Now()
		}
	}

	// Record transaction
	tx := Transaction{
		ID:       uuid.New().String(),
		BudgetID: budgetID,
		Amount:   amount,
		Timestamp: time.Now(),
		Surface:  surface,
		ModelID:  modelID,
	}
	bm.txLog = append(bm.txLog, tx)

	return nil
}

// CheckBudget checks if a requested amount would fit within the budget.
func (bm *BudgetManager) CheckBudget(ctx context.Context, budgetID string, requestedAmount float64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("budget manager: check budget: %w", err)
	}
	if requestedAmount < 0 {
		return fmt.Errorf("budget manager: requested amount must be non-negative")
	}

	bm.mu.RLock()
	defer bm.mu.RUnlock()

	b, ok := bm.budgets[budgetID]
	if !ok {
		return fmt.Errorf("budget manager: budget %q not found", budgetID)
	}

	if b.CurrentSpend+requestedAmount > b.LimitUSD {
		remaining := b.LimitUSD - b.CurrentSpend
		return fmt.Errorf("budget manager: insufficient budget; requested %.4f, remaining %.4f", requestedAmount, remaining)
	}
	return nil
}

// GetTransactions returns all transactions for a budget.
func (bm *BudgetManager) GetTransactions(ctx context.Context, budgetID string) []Transaction {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]Transaction, 0)
	for _, tx := range bm.txLog {
		if tx.BudgetID == budgetID {
			result = append(result, tx)
		}
	}
	return result
}

// ListBudgets returns all budgets.
func (bm *BudgetManager) ListBudgets(ctx context.Context) []Budget {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]Budget, 0, len(bm.budgets))
	for _, b := range bm.budgets {
		result = append(result, *b)
	}
	return result
}

// DeleteBudget deletes a budget by ID.
func (bm *BudgetManager) DeleteBudget(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("budget manager: delete budget: %w", err)
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, ok := bm.budgets[id]; !ok {
		return fmt.Errorf("budget manager: budget %q not found", id)
	}

	delete(bm.budgets, id)

	// Note: transactions are kept for audit purposes
	return nil
}

// computeResetAt calculates the next reset time based on budget period.
func (bm *BudgetManager) computeResetAt(now time.Time, period BudgetPeriod) time.Time {
	switch period {
	case BudgetDaily:
		return now.Add(24 * time.Hour)
	case BudgetMonthly:
		return now.AddDate(0, 1, 0)
	default:
		return now.Add(24 * time.Hour)
	}
}

// resetIfNeeded checks if the budget has passed its reset time and resets it.
func (bm *BudgetManager) resetIfNeeded(b *Budget) {
	if time.Now().After(b.ResetAt) {
		b.CurrentSpend = 0
		b.ResetAt = bm.computeResetAt(time.Now(), b.Period)
		for i := range b.Alerts {
			b.Alerts[i].Triggered = false
			b.Alerts[i].TriggeredAt = time.Time{}
		}
	}
}
