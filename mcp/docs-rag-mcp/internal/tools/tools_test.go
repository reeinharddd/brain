package tools

import (
	"context"
	"testing"
	"time"

	"github.com/reeinharrrd/brain/mcp/docs-rag-mcp/internal/indexer"
)

// TestDocsSearch tests the docs_search tool.
func TestDocsSearch(t *testing.T) {
	// Create a test indexer
	idx, err := indexer.NewIndexer("/tmp")
	if err != nil {
		t.Fatalf("failed to create indexer: %v", err)
	}

	tests := []struct {
		name      string
		req       SearchRequest
		wantError bool
	}{
		{
			name:      "empty query",
			req:       SearchRequest{Query: ""},
			wantError: true,
		},
		{
			name:      "valid query",
			req:       SearchRequest{Query: "authentication", Limit: 5},
			wantError: false,
		},
		{
			name:      "query with domain",
			req:       SearchRequest{Query: "daemon", Domain: "architecture"},
			wantError: false,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := DocsSearch(ctx, idx, tt.req)
			if (resp.Error != "") != tt.wantError {
				t.Fatalf("got error: %v, wantError: %v", resp.Error, tt.wantError)
			}
		})
	}
}

// TestDocsStatus tests the docs_status tool.
func TestDocsStatus(t *testing.T) {
	idx, err := indexer.NewIndexer("/tmp")
	if err != nil {
		t.Fatalf("failed to create indexer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp := DocsStatus(ctx, idx)

	if resp.Error != "" {
		t.Fatalf("DocsStatus returned error: %v", resp.Error)
	}

	if resp.IndexStatus.State == "" {
		t.Errorf("status missing state field")
	}
}

// TestDocsRebuild tests the docs_rebuild tool.
func TestDocsRebuild_DevAllowed(t *testing.T) {
	idx, err := indexer.NewIndexer("/tmp")
	if err != nil {
		t.Fatalf("failed to create indexer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp := DocsRebuild(ctx, idx, "development")

	// Should succeed or fail due to missing docs, not due to env check
	if resp.Error != "" && resp.Error == "rebuild disabled in production" {
		t.Fatalf("rebuild should be allowed in development")
	}
}

// TestDocsRebuild_ProdBlocked tests that rebuild is blocked in production.
func TestDocsRebuild_ProdBlocked(t *testing.T) {
	idx, err := indexer.NewIndexer("/tmp")
	if err != nil {
		t.Fatalf("failed to create indexer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp := DocsRebuild(ctx, idx, "production")

	if resp.Error == "" {
		t.Fatal("rebuild should be blocked in production")
	}

	if resp.Error != "rebuild disabled in production" {
		t.Errorf("expected rebuild blocked error, got: %v", resp.Error)
	}
}

// TestValidateTools tests tool validation.
func TestValidateTools(t *testing.T) {
	resp := ValidateTools()

	if !resp.Valid {
		t.Fatalf("tool validation failed: %v", resp.Errors)
	}

	if len(resp.Errors) > 0 {
		t.Logf("validation errors: %v", resp.Errors)
	}
}

// TestSearchRequest_Validation tests search request validation.
func TestSearchRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		limit   int
		wantErr bool
	}{
		{
			name:    "empty query",
			query:   "",
			limit:   10,
			wantErr: true,
		},
		{
			name:    "valid query",
			query:   "daemon architecture",
			limit:   10,
			wantErr: false,
		},
		{
			name:    "zero limit uses default",
			query:   "test",
			limit:   0,
			wantErr: false,
		},
	}

	idx, _ := indexer.NewIndexer("/tmp")
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SearchRequest{Query: tt.query, Limit: tt.limit}
			resp := DocsSearch(ctx, idx, req)
			if (resp.Error != "") != tt.wantErr {
				t.Fatalf("got error: %v, wantErr: %v", resp.Error, tt.wantErr)
			}
		})
	}
}
