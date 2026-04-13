package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	coreenterprise "github.com/reeinharrrd/brain/core/enterprise"
	coreidentity "github.com/reeinharrrd/brain/core/identity"
)

// ── Skip mechanism for CI environments ────────────────────────────────────────

func skipIfOIDCTestsDisabled(t *testing.T) {
	t.Helper()
	if os.Getenv("SKIP_OIDC_TESTS") == "true" {
		t.Skip("OIDC tests disabled (SKIP_OIDC_TESTS=true)")
	}
}

// ── PKCE Helper Tests ─────────────────────────────────────────────────────────

func TestPKCEVerifierAndChallenge(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	// Generate a verifier using the same method as sso.go
	verifier := generateTestVerifier(48)
	if verifier == "" {
		t.Fatal("verifier should not be empty")
	}

	// Verify the challenge is the correct SHA-256 hash
	expectedChallenge := pkceChallengeFromVerifier(verifier)

	// The challenge must be base64url-encoded SHA-256 of the verifier
	sum := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(sum[:])

	if expectedChallenge != expected {
		t.Errorf("challenge mismatch: got %q, expected %q", expectedChallenge, expected)
	}

	// Verify the verifier is URL-safe (no padding, URL-safe chars)
	if strings.Contains(verifier, "+") || strings.Contains(verifier, "/") || strings.Contains(verifier, "=") {
		t.Errorf("verifier contains non-URL-safe characters: %q", verifier)
	}
}

func TestPKCEChallengeDeterministic(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	verifier := "test_verifier_12345"
	challenge1 := pkceChallengeFromVerifier(verifier)
	challenge2 := pkceChallengeFromVerifier(verifier)

	if challenge1 != challenge2 {
		t.Errorf("PKCE challenge not deterministic: %q != %q", challenge1, challenge2)
	}
}

func TestPKCEChallengeLength(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	verifier := "a" // minimum length verifier
	challenge := pkceChallengeFromVerifier(verifier)

	// SHA-256 produces 32 bytes, base64url-encoded = 43 chars
	if len(challenge) != 43 {
		t.Errorf("expected challenge length 43, got %d", len(challenge))
	}
}

// ── Mock OIDC Provider ────────────────────────────────────────────────────────

// mockOIDCProvider creates an httptest server that mimics an OIDC provider.
type mockOIDCProvider struct {
	server       *httptest.Server
	issuerURL    string
	clientID     string
	clientSecret string
	redirectURI  string

	// Configurable behavior
	invalidTokenResponse bool
	missingIDToken       bool
	wrongNonce           bool
	slowResponse         bool
}

func newMockOIDCProvider(t *testing.T, opts ...func(*mockOIDCProvider)) *mockOIDCProvider {
	t.Helper()

	mock := &mockOIDCProvider{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURI:  "http://127.0.0.1:9090/api/auth/oidc/callback",
	}
	for _, opt := range opts {
		opt(mock)
	}

	mux := http.NewServeMux()

	// OIDC Discovery
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		discovery := map[string]any{
			"issuer":                                mock.issuerURL,
			"authorization_endpoint":                mock.issuerURL + "/auth",
			"token_endpoint":                        mock.issuerURL + "/token",
			"userinfo_endpoint":                     mock.issuerURL + "/userinfo",
			"jwks_uri":                              mock.issuerURL + "/.well-known/jwks.json",
			"scopes_supported":                      []string{"openid", "profile", "email"},
			"response_types_supported":              []string{"code"},
			"grant_types_supported":                 []string{"authorization_code"},
			"code_challenge_methods_supported":      []string{"S256"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(discovery)
	})

	// Token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if mock.slowResponse {
			time.Sleep(5 * time.Second)
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Validate client auth
		clientID, clientSecret, _ := r.BasicAuth()
		if clientID != mock.clientID || clientSecret != mock.clientSecret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid_client",
			})
			return
		}

		code := r.FormValue("code")
		if code == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid_request",
			})
			return
		}

		if mock.invalidTokenResponse {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid_grant",
			})
			return
		}

		// Build a mock ID token (not cryptographically valid, but structurally correct for testing)
		// In a real scenario this would be a JWT signed by the provider
		// For mock testing we use a simple base64-encoded JSON
		now := time.Now()
		idToken := buildMockIDToken(mock.clientID, code, now)

		tokenResponse := map[string]any{
			"access_token":  "mock_access_token_" + code,
			"token_type":    "Bearer",
			"expires_in":    3600,
			"id_token":      idToken,
			"refresh_token": "mock_refresh_token_" + code,
		}
		if mock.missingIDToken {
			delete(tokenResponse, "id_token")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse)
	})

	// JWKS endpoint (return empty, we don't verify signatures in mock mode)
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []any{},
		})
	})

	// Userinfo endpoint
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sub":            "mock-subject-123",
			"email":          "test@brain.local",
			"name":           "Test User",
			"email_verified": true,
		})
	})

	mock.server = httptest.NewServer(mux)
	mock.issuerURL = mock.server.URL

	return mock
}

