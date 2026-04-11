package curator

import (
	"context"
	"testing"
	"time"
)

func TestCleanupAdvisor(t *testing.T) {
	now := time.Now()

	t.Run("stale entries flagged", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 3)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				AccessCount:  10,
				LastAccessed: now.Add(-2 * time.Hour),
				Content:      "some content",
			},
		}

		suggestions, err := advisor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].Reason != "stale" {
			t.Errorf("expected stale reason, got %s", suggestions[0].Reason)
		}
	})

	t.Run("low-utility entries flagged", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 5)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				AccessCount:  2,
				LastAccessed: now.Add(-40 * time.Minute), // > 30 min (staleThreshold/2)
				Content:      "some content",
			},
		}

		suggestions, err := advisor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].Reason != "low-utility" {
			t.Errorf("expected low-utility reason, got %s", suggestions[0].Reason)
		}
	})

	t.Run("recent frequent entries not flagged", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 3)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				AccessCount:  10,
				LastAccessed: now.Add(-10 * time.Minute),
				Content:      "some content",
			},
		}

		suggestions, err := advisor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("expected no suggestions, got %d", len(suggestions))
		}
	})

	t.Run("recent low access entries not flagged", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 5)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				AccessCount:  1,
				LastAccessed: now.Add(-10 * time.Minute), // < 30 min (staleThreshold/2)
				Content:      "some content",
			},
		}

		suggestions, err := advisor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("expected no suggestions, got %d", len(suggestions))
		}
	})

	t.Run("empty entries", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 3)

		suggestions, err := advisor.Analyze(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if suggestions == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 3)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				AccessCount:  10,
				LastAccessed: now.Add(-2 * time.Hour),
				Content:      "some content",
			},
			"entry2": {
				ID:           "entry2",
				AccessCount:  5,
				LastAccessed: now.Add(-3 * time.Hour),
				Content:      "more content",
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := advisor.Analyze(ctx, entries)
		if err == nil {
			t.Error("expected error on cancelled context")
		}
	})

	t.Run("size calculation", func(t *testing.T) {
		advisor := NewCleanupAdvisor(time.Hour, 3)
		content := "this is some test content"
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				AccessCount:  10,
				LastAccessed: now.Add(-2 * time.Hour),
				Content:      content,
			},
		}

		suggestions, err := advisor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].Size != len(content) {
			t.Errorf("expected size %d, got %d", len(content), suggestions[0].Size)
		}
	})
}
