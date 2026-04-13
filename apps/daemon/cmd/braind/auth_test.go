package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	coreidentity "github.com/reeinharrrd/brain/core/identity"
)

func setupTestAuth(t *testing.T) *BrainDaemon {
	os.Setenv("BRAIN_AUTH_REQUIRED", "true")
	os.Setenv("BRAIN_AUTH_MODE", "bootstrap")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_EMAIL", "test@brain.local")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_PASSWORD", "secret123")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_NAME", "Test User")
	os.Setenv("BRAIN_AUTH_BOOTSTRAP_ROLE", "owner")

	d := &BrainDaemon{
		auth: buildIdentityManager("test"),
	}
	return d
}

func TestAuthHandlersLoginStatusLogout(t *testing.T) {
	d := setupTestAuth(t)

	loginBody := `{"email":"test@brain.local","password":"secret123"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()

	d.handleAuthLogin(loginRR, loginReq)

	if loginRR.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", loginRR.Code, loginRR.Body.String())
	}

	var loginResp authLoginResponse
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("login response parse error: %v", err)
	}
	if !loginResp.Success {
		t.Fatal("login response success flag not set")
	}
	if loginResp.Token == "" {
		t.Fatal("login response missing token")
	}
	if loginResp.RefreshToken == "" {
		t.Fatal("login response missing refresh token")
	}
	if loginResp.User.Email != "test@brain.local" {
		t.Fatalf("unexpected user email: %s", loginResp.User.Email)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	statusRR := httptest.NewRecorder()

	d.handleAuthStatus(statusRR, statusReq)

	if statusRR.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", statusRR.Code, statusRR.Body.String())
	}

	var statusResp coreidentity.Status
	if err := json.Unmarshal(statusRR.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("status response parse error: %v", err)
	}
	if !statusResp.Authenticated {
		t.Fatal("status should show authenticated")
	}
	if statusResp.User == nil || statusResp.User.Email != "test@brain.local" {
		t.Fatal("status response missing user info")
	}
	if len(statusResp.User.Capabilities) == 0 {
		t.Fatal("owner role should have capabilities")
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	logoutRR := httptest.NewRecorder()

	d.handleAuthLogout(logoutRR, logoutReq)

	if logoutRR.Code != http.StatusOK {
		t.Fatalf("logout expected 200, got %d: %s", logoutRR.Code, logoutRR.Body.String())
	}

	statusAfterLogout := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	statusAfterLogout.Header.Set("Authorization", "Bearer "+loginResp.Token)
	statusAfterLogoutRR := httptest.NewRecorder()

	d.handleAuthStatus(statusAfterLogoutRR, statusAfterLogout)

	var statusAfterLogoutResp coreidentity.Status
	if err := json.Unmarshal(statusAfterLogoutRR.Body.Bytes(), &statusAfterLogoutResp); err != nil {
		t.Fatalf("post-logout status parse error: %v", err)
	}
	if statusAfterLogoutResp.Authenticated {
		t.Fatal("status should show unauthenticated after logout")
	}
}

func TestAuthLoginInvalidCredentials(t *testing.T) {
	d := setupTestAuth(t)

	body := `{"email":"test@brain.local","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	d.handleAuthLogin(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if !strings.Contains(resp["error"], "invalid email or password") {
		t.Fatalf("expected 'invalid email or password', got: %s", resp["error"])
	}
}

