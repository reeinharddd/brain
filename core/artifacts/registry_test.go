package artifacts

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func newTestKey(kind, id string) ArtifactKey {
	return ArtifactKey{Kind: kind, ID: id}
}

func newTestEnvelope(kind, id, version string) ArtifactEnvelope {
	return ArtifactEnvelope{
		ID:      id,
		Kind:    kind,
		Version: version,
		Name:    fmt.Sprintf("%s-%s", kind, id),
		Scope:   "org",
		State:   "active",
	}
}

func registerTestArtifact(t *testing.T, r *ArtifactRegistry, kind, id, version string) {
	t.Helper()
	ctx := context.Background()
	key := newTestKey(kind, id)
	env := newTestEnvelope(kind, id, version)
	if err := r.Register(ctx, key, env); err != nil {
		t.Fatalf("Register(%s:%s): %v", kind, id, err)
	}
}

// --- Register and Get ---

func TestRegistry_RegisterAndGet(t *testing.T) {
	ctx := context.Background()

	t.Run("register and retrieve", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("skill", "test-skill")
		env := ArtifactEnvelope{
			ID:          "test-skill",
			Kind:        "skill",
			Version:     "1.0.0",
			Name:        "Test Skill",
			Description: "A test skill",
			Scope:       "org",
			Visibility:  "public",
		}

		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("Register: %v", err)
		}

		record, err := r.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if record.Envelope.ID != "test-skill" {
			t.Errorf("expected ID test-skill, got %s", record.Envelope.ID)
		}
		if record.Envelope.Version != "1.0.0" {
			t.Errorf("expected version 1.0.0, got %s", record.Envelope.Version)
		}
		if record.Envelope.State != "active" {
			t.Errorf("expected state active, got %s", record.Envelope.State)
		}
		if record.Envelope.Checksum == "" {
			t.Error("expected non-empty checksum")
		}
	})

	t.Run("duplicate registration", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("skill", "dup")
		env := newTestEnvelope("skill", "dup", "1.0.0")

		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("first Register: %v", err)
		}
		if err := r.Register(ctx, key, env); err == nil {
			t.Error("expected error on duplicate registration")
		}
	})

	t.Run("auto-generated fields", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("mcp", "auto")
		env := ArtifactEnvelope{
			ID:   "auto",
			Kind: "mcp",
		}

		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("Register: %v", err)
		}

		record, _ := r.Get(ctx, key)
		if record.Envelope.Version == "" {
			t.Error("expected default version")
		}
		if record.Envelope.State == "" {
			t.Error("expected default state")
		}
		if record.Envelope.Checksum == "" {
			t.Error("expected default checksum")
		}
		if record.Envelope.CreatedAt.IsZero() {
			t.Error("expected created timestamp")
		}
	})

	t.Run("get non-existent", func(t *testing.T) {
		r := NewArtifactRegistry()
		_, err := r.Get(ctx, newTestKey("skill", "nonexistent"))
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		ctx2, cancel := context.WithCancel(ctx)
		cancel()

		key := newTestKey("skill", "cancelled")
		env := newTestEnvelope("skill", "cancelled", "1.0.0")
		if err := r.Register(ctx2, key, env); err == nil {
			t.Error("expected context cancelled error")
		}
		if _, err := r.Get(ctx2, key); err == nil {
			t.Error("expected context cancelled error")
		}
	})
}

// --- List ---

