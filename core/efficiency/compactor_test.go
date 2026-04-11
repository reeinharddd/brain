package efficiency

import (
	"context"
	"strings"
	"testing"
)

func TestCompactor_TokenEstimation(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minTokens int
		maxTokens int
	}{
		{
			name:     "empty text",
			text:     "",
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:     "short text",
			text:     "hello",
			minTokens: 1,
			maxTokens: 2,
		},
		{
			name:     "medium text",
			text:     "hello world this is a test of token estimation",
			minTokens: 10,
			maxTokens: 15,
		},
		{
			name:     "long text",
			text:     strings.Repeat("a", 400),
			minTokens: 90,
			maxTokens: 110,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompactor(100, "moderate")
			estimated := c.EstimateTokens(tt.text)

			if estimated < tt.minTokens || estimated > tt.maxTokens {
				t.Errorf("token estimate %d outside expected range [%d, %d]", estimated, tt.minTokens, tt.maxTokens)
			}
		})
	}
}

func TestCompactor_CompactionReducesTokens(t *testing.T) {
	content := strings.Repeat("This is a test line of text. ", 100)
	c := NewCompactor(100, "moderate")
	currentTokens := c.EstimateTokens(content)

	ctx := context.Background()
	result, compacted := c.Compact(ctx, content, currentTokens)

	if result.CompactedTokens >= result.OriginalTokens {
		t.Errorf("expected compaction to reduce tokens, got original=%d, compacted=%d",
			result.OriginalTokens, result.CompactedTokens)
	}

	if len(compacted) >= len(content) {
		t.Errorf("expected compacted content to be shorter, got original=%d, compacted=%d",
			len(content), len(compacted))
	}
}

func TestCompactor_StrategyOrdering(t *testing.T) {
	content := strings.Repeat("This is a test line of text. ", 200)
	comp := NewCompactor(100, "moderate")
	currentTokens := comp.EstimateTokens(content)

	ctx := context.Background()

	// Set maxTokens between moderate and minimal targets so all strategies trigger compaction
	// currentTokens ~1500, aggressive targets 375, moderate 750, minimal 1125
	// With maxTokens=800, all must compact but produce different results
	aggressive := NewCompactor(800, "aggressive")
	moderate := NewCompactor(800, "moderate")
	minimal := NewCompactor(800, "minimal")

	aggResult, _ := aggressive.Compact(ctx, content, currentTokens)
	modResult, _ := moderate.Compact(ctx, content, currentTokens)
	minResult, _ := minimal.Compact(ctx, content, currentTokens)

	// Aggressive should save more than moderate, moderate more than minimal
	if aggResult.CompactedTokens >= modResult.CompactedTokens {
		t.Errorf("expected aggressive to produce fewer tokens than moderate: aggressive=%d, moderate=%d",
			aggResult.CompactedTokens, modResult.CompactedTokens)
	}

	if modResult.CompactedTokens >= minResult.CompactedTokens {
		t.Errorf("expected moderate to produce fewer tokens than minimal: moderate=%d, minimal=%d",
			modResult.CompactedTokens, minResult.CompactedTokens)
	}
}

