package supabase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const testJWTSecret = "test-super-secret-jwt-key-with-at-least-32-characters-long"

func makeTestJWT(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}
	return signed
}

func newTestClient(handler http.Handler, jwtSecret string) *Client {
	server := httptest.NewServer(handler)
	c := NewClient(server.URL, "test-anon-key", jwtSecret)
	return c
}

// ---------------------------------------------------------------------------
// NewClient
// ---------------------------------------------------------------------------

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c := NewClient("http://localhost:9999/", "key", "secret")
	if c.baseURL != "http://localhost:9999" {
		t.Errorf("expected baseURL to trim trailing slash, got %q", c.baseURL)
	}
}

// ---------------------------------------------------------------------------
// SignUp
// ---------------------------------------------------------------------------

func TestSignUp_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/signup" {
			t.Errorf("expected /signup, got %s", r.URL.Path)
		}

		var req SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Email != "alice@example.com" {
			t.Errorf("unexpected email: %s", req.Email)
		}

		resp := SignInResponse{
			AccessToken:  "eyJ.test.token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			RefreshToken: "refresh_abc",
			User: User{
				ID:    "user-123",
				Email: "alice@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := newTestClient(handler, testJWTSecret)
	resp, err := c.SignUp(t.Context(), "alice@example.com", "secure-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "eyJ.test.token" {
		t.Errorf("unexpected access token: %s", resp.AccessToken)
	}
	if resp.User.Email != "alice@example.com" {
		t.Errorf("unexpected user email: %s", resp.User.Email)
	}
}

func TestSignUp_EmptyEmail(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", "secret")
	_, err := c.SignUp(t.Context(), "", "password")
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestSignUp_EmptyPassword(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", "secret")
	_, err := c.SignUp(t.Context(), "test@example.com", "")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestSignUp_EmailAlreadyExists(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"msg": "User already registered",
		})
	})

	c := newTestClient(handler, testJWTSecret)
	_, err := c.SignUp(t.Context(), "dup@example.com", "password")
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrEmailExists {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SignIn
// ---------------------------------------------------------------------------

func TestSignIn_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("expected /token, got %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "grant_type=password") {
			t.Errorf("expected grant_type=password in query, got %s", r.URL.RawQuery)
		}

		resp := SignInResponse{
			AccessToken:  "eyJ.access.token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
			RefreshToken: "refresh_xyz",
			User: User{
				ID:    "user-456",
				Email: "bob@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := newTestClient(handler, testJWTSecret)
	resp, err := c.SignIn(t.Context(), "bob@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RefreshToken != "refresh_xyz" {
		t.Errorf("unexpected refresh token: %s", resp.RefreshToken)
	}
	if resp.User.ID != "user-456" {
		t.Errorf("unexpected user id: %s", resp.User.ID)
	}
}

func TestSignIn_InvalidCredentials(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"msg": "Invalid login credentials",
		})
	})

	c := newTestClient(handler, testJWTSecret)
	_, err := c.SignIn(t.Context(), "bob@example.com", "wrong")
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SignInWithOAuth
// ---------------------------------------------------------------------------

func TestSignInWithOAuth_Success(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", "secret")
	resp, err := c.SignInWithOAuth(t.Context(), "google")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Provider != "google" {
		t.Errorf("expected provider google, got %s", resp.Provider)
	}
	if !strings.Contains(resp.AuthURL, "/authorize") {
		t.Errorf("auth URL missing /authorize: %s", resp.AuthURL)
	}
	if !strings.Contains(resp.AuthURL, "provider=google") {
		t.Errorf("auth URL missing provider param: %s", resp.AuthURL)
	}
}

