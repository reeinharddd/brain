package autoevolve

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTelemetryAccumulator_Record(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	event := UsageTelemetry{
		Timestamp:    time.Now(),
		Surface:      "vscode",
		ActionType:   "skill_used",
		ArtifactKind: "skill",
		ArtifactID:   "test-skill",
		Success:      true,
		Duration:     100 * time.Millisecond,
		TokensUsed:   500,
		TokensWasted: 10,
	}

	if err := acc.Record(ctx, event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	if got := acc.Count(); got != 1 {
		t.Fatalf("Count() = %d, want 1", got)
	}
}

func TestTelemetryAccumulator_GetEvents(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	now := time.Now()
	events := []UsageTelemetry{
		{Timestamp: now, Surface: "vscode", ActionType: "skill_used", ArtifactID: "skill-1", Success: true},
		{Timestamp: now.Add(1 * time.Second), Surface: "cli", ActionType: "mcp_called", ArtifactID: "mcp-1", Success: false},
	}

	for _, e := range events {
		if err := acc.Record(ctx, e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	got := acc.GetEvents(ctx)
	if len(got) != 2 {
		t.Fatalf("GetEvents() returned %d events, want 2", len(got))
	}

	// Verify order
	if got[0].ArtifactID != "skill-1" {
		t.Errorf("GetEvents()[0].ArtifactID = %q, want %q", got[0].ArtifactID, "skill-1")
	}
	if got[1].ArtifactID != "mcp-1" {
		t.Errorf("GetEvents()[1].ArtifactID = %q, want %q", got[1].ArtifactID, "mcp-1")
	}
}

func TestTelemetryAccumulator_FailureTracking(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		event   UsageTelemetry
		wantID  string
		wantFails int
	}{
		{
			name:   "successful call",
			event:  UsageTelemetry{Timestamp: now, ArtifactID: "skill-a", Success: true},
			wantID: "skill-a",
		},
		{
			name:    "first failure",
			event:   UsageTelemetry{Timestamp: now.Add(1 * time.Second), ArtifactID: "skill-a", Success: false, ErrorType: "timeout"},
			wantID:  "skill-a",
			wantFails: 1,
		},
		{
			name:    "second failure",
			event:   UsageTelemetry{Timestamp: now.Add(2 * time.Second), ArtifactID: "skill-a", Success: false, ErrorType: "timeout"},
			wantID:  "skill-a",
			wantFails: 2,
		},
		{
			name:   "success resets nothing but increments attempts",
			event:  UsageTelemetry{Timestamp: now.Add(3 * time.Second), ArtifactID: "skill-a", Success: true},
			wantID: "skill-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := acc.Record(ctx, tt.event); err != nil {
				t.Fatalf("Record() error = %v", err)
			}
		})
	}

	stats := acc.GetFailureStats(ctx)
	sa, ok := stats["skill-a"]
	if !ok {
		t.Fatal("GetFailureStats() missing skill-a")
	}

	if sa.TotalAttempts != 4 {
		t.Errorf("TotalAttempts = %d, want 4", sa.TotalAttempts)
	}
	if sa.Failures != 2 {
		t.Errorf("Failures = %d, want 2", sa.Failures)
	}
	if sa.ErrorTypes["timeout"] != 2 {
		t.Errorf("ErrorTypes[timeout] = %d, want 2", sa.ErrorTypes["timeout"])
	}
	wantRate := 2.0 / 4.0
	if sa.FailureRate != wantRate {
		t.Errorf("FailureRate = %f, want %f", sa.FailureRate, wantRate)
	}
}

