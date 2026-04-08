package environment

import (
	"os"
	"strings"
)

const (
	Development = "development"
	Staging     = "staging"
	Production  = "production"
	ProdSafe    = "prod-safe"
	DevOnly     = "dev-only"
)

// Current returns the active Brain environment.
// The default is production-safe behavior.
func Current() string {
	return Normalize(os.Getenv("BRAIN_ENV"))
}

// Normalize canonicalizes the environment value and falls back to production.
func Normalize(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case Development, Staging, Production:
		return value
	case "":
		return Production
	default:
		return Production
	}
}

// AllowsVisibility reports whether a component can be loaded in the given environment.
func AllowsVisibility(visibility string, environment string) bool {
	normalizedVisibility := strings.ToLower(strings.TrimSpace(visibility))
	if normalizedVisibility == "" {
		normalizedVisibility = ProdSafe
	}

	switch normalizedVisibility {
	case ProdSafe:
		return true
	case DevOnly:
		return environment == Development
	default:
		return false
	}
}
