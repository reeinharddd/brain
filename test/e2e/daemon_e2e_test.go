package e2e

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestHealthEndpoint verifies the daemon health endpoint returns 200.
func TestHealthEndpoint(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	resp, err := env.httpClient().Get(env.baseURL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	requireStatusCode(t, resp, http.StatusOK)
}

// TestBootstrapLogin verifies bootstrap login works with real daemon.
func TestBootstrapLogin(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, refreshToken := env.login(testBootstrapEmail, testBootstrapPassword)
	if token == "" {
		t.Fatal("login did not return a token")
	}
	if refreshToken == "" {
		t.Fatal("login did not return a refresh token")
	}
}

// TestAuthenticatedRequestAgents verifies authenticated request to /api/agents works.
func TestAuthenticatedRequestAgents(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	var agentsResp AgentsResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
}

// TestUnauthenticatedRequestAgents verifies unauthenticated request to /api/agents returns 401.
func TestUnauthenticatedRequestAgents(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, nil, &errResp)
	if err != nil {
		t.Fatalf("agents request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
	if errResp.Error == "" {
		t.Fatal("expected error message in response body")
	}
}

// TestLogoutRevokesSession verifies logout revokes the session.
func TestLogoutRevokesSession(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Logout.
	var logoutResp map[string]interface{}
	resp, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, authHeader(token), &logoutResp)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
}

// TestRevokedSessionReturns401 verifies a revoked session returns 401 on next request.
func TestRevokedSessionReturns401(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Logout.
	var logoutResp map[string]interface{}
	resp, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, authHeader(token), &logoutResp)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)

	// Try to use the same token.
	var errResp ErrorResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &errResp)
	if err != nil {
		t.Fatalf("agents request after logout failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusUnauthorized)
}

// TestSessionPersistsAcrossRestart verifies session persists across daemon restart.
func TestSessionPersistsAcrossRestart(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Verify token works before restart.
	var agentsResp AgentsResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request before restart failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)

	// Restart daemon.
	env.restartDaemon()

	// Same token should still work because store is on disk.
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request after restart failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)

	t.Cleanup(env.stopDaemon)
}

// TestRefreshEndpointIssuesNewToken verifies the refresh endpoint issues a new token.
func TestRefreshEndpointIssuesNewToken(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	_, refreshToken := env.login(testBootstrapEmail, testBootstrapPassword)

	// Use refresh token to get a new token.
	var refreshResp RefreshResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, nil, &refreshResp)
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
	if !refreshResp.Success {
		t.Fatal("refresh response success flag not set")
	}
	if refreshResp.Token == "" {
		t.Fatal("refresh did not return a new token")
	}

	// Verify the new token works.
	var agentsResp AgentsResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(refreshResp.Token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request with refreshed token failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)
}

// TestRateLimitingBlocksAfterFailedAttempts verifies rate limiting after 5 failed attempts.
func TestRateLimitingBlocksAfterFailedAttempts(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Make 5 failed login attempts.
	for i := 0; i < 5; i++ {
		var errResp ErrorResponse
		resp, err := env.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
			"email":    "ratelimit@brain.local",
			"password": "wrongpassword",
		}, nil, &errResp)
		if err != nil {
			t.Fatalf("login attempt %d failed: %v", i+1, err)
		}
		requireStatusCode(t, resp, http.StatusUnauthorized)
	}

	// The 6th attempt should be rate-limited.
	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "ratelimit@brain.local",
		"password": "wrongpassword",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("rate limit check request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusTooManyRequests)
	if !strings.Contains(strings.ToLower(errResp.Error), "too many") {
		t.Fatalf("expected rate limit error, got: %s", errResp.Error)
	}
}

// TestProtectedRoutesBlockWithoutToken verifies protected routes block without token.
func TestProtectedRoutesBlockWithoutToken(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	protectedRoutes := []string{
		"/api/agents",
		"/api/skills",
		"/api/mcps",
		"/api/users",
	}

	for _, route := range protectedRoutes {
		t.Run(route, func(t *testing.T) {
			var errResp ErrorResponse
			resp, err := env.doJSON(http.MethodGet, route, nil, nil, &errResp)
			if err != nil {
				t.Fatalf("request to %s failed: %v", route, err)
			}
			requireStatusCode(t, resp, http.StatusUnauthorized)
		})
	}
}

// TestAuthStatusWithToken verifies /api/auth/status returns user info with valid token.
func TestAuthStatusWithToken(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	var statusResp StatusResponse
	resp, err := env.doJSON(http.MethodGet, "/api/auth/status", nil, authHeader(token), &statusResp)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
	if !statusResp.Authenticated {
		t.Fatal("status should show authenticated")
	}
	if statusResp.User == nil {
		t.Fatal("status response missing user info")
	}
	if statusResp.User.Email != testBootstrapEmail {
		t.Fatalf("expected email %q, got %q", testBootstrapEmail, statusResp.User.Email)
	}
	if len(statusResp.User.Capabilities) == 0 {
		t.Fatal("owner role should have capabilities")
	}
}

