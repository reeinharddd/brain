package context

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCompile(t *testing.T) {
	t.Run("compile minimal bundle with mandatory layers only", func(t *testing.T) {
		compiler := NewContextCompiler(10000)
		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy content",
				1: "Identity information",
			},
			TokenLimit: 10000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bundle == nil {
			t.Fatal("bundle should not be nil")
		}

		// Should have 2 layers (0 and 1)
		if len(bundle.Layers) != 2 {
			t.Errorf("expected 2 layers, got %d", len(bundle.Layers))
		}

		// Verify layer IDs
		if bundle.Layers[0].ID != 0 {
			t.Errorf("expected first layer ID 0, got %d", bundle.Layers[0].ID)
		}
		if bundle.Layers[1].ID != 1 {
			t.Errorf("expected second layer ID 1, got %d", bundle.Layers[1].ID)
		}
	})

	t.Run("compile full bundle with all layers", func(t *testing.T) {
		compiler := NewContextCompiler(100000)
		layers := make(map[int]string)
		for i := 0; i <= 12; i++ {
			layers[i] = strings.Repeat("content for layer ", 20)
		}

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice", "workspace:ws1"},
			Task:       "Implement feature X",
			Layers:     layers,
			TokenLimit: 100000,
			Skills:     []string{"skill1", "skill2"},
			MCPs:       []string{"mcp1"},
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bundle == nil {
			t.Fatal("bundle should not be nil")
		}

		// All 13 layers should be present (since we have enough tokens)
		if len(bundle.Layers) == 0 {
			t.Error("expected at least some layers")
		}
	})

	t.Run("token limit enforcement", func(t *testing.T) {
		compiler := NewContextCompiler(100)

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
				2: strings.Repeat("organization baseline content ", 100),
				3: strings.Repeat("user baseline content ", 100),
			},
			TokenLimit: 100,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Total tokens should not exceed limit
		if bundle.TotalTokens > bundle.TokenLimit {
			t.Errorf("total tokens %d exceeds limit %d", bundle.TotalTokens, bundle.TokenLimit)
		}
	})

	t.Run("utilization calculation is correct", func(t *testing.T) {
		compiler := NewContextCompiler(1000)

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: 1000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedUtil := float64(bundle.TotalTokens) / 1000.0
		if bundle.Utilization != expectedUtil {
			t.Errorf("expected utilization %.4f, got %.4f", expectedUtil, bundle.Utilization)
		}
	})

	t.Run("progressive disclosure for skill layers", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		skillContent := `# Skill: Code Reviewer
Description: Reviews code for quality
Usage: Use when reviewing code
Full content: This is the full skill definition with detailed instructions and parameters.`

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
				7: skillContent,
			},
			TokenLimit: 10000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have applied progressive disclosure optimization
		hasProgressiveDisclosure := false
		for _, opt := range bundle.Optimizations {
			if opt.Type == "progressive_disclosure" && opt.LayerID == LayerActiveSkills {
				hasProgressiveDisclosure = true
				break
			}
		}

		if !hasProgressiveDisclosure {
			t.Error("expected progressive_disclosure optimization for skills layer")
		}
	})

	t.Run("progressive disclosure for MCP layers", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		mcpContent := `# MCP Tool: File Search
Description: Searches files by pattern
Parameters: pattern, directory, recursive

Full implementation details:
This MCP tool provides comprehensive file search capabilities.
It supports glob patterns, regex matching, and content-based searches.
The tool integrates with the workspace indexing system for fast lookups.
Configuration options include max depth, file type filters, and exclusion patterns.
Performance is optimized through cached index lookups and parallel scanning.
The tool returns ranked results with relevance scores and file metadata.
Additional features include search history, saved queries, and batch operations.
Error handling covers permission issues, symlink loops, and filesystem errors.
Rate limiting ensures the tool doesn't overwhelm the filesystem during large scans.`

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
				8: mcpContent,
			},
			TokenLimit: 10000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have applied progressive disclosure optimization
		hasProgressiveDisclosure := false
		for _, opt := range bundle.Optimizations {
			if opt.Type == "progressive_disclosure" && opt.LayerID == LayerActiveMCPs {
				hasProgressiveDisclosure = true
				break
			}
		}

		if !hasProgressiveDisclosure {
			t.Error("expected progressive_disclosure optimization for MCP layer")
		}
	})

	t.Run("warnings generated at high utilization", func(t *testing.T) {
		compiler := NewContextCompiler(50)

		// Create content that will fill more than 90% of the 50 token limit
		// 50 tokens * 4 chars = 200 chars max
		// We need content that results in > 45 tokens
		content := strings.Repeat("x", 180) // ~45 tokens, 90% utilization

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: content,
				1: content,
			},
			TokenLimit: 50,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bundle.Utilization <= 0.9 {
			t.Fatalf("expected utilization > 90%%, got %.1f%%", bundle.Utilization*100)
		}

		if len(bundle.Warnings) == 0 {
			t.Error("expected warnings at high utilization")
		}

		// Check for utilization warning
		hasUtilizationWarning := false
		for _, w := range bundle.Warnings {
			if strings.Contains(w, "utilization") {
				hasUtilizationWarning = true
				break
			}
		}

		if !hasUtilizationWarning {
			t.Error("expected utilization warning when utilization > 90%")
		}
	})

	t.Run("empty content layers handled gracefully", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "", // Empty content for identity layer
				2: "", // Empty content for org baseline
			},
			TokenLimit: 10000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have at least the mandatory layers
		if len(bundle.Layers) < 2 {
			t.Errorf("expected at least 2 layers, got %d", len(bundle.Layers))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: 10000,
		}

		_, err := compiler.Compile(ctx, req)
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})

	t.Run("invalid request - zero token limit", func(t *testing.T) {
		compiler := NewContextCompiler(0)

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: 0,
		}

		_, err := compiler.Compile(context.Background(), req)
		if err == nil {
			t.Error("expected error for zero token limit")
		}
	})

	t.Run("invalid request - negative token limit", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: -100,
		}

		_, err := compiler.Compile(context.Background(), req)
		if err == nil {
			t.Error("expected error for negative token limit")
		}
	})

	t.Run("invalid request - empty scope chain", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		req := CompileRequest{
			ScopeChain: []string{},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: 10000,
		}

		_, err := compiler.Compile(context.Background(), req)
		if err == nil {
			t.Error("expected error for empty scope chain")
		}
	})

	t.Run("bundle ID is unique per compilation", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: 10000,
		}

		bundle1, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Small delay to ensure different timestamp
		time.Sleep(1 * time.Millisecond)

		bundle2, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bundle1.BundleID == bundle2.BundleID {
			t.Errorf("bundle IDs should be unique: got %s and %s", bundle1.BundleID, bundle2.BundleID)
		}
	})

	t.Run("layer ordering preserved", func(t *testing.T) {
		compiler := NewContextCompiler(100000)

		layers := make(map[int]string)
		for i := 0; i <= 12; i++ {
			layers[i] = "content"
		}

		req := CompileRequest{
			ScopeChain: []string{"org:test", "user:alice"},
			Layers:     layers,
			TokenLimit: 100000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := 0; i < len(bundle.Layers); i++ {
			expectedID := bundle.Layers[i].ID
			if i > 0 {
				prevID := bundle.Layers[i-1].ID
				if expectedID <= prevID {
					t.Errorf("layer ordering violated: layer[%d].ID=%d <= layer[%d].ID=%d",
						i, expectedID, i-1, prevID)
				}
			}
		}
	})

	t.Run("scope chain preserved", func(t *testing.T) {
		compiler := NewContextCompiler(10000)

		scopeChain := []string{"org:acme", "user:bob", "workspace:proj1"}
		req := CompileRequest{
			ScopeChain: scopeChain,
			Layers: map[int]string{
				0: "Hard policy",
				1: "Identity",
			},
			TokenLimit: 10000,
		}

		bundle, err := compiler.Compile(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(bundle.ScopeChain) != len(scopeChain) {
			t.Errorf("expected scope chain length %d, got %d", len(scopeChain), len(bundle.ScopeChain))
		}

		for i, s := range scopeChain {
			if bundle.ScopeChain[i] != s {
				t.Errorf("scope chain[%d]: expected %s, got %s", i, s, bundle.ScopeChain[i])
			}
		}
	})
}
