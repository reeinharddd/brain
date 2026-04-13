package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	coreidentity "github.com/reeinharrrd/brain/core/identity"
	brainenv "github.com/reeinharrrd/brain/daemon/internal/environment"
)

// rateLimiter tracks login attempts per identifier and enforces a cooldown.
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*attemptRecord
	maxFail  int
	window   time.Duration
	cooldown time.Duration
}

type attemptRecord struct {
	fails      int
	lastFail   time.Time
	lockedUntil time.Time
}

func newRateLimiter(maxFail int, window, cooldown time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string]*attemptRecord),
		maxFail:  maxFail,
		window:   window,
		cooldown: cooldown,
	}
}

func (rl *rateLimiter) allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	rec, exists := rl.attempts[identifier]
	if !exists {
		rl.attempts[identifier] = &attemptRecord{fails: 0}
		return true
	}

	if !rec.lockedUntil.IsZero() && now.Before(rec.lockedUntil) {
		return false
	}

	if now.Sub(rec.lastFail) > rl.window {
		rec.fails = 0
		rec.lockedUntil = time.Time{}
		return true
	}

	if rec.fails >= rl.maxFail {
		rec.lockedUntil = now.Add(rl.cooldown)
		rec.fails = 0
		return false
	}

	return true
}

func (rl *rateLimiter) recordFailure(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	rec, exists := rl.attempts[identifier]
	if !exists {
		rec = &attemptRecord{}
		rl.attempts[identifier] = rec
	}

	rec.fails++
	rec.lastFail = now

	if rec.fails >= rl.maxFail {
		rec.lockedUntil = now.Add(rl.cooldown)
	}
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	for id, rec := range rl.attempts {
		if !rec.lockedUntil.IsZero() && now.After(rec.lockedUntil) {
			delete(rl.attempts, id)
			continue
		}
		if now.Sub(rec.lastFail) > rl.window*2 {
			delete(rl.attempts, id)
		}
	}
}

var globalLoginLimiter = newRateLimiter(
	5,
	10*time.Minute,
	5*time.Minute,
)

func sanitizeAuthError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, coreidentity.ErrInvalidCredentials):
		return "invalid email or password"
	case errors.Is(err, coreidentity.ErrEmptyEmail):
		return "email is required"
	case errors.Is(err, coreidentity.ErrEmptyPassword):
		return "password is required"
	case errors.Is(err, coreidentity.ErrSessionExpired):
		return "session expired, please login again"
	case errors.Is(err, coreidentity.ErrSessionNotFound):
		return "session not found"
	case errors.Is(err, coreidentity.ErrEmptyToken):
		return "authentication token is required"
	case errors.Is(err, coreidentity.ErrAuthDisabled):
		return "authentication is disabled"
	case errors.Is(err, coreidentity.ErrAuthenticationMode):
		return "use OIDC login instead"
	default:
		return "authentication service error"
	}
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isValidRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case string(coreidentity.RoleOwner), string(coreidentity.RoleAdmin),
		string(coreidentity.RoleMember), string(coreidentity.RoleViewer):
		return true
	default:
		return false
	}
}

type authLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authLoginResponse struct {
	Success          bool                     `json:"success"`
	Mode             coreidentity.Mode        `json:"mode"`
	Required         bool                     `json:"required"`
	Token            string                   `json:"token"`
	RefreshToken     string                   `json:"refresh_token"`
	ExpiresAt        time.Time                `json:"expires_at"`
	RefreshExpiresAt time.Time                `json:"refresh_expires_at"`
	User             coreidentity.User        `json:"user"`
	Capabilities     []coreidentity.Capability `json:"capabilities"`
	AllowedSections  []string                 `json:"allowed_sections"`
}

type authLogoutResponse struct {
	Success bool `json:"success"`
}

func buildIdentityManager(environment string) *coreidentity.Manager {
	required := parseBoolEnv("BRAIN_AUTH_REQUIRED", false)
	mode := parseAuthMode(os.Getenv("BRAIN_AUTH_MODE"))
	if mode == "" {
		mode = coreidentity.ModeBootstrap
	}

	allowAnonymous := !required || environment == brainenv.Development
	if mode == coreidentity.ModeAnonymous {
		allowAnonymous = true
	}

	manager := coreidentity.NewManager(coreidentity.Config{
		Mode:              mode,
		Required:          required,
		AllowAnonymous:    allowAnonymous,
		StorePath:         defaultAuthStorePath(),
		BootstrapEmail:    strings.TrimSpace(os.Getenv("BRAIN_AUTH_BOOTSTRAP_EMAIL")),
		BootstrapPassword: strings.TrimSpace(os.Getenv("BRAIN_AUTH_BOOTSTRAP_PASSWORD")),
		BootstrapName:     strings.TrimSpace(os.Getenv("BRAIN_AUTH_BOOTSTRAP_NAME")),
		BootstrapRole:     parseAuthRole(os.Getenv("BRAIN_AUTH_BOOTSTRAP_ROLE")),
		SessionTTL:        parseAuthTTL(os.Getenv("BRAIN_AUTH_SESSION_TTL_HOURS")),
		RefreshTTL:        parseRefreshTTL(os.Getenv("BRAIN_AUTH_REFRESH_TTL_DAYS")),
	})

	if store, err := coreidentity.NewStore(defaultAuthStorePath()); err == nil {
		manager.SetStore(store)
	}

	return manager
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func parseAuthMode(raw string) coreidentity.Mode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(coreidentity.ModeOIDC):
		return coreidentity.ModeOIDC
	case string(coreidentity.ModeAnonymous):
		return coreidentity.ModeAnonymous
	default:
		return coreidentity.ModeBootstrap
	}
}

