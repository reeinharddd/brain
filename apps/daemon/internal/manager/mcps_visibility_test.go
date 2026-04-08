package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMCPsManagerSkipsDevOnlyInProduction(t *testing.T) {
	brainRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(brainRoot, "artifacts", "mcps"), 0755); err != nil {
		t.Fatalf("Failed to create mcp dir: %v", err)
	}

	registry := `mcps:
  prod-safe-mcp:
    package: "@example/prod-safe"
    profile: [standard]
    visibility: prod-safe
  dev-only-mcp:
    package: "@example/dev-only"
    profile: [standard]
    visibility: dev-only
`
	if err := os.WriteFile(filepath.Join(brainRoot, "artifacts", "mcps", "registry.yml"), []byte(registry), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	logCh := make(chan string, 20)
	manager := NewMCPsManager(brainRoot, "production", logCh)
	if err := manager.Load(context.Background()); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	all := manager.GetAll(context.Background())
	if len(all) != 1 {
		t.Fatalf("Expected 1 MCP in production, got %d", len(all))
	}

	if got := manager.GetByID(context.Background(), "dev-only-mcp"); got != nil {
		t.Fatal("dev-only MCP should not be loaded in production")
	}

	if got := manager.GetByID(context.Background(), "prod-safe-mcp"); got == nil {
		t.Fatal("prod-safe MCP should be loaded in production")
	}
}

func TestMCPsManagerKeepsDevOnlyInDevelopment(t *testing.T) {
	brainRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(brainRoot, "artifacts", "mcps"), 0755); err != nil {
		t.Fatalf("Failed to create mcp dir: %v", err)
	}

	registry := `mcps:
  dev-only-mcp:
    package: "@example/dev-only"
    profile: [standard]
    visibility: dev-only
`
	if err := os.WriteFile(filepath.Join(brainRoot, "artifacts", "mcps", "registry.yml"), []byte(registry), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	logCh := make(chan string, 20)
	manager := NewMCPsManager(brainRoot, "development", logCh)
	if err := manager.Load(context.Background()); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := manager.GetByID(context.Background(), "dev-only-mcp"); got == nil {
		t.Fatal("dev-only MCP should be loaded in development")
	}
}
