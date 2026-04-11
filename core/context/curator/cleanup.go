package curator

import (
	"context"
	"fmt"
	"time"
)

// CleanupSuggestion represents content recommended for archival/removal
type CleanupSuggestion struct {
	ID           string
	Reason       string // "stale", "low-utility", "superseded"
	LastAccessed time.Time
	AccessCount  int
	Age          time.Duration
	Size         int
}

// CleanupAdvisor recommends content for cleanup
type CleanupAdvisor struct {
	staleThreshold time.Duration
	minAccessCount int
}

// NewCleanupAdvisor creates a new cleanup advisor
func NewCleanupAdvisor(staleThreshold time.Duration, minAccessCount int) *CleanupAdvisor {
	return &CleanupAdvisor{
		staleThreshold: staleThreshold,
		minAccessCount: minAccessCount,
	}
}

// Analyze finds entries that should be cleaned up
func (a *CleanupAdvisor) Analyze(ctx context.Context, entries map[string]AccessRecord) ([]CleanupSuggestion, error) {
	if entries == nil {
		return []CleanupSuggestion{}, nil
	}

	var suggestions []CleanupSuggestion
	now := time.Now()

	for id, record := range entries {
		select {
		case <-ctx.Done():
			return suggestions, fmt.Errorf("cleanup analysis cancelled: %w", ctx.Err())
		default:
		}

		age := now.Sub(record.LastAccessed)
		size := len(record.Content)

		// Check for stale entries
		if age > a.staleThreshold {
			suggestions = append(suggestions, CleanupSuggestion{
				ID:           id,
				Reason:       "stale",
				LastAccessed: record.LastAccessed,
				AccessCount:  record.AccessCount,
				Age:          age,
				Size:         size,
			})
			continue
		}

		// Check for low-utility entries
		if record.AccessCount < a.minAccessCount && age > a.staleThreshold/2 {
			suggestions = append(suggestions, CleanupSuggestion{
				ID:           id,
				Reason:       "low-utility",
				LastAccessed: record.LastAccessed,
				AccessCount:  record.AccessCount,
				Age:          age,
				Size:         size,
			})
		}
	}

	return suggestions, nil
}
