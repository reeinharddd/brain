package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	coreenterprise "github.com/reeinharrrd/brain/core/enterprise"
	coreidentity "github.com/reeinharrrd/brain/core/identity"
	brainenv "github.com/reeinharrrd/brain/daemon/internal/environment"
)

type oidcStartResponse struct {
	Success          bool              `json:"success"`
	Provider         string            `json:"provider"`
	State            string            `json:"state"`
	AuthorizationURL string            `json:"authorization_url"`
	ExpiresAt        time.Time         `json:"expires_at"`
}

type oidcCallbackResponse struct {
	Success          bool               `json:"success"`
	State            string             `json:"state"`
	Token            string             `json:"token,omitempty"`
	RefreshToken     string             `json:"refresh_token,omitempty"`
	ExpiresAt        time.Time          `json:"expires_at,omitempty"`
	RefreshExpiresAt time.Time          `json:"refresh_expires_at,omitempty"`
	User             *coreidentity.User `json:"user,omitempty"`
	Message          string             `json:"message,omitempty"`
}

type oidcPollResponse struct {
	Ready   bool                `json:"ready"`
	State   string              `json:"state"`
	Session *oidcCallbackResponse `json:"session,omitempty"`
	Message string              `json:"message,omitempty"`
}

func buildOIDCManager(environment string) *coreenterprise.SSOManager {
	issuerURL := strings.TrimSpace(os.Getenv("BRAIN_OIDC_ISSUER_URL"))
	clientID := strings.TrimSpace(os.Getenv("BRAIN_OIDC_CLIENT_ID"))
	redirectURL := strings.TrimSpace(os.Getenv("BRAIN_OIDC_REDIRECT_URL"))
	if redirectURL == "" {
		redirectURL = "http://127.0.0.1:9090/api/auth/oidc/callback"
	}

	enabled := parseBoolEnv("BRAIN_OIDC_ENABLED", false) || parseAuthMode(os.Getenv("BRAIN_AUTH_MODE")) == coreidentity.ModeOIDC
	if environment == brainenv.Development && issuerURL == "" {
		enabled = false
	}

	return coreenterprise.NewSSOManager(coreenterprise.SSOConfig{
		Provider:       parseOIDCProvider(os.Getenv("BRAIN_OIDC_PROVIDER")),
		IssuerURL:      issuerURL,
		ClientID:       clientID,
		ClientSecret:   strings.TrimSpace(os.Getenv("BRAIN_OIDC_CLIENT_SECRET")),
		RedirectURL:    redirectURL,
		Scopes:         parseOIDCScopes(os.Getenv("BRAIN_OIDC_SCOPES")),
		Enabled:        enabled,
		LoginHint:      strings.TrimSpace(os.Getenv("BRAIN_OIDC_LOGIN_HINT")),
		Prompt:         strings.TrimSpace(os.Getenv("BRAIN_OIDC_PROMPT")),
		TransactionTTL: parseOIDCTTL(os.Getenv("BRAIN_OIDC_TRANSACTION_TTL_MINUTES")),
	})
}

func defaultAuthStorePath() string {
	if override := strings.TrimSpace(os.Getenv("BRAIN_AUTH_STORE_PATH")); override != "" {
		return override
	}
	home, err := os.UserConfigDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "brain", "auth.sqlite")
	}
	return filepath.Join(home, "brain", "auth.sqlite")
}

func parseRefreshTTL(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 30 * 24 * time.Hour
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 {
		return 30 * 24 * time.Hour
	}
	return time.Duration(days) * 24 * time.Hour
}

func parseOIDCTTL(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 10 * time.Minute
	}
	minutes, err := strconv.Atoi(raw)
	if err != nil || minutes <= 0 {
		return 10 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

func parseOIDCProvider(raw string) coreenterprise.SSOProvider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(coreenterprise.ProviderOkta):
		return coreenterprise.ProviderOkta
	case string(coreenterprise.ProviderAuth0):
		return coreenterprise.ProviderAuth0
	case string(coreenterprise.ProviderKeycloak):
		return coreenterprise.ProviderKeycloak
	case string(coreenterprise.ProviderGoogle):
		return coreenterprise.ProviderGoogle
	case string(coreenterprise.ProviderGitHub):
		return coreenterprise.ProviderGitHub
	default:
		return coreenterprise.ProviderLogto
	}
}

func parseOIDCScopes(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{"openid", "profile", "email"}
	}
	return result
}

func (d *BrainDaemon) handleOIDCStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.oidc == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "oidc service unavailable"})
		return
	}
	if !d.oidc.GetConfig().Enabled {
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "oidc login is not enabled"})
		return
	}

	loginHint := strings.TrimSpace(r.URL.Query().Get("login_hint"))
	if loginHint == "" {
		loginHint = strings.TrimSpace(r.URL.Query().Get("email"))
	}
	request, err := d.oidc.StartLogin(r.Context(), loginHint)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to start oidc login: " + sanitizeAuthError(err)})
		return
	}

	_ = json.NewEncoder(w).Encode(oidcStartResponse{
		Success:          true,
		Provider:         string(d.oidc.GetConfig().Provider),
		State:            request.State,
		AuthorizationURL: request.AuthorizationURL,
		ExpiresAt:        request.ExpiresAt,
	})
}

