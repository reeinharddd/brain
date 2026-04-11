package context

import (
	"testing"
)

func TestProgressiveDisclosure(t *testing.T) {
	t.Run("register and get summary", func(t *testing.T) {
		pd := NewProgressiveDisclosure()
		content := ProgressiveContent{
			Summary:  "This is a summary",
			Full:     "This is the full content with all details",
			Triggers: []string{"detail", "full"},
		}

		pd.Register("test-skill", content)

		summary := pd.GetSummary("test-skill")
		if summary != "This is a summary" {
			t.Errorf("expected summary 'This is a summary', got '%s'", summary)
		}
	})

	t.Run("register and get full content", func(t *testing.T) {
		pd := NewProgressiveDisclosure()
		content := ProgressiveContent{
			Summary:  "Summary",
			Full:     "Full detailed content",
			Triggers: []string{"detail"},
		}

		pd.Register("test-skill", content)

		full := pd.GetFull("test-skill")
		if full != "Full detailed content" {
			t.Errorf("expected full 'Full detailed content', got '%s'", full)
		}
	})

	t.Run("should reveal with matching trigger keywords", func(t *testing.T) {
		pd := NewProgressiveDisclosure()
		content := ProgressiveContent{
			Summary:  "Summary",
			Full:     "Full content",
			Triggers: []string{"detail", "explain", "full"},
		}

		pd.Register("test-skill", content)

		tests := []struct {
			name  string
			query string
			want  bool
		}{
			{"exact match", "detail", true},
			{"substring match", "please explain this", true},
			{"case insensitive", "DETAIL please", true},
			{"mixed case trigger", "can you show me the Full version", true},
			{"no match", "something else", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := pd.ShouldReveal("test-skill", tt.query)
				if got != tt.want {
					t.Errorf("ShouldReveal(%q): expected %v, got %v", tt.query, tt.want, got)
				}
			})
		}
	})

	t.Run("should reveal without matching keywords returns false", func(t *testing.T) {
		pd := NewProgressiveDisclosure()
		content := ProgressiveContent{
			Summary:  "Summary",
			Full:     "Full content",
			Triggers: []string{"specific-keyword"},
		}

		pd.Register("test-skill", content)

		got := pd.ShouldReveal("test-skill", "this query has no matching keywords")
		if got {
			t.Error("ShouldReveal should return false for non-matching query")
		}
	})

	t.Run("content for disclosure factory method", func(t *testing.T) {
		pd := NewProgressiveDisclosure()

		content := pd.ContentForDisclosure(
			"skill-1",
			"Skill summary",
			"Full skill content with details",
			[]string{"explain", "details"},
		)

		if content.Summary != "Skill summary" {
			t.Errorf("expected summary 'Skill summary', got '%s'", content.Summary)
		}
		if content.Full != "Full skill content with details" {
			t.Errorf("expected full content, got '%s'", content.Full)
		}
		if len(content.Triggers) != 2 {
			t.Errorf("expected 2 triggers, got %d", len(content.Triggers))
		}
		if content.Triggers[0] != "explain" {
			t.Errorf("expected first trigger 'explain', got '%s'", content.Triggers[0])
		}
	})

	t.Run("missing ID handling returns empty strings", func(t *testing.T) {
		pd := NewProgressiveDisclosure()

		// Register one item
		pd.Register("existing", ProgressiveContent{
			Summary:  "Summary",
			Full:     "Full",
			Triggers: []string{"trigger"},
		})

		// Try to get content for non-existing IDs
		summary := pd.GetSummary("non-existing")
		if summary != "" {
			t.Errorf("expected empty summary for non-existing ID, got '%s'", summary)
		}

		full := pd.GetFull("non-existing")
		if full != "" {
			t.Errorf("expected empty full content for non-existing ID, got '%s'", full)
		}

		// ShouldReveal for non-existing ID
		reveal := pd.ShouldReveal("non-existing", "trigger")
		if reveal {
			t.Error("ShouldReveal should return false for non-existing ID")
		}
	})

	t.Run("empty triggers list returns false", func(t *testing.T) {
		pd := NewProgressiveDisclosure()
		content := ProgressiveContent{
			Summary:  "Summary",
			Full:     "Full content",
			Triggers: []string{},
		}

		pd.Register("no-triggers", content)

		got := pd.ShouldReveal("no-triggers", "any query")
		if got {
			t.Error("ShouldReveal should return false for empty triggers list")
		}
	})

	t.Run("multiple registrations", func(t *testing.T) {
		pd := NewProgressiveDisclosure()

		contents := []struct {
			id      string
			summary string
			full    string
		}{
			{"skill-1", "Skill 1 summary", "Skill 1 full"},
			{"skill-2", "Skill 2 summary", "Skill 2 full"},
			{"mcp-1", "MCP 1 summary", "MCP 1 full"},
		}

		for _, c := range contents {
			pd.Register(c.id, ProgressiveContent{
				Summary:  c.summary,
				Full:     c.full,
				Triggers: []string{c.id},
			})
		}

		// Verify all contents are accessible
		for _, c := range contents {
			summary := pd.GetSummary(c.id)
			if summary != c.summary {
				t.Errorf("summary for %s: expected '%s', got '%s'", c.id, c.summary, summary)
			}

			full := pd.GetFull(c.id)
			if full != c.full {
				t.Errorf("full for %s: expected '%s', got '%s'", c.id, c.full, full)
			}
		}
	})
}