func TestRegistry_List(t *testing.T) {
	ctx := context.Background()
	r := NewArtifactRegistry()

	registerTestArtifact(t, r, "skill", "a", "1.0.0")
	registerTestArtifact(t, r, "skill", "b", "1.0.0")
	registerTestArtifact(t, r, "mcp", "x", "1.0.0")
	registerTestArtifact(t, r, "agent", "y", "1.0.0")

	t.Run("list by kind", func(t *testing.T) {
		keys, err := r.List(ctx, "skill")
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(keys) != 2 {
			t.Fatalf("expected 2 skills, got %d", len(keys))
		}
	})

	t.Run("list non-existent kind", func(t *testing.T) {
		keys, err := r.List(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(keys) != 0 {
			t.Errorf("expected 0 artifacts, got %d", len(keys))
		}
	})

	t.Run("list all", func(t *testing.T) {
		keys, err := r.List(ctx, "")
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(keys) != 4 {
			t.Fatalf("expected 4 artifacts, got %d", len(keys))
		}
	})

	t.Run("list sorted", func(t *testing.T) {
		keys, err := r.List(ctx, "")
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		// Should be sorted by kind then ID
		for i := 1; i < len(keys); i++ {
			if keys[i].Kind < keys[i-1].Kind {
				t.Errorf("not sorted by kind at index %d", i)
			} else if keys[i].Kind == keys[i-1].Kind && keys[i].ID < keys[i-1].ID {
				t.Errorf("not sorted by ID at index %d", i)
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		_, err := r.List(ctx2, "skill")
		if err == nil {
			t.Error("expected context cancelled error")
		}
	})
}

// --- Dependencies ---

func TestRegistry_AddDependency(t *testing.T) {
	ctx := context.Background()

	t.Run("add dependency", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "parent", "1.0.0")
		registerTestArtifact(t, r, "skill", "child", "1.0.0")

		parentKey := newTestKey("skill", "parent")
		dep := ArtifactDependency{Kind: "skill", ID: "child", VersionReq: ">=1.0.0"}
		if err := r.AddDependency(ctx, parentKey, dep); err != nil {
			t.Fatalf("AddDependency: %v", err)
		}

		record, err := r.Get(ctx, parentKey)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if len(record.Dependencies) != 1 {
			t.Fatalf("expected 1 dependency, got %d", len(record.Dependencies))
		}
		if record.Dependencies[0].ID != "child" {
			t.Errorf("expected dependency on child, got %s", record.Dependencies[0].ID)
		}
	})

	t.Run("duplicate dependency", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")
		registerTestArtifact(t, r, "skill", "b", "1.0.0")

		key := newTestKey("skill", "a")
		dep := ArtifactDependency{Kind: "skill", ID: "b"}
		if err := r.AddDependency(ctx, key, dep); err != nil {
			t.Fatalf("first AddDependency: %v", err)
		}
		if err := r.AddDependency(ctx, key, dep); err == nil {
			t.Error("expected duplicate dependency error")
		}
	})

	t.Run("dependency on non-existent artifact", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		key := newTestKey("nonexistent", "ghost")
		dep := ArtifactDependency{Kind: "skill", ID: "a"}
		if err := r.AddDependency(ctx, key, dep); err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		if err := r.AddDependency(ctx2, newTestKey("skill", "a"), ArtifactDependency{}); err == nil {
			t.Error("expected context cancelled error")
		}
	})
}

