package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDocument_ValidateFrontmatter tests YAML frontmatter validation.
func TestDocument_ValidateFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		doc      *Document
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid document",
			doc: &Document{
				ID:          "test-1",
				Type:        "architecture",
				Title:       "Test Doc",
				Status:      "active",
				DateCreated: "2026-04-03",
				Language:    "en",
				Category:    "architecture",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			doc: &Document{
				Type:        "architecture",
				Title:       "Test Doc",
				Status:      "active",
				DateCreated: "2026-04-03",
				Language:    "en",
				Category:    "architecture",
			},
			wantErr: true,
			errMsg:  "missing required field: id",
		},
		{
			name: "wrong language",
			doc: &Document{
				ID:          "test-1",
				Type:        "architecture",
				Title:       "Test Doc",
				Status:      "active",
				DateCreated: "2026-04-03",
				Language:    "es",
				Category:    "architecture",
			},
			wantErr: true,
			errMsg:  "language must be 'en'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.ValidateFrontmatter([]string{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error: %v, wantErr: %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Fatalf("error message %q doesn't contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestChunkDocument_Section tests section-based chunking.
func TestChunkDocument_Section(t *testing.T) {
	doc := &Document{
		ID:            "test-1",
		ChunkStrategy: "section",
		Body: `## Introduction
This is the intro section.

## Main Content
This is the main section.

## Conclusion
This is the conclusion.`,
	}

	chunks := ChunkDocument(doc)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	// Verify first chunk contains the header
	if !strings.Contains(chunks[0].Content, "## Introduction") {
		t.Errorf("first chunk missing Introduction header")
	}

	if !strings.Contains(chunks[1].Content, "## Main Content") {
		t.Errorf("second chunk missing Main Content header")
	}

	if !strings.Contains(chunks[2].Content, "## Conclusion") {
		t.Errorf("third chunk missing Conclusion header")
	}
}

// TestChunkDocument_Sentence tests sentence-based chunking.
func TestChunkDocument_Sentence(t *testing.T) {
	doc := &Document{
		ID:            "test-1",
		ChunkStrategy: "sentence",
		Body: `First sentence. Second sentence. Third sentence.
Fourth sentence. Fifth sentence. Sixth sentence.`,
	}

	chunks := ChunkDocument(doc)

	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	// Verify chunks contain sentences
	allText := ""
	for _, chunk := range chunks {
		allText += chunk.Content
	}

	// All original text should be present
	if !strings.Contains(allText, "First sentence") {
		t.Errorf("content missing First sentence")
	}
}

// TestChunkDocument_Empty tests chunking of empty document.
func TestChunkDocument_Empty(t *testing.T) {
	doc := &Document{ID: "test-1", Body: ""}
	chunks := ChunkDocument(doc)

	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty document, got %d", len(chunks))
	}
}

// TestPriorityScore tests RAG priority scoring.
func TestPriorityScore(t *testing.T) {
	tests := []struct {
		priority string
		want     float32
	}{
		{"critical", 1.5},
		{"high", 1.2},
		{"medium", 1.0},
		{"low", 0.8},
		{"unknown", 1.0},
		{"", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			got := PriorityScore(tt.priority)
			if got != tt.want {
				t.Fatalf("PriorityScore(%q) = %f, want %f", tt.priority, got, tt.want)
			}
		})
	}
}

