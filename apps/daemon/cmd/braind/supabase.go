package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	corsupabase "github.com/reeinharrrd/brain/core/supabase"
	coreidentity "github.com/reeinharrrd/brain/core/identity"
)

// ---------------------------------------------------------------------------
// Responses
// ---------------------------------------------------------------------------

type supabaseLoginResponse struct {
	Success          bool                  `json:"success"`
	Mode             string                `json:"mode"`
	Required         bool                  `json:"required"`
	Token            string                `json:"token"`
	RefreshToken     string                `json:"refresh_token"`
	ExpiresAt        time.Time             `json:"expires_at"`
	RefreshExpiresAt time.Time             `json:"refresh_expires_at"`
	User             coreidentity.User     `json:"user"`
	Capabilities     []coreidentity.Capability `json:"capabilities"`
	AllowedSections  []string              `json:"allowed_sections"`
}

type supabaseOAuthStartResponse struct {
	Success   bool   `json:"success"`
	Provider  string `json:"provider"`
	AuthURL   string `json:"auth_url"`
}

type supabaseUserResponse struct {
	Success   bool              `json:"success"`
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	Name      string            `json:"name"`
	Role      string            `json:"role"`
	Provider  string            `json:"provider"`
	CreatedAt string            `json:"created_at,omitempty"`
}

// ---------------------------------------------------------------------------
// Supabase client builder
// ---------------------------------------------------------------------------

func (d *BrainDaemon) buildSupabaseClient() (*corsupabase.Client, error) {
	if d == nil {
		return nil, errors.New("daemon is nil")
	}

	baseURL := strings.TrimSpace(os.Getenv("BRAIN_SUPABASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("BRAIN_SUPABASE_ANON_KEY"))
	jwtSecret := strings.TrimSpace(os.Getenv("BRAIN_SUPABASE_JWT_SECRET"))

	if baseURL == "" {
		return nil, errors.New("BRAIN_SUPABASE_URL is not set")
	}
	if apiKey == "" {
		return nil, errors.New("BRAIN_SUPABASE_ANON_KEY is not set")
	}
	if jwtSecret == "" {
		return nil, errors.New("BRAIN_SUPABASE_JWT_SECRET is not set")
	}

	return corsupabase.NewClient(baseURL, apiKey, jwtSecret), nil
}

// ---------------------------------------------------------------------------
// Handler: POST /api/auth/supabase/signup
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseSignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	supClient, err := d.buildSupabaseClient()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "supabase not configured: " + err.Error()})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "email is required"})
		return
	}
	if !isValidEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid email format"})
		return
	}
	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "password is required"})
		return
	}

	if !globalLoginLimiter.allow(req.Email) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "too many attempts, please try again later",
		})
		return
	}

	// Forward to Supabase.
	supResp, err := supClient.SignUp(r.Context(), req.Email, req.Password)
	if err != nil {
		globalLoginLimiter.recordFailure(req.Email)

		switch {
		case errors.Is(err, corsupabase.ErrEmailExists):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "email already registered"})
		case errors.Is(err, corsupabase.ErrWeakPassword):
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "password does not meet security requirements"})
		default:
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "sign up failed"})
		}
		return
	}

	// Verify the Supabase JWT.
	_, err = supClient.VerifyToken(r.Context(), supResp.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "token verification failed"})
		return
	}

	globalLoginLimiter.mu.Lock()
	delete(globalLoginLimiter.attempts, req.Email)
	globalLoginLimiter.mu.Unlock()

	// Resolve user role.
	role := parseAuthRole(os.Getenv("BRAIN_SUPABASE_DEFAULT_ROLE"))
	if role == "" {
		role = coreidentity.RoleMember
	}

	identityUser := coreidentity.User{
		ID:         supResp.User.ID,
		Email:      supResp.User.Email,
		Name:       supResp.User.Name(),
		Role:       role,
		Provider:   "supabase",
		LastSeenAt: time.Now().UTC(),
	}
	identityUser.Capabilities = coreidentity.CapabilitiesForRole(role)
	identityUser.Sections = coreidentity.SectionsForRole(role)

	// Issue internal session.
	session, err := d.auth.IssueSession(r.Context(), identityUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(supabaseLoginResponse{
		Success:          true,
		Mode:             "supabase",
		Required:         d.auth.Config().Required,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		User:             session.User,
		Capabilities:     session.User.Capabilities,
		AllowedSections:  session.User.Sections,
	})
}