func TestRegistry_ResolveDependencies(t *testing.T) {
	ctx := context.Background()

	t.Run("simple chain", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")
		registerTestArtifact(t, r, "skill", "b", "1.0.0")
		registerTestArtifact(t, r, "skill", "c", "1.0.0")

		// a -> b -> c
		_ = r.AddDependency(ctx, newTestKey("skill", "b"), ArtifactDependency{Kind: "skill", ID: "c"})
		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{Kind: "skill", ID: "b"})

		deps, err := r.ResolveDependencies(ctx, newTestKey("skill", "a"))
		if err != nil {
			t.Fatalf("ResolveDependencies: %v", err)
		}
		// Should include b and c
		if len(deps) != 2 {
			t.Fatalf("expected 2 deps, got %d", len(deps))
		}
	})

	t.Run("no dependencies", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "solo", "1.0.0")

		deps, err := r.ResolveDependencies(ctx, newTestKey("skill", "solo"))
		if err != nil {
			t.Fatalf("ResolveDependencies: %v", err)
		}
		if len(deps) != 0 {
			t.Errorf("expected 0 deps, got %d", len(deps))
		}
	})

	t.Run("diamond dependency", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "top", "1.0.0")
		registerTestArtifact(t, r, "skill", "left", "1.0.0")
		registerTestArtifact(t, r, "skill", "right", "1.0.0")
		registerTestArtifact(t, r, "skill", "bottom", "1.0.0")

		// top -> left, top -> right, left -> bottom, right -> bottom
		_ = r.AddDependency(ctx, newTestKey("skill", "left"), ArtifactDependency{Kind: "skill", ID: "bottom"})
		_ = r.AddDependency(ctx, newTestKey("skill", "right"), ArtifactDependency{Kind: "skill", ID: "bottom"})
		_ = r.AddDependency(ctx, newTestKey("skill", "top"), ArtifactDependency{Kind: "skill", ID: "left"})
		_ = r.AddDependency(ctx, newTestKey("skill", "top"), ArtifactDependency{Kind: "skill", ID: "right"})

		deps, err := r.ResolveDependencies(ctx, newTestKey("skill", "top"))
		if err != nil {
			t.Fatalf("ResolveDependencies: %v", err)
		}
		// bottom should appear only once
		bottomCount := 0
		for _, d := range deps {
			if d.ID == "bottom" {
				bottomCount++
			}
		}
		if bottomCount != 1 {
			t.Errorf("expected bottom to appear once, appeared %d times", bottomCount)
		}
		if len(deps) != 3 {
			t.Errorf("expected 3 unique deps (left, right, bottom), got %d", len(deps))
		}
	})

	t.Run("non-existent key", func(t *testing.T) {
		r := NewArtifactRegistry()
		_, err := r.ResolveDependencies(ctx, newTestKey("skill", "missing"))
		if err == nil {
			t.Error("expected error for non-existent key")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		_, err := r.ResolveDependencies(ctx2, newTestKey("skill", "a"))
		if err == nil {
			t.Error("expected context cancelled error")
		}
	})

	t.Run("missing non-optional dependency", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")
		// a depends on missing (not registered)
		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{
			Kind: "skill", ID: "missing",
		})

		deps, err := r.ResolveDependencies(ctx, newTestKey("skill", "a"))
		if err != nil {
			t.Fatalf("ResolveDependencies: %v", err)
		}
		// Should still resolve (no hard failure), but dep is missing
		if len(deps) != 0 {
			t.Errorf("expected 0 deps, got %d", len(deps))
		}
	})

	t.Run("missing optional dependency", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")
		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{
			Kind: "skill", ID: "optional-missing", Optional: true,
		})

		deps, err := r.ResolveDependencies(ctx, newTestKey("skill", "a"))
		if err != nil {
			t.Fatalf("ResolveDependencies: %v", err)
		}
		if len(deps) != 0 {
			t.Errorf("expected 0 deps, got %d", len(deps))
		}
	})
}

func TestRegistry_ReverseDeps(t *testing.T) {
	ctx := context.Background()
	r := NewArtifactRegistry()

	registerTestArtifact(t, r, "skill", "shared", "1.0.0")
	registerTestArtifact(t, r, "skill", "a", "1.0.0")
	registerTestArtifact(t, r, "skill", "b", "1.0.0")

	// a -> shared, b -> shared
	_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{Kind: "skill", ID: "shared"})
	_ = r.AddDependency(ctx, newTestKey("skill", "b"), ArtifactDependency{Kind: "skill", ID: "shared"})

	rev, err := r.ReverseDeps(ctx, newTestKey("skill", "shared"))
	if err != nil {
		t.Fatalf("ReverseDeps: %v", err)
	}
	if len(rev) != 2 {
		t.Fatalf("expected 2 reverse deps, got %d", len(rev))
	}
}

// --- Usage Metrics ---

