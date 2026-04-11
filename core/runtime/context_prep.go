package runtime

import (
	"context"
)

// ContextPrep prepares context for a given model tier
type ContextPrep struct{}

// NewContextPrep creates a new ContextPrep
func NewContextPrep() *ContextPrep {
	return &ContextPrep{}
}

// PrepareContext adjusts context based on model tier
func (cp *ContextPrep) PrepareContext(ctx context.Context, content string, model *ModelCapability) string {
	select {
	case <-ctx.Done():
		return ""
	default:
	}

	if model == nil {
		return content
	}

	switch model.Tier {
	case Tier1Constrained:
		return cp.aggressiveCompress(content, model.MaxContextTokens)
	case Tier2Standard:
		return cp.moderateCompress(content, model.MaxContextTokens)
	case Tier3Advanced:
		return content // No compression
	default:
		return content
	}
}

// aggressiveCompress keeps first 25% of content or up to maxTokens
func (cp *ContextPrep) aggressiveCompress(content string, maxTokens int) string {
	if content == "" {
		return ""
	}

	// Token estimation: ~1 token per 4 characters
	maxChars := maxTokens * 4
	contentLen := len(content)

	// Keep first 25% of content
	limit := contentLen / 4
	if limit > maxChars {
		limit = maxChars
	}

	if limit >= contentLen {
		return content
	}

	return content[:limit]
}

// moderateCompress keeps first 50% of content or up to maxTokens
func (cp *ContextPrep) moderateCompress(content string, maxTokens int) string {
	if content == "" {
		return ""
	}

	// Token estimation: ~1 token per 4 characters
	maxChars := maxTokens * 4
	contentLen := len(content)

	// Keep first 50% of content
	limit := contentLen / 2
	if limit > maxChars {
		limit = maxChars
	}

	if limit >= contentLen {
		return content
	}

	return content[:limit]
}