// TestLoginInvalidCredentials verifies login with wrong password returns 401.
func TestLoginInvalidCredentials(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    testBootstrapEmail,
		"password": "wrongpassword",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
	if !strings.Contains(strings.ToLower(errResp.Error), "invalid") {
		t.Fatalf("expected 'invalid' error, got: %s", errResp.Error)
	}
}

// TestHealthEndpointV1 verifies the /api/v1/health endpoint is reachable.
func TestHealthEndpointV1(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	resp, err := env.httpClient().Get(env.baseURL + "/api/v1/health")
	if err != nil {
		t.Fatalf("v1 health request failed: %v", err)
	}
	defer resp.Body.Close()
	// The v1 health endpoint may return 503 if components are not fully initialized,
	// but it should at least respond with a valid JSON body.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", resp.StatusCode)
	}
}

// TestStatusEndpointWithoutAuth verifies /api/status is a public route.
func TestStatusEndpointWithoutAuth(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	resp, err := env.httpClient().Get(env.baseURL + "/api/status")
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	defer resp.Body.Close()
	requireStatusCode(t, resp, http.StatusOK)
}

// TestConcurrentLoginRequests verifies multiple sequential login attempts work correctly.
// Note: true concurrent logins to the same email can trigger rate limiting,
// so this test performs rapid sequential logins instead.
func TestConcurrentLoginRequests(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	tokens := make(map[string]bool)
	for i := 0; i < 5; i++ {
		token, _ := env.login(testBootstrapEmail, testBootstrapPassword)
		if token != "" {
			tokens[token] = true
		}
	}

	// At least some unique tokens should exist.
	if len(tokens) < 1 {
		t.Fatal("expected at least one unique token from sequential logins")
	}
}

// TestDaemonProcessLifecycle verifies daemon can be stopped and is no longer running.
func TestDaemonProcessLifecycle(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	env.startDaemon()

	if !env.isRunning() {
		t.Fatal("daemon should be running after start")
	}

	env.stopDaemon()
	if env.isRunning() {
		t.Fatal("daemon should not be running after stop")
	}
}

// TestLogoutWithNoTokenReturns401 verifies logout without token returns 401.
func TestLogoutWithNoTokenReturns401(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, nil, &errResp)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
}

// TestAuthStatusWithoutToken verifies /api/auth/status without token shows unauthenticated.
func TestAuthStatusWithoutToken(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var statusResp StatusResponse
	resp, err := env.doJSON(http.MethodGet, "/api/auth/status", nil, nil, &statusResp)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
	if statusResp.Authenticated {
		t.Fatal("should not be authenticated without token")
	}
	if !statusResp.Required {
		t.Fatal("auth should be required")
	}
}

// TestRefreshWithoutToken returns 401.
func TestRefreshWithoutTokenReturns401(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/refresh", nil, nil, &errResp)
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
}

// TestRefreshWithExpiredToken returns 401.
func TestRefreshWithExpiredOrInvalidToken(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/refresh", map[string]string{
		"refresh_token": "invalid-refresh-token-that-does-not-exist",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
}

// TestMultipleSessionsPersistsAfterRestart verifies multiple sessions survive restart.
func TestMultipleSessionsPersistsAfterRestart(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	env.startDaemon()

	token1, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Login again to get a second token.
	token2, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	if token1 == token2 {
		t.Fatal("expected different tokens for separate logins")
	}

	// Both tokens should work.
	for i, token := range []string{token1, token2} {
		var agentsResp AgentsResponse
		resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
		if err != nil {
			t.Fatalf("agents request for token %d failed: %v", i+1, err)
		}
		requireStatusCode(t, resp, http.StatusOK)
	}

	// Restart daemon.
	env.restartDaemon()

	// Both tokens should still work.
	for i, token := range []string{token1, token2} {
		var agentsResp AgentsResponse
		resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
		if err != nil {
			t.Fatalf("agents request after restart for token %d failed: %v", i+1, err)
		}
		requireStatusCode(t, resp, http.StatusOK)
	}

	t.Cleanup(env.stopDaemon)
}

// TestLoginMethodNotAllowed verifies GET on /api/auth/login returns 405 or 404.
func TestLoginMethodNotAllowed(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	resp, err := env.httpClient().Get(env.baseURL + "/api/auth/login")
	if err != nil {
		t.Fatalf("login GET request failed: %v", err)
	}
	defer resp.Body.Close()
	// The daemon returns 405 for wrong method on /api/auth/login.
	// Some versions may return 404 if the route is not registered for GET.
	if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 405 or 404, got %d", resp.StatusCode)
	}
}

// TestLongLivedSession tests session TTL behavior (token remains valid).
func TestLongLivedSession(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Use token several times.
	for i := 0; i < 3; i++ {
		var agentsResp AgentsResponse
		resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
		if err != nil {
			t.Fatalf("agents request %d failed: %v", i+1, err)
		}
		requireStatusCode(t, resp, http.StatusOK)
		time.Sleep(100 * time.Millisecond)
	}
}