func TestRegistry_RecordUsage(t *testing.T) {
	ctx := context.Background()

	t.Run("record usage", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "tracked", "1.0.0")
		key := newTestKey("skill", "tracked")

		if err := r.RecordUsage(ctx, key, "chat", 5*time.Second, true); err != nil {
			t.Fatalf("RecordUsage: %v", err)
		}
		if err := r.RecordUsage(ctx, key, "chat", 3*time.Second, true); err != nil {
			t.Fatalf("RecordUsage: %v", err)
		}
		if err := r.RecordUsage(ctx, key, "api", 10*time.Second, false); err != nil {
			t.Fatalf("RecordUsage: %v", err)
		}

		m, err := r.GetUsageMetrics(ctx, key)
		if err != nil {
			t.Fatalf("GetUsageMetrics: %v", err)
		}
		if m.TotalActivations != 3 {
			t.Errorf("expected 3 activations, got %d", m.TotalActivations)
		}
		if m.AvgDuration != 6*time.Second {
			t.Errorf("expected avg 6s, got %v", m.AvgDuration)
		}
		if m.SuccessRate != 2.0/3.0 {
			t.Errorf("expected success rate %.4f, got %.4f", 2.0/3.0, m.SuccessRate)
		}
		if m.SurfacesUsed["chat"] != 2 {
			t.Errorf("expected chat=2, got %d", m.SurfacesUsed["chat"])
		}
		if m.SurfacesUsed["api"] != 1 {
			t.Errorf("expected api=1, got %d", m.SurfacesUsed["api"])
		}
	})

	t.Run("usage on non-existent", func(t *testing.T) {
		r := NewArtifactRegistry()
		err := r.RecordUsage(ctx, newTestKey("skill", "missing"), "x", 0, true)
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})

	t.Run("get metrics on non-existent", func(t *testing.T) {
		r := NewArtifactRegistry()
		_, err := r.GetUsageMetrics(ctx, newTestKey("skill", "missing"))
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		err := r.RecordUsage(ctx2, newTestKey("skill", "a"), "x", 0, true)
		if err == nil {
			t.Error("expected context cancelled error")
		}
		_, err = r.GetUsageMetrics(ctx2, newTestKey("skill", "a"))
		if err == nil {
			t.Error("expected context cancelled error")
		}
	})
}

// --- Deprecate ---

func TestRegistry_Deprecate(t *testing.T) {
	ctx := context.Background()

	t.Run("deprecate artifact", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "old", "1.0.0")
		key := newTestKey("skill", "old")

		if err := r.Deprecate(ctx, key); err != nil {
			t.Fatalf("Deprecate: %v", err)
		}

		record, err := r.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if record.Envelope.State != "deprecated" {
			t.Errorf("expected state deprecated, got %s", record.Envelope.State)
		}
	})

	t.Run("deprecate non-existent", func(t *testing.T) {
		r := NewArtifactRegistry()
		err := r.Deprecate(ctx, newTestKey("skill", "missing"))
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		if err := r.Deprecate(ctx2, newTestKey("skill", "a")); err == nil {
			t.Error("expected context cancelled error")
		}
	})
}

// --- Rollback ---

func TestRegistry_Rollback(t *testing.T) {
	ctx := context.Background()

	t.Run("rollback to previous version", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("skill", "rb")
		env := ArtifactEnvelope{
			ID:      "rb",
			Kind:    "skill",
			Version: "2.0.0",
		}
		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("Register: %v", err)
		}

		// Add a previous version to history
		r.mu.Lock()
		record := r.artifacts[key]
		record.VersionHistory = append(record.VersionHistory, VersionEntry{
			Version:   "1.0.0",
			CreatedAt: time.Now().Add(-time.Hour),
			Checksum:  "old-checksum",
		})
		r.mu.Unlock()

		if err := r.Rollback(ctx, key, "1.0.0"); err != nil {
			t.Fatalf("Rollback: %v", err)
		}

		record, err := r.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if record.Envelope.Version != "1.0.0" {
			t.Errorf("expected version 1.0.0, got %s", record.Envelope.Version)
		}
		if record.Envelope.Checksum != "old-checksum" {
			t.Errorf("expected checksum old-checksum, got %s", record.Envelope.Checksum)
		}
	})

	t.Run("rollback to non-existent version", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "rb2", "1.0.0")
		key := newTestKey("skill", "rb2")

		err := r.Rollback(ctx, key, "9.9.9")
		if err == nil {
			t.Error("expected error for non-existent version")
		}
	})

	t.Run("rollback non-existent artifact", func(t *testing.T) {
		r := NewArtifactRegistry()
		err := r.Rollback(ctx, newTestKey("skill", "missing"), "1.0.0")
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		if err := r.Rollback(ctx2, newTestKey("skill", "a"), "1.0.0"); err == nil {
			t.Error("expected context cancelled error")
		}
	})
}

// --- FindVersion ---

