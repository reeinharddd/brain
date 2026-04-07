package main

import (
	"context"
	"fmt"
	"time"

	"github.com/reeinharrrd/brain/daemon/internal/api/handlers"
)

// StubIndexer is a test implementation of the indexer interface
type StubIndexer struct{}

// Search returns mock search results for testing
func (s *StubIndexer) Search(ctx context.Context, query string, limit int, domain string) ([]handlers.SearchResult, int, error) {
	// Simulate search latency
	time.Sleep(100 * time.Millisecond)

	// Return mock results for testing
	results := []handlers.SearchResult{
		{
			Title:       fmt.Sprintf("Architecture Overview: %s", query),
			Path:        "docs/architecture/system-design.md",
			Category:    "architecture",
			RagPriority: "1",
			Score:       0.98,
			Snippet:     "This is the main architecture document explaining the system design and components.",
		},
		{
			Title:       fmt.Sprintf("Testing Guide for %s", query),
			Path:        "docs/testing/strategies.md",
			Category:    "testing",
			RagPriority: "2",
			Score:       0.92,
			Snippet:     "Comprehensive testing strategies and best practices for validating functionality.",
		},
		{
			Title:       fmt.Sprintf("Security Considerations: %s", query),
			Path:        "docs/security/audit.md",
			Category:    "security",
			RagPriority: "2",
			Score:       0.87,
			Snippet:     "Security audit and vulnerability assessment for the system.",
		},
	}

	// Limit results
	if limit < len(results) {
		results = results[:limit]
	}

	return results, len(results), nil
}

// GetStatus returns the current index health status
func (s *StubIndexer) GetStatus() (handlers.IndexHealth, error) {
	return handlers.IndexHealth{
		State:           "ready",
		DocumentCount:   78,
		ChunkCount:      512,
		LastRebuildTime: time.Now().Add(-2 * time.Hour),
		QdrantHealth:    "healthy",
		Errors:          []string{},
	}, nil
}

// EnsureIndexBuilt ensures the index is built (no-op for stub)
func (s *StubIndexer) EnsureIndexBuilt(ctx context.Context) error {
	return nil
}
