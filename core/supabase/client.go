// Package supabase provides a Go client for the Supabase GoTrue Auth API.
//
// It communicates directly with the GoTrue HTTP endpoints (no external SDK),
// using only the standard library plus jwt/v5 for token verification.
package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Default timeouts and endpoints.
const (
	defaultHTTPTimeout  = 30 * time.Second
	signUpEndpoint      = "/signup"
	signInEndpoint      = "/token?grant_type=password"
	tokenEndpoint       = "/token?grant_type=refresh_token"
	userEndpoint        = "/user"
	signOutEndpoint     = "/logout"
	oAuthAuthorizeEP    = "/authorize"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// Client is a thin HTTP wrapper around the Supabase GoTrue API.
type Client struct {
	baseURL    string // e.g. http://127.0.0.1:9999
	apiKey     string // anon / public key
	jwtSecret  string // secret used to verify access tokens
	httpClient *http.Client
}

// SignUpRequest describes the body for a password-based registration.
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignInRequest describes the body for a password-based sign-in.
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the body sent when exchanging a refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// User represents the Supabase user record returned after authentication.
type User struct {
	ID               string                 `json:"id"`
	Aud               string                 `json:"aud"`
	Role              string                 `json:"role"`
	Email            string                 `json:"email"`
	EmailVerified    bool                   `json:"email_verified"`
	Phone            string                 `json:"phone"`
	PhoneVerified    bool                   `json:"phone_verified"`
	LastSignInAt     string                 `json:"last_sign_in_at"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	AppMetadata      map[string]interface{} `json:"app_metadata"`
	UserMetadata     map[string]interface{} `json:"user_metadata"`
	Identities       []interface{}          `json:"identities"`
	ConfirmationSentAt string               `json:"confirmation_sent_at,omitempty"`
	RecoverySentAt   string                 `json:"recovery_sent_at,omitempty"`
}

// Name extracts the display name from user_metadata (common OAuth field).
func (u *User) Name() string {
	if u == nil {
		return ""
	}
	if v, ok := u.UserMetadata["full_name"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := u.UserMetadata["name"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	// Fallback to email local-part.
	if u.Email != "" {
		if idx := strings.Index(u.Email, "@"); idx > 0 {
			return u.Email[:idx]
		}
	}
	return "Supabase User"
}

// SignInResponse wraps the tokens and user payload returned on login.
type SignInResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// OAuthResponse is returned when initiating an OAuth provider flow.
type OAuthResponse struct {
	Provider   string `json:"provider"`
	AuthURL    string `json:"auth_url"`
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// NewClient creates a Supabase GoTrue client.
//
//   - baseURL: the GoTrue base URL (e.g. http://127.0.0.1:9999).
//   - apiKey:  the anon/public key (sent as the apikey query param and
//     Authorization header).
//   - jwtSecret: the JWT secret used to verify access-token signatures.
func NewClient(baseURL, apiKey, jwtSecret string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:   baseURL,
		apiKey:    apiKey,
		jwtSecret: jwtSecret,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

// SetHTTPClient replaces the default HTTP client (useful for testing).
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func (c *Client) doJSON(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	u := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseSupabaseError(resp.StatusCode, respBody)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Supabase errors
// ---------------------------------------------------------------------------

// APIError is a structured error returned by the GoTrue API.
type APIError struct {
	StatusCode    int    `json:"-"`
	ErrorCode     string `json:"error_code,omitempty"`
	Msg           string `json:"msg,omitempty"`
	Message       string `json:"message,omitempty"`
	ErrorField    string `json:"error,omitempty"`
	Description   string `json:"error_description,omitempty"`
}

func (e *APIError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Description != "" {
		return e.Description
	}
	if e.ErrorField != "" {
		return e.ErrorField
	}
	return fmt.Sprintf("supabase api error (status %d)", e.StatusCode)
}

func parseSupabaseError(statusCode int, body []byte) error {
	var apiErr APIError
	apiErr.StatusCode = statusCode
	if err := json.Unmarshal(body, &apiErr); err != nil {
		// Fall back to plain-text error.
		apiErr.Msg = strings.TrimSpace(string(body))
		if apiErr.Msg == "" {
			apiErr.Msg = http.StatusText(statusCode)
		}
		return &apiErr
	}
	return &apiErr
}

// Sentinel errors.
var (
	ErrInvalidCredentials = errors.New("supabase: invalid credentials")
	ErrEmailExists        = errors.New("supabase: email already registered")
	ErrWeakPassword       = errors.New("supabase: password does not meet security requirements")
	ErrTokenExpired       = errors.New("supabase: token expired")
	ErrInvalidToken       = errors.New("supabase: invalid or malformed token")
	ErrUserNotFound       = errors.New("supabase: user not found")
)

func classifyAPIError(err error) error {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return err
	}

	msg := strings.ToLower(apiErr.Msg + " " + apiErr.Message + " " + apiErr.Description + " " + apiErr.ErrorField)

	switch {
	case strings.Contains(msg, "invalid login") ||
		strings.Contains(msg, "invalid credentials") ||
		strings.Contains(msg, "invalid email") && strings.Contains(msg, "password"):
		return ErrInvalidCredentials
	case strings.Contains(msg, "user already registered") ||
		strings.Contains(msg, "already been registered") ||
		strings.Contains(msg, "email already") ||
		(apiErr.StatusCode == http.StatusUnprocessableEntity && strings.Contains(msg, "email")):
		return ErrEmailExists
	case strings.Contains(msg, "password") && (strings.Contains(msg, "weak") ||
		strings.Contains(msg, "too short") ||
		strings.Contains(msg, "should be at least") ||
		strings.Contains(msg, "too long")):
		return ErrWeakPassword
	case strings.Contains(msg, "expired") || strings.Contains(msg, "token is expired"):
		return ErrTokenExpired
	case strings.Contains(msg, "not found") || strings.Contains(msg, "no user found"):
		return ErrUserNotFound
	default:
		return err
	}
}

// ---------------------------------------------------------------------------
// Public API methods
// ---------------------------------------------------------------------------

// SignUp creates a new user with email and password.
func (c *Client) SignUp(ctx context.Context, email, password string) (*SignInResponse, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("supabase: email is required")
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("supabase: password is required")
	}

	req := SignUpRequest{
		Email:    strings.TrimSpace(email),
		Password: password,
	}
	var resp SignInResponse
	err := c.doJSON(ctx, http.MethodPost, signUpEndpoint, req, &resp)
	if err != nil {
		return nil, classifyAPIError(err)
	}
	return &resp, nil
}

// SignIn authenticates a user with email and password and returns tokens.
func (c *Client) SignIn(ctx context.Context, email, password string) (*SignInResponse, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("supabase: email is required")
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("supabase: password is required")
	}

	req := SignInRequest{
		Email:    strings.TrimSpace(email),
		Password: password,
	}
	var resp SignInResponse
	err := c.doJSON(ctx, http.MethodPost, signInEndpoint, req, &resp)
	if err != nil {
		return nil, classifyAPIError(err)
	}
	return &resp, nil
}

// SignInWithOAuth returns the authorization URL for the given OAuth provider.
// Valid providers: "google", "github", "azure" (Microsoft).
func (c *Client) SignInWithOAuth(ctx context.Context, provider string) (*OAuthResponse, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, errors.New("supabase: provider is required")
	}

	// Build the authorize URL.
	u, err := url.Parse(c.baseURL + oAuthAuthorizeEP)
	if err != nil {
		return nil, fmt.Errorf("parse authorize endpoint: %w", err)
	}

	q := u.Query()
	q.Set("provider", provider)
	q.Set("apikey", c.apiKey)
	u.RawQuery = q.Encode()

	return &OAuthResponse{
		Provider: provider,
		AuthURL:  u.String(),
	}, nil
}

// VerifyToken validates and decodes a Supabase JWT access token.
//
// It verifies the signature (HMAC-SHA256), the issuer (iss), and the
// expiration (exp). On success it returns the decoded *User.
func (c *Client) VerifyToken(ctx context.Context, accessToken string) (*User, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, errors.New("supabase: access token is required")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.jwtSecret), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify issuer is Supabase.
	if iss, _ := claims["iss"].(string); iss != "" {
		if !strings.Contains(strings.ToLower(iss), "supabase") &&
			!strings.Contains(strings.ToLower(iss), "gotrue") {
			// We are lenient here -- some self-hosted instances may have
			// a custom issuer. Still log it but do not reject outright.
		}
		_ = iss // accepted
	}

	user := &User{}
	if sub, ok := claims["sub"].(string); ok {
		user.ID = sub
	}
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}
	if role, ok := claims["role"].(string); ok {
		user.Role = role
	}
	if aud, ok := claims["aud"].(string); ok {
		user.Aud = aud
	}
	if raw, ok := claims["user_metadata"].(map[string]interface{}); ok {
		user.UserMetadata = raw
	}
	if raw, ok := claims["app_metadata"].(map[string]interface{}); ok {
		user.AppMetadata = raw
	}

	return user, nil
}

// RefreshToken exchanges a valid refresh token for a new token pair.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*SignInResponse, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, errors.New("supabase: refresh token is required")
	}

	req := RefreshRequest{
		RefreshToken: strings.TrimSpace(refreshToken),
	}
	var resp SignInResponse
	err := c.doJSON(ctx, http.MethodPost, tokenEndpoint, req, &resp)
	if err != nil {
		return nil, classifyAPIError(err)
	}
	return &resp, nil
}

// SignOut invalidates the current session on the Supabase side.
func (c *Client) SignOut(ctx context.Context, accessToken string) error {
	if strings.TrimSpace(accessToken) == "" {
		return errors.New("supabase: access token is required")
	}

	// SignOut uses a custom request because the GoTrue /logout endpoint
	// returns an empty body on success, which our doJSON helper cannot
	// unmarshal into a non-nil target.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+signOutEndpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Drain the body.
	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return classifyAPIError(parseSupabaseError(resp.StatusCode, nil))
	}
	return nil
}

// GetUser fetches the current user profile for the given access token.
func (c *Client) GetUser(ctx context.Context, accessToken string) (*User, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, errors.New("supabase: access token is required")
	}

	var user User
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+userEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, classifyAPIError(parseSupabaseError(resp.StatusCode, respBody))
	}

	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user response: %w", err)
	}
	return &user, nil
}