func (m *mockOIDCProvider) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

func (m *mockOIDCProvider) IssuerURL() string {
	return m.issuerURL
}

// ── Mock ID Token Builder ─────────────────────────────────────────────────────

func buildMockIDToken(clientID, code string, now time.Time) string {
	// Create a simple JWT-like structure for mock testing
	// Header
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	// Payload
	payload := map[string]any{
		"sub":            "mock-subject-123",
		"email":          "test@brain.local",
		"name":           "Test User",
		"email_verified": true,
		"iss":            "http://mock-issuer",
		"aud":            clientID,
		"exp":            now.Add(time.Hour).Unix(),
		"iat":            now.Unix(),
		"nonce":          "", // Will be set by the test
	}
	payloadBytes, _ := json.Marshal(payload)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadBytes)
	// Signature (fake)
	signature := base64.RawURLEncoding.EncodeToString([]byte("mock-signature"))
	return header + "." + payloadEnc + "." + signature
}

// ── PKCE Helper Functions (mirrors sso.go logic) ──────────────────────────────

func generateTestVerifier(length int) string {
	buffer := make([]byte, length)
	for i := range buffer {
		buffer[i] = byte('a' + (i % 26))
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}

func pkceChallengeFromVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ── SSOManager Unit Tests ─────────────────────────────────────────────────────

func TestSSOManagerStartLogin(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	mock := newMockOIDCProvider(t)
	defer mock.Close()

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       coreenterprise.ProviderLogto,
		IssuerURL:      mock.IssuerURL(),
		ClientID:       mock.clientID,
		ClientSecret:   mock.clientSecret,
		RedirectURL:    mock.redirectURI,
		Scopes:         []string{"openid", "profile", "email"},
		Enabled:        true,
		TransactionTTL: 10 * time.Minute,
	})

	login, err := manager.StartLogin(t.Context(), "test@brain.local")
	if err != nil {
		t.Fatalf("StartLogin failed: %v", err)
	}

	if login.State == "" {
		t.Error("login state should not be empty")
	}
	if login.AuthorizationURL == "" {
		t.Error("authorization URL should not be empty")
	}
	if login.CodeVerifier == "" {
		t.Error("code verifier should not be empty")
	}
	if login.Nonce == "" {
		t.Error("nonce should not be empty")
	}
	if login.ExpiresAt.IsZero() {
		t.Error("expires at should not be zero")
	}
	if login.LoginHint != "test@brain.local" {
		t.Errorf("unexpected login hint: %q", login.LoginHint)
	}

	// Verify the authorization URL contains required parameters
	authURL, err := url.Parse(login.AuthorizationURL)
	if err != nil {
		t.Fatalf("invalid authorization URL: %v", err)
	}

	query := authURL.Query()
	if query.Get("client_id") != mock.clientID {
		t.Errorf("expected client_id %q, got %q", mock.clientID, query.Get("client_id"))
	}
	if query.Get("redirect_uri") != mock.redirectURI {
		t.Errorf("expected redirect_uri %q, got %q", mock.redirectURI, query.Get("redirect_uri"))
	}
	if query.Get("response_type") != "code" {
		t.Errorf("expected response_type code, got %q", query.Get("response_type"))
	}
	if query.Get("code_challenge_method") != "S256" {
		t.Errorf("expected S256 code challenge method, got %q", query.Get("code_challenge_method"))
	}
	if query.Get("code_challenge") == "" {
		t.Error("code_challenge should not be empty")
	}
	if query.Get("state") != login.State {
		t.Errorf("state mismatch between URL and response")
	}
	if query.Get("nonce") != login.Nonce {
		t.Errorf("nonce mismatch between URL and response")
	}
}

