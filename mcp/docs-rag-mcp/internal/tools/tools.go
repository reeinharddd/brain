package tools

import (
	"context"
	"fmt"

	"github.com/reeinharrrd/brain/mcp/docs-rag-mcp/internal/indexer"
)

// SearchRequest represents a search tool request.
type SearchRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Domain string `json:"domain,omitempty"`
}

// SearchResponse represents search results.
type SearchResponse struct {
	Results   []indexer.SearchResult `json:"results"`
	Metadata  SearchMetadata         `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
}

// SearchMetadata contains search metadata.
type SearchMetadata struct {
	TotalIndexed  int     `json:"total_indexed"`
	QueryTimeMs   int64   `json:"query_time_ms"`
	IndexStatus   string  `json:"index_status"`
	ResultsCount  int     `json:"results_count"`
}

// StatusRequest represents a status check request.
type StatusRequest struct{}

// StatusResponse returns the current index status.
type StatusResponse struct {
	IndexStatus indexer.IndexStatus `json:"index_status"`
	Error       string              `json:"error,omitempty"`
}

// RebuildRequest requests an index rebuild.
type RebuildRequest struct {
	Domains []string `json:"domains,omitempty"`
}

// RebuildResponse returns rebuild status.
type RebuildResponse struct {
	Success       bool   `json:"success"`
	DocumentCount int    `json:"document_count"`
	Duration      string `json:"duration"`
	Error         string `json:"error,omitempty"`
}

// DocsSearch performs a semantic search over documentation.
func DocsSearch(ctx context.Context, idx *indexer.Indexer, req SearchRequest) SearchResponse {
	if req.Query == "" {
		return SearchResponse{
			Error: "query cannot be empty",
		}
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Search using the indexer
	results, err := idx.Search(ctx, req.Query, req.Limit, req.Domain)
	if err != nil {
		return SearchResponse{
			Error: fmt.Sprintf("search failed: %v", err),
		}
	}

	status := idx.GetStatus()

	return SearchResponse{
		Results: results,
		Metadata: SearchMetadata{
			TotalIndexed: status.DocumentCount,
			IndexStatus:  status.State,
			ResultsCount: len(results),
		},
	}
}

// DocsStatus returns the current index status.
func DocsStatus(ctx context.Context, idx *indexer.Indexer) StatusResponse {
	status := idx.GetStatus()

	return StatusResponse{
		IndexStatus: status,
	}
}

// DocsRebuild rebuilds the document index (dev-only).
func DocsRebuild(ctx context.Context, idx *indexer.Indexer, brainEnv string) RebuildResponse {
	// Only allow rebuild in development
	if brainEnv == "production" {
		return RebuildResponse{
			Error: "rebuild disabled in production",
		}
	}

	// Force rebuild by calling Build directly
	err := idx.Build(ctx)
	if err != nil {
		return RebuildResponse{
			Error: fmt.Sprintf("rebuild failed: %v", err),
		}
	}

	status := idx.GetStatus()

	return RebuildResponse{
		Success:       true,
		DocumentCount: status.DocumentCount,
		Duration:      status.LastRebuildTime,
	}
}

// ValidationResponse represents validation result.
type ValidationResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// ValidateTools checks that all tools are properly defined.
func ValidateTools() ValidationResponse {
	toolErrors := []string{}

	// Verify each tool can be called
	// This is a placeholder for tool validation

	return ValidationResponse{
		Valid:  len(toolErrors) == 0,
		Errors: toolErrors,
	}
}
