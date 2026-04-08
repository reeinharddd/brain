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

func TestPathInDomainSupportsCanonicalAndLegacy(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		domain string
		want   bool
	}{
		{name: "canonical skill path", path: "artifacts/skills/registry.yml", domain: "skills", want: true},
		{name: "legacy skill path", path: "skills/registry.yml", domain: "skills", want: true},
		{name: "canonical mcp path", path: "artifacts/mcps/registry.yml", domain: "mcps", want: true},
		{name: "legacy mcp path", path: "mcp/registry.yml", domain: "mcps", want: true},
		{name: "non matching path", path: "docs/README.md", domain: "skills", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PathInDomain(tc.path, tc.domain)
			if got != tc.want {
				t.Fatalf("PathInDomain(%q, %q) = %v, want %v", tc.path, tc.domain, got, tc.want)
			}
		})
	}
}

func TestCanonicalizePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		domain string
		want   string
	}{
		{
			name:   "legacy relative agent path",
			path:   "agents/architect.md",
			domain: "agents",
			want:   "artifacts/agents/architect.md",
		},
		{
			name:   "legacy absolute agent path",
			path:   "/tmp/repo/agents/architect.md",
			domain: "agents",
			want:   "artifacts/agents/architect.md",
		},
		{
			name:   "already canonical path",
			path:   "artifacts/agents/architect.md",
			domain: "agents",
			want:   "artifacts/agents/architect.md",
		},
		{
			name:   "legacy mcp path",
			path:   "mcp/registry.yml",
			domain: "mcps",
			want:   "artifacts/mcps/registry.yml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CanonicalizePath(tc.path, tc.domain)
			if got != tc.want {
				t.Fatalf("CanonicalizePath(%q, %q) = %q, want %q", tc.path, tc.domain, got, tc.want)
			}
		})
	}
}
