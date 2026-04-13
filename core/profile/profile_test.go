package profile

import (
	"os"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clearEnv unsets every BRAIN_* variable used by the profile package so that
// tests run in a predictable environment. It returns a cleanup function.
func clearEnv() func() {
	keys := []string{
		"BRAIN_PROFILE",
		"BRAIN_DATABASE_DSN",
		"BRAIN_DATABASE_BACKEND",
		"BRAIN_DATABASE_MAX_OPEN_CONNS",
		"BRAIN_DATABASE_MAX_IDLE_CONNS",
		"BRAIN_AUTH_MODE",
		"BRAIN_AUTH_REQUIRED",
		"BRAIN_AUTH_SSO_REQUIRED",
		"BRAIN_AUTH_BOOTSTRAP_EMAIL",
		"BRAIN_AUTH_BOOTSTRAP_PASSWORD",
		"BRAIN_AUTH_BOOTSTRAP_NAME",
		"BRAIN_AUTH_BOOTSTRAP_ROLE",
		"BRAIN_AUTH_SESSION_TTL_HOURS",
		"BRAIN_AUTH_OIDC_ISSUER",
		"BRAIN_AUTH_OIDC_CLIENT_ID",
		"BRAIN_AUTH_OIDC_CLIENT_SECRET",
		"BRAIN_AUTH_OIDC_REDIRECT_URL",
		"BRAIN_SESSION_BACKEND",
		"BRAIN_REDIS_URL",
		"BRAIN_TELEMETRY_ENABLED",
		"BRAIN_LOG_FORMAT",
		"BRAIN_LOG_LEVEL",
		"BRAIN_OTLP_ENDPOINT",
		"BRAIN_OTEL_ENABLED",
		"BRAIN_METRICS_PORT",
		"BRAIN_HOST",
		"BRAIN_PORT",
		"BRAIN_MULTI_TENANT_ENABLED",
		"BRAIN_TENANT_HEADER",
		"BRAIN_TENANT_ISOLATION_MODE",
		"BRAIN_DATA_DIR",
	}
	old := make(map[string]string)
	for _, k := range keys {
		old[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	return func() {
		for k, v := range old {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Profile.Validate
// ---------------------------------------------------------------------------

func TestProfileValidate_KnownProfiles(t *testing.T) {
	for _, p := range AllProfiles() {
		if err := p.Validate(); err != nil {
			t.Errorf("profile %q should be valid: %v", p, err)
		}
	}
}

func TestProfileValidate_UnknownProfile(t *testing.T) {
	p := Profile("nonexistent")
	if err := p.Validate(); err == nil {
		t.Error("expected error for unknown profile, got nil")
	}
}

// ---------------------------------------------------------------------------
// Profile.RequiresExternalService
// ---------------------------------------------------------------------------

func TestProfileRequiresExternalService(t *testing.T) {
	tests := []struct {
		profile Profile
		want    bool
	}{
		{Local, false},
		{SelfHosted, true},
		{Cloud, true},
	}
	for _, tt := range tests {
		if got := tt.profile.RequiresExternalService(); got != tt.want {
			t.Errorf("%s.RequiresExternalService() = %v, want %v", tt.profile, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Profile.DatabaseBackend
// ---------------------------------------------------------------------------

func TestProfileDatabaseBackend(t *testing.T) {
	tests := []struct {
		profile Profile
		want    string
	}{
		{Local, "sqlite"},
		{SelfHosted, "postgres"},
		{Cloud, "postgres"},
		{Profile("unknown"), "auto"},
	}
	for _, tt := range tests {
		if got := tt.profile.DatabaseBackend(); got != tt.want {
			t.Errorf("%s.DatabaseBackend() = %q, want %q", tt.profile, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Profile.SessionBackend
// ---------------------------------------------------------------------------

func TestProfileSessionBackend(t *testing.T) {
	tests := []struct {
		profile Profile
		want    string
	}{
		{Local, "memory"},
		{SelfHosted, "redis"},
		{Cloud, "redis"},
		{Profile("unknown"), "memory"},
	}
	for _, tt := range tests {
		if got := tt.profile.SessionBackend(); got != tt.want {
			t.Errorf("%s.SessionBackend() = %q, want %q", tt.profile, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Profile.IsMultiTenant
// ---------------------------------------------------------------------------

func TestProfileIsMultiTenant(t *testing.T) {
	if Local.IsMultiTenant() {
		t.Error("Local should not be multi-tenant")
	}
	if SelfHosted.IsMultiTenant() {
		t.Error("SelfHosted should not be multi-tenant by default")
	}
	if !Cloud.IsMultiTenant() {
		t.Error("Cloud should be multi-tenant")
	}
}

// ---------------------------------------------------------------------------
// Profile.TelemetryLevel
// ---------------------------------------------------------------------------

func TestProfileTelemetryLevel(t *testing.T) {
	tests := []struct {
		profile Profile
		want    string
	}{
		{Local, "minimal"},
		{SelfHosted, "full"},
		{Cloud, "full"},
	}
	for _, tt := range tests {
		if got := tt.profile.TelemetryLevel(); got != tt.want {
			t.Errorf("%s.TelemetryLevel() = %q, want %q", tt.profile, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Profile.AuthRequired
// ---------------------------------------------------------------------------

func TestProfileAuthRequired(t *testing.T) {
	tests := []struct {
		profile Profile
		want    bool
	}{
		{Local, false},
		{SelfHosted, true},
		{Cloud, true},
	}
	for _, tt := range tests {
		if got := tt.profile.AuthRequired(); got != tt.want {
			t.Errorf("%s.AuthRequired() = %v, want %v", tt.profile, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Profile.DefaultPort
// ---------------------------------------------------------------------------

func TestProfileDefaultPort(t *testing.T) {
	if got := Local.DefaultPort(); got != 9090 {
		t.Errorf("Local.DefaultPort() = %d, want 9090", got)
	}
	if got := SelfHosted.DefaultPort(); got != 0 {
		t.Errorf("SelfHosted.DefaultPort() = %d, want 0", got)
	}
	if got := Cloud.DefaultPort(); got != 0 {
		t.Errorf("Cloud.DefaultPort() = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// DetectProfile
// ---------------------------------------------------------------------------

func TestDetectProfile_Default(t *testing.T) {
	defer clearEnv()()
	// BRAIN_PROFILE is unset
	got := DetectProfile()
	if got != Local {
		t.Errorf("DetectProfile() = %q, want %q", got, Local)
	}
}

func TestDetectProfile_FromEnv(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_PROFILE", "cloud")
	got := DetectProfile()
	if got != Cloud {
		t.Errorf("DetectProfile() = %q, want %q", got, Cloud)
	}
}

func TestDetectProfile_CaseInsensitive(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_PROFILE", "SELFHOSTED")
	got := DetectProfile()
	if got != SelfHosted {
		t.Errorf("DetectProfile() = %q, want %q", got, SelfHosted)
	}
}

func TestDetectProfile_UnknownValue(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_PROFILE", "enterprise")
	got := DetectProfile()
	if got != Profile("enterprise") {
		t.Errorf("DetectProfile() = %q, want %q", got, "enterprise")
	}
	// But Validate should reject it
	if err := got.Validate(); err == nil {
		t.Error("unknown profile from env should fail validation")
	}
}

// ---------------------------------------------------------------------------
// LoadConfig -- local profile defaults
// ---------------------------------------------------------------------------

func TestLoadConfig_Local(t *testing.T) {
	defer clearEnv()()
	cfg := LoadConfig(Local)

	if cfg.Profile != Local {
		t.Errorf("profile = %q, want %q", cfg.Profile, Local)
	}
	if cfg.Database.Backend != "sqlite" {
		t.Errorf("database backend = %q, want %q", cfg.Database.Backend, "sqlite")
	}
	if cfg.Database.MaxOpenConns != 1 {
		t.Errorf("max open conns = %d, want 1", cfg.Database.MaxOpenConns)
	}
	if cfg.Auth.Mode != "bootstrap" {
		t.Errorf("auth mode = %q, want %q", cfg.Auth.Mode, "bootstrap")
	}
	if cfg.Auth.Required {
		t.Error("auth should not be required for local")
	}
	if !cfg.Auth.AllowAnonymous {
		t.Error("anonymous should be allowed for local")
	}
	if cfg.Session.Backend != "memory" {
		t.Errorf("session backend = %q, want %q", cfg.Session.Backend, "memory")
	}
	if cfg.Telemetry.Level != "minimal" {
		t.Errorf("telemetry level = %q, want %q", cfg.Telemetry.Level, "minimal")
	}
	if cfg.Telemetry.LogFormat != "console" {
		t.Errorf("log format = %q, want %q", cfg.Telemetry.LogFormat, "console")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("server port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("server host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.MultiTenant.Enabled {
		t.Error("multi-tenant should be disabled for local")
	}
}

// ---------------------------------------------------------------------------
// LoadConfig -- selfhosted profile defaults
// ---------------------------------------------------------------------------

func TestLoadConfig_SelfHosted(t *testing.T) {
	defer clearEnv()()
	cfg := LoadConfig(SelfHosted)

	if cfg.Profile != SelfHosted {
		t.Errorf("profile = %q, want %q", cfg.Profile, SelfHosted)
	}
	if cfg.Database.Backend != "postgres" {
		t.Errorf("database backend = %q, want %q", cfg.Database.Backend, "postgres")
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("max open conns = %d, want 25", cfg.Database.MaxOpenConns)
	}
	if cfg.Auth.Mode != "oidc" {
		t.Errorf("auth mode = %q, want %q", cfg.Auth.Mode, "oidc")
	}
	if !cfg.Auth.Required {
		t.Error("auth should be required for selfhosted")
	}
	if !cfg.Auth.SSORequired {
		t.Error("SSO should be required for selfhosted")
	}
	if cfg.Session.Backend != "redis" {
		t.Errorf("session backend = %q, want %q", cfg.Session.Backend, "redis")
	}
	if cfg.Telemetry.Level != "full" {
		t.Errorf("telemetry level = %q, want %q", cfg.Telemetry.Level, "full")
	}
	if cfg.Telemetry.LogFormat != "json" {
		t.Errorf("log format = %q, want %q", cfg.Telemetry.LogFormat, "json")
	}
	if cfg.MultiTenant.Enabled {
		t.Error("multi-tenant should be disabled for selfhosted")
	}
}

// ---------------------------------------------------------------------------
// LoadConfig -- cloud profile defaults
// ---------------------------------------------------------------------------

func TestLoadConfig_Cloud(t *testing.T) {
	defer clearEnv()()
	cfg := LoadConfig(Cloud)

	if cfg.Profile != Cloud {
		t.Errorf("profile = %q, want %q", cfg.Profile, Cloud)
	}
	if cfg.Database.Backend != "postgres" {
		t.Errorf("database backend = %q, want %q", cfg.Database.Backend, "postgres")
	}
	if cfg.Database.MaxOpenConns != 50 {
		t.Errorf("max open conns = %d, want 50", cfg.Database.MaxOpenConns)
	}
	if cfg.Auth.Mode != "oidc" {
		t.Errorf("auth mode = %q, want %q", cfg.Auth.Mode, "oidc")
	}
	if !cfg.Auth.Required {
		t.Error("auth should be required for cloud")
	}
	if cfg.Session.Backend != "redis" {
		t.Errorf("session backend = %q, want %q", cfg.Session.Backend, "redis")
	}
	if cfg.Telemetry.Level != "full" {
		t.Errorf("telemetry level = %q, want %q", cfg.Telemetry.Level, "full")
	}
	if !cfg.Telemetry.OTLPEnabled {
		t.Error("OTLP should be enabled by default for cloud")
	}
	if !cfg.MultiTenant.Enabled {
		t.Error("multi-tenant should be enabled for cloud")
	}
	if cfg.MultiTenant.TenantHeader != "X-Brain-Tenant-ID" {
		t.Errorf("tenant header = %q, want %q", cfg.MultiTenant.TenantHeader, "X-Brain-Tenant-ID")
	}
	if cfg.MultiTenant.IsolationMode != "rls" {
		t.Errorf("isolation mode = %q, want %q", cfg.MultiTenant.IsolationMode, "rls")
	}
	if cfg.Server.Port != 0 {
		t.Errorf("server port should default to 0 (env-driven) for cloud, got %d", cfg.Server.Port)
	}
}

// ---------------------------------------------------------------------------
// LoadConfig -- unknown profile panics
// ---------------------------------------------------------------------------

func TestLoadConfig_UnknownPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("LoadConfig with unknown profile should panic")
		}
	}()
	LoadConfig(Profile("bogus"))
}

// ---------------------------------------------------------------------------
// Env var overrides
// ---------------------------------------------------------------------------

func TestResolveConfig_ProfileFromEnv(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_PROFILE", "cloud")
	rc := ResolveConfig()
	if rc.Profile != Cloud {
		t.Errorf("resolved profile = %q, want %q", rc.Profile, Cloud)
	}
}

func TestResolveConfig_PortOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_PORT", "8080")
	rc := ResolveConfig()
	if rc.Server.Port != 8080 {
		t.Errorf("server port = %d, want 8080", rc.Server.Port)
	}
}

func TestResolveConfig_HostOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_HOST", "::")
	rc := ResolveConfig()
	if rc.Server.Host != "::" {
		t.Errorf("server host = %q, want %q", rc.Server.Host, "::")
	}
}

func TestResolveConfig_DatabaseDSNOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_DATABASE_DSN", "postgres://user:pass@db:5432/brain")
	rc := ResolveConfig()
	if rc.Database.DSN != "postgres://user:pass@db:5432/brain" {
		t.Errorf("database DSN = %q, want %q", rc.Database.DSN, "postgres://user:pass@db:5432/brain")
	}
}

func TestResolveConfig_DatabaseBackendOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_DATABASE_BACKEND", "sqlite")
	rc := ResolveConfig()
	if rc.Database.Backend != "sqlite" {
		t.Errorf("database backend = %q, want %q", rc.Database.Backend, "sqlite")
	}
}

func TestResolveConfig_AuthModeOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_AUTH_MODE", "anonymous")
	rc := ResolveConfig()
	if rc.Auth.Mode != "anonymous" {
		t.Errorf("auth mode = %q, want %q", rc.Auth.Mode, "anonymous")
	}
}

func TestResolveConfig_AuthRequiredOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_AUTH_REQUIRED", "true")
	rc := ResolveConfig()
	if !rc.Auth.Required {
		t.Error("auth should be required after override")
	}
}

func TestResolveConfig_TelemetryOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_TELEMETRY_ENABLED", "false")
	rc := ResolveConfig()
	if rc.Telemetry.Enabled {
		t.Error("telemetry should be disabled after override")
	}
}

func TestResolveConfig_RedisURLOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_REDIS_URL", "redis://my-redis:6380")
	rc := ResolveConfig()
	if rc.Session.RedisURL != "redis://my-redis:6380" {
		t.Errorf("redis URL = %q, want %q", rc.Session.RedisURL, "redis://my-redis:6380")
	}
}

func TestResolveConfig_SessionBackendOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_SESSION_BACKEND", "sqlite")
	rc := ResolveConfig()
	if rc.Session.Backend != "sqlite" {
		t.Errorf("session backend = %q, want %q", rc.Session.Backend, "sqlite")
	}
}

func TestResolveConfig_OTLPEndpointOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_OTLP_ENDPOINT", "collector:4317")
	rc := ResolveConfig()
	if rc.Telemetry.OTLPEndpoint != "collector:4317" {
		t.Errorf("OTLP endpoint = %q, want %q", rc.Telemetry.OTLPEndpoint, "collector:4317")
	}
}

func TestResolveConfig_LogFormatOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_LOG_FORMAT", "json")
	rc := ResolveConfig()
	if rc.Telemetry.LogFormat != "json" {
		t.Errorf("log format = %q, want %q", rc.Telemetry.LogFormat, "json")
	}
}

func TestResolveConfig_MultiTenantOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_MULTI_TENANT_ENABLED", "true")
	os.Setenv("BRAIN_TENANT_HEADER", "X-Custom-Tenant")
	os.Setenv("BRAIN_TENANT_ISOLATION_MODE", "schema")
	rc := ResolveConfig()
	if !rc.MultiTenant.Enabled {
		t.Error("multi-tenant should be enabled after override")
	}
	if rc.MultiTenant.TenantHeader != "X-Custom-Tenant" {
		t.Errorf("tenant header = %q, want %q", rc.MultiTenant.TenantHeader, "X-Custom-Tenant")
	}
	if rc.MultiTenant.IsolationMode != "schema" {
		t.Errorf("isolation mode = %q, want %q", rc.MultiTenant.IsolationMode, "schema")
	}
}

func TestResolveConfig_MetricsPortOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_METRICS_PORT", "9100")
	rc := ResolveConfig()
	if rc.Telemetry.MetricsPort != 9100 {
		t.Errorf("metrics port = %d, want 9100", rc.Telemetry.MetricsPort)
	}
}

func TestResolveConfig_SessionTTLOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_AUTH_SESSION_TTL_HOURS", "24")
	rc := ResolveConfig()
	if rc.Auth.SessionTTL != 24*time.Hour {
		t.Errorf("session TTL = %v, want %v", rc.Auth.SessionTTL, 24*time.Hour)
	}
}

func TestResolveConfig_BootstrapCredentialsOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_EMAIL", "admin@corp.io")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_PASSWORD", "s3cret")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_NAME", "Admin User")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_ROLE", "admin")
	rc := ResolveConfig()
	if rc.Auth.BootstrapEmail != "admin@corp.io" {
		t.Errorf("bootstrap email = %q, want %q", rc.Auth.BootstrapEmail, "admin@corp.io")
	}
	if rc.Auth.BootstrapPassword != "s3cret" {
		t.Errorf("bootstrap password = %q, want %q", rc.Auth.BootstrapPassword, "s3cret")
	}
	if rc.Auth.BootstrapName != "Admin User" {
		t.Errorf("bootstrap name = %q, want %q", rc.Auth.BootstrapName, "Admin User")
	}
	if rc.Auth.BootstrapRole != "admin" {
		t.Errorf("bootstrap role = %q, want %q", rc.Auth.BootstrapRole, "admin")
	}
}

func TestResolveConfig_OIDCOverrides(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_AUTH_OIDC_ISSUER", "https://idp.corp.io/oidc")
	os.Setenv("BRAIN_AUTH_OIDC_CLIENT_ID", "client-123")
	os.Setenv("BRAIN_AUTH_OIDC_CLIENT_SECRET", "secret-456")
	os.Setenv("BRAIN_AUTH_OIDC_REDIRECT_URL", "https://brain.corp.io/callback")
	rc := ResolveConfig()
	if rc.Auth.OIDCIssuer != "https://idp.corp.io/oidc" {
		t.Errorf("OIDC issuer = %q, want %q", rc.Auth.OIDCIssuer, "https://idp.corp.io/oidc")
	}
	if rc.Auth.OIDCClientID != "client-123" {
		t.Errorf("OIDC client ID = %q, want %q", rc.Auth.OIDCClientID, "client-123")
	}
	if rc.Auth.OIDCClientSecret != "secret-456" {
		t.Errorf("OIDC client secret = %q, want %q", rc.Auth.OIDCClientSecret, "secret-456")
	}
	if rc.Auth.OIDCRedirectURL != "https://brain.corp.io/callback" {
		t.Errorf("OIDC redirect URL = %q, want %q", rc.Auth.OIDCRedirectURL, "https://brain.corp.io/callback")
	}
}

func TestResolveConfig_LogLevelOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_LOG_LEVEL", "debug")
	rc := ResolveConfig()
	if rc.Telemetry.LogLevel != "debug" {
		t.Errorf("log level = %q, want %q", rc.Telemetry.LogLevel, "debug")
	}
}

func TestResolveConfig_DatabaseConnPoolOverride(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_DATABASE_MAX_OPEN_CONNS", "100")
	os.Setenv("BRAIN_DATABASE_MAX_IDLE_CONNS", "20")
	rc := ResolveConfig()
	if rc.Database.MaxOpenConns != 100 {
		t.Errorf("max open conns = %d, want 100", rc.Database.MaxOpenConns)
	}
	if rc.Database.MaxIdleConns != 20 {
		t.Errorf("max idle conns = %d, want 20", rc.Database.MaxIdleConns)
	}
}

// ---------------------------------------------------------------------------
// Config merging: verify that non-overridden fields retain profile defaults
// ---------------------------------------------------------------------------

func TestResolveConfig_MergingPreservesProfileDefaults(t *testing.T) {
	defer clearEnv()()
	os.Setenv("BRAIN_PORT", "7070") // override only the port
	os.Setenv("BRAIN_PROFILE", "selfhosted")

	rc := ResolveConfig()

	// Overridden
	if rc.Server.Port != 7070 {
		t.Errorf("port = %d, want 7070", rc.Server.Port)
	}
	// Still from profile defaults
	if rc.Database.Backend != "postgres" {
		t.Errorf("database backend = %q, want %q", rc.Database.Backend, "postgres")
	}
	if rc.Database.MaxOpenConns != 25 {
		t.Errorf("max open conns = %d, want 25", rc.Database.MaxOpenConns)
	}
	if rc.Auth.Mode != "oidc" {
		t.Errorf("auth mode = %q, want %q", rc.Auth.Mode, "oidc")
	}
	if rc.Session.Backend != "redis" {
		t.Errorf("session backend = %q, want %q", rc.Session.Backend, "redis")
	}
	if rc.Telemetry.LogFormat != "json" {
		t.Errorf("log format = %q, want %q", rc.Telemetry.LogFormat, "json")
	}
	if rc.MultiTenant.Enabled {
		t.Error("multi-tenant should still be disabled for selfhosted")
	}
}

// ---------------------------------------------------------------------------
// Test that LoadConfig panics on invalid profile
// ---------------------------------------------------------------------------

func TestAllProfiles_ContainsExpected(t *testing.T) {
	all := AllProfiles()
	if len(all) != 3 {
		t.Fatalf("AllProfiles() returned %d profiles, want 3", len(all))
	}
	expected := map[Profile]bool{Local: true, SelfHosted: true, Cloud: true}
	for _, p := range all {
		if !expected[p] {
			t.Errorf("unexpected profile in AllProfiles(): %q", p)
		}
	}
}

// ---------------------------------------------------------------------------
// String representation
// ---------------------------------------------------------------------------

func TestProfileString(t *testing.T) {
	tests := []struct {
		profile Profile
		want    string
	}{
		{Local, "local"},
		{SelfHosted, "selfhosted"},
		{Cloud, "cloud"},
	}
	for _, tt := range tests {
		if got := tt.profile.String(); got != tt.want {
			t.Errorf("%s.String() = %q, want %q", tt.profile, got, tt.want)
		}
	}
}