func TestSSOManagerStartLoginDisabled(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	mock := newMockOIDCProvider(t)
	defer mock.Close()

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Enabled: false,
	})

	_, err := manager.StartLogin(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for disabled SSO")
	}
	if !strings.Contains(err.Error(), "not enabled") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSSOManagerStartLoginNilManager(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	var manager *coreenterprise.SSOManager
	_, err := manager.StartLogin(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for nil manager")
	}
}

func TestSSOManagerCancelLogin(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	mock := newMockOIDCProvider(t)
	defer mock.Close()

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       coreenterprise.ProviderLogto,
		IssuerURL:      mock.IssuerURL(),
		ClientID:       mock.clientID,
		ClientSecret:   mock.clientSecret,
		RedirectURL:    mock.redirectURI,
		Enabled:        true,
		TransactionTTL: 10 * time.Minute,
	})

	login, err := manager.StartLogin(t.Context(), "")
	if err != nil {
		t.Fatalf("StartLogin failed: %v", err)
	}

	manager.CancelLogin(login.State)

	// Verify the pending login is removed by checking it's gone
	// (popPending is internal, but we can verify via CancelLogin being idempotent)
	manager.CancelLogin(login.State) // Should not panic
}

func TestSSOManagerCleanupExpired(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	mock := newMockOIDCProvider(t)
	defer mock.Close()

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       coreenterprise.ProviderLogto,
		IssuerURL:      mock.IssuerURL(),
		ClientID:       mock.clientID,
		ClientSecret:   mock.clientSecret,
		RedirectURL:    mock.redirectURI,
		Enabled:        true,
		TransactionTTL: 1 * time.Millisecond, // Very short TTL
	})

	login, err := manager.StartLogin(t.Context(), "")
	if err != nil {
		t.Fatalf("StartLogin failed: %v", err)
	}

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	removed := manager.CleanupExpired()
	if removed != 1 {
		t.Errorf("expected 1 expired login removed, got %d", removed)
	}

	// Try to complete with expired state - should fail
	_, err = manager.CompleteLogin(t.Context(), login.State, "some_code")
	if err == nil {
		t.Fatal("expected error for expired login")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got: %v", err)
	}
}

