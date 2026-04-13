package manager

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseSkillsListOutput(t *testing.T) {
	output := `███████╗██╗  ██╗██╗██╗     ██╗     ███████╗
██╔════╝██║ ██╔╝██║██║     ██║     ██╔════╝
███████╗█████╔╝ ██║██║     ██║     ███████╗
╚════██║██╔═██╗ ██║██║     ██║     ╚════██║
███████║██║  ██╗██║███████╗███████╗███████║
╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚══════╝

┌   skills 
│
◇  Source: https://github.com/vercel-labs/agent-skills.git
│
◇  Found 2 skills
│
◇  Available Skills
│
│    vercel-composition-patterns
│
│      React composition patterns that scale. Use when refactoring components.
│      Includes React 19 API changes.
│
│    deploy-to-vercel
│
│      Deploy applications and websites to Vercel.
│
└  Use --skill <name> to install specific skills`

	records := parseSkillsListOutput(output)
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	if records[0].ID != "vercel-composition-patterns" {
		t.Fatalf("unexpected first record ID: %q", records[0].ID)
	}
	if records[0].Name != "vercel-composition-patterns" {
		t.Fatalf("unexpected first record name: %q", records[0].Name)
	}
	if records[0].Description == "" || records[0].Description == "React composition patterns that scale. Use when refactoring components." {
		// The parser should include both wrapped lines.
	}
	if records[0].Description != "React composition patterns that scale. Use when refactoring components. Includes React 19 API changes." {
		t.Fatalf("unexpected first record description: %q", records[0].Description)
	}

	if records[1].ID != "deploy-to-vercel" {
		t.Fatalf("unexpected second record ID: %q", records[1].ID)
	}
	if records[1].Description != "Deploy applications and websites to Vercel." {
		t.Fatalf("unexpected second record description: %q", records[1].Description)
	}
}

func TestDiscoverSkillRootsIncludesHiddenTrees(t *testing.T) {
	root := t.TempDir()

	paths := []string{
		filepath.Join(root, ".agents", "skills", "skill-one"),
		filepath.Join(root, ".github", "skills", "skill-two"),
	}
	for _, skillRoot := range paths {
		if err := os.MkdirAll(skillRoot, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", skillRoot, err)
		}
		if err := os.WriteFile(filepath.Join(skillRoot, "SKILL.md"), []byte("---\nname: test\n---\nbody"), 0o644); err != nil {
			t.Fatalf("write skill md %s: %v", skillRoot, err)
		}
	}

	roots, err := discoverSkillRoots(root)
	if err != nil {
		t.Fatalf("discoverSkillRoots: %v", err)
	}

	if !reflect.DeepEqual(roots, paths) {
		t.Fatalf("unexpected roots:\nwant: %v\n got: %v", paths, roots)
	}
}

func TestParseSkillInstallRecord(t *testing.T) {
	root := t.TempDir()
	markdown := `---
name: Example Skill
description: Helpful skill
version: 2.1.0
category: automation
metadata:
  internal: true
---

# Example Skill

Do useful things.
`
	if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte(markdown), 0o644); err != nil {
		t.Fatalf("write skill md: %v", err)
	}

	record, err := parseSkillInstallRecord(root)
	if err != nil {
		t.Fatalf("parseSkillInstallRecord: %v", err)
	}

	if record.ID != filepath.Base(root) {
		t.Fatalf("unexpected ID: %q", record.ID)
	}
	if record.Name != "Example Skill" {
		t.Fatalf("unexpected name: %q", record.Name)
	}
	if record.Description != "Helpful skill" {
		t.Fatalf("unexpected description: %q", record.Description)
	}
	if record.Version != "2.1.0" {
		t.Fatalf("unexpected version: %q", record.Version)
	}
	if record.Category != "automation" {
		t.Fatalf("unexpected category: %q", record.Category)
	}
	if !record.Internal {
		t.Fatal("expected internal metadata to be true")
	}
	if len(record.Files) != 1 || record.Files[0] != "SKILL.md" {
		t.Fatalf("unexpected files: %v", record.Files)
	}
}