func TestRegistry_FindVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("find latest", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("skill", "ver")
		env := ArtifactEnvelope{
			ID:      "ver",
			Kind:    "skill",
			Version: "1.0.0",
		}
		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("Register: %v", err)
		}

		r.mu.Lock()
		record := r.artifacts[key]
		record.VersionHistory = append(record.VersionHistory,
			VersionEntry{Version: "2.0.0", CreatedAt: time.Now(), Checksum: "chk2"},
			VersionEntry{Version: "1.5.0", CreatedAt: time.Now(), Checksum: "chk15"},
		)
		r.mu.Unlock()

		rec, err := r.FindVersion(ctx, key, "latest")
		if err != nil {
			t.Fatalf("FindVersion: %v", err)
		}
		if rec.Envelope.Version != "2.0.0" {
			t.Errorf("expected version 2.0.0, got %s", rec.Envelope.Version)
		}
	})

	t.Run("find by constraint", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("skill", "ver2")
		env := ArtifactEnvelope{
			ID:      "ver2",
			Kind:    "skill",
			Version: "1.0.0",
		}
		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("Register: %v", err)
		}

		r.mu.Lock()
		record := r.artifacts[key]
		record.VersionHistory = append(record.VersionHistory,
			VersionEntry{Version: "2.0.0", CreatedAt: time.Now(), Checksum: "chk2"},
			VersionEntry{Version: "1.5.0", CreatedAt: time.Now(), Checksum: "chk15"},
		)
		r.mu.Unlock()

		rec, err := r.FindVersion(ctx, key, ">=1.0.0,<2.0.0")
		if err != nil {
			t.Fatalf("FindVersion: %v", err)
		}
		if rec.Envelope.Version != "1.5.0" {
			t.Errorf("expected version 1.5.0, got %s", rec.Envelope.Version)
		}
	})

	t.Run("no matching version", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "ver3", "1.0.0")
		key := newTestKey("skill", "ver3")

		_, err := r.FindVersion(ctx, key, ">=5.0.0")
		if err == nil {
			t.Error("expected error for no matching version")
		}
	})

	t.Run("non-existent artifact", func(t *testing.T) {
		r := NewArtifactRegistry()
		_, err := r.FindVersion(ctx, newTestKey("skill", "missing"), "latest")
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})
}

// --- ValidateDependencies ---

func TestRegistry_ValidateDependencies(t *testing.T) {
	ctx := context.Background()

	t.Run("all valid", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")
		registerTestArtifact(t, r, "skill", "b", "1.0.0")

		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{
			Kind: "skill", ID: "b", VersionReq: ">=1.0.0",
		})

		errors := r.ValidateDependencies(ctx)
		if len(errors) != 0 {
			t.Errorf("expected 0 validation errors, got %d", len(errors))
			for _, e := range errors {
				t.Logf("  %s", e)
			}
		}
	})

	t.Run("missing required dependency", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{
			Kind: "skill", ID: "missing",
		})

		errors := r.ValidateDependencies(ctx)
		if len(errors) != 1 {
			t.Fatalf("expected 1 validation error, got %d", len(errors))
		}
		if errors[0].Dep.ID != "missing" {
			t.Errorf("expected missing dependency on 'missing', got %s", errors[0].Dep.ID)
		}
	})

	t.Run("optional missing dependency passes", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")

		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{
			Kind: "skill", ID: "optional", Optional: true,
		})

		errors := r.ValidateDependencies(ctx)
		if len(errors) != 0 {
			t.Errorf("expected 0 errors for optional dep, got %d", len(errors))
		}
	})

	t.Run("version constraint violation", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "a", "1.0.0")
		registerTestArtifact(t, r, "skill", "b", "0.5.0")

		_ = r.AddDependency(ctx, newTestKey("skill", "a"), ArtifactDependency{
			Kind: "skill", ID: "b", VersionReq: ">=1.0.0",
		})

		errors := r.ValidateDependencies(ctx)
		if len(errors) != 1 {
			t.Fatalf("expected 1 validation error, got %d", len(errors))
		}
	})

	t.Run("no dependencies to validate", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "solo", "1.0.0")

		errors := r.ValidateDependencies(ctx)
		if len(errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(errors))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		r := NewArtifactRegistry()
		ctx2, cancel := context.WithCancel(ctx)
		cancel()

		errors := r.ValidateDependencies(ctx2)
		if len(errors) != 1 {
			t.Errorf("expected 1 error for cancelled context, got %d", len(errors))
		}
	})
}