func TestAuthLoginMissingFields(t *testing.T) {
	d := setupTestAuth(t)

	cases := []struct {
		name          string
		body          string
		expectedCode  int
		expectedError string
	}{
		{"empty body", "", http.StatusBadRequest, "invalid request body"},
		{"missing email", `{"password":"secret123"}`, http.StatusBadRequest, "email is required"},
		{"missing password", `{"email":"test@brain.local"}`, http.StatusBadRequest, "password is required"},
		{"invalid email", `{"email":"notanemail","password":"secret123"}`, http.StatusBadRequest, "invalid email format"},
		{"empty email", `{"email":"","password":"secret123"}`, http.StatusBadRequest, "email is required"},
		{"empty password", `{"email":"test@brain.local","password":""}`, http.StatusBadRequest, "password is required"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			d.handleAuthLogin(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("expected %d, got %d: %s", tc.expectedCode, rr.Code, rr.Body.String())
			}

			var resp map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response parse error: %v", err)
			}
			if !strings.Contains(resp["error"], tc.expectedError) {
				t.Errorf("expected error containing %q, got: %s", tc.expectedError, resp["error"])
			}
		})
	}
}

func TestAuthLoginRateLimiting(t *testing.T) {
	d := setupTestAuth(t)

	body := `{"email":"ratelimit@brain.local","password":"wrongpassword"}`

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		d.handleAuthLogin(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d: %s", i+1, rr.Code, rr.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	d.handleAuthLogin(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after rate limit, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if !strings.Contains(resp["error"], "too many login attempts") {
		t.Fatalf("expected rate limit error, got: %s", resp["error"])
	}
}

