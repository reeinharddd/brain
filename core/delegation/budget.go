package delegation

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Budget error types
var (
	ErrTokenLimitExceeded    = errors.New("delegation: token limit exceeded")
	ErrCostLimitExceeded     = errors.New("delegation: cost limit exceeded")
	ErrDurationLimitExceeded = errors.New("delegation: duration limit exceeded")
	ErrRetryLimitExceeded    = errors.New("delegation: retry limit exceeded")
)

// BudgetTracker tracks delegation execution budget
type BudgetTracker struct {
	mu           sync.RWMutex
	budget       DelegationBudget
	tokensUsed   int
	costUsed     float64
	durationUsed time.Duration
	retriesUsed  int
	startTime    time.Time
}

// NewBudgetTracker creates a new budget tracker with the given budget limits
func NewBudgetTracker(budget DelegationBudget) *BudgetTracker {
	return &BudgetTracker{
		budget:    budget,
		startTime: time.Now(),
	}
}

// RecordTokenUsage records token usage and returns an error if budget is exceeded
func (bt *BudgetTracker) RecordTokenUsage(tokens int) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if tokens < 0 {
		return fmt.Errorf("delegation: cannot record negative token usage: %d", tokens)
	}

	bt.tokensUsed += tokens

	if bt.budget.MaxTokens > 0 && bt.tokensUsed > bt.budget.MaxTokens {
		return fmt.Errorf("token usage %d exceeds limit %d: %w", bt.tokensUsed, bt.budget.MaxTokens, ErrTokenLimitExceeded)
	}

	return nil
}

// RecordCost records cost usage and returns an error if budget is exceeded
func (bt *BudgetTracker) RecordCost(cost float64) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if cost < 0 {
		return fmt.Errorf("delegation: cannot record negative cost: %f", cost)
	}

	bt.costUsed += cost

	if bt.budget.MaxCostUSD > 0 && bt.costUsed > bt.budget.MaxCostUSD {
		return fmt.Errorf("cost %f exceeds limit %f: %w", bt.costUsed, bt.budget.MaxCostUSD, ErrCostLimitExceeded)
	}

	return nil
}

// RecordRetry records a retry attempt and returns an error if budget is exceeded
func (bt *BudgetTracker) RecordRetry() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.retriesUsed++

	if bt.budget.MaxRetries > 0 && bt.retriesUsed > bt.budget.MaxRetries {
		return fmt.Errorf("retries %d exceeds limit %d: %w", bt.retriesUsed, bt.budget.MaxRetries, ErrRetryLimitExceeded)
	}

	return nil
}

// CheckBudget checks if the current budget usage exceeds any limits
func (bt *BudgetTracker) CheckBudget() error {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.budget.MaxTokens > 0 && bt.tokensUsed > bt.budget.MaxTokens {
		return fmt.Errorf("token usage %d exceeds limit %d: %w", bt.tokensUsed, bt.budget.MaxTokens, ErrTokenLimitExceeded)
	}

	if bt.budget.MaxCostUSD > 0 && bt.costUsed > bt.budget.MaxCostUSD {
		return fmt.Errorf("cost %f exceeds limit %f: %w", bt.costUsed, bt.budget.MaxCostUSD, ErrCostLimitExceeded)
	}

	elapsed := time.Since(bt.startTime)
	if bt.budget.MaxDuration > 0 && elapsed > bt.budget.MaxDuration {
		return fmt.Errorf("duration %v exceeds limit %v: %w", elapsed, bt.budget.MaxDuration, ErrDurationLimitExceeded)
	}

	if bt.budget.MaxRetries > 0 && bt.retriesUsed > bt.budget.MaxRetries {
		return fmt.Errorf("retries %d exceeds limit %d: %w", bt.retriesUsed, bt.budget.MaxRetries, ErrRetryLimitExceeded)
	}

	return nil
}

// GetRemaining returns the remaining budget for each category
func (bt *BudgetTracker) GetRemaining() map[string]interface{} {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	elapsed := time.Since(bt.startTime)

	remaining := make(map[string]interface{})

	if bt.budget.MaxTokens > 0 {
		remaining["tokens"] = bt.budget.MaxTokens - bt.tokensUsed
	} else {
		remaining["tokens"] = -1 // unlimited
	}

	if bt.budget.MaxCostUSD > 0 {
		remaining["cost"] = bt.budget.MaxCostUSD - bt.costUsed
	} else {
		remaining["cost"] = -1.0 // unlimited
	}

	if bt.budget.MaxDuration > 0 {
		remaining["duration"] = bt.budget.MaxDuration - elapsed
	} else {
		remaining["duration"] = time.Duration(-1) // unlimited
	}

	if bt.budget.MaxRetries > 0 {
		remaining["retries"] = bt.budget.MaxRetries - bt.retriesUsed
	} else {
		remaining["retries"] = -1 // unlimited
	}

	remaining["tokens_used"] = bt.tokensUsed
	remaining["cost_used"] = bt.costUsed
	remaining["duration_used"] = elapsed
	remaining["retries_used"] = bt.retriesUsed

	return remaining
}

// IsExceeded returns true if any budget limit has been exceeded
func (bt *BudgetTracker) IsExceeded() bool {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.budget.MaxTokens > 0 && bt.tokensUsed > bt.budget.MaxTokens {
		return true
	}

	if bt.budget.MaxCostUSD > 0 && bt.costUsed > bt.budget.MaxCostUSD {
		return true
	}

	elapsed := time.Since(bt.startTime)
	if bt.budget.MaxDuration > 0 && elapsed > bt.budget.MaxDuration {
		return true
	}

	if bt.budget.MaxRetries > 0 && bt.retriesUsed > bt.budget.MaxRetries {
		return true
	}

	return false
}

// GetTokensUsed returns the current token usage
func (bt *BudgetTracker) GetTokensUsed() int {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.tokensUsed
}

// GetCostUsed returns the current cost usage
func (bt *BudgetTracker) GetCostUsed() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.costUsed
}

// GetRetriesUsed returns the current retry count
func (bt *BudgetTracker) GetRetriesUsed() int {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.retriesUsed
}
