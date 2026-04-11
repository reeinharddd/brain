package curator

import (
	"context"
	"sync"
	"testing"
	"time"
)

func newTestAutoDreamConfig() CuratorConfig {
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

func TestAutoDreamService(t *testing.T) {
	now := time.Now()

	t.Run("session recording increments counter", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)

		err := service.RecordSession(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		status := service.GetStatus()
		if status["session_counter"] != 1 {
			t.Errorf("expected session counter 1, got %v", status["session_counter"])
		}
	})

	t.Run("tick triggers curation when conditions met", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)
		// Set minHoursDream to 0 to allow immediate triggering
		service.minHoursDream = 0
		service.minSessions = 1

		// Record enough sessions
		err := service.RecordSession(context.Background())
		if err != nil {
			t.Fatalf("unexpected error recording session: %v", err)
		}

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

		report, err := service.Tick(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report == nil {
			t.Fatal("expected report, got nil")
		}
	})

	t.Run("tick skips when too soon", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)
		service.minHoursDream = time.Hour // Require 1 hour between dreams
		service.minSessions = 1
		service.lastDreamTime = time.Now() // Set last dream to now

		// Record enough sessions
		err := service.RecordSession(context.Background())
		if err != nil {
			t.Fatalf("unexpected error recording session: %v", err)
		}

		contentEntries := map[string]string{
			"entry1": "test content",
		}
		accessRecords := map[string]AccessRecord{}

		report, err := service.Tick(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report != nil {
			t.Error("expected nil report when too soon")
		}
	})

	t.Run("tick skips when not enough sessions", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)
		service.minHoursDream = 0
		service.minSessions = 5 // Require 5 sessions

		contentEntries := map[string]string{
			"entry1": "test content",
		}
		accessRecords := map[string]AccessRecord{}

		// Only record 2 sessions
		service.RecordSession(context.Background())
		service.RecordSession(context.Background())

		report, err := service.Tick(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report != nil {
			t.Error("expected nil report when not enough sessions")
		}
	})

	t.Run("shouldTrigger logic", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)
		service.minHoursDream = 0
		service.minSessions = 2

		// Should not trigger without sessions
		if service.ShouldTrigger() {
			t.Error("should not trigger without enough sessions")
		}

		// Record sessions
		service.RecordSession(context.Background())
		if service.ShouldTrigger() {
			t.Error("should not trigger with only 1 session")
		}

		service.RecordSession(context.Background())
		if !service.ShouldTrigger() {
			t.Error("should trigger with enough sessions and minHoursDream=0")
		}
	})

	t.Run("status reporting", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)

		status := service.GetStatus()

		if _, ok := status["last_dream_time"]; !ok {
			t.Error("missing last_dream_time in status")
		}
		if _, ok := status["session_counter"]; !ok {
			t.Error("missing session_counter in status")
		}
		if _, ok := status["should_trigger"]; !ok {
			t.Error("missing should_trigger in status")
		}
	})

	t.Run("concurrent session recording", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)

		var wg sync.WaitGroup
		numGoroutines := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := service.RecordSession(context.Background())
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}()
		}

		wg.Wait()

		status := service.GetStatus()
		if status["session_counter"] != numGoroutines {
			t.Errorf("expected session counter %d, got %v", numGoroutines, status["session_counter"])
		}
	})

	t.Run("tick resets counters after successful run", func(t *testing.T) {
		cfg := newTestAutoDreamConfig()
		service := NewAutoDreamService(cfg, time.Hour)
		service.minHoursDream = 0
		service.minSessions = 1

		// Record 3 sessions
		for i := 0; i < 3; i++ {
			service.RecordSession(context.Background())
		}

		contentEntries := map[string]string{}
		accessRecords := map[string]AccessRecord{}

		report, err := service.Tick(context.Background(), contentEntries, accessRecords)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report == nil {
			t.Fatal("expected report, got nil")
		}

		// Verify counters reset
		status := service.GetStatus()
		if status["session_counter"] != 0 {
			t.Errorf("expected session counter 0 after tick, got %v", status["session_counter"])
		}
	})
}
