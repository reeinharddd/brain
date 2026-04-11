package autoevolve

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAutoEvolveEngine_EnableDisable(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	engine := NewAutoEvolveEngine(acc)

	if engine.IsEnabled() {
		t.Error("New engine should be disabled by default")
	}

	engine.Enable()
	if !engine.IsEnabled() {
		t.Error("Engine should be enabled after Enable()")
	}

	engine.Disable()
	if engine.IsEnabled() {
		t.Error("Engine should be disabled after Disable()")
	}
}

func TestAutoEvolveEngine_RunAnalysis(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	// Add some telemetry
	if err := acc.Record(ctx, UsageTelemetry{
		Timestamp:    time.Now(),
		ArtifactKind: "skill",
		ArtifactID:   "test-skill",
		Success:      true,
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	engine := NewAutoEvolveEngine(acc)
	engine.Enable()

	report, err := engine.RunAnalysis(ctx)
	if err != nil {
		t.Fatalf("RunAnalysis() error = %v", err)
	}

	if report == nil {
		t.Fatal("RunAnalysis() returned nil report")
	}
}

func TestAutoEvolveEngine_DisabledBlocksAnalysis(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)
	// Don't enable

	report, err := engine.RunAnalysis(ctx)
	if err == nil {
		t.Error("RunAnalysis() on disabled engine should return error")
	}
	if report != nil {
		t.Error("RunAnalysis() on disabled engine should return nil report")
	}
}

func TestAutoEvolveEngine_GetPendingRecommendations(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	// Create conditions that generate recommendations
	for i := 0; i < 6; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			ActionType: "search",
			ArtifactID: "needed",
			Surface:    "vscode",
			Success:    false,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	engine := NewAutoEvolveEngine(acc)
	engine.Enable()

	_, err := engine.RunAnalysis(ctx)
	if err != nil {
		t.Fatalf("RunAnalysis() error = %v", err)
	}

	pending := engine.GetPendingRecommendations(ctx)
	if len(pending) == 0 {
		t.Error("Expected pending recommendations, got none")
	}

	// All should be pending (Approved == nil)
	for _, rec := range pending {
		if rec.Approved != nil {
			t.Error("Pending recommendation should have Approved == nil")
		}
	}
}

func TestAutoEvolveEngine_ApproveReject(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)

	if err := engine.ApproveRecommendation(ctx, "rec-1"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}

	if err := engine.RejectRecommendation(ctx, "rec-2"); err != nil {
		t.Fatalf("RejectRecommendation() error = %v", err)
	}
}

func TestAutoEvolveEngine_ApplyApproved(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)

	// Approve some, reject others
	if err := engine.ApproveRecommendation(ctx, "rec-1"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}
	if err := engine.ApproveRecommendation(ctx, "rec-2"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}
	if err := engine.RejectRecommendation(ctx, "rec-3"); err != nil {
		t.Fatalf("RejectRecommendation() error = %v", err)
	}

	applied, err := engine.ApplyApproved(ctx)
	if err != nil {
		t.Fatalf("ApplyApproved() error = %v", err)
	}

	if len(applied) != 2 {
		t.Fatalf("ApplyApproved() returned %d items, want 2", len(applied))
	}

	// Should only contain approved ones
	found1, found2 := false, false
	for _, id := range applied {
		if id == "rec-1" {
			found1 = true
		}
		if id == "rec-2" {
			found2 = true
		}
		if id == "rec-3" {
			t.Error("rec-3 (rejected) should not be in applied list")
		}
	}
	if !found1 || !found2 {
		t.Errorf("Applied list missing rec-1 or rec-2, got %v", applied)
	}
}

func TestAutoEvolveEngine_GetHistory(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)

	// Approve and apply
	if err := engine.ApproveRecommendation(ctx, "rec-a"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}
	if err := engine.ApproveRecommendation(ctx, "rec-b"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}

	applied, err := engine.ApplyApproved(ctx)
	if err != nil {
		t.Fatalf("ApplyApproved() error = %v", err)
	}

	history := engine.GetHistory(ctx)
	if len(history) != len(applied) {
		t.Errorf("GetHistory() returned %d items, want %d", len(history), len(applied))
	}

	// Verify history contains the applied items
	for _, id := range applied {
		found := false
		for _, h := range history {
			if h == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("History missing %q", id)
		}
	}
}

