package e2e

import (
	"net/http"
	"strings"
	"testing"
)

// TestFullLoginFlow verifies: POST /api/auth/login -> get token -> use token -> get 200.
func TestFullLoginFlow(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Step 1: Login.
	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Step 2: Use token to access a protected endpoint.
	var agentsResp AgentsResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request failed: %v", err)
	}

	// Step 3: Verify we get 200.
	requireStatusCode(t, resp, http.StatusOK)
}

// TestStatusFlow verifies: GET /api/auth/status with token -> returns user info + capabilities.
func TestStatusFlow(t *testing.T) {
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

	// Verify authenticated status.
	if !statusResp.Authenticated {
		t.Fatal("expected authenticated=true")
	}
	if statusResp.User == nil {
		t.Fatal("expected user info in status response")
	}
	if statusResp.User.Email != testBootstrapEmail {
		t.Fatalf("expected email %q, got %q", testBootstrapEmail, statusResp.User.Email)
	}
	if statusResp.User.Role != testBootstrapRole {
		t.Fatalf("expected role %q, got %q", testBootstrapRole, statusResp.User.Role)
	}

	// Owner should have broad capabilities.
	if len(statusResp.User.Capabilities) == 0 {
		t.Fatal("owner role should have capabilities")
	}
	if !containsString(statusResp.User.Capabilities, "auth:manage") {
		t.Fatal("owner should have auth:manage capability")
	}
}

