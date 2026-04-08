package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRootFilePathUsesHomeDirectory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	got := configRootFilePath()
	want := filepath.Join(homeDir, ".config", "brain", "root")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestReadConfiguredRootValidation(t *testing.T) {
	t.Run("returns configured root when manifest exists", func(t *testing.T) {
		homeDir := t.TempDir()
		t.Setenv("HOME", homeDir)

		root := filepath.Join(homeDir, "workspace", "brain")
		if err := os.MkdirAll(filepath.Join(root, "apps", "cli", "cmd", "brain"), 0755); err != nil {
			t.Fatalf("mkdir root: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(root, "apps", "daemon", "cmd", "braind"), 0755); err != nil {
			t.Fatalf("mkdir daemon app tree: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "manifest.yml"), []byte("version: 1"), 0644); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "apps", "cli", "cmd", "brain", "main.go"), []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write cli main: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "apps", "daemon", "cmd", "braind", "main.go"), []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write daemon main: %v", err)
		}

		configPath := configRootFilePath()
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}
		if err := os.WriteFile(configPath, []byte(root+"\n"), 0644); err != nil {
			t.Fatalf("write config root: %v", err)
		}

		if got := readConfiguredRoot(); got != root {
			t.Fatalf("expected %q, got %q", root, got)
		}
	})

	t.Run("returns empty when configured path is invalid", func(t *testing.T) {
		homeDir := t.TempDir()
		t.Setenv("HOME", homeDir)

		root := filepath.Join(homeDir, "workspace", "brain")
		configPath := configRootFilePath()
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}
		if err := os.WriteFile(configPath, []byte(root+"\n"), 0644); err != nil {
			t.Fatalf("write config root: %v", err)
		}

		if got := readConfiguredRoot(); got != "" {
			t.Fatalf("expected empty configured root, got %q", got)
		}
	})
}

func TestResolveBrainRootPrefersEnvironment(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	envRoot := filepath.Join(homeDir, "env-brain")
	if err := os.MkdirAll(filepath.Join(envRoot, "apps", "cli", "cmd", "brain"), 0755); err != nil {
		t.Fatalf("mkdir env root: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(envRoot, "apps", "daemon", "cmd", "braind"), 0755); err != nil {
		t.Fatalf("mkdir env daemon tree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "manifest.yml"), []byte("version: 1"), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "apps", "cli", "cmd", "brain", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("write env cli main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "apps", "daemon", "cmd", "braind", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("write env daemon main: %v", err)
	}

	t.Setenv("BRAIN_ROOT", envRoot)
	if got := resolveBrainRoot(); got != envRoot {
		t.Fatalf("expected resolveBrainRoot=%q, got %q", envRoot, got)
	}
}