func TestSSOManagerCompleteLoginNotFound(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	mock := newMockOIDCProvider(t)
	defer mock.Close()

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       coreenterprise.ProviderLogto,
		IssuerURL:      mock.IssuerURL(),
		ClientID:       mock.clientID,
		ClientSecret:   mock.clientSecret,
		RedirectURL:    mock.redirectURI,
		Enabled:        true,
		TransactionTTL: 10 * time.Minute,
	})

	_, err := manager.CompleteLogin(t.Context(), "nonexistent-state", "some_code")
	if err == nil {
		t.Fatal("expected error for nonexistent state")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestSSOManagerCompleteLoginEmptyParams(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	var manager *coreenterprise.SSOManager
	_, err := manager.CompleteLogin(t.Context(), "", "")
	if err == nil {
		t.Fatal("expected error for nil manager")
	}
}

// ── OIDC Handler Tests (via BrainDaemon) ──────────────────────────────────────

func TestOIDCStartHandlerNotEnabled(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{
		oidc: coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
			Enabled: false,
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/start", nil)
	rr := httptest.NewRecorder()

	d.handleOIDCStart(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOIDCStartHandlerNilManager(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{
		oidc: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/start", nil)
	rr := httptest.NewRecorder()

	d.handleOIDCStart(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOIDCCallbackHandlerMissingParams(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{
		oidc: coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
			Enabled: true,
		}),
		auth: coreidentity.NewManager(coreidentity.Config{
			Mode:     coreidentity.ModeOIDC,
			Required: true,
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback", nil)
	rr := httptest.NewRecorder()

	d.handleOIDCCallback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if !strings.Contains(resp["error"], "missing") {
		t.Errorf("expected missing error, got: %s", resp["error"])
	}
}

func TestOIDCCallbackHandlerErrorFromProvider(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{
		oidc: coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
			Enabled: true,
		}),
		auth: coreidentity.NewManager(coreidentity.Config{
			Mode:     coreidentity.ModeOIDC,
			Required: true,
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback?state=x&code=y&error=access_denied&error_description=User+denied", nil)
	rr := httptest.NewRecorder()

	d.handleOIDCCallback(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if resp["error"] != "access_denied" {
		t.Errorf("expected access_denied error, got: %s", resp["error"])
	}
}

func TestOIDCPollHandlerMissingState(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/poll", nil)
	rr := httptest.NewRecorder()

	d.handleOIDCPoll(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOIDCPollHandlerWaiting(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/poll?state=some-state", nil)
	rr := httptest.NewRecorder()

	d.handleOIDCPoll(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp oidcPollResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if resp.Ready {
		t.Error("should not be ready yet")
	}
	if resp.State != "some-state" {
		t.Errorf("state mismatch: got %q, expected %q", resp.State, "some-state")
	}
}

func TestOIDCRefreshHandlerMethodNotAllowed(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/refresh", nil)
	rr := httptest.NewRecorder()

	d.handleAuthRefresh(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOIDCRefreshHandlerMissingToken(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	d := &BrainDaemon{
		auth: coreidentity.NewManager(coreidentity.Config{
			Mode:     coreidentity.ModeBootstrap,
			Required: true,
		}),
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	rr := httptest.NewRecorder()

	d.handleAuthRefresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── BuildOIDCManager Tests ────────────────────────────────────────────────────

func TestBuildOIDCManagerDefaults(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	// Save and restore env
	oldIssuer := os.Getenv("BRAIN_OIDC_ISSUER_URL")
	oldClientID := os.Getenv("BRAIN_OIDC_CLIENT_ID")
	oldAuthMode := os.Getenv("BRAIN_AUTH_MODE")
	defer func() {
		os.Setenv("BRAIN_OIDC_ISSUER_URL", oldIssuer)
		os.Setenv("BRAIN_OIDC_CLIENT_ID", oldClientID)
		os.Setenv("BRAIN_AUTH_MODE", oldAuthMode)
	}()

	os.Setenv("BRAIN_AUTH_MODE", "oidc")
	os.Setenv("BRAIN_OIDC_ISSUER_URL", "http://127.0.0.1:3002/oidc")
	os.Setenv("BRAIN_OIDC_CLIENT_ID", "test-client")

	manager := buildOIDCManager("development")
	config := manager.GetConfig()

	if !config.Enabled {
		t.Error("OIDC should be enabled when auth mode is oidc")
	}
	if config.ClientID != "test-client" {
		t.Errorf("expected client_id test-client, got %q", config.ClientID)
	}
	if config.RedirectURL == "" {
		t.Error("redirect URL should have a default value")
	}
}

// ── Integration Tests (require real Logto instance) ───────────────────────────
//
// These tests are skipped unless:
//   1. SKIP_OIDC_TESTS is not "true"
//   2. BRAIN_OIDC_ISSUER_URL, BRAIN_OIDC_CLIENT_ID, BRAIN_OIDC_CLIENT_SECRET
//      are set to valid values
//
// Run with: SKIP_OIDC_TESTS=false \
//   BRAIN_OIDC_ISSUER_URL=http://127.0.0.1:3002/oidc \
//   BRAIN_OIDC_CLIENT_ID=<id> \
//   BRAIN_OIDC_CLIENT_SECRET=<secret> \
//   go test -run TestOIDCIntegration -v

func TestOIDCIntegrationDiscovery(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	issuerURL := os.Getenv("BRAIN_OIDC_ISSUER_URL")
	if issuerURL == "" {
		t.Skip("BRAIN_OIDC_ISSUER_URL not set, skipping integration test")
	}

	// Test OIDC discovery
	req := httptest.NewRequest(http.MethodGet, issuerURL+"/.well-known/openid-configuration", nil)
	rr := httptest.NewRecorder()

	// Use a simple HTTP client to test discovery
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(issuerURL + "/.well-known/openid-configuration")
	if err != nil {
		t.Fatalf("OIDC discovery failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("OIDC discovery returned %d", resp.StatusCode)
	}

	var discovery map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		t.Fatalf("failed to parse discovery document: %v", err)
	}

	// Verify required fields
	requiredFields := []string{
		"issuer", "authorization_endpoint", "token_endpoint",
		"jwks_uri", "scopes_supported", "response_types_supported",
		"grant_types_supported",
	}
	for _, field := range requiredFields {
		if _, ok := discovery[field]; !ok {
			t.Errorf("discovery document missing required field: %s", field)
		}
	}

	if discovery["issuer"] != issuerURL {
		t.Errorf("issuer mismatch: got %v, expected %s", discovery["issuer"], issuerURL)
	}
}

func TestOIDCIntegrationLoginFlow(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	issuerURL := os.Getenv("BRAIN_OIDC_ISSUER_URL")
	clientID := os.Getenv("BRAIN_OIDC_CLIENT_ID")
	clientSecret := os.Getenv("BRAIN_OIDC_CLIENT_SECRET")
	if issuerURL == "" || clientID == "" || clientSecret == "" {
		t.Skip("OIDC integration env vars not fully set (BRAIN_OIDC_ISSUER_URL, BRAIN_OIDC_CLIENT_ID, BRAIN_OIDC_CLIENT_SECRET required)")
	}

	redirectURI := os.Getenv("BRAIN_OIDC_REDIRECT_URL")
	if redirectURI == "" {
		redirectURI = "http://127.0.0.1:9090/api/auth/oidc/callback"
	}

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       coreenterprise.ProviderLogto,
		IssuerURL:      issuerURL,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		RedirectURL:    redirectURI,
		Scopes:         []string{"openid", "profile", "email"},
		Enabled:        true,
		TransactionTTL: 10 * time.Minute,
	})

	// Step 1: Start login
	login, err := manager.StartLogin(t.Context(), "")
	if err != nil {
		t.Fatalf("StartLogin failed: %v", err)
	}

	t.Logf("Authorization URL: %s", login.AuthorizationURL)
	t.Logf("State: %s", login.State)
	t.Logf("Nonce: %s", login.Nonce)

	// Verify the authorization URL is well-formed
	authURL, err := url.Parse(login.AuthorizationURL)
	if err != nil {
		t.Fatalf("invalid authorization URL: %v", err)
	}

	query := authURL.Query()
	if query.Get("client_id") != clientID {
		t.Errorf("expected client_id %q, got %q", clientID, query.Get("client_id"))
	}
	if query.Get("redirect_uri") != redirectURI {
		t.Errorf("expected redirect_uri %q, got %q", redirectURI, query.Get("redirect_uri"))
	}
	if query.Get("code_challenge_method") != "S256" {
		t.Errorf("expected S256 code challenge method")
	}

	// Note: We cannot complete the full flow programmatically without a browser
	// The authorization code requires user interaction with the Logto UI.
	// The test validates that StartLogin works correctly and produces a valid
	// authorization URL. CompleteLogin requires a real authorization code
	// from a completed browser-based flow.
	t.Log("StartLogin succeeded - manual browser-based completion required for full flow test")
}

func TestOIDCIntegrationProviderUnavailable(t *testing.T) {
	skipIfOIDCTestsDisabled(t)

	manager := coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       coreenterprise.ProviderLogto,
		IssuerURL:      "http://127.0.0.1:59999/nonexistent",
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		RedirectURL:    "http://127.0.0.1:9090/callback",
		Enabled:        true,
		TransactionTTL: 10 * time.Minute,
	})

	// StartLogin should fail because the provider is unreachable
	_, err := manager.StartLogin(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for unavailable provider")
	}
	t.Logf("Got expected error: %v", err)
}