func TestAutoEvolveEngine_GetStatus(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)

	// Add telemetry
	if err := acc.Record(ctx, UsageTelemetry{
		Timestamp: time.Now(),
		Surface:   "vscode",
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	engine.Enable()

	// Run analysis to generate a report
	if _, err := engine.RunAnalysis(ctx); err != nil {
		t.Fatalf("RunAnalysis() error = %v", err)
	}

	// Approve a recommendation
	if err := engine.ApproveRecommendation(ctx, "test-rec"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}

	status := engine.GetStatus(ctx)

	if enabled, ok := status["enabled"].(bool); !ok || !enabled {
		t.Error("Status should show enabled = true")
	}
	if count, ok := status["telemetry_count"].(int); !ok || count != 1 {
		t.Errorf("Status telemetry_count = %v, want 1", status["telemetry_count"])
	}
	if count, ok := status["approved_count"].(int); !ok || count != 1 {
		t.Errorf("Status approved_count = %v, want 1", status["approved_count"])
	}
}

func TestAutoEvolveEngine_ConcurrentApproval(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)

	var wg sync.WaitGroup
	n := 50
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			recID := "concurrent-rec"
			if idx%2 == 0 {
				if err := engine.ApproveRecommendation(ctx, recID); err != nil {
					t.Errorf("ApproveRecommendation() error = %v", err)
				}
			} else {
				if err := engine.RejectRecommendation(ctx, recID); err != nil {
					t.Errorf("RejectRecommendation() error = %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Should not panic or deadlock
	status := engine.GetStatus(ctx)
	if _, ok := status["enabled"]; !ok {
		t.Error("GetStatus() missing 'enabled' key after concurrent operations")
	}
}

func TestAutoEvolveEngine_ContextCancellation(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	engine := NewAutoEvolveEngine(acc)
	engine.Enable()

	_, err := engine.RunAnalysis(ctx)
	if err == nil {
		t.Error("RunAnalysis() with cancelled context should return error")
	}

	if err := engine.ApproveRecommendation(ctx, "rec-1"); err == nil {
		t.Error("ApproveRecommendation() with cancelled context should return error")
	}

	if err := engine.RejectRecommendation(ctx, "rec-1"); err == nil {
		t.Error("RejectRecommendation() with cancelled context should return error")
	}

	_, err = engine.ApplyApproved(ctx)
	if err == nil {
		t.Error("ApplyApproved() with cancelled context should return error")
	}

	history := engine.GetHistory(ctx)
	if history != nil {
		t.Error("GetHistory() with cancelled context should return nil")
	}
}

func TestAutoEvolveEngine_ApplyApproved_OnlyAppliesApproved(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	engine := NewAutoEvolveEngine(acc)

	// Mix of approved, rejected, and pending
	if err := engine.ApproveRecommendation(ctx, "approved-1"); err != nil {
		t.Fatalf("ApproveRecommendation() error = %v", err)
	}
	if err := engine.RejectRecommendation(ctx, "rejected-1"); err != nil {
		t.Fatalf("RejectRecommendation() error = %v", err)
	}

	applied, err := engine.ApplyApproved(ctx)
	if err != nil {
		t.Fatalf("ApplyApproved() error = %v", err)
	}

	if len(applied) != 1 {
		t.Fatalf("ApplyApproved() returned %d items, want 1", len(applied))
	}
	if applied[0] != "approved-1" {
		t.Errorf("ApplyApproved()[0] = %q, want %q", applied[0], "approved-1")
	}

	// Apply again - should apply the same approved one again
	_, err = engine.ApplyApproved(ctx)
	if err != nil {
		t.Fatalf("ApplyApproved() second call error = %v", err)
	}

	history := engine.GetHistory(ctx)
	// History should have 2 entries total
	if len(history) != 2 {
		t.Errorf("GetHistory() returned %d items, want 2", len(history))
	}
}
