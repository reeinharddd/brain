package artifacts

import (
	"os"
	"path/filepath"
	"strings"
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

func DomainPrefixes(domain string) []string {
	canonical := filepath.ToSlash(filepath.Join("artifacts", domain)) + "/"
	legacy := legacyDomainSegment(domain) + "/"

	if canonical == legacy {
		return []string{canonical}
	}

	return []string{canonical, legacy}
}

func PathInDomain(path, domain string) bool {
	normalized := normalizePath(path)
	for _, prefix := range DomainPrefixes(domain) {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	return false
}

func CanonicalizePath(path, domain string) string {
	normalized := normalizePath(path)
	canonicalPrefix := filepath.ToSlash(filepath.Join("artifacts", domain)) + "/"

	if idx := strings.Index(normalized, "/"+canonicalPrefix); idx >= 0 {
		return normalized[idx+1:]
	}
	if strings.HasPrefix(normalized, canonicalPrefix) {
		return normalized
	}

	legacyPrefix := legacyDomainSegment(domain) + "/"
	if idx := strings.Index(normalized, "/"+legacyPrefix); idx >= 0 {
		return canonicalPrefix + normalized[idx+len("/"+legacyPrefix):]
	}
	if strings.HasPrefix(normalized, legacyPrefix) {
		return canonicalPrefix + strings.TrimPrefix(normalized, legacyPrefix)
	}

	return normalized
}

func normalizePath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(path))
	normalized = strings.TrimPrefix(normalized, "./")
	return normalized
}

func legacyDomainSegment(domain string) string {
	switch domain {
	case "mcps":
		return "mcp"
	default:
		return domain
	}
}

func legacyDomainDir(root, domain string) string {
	return filepath.Join(root, legacyDomainSegment(domain))
}