func TestTelemetryAccumulator_FailureTracking_ErrorTypes(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	events := []UsageTelemetry{
		{Timestamp: now, ArtifactID: "skill-b", Success: false, ErrorType: "timeout"},
		{Timestamp: now.Add(1 * time.Second), ArtifactID: "skill-b", Success: false, ErrorType: "oom"},
		{Timestamp: now.Add(2 * time.Second), ArtifactID: "skill-b", Success: false, ErrorType: "timeout"},
	}

	for _, e := range events {
		if err := acc.Record(ctx, e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	stats := acc.GetFailureStats(ctx)
	sb := stats["skill-b"]
	if sb.ErrorTypes["timeout"] != 2 {
		t.Errorf("ErrorTypes[timeout] = %d, want 2", sb.ErrorTypes["timeout"])
	}
	if sb.ErrorTypes["oom"] != 1 {
		t.Errorf("ErrorTypes[oom] = %d, want 1", sb.ErrorTypes["oom"])
	}
}

func TestTelemetryAccumulator_GapDetection(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	events := []UsageTelemetry{
		{Timestamp: now, ActionType: "search", ArtifactID: "missing-skill", Surface: "vscode", Success: false},
		{Timestamp: now.Add(1 * time.Second), ActionType: "search", ArtifactID: "missing-skill", Surface: "cli", Success: false},
		{Timestamp: now.Add(2 * time.Second), ActionType: "search", ArtifactID: "missing-skill", Surface: "vscode", Success: false},
	}

	for _, e := range events {
		if err := acc.Record(ctx, e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	gaps := acc.GetGapStats(ctx)
	ms, ok := gaps["missing-skill"]
	if !ok {
		t.Fatal("GetGapStats() missing missing-skill")
	}

	if ms.SearchCount != 3 {
		t.Errorf("SearchCount = %d, want 3", ms.SearchCount)
	}
	if ms.TopSurfaces["vscode"] != 2 {
		t.Errorf("TopSurfaces[vscode] = %d, want 2", ms.TopSurfaces["vscode"])
	}
	if ms.TopSurfaces["cli"] != 1 {
		t.Errorf("TopSurfaces[cli] = %d, want 1", ms.TopSurfaces["cli"])
	}
	if !ms.FirstSeen.Equal(now) {
		t.Errorf("FirstSeen = %v, want %v", ms.FirstSeen, now)
	}
}

func TestTelemetryAccumulator_WasteTracking(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	events := []UsageTelemetry{
		{Timestamp: now, Surface: "vscode", TokensUsed: 1000, TokensWasted: 600},
		{Timestamp: now.Add(1 * time.Second), Surface: "vscode", TokensUsed: 500, TokensWasted: 300},
		{Timestamp: now.Add(2 * time.Second), Surface: "cli", TokensUsed: 200, TokensWasted: 50},
	}

	for _, e := range events {
		if err := acc.Record(ctx, e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	waste := acc.GetWasteStats(ctx)

	vs := waste["vscode"]
	if vs.TotalWasted != 900 {
		t.Errorf("vscode TotalWasted = %d, want 900", vs.TotalWasted)
	}
	if vs.TotalUsed != 1500 {
		t.Errorf("vscode TotalUsed = %d, want 1500", vs.TotalUsed)
	}
	if vs.Sessions != 2 {
		t.Errorf("vscode Sessions = %d, want 2", vs.Sessions)
	}
	wantRate := 900.0 / 1500.0
	if vs.WasteRate != wantRate {
		t.Errorf("vscode WasteRate = %f, want %f", vs.WasteRate, wantRate)
	}

	cs := waste["cli"]
	if cs.TotalWasted != 50 {
		t.Errorf("cli TotalWasted = %d, want 50", cs.TotalWasted)
	}
}

func TestTelemetryAccumulator_MaxEventsLimit(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(3)
	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 5; i++ {
		e := UsageTelemetry{Timestamp: now.Add(time.Duration(i) * time.Second), ArtifactID: "skill"}
		if err := acc.Record(ctx, e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	if got := acc.Count(); got != 3 {
		t.Fatalf("Count() = %d, want 3", got)
	}

	events := acc.GetEvents(ctx)
	// Oldest two should be dropped, so first event is i=2
	if events[0].Timestamp != now.Add(2*time.Second) {
		t.Errorf("Expected first event at t+2s, got %v", events[0].Timestamp)
	}
}

func TestTelemetryAccumulator_ConcurrentRecording(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(1000)
	ctx := context.Background()
	now := time.Now()

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			e := UsageTelemetry{
				Timestamp:  now,
				ArtifactID: "concurrent-skill",
				Surface:    "vscode",
				Success:    idx%2 == 0,
			}
			if err := acc.Record(ctx, e); err != nil {
				t.Errorf("Record() error = %v", err)
			}
		}(i)
	}

	wg.Wait()

	if got := acc.Count(); got != n {
		t.Fatalf("Count() = %d, want %d", got, n)
	}
}

func TestTelemetryAccumulator_Clear(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	events := []UsageTelemetry{
		{Timestamp: time.Now(), ArtifactID: "skill-a", Success: true},
		{Timestamp: time.Now(), ActionType: "search", ArtifactID: "missing", Surface: "vscode", Success: false},
		{Timestamp: time.Now(), Surface: "vscode", TokensUsed: 100, TokensWasted: 50},
	}

	for _, e := range events {
		if err := acc.Record(ctx, e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	if acc.Count() != 3 {
		t.Fatalf("Count() = %d, want 3 before Clear", acc.Count())
	}

	acc.Clear()

	if acc.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after Clear", acc.Count())
	}
	if len(acc.GetFailureStats(ctx)) != 0 {
		t.Error("GetFailureStats() should be empty after Clear")
	}
	if len(acc.GetGapStats(ctx)) != 0 {
		t.Error("GetGapStats() should be empty after Clear")
	}
	if len(acc.GetWasteStats(ctx)) != 0 {
		t.Error("GetWasteStats() should be empty after Clear")
	}
}

func TestTelemetryAccumulator_ContextCancellation(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := acc.Record(ctx, UsageTelemetry{}); err == nil {
		t.Error("Record() with cancelled context should return error")
	}

	events := acc.GetEvents(ctx)
	if events != nil {
		t.Error("GetEvents() with cancelled context should return nil")
	}

	failure := acc.GetFailureStats(ctx)
	if failure != nil {
		t.Error("GetFailureStats() with cancelled context should return nil")
	}

	gaps := acc.GetGapStats(ctx)
	if gaps != nil {
		t.Error("GetGapStats() with cancelled context should return nil")
	}

	waste := acc.GetWasteStats(ctx)
	if waste != nil {
		t.Error("GetWasteStats() with cancelled context should return nil")
	}
}
