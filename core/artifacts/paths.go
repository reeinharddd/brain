package artifacts

import (
	"os"
	"path/filepath"
)

type Locator struct {
	Root string
}

func NewLocator(root string) Locator {
	return Locator{Root: root}
}

func (l Locator) DomainDir(domain string) string {
	candidates := []string{
		filepath.Join(l.Root, "artifacts", domain),
		legacyDomainDir(l.Root, domain),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return candidates[0]
}

func (l Locator) DomainFile(domain, name string) string {
	candidates := []string{
		filepath.Join(l.Root, "artifacts", domain, name),
		filepath.Join(legacyDomainDir(l.Root, domain), name),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return candidates[0]
}

func (l Locator) FirstExisting(paths ...string) []string {
	var existing []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	return existing
}

func (l Locator) FirstExistingGlob(patterns ...string) []string {
	for _, pattern := range patterns {
		entries, err := filepath.Glob(pattern)
		if err == nil && len(entries) > 0 {
			return entries
		}
	}
	return nil
}

func (l Locator) ReadFirstExistingFile(paths ...string) (string, []byte, error) {
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return path, content, nil
		}
	}
	return "", nil, os.ErrNotExist
}

func (l Locator) CanonicalRelative(domain, name string) string {
	return filepath.Join("artifacts", domain, name)
}

func legacyDomainDir(root, domain string) string {
	switch domain {
	case "mcps":
		return filepath.Join(root, "mcp")
	case "providers":
		return filepath.Join(root, "providers")
	default:
		return filepath.Join(root, domain)
	}
}
