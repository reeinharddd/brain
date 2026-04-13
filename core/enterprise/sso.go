package enterprise

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// SSOProvider defines supported SSO providers.
type SSOProvider string

const (
	ProviderOkta     SSOProvider = "okta"
	ProviderAuth0    SSOProvider = "auth0"
	ProviderKeycloak SSOProvider = "keycloak"
	ProviderGoogle   SSOProvider = "google"
	ProviderGitHub   SSOProvider = "github"
	ProviderLogto    SSOProvider = "logto"
)

// SSOConfig holds OIDC configuration for a provider-backed login flow.
type SSOConfig struct {
	Provider      SSOProvider
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	Enabled       bool
	LoginHint     string
	Prompt        string
	TransactionTTL time.Duration
}

// SSOUser represents an authenticated OIDC user resolved from the provider.
type SSOUser struct {
	ID           string            `json:"id"`
	Subject      string            `json:"subject"`
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	Groups       []string          `json:"groups"`
	Attributes   map[string]string  `json:"attributes"`
	AuthTime     time.Time         `json:"auth_time"`
	Expiry       time.Time         `json:"expiry"`
	Issuer       string            `json:"issuer"`
	Provider     SSOProvider       `json:"provider"`
	EmailVerified bool              `json:"email_verified"`
}

// LoginRequest is a pending authorization-code flow that the UI or CLI can complete.
type LoginRequest struct {
	State           string
	AuthorizationURL string
	CodeVerifier     string
	Nonce           string
	ExpiresAt       time.Time
	LoginHint       string
}

// SSOManager handles OIDC login initiation and callback completion.
type SSOManager struct {
	mu      sync.RWMutex
	config  SSOConfig
	pending map[string]*LoginRequest
	completed map[string]*SSOUser
}

