package main

import "testing"

func TestFormatTags(t *testing.T) {
	t.Run("returns empty for non-slice input", func(t *testing.T) {
		if got := formatTags("not-a-slice"); got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("formats string tags with commas", func(t *testing.T) {
		input := []interface{}{"go", "testing", "brain"}
		if got := formatTags(input); got != "go, testing, brain" {
			t.Fatalf("unexpected tags output: %q", got)
		}
	})
}

func TestFormatTagsSkipsNonStringValues(t *testing.T) {
	input := []interface{}{"go", 42, true, "cli"}
	if got := formatTags(input); got != "go, cli" {
		t.Fatalf("expected only string tags, got %q", got)
	}
}
