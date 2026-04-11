package autoevolve

import (
	"context"
	"testing"
	"time"
)

func TestAnalyzer_Analyze_FullAnalysis(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(1000)
	ctx := context.Background()
	now := time.Now()

	// Record events for top skills
	for i := 0; i < 20; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			Surface:      "vscode",
			ActionType:   "skill_used",
			ArtifactKind: "skill",
			ArtifactID:   "popular-skill",
			Success:      true,
			Duration:     200 * time.Millisecond,
			TokensUsed:   300,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Record events for a failed skill
	for i := 0; i < 10; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			Surface:      "vscode",
			ActionType:   "skill_used",
			ArtifactKind: "skill",
			ArtifactID:   "failing-skill",
			Success:      i < 5, // 50% failure rate
			Duration:     100 * time.Millisecond,
			TokensUsed:   200,
			ErrorType:    "timeout",
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Record events for missing skill searches
	for i := 0; i < 7; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			Surface:    "cli",
			ActionType: "search",
			ArtifactID: "wanted-skill",
			Success:    false,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Record events for token waste
	for i := 0; i < 5; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			Surface:      "wasteful-surface",
			ActionType:   "skill_used",
			ArtifactKind: "skill",
			ArtifactID:   "some-skill",
			Success:      true,
			TokensUsed:   100,
			TokensWasted: 80, // 80% waste
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	analyzer := NewAnalyzer(acc)

	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if len(report.TopSkills) == 0 {
		t.Error("Expected top skills, got none")
	}

	if len(report.FailedSkills) == 0 {
		t.Error("Expected failed skills, got none")
	}

	if len(report.MissingSkills) == 0 {
		t.Error("Expected missing skills, got none")
	}

	if len(report.TokenWaste) == 0 {
		t.Error("Expected token waste findings, got none")
	}

	if len(report.Recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}
}

func TestAnalyzer_TopSkillsIdentification(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	// Skill A: 10 activations
	for i := 0; i < 10; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			ArtifactKind: "skill",
			ArtifactID:   "skill-a",
			Success:      true,
			Duration:     100 * time.Millisecond,
			TokensUsed:   100,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Skill B: 5 activations
	for i := 0; i < 5; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			ArtifactKind: "skill",
			ArtifactID:   "skill-b",
			Success:      true,
			Duration:     200 * time.Millisecond,
			TokensUsed:   200,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	analyzer := NewAnalyzer(acc)
	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if len(report.TopSkills) < 2 {
		t.Fatalf("Expected at least 2 top skills, got %d", len(report.TopSkills))
	}

	if report.TopSkills[0].ID != "skill-a" {
		t.Errorf("Top skill should be skill-a, got %s", report.TopSkills[0].ID)
	}
	if report.TopSkills[0].Activations != 10 {
		t.Errorf("skill-a activations = %d, want 10", report.TopSkills[0].Activations)
	}
	if report.TopSkills[1].ID != "skill-b" {
		t.Errorf("Second skill should be skill-b, got %s", report.TopSkills[1].ID)
	}
	if report.TopSkills[1].Activations != 5 {
		t.Errorf("skill-b activations = %d, want 5", report.TopSkills[1].Activations)
	}
}

func TestAnalyzer_FailedSkillsDetection(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	// Skill with 50% failure rate
	for i := 0; i < 10; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			ArtifactID: "unreliable-skill",
			Surface:    "vscode",
			Success:    i < 4, // 6 failures out of 10 = 60% failure rate
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	analyzer := NewAnalyzer(acc)
	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if len(report.FailedSkills) == 0 {
		t.Fatal("Expected failed skills, got none")
	}

	found := false
	for _, fs := range report.FailedSkills {
		if fs.ID == "unreliable-skill" {
			found = true
			if fs.SuccessRate > 0.5 {
				t.Errorf("unreliable-skill SuccessRate = %f, expected <= 0.4", fs.SuccessRate)
			}
		}
	}
	if !found {
		t.Error("unreliable-skill not found in failed skills")
	}
}

func TestAnalyzer_MissingSkillsDetection(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	// Frequently searched but not found (>= 5)
	for i := 0; i < 8; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			ActionType: "search",
			ArtifactID: "missing-skill",
			Surface:    "vscode",
			Success:    false,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Infrequently searched (< 5, should not appear)
	for i := 0; i < 3; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			ActionType: "search",
			ArtifactID: "rare-search",
			Surface:    "cli",
			Success:    false,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	analyzer := NewAnalyzer(acc)
	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	found := false
	for _, ms := range report.MissingSkills {
		if ms.Query == "missing-skill" {
			found = true
			if ms.SearchCount != 8 {
				t.Errorf("missing-skill SearchCount = %d, want 8", ms.SearchCount)
			}
		}
		if ms.Query == "rare-search" {
			t.Error("rare-search should not appear (only 3 searches, need >= 5)")
		}
	}
	if !found {
		t.Error("missing-skill not detected in missing skills")
	}
}

