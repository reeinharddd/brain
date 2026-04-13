package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/reeinharrrd/brain/daemon/internal/api/handlers"
)

// docChunk represents a section of a markdown file
type docChunk struct {
	Title       string
	Path        string
	Category    string
	RagPriority string
	Content     string
	Terms       map[string]int // term frequency
}

// DocsIndexer is a real in-memory documentation indexer
type DocsIndexer struct {
	mu       sync.RWMutex
	chunks   []docChunk
	docCount int
	builtAt  time.Time
	brainRoot string
	indexed  bool
	indexErr error
}

// NewDocsIndexer creates a new indexer pointing at the given brain root.
func NewDocsIndexer(brainRoot string) *DocsIndexer {
	return &DocsIndexer{
		brainRoot: brainRoot,
	}
}

// categoryFromPath derives a category from the docs/ subdirectory structure.
func categoryFromPath(relPath string) string {
	dir := filepath.Dir(relPath)
	if dir == "." {
		return "docs"
	}
	parts := strings.Split(dir, string(filepath.Separator))
	if len(parts) > 0 {
		return parts[0]
	}
	return "docs"
}

// ragPriorityFromCategory assigns a retrieval priority based on category.
func ragPriorityFromCategory(cat string) string {
	switch cat {
	case "architecture", "adr":
		return "1"
	case "skills":
		return "2"
	case "testing":
		return "2"
	case "templates":
		return "3"
	default:
		return "3"
	}
}

// tokenize splits text into lowercased terms (alpha-numeric only).
func tokenize(text string) []string {
	var terms []string
	var buf strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(unicode.ToLower(r))
		} else {
			if buf.Len() > 0 {
				terms = append(terms, buf.String())
				buf.Reset()
			}
		}
	}
	if buf.Len() > 0 {
		terms = append(terms, buf.String())
	}
	return terms
}

// termFrequency builds a term frequency map for a list of terms.
func termFrequency(terms []string) map[string]int {
	tf := make(map[string]int)
	for _, t := range terms {
		if len(t) < 2 {
			continue
		}
		tf[t]++
	}
	return tf
}

// stopWords is a minimal set of common English stop words.
var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
	"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "shall": true,
	"can": true, "need": true, "must": true, "it": true, "its": true, "this": true,
	"that": true, "these": true, "those": true, "i": true, "you": true, "he": true,
	"she": true, "we": true, "they": true, "what": true, "which": true, "who": true,
	"when": true, "where": true, "why": true, "how": true, "not": true, "no": true,
	"as": true, "if": true, "so": true, "than": true, "then": true, "too": true,
	"very": true, "just": true, "about": true, "up": true, "out": true, "into": true,
	"over": true, "after": true, "before": true, "between": true, "under": true,
	"again": true, "further": true, "once": true, "here": true, "there": true,
	"all": true, "each": true, "every": true, "both": true, "few": true, "more": true,
	"most": true, "other": true, "some": true, "such": true, "only": true, "own": true,
	"same": true, "also": true, "back": true, "still": true, "even": true, "any": true,
}

// filterStopWords removes stop words from a term frequency map.
func filterStopWords(tf map[string]int) map[string]int {
	result := make(map[string]int)
	for term, count := range tf {
		if !stopWords[term] {
			result[term] = count
		}
	}
	return result
}