func TestAuthLoginMethodNotAllowed(t *testing.T) {
	d := setupTestAuth(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rr := httptest.NewRecorder()

	d.handleAuthLogin(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthLogoutWithoutToken(t *testing.T) {
	d := setupTestAuth(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()

	d.handleAuthLogout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthStatusWithoutToken(t *testing.T) {
	d := setupTestAuth(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	rr := httptest.NewRecorder()

	d.handleAuthStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp coreidentity.Status
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if resp.Authenticated {
		t.Fatal("should not be authenticated without token")
	}
	if resp.Required != true {
		t.Fatal("auth should be required")
	}
}

func TestAuthRefreshWithoutToken(t *testing.T) {
	d := setupTestAuth(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	rr := httptest.NewRecorder()

	d.handleAuthRefresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthLoginInternalErrorNotExposed(t *testing.T) {
	d := &BrainDaemon{
		auth: coreidentity.NewManager(coreidentity.Config{
			Mode:     coreidentity.ModeBootstrap,
			Required: true,
		}),
	}

	body := `{"email":"test@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	d.handleAuthLogin(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response parse error: %v", err)
	}
	if strings.Contains(resp["error"], "nil") || strings.Contains(resp["error"], "panic") || strings.Contains(resp["error"], "internal") {
		t.Fatalf("internal error details exposed: %s", resp["error"])
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := newRateLimiter(2, 100*time.Millisecond, 50*time.Millisecond)

	identifier := "test@example.com"

	if !rl.allow(identifier) {
		t.Fatal("first attempt should be allowed")
	}
	rl.recordFailure(identifier)

	if !rl.allow(identifier) {
		t.Fatal("second attempt should be allowed")
	}
	rl.recordFailure(identifier)

	if rl.allow(identifier) {
		t.Fatal("third attempt should be blocked")
	}

	time.Sleep(200 * time.Millisecond)
	rl.cleanup()

	if !rl.allow(identifier) {
		t.Fatal("should be allowed after cleanup and cooldown")
	}
}

func TestIsValidEmail(t *testing.T) {
	cases := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{"notanemail", false},
		{"", false},
		{"@missing-local.com", false},
	}

	for _, tc := range cases {
		t.Run(tc.email, func(t *testing.T) {
			result := isValidEmail(tc.email)
			if result != tc.expected {
				t.Errorf("isValidEmail(%q) = %v, expected %v", tc.email, result, tc.expected)
			}
		})
	}
}

func TestSanitizeAuthError(t *testing.T) {
	cases := []struct {
		err      error
		expected string
	}{
		{coreidentity.ErrInvalidCredentials, "invalid email or password"},
		{coreidentity.ErrEmptyEmail, "email is required"},
		{coreidentity.ErrEmptyPassword, "password is required"},
		{coreidentity.ErrSessionExpired, "session expired, please login again"},
		{coreidentity.ErrSessionNotFound, "session not found"},
		{coreidentity.ErrEmptyToken, "authentication token is required"},
		{coreidentity.ErrAuthDisabled, "authentication is disabled"},
		{coreidentity.ErrAuthenticationMode, "use OIDC login instead"},
		{fmt.Errorf("some internal error"), "authentication service error"},
		{nil, ""},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			result := sanitizeAuthError(tc.err)
			if result != tc.expected {
				t.Errorf("sanitizeAuthError(%v) = %q, expected %q", tc.err, result, tc.expected)
			}
		})
	}
}

func TestInviteConsumePublicRoute(t *testing.T) {
	d := setupTestAuth(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"test@brain.local","password":"secret123"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	d.handleAuthLogin(createRR, createReq)

	var loginResp authLoginResponse
	if err := json.Unmarshal(createRR.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("login response parse error: %v", err)
	}

	inviteReq := httptest.NewRequest(http.MethodPost, "/api/invites", strings.NewReader(`{"email":"newuser@example.com","role":"member"}`))
	inviteReq.Header.Set("Content-Type", "application/json")
	inviteReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	inviteRR := httptest.NewRecorder()

	d.handleInviteCreate(inviteRR, inviteReq)

	if inviteRR.Code != http.StatusOK {
		t.Fatalf("invite create expected 200, got %d: %s", inviteRR.Code, inviteRR.Body.String())
	}

	var inviteResp map[string]any
	if err := json.Unmarshal(inviteRR.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("invite response parse error: %v", err)
	}

	inviteToken := inviteResp["invite"].(map[string]any)["token"].(string)

	consumeReq := httptest.NewRequest(http.MethodPost, "/api/invites/consume", strings.NewReader(`{"token":"`+inviteToken+`"}`))
	consumeReq.Header.Set("Content-Type", "application/json")
	consumeRR := httptest.NewRecorder()

	d.handleInviteConsume(consumeRR, consumeReq)

	if consumeRR.Code != http.StatusOK {
		t.Fatalf("invite consume expected 200, got %d: %s", consumeRR.Code, consumeRR.Body.String())
	}
}

func TestUserRoleUpdateValidation(t *testing.T) {
	d := setupTestAuth(t)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"test@brain.local","password":"secret123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	d.handleAuthLogin(loginRR, loginReq)

	var loginResp authLoginResponse
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("login response parse error: %v", err)
	}

	cases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{"valid role", `{"user_id":"test-user","role":"admin"}`, http.StatusBadRequest},
		{"invalid role", `{"user_id":"test-user","role":"superadmin"}`, http.StatusBadRequest},
		{"missing user_id", `{"role":"admin"}`, http.StatusBadRequest},
		{"empty role", `{"user_id":"test-user"}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/users/test-user/role", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+loginResp.Token)
			rr := httptest.NewRecorder()

			d.handleUserRoleUpdate(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("expected %d, got %d: %s", tc.expectedCode, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestInviteCreateValidation(t *testing.T) {
	d := setupTestAuth(t)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"test@brain.local","password":"secret123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	d.handleAuthLogin(loginRR, loginReq)

	var loginResp authLoginResponse
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("login response parse error: %v", err)
	}

	cases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{"valid", `{"email":"new@example.com"}`, http.StatusOK},
		{"invalid email", `{"email":"notanemail"}`, http.StatusBadRequest},
		{"missing email", `{}`, http.StatusBadRequest},
		{"with valid role", `{"email":"new@example.com","role":"viewer"}`, http.StatusOK},
		{"with invalid role", `{"email":"new@example.com","role":"god"}`, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/invites", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+loginResp.Token)
			rr := httptest.NewRecorder()

			d.handleInviteCreate(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("expected %d, got %d: %s", tc.expectedCode, rr.Code, rr.Body.String())
			}
		})
	}
}
