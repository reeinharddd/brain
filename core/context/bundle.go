package context

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CompileRequest describes what context to compile
type CompileRequest struct {
	ScopeChain []string          // org:x, user:y, workspace:z
	Task       string            // Current task description
	TokenLimit int               // Maximum tokens allowed
	Model      string            // Target model (for context sizing)
	MinTier    int               // Minimum capability tier
	Layers     map[int]string    // Layer ID -> content
	Skills     []string          // Active skill IDs
	MCPs       []string          // Active MCP tool IDs
}

// AppliedOptimization describes one optimization applied
type AppliedOptimization struct {
	Type         string  // progressive_disclosure, compaction, truncation, deduplication
	LayerID      int
	LayerName    string
	TokensSaved  int
	AccuracyRisk float64 // 0.0-1.0
}

// CompiledBundle is the output of the context compiler
type CompiledBundle struct {
	BundleID      string
	CompiledAt    time.Time
	ScopeChain    []string
	Layers        []ContextLayer // Included layers (in order)
	TotalTokens   int
	TokenLimit    int
	Utilization   float64 // TotalTokens / TokenLimit
	Optimizations []AppliedOptimization
	Warnings      []string // Non-critical issues
}

// ContextCompiler compiles context bundles
type ContextCompiler struct {
	tokenLimit int
}

// NewContextCompiler creates a new context compiler with the given token limit
func NewContextCompiler(tokenLimit int) *ContextCompiler {
	return &ContextCompiler{
		tokenLimit: tokenLimit,
	}
}

// Compile compiles a context bundle from the given request
func (c *ContextCompiler) Compile(ctx context.Context, req CompileRequest) (*CompiledBundle, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// Step 1: Validate request
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	tokenLimit := req.TokenLimit
	if tokenLimit <= 0 {
		tokenLimit = c.tokenLimit
	}
	if tokenLimit <= 0 {
		return nil, errors.New("token limit must be greater than 0")
	}

	// Initialize bundle
	bundle := &CompiledBundle{
		CompiledAt:  time.Now(),
		ScopeChain:  req.ScopeChain,
		TokenLimit:  tokenLimit,
		Optimizations: make([]AppliedOptimization, 0),
		Warnings:    make([]string, 0),
	}

	layers := make([]ContextLayer, 0, 13)
	currentTokens := 0

	// Step 2: Add mandatory layers (0-1) — always included, never compressed
	for i := 0; i <= 1; i++ {
		content := req.Layers[i]
		layer := c.addLayer(i, content, &layers)
		layers = append(layers, layer)
		currentTokens += layer.TokenCount
	}

	// Track token savings from optimizations
	totalTokensSaved := 0

	// Step 3: For each remaining layer (2-12)
	for i := 2; i <= MaxLayerID; i++ {
		// Check context cancellation periodically
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled during compilation: %w", err)
		}

		content := req.Layers[i]
		if content == "" {
			continue // Skip empty layers
		}

		layer := c.addLayer(i, content, &layers)
		layerTokens := layer.TokenCount

		// Check if adding this layer would exceed the limit
		if currentTokens+layerTokens > tokenLimit && layer.Compressible {
			// Apply compaction
			budget := tokenLimit - currentTokens
			if budget < 0 {
				budget = 0
			}
			compacted := compactContent(content, budget)
			compactedTokens := len(compacted) / 4
			saved := layerTokens - compactedTokens
			if saved < 0 {
				saved = 0
			}
			layer.Content = compacted
			layer.TokenCount = compactedTokens
			layerTokens = compactedTokens

			if saved > 0 {
				bundle.Optimizations = append(bundle.Optimizations, AppliedOptimization{
					Type:         "compaction",
					LayerID:      i,
					LayerName:    layer.Name,
					TokensSaved:  saved,
					AccuracyRisk: 0.15,
				})
				totalTokensSaved += saved
			}
		}

		// Apply progressive disclosure for skills/MCP layers (7-8)
		if i == LayerActiveSkills || i == LayerActiveMCPs {
			if layer.Compressible && len(content) > 0 {
				summary := generateSummary(content)
				summaryTokens := len(summary) / 4
				saved := layerTokens - summaryTokens

				layer.Content = summary
				layer.TokenCount = summaryTokens
				layerTokens = summaryTokens

				if saved > 0 {
					bundle.Optimizations = append(bundle.Optimizations, AppliedOptimization{
						Type:         "progressive_disclosure",
						LayerID:      i,
						LayerName:    layer.Name,
						TokensSaved:  saved,
						AccuracyRisk: 0.1,
					})
					totalTokensSaved += saved
				}
			}
		}

		layers = append(layers, layer)
		currentTokens += layerTokens
	}

	bundle.Layers = layers
	bundle.TotalTokens = currentTokens

	// Step 4: Calculate utilization
	if tokenLimit > 0 {
		bundle.Utilization = float64(currentTokens) / float64(tokenLimit)
	}

	// Step 5: Generate bundle ID
	bundle.BundleID = generateBundleID(bundle)

	// Step 6: Add warnings if utilization > 90%
	if bundle.Utilization > 0.9 {
		bundle.Warnings = append(bundle.Warnings,
			fmt.Sprintf("context utilization at %.1f%%, approaching token limit", bundle.Utilization*100))
	}

	// Add warning if optimizations with accuracy risk were applied
	for _, opt := range bundle.Optimizations {
		if opt.AccuracyRisk > 0.1 {
			bundle.Warnings = append(bundle.Warnings,
				fmt.Sprintf("%s applied to layer %d (%s) with %.0f%% accuracy risk",
					opt.Type, opt.LayerID, opt.LayerName, opt.AccuracyRisk*100))
		}
	}

	return bundle, nil
}