func TestSignInWithOAuth_EmptyProvider(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", "secret")
	_, err := c.SignInWithOAuth(t.Context(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// VerifyToken
// ---------------------------------------------------------------------------

func TestVerifyToken_Success(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":   "user-789",
		"email": "carol@example.com",
		"role":  "authenticated",
		"aud":   "authenticated",
		"iss":   "https://xyz.supabase.co/auth/v1",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"user_metadata": map[string]interface{}{
			"full_name": "Carol Danvers",
		},
	}
	rawToken := makeTestJWT(t, claims)

	c := NewClient("http://localhost:9999", "key", testJWTSecret)
	user, err := c.VerifyToken(t.Context(), rawToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-789" {
		t.Errorf("expected user-789, got %s", user.ID)
	}
	if user.Email != "carol@example.com" {
		t.Errorf("expected carol@example.com, got %s", user.Email)
	}
	if user.Name() != "Carol Danvers" {
		t.Errorf("expected Carol Danvers, got %s", user.Name())
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "user-exp",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	rawToken := makeTestJWT(t, claims)

	c := NewClient("http://localhost:9999", "key", testJWTSecret)
	_, err := c.VerifyToken(t.Context(), rawToken)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyToken_InvalidSignature(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "user-bad",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	// Sign with a different secret.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	rawToken, _ := token.SignedString([]byte("wrong-secret"))

	c := NewClient("http://localhost:9999", "key", testJWTSecret)
	_, err := c.VerifyToken(t.Context(), rawToken)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestVerifyToken_EmptyToken(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", testJWTSecret)
	_, err := c.VerifyToken(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

// ---------------------------------------------------------------------------
// RefreshToken
// ---------------------------------------------------------------------------

func TestRefreshToken_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "grant_type=refresh_token") {
			t.Errorf("expected grant_type=refresh_token, got %s", r.URL.RawQuery)
		}
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.RefreshToken != "valid_refresh" {
			t.Errorf("unexpected refresh token: %s", req.RefreshToken)
		}

		resp := SignInResponse{
			AccessToken:  "eyJ.new.access",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			RefreshToken: "new_refresh",
			User: User{
				ID:    "user-789",
				Email: "refreshed@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := newTestClient(handler, testJWTSecret)
	resp, err := c.RefreshToken(t.Context(), "valid_refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "eyJ.new.access" {
		t.Errorf("unexpected access token: %s", resp.AccessToken)
	}
	if resp.RefreshToken != "new_refresh" {
		t.Errorf("unexpected refresh token: %s", resp.RefreshToken)
	}
}

func TestRefreshToken_Empty(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", testJWTSecret)
	_, err := c.RefreshToken(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for empty refresh token")
	}
}

// ---------------------------------------------------------------------------
// SignOut
// ---------------------------------------------------------------------------

func TestSignOut_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/logout" {
			t.Errorf("expected /logout, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})

	c := newTestClient(handler, testJWTSecret)
	err := c.SignOut(t.Context(), "valid_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignOut_EmptyToken(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", testJWTSecret)
	err := c.SignOut(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

// ---------------------------------------------------------------------------
// GetUser
// ---------------------------------------------------------------------------

func TestGetUser_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("expected /user, got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid_user_token" {
			t.Errorf("unexpected Authorization: %s", auth)
		}

		user := User{
			ID:    "user-get",
			Email: "getuser@example.com",
			UserMetadata: map[string]interface{}{
				"name": "Get User",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(user)
	})

	c := newTestClient(handler, testJWTSecret)
	user, err := c.GetUser(t.Context(), "valid_user_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-get" {
		t.Errorf("expected user-get, got %s", user.ID)
	}
	if user.Name() != "Get User" {
		t.Errorf("expected Get User, got %s", user.Name())
	}
}

// ---------------------------------------------------------------------------
// User.Name()
// ---------------------------------------------------------------------------

func TestUserName_FromFullName(t *testing.T) {
	u := &User{
		UserMetadata: map[string]interface{}{"full_name": "Full Name"},
	}
	if u.Name() != "Full Name" {
		t.Errorf("expected Full Name, got %s", u.Name())
	}
}

func TestUserName_FromName(t *testing.T) {
	u := &User{
		UserMetadata: map[string]interface{}{"name": "Short Name"},
	}
	if u.Name() != "Short Name" {
		t.Errorf("expected Short Name, got %s", u.Name())
	}
}

func TestUserName_FromEmail(t *testing.T) {
	u := &User{
		Email: "dave@example.com",
	}
	if u.Name() != "dave" {
		t.Errorf("expected dave, got %s", u.Name())
	}
}

func TestUserName_DefaultFallback(t *testing.T) {
	u := &User{}
	if u.Name() != "Supabase User" {
		t.Errorf("expected Supabase User, got %s", u.Name())
	}
}

func TestUserName_NilReceiver(t *testing.T) {
	var u *User
	// Should not panic.
	_ = u.Name()
}

// ---------------------------------------------------------------------------
// APIError
// ---------------------------------------------------------------------------

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  APIError
		want string
	}{
		{"msg", APIError{Msg: "something broke"}, "something broke"},
		{"message", APIError{Message: "api message"}, "api message"},
		{"description", APIError{Description: "desc"}, "desc"},
		{"error field", APIError{ErrorField: "invalid_grant"}, "invalid_grant"},
		{"fallback", APIError{StatusCode: 500}, "supabase api error (status 500)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// classifyAPIError
// ---------------------------------------------------------------------------

func TestClassifyAPIError_InvalidCredentials(t *testing.T) {
	err := &APIError{StatusCode: 400, Msg: "Invalid login credentials"}
	got := classifyAPIError(err)
	if got != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", got)
	}
}

func TestClassifyAPIError_WeakPassword(t *testing.T) {
	err := &APIError{StatusCode: 400, Msg: "Password should be at least 6 characters"}
	got := classifyAPIError(err)
	if got != ErrWeakPassword {
		t.Errorf("expected ErrWeakPassword, got %v", got)
	}
}

func TestClassifyAPIError_TokenExpired(t *testing.T) {
	err := &APIError{StatusCode: 400, Msg: "Token is expired"}
	got := classifyAPIError(err)
	if got != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", got)
	}
}

func TestClassifyAPIError_UserNotFound(t *testing.T) {
	err := &APIError{StatusCode: 404, Msg: "User not found"}
	got := classifyAPIError(err)
	if got != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", got)
	}
}

func TestClassifyAPIError_Passthrough(t *testing.T) {
	err := &APIError{StatusCode: 500, Msg: "Internal server error"}
	got := classifyAPIError(err)
	if got == ErrInvalidCredentials || got == ErrEmailExists ||
		got == ErrWeakPassword || got == ErrTokenExpired || got == ErrUserNotFound {
		t.Errorf("expected original error, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// SetHTTPClient
// ---------------------------------------------------------------------------

func TestSetHTTPClient(t *testing.T) {
	c := NewClient("http://localhost:9999", "key", "secret")
	custom := &http.Client{Timeout: 5 * time.Second}
	c.SetHTTPClient(custom)
	if c.httpClient != custom {
		t.Error("expected httpClient to be replaced")
	}
}