// parseMarkdownChunks reads a markdown file and splits it into sections by headings.
func parseMarkdownChunks(path string, docsRoot string) ([]docChunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	relPath, err := filepath.Rel(docsRoot, path)
	if err != nil {
		relPath = path
	}
	category := categoryFromPath(relPath)
	ragPriority := ragPriorityFromCategory(category)

	var chunks []docChunk
	scanner := bufio.NewScanner(f)
	// Increase buffer for large files
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	var currentTitle string
	var currentContent strings.Builder
	lineNum := 0
	inFrontMatter := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Handle YAML front matter
		if lineNum == 1 && strings.TrimSpace(line) == "---" {
			inFrontMatter = true
			continue
		}
		if inFrontMatter {
			if strings.TrimSpace(line) == "---" {
				inFrontMatter = false
			}
			continue
		}

		// Detect heading lines
		if strings.HasPrefix(line, "#") {
			// Save previous chunk
			if currentTitle != "" && currentContent.Len() > 0 {
				content := strings.TrimSpace(currentContent.String())
				if content != "" {
					terms := filterStopWords(termFrequency(tokenize(content)))
					chunks = append(chunks, docChunk{
						Title:       currentTitle,
						Path:        relPath,
						Category:    category,
						RagPriority: ragPriority,
						Content:     content,
						Terms:       terms,
					})
				}
			}

			// Start new chunk
			currentTitle = strings.TrimLeft(line, "# ")
			currentContent.Reset()
		} else {
			if currentTitle == "" {
				// Content before first heading goes into a default chunk
				currentTitle = filepath.Base(relPath)
			}
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save last chunk
	if currentTitle != "" && currentContent.Len() > 0 {
		content := strings.TrimSpace(currentContent.String())
		if content != "" {
			terms := filterStopWords(termFrequency(tokenize(content)))
			chunks = append(chunks, docChunk{
				Title:       currentTitle,
				Path:        relPath,
				Category:    category,
				RagPriority: ragPriority,
				Content:     content,
				Terms:       terms,
			})
		}
	}

	return chunks, scanner.Err()
}

// buildIndex scans the docs directory and builds the search index.
func (d *DocsIndexer) buildIndex(ctx context.Context) error {
	docsRoot := filepath.Join(d.brainRoot, "docs")

	// Check if docs directory exists
	if _, err := os.Stat(docsRoot); os.IsNotExist(err) {
		d.mu.Lock()
		d.chunks = nil
		d.docCount = 0
		d.indexed = true
		d.indexErr = nil
		d.builtAt = time.Now()
		d.mu.Unlock()
		return nil
	}

	var allChunks []docChunk
	var docCount int

	err := filepath.Walk(docsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		chunks, parseErr := parseMarkdownChunks(path, docsRoot)
		if parseErr != nil {
			// Log but continue with other files
			fmt.Fprintf(os.Stderr, "[docs-indexer] warning: failed to parse %s: %v\n", path, parseErr)
			return nil
		}

		docCount++
		allChunks = append(allChunks, chunks...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking docs directory: %w", err)
	}

	d.mu.Lock()
	d.chunks = allChunks
	d.docCount = docCount
	d.indexed = true
	d.indexErr = nil
	d.builtAt = time.Now()
	d.mu.Unlock()

	return nil
}

// scoreChunk computes a relevance score for a chunk against query terms.
// Uses TF-like scoring with IDF weighting across the corpus.
func scoreChunk(chunk docChunk, queryTerms []string, docFreq map[string]int, totalDocs int) float32 {
	if len(chunk.Terms) == 0 || len(queryTerms) == 0 {
		return 0
	}

	var score float64
	matchCount := 0

	for _, qt := range queryTerms {
		tf, ok := chunk.Terms[qt]
		if !ok {
			continue
		}
		matchCount++

		// TF component: log(1 + tf) to dampen high frequencies
		tfScore := math.Log(1.0 + float64(tf))

		// IDF component: log(N / df) where N is total docs and df is doc frequency
		df := docFreq[qt]
		if df == 0 {
			df = 1
		}
		idfScore := math.Log(1.0 + float64(totalDocs)/float64(df))

		score += tfScore * idfScore
	}

	// Bonus for matching multiple query terms
	if matchCount > 1 {
		score *= 1.0 + 0.25*float64(matchCount-1)
	}

	// Bonus for title matches
	queryLower := strings.Join(queryTerms, " ")
	titleLower := strings.ToLower(chunk.Title)
	if strings.Contains(titleLower, queryLower) {
		score *= 1.5
	} else {
		for _, qt := range queryTerms {
			if strings.Contains(titleLower, qt) {
				score *= 1.2
				break
			}
		}
	}

	// Normalize to 0-1 range
	normalized := 1.0 - math.Exp(-score*0.5)

	// Apply rag priority boost (lower number = higher priority)
	ragBoost := 1.0
	switch chunk.RagPriority {
	case "1":
		ragBoost = 1.15
	case "2":
		ragBoost = 1.05
	}

	return float32(normalized * ragBoost)
}

// buildDocFreq computes document frequency for each term across all chunks.
func buildDocFreq(chunks []docChunk) map[string]int {
	df := make(map[string]int)
	for _, c := range chunks {
		seen := make(map[string]bool)
		for term := range c.Terms {
			if !seen[term] {
				df[term]++
				seen[term] = true
			}
		}
	}
	return df
}

// extractSnippet returns a relevant snippet from the chunk content matching the query.
func extractSnippet(content string, queryTerms []string, maxLen int) string {
	if len(content) == 0 {
		return ""
	}

	// Find the best matching sentence
	sentences := splitSentences(content)
	bestIdx := -1
	bestScore := 0

	for i, sent := range sentences {
		sentLower := strings.ToLower(sent)
		sentTerms := tokenize(sentLower)
		sentTermSet := make(map[string]bool)
		for _, t := range sentTerms {
			sentTermSet[t] = true
		}

		s := 0
		for _, qt := range queryTerms {
			if sentTermSet[qt] {
				s++
			}
		}
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}

	if bestIdx < 0 {
		bestIdx = 0
	}

	// Build snippet from the best sentence, with context
	start := bestIdx
	if start > 1 {
		start -= 1
	}
	end := bestIdx + 1
	if end < len(sentences) {
		end += 1
	}

	snippet := strings.Join(sentences[start:end], " ")

	// Truncate if needed
	if len(snippet) > maxLen {
		// Find a clean break point
		truncated := snippet[:maxLen]
		if idx := strings.LastIndexAny(truncated, ".,;:"); idx > maxLen/2 {
			truncated = truncated[:idx+1]
		}
		return truncated + "..."
	}

	return snippet
}

// splitSentences splits text into sentences (simple heuristic).
func splitSentences(text string) []string {
	var sentences []string
	var buf strings.Builder

	for _, r := range text {
		buf.WriteRune(r)
		if r == '.' || r == '!' || r == '?' || r == '\n' {
			s := strings.TrimSpace(buf.String())
			if len(s) > 0 {
				sentences = append(sentences, s)
			}
			buf.Reset()
		}
	}
	if buf.Len() > 0 {
		s := strings.TrimSpace(buf.String())
		if len(s) > 0 {
			sentences = append(sentences, s)
		}
	}

	return sentences
}

// Search performs a text search across the indexed documentation.
func (d *DocsIndexer) Search(ctx context.Context, query string, limit int, domain string) ([]handlers.SearchResult, int, error) {
	if limit <= 0 {
		limit = 10
	}

	// Check context
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
	}

	d.mu.RLock()
	if !d.indexed {
		d.mu.RUnlock()
		// Build index on demand if not yet indexed
		if err := d.EnsureIndexBuilt(ctx); err != nil {
			return nil, 0, err
		}
		d.mu.RLock()
	}

	chunks := make([]docChunk, len(d.chunks))
	copy(chunks, d.chunks)
	totalDocs := d.docCount
	d.mu.RUnlock()

	// Tokenize query
	queryTerms := filterStopWords(termFrequency(tokenize(strings.ToLower(query))))
	if len(queryTerms) == 0 {
		return []handlers.SearchResult{}, 0, nil
	}

	queryTermList := make([]string, 0, len(queryTerms))
	for term := range queryTerms {
		queryTermList = append(queryTermList, term)
	}

	// Build document frequency for IDF
	docFreq := buildDocFreq(chunks)

	// Filter by domain if specified
	if domain != "" {
		filtered := chunks[:0]
		for _, c := range chunks {
			if c.Category == domain {
				filtered = append(filtered, c)
			}
		}
		chunks = filtered
	}

	// Score and rank
	type scored struct {
		chunk docChunk
		score float32
	}
	var scoredChunks []scored

	for _, c := range chunks {
		s := scoreChunk(c, queryTermList, docFreq, totalDocs)
		if s > 0 {
			scoredChunks = append(scoredChunks, scored{chunk: c, score: s})
		}
	}

	// Sort by score descending
	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].score > scoredChunks[j].score
	})

	// Apply limit
	if limit < len(scoredChunks) {
		scoredChunks = scoredChunks[:limit]
	}

	// Build results
	results := make([]handlers.SearchResult, len(scoredChunks))
	for i, sc := range scoredChunks {
		results[i] = handlers.SearchResult{
			Title:       sc.chunk.Title,
			Path:        sc.chunk.Path,
			Category:    sc.chunk.Category,
			RagPriority: sc.chunk.RagPriority,
			Score:       sc.score,
			Snippet:     extractSnippet(sc.chunk.Content, queryTermList, 200),
		}
	}

	return results, len(results), nil
}

// GetStatus returns the current index health status.
func (d *DocsIndexer) GetStatus() (handlers.IndexHealth, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	state := "not_built"
	if d.indexed {
		state = "built"
		if d.indexErr != nil {
			state = "error"
		}
	}

	var errors []string
	if d.indexErr != nil {
		errors = []string{d.indexErr.Error()}
	}

	return handlers.IndexHealth{
		State:           state,
		DocumentCount:   d.docCount,
		ChunkCount:      len(d.chunks),
		LastRebuildTime: d.builtAt,
		QdrantHealth:    "unavailable",
		Errors:          errors,
	}, nil
}

// EnsureIndexBuilt builds the documentation index.
func (d *DocsIndexer) EnsureIndexBuilt(ctx context.Context) error {
	d.mu.RLock()
	if d.indexed && d.indexErr == nil {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	// Allow only one build at a time
	d.mu.Lock()
	if d.indexed && d.indexErr == nil {
		d.mu.Unlock()
		return nil
	}
	d.mu.Unlock()

	err := d.buildIndex(ctx)

	d.mu.Lock()
	d.indexErr = err
	if err == nil {
		d.indexed = true
	}
	d.mu.Unlock()

	return err
}
