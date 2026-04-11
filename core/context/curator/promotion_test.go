package curator

import (
	"context"
	"testing"
	"time"
)

func TestPromoter(t *testing.T) {
	now := time.Now()

	t.Run("frequently accessed entries get promoted", func(t *testing.T) {
		promoter := NewPromoter(5, time.Hour)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  10,
				LastAccessed: now.Add(-10 * time.Minute),
			},
		}

		suggestions, err := promoter.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Errorf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].Destination != "semantic" {
			t.Errorf("expected semantic destination, got %s", suggestions[0].Destination)
		}
	})

	t.Run("infrequently accessed entries don't promote", func(t *testing.T) {
		promoter := NewPromoter(5, time.Hour)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  2,
				LastAccessed: now.Add(-10 * time.Minute),
			},
		}

		suggestions, err := promoter.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("expected no suggestions, got %d", len(suggestions))
		}
	})

	t.Run("old entries don't promote", func(t *testing.T) {
		promoter := NewPromoter(5, time.Hour)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  10,
				LastAccessed: now.Add(-2 * time.Hour),
			},
		}

		suggestions, err := promoter.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("expected no suggestions for old entry, got %d", len(suggestions))
		}
	})

	t.Run("different source types map to correct destinations", func(t *testing.T) {
		tests := []struct {
			name        string
			source      string
			wantDest    string
		}{
			{"episodic maps to semantic", "episodic", "semantic"},
			{"task-local maps to structured", "task-local", "structured"},
			{"workspace maps to long-term", "workspace", "long-term"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				promoter := NewPromoter(5, time.Hour)
				entries := map[string]AccessRecord{
					"entry1": {
						ID:           "entry1",
						Source:       tt.source,
						AccessCount:  10,
						LastAccessed: now.Add(-10 * time.Minute),
					},
				}

				suggestions, err := promoter.Analyze(context.Background(), entries)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(suggestions) != 1 {
					t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
				}
				if suggestions[0].Destination != tt.wantDest {
					t.Errorf("expected destination %s, got %s", tt.wantDest, suggestions[0].Destination)
				}
			})
		}
	})

	t.Run("long-term sources don't promote", func(t *testing.T) {
		promoter := NewPromoter(5, time.Hour)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "structured",
				AccessCount:  10,
				LastAccessed: now.Add(-10 * time.Minute),
			},
		}

		suggestions, err := promoter.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("expected no suggestions for long-term source, got %d", len(suggestions))
		}
	})

	t.Run("empty entries", func(t *testing.T) {
		promoter := NewPromoter(5, time.Hour)

		suggestions, err := promoter.Analyze(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if suggestions == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		promoter := NewPromoter(5, time.Hour)
		entries := map[string]AccessRecord{
			"entry1": {
				ID:           "entry1",
				Source:       "episodic",
				AccessCount:  10,
				LastAccessed: now.Add(-10 * time.Minute),
			},
			"entry2": {
				ID:           "entry2",
				Source:       "task-local",
				AccessCount:  15,
				LastAccessed: now.Add(-5 * time.Minute),
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := promoter.Analyze(ctx, entries)
		if err == nil {
			t.Error("expected error on cancelled context")
		}
	})
}
