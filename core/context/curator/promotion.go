package curator

import (
	"context"
	"fmt"
	"time"
)

// AccessRecord tracks access patterns for context entries
type AccessRecord struct {
	ID           string
	Source       string
	AccessCount  int
	LastAccessed time.Time
	Content      string
}

// PromotionSuggestion represents content that should move from short-term to long-term
type PromotionSuggestion struct {
	ID           string
	Source       string // "episodic", "task-local", "workspace"
	Destination  string // "structured", "semantic", "long-term"
	Reason       string
	AccessCount  int
	LastAccessed time.Time
}

// Promoter analyzes access patterns to suggest promotions
type Promoter struct {
	minAccessCount int
	recencyWindow  time.Duration
}

// NewPromoter creates a new promoter
func NewPromoter(minAccessCount int, recencyWindow time.Duration) *Promoter {
	return &Promoter{
		minAccessCount: minAccessCount,
		recencyWindow:  recencyWindow,
	}
}

// Analyze finds entries that should be promoted to long-term storage
func (p *Promoter) Analyze(ctx context.Context, entries map[string]AccessRecord) ([]PromotionSuggestion, error) {
	if entries == nil {
		return []PromotionSuggestion{}, nil
	}

	var suggestions []PromotionSuggestion
	now := time.Now()

	for id, record := range entries {
		select {
		case <-ctx.Done():
			return suggestions, fmt.Errorf("promotion analysis cancelled: %w", ctx.Err())
		default:
		}

		// Check if entry meets access count threshold
		if record.AccessCount < p.minAccessCount {
			continue
		}

		// Check if entry was accessed recently
		age := now.Sub(record.LastAccessed)
		if age > p.recencyWindow {
			continue
		}

		// Check if source is short-term
		if !p.isShortTermSource(record.Source) {
			continue
		}

		suggestions = append(suggestions, PromotionSuggestion{
			ID:           id,
			Source:       record.Source,
			Destination:  p.mapDestination(record.Source),
			Reason:       fmt.Sprintf("Frequently accessed (%d times) in short-term source %s", record.AccessCount, record.Source),
			AccessCount:  record.AccessCount,
			LastAccessed: record.LastAccessed,
		})
	}

	return suggestions, nil
}

// isShortTermSource checks if a source is considered short-term
func (p *Promoter) isShortTermSource(source string) bool {
	shortTermSources := map[string]bool{
		"episodic":   true,
		"task-local": true,
		"workspace":  true,
	}
	return shortTermSources[source]
}

// mapDestination maps a short-term source to its appropriate long-term destination
func (p *Promoter) mapDestination(source string) string {
	destMap := map[string]string{
		"episodic":   "semantic",
		"task-local": "structured",
		"workspace":  "long-term",
	}
	if dest, ok := destMap[source]; ok {
		return dest
	}
	return "long-term"
}
