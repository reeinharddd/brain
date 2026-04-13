// Package profile (config.go) -- profile-aware configuration resolution.
//
// ResolvedConfig merges profile defaults with environment-variable overrides.
// The resolution order is:
//
//  1. Profile defaults (hard-coded per profile)
//  2. Environment variable overrides (BRAIN_* vars always win)
package profile

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Backend       string        // "sqlite" | "postgres"
	DSN           string        // connection string or file path
	MaxOpenConns  int           // 0 = use profile default
	MaxIdleConns  int           // 0 = use profile default
	ConnMaxLifetime time.Duration // 0 = use profile default
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	Mode              string // "bootstrap" | "oidc" | "anonymous"
	Required          bool
	AllowAnonymous    bool
	SSORequired       bool
	BootstrapEmail    string
	BootstrapPassword string
	BootstrapName     string
	BootstrapRole     string
	SessionTTL        time.Duration
	RefreshTTL        time.Duration
	OIDCIssuer        string
	OIDCClientID      string
	OIDCClientSecret  string
	OIDCRedirectURL   string
}

// SessionConfig holds session / cache settings.
type SessionConfig struct {
	Backend      string // "memory" | "sqlite" | "redis"
	RedisURL     string // only used when Backend == "redis"
	RedisPrefix  string
	SessionTTL   time.Duration
}

// TelemetryConfig holds observability settings.
type TelemetryConfig struct {
	Enabled        bool
	Level          string // "minimal" | "full"
	LogFormat      string // "console" | "json"
	LogLevel       string // "info" | "debug" | "warn" | "error"
	OTLPEndpoint   string
	OTLPEnabled    bool
	MetricsEnabled bool
	MetricsPort    int
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string
	Port int
}

// MultiTenantConfig holds multi-tenant isolation settings (only relevant for cloud).
type MultiTenantConfig struct {
	Enabled         bool
	TenantHeader    string // header that carries tenant ID
	IsolationMode   string // "rls" (row-level security) | "schema" | "database"
}

// ProfileConfig is the complete infrastructure configuration for a single profile.
type ProfileConfig struct {
	Profile     Profile
	Database    DatabaseConfig
	Auth        AuthConfig
	Session     SessionConfig
	Telemetry   TelemetryConfig
	Server      ServerConfig
	MultiTenant MultiTenantConfig
}

// ResolvedConfig is the final configuration after merging profile defaults with
// environment variable overrides. It is what the daemon should consume at startup.
type ResolvedConfig struct {
	Profile     Profile
	Database    DatabaseConfig
	Auth        AuthConfig
	Session     SessionConfig
	Telemetry   TelemetryConfig
	Server      ServerConfig
	MultiTenant MultiTenantConfig
}