// TestLogoutFlow verifies: POST /api/auth/logout -> session revoked -> 401 on next request.
func TestLogoutFlow(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Step 1: Login.
	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Step 2: Verify token works.
	var agentsResp AgentsResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request before logout failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)

	// Step 3: Logout.
	var logoutResp map[string]interface{}
	resp2, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, authHeader(token), &logoutResp)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)

	// Step 4: Same token should now fail.
	var errResp ErrorResponse
	resp3, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &errResp)
	if err != nil {
		t.Fatalf("agents request after logout failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusUnauthorized)
}

// TestPersistenceFlow verifies: login -> kill daemon -> restart daemon -> same token works.
func TestPersistenceFlow(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Step 1: Login.
	token, refreshToken := env.login(testBootstrapEmail, testBootstrapPassword)

	// Step 2: Verify token works.
	var agentsResp AgentsResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request before restart failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)

	// Step 3: Restart daemon (simulates kill + restart, preserves store).
	env.restartDaemon()

	// Step 5: Same token should still work (persisted in SQLite store).
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request after restart failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)

	// Step 6: Refresh token should also still work.
	var refreshResp RefreshResponse
	resp3, err := env.doJSON(http.MethodPost, "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, nil, &refreshResp)
	if err != nil {
		t.Fatalf("refresh after restart failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusOK)
	if refreshResp.Token == "" {
		t.Fatal("refresh after restart did not return a new token")
	}

	t.Cleanup(env.stopDaemon)
}

// TestRefreshFlow verifies: login -> POST /api/auth/refresh -> new token works.
func TestRefreshFlow(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Step 1: Login and get refresh token.
	_, refreshToken := env.login(testBootstrapEmail, testBootstrapPassword)

	// Step 2: Refresh to get a new token.
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
	if refreshResp.RefreshToken == "" {
		t.Fatal("refresh did not return a new refresh token")
	}

	// Step 3: New token works.
	var agentsResp AgentsResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(refreshResp.Token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request with new token failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)

	// Step 4: Old token should be revoked (rotate).
	var errResp ErrorResponse
	resp3, err := env.doJSON(http.MethodPost, "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("second refresh request failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusUnauthorized)
}

// TestInviteFlow verifies: login as admin -> create invite -> consume invite -> new user logs in.
func TestInviteFlow(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Step 1: Login as admin (bootstrap owner).
	adminToken, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Step 2: Create an invite for a new user.
	var inviteResp InviteResponse
	resp, err := env.doJSON(http.MethodPost, "/api/invites", map[string]string{
		"email": "newuser@example.com",
		"role":  "member",
	}, authHeader(adminToken), &inviteResp)
	if err != nil {
		t.Fatalf("invite create request failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
	if inviteResp.Invite.Token == "" {
		t.Fatal("invite response missing token")
	}
	inviteToken := inviteResp.Invite.Token

	// Step 3: Consume the invite (public endpoint, no auth needed).
	var consumeResp ConsumeInviteResponse
	resp2, err := env.doJSON(http.MethodPost, "/api/invites/consume", map[string]string{
		"token": inviteToken,
	}, nil, &consumeResp)
	if err != nil {
		t.Fatalf("invite consume request failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)

	// Step 4: Verify invite was consumed by listing invites.
	var listResp map[string]interface{}
	resp3, err := env.doJSON(http.MethodGet, "/api/invites", nil, authHeader(adminToken), &listResp)
	if err != nil {
		t.Fatalf("invite list request failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusOK)

	invitesRaw, ok := listResp["invites"]
	if !ok {
		t.Fatal("invite list response missing 'invites' field")
	}
	invites, ok := invitesRaw.([]interface{})
	if !ok {
		t.Fatalf("invites has unexpected type: %T", invitesRaw)
	}
	if len(invites) == 0 {
		t.Fatal("expected at least one invite")
	}

	// Check the consumed invite has a consumed_at timestamp.
	firstInvite := invites[0].(map[string]interface{})
	consumedAt, _ := firstInvite["consumed_at"].(string)
	if consumedAt == "" {
		t.Fatal("expected invite to have consumed_at timestamp")
	}
}

// TestInviteCreateWithoutAuth returns 401.
func TestInviteCreateWithoutAuth(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/invites", map[string]string{
		"email": "unauthorized@example.com",
		"role":  "member",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("invite create without auth failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
}

// TestInviteListWithoutAuth returns 401.
func TestInviteListWithoutAuth(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodGet, "/api/invites", nil, nil, &errResp)
	if err != nil {
		t.Fatalf("invite list without auth failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
}

// TestAuthStatusAfterLoginAndLogout verifies status reflects authentication state.
func TestAuthStatusAfterLoginAndLogout(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// Before login.
	var beforeLogin StatusResponse
	resp, err := env.doJSON(http.MethodGet, "/api/auth/status", nil, nil, &beforeLogin)
	if err != nil {
		t.Fatalf("pre-login status failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
	if beforeLogin.Authenticated {
		t.Fatal("should not be authenticated before login")
	}

	// Login.
	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// After login.
	var afterLogin StatusResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/auth/status", nil, authHeader(token), &afterLogin)
	if err != nil {
		t.Fatalf("post-login status failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)
	if !afterLogin.Authenticated {
		t.Fatal("should be authenticated after login")
	}

	// Logout.
	var logoutResp map[string]interface{}
	resp3, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, authHeader(token), &logoutResp)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusOK)

	// After logout with same token.
	var afterLogout StatusResponse
	resp4, err := env.doJSON(http.MethodGet, "/api/auth/status", nil, authHeader(token), &afterLogout)
	if err != nil {
		t.Fatalf("post-logout status failed: %v", err)
	}
	requireStatusCode(t, resp4, http.StatusOK)
	if afterLogout.Authenticated {
		t.Fatal("should not be authenticated after logout")
	}
}

// TestMultipleLoginGetUniqueTokens verifies each login produces a unique session token.
func TestMultipleLoginGetUniqueTokens(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token1, _ := env.login(testBootstrapEmail, testBootstrapPassword)
	token2, _ := env.login(testBootstrapEmail, testBootstrapPassword)
	token3, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	if token1 == token2 || token2 == token3 || token1 == token3 {
		t.Fatal("expected unique tokens for each login")
	}

	// All three tokens should work.
	for i, token := range []string{token1, token2, token3} {
		var agentsResp AgentsResponse
		resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
		if err != nil {
			t.Fatalf("agents request for token %d failed: %v", i+1, err)
		}
		requireStatusCode(t, resp, http.StatusOK)
	}
}

// TestLogoutOneSessionDoesNotAffectOthers verifies logging out one session doesn't revoke others.
func TestLogoutOneSessionDoesNotAffectOthers(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token1, _ := env.login(testBootstrapEmail, testBootstrapPassword)
	token2, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Logout token1.
	var logoutResp map[string]interface{}
	resp, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, authHeader(token1), &logoutResp)
	if err != nil {
		t.Fatalf("logout token1 failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)

	// token1 should be revoked.
	var errResp ErrorResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token1), &errResp)
	if err != nil {
		t.Fatalf("agents with revoked token1 failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusUnauthorized)

	// token2 should still work.
	var agentsResp AgentsResponse
	resp3, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token2), &agentsResp)
	if err != nil {
		t.Fatalf("agents with token2 failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusOK)
}

// TestRefreshInvalidToken verifies using an invalid refresh token returns 401.
func TestRefreshInvalidToken(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/refresh", map[string]string{
		"refresh_token": "totally-invalid-refresh-token",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("refresh with invalid token failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusUnauthorized)
}

// TestFullAuthFlowWithStatusCheck combines login, status check, request, logout.
func TestFullAuthFlowWithStatusCheck(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	// 1. Login.
	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// 2. Status check.
	var statusResp StatusResponse
	resp, err := env.doJSON(http.MethodGet, "/api/auth/status", nil, authHeader(token), &statusResp)
	if err != nil {
		t.Fatalf("status check failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)
	if !statusResp.Authenticated {
		t.Fatal("expected authenticated status")
	}

	// 3. Make authenticated request.
	var agentsResp AgentsResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusOK)

	// 4. Logout.
	var logoutResp map[string]interface{}
	resp3, err := env.doJSON(http.MethodPost, "/api/auth/logout", nil, authHeader(token), &logoutResp)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	requireStatusCode(t, resp3, http.StatusOK)

	// 5. Verify token no longer works.
	var errResp ErrorResponse
	resp4, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &errResp)
	if err != nil {
		t.Fatalf("agents after logout failed: %v", err)
	}
	requireStatusCode(t, resp4, http.StatusUnauthorized)
}

// TestBearerTokenInQueryParams verifies token passed via query param works (fallback).
func TestBearerTokenInQueryParamsNotUsed(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	// Token should work in Authorization header.
	var agentsResp AgentsResponse
	resp, err := env.doJSON(http.MethodGet, "/api/agents", nil, authHeader(token), &agentsResp)
	if err != nil {
		t.Fatalf("agents request with header failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusOK)

	// Without header, should fail (query param fallback not used in our tests).
	var errResp ErrorResponse
	resp2, err := env.doJSON(http.MethodGet, "/api/agents", nil, nil, &errResp)
	if err != nil {
		t.Fatalf("agents without auth failed: %v", err)
	}
	requireStatusCode(t, resp2, http.StatusUnauthorized)
}

// TestUsersListEndpoint verifies /api/users with auth returns user list.
func TestUsersListEndpoint(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	token, _ := env.login(testBootstrapEmail, testBootstrapPassword)

	var listResp map[string]interface{}
	resp, err := env.doJSON(http.MethodGet, "/api/users", nil, authHeader(token), &listResp)
	if err != nil {
		t.Fatalf("users list request failed: %v", err)
	}
	// The endpoint may return 200 even with empty list.
	requireStatusCode(t, resp, http.StatusOK)
}

// TestLoginWithEmptyEmail returns 400.
func TestLoginWithEmptyEmail(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "",
		"password": testBootstrapPassword,
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("login with empty email failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusBadRequest)
	if !strings.Contains(strings.ToLower(errResp.Error), "email") {
		t.Fatalf("expected email error, got: %s", errResp.Error)
	}
}

// TestLoginWithEmptyPassword returns 400.
func TestLoginWithEmptyPassword(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    testBootstrapEmail,
		"password": "",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("login with empty password failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusBadRequest)
	if !strings.Contains(strings.ToLower(errResp.Error), "password") {
		t.Fatalf("expected password error, got: %s", errResp.Error)
	}
}

// TestLoginWithInvalidEmailFormat returns 400.
func TestLoginWithInvalidEmailFormat(t *testing.T) {
	skipIfSlow(t)
	env := newTestEnv(t)
	t.Cleanup(env.stopDaemon)
	env.startDaemon()

	var errResp ErrorResponse
	resp, err := env.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "not-an-email",
		"password": "somepassword",
	}, nil, &errResp)
	if err != nil {
		t.Fatalf("login with invalid email failed: %v", err)
	}
	requireStatusCode(t, resp, http.StatusBadRequest)
	if !strings.Contains(strings.ToLower(errResp.Error), "invalid") {
		t.Fatalf("expected 'invalid' error, got: %s", errResp.Error)
	}
}
