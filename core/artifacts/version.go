package artifacts

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ParseVersion parses a version string like "1.2.3" into a slice of ints.
// Pre-release suffixes (e.g. "-beta") are stripped before parsing.
func ParseVersion(v string) ([]int, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, fmt.Errorf("empty version string")
	}

	// Strip pre-release and build metadata
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("invalid version %q: empty component", v)
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid version %q: %w", v, err)
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid version %q: negative component %q", v, p)
		}
		result = append(result, n)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("invalid version %q: no components", v)
	}
	return result, nil
}

// CompareVersions compares two version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Versions are compared component by component; shorter versions are padded with zeros.
func CompareVersions(a, b string) int {
	pa, errA := ParseVersion(a)
	pb, errB := ParseVersion(b)

	// If either fails to parse, fall back to string comparison
	if errA != nil || errB != nil {
		if a == b {
			return 0
		}
		if a < b {
			return -1
		}
		return 1
	}

	// Pad to same length
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for len(pa) < maxLen {
		pa = append(pa, 0)
	}
	for len(pb) < maxLen {
		pb = append(pb, 0)
	}

	for i := 0; i < maxLen; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

// MatchVersion checks whether a version string matches a constraint.
// Supported constraint formats:
//   - "latest" — always matches
//   - exact match: "1.2.3"
//   - comparison: ">=1.0.0", ">1.0.0", "<=2.0.0", "<2.0.0"
//   - compound: ">=1.0.0,<2.0.0" (comma-separated, all must match)
func MatchVersion(version, constraint string) bool {
	version = strings.TrimSpace(version)
	constraint = strings.TrimSpace(constraint)

	if constraint == "" || constraint == "latest" {
		return true
	}

	// Compound constraints separated by commas
	parts := strings.Split(constraint, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !matchSingleConstraint(version, part) {
			return false
		}
	}
	return true
}

func matchSingleConstraint(version, constraint string) bool {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" || constraint == "latest" {
		return true
	}

	// Determine the operator prefix
	var op string
	var cVersion string
	switch {
	case strings.HasPrefix(constraint, ">="):
		op = ">="
		cVersion = strings.TrimPrefix(constraint, ">=")
	case strings.HasPrefix(constraint, "<="):
		op = "<="
		cVersion = strings.TrimPrefix(constraint, "<=")
	case strings.HasPrefix(constraint, "!="):
		op = "!="
		cVersion = strings.TrimPrefix(constraint, "!=")
	case strings.HasPrefix(constraint, ">"):
		op = ">"
		cVersion = strings.TrimPrefix(constraint, ">")
	case strings.HasPrefix(constraint, "<"):
		op = "<"
		cVersion = strings.TrimPrefix(constraint, "<")
	case strings.HasPrefix(constraint, "="):
		op = "="
		cVersion = strings.TrimPrefix(constraint, "=")
	default:
		// Exact match
		return CompareVersions(version, constraint) == 0
	}

	cmp := CompareVersions(version, cVersion)
	switch op {
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	case "!=":
		return cmp != 0
	case "=":
		return cmp == 0
	default:
		return false
	}
}

// SelectBestVersion selects the best (highest) version from a list that matches the constraint.
// Returns empty string if no version matches.
func SelectBestVersion(versions []string, constraint string) string {
	var candidates []string
	for _, v := range versions {
		if MatchVersion(v, constraint) {
			candidates = append(candidates, v)
		}
	}
	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return CompareVersions(candidates[i], candidates[j]) > 0
	})
	return candidates[0]
}