// LoadConfig returns the full ProfileConfig for a given profile.
// It panics when the profile fails validation -- callers that need graceful
// error handling should call profile.Validate() first.
func LoadConfig(p Profile) *ProfileConfig {
	if err := p.Validate(); err != nil {
		panic(err)
	}

	cfg := &ProfileConfig{Profile: p}

	switch p {
	case Local:
		cfg.Database = DatabaseConfig{
			Backend:       "sqlite",
			DSN:           defaultSQLiteDSN(),
			MaxOpenConns:  1,
			MaxIdleConns:  1,
			ConnMaxLifetime: 0, // unlimited for sqlite
		}
		cfg.Auth = AuthConfig{
			Mode:           "bootstrap",
			Required:       false,
			AllowAnonymous: true,
			SSORequired:    false,
			BootstrapEmail: envOr("BRAIN_AUTH_BOOTSTRAP_EMAIL", "owner@brain.local"),
			BootstrapRole:  "owner",
			SessionTTL:     12 * time.Hour,
			RefreshTTL:     30 * 24 * time.Hour,
		}
		cfg.Session = SessionConfig{
			Backend:    "memory",
			SessionTTL: 12 * time.Hour,
		}
		cfg.Telemetry = TelemetryConfig{
			Enabled:        envBool("BRAIN_TELEMETRY_ENABLED", true),
			Level:          "minimal",
			LogFormat:      "console",
			LogLevel:       "info",
			OTLPEnabled:    false,
			MetricsEnabled: false,
		}
		cfg.Server = ServerConfig{
			Host: "127.0.0.1",
			Port: 9090,
		}
		cfg.MultiTenant = MultiTenantConfig{Enabled: false}

	case SelfHosted:
		cfg.Database = DatabaseConfig{
			Backend:       "postgres",
			DSN:           envOr("BRAIN_DATABASE_DSN", ""),
			MaxOpenConns:  25,
			MaxIdleConns:  5,
			ConnMaxLifetime: 30 * time.Minute,
		}
		cfg.Auth = AuthConfig{
			Mode:           "oidc",
			Required:       true,
			AllowAnonymous: false,
			SSORequired:    true,
			OIDCIssuer:     envOr("BRAIN_AUTH_OIDC_ISSUER", ""),
			OIDCClientID:   envOr("BRAIN_AUTH_OIDC_CLIENT_ID", ""),
			OIDCRedirectURL: envOr("BRAIN_AUTH_OIDC_REDIRECT_URL", "http://127.0.0.1:9090/api/auth/oidc/callback"),
			SessionTTL:     12 * time.Hour,
			RefreshTTL:     30 * 24 * time.Hour,
		}
		cfg.Session = SessionConfig{
			Backend:    "redis",
			RedisURL:   envOr("BRAIN_REDIS_URL", "redis://localhost:6379"),
			RedisPrefix: "brain:",
			SessionTTL: 12 * time.Hour,
		}
		cfg.Telemetry = TelemetryConfig{
			Enabled:        envBool("BRAIN_TELEMETRY_ENABLED", true),
			Level:          "full",
			LogFormat:      "json",
			LogLevel:       "info",
			OTLPEndpoint:   envOr("BRAIN_OTLP_ENDPOINT", "localhost:4318"),
			OTLPEnabled:    envBool("BRAIN_OTEL_ENABLED", false),
			MetricsEnabled: true,
			MetricsPort:    9091,
		}
		cfg.Server = ServerConfig{
			Host: "0.0.0.0",
			Port: envPort("BRAIN_PORT", 0),
		}
		cfg.MultiTenant = MultiTenantConfig{Enabled: false}

	case Cloud:
		cfg.Database = DatabaseConfig{
			Backend:       "postgres",
			DSN:           envOr("BRAIN_DATABASE_DSN", ""),
			MaxOpenConns:  50,
			MaxIdleConns:  10,
			ConnMaxLifetime: 30 * time.Minute,
		}
		cfg.Auth = AuthConfig{
			Mode:           "oidc",
			Required:       true,
			AllowAnonymous: false,
			SSORequired:    true,
			OIDCIssuer:     envOr("BRAIN_AUTH_OIDC_ISSUER", ""),
			OIDCClientID:   envOr("BRAIN_AUTH_OIDC_CLIENT_ID", ""),
			OIDCRedirectURL: envOr("BRAIN_AUTH_OIDC_REDIRECT_URL", ""),
			SessionTTL:     8 * time.Hour,
			RefreshTTL:     7 * 24 * time.Hour,
		}
		cfg.Session = SessionConfig{
			Backend:    "redis",
			RedisURL:   envOr("BRAIN_REDIS_URL", "redis://localhost:6379"),
			RedisPrefix: "brain:",
			SessionTTL: 8 * time.Hour,
		}
		cfg.Telemetry = TelemetryConfig{
			Enabled:        envBool("BRAIN_TELEMETRY_ENABLED", true),
			Level:          "full",
			LogFormat:      "json",
			LogLevel:       "info",
			OTLPEndpoint:   envOr("BRAIN_OTLP_ENDPOINT", "localhost:4318"),
			OTLPEnabled:    envBool("BRAIN_OTEL_ENABLED", true),
			MetricsEnabled: true,
			MetricsPort:    9091,
		}
		cfg.Server = ServerConfig{
			Host: "0.0.0.0",
			Port: envPort("BRAIN_PORT", 0),
		}
		cfg.MultiTenant = MultiTenantConfig{
			Enabled:       true,
			TenantHeader:  "X-Brain-Tenant-ID",
			IsolationMode: "rls",
		}
	}

	return cfg
}

// ResolveConfig is the main entry point for the daemon. It detects the active
// profile, loads its defaults, and applies environment variable overrides to
// produce the final ResolvedConfig.
func ResolveConfig() *ResolvedConfig {
	p := DetectProfile()
	profileCfg := LoadConfig(p)
	return applyEnvOverrides(profileCfg)
}

