package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SearchRequest holds search query parameters
type SearchRequest struct {
	Query  string `json:"q"`
	Limit  int    `json:"limit"`
	Domain string `json:"domain"`
}

// SearchResult is a single search result
type SearchResult struct {
	Title       string  `json:"title"`
	Path        string  `json:"path"`
	Category    string  `json:"category"`
	RagPriority string  `json:"rag_priority"`
	Score       float32 `json:"score"`
	Snippet     string  `json:"snippet"`
}

// SearchMetadata holds search response metadata
type SearchMetadata struct {
	TotalIndexed  int    `json:"total_indexed"`
	QueryTimeMs   int    `json:"query_time_ms"`
	IndexStatus   string `json:"index_status"`
	ResultsCount  int    `json:"results_count"`
}

// SearchResponse is the API response for search
type SearchResponse struct {
	Results  []SearchResult  `json:"results"`
	Metadata SearchMetadata  `json:"metadata"`
	Error    string          `json:"error,omitempty"`
}

// IndexHealth represents Qdrant health status
type IndexHealth struct {
	State            string    `json:"state"`
	DocumentCount    int       `json:"document_count"`
	ChunkCount       int       `json:"chunk_count"`
	LastRebuildTime  time.Time `json:"last_rebuild_time"`
	QdrantHealth     string    `json:"qdrant_health"`
	Errors           []string  `json:"errors"`
}

// StatusResponse is the API response for status
type StatusResponse struct {
	IndexStatus IndexHealth `json:"index_status"`
	Error       string      `json:"error,omitempty"`
}

// RebuildRequest holds rebuild parameters
type RebuildRequest struct {
	Domains []string `json:"domains,omitempty"`
}

// RebuildResponse is the API response for rebuild
type RebuildResponse struct {
	Success       bool          `json:"success"`
	DocumentCount int           `json:"document_count"`
	Duration      string        `json:"duration"`
	Error         string        `json:"error,omitempty"`
}

// DocsHandler handles documentation API endpoints
type DocsHandler struct {
	indexer interface {
		Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, int, error)
		GetStatus() (IndexHealth, error)
		EnsureIndexBuilt(ctx context.Context) error
	}
	brainEnv string
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler(indexer interface {
	Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, int, error)
	GetStatus() (IndexHealth, error)
	EnsureIndexBuilt(ctx context.Context) error
}, brainEnv string) *DocsHandler {
	return &DocsHandler{
		indexer:  indexer,
		brainEnv: brainEnv,
	}
}

// Search handles GET /api/docs/search
func (h *DocsHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SearchResponse{
			Error: "method not allowed",
		})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SearchResponse{
			Error: "query required",
		})
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	domain := r.URL.Query().Get("domain")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Ensure index is built
	if err := h.indexer.EnsureIndexBuilt(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SearchResponse{
			Error: fmt.Sprintf("index build failed: %v", err),
		})
		return
	}

	startTime := time.Now()
	results, _, err := h.indexer.Search(ctx, query, limit, domain)
	queryTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SearchResponse{
			Results: []SearchResult{},
			Metadata: SearchMetadata{
				QueryTimeMs:  queryTime,
				IndexStatus:  "error",
				ResultsCount: 0,
			},
			Error: fmt.Sprintf("search failed: %v", err),
		})
		return
	}

	status, _ := h.indexer.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SearchResponse{
		Results: results,
		Metadata: SearchMetadata{
			TotalIndexed:  status.DocumentCount,
			QueryTimeMs:   queryTime,
			IndexStatus:   status.State,
			ResultsCount:  len(results),
		},
	})
}

// Status handles GET /api/docs/status
func (h *DocsHandler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(StatusResponse{
			Error: "method not allowed",
		})
		return
	}

	status, err := h.indexer.GetStatus()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StatusResponse{
			IndexHealth{"not_built", 0, 0, time.Time{}, "unavailable", []string{err.Error()}},
			fmt.Sprintf("status check failed: %v", err),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		IndexStatus: status,
	})
}

// Rebuild handles POST /api/docs/rebuild (dev-only)
func (h *DocsHandler) Rebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(RebuildResponse{
			Error: "method not allowed",
		})
		return
	}

	// Block rebuild in production
	if h.brainEnv == "production" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(RebuildResponse{
			Success: false,
			Error:   "rebuild blocked in production",
		})
		return
	}

	var req RebuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = RebuildRequest{}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	if err := h.indexer.EnsureIndexBuilt(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RebuildResponse{
			Success: false,
			Error:   fmt.Sprintf("rebuild failed: %v", err),
		})
		return
	}

	status, _ := h.indexer.GetStatus()
	duration := time.Since(startTime)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RebuildResponse{
		Success:       true,
		DocumentCount: status.DocumentCount,
		Duration:      duration.String(),
	})
}

// RegisterRoutes registers documentation API routes
func (h *DocsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/docs/search", h.Search)
	mux.HandleFunc("GET /api/docs/status", h.Status)
	mux.HandleFunc("POST /api/docs/rebuild", h.Rebuild)
}
