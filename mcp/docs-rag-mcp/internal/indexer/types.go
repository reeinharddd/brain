// Package indexer provides document indexing and searching capabilities.
package indexer

import (
	"fmt"
	"time"
)

// Document represents a single markdown document with metadata.
type Document struct {
	ID              string   `yaml:"id"`
	Type            string   `yaml:"type"`
	Title           string   `yaml:"title"`
	Status          string   `yaml:"status"`
	Version         string   `yaml:"version"`
	DateCreated     string   `yaml:"date_created"`
	Language        string   `yaml:"language"`
	Category        string   `yaml:"category"`
	RelatedTo       []string `yaml:"related_to,omitempty"`
	Keywords        []string `yaml:"keywords,omitempty"`
	RAGPriority     string   `yaml:"rag_priority,omitempty"` // critical, high, medium, low
	ChunkStrategy   string   `yaml:"chunk_strategy,omitempty"` // section, sentence
	Path            string   // File system path
	Body            string   // Markdown content (without frontmatter)
	LastModified    time.Time
}

// ManifestDomain defines rules for a documentation domain.
type ManifestDomain struct {
	Purpose                string                 `json:"purpose"`
	Required               bool                   `json:"required"`
	Subdirs                []string               `json:"subdirs,omitempty"`
	Files                  []string               `json:"files,omitempty"`
	Template               string                 `json:"template,omitempty"`
	TemplateLocation       string                 `json:"template_location,omitempty"`
	ValidationLevel        string                 `json:"validation_level,omitempty"`
	Rules                  map[string]interface{} `json:"rules,omitempty"`
}

// Manifest represents the documentation structure contract.
type Manifest struct {
	Version      string                    `json:"version"`
	LastUpdated  string                    `json:"last_updated,omitempty"`
	Purpose      string                    `json:"purpose,omitempty"`
	RootFiles    map[string]interface{}    `json:"root_files,omitempty"`
	Domains      map[string]ManifestDomain `json:"domains"`
	GlobalRules  GlobalValidationRules     `json:"global_rules,omitempty"`
}

// GlobalValidationRules defines validation rules applied globally.
type GlobalValidationRules struct {
	RequiredFrontmatterFields []string `json:"required_frontmatter_fields,omitempty"`
	EnglishOnly               bool     `json:"english_only,omitempty"`
	NoEmojis                  bool     `json:"no_emojis,omitempty"`
	FilenameFormat            string   `json:"filename_format,omitempty"`
}

// SearchResult represents a search result with ranking.
type SearchResult struct {
	DocumentID   string  `json:"document_id"`
	Title        string  `json:"title"`
	Path         string  `json:"path"`
	Score        float32 `json:"score"`
	RAGPriority  string  `json:"rag_priority"`
	Snippet      string  `json:"snippet"` // 200 char excerpt with context
	Category     string  `json:"category"`
	RelatedCount int     `json:"related_count"`
}

// IndexStatus represents the current state of the index.
type IndexStatus struct {
	State           string    `json:"state"` // building, ready, error
	DocumentCount   int       `json:"document_count"`
	LastRebuildTime string    `json:"last_rebuild_time"`
	IndexSizeBytes  int64     `json:"index_size_bytes"`
	QdrantHealth    string    `json:"qdrant_health"` // healthy, unhealthy
	Errors          []string  `json:"errors"`
}

// ValidationError represents a document validation failure.
type ValidationError struct {
	Path    string
	Field   string
	Message string
	Error   error
}

// PriorityScore maps RAG priority to numeric score for ranking.
func PriorityScore(priority string) float32 {
	switch priority {
	case "critical":
		return 1.5
	case "high":
		return 1.2
	case "medium":
		return 1.0
	case "low":
		return 0.8
	default:
		return 1.0
	}
}

// ValidateFrontmatter checks that a document has all required fields.
func (d *Document) ValidateFrontmatter(requiredFields []string) error {
	if d.ID == "" {
		return fmt.Errorf("missing required field: id")
	}
	if d.Type == "" {
		return fmt.Errorf("missing required field: type")
	}
	if d.Title == "" {
		return fmt.Errorf("missing required field: title")
	}
	if d.Status == "" {
		return fmt.Errorf("missing required field: status")
	}
	if d.DateCreated == "" {
		return fmt.Errorf("missing required field: date_created")
	}
	if d.Language != "en" {
		return fmt.Errorf("language must be 'en', found: %s", d.Language)
	}
	if d.Category == "" {
		return fmt.Errorf("missing required field: category")
	}
	return nil
}

// SetDefaults sets default values for optional fields.
func (d *Document) SetDefaults() {
	if d.RAGPriority == "" {
		d.RAGPriority = "medium"
	}
	if d.ChunkStrategy == "" {
		d.ChunkStrategy = "section"
	}
	if d.Language == "" {
		d.Language = "en"
	}
}