func TestAnalyzer_TokenWasteIdentification(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	// High waste surface: 70% waste rate
	for i := 0; i < 10; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			Surface:      "high-waste",
			TokensUsed:   100,
			TokensWasted: 70,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Low waste surface: 20% waste rate
	for i := 0; i < 10; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			Surface:      "low-waste",
			TokensUsed:   100,
			TokensWasted: 20,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	analyzer := NewAnalyzer(acc)
	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	foundHigh := false
	foundLow := false
	for _, tw := range report.TokenWaste {
		if tw.Surface == "high-waste" {
			foundHigh = true
		}
		if tw.Surface == "low-waste" {
			foundLow = true
		}
	}

	if !foundHigh {
		t.Error("high-waste surface should be in token waste findings")
	}
	if foundLow {
		t.Error("low-waste surface should NOT be in token waste findings (20% < 50% threshold)")
	}
}

func TestAnalyzer_RecommendationGeneration(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()
	now := time.Now()

	// Generate conditions for all recommendation types
	// Missing skill -> new_skill
	for i := 0; i < 6; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			ActionType: "search",
			ArtifactID: "needed-skill",
			Surface:    "vscode",
			Success:    false,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Failed skill -> update_skill
	for i := 0; i < 10; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:  now,
			ArtifactID: "broken-skill",
			Surface:    "vscode",
			Success:    false,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Token waste -> optimize_context
	for i := 0; i < 5; i++ {
		if err := acc.Record(ctx, UsageTelemetry{
			Timestamp:    now,
			Surface:      "wasteful",
			TokensUsed:   100,
			TokensWasted: 80,
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	analyzer := NewAnalyzer(acc)
	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	types := make(map[string]int)
	for _, r := range report.Recommendations {
		types[r.Type]++
	}

	if types["new_skill"] == 0 {
		t.Error("Expected new_skill recommendation")
	}
	if types["update_skill"] == 0 {
		t.Error("Expected update_skill recommendation")
	}
	if types["optimize_context"] == 0 {
		t.Error("Expected optimize_context recommendation")
	}

	// Verify recommendation fields
	for _, r := range report.Recommendations {
		if r.Title == "" {
			t.Error("Recommendation missing Title")
		}
		if r.Description == "" {
			t.Error("Recommendation missing Description")
		}
		if r.Impact == "" {
			t.Error("Recommendation missing Impact")
		}
		if r.Effort == "" {
			t.Error("Recommendation missing Effort")
		}
		if r.Confidence < 0 || r.Confidence > 1 {
			t.Errorf("Recommendation Confidence = %f, should be 0.0-1.0", r.Confidence)
		}
	}
}

func TestAnalyzer_EmptyTelemetryHandling(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	analyzer := NewAnalyzer(acc)
	report, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if report == nil {
		t.Fatal("Analyze() returned nil report")
	}

	if len(report.TopSkills) != 0 {
		t.Error("Expected no top skills with empty telemetry")
	}
	if len(report.FailedSkills) != 0 {
		t.Error("Expected no failed skills with empty telemetry")
	}
	if len(report.MissingSkills) != 0 {
		t.Error("Expected no missing skills with empty telemetry")
	}
	if len(report.TokenWaste) != 0 {
		t.Error("Expected no token waste findings with empty telemetry")
	}
}

func TestAnalyzer_ContextCancellation(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	analyzer := NewAnalyzer(acc)

	report, err := analyzer.Analyze(ctx)
	if err == nil {
		t.Error("Analyze() with cancelled context should return error")
	}
	if report != nil {
		t.Error("Analyze() with cancelled context should return nil report")
	}

	reports := analyzer.GetReports(ctx)
	if reports != nil {
		t.Error("GetReports() with cancelled context should return nil")
	}
}

func TestAnalyzer_GetReports(t *testing.T) {
	t.Parallel()
	acc := NewTelemetryAccumulator(100)
	ctx := context.Background()

	analyzer := NewAnalyzer(acc)

	// Run analysis twice
	for i := 0; i < 2; i++ {
		if _, err := analyzer.Analyze(ctx); err != nil {
			t.Fatalf("Analyze() error = %v", err)
		}
	}

	reports := analyzer.GetReports(ctx)
	if len(reports) != 2 {
		t.Fatalf("GetReports() returned %d reports, want 2", len(reports))
	}
}
