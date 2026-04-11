package autoevolve

import (
	"context"
	"fmt"
	"sync"
)

// AutoEvolveEngine orchestrates the self-improvement loop
type AutoEvolveEngine struct {
	mu        sync.RWMutex
	telemetry *TelemetryAccumulator
	analyzer  *Analyzer
	enabled   bool
	approved  map[string]bool // recommendation ID -> approval
	history   []string        // applied recommendation IDs
}

func NewAutoEvolveEngine(telemetry *TelemetryAccumulator) *AutoEvolveEngine {
	return &AutoEvolveEngine{
		telemetry: telemetry,
		analyzer:  NewAnalyzer(telemetry),
		enabled:   false,
		approved:  make(map[string]bool),
		history:   make([]string, 0),
	}
}

func (e *AutoEvolveEngine) Enable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = true
}

func (e *AutoEvolveEngine) Disable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = false
}

func (e *AutoEvolveEngine) IsEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enabled
}

func (e *AutoEvolveEngine) RunAnalysis(ctx context.Context) (*AnalysisReport, error) {
	e.mu.RLock()
	enabled := e.enabled
	e.mu.RUnlock()

	if !enabled {
		return nil, fmt.Errorf("autoevolve engine is disabled")
	}

	report, err := e.analyzer.Analyze(ctx)
	if err != nil {
		return nil, fmt.Errorf("run analysis: %w", err)
	}
	return report, nil
}

func (e *AutoEvolveEngine) GetPendingRecommendations(ctx context.Context) []Recommendation {
	reports := e.analyzer.GetReports(ctx)

	var pending []Recommendation
	for _, report := range reports {
		for _, rec := range report.Recommendations {
			if rec.Approved == nil {
				pending = append(pending, rec)
			}
		}
	}
	return pending
}

func (e *AutoEvolveEngine) ApproveRecommendation(ctx context.Context, recommendationID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("approve recommendation: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.approved[recommendationID] = true
	return nil
}

func (e *AutoEvolveEngine) RejectRecommendation(ctx context.Context, recommendationID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("reject recommendation: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.approved[recommendationID] = false
	return nil
}

func (e *AutoEvolveEngine) ApplyApproved(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("apply approved: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	var applied []string
	for recID, isApproved := range e.approved {
		if isApproved {
			// Simulate applying the recommendation
			e.history = append(e.history, recID)
			applied = append(applied, recID)
		}
	}

	return applied, nil
}

func (e *AutoEvolveEngine) GetHistory(ctx context.Context) []string {
	if err := ctx.Err(); err != nil {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]string, len(e.history))
	copy(result, e.history)
	return result
}

func (e *AutoEvolveEngine) GetStatus(ctx context.Context) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":           e.enabled,
		"telemetry_count":   e.telemetry.Count(),
		"approved_count":    len(e.approved),
		"history_count":     len(e.history),
		"pending_approvals": 0,
	}

	// Count pending approvals
	reports := e.analyzer.GetReports(ctx)
	for _, report := range reports {
		for _, rec := range report.Recommendations {
			if rec.Approved == nil {
				status["pending_approvals"] = status["pending_approvals"].(int) + 1
			}
		}
	}

	return status
}