func (d *BrainDaemon) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.oidc == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "oidc service unavailable"})
		return
	}

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if state == "" || code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing code or state"})
		return
	}

	if errMsg := strings.TrimSpace(r.URL.Query().Get("error")); errMsg != "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
		return
	}

	providerUser, err := d.oidc.CompleteLogin(r.Context(), state, code)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "oidc provider error: " + sanitizeAuthError(err)})
		return
	}

	user := d.resolveOIDCUser(r.Context(), providerUser)
	session, err := d.auth.IssueSession(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create session"})
		return
	}

	d.cacheOIDCSession(state, session)
	_ = json.NewEncoder(w).Encode(oidcCallbackResponse{
		Success:          true,
		State:            state,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		User:             &session.User,
		Message:          "login complete",
	})
}

func (d *BrainDaemon) handleOIDCPoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "oidc service unavailable"})
		return
	}

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing state"})
		return
	}

	if session, ok := d.consumeOIDCSession(state); ok && session != nil {
		_ = json.NewEncoder(w).Encode(oidcPollResponse{
			Ready:   true,
			State:   state,
			Session: &oidcCallbackResponse{Success: true, State: state, Token: session.Token, RefreshToken: session.RefreshToken, ExpiresAt: session.ExpiresAt, RefreshExpiresAt: session.RefreshExpiresAt, User: &session.User, Message: "login complete"},
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(oidcPollResponse{Ready: false, State: state, Message: "waiting for login callback"})
}

func (d *BrainDaemon) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
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

	refreshToken := d.extractRefreshToken(r)
	if refreshToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh token is required"})
		return
	}

	session, err := d.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, coreidentity.ErrSessionExpired):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh token expired, please login again"})
		case errors.Is(err, coreidentity.ErrSessionNotFound):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh token not found"})
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "refresh failed"})
		}
		return
	}

	_ = json.NewEncoder(w).Encode(oidcCallbackResponse{
		Success:          true,
		State:            "refresh",
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		User:             &session.User,
		Message:          "session refreshed",
	})
}

func (d *BrainDaemon) resolveOIDCUser(ctx context.Context, providerUser *coreenterprise.SSOUser) coreidentity.User {
	if providerUser == nil {
		return coreidentity.User{}
	}

	role := parseAuthRole(os.Getenv("BRAIN_OIDC_DEFAULT_ROLE"))
	if role == "" {
		role = coreidentity.RoleMember
	}

	if existing, err := d.lookupUserByIdentity(ctx, providerUser); err == nil {
		role = existing.Role
	} else if invite, err := d.lookupInviteByEmail(ctx, providerUser.Email); err == nil {
		role = invite.Role
		_ = d.auth.ConsumeInvite(ctx, invite.Token)
	}

	user := coreidentity.User{
		ID:           providerUser.ID,
		Email:        providerUser.Email,
		Name:         providerUser.Name,
		Role:         role,
		Provider:     string(providerUser.Provider),
		Subject:      providerUser.Subject,
		LastSeenAt:   time.Now().UTC(),
	}
	user.Capabilities = coreidentity.CapabilitiesForRole(role)
	user.Sections = coreidentity.SectionsForRole(role)
	_ = d.auth.UpsertUser(ctx, user)
	return user
}

func (d *BrainDaemon) lookupUserByIdentity(ctx context.Context, providerUser *coreenterprise.SSOUser) (*coreidentity.User, error) {
	if d == nil || d.auth == nil || providerUser == nil {
		return nil, fmt.Errorf("user lookup unavailable")
	}
	users, err := d.auth.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if strings.EqualFold(user.Email, providerUser.Email) || user.Subject == providerUser.Subject {
			copyUser := user
			return &copyUser, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (d *BrainDaemon) lookupInviteByEmail(ctx context.Context, email string) (*coreidentity.Invite, error) {
	if d == nil || d.auth == nil {
		return nil, fmt.Errorf("invite lookup unavailable")
	}
	invites, err := d.auth.ListInvites(ctx)
	if err != nil {
		return nil, err
	}
	for _, invite := range invites {
		if strings.EqualFold(invite.Email, email) && invite.ConsumedAt.IsZero() {
			copyInvite := invite
			return &copyInvite, nil
		}
	}
	return nil, fmt.Errorf("invite not found")
}

func (d *BrainDaemon) cacheOIDCSession(state string, session *coreidentity.Session) {
	if d == nil || session == nil {
		return
	}
	d.mu.Lock()
	if d.oidcSessions == nil {
		d.oidcSessions = make(map[string]*coreidentity.Session)
	}
	d.oidcSessions[strings.TrimSpace(state)] = session
	d.mu.Unlock()
}

func (d *BrainDaemon) consumeOIDCSession(state string) (*coreidentity.Session, bool) {
	if d == nil {
		return nil, false
	}
	state = strings.TrimSpace(state)
	if state == "" {
		return nil, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	session, ok := d.oidcSessions[state]
	if !ok {
		return nil, false
	}
	delete(d.oidcSessions, state)
	return session, true
}

func (d *BrainDaemon) extractRefreshToken(r *http.Request) string {
	if header := strings.TrimSpace(r.Header.Get("Authorization")); header != "" {
		if token, ok := strings.CutPrefix(header, "Bearer "); ok {
			return strings.TrimSpace(token)
		}
	}
	if token := strings.TrimSpace(r.URL.Query().Get("refresh_token")); token != "" {
		return token
	}
	if token := strings.TrimSpace(r.Header.Get("X-Brain-Refresh-Token")); token != "" {
		return token
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return ""
		}
	}
	return strings.TrimSpace(req.RefreshToken)
}