// ---------------------------------------------------------------------------
// Handler: POST /api/auth/supabase/signin
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseSignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	supClient, err := d.buildSupabaseClient()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "supabase not configured: " + err.Error()})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "email is required"})
		return
	}
	if !isValidEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid email format"})
		return
	}
	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "password is required"})
		return
	}

	if !globalLoginLimiter.allow(req.Email) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "too many attempts, please try again later",
		})
		return
	}

	// Forward credentials to Supabase.
	supResp, err := supClient.SignIn(r.Context(), req.Email, req.Password)
	if err != nil {
		globalLoginLimiter.recordFailure(req.Email)

		switch {
		case errors.Is(err, corsupabase.ErrInvalidCredentials):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid email or password"})
		default:
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "sign in failed"})
		}
		return
	}

	// Verify the Supabase JWT.
	supUser, err := supClient.VerifyToken(r.Context(), supResp.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "token verification failed"})
		return
	}

	globalLoginLimiter.mu.Lock()
	delete(globalLoginLimiter.attempts, req.Email)
	globalLoginLimiter.mu.Unlock()

	// Resolve user role.
	role := parseAuthRole(os.Getenv("BRAIN_SUPABASE_DEFAULT_ROLE"))
	if role == "" {
		role = coreidentity.RoleMember
	}

	identityUser := coreidentity.User{
		ID:         supUser.ID,
		Email:      supUser.Email,
		Name:       supUser.Name(),
		Role:       role,
		Provider:   "supabase",
		LastSeenAt: time.Now().UTC(),
	}
	identityUser.Capabilities = coreidentity.CapabilitiesForRole(role)
	identityUser.Sections = coreidentity.SectionsForRole(role)

	// Issue internal session.
	session, err := d.auth.IssueSession(r.Context(), identityUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(supabaseLoginResponse{
		Success:          true,
		Mode:             "supabase",
		Required:         d.auth.Config().Required,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		User:             session.User,
		Capabilities:     session.User.Capabilities,
		AllowedSections:  session.User.Sections,
	})
}

// ---------------------------------------------------------------------------
// Handler: GET /api/auth/supabase/oauth/{provider}
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseOAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	supClient, err := d.buildSupabaseClient()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "supabase not configured: " + err.Error()})
		return
	}

	// Extract provider from the URL path: /api/auth/supabase/oauth/google
	path := strings.TrimPrefix(r.URL.Path, "/api/auth/supabase/oauth/")
	provider := strings.TrimSpace(strings.Trim(path, "/"))
	if provider == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "oauth provider is required"})
		return
	}

	oauthResp, err := supClient.SignInWithOAuth(r.Context(), provider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to start oauth flow: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(supabaseOAuthStartResponse{
		Success:  true,
		Provider: oauthResp.Provider,
		AuthURL:  oauthResp.AuthURL,
	})
}

// ---------------------------------------------------------------------------
// Handler: GET /api/auth/supabase/callback
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	// The OAuth callback delivers the access token as a URL fragment (#access_token=...).
	// For server-side flows, the frontend should pass the token as a query parameter
	// after extracting it from the hash fragment.
	accessToken := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if accessToken == "" {
		// Also check Authorization header as fallback.
		if auth := strings.TrimSpace(r.Header.Get("Authorization")); auth != "" {
			if token, ok := strings.CutPrefix(auth, "Bearer "); ok {
				accessToken = strings.TrimSpace(token)
			}
		}
	}
	if accessToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "access token is required"})
		return
	}

	supClient, err := d.buildSupabaseClient()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "supabase not configured: " + err.Error()})
		return
	}

	// Verify the Supabase JWT.
	supUser, err := supClient.VerifyToken(r.Context(), accessToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
		return
	}

	// Resolve user role.
	role := parseAuthRole(os.Getenv("BRAIN_SUPABASE_DEFAULT_ROLE"))
	if role == "" {
		role = coreidentity.RoleMember
	}

	identityUser := coreidentity.User{
		ID:         supUser.ID,
		Email:      supUser.Email,
		Name:       supUser.Name(),
		Role:       role,
		Provider:   "supabase",
		LastSeenAt: time.Now().UTC(),
	}
	identityUser.Capabilities = coreidentity.CapabilitiesForRole(role)
	identityUser.Sections = coreidentity.SectionsForRole(role)

	// Issue internal session.
	session, err := d.auth.IssueSession(r.Context(), identityUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(supabaseLoginResponse{
		Success:          true,
		Mode:             "supabase",
		Required:         d.auth.Config().Required,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		User:             session.User,
		Capabilities:     session.User.Capabilities,
		AllowedSections:  session.User.Sections,
	})
}

