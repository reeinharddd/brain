package curator

import (
	"context"
	"testing"
	"time"
)

func TestDeduplicationDetector(t *testing.T) {
	t.Run("exact duplicate detection via hash", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.9)
		entries := map[string]string{
			"entry1": "the quick brown fox jumps over the lazy dog",
			"entry2": "the quick brown fox jumps over the lazy dog",
			"entry3": "a completely different string",
		}

		findings, err := detector.Detect(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hashDuplicates := 0
		for _, f := range findings {
			if f.Method == "hash" {
				hashDuplicates++
			}
		}
		if hashDuplicates != 1 {
			t.Errorf("expected 1 hash duplicate, got %d", hashDuplicates)
		}
	})

	t.Run("no duplicates in unique entries", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.9)
		entries := map[string]string{
			"entry1": "apple banana cherry",
			"entry2": "dog elephant frog",
			"entry3": "guitar harmony jazz",
		}

		findings, err := detector.Detect(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(findings) != 0 {
			t.Errorf("expected no findings, got %d", len(findings))
		}
	})

	t.Run("similarity-based detection", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.5)
		entries := map[string]string{
			"entry1": "the quick brown fox jumps over the lazy dog",
			"entry2": "the quick brown fox runs over the lazy cat",
		}

		findings, err := detector.Detect(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		simFindings := 0
		for _, f := range findings {
			if f.Method == "similarity" {
				simFindings++
			}
		}
		if simFindings != 1 {
			t.Errorf("expected 1 similarity finding, got %d", simFindings)
		}
	})

	t.Run("threshold behavior above", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.3)
		entries := map[string]string{
			"entry1": "hello world foo bar",
			"entry2": "hello world baz qux",
		}

		findings, err := detector.Detect(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(findings) < 1 {
			t.Error("expected findings above threshold")
		}
	})

	t.Run("threshold behavior below", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.95)
		entries := map[string]string{
			"entry1": "hello world foo bar",
			"entry2": "hello world baz qux",
		}

		findings, err := detector.Detect(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(findings) != 0 {
			t.Errorf("expected no findings below threshold, got %d", len(findings))
		}
	})

	t.Run("empty entries", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.5)

		findings, err := detector.Detect(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if findings == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("single entry", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.5)
		entries := map[string]string{
			"entry1": "only one entry",
		}

		findings, err := detector.Detect(context.Background(), entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(findings) != 0 {
			t.Errorf("expected no findings for single entry, got %d", len(findings))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.5)
		entries := map[string]string{
			"entry1": "content one",
			"entry2": "content two",
			"entry3": "content three",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := detector.Detect(ctx, entries)
		if err == nil {
			t.Error("expected error on cancelled context")
		}
	})
}

func TestDeduplicationDetector_Performance(t *testing.T) {
	t.Run("large entry set completes in time", func(t *testing.T) {
		detector := NewDeduplicationDetector(0.7)
		entries := make(map[string]string)
		for i := 0; i < 50; i++ {
			entries[string(rune(i))] = "sample content for entry number"
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := detector.Detect(ctx, entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
