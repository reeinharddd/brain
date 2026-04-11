package curator

import (
	"context"
	"fmt"
	"strings"
)

// CompactionSuggestion represents a suggestion to compact content
type CompactionSuggestion struct {
	ID              string
	CurrentTokens   int
	SuggestedTokens int
	Method          string // "summarize", "truncate", "remove_examples"
	SavingsTokens   int
	Reason          string
}

// Compactor analyzes context for compaction opportunities
type Compactor struct {
	maxTokensPerLayer int
}

// NewCompactor creates a new compactor
func NewCompactor(maxTokensPerLayer int) *Compactor {
	return &Compactor{
		maxTokensPerLayer: maxTokensPerLayer,
	}
}

// Analyze finds entries that exceed token limits and suggests compaction
func (c *Compactor) Analyze(ctx context.Context, entries map[string]string) ([]CompactionSuggestion, error) {
	if entries == nil {
		return []CompactionSuggestion{}, nil
	}

	var suggestions []CompactionSuggestion

	for id, content := range entries {
		select {
		case <-ctx.Done():
			return suggestions, fmt.Errorf("compaction analysis cancelled: %w", ctx.Err())
		default:
		}

		tokens := c.estimateTokens(content)
		if tokens > c.maxTokensPerLayer {
			var method string
			var suggested int
			if tokens > c.maxTokensPerLayer*2 {
				method = "summarize"
				suggested = c.maxTokensPerLayer
			} else {
				method = "truncate"
				suggested = c.maxTokensPerLayer
			}

			suggestions = append(suggestions, CompactionSuggestion{
				ID:              id,
				CurrentTokens:   tokens,
				SuggestedTokens: suggested,
				Method:          method,
				SavingsTokens:   tokens - suggested,
				Reason:          fmt.Sprintf("Entry %s exceeds token limit (%d > %d)", id, tokens, c.maxTokensPerLayer),
			})
		}
	}

	return suggestions, nil
}

// estimateTokens estimates the token count for content
func (c *Compactor) estimateTokens(content string) int {
	if content == "" {
		return 0
	}
	// Simple estimation: ~1 token per word for typical English text
	words := strings.Fields(content)
	return len(words)
}
