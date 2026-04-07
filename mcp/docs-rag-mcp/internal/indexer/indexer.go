package indexer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Indexer manages document indexing with lazy-load on first search.
type Indexer struct {
	brainRoot       string
	indexBuilt      bool
	buildInProgress bool
	buildMutex      sync.RWMutex
	lastBuildTime   time.Time
	docCount        int
	collectionName  string
	manifestPath    string
	documents       map[string]*Document // Store documents for metadata access
	docChunks       map[string][]Chunk   // Chunks for text-based search fallback
}

// NewIndexer creates a new indexer instance.
func NewIndexer(brainRoot string) (*Indexer, error) {
	if brainRoot == "" {
		return nil, fmt.Errorf("brain root path cannot be empty")
	}

	return &Indexer{
		brainRoot:      brainRoot,
		collectionName: "brain_docs",
		manifestPath:   brainRoot + "/docs-manifest.json",
	}, nil
}

// EnsureIndexBuilt checks if index is built, and builds it if necessary.
// This is the lazy-load entry point - called before first search.
func (i *Indexer) EnsureIndexBuilt(ctx context.Context) error {
	i.buildMutex.Lock()

	// Check if already building or built
	if i.indexBuilt {
		i.buildMutex.Unlock()
		return nil
	}

	if i.buildInProgress {
		i.buildMutex.Unlock()
		// Wait for ongoing build
		return i.waitForBuild(ctx)
	}

	i.buildInProgress = true
	i.buildMutex.Unlock()

	// Perform the actual build
	err := i.Build(ctx)

	i.buildMutex.Lock()
	defer i.buildMutex.Unlock()

	if err != nil {
		i.buildInProgress = false
		return fmt.Errorf("failed to build index: %w", err)
	}

	i.indexBuilt = true
	i.buildInProgress = false
	i.lastBuildTime = time.Now()

	return nil
}

// Build performs a full index build from scratch.
func (i *Indexer) Build(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "Starting index build from %s\n", i.brainRoot)
	startTime := time.Now()

	// Load manifest to understand document structure
	manifest, err := LoadManifest(i.manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Load all documents from docs/ directory
	docsDir := i.brainRoot + "/docs"
	docs, loadErrors := LoadDocumentsFromDir(docsDir)

	// Log any load errors but continue (graceful degradation)
	if len(loadErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d documents failed to load:\n", len(loadErrors))
		for _, ve := range loadErrors {
			fmt.Fprintf(os.Stderr, "  - %s: %v\n", ve.Path, ve.Error)
		}
	}

	if len(docs) == 0 {
		return fmt.Errorf("no documents found in %s", docsDir)
	}

	fmt.Fprintf(os.Stderr, "Loaded %d documents\n", len(docs))

	// Validate documents against manifest (warnings only - include all docs in index)
	validDocs := 0
	validationErrors := 0
	for _, doc := range docs {
		if err := ValidateDocumentAgainstManifest(doc, manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Document %s failed validation: %v\n", doc.ID, err)
			validationErrors++
		}
		validDocs++ // Count all docs regardless of validation
	}

	fmt.Fprintf(os.Stderr, "Validated: %d documents passed, %d failed\n", validDocs-validationErrors, validationErrors)

	// Chunk ALL documents (including those with validation warnings)
	totalChunks := 0
	docChunks := make(map[string][]Chunk)
	for _, doc := range docs {
		chunks := ChunkDocument(doc)
		docChunks[doc.ID] = chunks
		totalChunks += len(chunks)
	}

	fmt.Fprintf(os.Stderr, "Chunked documents into %d chunks\n", totalChunks)

	// Store chunks and documents for text-based search (until embedding service is available)
	i.buildMutex.Lock()
	i.docChunks = docChunks
	docMap := make(map[string]*Document)
	for _, doc := range docs {
		docMap[doc.ID] = doc
	}
	i.documents = docMap
	i.buildMutex.Unlock()

	i.docCount = len(docs)

	elapsed := time.Since(startTime)
	fmt.Fprintf(os.Stderr, "Index build completed in %v (%d docs, %d chunks)\n", elapsed, len(docs), totalChunks)

	return nil
}

// Search performs a semantic search across indexed documents.
func (i *Indexer) Search(ctx context.Context, query string, limit int, domain string) ([]SearchResult, error) {
	// Ensure index is built before searching
	if err := i.EnsureIndexBuilt(ctx); err != nil {
		return nil, fmt.Errorf("index initialization failed: %w", err)
	}

	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if limit == 0 {
		limit = 10 // Default limit
	}

	// Keyword search across chunks (fallback until embedding service is available)
	if limit == 0 {
		limit = 10
	}

	i.buildMutex.RLock()
	defer i.buildMutex.RUnlock()

	var results []SearchResult
	queryLower := strings.ToLower(query)

	for docID, chunks := range i.docChunks {
		doc := i.documents[docID]
		if doc == nil {
			continue
		}

		for _, chunk := range chunks {
			// Simple keyword matching
			if strings.Contains(strings.ToLower(chunk.Content), queryLower) {
				snippet := chunk.Content
				if len(snippet) > 200 {
					snippet = snippet[:200] + "..."
				}

				results = append(results, SearchResult{
					DocumentID: docID,
					Title:      doc.Title,
					Path:       doc.Path,
					Score:      1.0, // All matches are equally weighted in text search
					RAGPriority: doc.RAGPriority,
					Snippet:    snippet,
					Category:   doc.Category,
				})

				// Stop if we have enough results
				if len(results) >= limit {
					break
				}
			}
		}

		if len(results) >= limit {
			break
		}
	}

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetStatus returns the current indexer status.
func (i *Indexer) GetStatus() IndexStatus {
	i.buildMutex.RLock()
	defer i.buildMutex.RUnlock()

	state := "ready"
	if i.buildInProgress {
		state = "building"
	} else if !i.indexBuilt {
		state = "not-initialized"
	}

	return IndexStatus{
		State:           state,
		DocumentCount:   i.docCount,
		LastRebuildTime: i.lastBuildTime.Format(time.RFC3339),
		QdrantHealth:    "unknown", // TODO: Check Qdrant health
		Errors:          []string{},
	}
}

// waitForBuild waits for an ongoing build to complete.
func (i *Indexer) waitForBuild(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("timeout waiting for index build")
		case <-ticker.C:
			i.buildMutex.RLock()
			if !i.buildInProgress {
				completed := i.indexBuilt
				i.buildMutex.RUnlock()
				if completed {
					return nil
				}
				return fmt.Errorf("index build failed")
			}
			i.buildMutex.RUnlock()
		}
	}
}
