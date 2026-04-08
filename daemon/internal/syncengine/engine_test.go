package syncengine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reeinharrrd/brain/daemon/internal/manager"
	"github.com/reeinharrrd/brain/daemon/internal/manifest"
)

func TestRunSyncWritesCanonicalArtifacts(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	brainRoot := filepath.Join(tempHome, ".brain")
	mustMkdir(t, filepath.Join(brainRoot, "skills"))
	mustMkdir(t, filepath.Join(brainRoot, "mcp"))
	mustMkdir(t, filepath.Join(brainRoot, "artifacts", "rules", "modules"))
	mustMkdir(t, filepath.Join(brainRoot, "artifacts", "agents"))
	mustMkdir(t, filepath.Join(tempHome, "out"))

	writeFile(t, filepath.Join(brainRoot, "skills", "registry.yml"), []byte(`skills:
  alpha:
    type: internal
    path: ~/.brain/skills/alpha/
    purpose: Alpha skill
    when_to_use: Always
    capabilities: [inspect, report]
`))

	writeFile(t, filepath.Join(brainRoot, "mcp", "registry.yml"), []byte(`mcps:
  memory:
    package: "@modelcontextprotocol/server-memory"
    profile: [standard, full]
    required: true
    purpose: Memory backend
    setup: "npx -y @modelcontextprotocol/server-memory"
`))

	writeFile(t, filepath.Join(brainRoot, "artifacts", "rules", "canonical.md"), []byte("# Canonical Rules\n\nBase rules body.\n"))
	writeFile(t, filepath.Join(brainRoot, "artifacts", "rules", "modules", "testing.md"), []byte("## Testing\n\nWrite tests first.\n"))
	writeFile(t, filepath.Join(brainRoot, "artifacts", "agents", "architect.md"), []byte(`---
name: architect
description: Designs technical solutions and trade-offs
version: 2.1.0
model: claude-opus
temperature: 0.4
tags:
  - architecture
  - planning
maintained: true
---

# Architect Agent

Guidance text for the architect agent.
`))

	manifestPath := filepath.Join(brainRoot, "manifest.yml")
	writeFile(t, manifestPath, []byte(`version: "1.0"
settings:
  backup_enabled: false
  backup_dir: ~/.brain/backups
  backup_retention_days: 7
  dry_run_by_default: false
domains:
  skills:
    source: "skills/registry.yml"
    enabled: true
  mcp:
    source: "mcp/registry.yml"
    enabled: true
  rules:
    source: "artifacts/rules/canonical.md"
    enabled: true
  agents:
    source: "artifacts/agents/"
    enabled: true
targets:
  cli:
    enabled: true
    type: "brain-cli"
    output_dirs:
      skills: "~/out/skills.json"
      mcp: "~/out/mcps.json"
      agents: "~/out/agents.json"
      rules: "~/out/brain.instructions.md"
    managed_sections: false
`))

	m, err := manifest.Parse(manifestPath)
	if err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	logger := make(chan string, 64)
	skillsRegistry := manager.NewSkillsRegistry(brainRoot, logger)
	if err := skillsRegistry.Load(context.Background()); err != nil {
		t.Fatalf("load skills registry: %v", err)
	}

	engine := NewSyncEngine(m, logger, skillsRegistry)
	if err := engine.RunSync(); err != nil {
		t.Fatalf("run sync: %v", err)
	}

	assertFileExists(t, filepath.Join(tempHome, "out", "skills.json"))
	assertFileExists(t, filepath.Join(tempHome, "out", "mcps.json"))
	assertFileExists(t, filepath.Join(tempHome, "out", "agents.json"))
	assertFileExists(t, filepath.Join(tempHome, "out", "brain.instructions.md"))
}

func TestRunSyncDryRunDoesNotWriteArtifacts(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	brainRoot := filepath.Join(tempHome, ".brain")
	mustMkdir(t, filepath.Join(brainRoot, "skills"))
	mustMkdir(t, filepath.Join(brainRoot, "mcp"))
	mustMkdir(t, filepath.Join(brainRoot, "artifacts", "rules"))
	mustMkdir(t, filepath.Join(brainRoot, "artifacts", "agents"))

	writeFile(t, filepath.Join(brainRoot, "skills", "registry.yml"), []byte(`skills:
  alpha:
    type: internal
    path: ~/.brain/skills/alpha/
    purpose: Alpha skill
`))
	writeFile(t, filepath.Join(brainRoot, "mcp", "registry.yml"), []byte(`mcps:
  memory:
    package: "@modelcontextprotocol/server-memory"
    profile: [standard, full]
`))
	writeFile(t, filepath.Join(brainRoot, "artifacts", "rules", "canonical.md"), []byte("# Canonical\n"))
	writeFile(t, filepath.Join(brainRoot, "artifacts", "agents", "architect.md"), []byte(`---
name: architect
description: Designs technical solutions
---
`))
	manifestPath := filepath.Join(brainRoot, "manifest.yml")
	writeFile(t, manifestPath, []byte(`version: "1.0"
settings:
  backup_enabled: false
  backup_dir: ~/.brain/backups
  backup_retention_days: 7
  dry_run_by_default: false
domains:
  skills:
    source: "skills/registry.yml"
    enabled: true
  mcp:
    source: "mcp/registry.yml"
    enabled: true
  rules:
    source: "artifacts/rules/canonical.md"
    enabled: true
  agents:
    source: "artifacts/agents/"
    enabled: true
targets:
  cli:
    enabled: true
    type: "brain-cli"
    output_dirs:
      skills: "~/out/skills.json"
      mcp: "~/out/mcps.json"
      agents: "~/out/agents.json"
      rules: "~/out/brain.instructions.md"
    managed_sections: false
`))

	m, err := manifest.Parse(manifestPath)
	if err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	logger := make(chan string, 64)
	skillsRegistry := manager.NewSkillsRegistry(brainRoot, logger)
	if err := skillsRegistry.Load(context.Background()); err != nil {
		t.Fatalf("load skills registry: %v", err)
	}

	engine := NewSyncEngine(m, logger, skillsRegistry)
	if err := engine.RunSyncWithOptions(SyncOptions{DryRun: true}); err != nil {
		t.Fatalf("run dry sync: %v", err)
	}

	assertFileMissing(t, filepath.Join(tempHome, "out", "skills.json"))
	assertFileMissing(t, filepath.Join(tempHome, "out", "mcps.json"))
	assertFileMissing(t, filepath.Join(tempHome, "out", "agents.json"))
	assertFileMissing(t, filepath.Join(tempHome, "out", "brain.instructions.md"))
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s (%v)", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected file to be absent: %s", path)
	}
}
