package efficiency

import (
	"context"
	"strings"
)

// CompactionResult describes the result of a compaction operation
type CompactionResult struct {
	OriginalTokens  int
	CompactedTokens int
	SavingsPercent  float64
	Method          string // summarize, truncate, remove_redundant
	RiskLevel       string // none, low, medium, high
}

// Compactor handles context compaction
type Compactor struct {
	maxTokens int
	strategy  string // aggressive, moderate, minimal
}

// NewCompactor creates a new compactor with the given max tokens and strategy
func NewCompactor(maxTokens int, strategy string) *Compactor {
	return &Compactor{
		maxTokens: maxTokens,
		strategy:  strategy,
	}
}

// Compact applies compaction to the given content based on the configured strategy
func (c *Compactor) Compact(ctx context.Context, content string, currentTokens int) (*CompactionResult, string) {
	select {
	case <-ctx.Done():
		return &CompactionResult{
			OriginalTokens:  currentTokens,
			CompactedTokens: currentTokens,
			SavingsPercent:  0,
			Method:          "none",
			RiskLevel:       "none",
		}, content
	default:
	}

	if currentTokens <= c.maxTokens || len(content) == 0 {
		return &CompactionResult{
			OriginalTokens:  currentTokens,
			CompactedTokens: currentTokens,
			SavingsPercent:  0,
			Method:          "none",
			RiskLevel:       "none",
		}, content
	}

	// Determine target ratio based on strategy
	var targetRatio float64
	var riskLevel string
	var method string

	switch c.strategy {
	case "aggressive":
		targetRatio = 0.25
		riskLevel = "high"
		method = "truncate"
	case "moderate":
		targetRatio = 0.50
		riskLevel = "medium"
		method = "summarize"
	case "minimal":
		targetRatio = 0.75
		riskLevel = "low"
		method = "remove_redundant"
	default:
		targetRatio = 0.50
		riskLevel = "medium"
		method = "summarize"
	}

	targetTokens := int(float64(currentTokens) * targetRatio)

	// If under target, no compaction needed
	if targetTokens >= currentTokens {
		return &CompactionResult{
			OriginalTokens:  currentTokens,
			CompactedTokens: currentTokens,
			SavingsPercent:  0,
			Method:          "none",
			RiskLevel:       "none",
		}, content
	}

	var compacted string
	var compactedTokens int

	switch method {
	case "truncate":
		compacted = c.truncateToTokens(content, currentTokens, targetTokens)
		compactedTokens = c.EstimateTokens(compacted)
	case "summarize":
		compacted = c.Summarize(content)
		compactedTokens = c.EstimateTokens(compacted)
	case "remove_redundant":
		compacted = c.RemoveRedundant(content)
		compactedTokens = c.EstimateTokens(compacted)
	}

	// Ensure we don't exceed maxTokens
	if compactedTokens > c.maxTokens && method != "truncate" {
		compacted = c.truncateToTokens(content, currentTokens, c.maxTokens)
		compactedTokens = c.EstimateTokens(compacted)
		method = "truncate"
		riskLevel = "high"
	}

	savingsPercent := float64(currentTokens-compactedTokens) / float64(currentTokens) * 100

	return &CompactionResult{
		OriginalTokens:  currentTokens,
		CompactedTokens: compactedTokens,
		SavingsPercent:  savingsPercent,
		Method:          method,
		RiskLevel:       riskLevel,
	}, compacted
}

// EstimateTokens estimates the token count for the given text
// Uses ~1 token per 4 characters of English text as a simple approximation
func (c *Compactor) EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	// Simple approximation: 1 token per 4 characters
	// Round up for partial tokens
	tokens := len(text) / 4
	if len(text)%4 > 0 {
		tokens++
	}
	return tokens
}

// Summarize returns a truncated version of the content (placeholder summarization)
func (c *Compactor) Summarize(content string) string {
	if len(content) == 0 {
		return ""
	}

	// For summarization, keep the first and last portions
	// This simulates keeping key information from beginning and end
	charTarget := len(content) / 2

	if charTarget <= 0 {
		return content
	}

	halfTarget := charTarget / 2

	if len(content) <= charTarget {
		return content
	}

	// Keep first half of target from beginning, rest from end
	beginningLen := halfTarget
	if beginningLen > len(content) {
		beginningLen = len(content)
	}

	remainingTarget := charTarget - beginningLen
	endStart := len(content) - remainingTarget
	if endStart < beginningLen {
		endStart = beginningLen
	}

	result := content[:beginningLen]
	if endStart < len(content) {
		result += "..." + content[endStart:]
	}
	return result
}

// RemoveRedundant removes duplicate consecutive lines from the content
func (c *Compactor) RemoveRedundant(content string) string {
	if len(content) == 0 {
		return ""
	}

	lines := strings.Split(content, "\n")
	if len(lines) <= 1 {
		return content
	}

	var result []string
	seen := make(map[string]bool)
	prevLine := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			// Keep blank lines but collapse consecutive blank lines
			if prevLine == "" {
				continue
			}
			result = append(result, line)
			prevLine = trimmed
			continue
		}

		if !seen[trimmed] {
			result = append(result, line)
			seen[trimmed] = true
		}
		prevLine = trimmed
	}

	return strings.Join(result, "\n")
}

// truncateToTokens truncates content to fit within the target token count
func (c *Compactor) truncateToTokens(content string, currentTokens, targetTokens int) string {
	if targetTokens <= 0 {
		return ""
	}

	// Calculate approximate character limit
	// 1 token ≈ 4 characters
	charLimit := targetTokens * 4

	if len(content) <= charLimit {
		return content
	}

	return content[:charLimit]
}
