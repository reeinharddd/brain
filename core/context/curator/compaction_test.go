package curator

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCompactor(t *testing.T) {
	t.Run("entries over limit get suggestions", func(t *testing.T) {
		compactor := NewCompactor(10)
		entries := map[string]string{
			"entry1": "one two three four five six seven eight nine ten eleven twelve",
		}

		suggestions, err := compactor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Errorf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].ID != "entry1" {
			t.Errorf("expected entry1, got %s", suggestions[0].ID)
		}
	})

	t.Run("entries under limit don't get suggestions", func(t *testing.T) {
		compactor := NewCompactor(100)
		entries := map[string]string{
			"entry1": "short content",
		}

		suggestions, err := compactor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("expected no suggestions, got %d", len(suggestions))
		}
	})

	t.Run("summarize method for entries over 2x limit", func(t *testing.T) {
		compactor := NewCompactor(10)
		// Create content with > 20 words (2x limit)
		content := strings.Repeat("word ", 25)
		entries := map[string]string{
			"entry1": content,
		}

		suggestions, err := compactor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].Method != "summarize" {
			t.Errorf("expected summarize method, got %s", suggestions[0].Method)
		}
	})

	t.Run("truncate method for entries between 1x and 2x limit", func(t *testing.T) {
		compactor := NewCompactor(10)
		// Create content with 15 words (between 10 and 20)
		content := strings.Repeat("word ", 15)
		entries := map[string]string{
			"entry1": content,
		}

		suggestions, err := compactor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
		}
		if suggestions[0].Method != "truncate" {
			t.Errorf("expected truncate method, got %s", suggestions[0].Method)
		}
	})

	t.Run("empty entries", func(t *testing.T) {
		compactor := NewCompactor(10)

		suggestions, err := compactor.Analyze(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if suggestions == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		compactor := NewCompactor(5)
		entries := map[string]string{
			"entry1": "one two three four five six seven",
			"entry2": "another long entry with many words to process",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := compactor.Analyze(ctx, entries)
		if err == nil {
			t.Error("expected error on cancelled context")
		}
	})

	t.Run("savings calculation", func(t *testing.T) {
		compactor := NewCompactor(10)
		// Content with 30 words
		content := strings.Repeat("word ", 30)
		entries := map[string]string{
			"entry1": content,
		}

		suggestions, err := compactor.Analyze(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
		}
		expectedSavings := 30 - 10
		if suggestions[0].SavingsTokens != expectedSavings {
			t.Errorf("expected savings %d, got %d", expectedSavings, suggestions[0].SavingsTokens)
		}
	})
}

func TestCompactor_Performance(t *testing.T) {
	t.Run("large entry set completes in time", func(t *testing.T) {
		compactor := NewCompactor(50)
		entries := make(map[string]string)
		for i := 0; i < 100; i++ {
			entries[fmt.Sprintf("entry%d", i)] = strings.Repeat("word ", 60)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := compactor.Analyze(ctx, entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
