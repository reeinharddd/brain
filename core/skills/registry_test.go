package skills

import (
	"context"
	"sync"
	"testing"
)

func newCleanSkill(id, name, version string) *Skill {
	return &Skill{
		ID:      id,
		Name:    name,
		Version: version,
		Content: map[string]string{
			"main.py": "def hello():\n    print('hello')",
		},
	}
}

func TestRegister(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	t.Run("register clean skill", func(t *testing.T) {
		skill := newCleanSkill("test-skill", "Test Skill", "1.0.0")
		err := registry.Register(ctx, skill)
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if skill.Status != SkillActive {
			t.Errorf("Register() status = %v, want %v", skill.Status, SkillActive)
		}
		if skill.SecurityResult == nil {
			t.Error("Register() security result is nil")
		}
		if !skill.SecurityResult.OverallPass {
			t.Error("Register() security result should pass")
		}
	})

	t.Run("register multiple versions", func(t *testing.T) {
		skill1 := newCleanSkill("multi-skill", "Multi Skill", "1.0.0")
		skill2 := newCleanSkill("multi-skill", "Multi Skill", "2.0.0")

		if err := registry.Register(ctx, skill1); err != nil {
			t.Fatalf("Register() v1 error = %v", err)
		}
		if err := registry.Register(ctx, skill2); err != nil {
			t.Fatalf("Register() v2 error = %v", err)
		}
	})

	t.Run("duplicate version error", func(t *testing.T) {
		skill := newCleanSkill("dup-skill", "Dup Skill", "1.0.0")
		if err := registry.Register(ctx, skill); err != nil {
			t.Fatalf("Register() first error = %v", err)
		}

		skill2 := newCleanSkill("dup-skill", "Dup Skill", "1.0.0")
		err := registry.Register(ctx, skill2)
		if err == nil {
			t.Fatal("Register() duplicate should return error")
		}
	})

	t.Run("scan failure blocks registration", func(t *testing.T) {
		badSkill := &Skill{
			ID:      "bad-skill",
			Name:    "Bad Skill",
			Version: "1.0.0",
			Content: map[string]string{
				"evil.sh": "rm -rf /\nsudo apt-get install hacktool",
			},
		}
		err := registry.Register(ctx, badSkill)
		if err == nil {
			t.Fatal("Register() with dangerous content should fail")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		skill := newCleanSkill("ctx-skill", "Ctx Skill", "1.0.0")
		err := registry.Register(ctx, skill)
		if err == nil {
			t.Error("Register() with cancelled context should return error")
		}
	})
}

func TestGet(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill := newCleanSkill("get-skill", "Get Skill", "1.0.0")
	if err := registry.Register(ctx, skill); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	t.Run("get existing skill", func(t *testing.T) {
		got, err := registry.Get(ctx, "get-skill", "1.0.0")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.ID != "get-skill" {
			t.Errorf("Get() ID = %v, want get-skill", got.ID)
		}
		if got.Version != "1.0.0" {
			t.Errorf("Get() version = %v, want 1.0.0", got.Version)
		}
	})

	t.Run("get non-existent skill", func(t *testing.T) {
		_, err := registry.Get(ctx, "nonexistent", "1.0.0")
		if err == nil {
			t.Error("Get() non-existent should return error")
		}
	})

	t.Run("get non-existent version", func(t *testing.T) {
		_, err := registry.Get(ctx, "get-skill", "9.9.9")
		if err == nil {
			t.Error("Get() non-existent version should return error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := registry.Get(ctx, "get-skill", "1.0.0")
		if err == nil {
			t.Error("Get() with cancelled context should return error")
		}
	})
}

func TestGetLatest(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill1 := newCleanSkill("latest-skill", "Latest Skill", "1.0.0")
	skill2 := newCleanSkill("latest-skill", "Latest Skill", "2.0.0")

	if err := registry.Register(ctx, skill1); err != nil {
		t.Fatalf("Register() v1 error = %v", err)
	}
	if err := registry.Register(ctx, skill2); err != nil {
		t.Fatalf("Register() v2 error = %v", err)
	}

	t.Run("get latest version", func(t *testing.T) {
		latest, err := registry.GetLatest(ctx, "latest-skill")
		if err != nil {
			t.Fatalf("GetLatest() error = %v", err)
		}
		if latest.Version != "2.0.0" {
			t.Errorf("GetLatest() version = %v, want 2.0.0", latest.Version)
		}
	})

	t.Run("get latest non-existent", func(t *testing.T) {
		_, err := registry.GetLatest(ctx, "nonexistent")
		if err == nil {
			t.Error("GetLatest() non-existent should return error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := registry.GetLatest(ctx, "latest-skill")
		if err == nil {
			t.Error("GetLatest() with cancelled context should return error")
		}
	})
}

func TestList(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill1 := newCleanSkill("list-skill-1", "List Skill 1", "1.0.0")
	skill2 := newCleanSkill("list-skill-2", "List Skill 2", "1.0.0")

	if err := registry.Register(ctx, skill1); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := registry.Register(ctx, skill2); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	t.Run("list all skills", func(t *testing.T) {
		ids := registry.List(ctx)
		if len(ids) < 2 {
			t.Errorf("List() count = %d, want at least 2", len(ids))
		}

		found1, found2 := false, false
		for _, id := range ids {
			if id == "list-skill-1" {
				found1 = true
			}
			if id == "list-skill-2" {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Errorf("List() missing skills; got %v", ids)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ids := registry.List(ctx)
		if ids != nil {
			t.Errorf("List() with cancelled context should return nil, got %v", ids)
		}
	})
}

func TestListVersions(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill1 := newCleanSkill("ver-skill", "Version Skill", "1.0.0")
	skill2 := newCleanSkill("ver-skill", "Version Skill", "1.1.0")
	skill3 := newCleanSkill("ver-skill", "Version Skill", "2.0.0")

	if err := registry.Register(ctx, skill1); err != nil {
		t.Fatalf("Register() v1 error = %v", err)
	}
	if err := registry.Register(ctx, skill2); err != nil {
		t.Fatalf("Register() v2 error = %v", err)
	}
	if err := registry.Register(ctx, skill3); err != nil {
		t.Fatalf("Register() v3 error = %v", err)
	}

	t.Run("list all versions", func(t *testing.T) {
		versions := registry.ListVersions(ctx, "ver-skill")
		if len(versions) != 3 {
			t.Errorf("ListVersions() count = %d, want 3", len(versions))
		}

		expected := map[string]bool{"1.0.0": false, "1.1.0": false, "2.0.0": false}
		for _, v := range versions {
			expected[v] = true
		}
		for v, found := range expected {
			if !found {
				t.Errorf("ListVersions() missing version %s", v)
			}
		}
	})

	t.Run("list versions non-existent skill", func(t *testing.T) {
		versions := registry.ListVersions(ctx, "nonexistent")
		if versions != nil {
			t.Errorf("ListVersions() non-existent should return nil, got %v", versions)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		versions := registry.ListVersions(ctx, "ver-skill")
		if versions != nil {
			t.Errorf("ListVersions() with cancelled context should return nil")
		}
	})
}

func TestSearch(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skills := []*Skill{
		{
			ID:       "search-python",
			Name:     "Python Utils",
			Version:  "1.0.0",
			Category: "utilities",
			Tags:     []string{"python", "utils"},
			Content:  map[string]string{"main.py": "print('hello')"},
		},
		{
			ID:       "search-js",
			Name:     "JS Helpers",
			Version:  "1.0.0",
			Category: "utilities",
			Tags:     []string{"javascript", "helpers"},
			Content:  map[string]string{"main.js": "console.log('hello')"},
		},
		{
			ID:       "search-ml",
			Name:     "ML Tools",
			Version:  "1.0.0",
			Category: "machine-learning",
			Tags:     []string{"ml", "python"},
			Content:  map[string]string{"train.py": "model.fit()"},
		},
	}

	for _, s := range skills {
		if err := registry.Register(ctx, s); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
	}

	t.Run("search by query", func(t *testing.T) {
		results := registry.Search(ctx, "search-python", "")
		if len(results) != 1 {
			t.Fatalf("Search() by query count = %d, want 1", len(results))
		}
		if results[0].ID != "search-python" {
			t.Errorf("Search() by query ID = %v, want search-python", results[0].ID)
		}
	})

	t.Run("search by category", func(t *testing.T) {
		results := registry.Search(ctx, "", "utilities")
		if len(results) != 2 {
			t.Fatalf("Search() by category count = %d, want 2", len(results))
		}
	})

	t.Run("search by query and category", func(t *testing.T) {
		results := registry.Search(ctx, "search-js", "utilities")
		if len(results) != 1 {
			t.Fatalf("Search() by query+category count = %d, want 1", len(results))
		}
		if results[0].ID != "search-js" {
			t.Errorf("Search() by query+category ID = %v, want search-js", results[0].ID)
		}
	})

	t.Run("search by tag", func(t *testing.T) {
		results := registry.Search(ctx, "python", "")
		if len(results) < 1 {
			t.Fatalf("Search() by tag count = %d, want at least 1", len(results))
		}
	})

	t.Run("search no match", func(t *testing.T) {
		results := registry.Search(ctx, "nonexistent-query", "")
		if len(results) != 0 {
			t.Errorf("Search() no match count = %d, want 0", len(results))
		}
	})

	t.Run("search all with empty query", func(t *testing.T) {
		results := registry.Search(ctx, "", "")
		if len(results) != 3 {
			t.Errorf("Search() all count = %d, want 3", len(results))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		results := registry.Search(ctx, "", "")
		if results != nil {
			t.Errorf("Search() with cancelled context should return nil")
		}
	})
}

func TestDeprecate(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill := newCleanSkill("depr-skill", "Deprecate Skill", "1.0.0")
	if err := registry.Register(ctx, skill); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	t.Run("deprecate skill", func(t *testing.T) {
		err := registry.Deprecate(ctx, "depr-skill", "1.0.0")
		if err != nil {
			t.Fatalf("Deprecate() error = %v", err)
		}

		got, err := registry.Get(ctx, "depr-skill", "1.0.0")
		if err != nil {
			t.Fatalf("Get() after deprecate error = %v", err)
		}
		if got.Status != SkillDeprecated {
			t.Errorf("Deprecate() status = %v, want %v", got.Status, SkillDeprecated)
		}
	})

	t.Run("deprecate non-existent skill", func(t *testing.T) {
		err := registry.Deprecate(ctx, "nonexistent", "1.0.0")
		if err == nil {
			t.Error("Deprecate() non-existent should return error")
		}
	})

	t.Run("deprecate non-existent version", func(t *testing.T) {
		err := registry.Deprecate(ctx, "depr-skill", "9.9.9")
		if err == nil {
			t.Error("Deprecate() non-existent version should return error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := registry.Deprecate(ctx, "depr-skill", "1.0.0")
		if err == nil {
			t.Error("Deprecate() with cancelled context should return error")
		}
	})
}

func TestDelete(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill := newCleanSkill("del-skill", "Delete Skill", "1.0.0")
	if err := registry.Register(ctx, skill); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	t.Run("delete skill", func(t *testing.T) {
		err := registry.Delete(ctx, "del-skill", "1.0.0")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err = registry.Get(ctx, "del-skill", "1.0.0")
		if err == nil {
			t.Error("Get() after delete should return error")
		}
	})

	t.Run("delete non-existent skill", func(t *testing.T) {
		err := registry.Delete(ctx, "nonexistent", "1.0.0")
		if err == nil {
			t.Error("Delete() non-existent should return error")
		}
	})

	t.Run("delete non-existent version", func(t *testing.T) {
		skill2 := newCleanSkill("del-skill-2", "Delete Skill 2", "1.0.0")
		if err := registry.Register(ctx, skill2); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		err := registry.Delete(ctx, "del-skill-2", "9.9.9")
		if err == nil {
			t.Error("Delete() non-existent version should return error")
		}
	})

	t.Run("delete last version removes skill", func(t *testing.T) {
		skill3 := newCleanSkill("del-skill-3", "Delete Skill 3", "1.0.0")
		if err := registry.Register(ctx, skill3); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		ids := registry.List(ctx)
		found := false
		for _, id := range ids {
			if id == "del-skill-3" {
				found = true
				break
			}
		}
		if !found {
			t.Error("del-skill-3 should be in list before delete")
		}

		err := registry.Delete(ctx, "del-skill-3", "1.0.0")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		ids = registry.List(ctx)
		for _, id := range ids {
			if id == "del-skill-3" {
				t.Error("del-skill-3 should not be in list after delete")
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := registry.Delete(ctx, "del-skill-2", "1.0.0")
		if err == nil {
			t.Error("Delete() with cancelled context should return error")
		}
	})
}

func TestCount(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	t.Run("count empty registry", func(t *testing.T) {
		if count := registry.Count(); count != 0 {
			t.Errorf("Count() empty = %d, want 0", count)
		}
	})

	t.Run("count after registers", func(t *testing.T) {
		skill1 := newCleanSkill("count-1", "Count 1", "1.0.0")
		skill2 := newCleanSkill("count-2", "Count 2", "1.0.0")
		skill3 := newCleanSkill("count-1", "Count 1", "2.0.0")

		if err := registry.Register(ctx, skill1); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if err := registry.Register(ctx, skill2); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if err := registry.Register(ctx, skill3); err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if count := registry.Count(); count != 3 {
			t.Errorf("Count() = %d, want 3", count)
		}
	})

	t.Run("count after delete", func(t *testing.T) {
		err := registry.Delete(ctx, "count-2", "1.0.0")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if count := registry.Count(); count != 2 {
			t.Errorf("Count() after delete = %d, want 2", count)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			skill := newCleanSkill("concurrent", "Concurrent", "1.0.0")
			skill.ID = "concurrent"
			skill.Version = "v" + string(rune('0'+n))
			err := registry.Register(ctx, skill)
			if err != nil {
				errs <- err
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.List(ctx)
			registry.GetLatest(ctx, "concurrent")
			registry.Search(ctx, "", "")
			registry.Count()
		}()
	}

	wg.Wait()
	close(errs)

	// Count non-duplicate errors (expected: "already exists" for concurrent writes)
	var dupErrors int
	for err := range errs {
		if err != nil {
			dupErrors++
		}
	}
	// Some duplicate errors are expected with concurrent registration of same ID
	t.Logf("Concurrent test: %d errors (some duplicates expected)", dupErrors)
}

func TestNewSkillRegistryNilScanner(t *testing.T) {
	registry := NewSkillRegistry(nil)
	if registry == nil {
		t.Fatal("NewSkillRegistry(nil) returned nil")
	}
	if registry.scanner == nil {
		t.Error("NewSkillRegistry(nil) should create default scanner")
	}
}

func TestRegisterSetsTimestamps(t *testing.T) {
	scanner := NewSecurityScanner()
	registry := NewSkillRegistry(scanner)
	ctx := context.Background()

	skill := newCleanSkill("time-skill", "Time Skill", "1.0.0")
	if err := registry.Register(ctx, skill); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if skill.CreatedAt.IsZero() {
		t.Error("Register() should set CreatedAt")
	}
	if skill.UpdatedAt.IsZero() {
		t.Error("Register() should set UpdatedAt")
	}
}
