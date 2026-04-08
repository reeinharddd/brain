package artifacts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDomainFilePrefersCanonical(t *testing.T) {
	root := t.TempDir()
	canonical := filepath.Join(root, "artifacts", "skills")
	legacy := filepath.Join(root, "skills")
	if err := os.MkdirAll(canonical, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(legacy, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonical, "registry.yml"), []byte("skills: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "registry.yml"), []byte("skills: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got := NewLocator(root).DomainFile("skills", "registry.yml")
	want := filepath.Join(canonical, "registry.yml")
	if got != want {
		t.Fatalf("expected canonical path %s, got %s", want, got)
	}
}