// addLayer creates a ContextLayer from content and appends it to layers
func (c *ContextCompiler) addLayer(id int, content string, layers *[]ContextLayer) ContextLayer {
	layerDefs := LayerDefinitions()
	def := layerDefs[id]

	layer := ContextLayer{
		ID:            def.ID,
		Name:          def.Name,
		Content:       content,
		Compressible:  def.Compressible,
		AlwaysInclude: def.AlwaysInclude,
		Priority:      def.Priority,
		Tags:          def.Tags,
		Metadata:      def.Metadata,
		TokenCount:    len(content) / 4, // Rough token estimate: ~4 chars per token
	}

	if len(content) == 0 {
		layer.TokenCount = 0
	}

	return layer
}

// validateRequest validates the compile request
func validateRequest(req CompileRequest) error {
	if len(req.ScopeChain) == 0 {
		return errors.New("scope chain must not be empty")
	}
	if req.TokenLimit < 0 {
		return errors.New("token limit cannot be negative")
	}
	return nil
}

// compactContent reduces content size to fit within the token budget
func compactContent(content string, maxTokens int) string {
	if maxTokens <= 0 {
		return ""
	}

	maxChars := maxTokens * 4 // Rough estimate: 4 chars per token

	if len(content) <= maxChars {
		return content
	}

	// Truncate and add ellipsis, accounting for suffix length
	const suffix = " [...truncated]"
	suffixLen := len(suffix)
	if maxChars <= suffixLen+4 {
		// Not enough room for meaningful content + suffix
		if maxChars > 0 {
			return content[:maxChars]
		}
		return ""
	}
	return content[:maxChars-suffixLen] + suffix
}

// generateSummary creates a brief summary of content
func generateSummary(content string) string {
	lines := strings.Split(content, "\n")
	var summary strings.Builder
	lineCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if summary.Len() > 0 {
			summary.WriteString(" ")
		}
		summary.WriteString(line)
		lineCount++
		if lineCount >= 3 {
			break
		}
	}

	if lineCount > 3 {
		summary.WriteString(" [more...]")
	}

	return summary.String()
}

// generateBundleID creates a unique bundle ID using timestamp and content hash
func generateBundleID(bundle *CompiledBundle) string {
	h := sha256.New()
	h.Write([]byte(bundle.CompiledAt.Format(time.RFC3339Nano)))
	for _, layer := range bundle.Layers {
		h.Write([]byte(layer.Content))
	}
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("bundle_%s", hash[:16])
}