func TestCompactor_RemoveRedundant(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no duplicates",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "consecutive duplicates",
			input:    "line1\nline1\nline2\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "non-consecutive duplicates",
			input:    "line1\nline2\nline1\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "blank lines collapsed",
			input:    "line1\n\n\nline2",
			expected: "line1\n\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompactor(100, "minimal")
			result := c.RemoveRedundant(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCompactor_EmptyInput(t *testing.T) {
	c := NewCompactor(100, "moderate")
	ctx := context.Background()

	result, compacted := c.Compact(ctx, "", 0)

	if result.OriginalTokens != 0 {
		t.Errorf("expected 0 original tokens, got %d", result.OriginalTokens)
	}
	if result.CompactedTokens != 0 {
		t.Errorf("expected 0 compacted tokens, got %d", result.CompactedTokens)
	}
	if compacted != "" {
		t.Errorf("expected empty compacted content, got %q", compacted)
	}
}

func TestCompactor_ContextCancellation(t *testing.T) {
	c := NewCompactor(100, "moderate")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	content := "some content that would be compacted"
	currentTokens := c.EstimateTokens(content)

	result, compacted := c.Compact(ctx, content, currentTokens)

	if result.Method != "none" {
		t.Errorf("expected no compaction with cancelled context, got method %q", result.Method)
	}
	if compacted != content {
		t.Errorf("expected original content with cancelled context, got %q", compacted)
	}
}

func TestCompactor_NoCompactionNeeded(t *testing.T) {
	content := "short content"
	c := NewCompactor(1000, "moderate")
	currentTokens := c.EstimateTokens(content)

	ctx := context.Background()
	result, compacted := c.Compact(ctx, content, currentTokens)

	if result.Method != "none" {
		t.Errorf("expected no compaction, got method %q", result.Method)
	}
	if compacted != content {
		t.Errorf("expected original content, got %q", compacted)
	}
}

func TestCompactor_AggressiveStrategy(t *testing.T) {
	content := strings.Repeat("This is a test line of text. ", 100)
	c := NewCompactor(100, "aggressive")
	currentTokens := c.EstimateTokens(content)

	ctx := context.Background()
	result, _ := c.Compact(ctx, content, currentTokens)

	if result.Method != "truncate" {
		t.Errorf("expected method 'truncate' for aggressive, got %q", result.Method)
	}
	if result.RiskLevel != "high" {
		t.Errorf("expected risk level 'high' for aggressive, got %q", result.RiskLevel)
	}
	if result.SavingsPercent <= 0 {
		t.Errorf("expected positive savings, got %.2f", result.SavingsPercent)
	}
}

func TestCompactor_ModerateStrategy(t *testing.T) {
	content := strings.Repeat("This is a test line of text. ", 100)
	c := NewCompactor(500, "moderate") // maxTokens below currentTokens to trigger compaction
	currentTokens := c.EstimateTokens(content)

	ctx := context.Background()
	result, _ := c.Compact(ctx, content, currentTokens)

	if result.Method != "summarize" {
		t.Errorf("expected method 'summarize' for moderate, got %q", result.Method)
	}
	if result.RiskLevel != "medium" {
		t.Errorf("expected risk level 'medium' for moderate, got %q", result.RiskLevel)
	}
}

func TestCompactor_MinimalStrategy(t *testing.T) {
	// Create content with many redundant lines
	content := strings.Repeat("duplicate line\n", 50)
	c := NewCompactor(100, "minimal")
	currentTokens := c.EstimateTokens(content)

	ctx := context.Background()
	result, compacted := c.Compact(ctx, content, currentTokens)

	if result.Method != "remove_redundant" {
		t.Errorf("expected method 'remove_redundant' for minimal, got %q", result.Method)
	}
	if result.RiskLevel != "low" {
		t.Errorf("expected risk level 'low' for minimal, got %q", result.RiskLevel)
	}

	// After removing redundant lines, should be significantly shorter
	if len(compacted) >= len(content) {
		t.Errorf("expected compacted content to be shorter after removing duplicates, got original=%d, compacted=%d", len(content), len(compacted))
	}
}

func TestCompactor_Summarize(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "short string returns shortened version",
			input: "short",
		},
		{
			name:  "long string is truncated",
			input: strings.Repeat("a", 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompactor(100, "moderate")
			result := c.Summarize(tt.input)

			if len(result) > len(tt.input) {
				t.Errorf("summarized result longer than input")
			}

			if len(tt.input) > 20 {
				// For longer inputs, summarized content should start with the same prefix
				if !strings.HasPrefix(result, tt.input[:10]) {
					t.Errorf("summarized result should start with same prefix as input")
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
