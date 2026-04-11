package runtime

import (
	"context"
	"sync"
	"testing"
	"time"
)

func testModel(id string, tier CapabilityTier, local bool) ModelCapability {
	return ModelCapability{
		ModelID:          id,
		Provider:         "test-provider",
		Tier:             tier,
		MaxContextTokens: 4096,
		MaxOutputTokens:  1024,
		SupportsTools:    true,
		SupportsParallel: true,
		CostPer1KInput:   0.01,
		CostPer1KOutput:  0.02,
		LatencyP50:       100 * time.Millisecond,
		LatencyP99:       500 * time.Millisecond,
		IsLocal:          local,
	}
}

func TestModelRegistry_Register(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	model := testModel("model-1", Tier1Constrained, false)

	err := reg.Register(ctx, model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reg.Count() != 1 {
		t.Fatalf("expected count 1, got %d", reg.Count())
	}
}

func TestModelRegistry_RegisterDuplicate(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	model := testModel("model-1", Tier1Constrained, false)

	if err := reg.Register(ctx, model); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := reg.Register(ctx, model)
	if err == nil {
		t.Fatal("expected duplicate registration error")
	}
	if err != ErrModelAlreadyExists {
		t.Fatalf("expected ErrModelAlreadyExists, got %v", err)
	}
}

func TestModelRegistry_Get(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	model := testModel("model-1", Tier2Standard, false)
	if err := reg.Register(ctx, model); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := reg.Get(ctx, "model-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ModelID != "model-1" {
		t.Fatalf("expected model-1, got %s", got.ModelID)
	}
	if got.Tier != Tier2Standard {
		t.Fatalf("expected tier 2, got %d", got.Tier)
	}
}

func TestModelRegistry_GetNotFound(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	_, err := reg.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if err != ErrModelNotFound {
		t.Fatalf("expected ErrModelNotFound, got %v", err)
	}
}

func TestModelRegistry_List(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	models := []ModelCapability{
		testModel("m1", Tier1Constrained, false),
		testModel("m2", Tier2Standard, false),
		testModel("m3", Tier3Advanced, true),
	}

	for _, m := range models {
		if err := reg.Register(ctx, m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	list := reg.List(ctx)
	if len(list) != 3 {
		t.Fatalf("expected 3 models, got %d", len(list))
	}
}

func TestModelRegistry_GetByTier(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	models := []ModelCapability{
		testModel("m1", Tier1Constrained, false),
		testModel("m2", Tier1Constrained, false),
		testModel("m3", Tier2Standard, false),
		testModel("m4", Tier3Advanced, true),
	}

	for _, m := range models {
		if err := reg.Register(ctx, m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	t.Run("tier 1", func(t *testing.T) {
		tier1 := reg.GetByTier(ctx, Tier1Constrained)
		if len(tier1) != 2 {
			t.Fatalf("expected 2 tier 1 models, got %d", len(tier1))
		}
	})

	t.Run("tier 2", func(t *testing.T) {
		tier2 := reg.GetByTier(ctx, Tier2Standard)
		if len(tier2) != 1 {
			t.Fatalf("expected 1 tier 2 model, got %d", len(tier2))
		}
	})

	t.Run("tier 3", func(t *testing.T) {
		tier3 := reg.GetByTier(ctx, Tier3Advanced)
		if len(tier3) != 1 {
			t.Fatalf("expected 1 tier 3 model, got %d", len(tier3))
		}
	})
}

func TestModelRegistry_Delete(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	model := testModel("model-1", Tier1Constrained, false)
	if err := reg.Register(ctx, model); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := reg.Delete(ctx, "model-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reg.Count() != 0 {
		t.Fatalf("expected count 0, got %d", reg.Count())
	}

	_, err := reg.Get(ctx, "model-1")
	if err != ErrModelNotFound {
		t.Fatalf("expected ErrModelNotFound after delete, got %v", err)
	}
}

func TestModelRegistry_DeleteNotFound(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	err := reg.Delete(ctx, "nonexistent")
	if err != ErrModelNotFound {
		t.Fatalf("expected ErrModelNotFound, got %v", err)
	}
}

func TestModelRegistry_Count(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	if reg.Count() != 0 {
		t.Fatalf("expected count 0, got %d", reg.Count())
	}

	for i := 0; i < 5; i++ {
		m := testModel("m"+string(rune('0'+i)), Tier1Constrained, false)
		if err := reg.Register(ctx, m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reg.Count() != i+1 {
			t.Fatalf("expected count %d, got %d", i+1, reg.Count())
		}
	}
}

func TestModelRegistry_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)

		go func(n int) {
			defer wg.Done()
			m := testModel("concurrent-"+string(rune('A'+n%26)), Tier1Constrained, false)
			_ = reg.Register(ctx, m) // may fail with duplicate, that's ok
		}(i)

		go func(n int) {
			defer wg.Done()
			_, _ = reg.Get(ctx, "concurrent-"+string(rune('A'+n%26)))
		}(i)

		go func(n int) {
			defer wg.Done()
			_ = reg.List(ctx)
		}(i)
	}

	wg.Wait()
	// Just ensure no panics or data races occurred
	t.Logf("concurrent access completed, count=%d", reg.Count())
}

func TestModelRegistry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reg := NewModelRegistry()
	model := testModel("model-1", Tier1Constrained, false)

	err := reg.Register(ctx, model)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}

	_, err = reg.Get(ctx, "model-1")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}

	_ = reg.List(ctx)     // should return nil
	_ = reg.GetByTier(ctx, Tier1Constrained) // should return nil

	err = reg.Delete(ctx, "model-1")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
