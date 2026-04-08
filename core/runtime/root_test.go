package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBrainRootSupportsAppsTree(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, "apps", "cli", "cmd", "brain"),
		filepath.Join(root, "apps", "daemon", "cmd", "braind"),
	} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.yml"), []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "apps", "cli", "cmd", "brain", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "apps", "daemon", "cmd", "braind", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsBrainRoot(root) {
		t.Fatalf("expected %s to be a valid brain root", root)
	}
}
