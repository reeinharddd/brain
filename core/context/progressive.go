package context

import "strings"

// ProgressiveContent represents content with disclosure levels
type ProgressiveContent struct {
	LayerID  int
	Summary  string   // Always included
	Full     string   // Included on demand
	Triggers []string // Keywords that trigger full disclosure
}

// ProgressiveDisclosure manages on-demand content loading
type ProgressiveDisclosure struct {
	contents map[string]*ProgressiveContent // ID -> content
}

// NewProgressiveDisclosure creates a new progressive disclosure manager
func NewProgressiveDisclosure() *ProgressiveDisclosure {
	return &ProgressiveDisclosure{
		contents: make(map[string]*ProgressiveContent),
	}
}

// Register adds a content entry to the progressive disclosure manager
func (pd *ProgressiveDisclosure) Register(id string, content ProgressiveContent) {
	pd.contents[id] = &content
}

// GetSummary returns the summary for the given content ID
func (pd *ProgressiveDisclosure) GetSummary(id string) string {
	content, ok := pd.contents[id]
	if !ok {
		return ""
	}
	return content.Summary
}

// GetFull returns the full content for the given ID
func (pd *ProgressiveDisclosure) GetFull(id string) string {
	content, ok := pd.contents[id]
	if !ok {
		return ""
	}
	return content.Full
}

// ShouldReveal checks if the full content should be revealed based on query keywords
func (pd *ProgressiveDisclosure) ShouldReveal(id string, query string) bool {
	content, ok := pd.contents[id]
	if !ok {
		return false
	}

	if len(content.Triggers) == 0 {
		return false
	}

	queryLower := strings.ToLower(query)
	for _, trigger := range content.Triggers {
		if strings.Contains(queryLower, strings.ToLower(trigger)) {
			return true
		}
	}

	return false
}

// ContentForDisclosure creates a ProgressiveContent from the given parameters
func (pd *ProgressiveDisclosure) ContentForDisclosure(id, summary, full string, triggers []string) ProgressiveContent {
	return ProgressiveContent{
		Summary:  summary,
		Full:     full,
		Triggers: triggers,
	}
}