func NewSSOManager(config SSOConfig) *SSOManager {
	if config.TransactionTTL <= 0 {
		config.TransactionTTL = 10 * time.Minute
	}
	if len(config.Scopes) == 0 {
		config.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	return &SSOManager{
		config:   config,
		pending:  make(map[string]*LoginRequest),
		completed: make(map[string]*SSOUser),
	}
}

// GetConfig returns a copy of the OIDC configuration.
func (m *SSOManager) GetConfig() SSOConfig {
	if m == nil {
		return SSOConfig{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// StartLogin creates a PKCE authorization request and stores the pending state.
func (m *SSOManager) StartLogin(ctx context.Context, loginHint string) (*LoginRequest, error) {
	if m == nil {
		return nil, fmt.Errorf("sso manager is nil")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	config := m.GetConfig()
	if !config.Enabled {
		return nil, fmt.Errorf("sso is not enabled")
	}
	if strings.TrimSpace(config.IssuerURL) == "" {
		return nil, fmt.Errorf("oidc issuer url is required")
	}
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, fmt.Errorf("oidc client id is required")
	}
	if strings.TrimSpace(config.RedirectURL) == "" {
		return nil, fmt.Errorf("oidc redirect url is required")
	}

	provider, err := oidc.NewProvider(ctx, strings.TrimSpace(config.IssuerURL))
	if err != nil {
		return nil, fmt.Errorf("discover oidc provider: %w", err)
	}

	state, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	nonce, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	verifier, err := randomToken(48)
	if err != nil {
		return nil, err
	}
	challenge := pkceChallenge(verifier)

	endpoint := provider.Endpoint()
	authorizationURL, err := url.Parse(endpoint.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("parse oidc auth url: %w", err)
	}

	query := authorizationURL.Query()
	query.Set("response_type", "code")
	query.Set("client_id", config.ClientID)
	query.Set("redirect_uri", config.RedirectURL)
	query.Set("scope", strings.Join(config.Scopes, " "))
	query.Set("state", state)
	query.Set("nonce", nonce)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	if strings.TrimSpace(loginHint) != "" {
		query.Set("login_hint", strings.TrimSpace(loginHint))
	}
	if strings.TrimSpace(config.Prompt) != "" {
		query.Set("prompt", strings.TrimSpace(config.Prompt))
	}
	authorizationURL.RawQuery = query.Encode()

	request := &LoginRequest{
		State:           state,
		AuthorizationURL: authorizationURL.String(),
		CodeVerifier:     verifier,
		Nonce:           nonce,
		ExpiresAt:       time.Now().UTC().Add(config.TransactionTTL),
		LoginHint:       strings.TrimSpace(loginHint),
	}

	m.mu.Lock()
	m.pending[state] = request
	m.mu.Unlock()

	return request, nil
}

// CompleteLogin exchanges the authorization code for OIDC identity claims.
func (m *SSOManager) CompleteLogin(ctx context.Context, state, code string) (*SSOUser, error) {
	if m == nil {
		return nil, fmt.Errorf("sso manager is nil")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	state = strings.TrimSpace(state)
	code = strings.TrimSpace(code)
	if state == "" || code == "" {
		return nil, fmt.Errorf("state and code are required")
	}

	request, err := m.popPending(state)
	if err != nil {
		return nil, err
	}
	if time.Now().UTC().After(request.ExpiresAt) {
		return nil, fmt.Errorf("oidc login request expired")
	}

	config := m.GetConfig()
	provider, err := oidc.NewProvider(ctx, strings.TrimSpace(config.IssuerURL))
	if err != nil {
		return nil, fmt.Errorf("discover oidc provider: %w", err)
	}

	oauthConfig := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
	}

	token, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(request.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("exchange oidc code: %w", err)
	}

	idTokenRaw, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(idTokenRaw) == "" {
		return nil, fmt.Errorf("oidc token response missing id_token")
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})
	idToken, err := verifier.Verify(ctx, idTokenRaw)
	if err != nil {
		return nil, fmt.Errorf("verify oidc id token: %w", err)
	}
	if idToken.Nonce != request.Nonce {
		return nil, fmt.Errorf("oidc nonce mismatch")
	}

	var claims struct {
		Subject       string            `json:"sub"`
		Email         string            `json:"email"`
		Name          string            `json:"name"`
		EmailVerified bool              `json:"email_verified"`
		Groups        []string          `json:"groups"`
		Profile       string            `json:"profile"`
		Picture       string            `json:"picture"`
		UpdatedAt     string            `json:"updated_at"`
		Extra         map[string]string  `json:"-"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("decode oidc claims: %w", err)
	}

	attributes := map[string]string{
		"profile": claims.Profile,
		"picture": claims.Picture,
		"updated_at": claims.UpdatedAt,
	}

	user := &SSOUser{
		ID:           normalizeIdentityID(claims.Email, claims.Subject),
		Subject:      claims.Subject,
		Email:        strings.TrimSpace(claims.Email),
		Name:         strings.TrimSpace(claims.Name),
		Groups:       append([]string(nil), claims.Groups...),
		Attributes:   cleanAttributes(attributes),
		AuthTime:     time.Now().UTC(),
		Expiry:       idToken.Expiry,
		Issuer:       strings.TrimSpace(config.IssuerURL),
		Provider:     config.Provider,
		EmailVerified: claims.EmailVerified,
	}

	m.mu.Lock()
	m.completed[state] = user
	m.mu.Unlock()

	return user, nil
}

func (m *SSOManager) popPending(state string) (*LoginRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	request, exists := m.pending[state]
	if !exists {
		return nil, fmt.Errorf("oidc login request not found")
	}
	delete(m.pending, state)
	return request, nil
}

func (m *SSOManager) CancelLogin(state string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	delete(m.pending, strings.TrimSpace(state))
	delete(m.completed, strings.TrimSpace(state))
	m.mu.Unlock()
}

// ConsumeCompletedLogin returns the completed OIDC identity for the supplied state and clears it.
func (m *SSOManager) ConsumeCompletedLogin(state string) (*SSOUser, bool) {
	if m == nil {
		return nil, false
	}
	state = strings.TrimSpace(state)
	if state == "" {
		return nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.completed[state]
	if !ok {
		return nil, false
	}
	delete(m.completed, state)
	return user, true
}

func (m *SSOManager) CleanupExpired() int {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	removed := 0
	for state, request := range m.pending {
		if now.After(request.ExpiresAt) {
			delete(m.pending, state)
			removed++
		}
	}
	return removed
}

func randomToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func cleanAttributes(attributes map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range attributes {
		if strings.TrimSpace(value) == "" {
			continue
		}
		result[key] = strings.TrimSpace(value)
	}
	return result
}

func normalizeIdentityID(email, subject string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	subject = strings.TrimSpace(subject)
	if email != "" {
		return strings.NewReplacer("@", "-", ".", "-").Replace(email)
	}
	if subject != "" {
		return strings.NewReplacer("/", "-", ":", "-").Replace(subject)
	}
	return "brain-user"
}