// --- Size ---

func TestRegistry_Size(t *testing.T) {
	r := NewArtifactRegistry()
	if r.Size() != 0 {
		t.Errorf("expected size 0, got %d", r.Size())
	}

	registerTestArtifact(t, r, "skill", "a", "1.0.0")
	if r.Size() != 1 {
		t.Errorf("expected size 1, got %d", r.Size())
	}

	registerTestArtifact(t, r, "mcp", "b", "1.0.0")
	if r.Size() != 2 {
		t.Errorf("expected size 2, got %d", r.Size())
	}
}

// --- AddVersion ---

func TestRegistry_AddVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("add version", func(t *testing.T) {
		r := NewArtifactRegistry()
		registerTestArtifact(t, r, "skill", "ver", "1.0.0")
		key := newTestKey("skill", "ver")

		if err := r.AddVersion(ctx, key, "2.0.0", "new feature"); err != nil {
			t.Fatalf("AddVersion: %v", err)
		}

		record, err := r.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if record.Envelope.Version != "2.0.0" {
			t.Errorf("expected version 2.0.0, got %s", record.Envelope.Version)
		}
		if len(record.VersionHistory) != 2 {
			t.Errorf("expected 2 versions in history, got %d", len(record.VersionHistory))
		}
	})

	t.Run("add version to non-existent", func(t *testing.T) {
		r := NewArtifactRegistry()
		err := r.AddVersion(ctx, newTestKey("skill", "missing"), "1.0.0", "")
		if err == nil {
			t.Error("expected error for non-existent artifact")
		}
	})
}

// --- Concurrent Access ---

func TestRegistry_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	r := NewArtifactRegistry()

	// Pre-register artifacts
	for i := 0; i < 10; i++ {
		registerTestArtifact(t, r, "skill", fmt.Sprintf("conc-%d", i), "1.0.0")
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := newTestKey("skill", fmt.Sprintf("conc-%d", idx%10))
			_, err := r.Get(ctx, key)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := newTestKey("skill", fmt.Sprintf("conc-%d", idx))
			if err := r.RecordUsage(ctx, key, "surface", 1*time.Second, true); err != nil {
				errors <- err
			}
		}(i)
	}

	// Concurrent lists
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.List(ctx, "skill")
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestRegistry_ConcurrentDepAddition(t *testing.T) {
	ctx := context.Background()
	r := NewArtifactRegistry()

	registerTestArtifact(t, r, "skill", "root", "1.0.0")
	for i := 0; i < 20; i++ {
		registerTestArtifact(t, r, "skill", fmt.Sprintf("dep-%d", i), "1.0.0")
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			dep := ArtifactDependency{
				Kind: "skill",
				ID:   fmt.Sprintf("dep-%d", idx),
			}
			if err := r.AddDependency(ctx, newTestKey("skill", "root"), dep); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent dep addition error: %v", err)
	}
}

// --- Error Wrapping ---

func TestRegistry_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("ErrNotFound is returned for get", func(t *testing.T) {
		r := NewArtifactRegistry()
		_, err := r.Get(ctx, newTestKey("skill", "nope"))
		if err == nil {
			t.Fatal("expected error")
		}
		// Check that ErrNotFound is in the chain
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("ErrDuplicate is returned for register", func(t *testing.T) {
		r := NewArtifactRegistry()
		key := newTestKey("skill", "dup-err")
		env := newTestEnvelope("skill", "dup-err", "1.0.0")
		if err := r.Register(ctx, key, env); err != nil {
			t.Fatalf("first Register: %v", err)
		}
		if err := r.Register(ctx, key, env); err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- Get returns a copy ---

func TestRegistry_GetReturnsCopy(t *testing.T) {
	ctx := context.Background()
	r := NewArtifactRegistry()
	registerTestArtifact(t, r, "skill", "copy-test", "1.0.0")
	key := newTestKey("skill", "copy-test")

	// Get and modify
	record, _ := r.Get(ctx, key)
	record.Envelope.Name = "modified-name"

	// Get again and verify original is unchanged
	record2, _ := r.Get(ctx, key)
	if record2.Envelope.Name == "modified-name" {
		t.Error("Get should return a copy; original was modified through returned pointer")
	}
}
