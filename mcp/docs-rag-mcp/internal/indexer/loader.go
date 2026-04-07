// Package indexer provides document indexing and searching capabilities.
package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadDocument loads a markdown file and parses its YAML frontmatter and body.
func LoadDocument(path string) (*Document, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Read file contents
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	contentStr := string(content)

	// Parse frontmatter and body
	fm, body, err := parseFrontmatter(contentStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", path, err)
	}

	// Create document from frontmatter
	doc := &Document{
		Path: path,
		Body: body,
	}

	// Unmarshal YAML frontmatter into document
	if err := yaml.Unmarshal(fm, doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML frontmatter in %s: %w", path, err)
	}

	// Set defaults and get file stats
	doc.SetDefaults()

	fileInfo, err := os.Stat(path)
	if err == nil {
		doc.LastModified = fileInfo.ModTime()
	}

	// Validate required fields - but if missing, generate them
	requiredFields := []string{"id", "type", "title", "status", "date_created", "language", "category"}
	if err := doc.ValidateFrontmatter(requiredFields); err != nil {
		// Generate missing fields instead of failing
		if doc.ID == "" {
			// Use filename as ID
			doc.ID = strings.TrimSuffix(filepath.Base(path), ".md")
		}
		if doc.Title == "" {
			doc.Title = doc.ID
		}
		if doc.Type == "" {
			doc.Type = "guide"
		}
		if doc.Status == "" {
			doc.Status = "active"
		}
		if doc.DateCreated == "" {
			doc.DateCreated = time.Now().Format("2006-01-02T15:04:05Z07:00")
		}
		if doc.Language == "" {
			doc.Language = "en"
		}
		if doc.Category == "" {
			doc.Category = "documentation"
		}
	}

	return doc, nil
}

// parseFrontmatter extracts YAML frontmatter and body from markdown content.
// Expected format:
//
//	---
//	key: value
//	---
//	# Body content
func parseFrontmatter(content string) ([]byte, string, error) {
	content = strings.TrimSpace(content)

	// Must start with ---
	if !strings.HasPrefix(content, "---") {
		return nil, "", fmt.Errorf("markdown must start with YAML frontmatter (---)")
	}

	// Find closing ---
	remainder := content[3:] // Skip opening ---
	closingIdx := strings.Index(remainder, "---")
	if closingIdx == -1 {
		return nil, "", fmt.Errorf("frontmatter closing delimiter (---) not found")
	}

	// Extract frontmatter and body
	frontmatter := strings.TrimSpace(remainder[:closingIdx])
	body := strings.TrimSpace(remainder[closingIdx+3:])

	return []byte(frontmatter), body, nil
}

// LoadDocumentsFromDir recursively loads all markdown files from a directory.
func LoadDocumentsFromDir(dir string) ([]*Document, []ValidationError) {
	var docs []*Document
	var errors []ValidationError

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Load document
		doc, err := LoadDocument(path)
		if err != nil {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "failed to load document",
				Error:   err,
			})
			return nil // Continue with other files
		}

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		errors = append(errors, ValidationError{
			Path:    dir,
			Message: "failed to walk directory",
			Error:   err,
		})
	}

	return docs, errors
}

// LoadManifest loads the documentation manifest from docs-manifest.json.
func LoadManifest(path string) (*Manifest, error) {
	if path == "" {
		return nil, fmt.Errorf("manifest path cannot be empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file %s: %w", path, err)
	}

	manifest := &Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// Set defaults
	if manifest.GlobalRules.EnglishOnly {
		manifest.GlobalRules.EnglishOnly = true
	}
	if manifest.GlobalRules.NoEmojis {
		manifest.GlobalRules.NoEmojis = true
	}

	return manifest, nil
}

// ValidateDocumentAgainstManifest checks if a document conforms to manifest rules.
func ValidateDocumentAgainstManifest(doc *Document, manifest *Manifest) error {
	// Check category exists in manifest
	if _, exists := manifest.Domains[doc.Category]; !exists {
		return fmt.Errorf("document category '%s' not found in manifest", doc.Category)
	}

	// Check that status is valid
	validStatuses := map[string]bool{"active": true, "draft": true, "review": true, "deprecated": true, "archived": true}
	if !validStatuses[doc.Status] {
		return fmt.Errorf("invalid status '%s' (must be: active, draft, review, deprecated, archived)", doc.Status)
	}

	// Check English-only rule
	if manifest.GlobalRules.EnglishOnly {
		if containsNonEnglish(doc.Body) {
			return fmt.Errorf("document contains non-English content")
		}
	}

	// Check no-emojis rule
	if manifest.GlobalRules.NoEmojis {
		if containsEmojis(doc.Body) {
			return fmt.Errorf("document contains emojis")
		}
	}

	return nil
}

// containsNonEnglish checks for common Spanish characters and words.
func containsNonEnglish(text string) bool {
	// Check for Spanish accented characters
	spanishChars := []string{"á", "é", "í", "ó", "ú", "ñ"}
	for _, char := range spanishChars {
		if strings.Contains(text, char) {
			return true
		}
	}

	// Check for Spanish punctuation
	if strings.Contains(text, "¿") || strings.Contains(text, "¡") {
		return true
	}

	return false
}

// containsEmojis checks for common emojis and emoji patterns.
func containsEmojis(text string) bool {
	// Check for emoji colon notation like :rocket:
	if strings.Contains(text, ":") {
		parts := strings.Split(text, ":")
		if len(parts) >= 3 {
			for i := 0; i < len(parts)-1; i++ {
				potentialEmoji := parts[i+1]
				emojiList := []string{"rocket", "check", "fire", "smile", "tada", "star", "heart", "x", "ok"}
				for _, emoji := range emojiList {
					if potentialEmoji == emoji {
						return true
					}
				}
			}
		}
	}

	// Check for common Unicode emoji ranges
	emojiPatterns := []string{
		"\U0001F300", // 🌀 and similar
		"\U0001F600", // 😀and similar
		"✓",          // checkmark
		"✗",          // cross
		"×",          // X mark
	}

	for _, pattern := range emojiPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}

	return false
}

// GetPathRelativeToRoot returns the path relative to the brain root.
func GetPathRelativeToRoot(root, fullPath string) string {
	if rel, err := filepath.Rel(root, fullPath); err == nil {
		return rel
	}
	return fullPath
}
