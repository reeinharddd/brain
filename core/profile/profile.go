// Package profile defines deployment profiles for the Brain daemon.
//
// Brain supports three deployment profiles that share a single codebase:
//
//   - local: Personal use, SQLite, single instance, no external dependencies.
//   - selfhosted: Corporate use, PostgreSQL, optional Redis, SSO required.
//   - cloud: Multi-tenant, PostgreSQL + Redis, horizontal scaling.
//
// The profile determines infrastructure defaults (database backend, session
// storage, auth requirements, telemetry level, etc.). Environment variables
// always override profile defaults via the ResolveConfig function in config.go.
package profile

import (
	"fmt"
	"os"
	"strings"
)

// Profile identifies a deployment posture for the Brain daemon.
type Profile string

// Known profiles.
const (
	Local      Profile = "local"
	SelfHosted Profile = "selfhosted"
	Cloud      Profile = "cloud"
)

// AllProfiles returns every known profile in a deterministic order.
func AllProfiles() []Profile {
	return []Profile{Local, SelfHosted, Cloud}
}

// String returns the canonical profile name.
func (p Profile) String() string { return string(p) }

// Validate returns an error when the profile is not one of the known values.
func (p Profile) Validate() error {
	switch p {
	case Local, SelfHosted, Cloud:
		return nil
	default:
		return fmt.Errorf("unknown profile %q (expected %q, %q, or %q)", p, Local, SelfHosted, Cloud)
	}
}

// RequiresExternalService reports whether the profile depends on services
// outside the local machine (PostgreSQL, Redis, OIDC IdP, etc.).
func (p Profile) RequiresExternalService() bool {
	return p == SelfHosted || p == Cloud
}

// DatabaseBackend returns the canonical database driver for the profile.
//
//   - "sqlite"   for local
//   - "postgres" for selfhosted and cloud
func (p Profile) DatabaseBackend() string {
	switch p {
	case Local:
		return "sqlite"
	case SelfHosted, Cloud:
		return "postgres"
	default:
		return "auto"
	}
}

// SessionBackend returns the canonical session / cache backend for the profile.
//
//   - "memory" for local
//   - "redis"  for selfhosted and cloud
func (p Profile) SessionBackend() string {
	switch p {
	case Local:
		return "memory"
	case SelfHosted, Cloud:
		return "redis"
	default:
		return "memory"
	}
}

// IsMultiTenant reports whether the profile requires multi-tenant isolation.
func (p Profile) IsMultiTenant() bool { return p == Cloud }

// TelemetryLevel returns the observability level for the profile.
//   - "minimal"   for local (console logs + health checks)
//   - "full"      for selfhosted and cloud (structured logs, metrics, traces)
func (p Profile) TelemetryLevel() string {
	if p == Local {
		return "minimal"
	}
	return "full"
}

// AuthRequired reports whether authentication is mandatory for the profile.
func (p Profile) AuthRequired() bool { return p != Local }

// DefaultPort returns the default HTTP listen port for the profile.
func (p Profile) DefaultPort() int {
	if p == Local {
		return 9090
	}
	return 0 // 0 means "use env var or further config"
}

// DetectProfile reads the BRAIN_PROFILE environment variable and returns the
// corresponding Profile. When the variable is unset or empty it defaults to
// Local. When the value does not match a known profile the raw value is still
// returned -- callers should call Validate() on the result.
func DetectProfile() Profile {
	raw := strings.TrimSpace(os.Getenv("BRAIN_PROFILE"))
	if raw == "" {
		return Local
	}
	return Profile(strings.ToLower(raw))
}