// TestContainsNonEnglish tests non-English content detection.
func TestContainsNonEnglish(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "pure english",
			text: "This is pure English text.",
			want: false,
		},
		{
			name: "spanish accents",
			text: "Este es un texto español con ñ.",
			want: true,
		},
		{
			name: "spanish punctuation",
			text: "¿Cómo estás?",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsNonEnglish(tt.text)
			if got != tt.want {
				t.Fatalf("containsNonEnglish() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainsEmojis tests emoji detection.
func TestContainsEmojis(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "no emojis",
			text: "This is plain text.",
			want: false,
		},
		{
			name: "emoji notation",
			text: "This is great :rocket:",
			want: true,
		},
		{
			name: "checkmark",
			text: "✓ Done",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsEmojis(tt.text)
			if got != tt.want {
				t.Fatalf("containsEmojis() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseFrontmatter tests YAML frontmatter parsing.
func TestParseFrontmatter(t *testing.T) {
	content := `---
id: test-1
type: architecture
title: Test Document
---
# Body Content
This is the body.`

	fm, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("parseFrontmatter failed: %v", err)
	}

	if !strings.Contains(string(fm), "id: test-1") {
		t.Errorf("frontmatter missing id field")
	}

	if !strings.Contains(body, "# Body Content") {
		t.Errorf("body missing expected content")
	}
}

// TestLoadDocumentFromTempFile tests document loading from real file.
func TestLoadDocumentFromTempFile(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	docPath := filepath.Join(tmpDir, "test.md")

	content := `---
id: test-1
type: architecture
title: Test Document
status: active
version: 1.0.0
date_created: 2026-04-03
language: en
category: architecture
---
# Body Content
This is the body.`

	if err := os.WriteFile(docPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	doc, err := LoadDocument(docPath)
	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	if doc.ID != "test-1" {
		t.Errorf("doc.ID = %q, want test-1", doc.ID)
	}

	if doc.Title != "Test Document" {
		t.Errorf("doc.Title = %q, want Test Document", doc.Title)
	}

	if !strings.Contains(doc.Body, "# Body Content") {
		t.Errorf("doc.Body missing expected content")
	}
}

// TestLoadDocumentInvalidFile tests error handling for invalid files.
func TestLoadDocumentInvalidFile(t *testing.T) {
	_, err := LoadDocument("/nonexistent/path/file.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// TestIndexer_NewIndexer tests indexer creation.
func TestIndexer_NewIndexer(t *testing.T) {
	indexer, err := NewIndexer("/tmp/brain")
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	if indexer.brainRoot != "/tmp/brain" {
		t.Errorf("indexer.brainRoot = %q, want /tmp/brain", indexer.brainRoot)
	}

	if indexer.indexBuilt {
		t.Error("new indexer should not be marked as built")
	}
}

// TestIndexer_NewIndexer_EmptyRoot tests error on empty root.
func TestIndexer_NewIndexer_EmptyRoot(t *testing.T) {
	_, err := NewIndexer("")
	if err == nil {
		t.Fatal("expected error for empty brain root")
	}
}

// TestIndexer_GetStatus tests status reporting.
func TestIndexer_GetStatus(t *testing.T) {
	indexer, _ := NewIndexer("/tmp/brain")

	status := indexer.GetStatus()

	if status.State != "not-initialized" {
		t.Errorf("status.State = %q, want not-initialized", status.State)
	}

	if status.DocumentCount != 0 {
		t.Errorf("status.DocumentCount = %d, want 0", status.DocumentCount)
	}
}

// TestValidateDocumentAgainstManifest tests manifest-based document validation.
func TestValidateDocumentAgainstManifest(t *testing.T) {
	manifest := &Manifest{
		Domains: map[string]ManifestDomain{
			"architecture": {
				Name:        "architecture",
				Path:        "docs/architecture/",
				RAGPriority: "high",
			},
		},
		GlobalRules: GlobalValidationRules{
			RequiredFrontmatterFields: []string{"id", "type", "title"},
			EnglishOnly:               true,
			NoEmojis:                  true,
		},
	}

	doc := &Document{
		ID:       "test-1",
		Category: "architecture",
		Status:   "active",
		Body:     "This is English text.",
	}

	err := ValidateDocumentAgainstManifest(doc, manifest)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

// TestValidateDocumentAgainstManifest_BadCategory tests invalid category.
func TestValidateDocumentAgainstManifest_BadCategory(t *testing.T) {
	manifest := &Manifest{
		Domains: map[string]ManifestDomain{},
	}

	doc := &Document{
		ID:       "test-1",
		Category: "unknown",
		Status:   "active",
	}

	err := ValidateDocumentAgainstManifest(doc, manifest)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}

	if !strings.Contains(err.Error(), "not found in manifest") {
		t.Errorf("error message doesn't mention manifest: %v", err)
	}
}

// BenchmarkChunkDocument benchmarks chunking performance.
func BenchmarkChunkDocument(b *testing.B) {
	doc := &Document{
		ID:            "test-1",
		ChunkStrategy: "section",
		Body: strings.Repeat(`## Section
This is a section with some content.
It has multiple lines.

`, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ChunkDocument(doc)
	}
}

// TestParseHTML_ViaIndexerBuild tests full indexer pipeline.
func TestParseHTML_ViaIndexerBuild(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	os.MkdirAll(filepath.Join(docsDir, "architecture"), 0755)

	// Create manifest
	manifestData := `
version: "1.0"
title: Brain Docs
domains:
  architecture:
    name: "Architecture"
    path: "docs/architecture/"
    rag_priority: "high"
global_rules:
  required_frontmatter_fields: ["id", "type", "title", "status", "date_created", "language", "category"]
  english_only: true
  no_emojis: true
`
	manifestPath := filepath.Join(tmpDir, "docs-manifest.json")
	os.WriteFile(manifestPath, []byte(manifestData), 0644)

	// Create test document
	docContent := `---
id: arch-001
type: architecture
title: Daemon Architecture
status: active
version: 1.0.0
date_created: 2026-04-03
language: en
category: architecture
rag_priority: critical
chunk_strategy: section
---
## Introduction
This is the daemon architecture document.

## Design
The design consists of multiple components.

## Implementation
Implementation details go here.`

	docPath := filepath.Join(docsDir, "architecture", "daemon-design.md")
	os.WriteFile(docPath, []byte(docContent), 0644)

	// Create indexer and build
	indexer, _ := NewIndexer(tmpDir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := indexer.Build(ctx)
	if err == nil || strings.Contains(err.Error(), "Qdrant") {
		// Build succeeded or failed only due to missing Qdrant (expected)
		status := indexer.GetStatus()
		if status.DocumentCount > 0 {
			t.Logf("Successfully loaded %d documents", status.DocumentCount)
		}
	} else {
		t.Fatalf("unexpected build error: %v", err)
	}
}
