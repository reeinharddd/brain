package curator

import (
	"context"
	"testing"
	"time"
)

func newTestConfig() CuratorConfig {
	return CuratorConfig{
		Enabled:            true,
		DryRun:             false,
		DedupThreshold:     0.7,
		MaxTokensPerLayer:  100,
		MinPromotionAccess: 5,
		PromotionRecency:   time.Hour,
		StaleThreshold:     time.Hour,
		MinCleanupAccess:   3,
	}
}

func TestCuratorService(t *testing.T) {
	now := time.Now()

	t.Run("full curator run with all analyzers", func(t *testing.T) {
		cfg := newTestConfig()
		curator := NewCuratorService(cfg)

		contentEntries := map[string]string{
			"entry1": "short content",
			"entry2": "duplicate content here",
			"entry3": "duplicate content here",
		}

		accessRecords := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  10,
				LastAccessed: now.Add(-10 * time.Minute),
				Content:      "short content",
			},
			"entry2": {
				ID:           "entry2",
				Source:       "task-local",
				AccessCount:  8,
				LastAccessed: now.Add(-5 * time.Minute),
				Content:      "duplicate content here",
			},
			"entry3": {
				ID:           "entry3",
				Source:       "workspace",
				AccessCount:  6,
				LastAccessed: now.Add(-2 * time.Minute),
				Content:      "duplicate content here",
			},
		}

		report, err := curator.Run(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.DryRun != false {
			t.Error("expected dry run to be false")
		}
		if report.MemoryState.TotalEntries != 3 {
			t.Errorf("expected 3 total entries, got %d", report.MemoryState.TotalEntries)
		}
		if len(report.Duplicates) < 1 {
			t.Error("expected at least 1 duplicate finding")
		}
	})

	t.Run("dry-run mode", func(t *testing.T) {
		cfg := newTestConfig()
		cfg.DryRun = true
		curator := NewCuratorService(cfg)

		contentEntries := map[string]string{
			"entry1": "test content",
		}

		accessRecords := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  10,
				LastAccessed: now.Add(-5 * time.Minute),
				Content:      "test content",
			},
		}

		report, err := curator.Run(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.DryRun != true {
			t.Error("expected dry run to be true")
		}
	})

	t.Run("disabled service error", func(t *testing.T) {
		cfg := newTestConfig()
		cfg.Enabled = false
		curator := NewCuratorService(cfg)

		_, err := curator.Run(context.Background(), nil, nil)
		if err == nil {
			t.Error("expected error for disabled service")
		}
	})

	t.Run("memory state assessment", func(t *testing.T) {
		cfg := newTestConfig()
		curator := NewCuratorService(cfg)

		contentEntries := map[string]string{
			"entry1": "content one",
			"entry2": "content two",
		}

		accessRecords := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  1,
				LastAccessed: now.Add(-30 * time.Minute),
				Content:      "content one",
			},
			"entry2": {
				ID:           "entry2",
				Source:       "task-local",
				AccessCount:  2,
				LastAccessed: now.Add(-10 * time.Minute),
				Content:      "content two",
			},
		}

		report, err := curator.Run(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.MemoryState.TotalEntries != 2 {
			t.Errorf("expected 2 total entries, got %d", report.MemoryState.TotalEntries)
		}
		if report.MemoryState.HealthScore < 0 || report.MemoryState.HealthScore > 1 {
			t.Errorf("health score out of range: %f", report.MemoryState.HealthScore)
		}
	})

	t.Run("token savings estimation", func(t *testing.T) {
		cfg := newTestConfig()
		cfg.MaxTokensPerLayer = 5
		curator := NewCuratorService(cfg)

		// Create content that exceeds limit
		contentEntries := map[string]string{
			"entry1": "one two three four five six seven eight nine ten eleven twelve",
		}

		accessRecords := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  1,
				LastAccessed: now.Add(-10 * time.Minute),
				Content:      "one two three four five six seven eight nine ten eleven twelve",
			},
		}

		report, err := curator.Run(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(report.Compactions) < 1 {
			t.Fatal("expected at least 1 compaction suggestion")
		}
		if report.TokenSavings <= 0 {
			t.Errorf("expected positive token savings, got %d", report.TokenSavings)
		}
	})

	t.Run("empty input handling", func(t *testing.T) {
		cfg := newTestConfig()
		curator := NewCuratorService(cfg)

		report, err := curator.Run(context.Background(), nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.MemoryState.TotalEntries != 0 {
			t.Errorf("expected 0 total entries, got %d", report.MemoryState.TotalEntries)
		}
		if len(report.Duplicates) != 0 {
			t.Errorf("expected no duplicates, got %d", len(report.Duplicates))
		}
		if len(report.Compactions) != 0 {
			t.Errorf("expected no compactions, got %d", len(report.Compactions))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		cfg := newTestConfig()
		curator := NewCuratorService(cfg)

		contentEntries := map[string]string{
			"entry1": "content one",
			"entry2": "content two",
			"entry3": "content three",
		}

		accessRecords := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  10,
				LastAccessed: now.Add(-5 * time.Minute),
				Content:      "content one",
			},
			"entry2": {
				ID:           "entry2",
				Source:       "task-local",
				AccessCount:  5,
				LastAccessed: now.Add(-10 * time.Minute),
				Content:      "content two",
			},
			"entry3": {
				ID:           "entry3",
				Source:       "workspace",
				AccessCount:  3,
				LastAccessed: now.Add(-15 * time.Minute),
				Content:      "content three",
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := curator.Run(ctx, contentEntries, accessRecords)
		if err == nil {
			t.Error("expected error on cancelled context")
		}
	})
}