// applyEnvOverrides copies a ProfileConfig into a ResolvedConfig and patches
// every field that has a corresponding BRAIN_* environment variable set.
func applyEnvOverrides(pc *ProfileConfig) *ResolvedConfig {
	rc := &ResolvedConfig{
		Profile:     pc.Profile,
		Database:    pc.Database,
		Auth:        pc.Auth,
		Session:     pc.Session,
		Telemetry:   pc.Telemetry,
		Server:      pc.Server,
		MultiTenant: pc.MultiTenant,
	}

	// --- Database overrides ---
	if v := os.Getenv("BRAIN_DATABASE_DSN"); v != "" {
		rc.Database.DSN = v
	}
	if v := os.Getenv("BRAIN_DATABASE_BACKEND"); v != "" {
		rc.Database.Backend = v
	}
	if v := os.Getenv("BRAIN_DATABASE_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rc.Database.MaxOpenConns = n
		}
	}
	if v := os.Getenv("BRAIN_DATABASE_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rc.Database.MaxIdleConns = n
		}
	}

	// --- Auth overrides ---
	if v := os.Getenv("BRAIN_AUTH_MODE"); v != "" {
		rc.Auth.Mode = strings.ToLower(v)
	}
	if v := os.Getenv("BRAIN_AUTH_REQUIRED"); v != "" {
		rc.Auth.Required = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("BRAIN_AUTH_SSO_REQUIRED"); v != "" {
		rc.Auth.SSORequired = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("BRAIN_AUTH_BOOTSTRAP_EMAIL"); v != "" {
		rc.Auth.BootstrapEmail = v
	}
	if v := os.Getenv("BRAIN_AUTH_BOOTSTRAP_PASSWORD"); v != "" {
		rc.Auth.BootstrapPassword = v
	}
	if v := os.Getenv("BRAIN_AUTH_BOOTSTRAP_NAME"); v != "" {
		rc.Auth.BootstrapName = v
	}
	if v := os.Getenv("BRAIN_AUTH_BOOTSTRAP_ROLE"); v != "" {
		rc.Auth.BootstrapRole = v
	}
	if v := os.Getenv("BRAIN_AUTH_SESSION_TTL_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil {
			rc.Auth.SessionTTL = time.Duration(h) * time.Hour
		}
	}
	if v := os.Getenv("BRAIN_AUTH_OIDC_ISSUER"); v != "" {
		rc.Auth.OIDCIssuer = v
	}
	if v := os.Getenv("BRAIN_AUTH_OIDC_CLIENT_ID"); v != "" {
		rc.Auth.OIDCClientID = v
	}
	if v := os.Getenv("BRAIN_AUTH_OIDC_CLIENT_SECRET"); v != "" {
		rc.Auth.OIDCClientSecret = v
	}
	if v := os.Getenv("BRAIN_AUTH_OIDC_REDIRECT_URL"); v != "" {
		rc.Auth.OIDCRedirectURL = v
	}

	// --- Session overrides ---
	if v := os.Getenv("BRAIN_SESSION_BACKEND"); v != "" {
		rc.Session.Backend = v
	}
	if v := os.Getenv("BRAIN_REDIS_URL"); v != "" {
		rc.Session.RedisURL = v
	}

	// --- Telemetry overrides ---
	if v := os.Getenv("BRAIN_TELEMETRY_ENABLED"); v != "" {
		rc.Telemetry.Enabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("BRAIN_LOG_FORMAT"); v != "" {
		rc.Telemetry.LogFormat = v
	}
	if v := os.Getenv("BRAIN_LOG_LEVEL"); v != "" {
		rc.Telemetry.LogLevel = v
	}
	if v := os.Getenv("BRAIN_OTLP_ENDPOINT"); v != "" {
		rc.Telemetry.OTLPEndpoint = v
	}
	if v := os.Getenv("BRAIN_OTEL_ENABLED"); v != "" {
		rc.Telemetry.OTLPEnabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("BRAIN_METRICS_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rc.Telemetry.MetricsPort = n
		}
	}

	// --- Server overrides ---
	if v := os.Getenv("BRAIN_HOST"); v != "" {
		rc.Server.Host = v
	}
	if v := os.Getenv("BRAIN_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rc.Server.Port = n
		}
	}

	// --- Multi-tenant overrides ---
	if v := os.Getenv("BRAIN_MULTI_TENANT_ENABLED"); v != "" {
		rc.MultiTenant.Enabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("BRAIN_TENANT_HEADER"); v != "" {
		rc.MultiTenant.TenantHeader = v
	}
	if v := os.Getenv("BRAIN_TENANT_ISOLATION_MODE"); v != "" {
		rc.MultiTenant.IsolationMode = v
	}

	return rc
}

// defaultSQLiteDSN returns a sensible default path for the local SQLite file.
func defaultSQLiteDSN() string {
	if dir := os.Getenv("BRAIN_DATA_DIR"); dir != "" {
		return dir + "/brain.db"
	}
	// Fall back to a standard location.
	if home, err := os.UserHomeDir(); err == nil {
		return home + "/.brain/brain.db"
	}
	return "brain.db"
}

// envOr returns the value of the environment variable when set and non-empty,
// otherwise the provided default.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envBool parses a boolean from the environment variable, falling back to the
// provided default when the variable is unset.
func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return strings.ToLower(v) == "true"
}

// envPort parses an integer port from the environment variable, falling back to
// the provided default when the variable is unset or invalid.
func envPort(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return fallback
}
