package runtime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestContextPrep_Tier1AggressiveCompression(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	content := strings.Repeat("A", 1000) // 1000 characters
	model := &ModelCapability{
		Tier:             Tier1Constrained,
		MaxContextTokens: 128, // ~512 chars
	}

	result := cp.PrepareContext(ctx, content, model)

	// Aggressive: 25% of 1000 = 250 chars, which is within 512 char limit
	expectedLen := 250
	if len(result) != expectedLen {
		t.Fatalf("expected length %d, got %d", expectedLen, len(result))
	}
}

func TestContextPrep_Tier2ModerateCompression(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	content := strings.Repeat("B", 1000) // 1000 characters
	model := &ModelCapability{
		Tier:             Tier2Standard,
		MaxContextTokens: 128, // ~512 chars
	}

	result := cp.PrepareContext(ctx, content, model)

	// Moderate: 50% of 1000 = 500 chars, within 512 char limit
	expectedLen := 500
	if len(result) != expectedLen {
		t.Fatalf("expected length %d, got %d", expectedLen, len(result))
	}
}

func TestContextPrep_Tier3FullContent(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	content := "This is the full content that should not be compressed."
	model := &ModelCapability{
		Tier:             Tier3Advanced,
		MaxContextTokens: 1024,
	}

	result := cp.PrepareContext(ctx, content, model)

	if result != content {
		t.Fatalf("expected full content, got %q", result)
	}
}

func TestContextPrep_RespectsMaxTokensLimit(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	t.Run("aggressive limited by maxTokens", func(t *testing.T) {
		content := strings.Repeat("C", 10000)
		model := &ModelCapability{
			Tier:             Tier1Constrained,
			MaxContextTokens: 100, // 400 chars max
		}

		result := cp.PrepareContext(ctx, content, model)

		// 25% of 10000 = 2500, but maxChars = 100 * 4 = 400, so limit is 400
		if len(result) != 400 {
			t.Fatalf("expected length 400, got %d", len(result))
		}
	})

	t.Run("moderate limited by maxTokens", func(t *testing.T) {
		content := strings.Repeat("D", 10000)
		model := &ModelCapability{
			Tier:             Tier2Standard,
			MaxContextTokens: 200, // 800 chars max
		}

		result := cp.PrepareContext(ctx, content, model)

		// 50% of 10000 = 5000, but maxChars = 200 * 4 = 800, so limit is 800
		if len(result) != 800 {
			t.Fatalf("expected length 800, got %d", len(result))
		}
	})
}

func TestContextPrep_EmptyContent(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	t.Run("tier 1 empty", func(t *testing.T) {
		model := &ModelCapability{
			Tier:             Tier1Constrained,
			MaxContextTokens: 1024,
		}

		result := cp.PrepareContext(ctx, "", model)
		if result != "" {
			t.Fatalf("expected empty string, got %q", result)
		}
	})

	t.Run("tier 2 empty", func(t *testing.T) {
		model := &ModelCapability{
			Tier:             Tier2Standard,
			MaxContextTokens: 1024,
		}

		result := cp.PrepareContext(ctx, "", model)
		if result != "" {
			t.Fatalf("expected empty string, got %q", result)
		}
	})

	t.Run("tier 3 empty", func(t *testing.T) {
		model := &ModelCapability{
			Tier:             Tier3Advanced,
			MaxContextTokens: 1024,
		}

		result := cp.PrepareContext(ctx, "", model)
		if result != "" {
			t.Fatalf("expected empty string, got %q", result)
		}
	})
}

func TestContextPrep_NoCompressionWhenContentFits(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	t.Run("tier 1 content fits", func(t *testing.T) {
		content := "Short content"
		model := &ModelCapability{
			Tier:             Tier1Constrained,
			MaxContextTokens: 1024, // plenty of room
		}

		result := cp.PrepareContext(ctx, content, model)
		// 25% of 13 chars = 3 chars (integer division)
		if len(result) != 3 {
			t.Fatalf("expected length 3 (25%% of 13), got %d", len(result))
		}
	})

	t.Run("tier 2 content fits", func(t *testing.T) {
		content := "Short content"
		model := &ModelCapability{
			Tier:             Tier2Standard,
			MaxContextTokens: 1024, // plenty of room
		}

		result := cp.PrepareContext(ctx, content, model)
		// 50% of 13 chars = 6 chars (integer division)
		if len(result) != 6 {
			t.Fatalf("expected length 6 (50%% of 13), got %d", len(result))
		}
	})
}

func TestContextPrep_NilModel(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	content := "Some content"
	result := cp.PrepareContext(ctx, content, nil)

	if result != content {
		t.Fatalf("expected original content, got %q", result)
	}
}

func TestContextPrep_ContextCancellation(t *testing.T) {
	cp := NewContextPrep()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	content := "Some content"
	model := &ModelCapability{
		Tier:             Tier1Constrained,
		MaxContextTokens: 1024,
		LatencyP50:       100 * time.Millisecond,
	}

	result := cp.PrepareContext(ctx, content, model)
	if result != "" {
		t.Fatalf("expected empty string on cancelled context, got %q", result)
	}
}

func TestContextPrep_CompressionPreservesContentStart(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	content := "IMPORTANT_START" + strings.Repeat("x", 10000)

	t.Run("aggressive keeps start", func(t *testing.T) {
		model := &ModelCapability{
			Tier:             Tier1Constrained,
			MaxContextTokens: 1024,
		}

		result := cp.PrepareContext(ctx, content, model)
		if !strings.HasPrefix(result, "IMPORTANT_START") {
			t.Fatal("aggressive compression should preserve content start")
		}
	})

	t.Run("moderate keeps start", func(t *testing.T) {
		model := &ModelCapability{
			Tier:             Tier2Standard,
			MaxContextTokens: 1024,
		}

		result := cp.PrepareContext(ctx, content, model)
		if !strings.HasPrefix(result, "IMPORTANT_START") {
			t.Fatal("moderate compression should preserve content start")
		}
	})
}

func TestContextPrep_AggressiveVsModerateRatio(t *testing.T) {
	cp := NewContextPrep()
	ctx := context.Background()

	content := strings.Repeat("Z", 1000)

	tier1Model := &ModelCapability{
		Tier:             Tier1Constrained,
		MaxContextTokens: 10000, // no maxTokens limit
	}
	tier2Model := &ModelCapability{
		Tier:             Tier2Standard,
		MaxContextTokens: 10000, // no maxTokens limit
	}

	result1 := cp.PrepareContext(ctx, content, tier1Model)
	result2 := cp.PrepareContext(ctx, content, tier2Model)

	// Tier 1: 25% = 250, Tier 2: 50% = 500
	if len(result1) != 250 {
		t.Fatalf("expected tier 1 length 250, got %d", len(result1))
	}
	if len(result2) != 500 {
		t.Fatalf("expected tier 2 length 500, got %d", len(result2))
	}
}