// ---------------------------------------------------------------------------
// Handler: POST /api/auth/supabase/refresh
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	supClient, err := d.buildSupabaseClient()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "supabase not configured: " + err.Error()})
		return
	}

	// Extract the Supabase refresh token.
	refreshToken := d.extractRefreshToken(r)
	if refreshToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh token is required"})
		return
	}

	// Refresh via Supabase.
	supResp, err := supClient.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, corsupabase.ErrTokenExpired):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh token expired, please login again"})
		case errors.Is(err, corsupabase.ErrInvalidCredentials):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid refresh token"})
		default:
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh failed"})
		}
		return
	}

	// Verify the new access token.
	supUser, err := supClient.VerifyToken(r.Context(), supResp.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "token verification failed"})
		return
	}

	// Resolve user role.
	role := parseAuthRole(os.Getenv("BRAIN_SUPABASE_DEFAULT_ROLE"))
	if role == "" {
		role = coreidentity.RoleMember
	}

	identityUser := coreidentity.User{
		ID:         supUser.ID,
		Email:      supUser.Email,
		Name:       supUser.Name(),
		Role:       role,
		Provider:   "supabase",
		LastSeenAt: time.Now().UTC(),
	}
	identityUser.Capabilities = coreidentity.CapabilitiesForRole(role)
	identityUser.Sections = coreidentity.SectionsForRole(role)

	// Issue a fresh internal session.
	session, err := d.auth.IssueSession(r.Context(), identityUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(supabaseLoginResponse{
		Success:          true,
		Mode:             "supabase",
		Required:         d.auth.Config().Required,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		User:             session.User,
		Capabilities:     session.User.Capabilities,
		AllowedSections:  session.User.Sections,
	})
}

// ---------------------------------------------------------------------------
// Handler: POST /api/auth/supabase/signout
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseSignOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	token := d.extractAuthToken(r)
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication required"})
		return
	}

	// Best-effort Supabase sign-out (does not block local logout).
	if supClient, err := d.buildSupabaseClient(); err == nil && supClient != nil {
		_ = supClient.SignOut(r.Context(), token)
	}

	// Always remove the local session.
	if err := d.auth.Logout(r.Context(), token); err != nil && !errors.Is(err, coreidentity.ErrSessionNotFound) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "logout failed"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "signed out",
	})
}

// ---------------------------------------------------------------------------
// Handler: GET /api/auth/supabase/user
// ---------------------------------------------------------------------------

func (d *BrainDaemon) handleSupabaseUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}

	// First try the daemon's internal session.
	token := d.extractAuthToken(r)
	if token != "" {
		session, err := d.auth.Authenticate(r.Context(), token)
		if err == nil && session != nil {
			user := session.User
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(supabaseUserResponse{
				Success:   true,
				ID:        user.ID,
				Email:     user.Email,
				Name:      user.Name,
				Role:      string(user.Role),
				Provider:  user.Provider,
				CreatedAt: user.LastSeenAt.Format(time.RFC3339),
			})
			return
		}
	}

	// Fallback: verify Supabase JWT directly.
	accessToken := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if accessToken == "" && token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication required"})
		return
	}
	if accessToken == "" {
		accessToken = token
	}

	supClient, err := d.buildSupabaseClient()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "supabase not configured: " + err.Error()})
		return
	}

	supUser, err := supClient.VerifyToken(r.Context(), accessToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(supabaseUserResponse{
		Success:   true,
		ID:        supUser.ID,
		Email:     supUser.Email,
		Name:      supUser.Name(),
		Role:      supUser.Role,
		Provider:  "supabase",
		CreatedAt: supUser.CreatedAt,
	})
}