func parseAuthRole(raw string) coreidentity.Role {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	switch trimmed {
	case string(coreidentity.RoleAdmin):
		return coreidentity.RoleAdmin
	case string(coreidentity.RoleMember):
		return coreidentity.RoleMember
	case string(coreidentity.RoleViewer):
		return coreidentity.RoleViewer
	case string(coreidentity.RoleOwner):
		return coreidentity.RoleOwner
	default:
		return ""
	}
}

func parseAuthTTL(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 12 * time.Hour
	}
	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return 12 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

func (d *BrainDaemon) isPublicRoute(method, path string) bool {
	if method == http.MethodOptions {
		return true
	}

	switch {
	case path == "/health":
		return true
	case path == "/api/v1/health":
		return true
	case path == "/metrics":
		return true
	case path == "/api/v1/traces":
		return true
	case path == "/api/status":
		return true
	case path == "/api/providers/available":
		return true
	case strings.HasPrefix(path, "/api/auth/"):
		return true
	case strings.HasPrefix(path, "/api/auth/supabase/"):
		return true
	case path == "/api/invites/consume":
		return true
	default:
		return false
	}
}

func (d *BrainDaemon) extractAuthToken(r *http.Request) string {
	if header := strings.TrimSpace(r.Header.Get("Authorization")); header != "" {
		if token, ok := strings.CutPrefix(header, "Bearer "); ok {
			return strings.TrimSpace(token)
		}
	}

	if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
		return token
	}

	return strings.TrimSpace(r.Header.Get("X-Brain-Token"))
}

func (d *BrainDaemon) authorizeRequest(w http.ResponseWriter, r *http.Request) bool {
	if d == nil || d.auth == nil || !d.auth.Config().Required {
		return true
	}

	token := d.extractAuthToken(r)
	session, err := d.auth.Authenticate(r.Context(), token)
	if err == nil && session != nil {
		return true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "authentication required",
	})
	return false
}

func (d *BrainDaemon) authorizeCapability(w http.ResponseWriter, r *http.Request, required coreidentity.Capability) bool {
	if !d.authorizeRequest(w, r) {
		return false
	}

	status := d.authStatusForRequest(r)
	if status.User == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
		return false
	}

	for _, capability := range status.User.Capabilities {
		if capability == required {
			return true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
	return false
}

func (d *BrainDaemon) authStatusForRequest(r *http.Request) coreidentity.Status {
	if d == nil || d.auth == nil {
		return coreidentity.Status{}
	}
	return d.auth.Status(r.Context(), d.extractAuthToken(r))
}

func (d *BrainDaemon) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
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

	var req authLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = req.Password

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
			"error": "too many login attempts, please try again later",
		})
		return
	}

	session, err := d.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		globalLoginLimiter.recordFailure(req.Email)

		switch {
		case errors.Is(err, coreidentity.ErrInvalidCredentials):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid email or password"})
		case errors.Is(err, coreidentity.ErrAuthDisabled):
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication is disabled"})
		case errors.Is(err, coreidentity.ErrAuthenticationMode):
			w.WriteHeader(http.StatusNotImplemented)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "use OIDC login instead"})
		case errors.Is(err, coreidentity.ErrEmptyEmail), errors.Is(err, coreidentity.ErrEmptyPassword):
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": sanitizeAuthError(err)})
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service error"})
		}
		return
	}

	globalLoginLimiter.mu.Lock()
	delete(globalLoginLimiter.attempts, req.Email)
	globalLoginLimiter.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authLoginResponse{
		Success:          true,
		Mode:             d.auth.Config().Mode,
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

func (d *BrainDaemon) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
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

	if err := d.auth.Logout(r.Context(), token); err != nil && !errors.Is(err, coreidentity.ErrSessionNotFound) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "logout failed"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authLogoutResponse{Success: true})
}

func (d *BrainDaemon) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	status := d.authStatusForRequest(r)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
