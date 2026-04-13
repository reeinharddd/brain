package identity

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Mode describes how the daemon authenticates users.
type Mode string

const (
	ModeBootstrap Mode = "bootstrap"
	ModeOIDC      Mode = "oidc"
	ModeAnonymous Mode = "anonymous"
)

// Role describes the authority level granted to a session.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// Capability is a fine-grained permission granted to a user session.
type Capability string

const (
	CapabilityAuthManage      Capability = "auth:manage"
	CapabilityInfraRead       Capability = "infra:read"
	CapabilityInfraWrite      Capability = "infra:write"
	CapabilityArtifactsRead   Capability = "artifacts:read"
	CapabilityArtifactsWrite  Capability = "artifacts:write"
	CapabilityArtifactsInstall Capability = "artifacts:install"
	CapabilityLogsRead        Capability = "logs:read"
	CapabilitySyncRun         Capability = "sync:run"
	CapabilityMCPManage       Capability = "mcp:manage"
	CapabilityAgentManage     Capability = "agents:manage"
	CapabilityPolicyManage    Capability = "policy:manage"
)

// Config controls how the identity manager behaves.
type Config struct {
	Mode              Mode
	Required          bool
	AllowAnonymous    bool
	StorePath         string
	BootstrapEmail    string
	BootstrapPassword string
	BootstrapName     string
	BootstrapRole     Role
	SessionTTL        time.Duration
	RefreshTTL        time.Duration
}

// User describes the authenticated identity attached to a session.
type User struct {
	ID           string       `json:"id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	Role         Role         `json:"role"`
	Capabilities []Capability `json:"capabilities"`
	Sections     []string     `json:"sections"`
	Provider     string       `json:"provider,omitempty"`
	Subject      string       `json:"subject,omitempty"`
	LastSeenAt   time.Time    `json:"last_seen_at,omitempty"`
}

// Session contains the active bearer token and its user payload.
type Session struct {
	Token     string    `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at,omitempty"`
	LastUsed  time.Time `json:"last_used"`
}

// SessionView is a sanitized session description for status endpoints.
type SessionView struct {
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
}

// Status captures the current authentication posture for the daemon.
type Status struct {
	Required        bool         `json:"required"`
	Mode            Mode         `json:"mode"`
	Authenticated   bool         `json:"authenticated"`
	User            *User        `json:"user,omitempty"`
	Session         *SessionView `json:"session,omitempty"`
	AllowedSections []string     `json:"allowed_sections,omitempty"`
	ActiveSessions  int          `json:"active_sessions"`
	Message         string       `json:"message,omitempty"`
}

var (
	ErrAuthDisabled        = errors.New("authentication is disabled")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrAuthenticationMode  = errors.New("authentication mode is not supported")
	ErrEmptyToken          = errors.New("token is empty")
	ErrEmptyEmail          = errors.New("email is empty")
	ErrEmptyPassword       = errors.New("password is empty")
)

// Manager maintains in-memory sessions and resolves capability sets.
type Manager struct {
	mu       sync.RWMutex
	config   Config
	sessions map[string]*Session
	store    SessionStore
}

// NewManager constructs an identity manager with safe defaults.
func NewManager(config Config) *Manager {
	if config.Mode == "" {
		config.Mode = ModeBootstrap
	}
	if config.SessionTTL <= 0 {
		config.SessionTTL = 12 * time.Hour
	}
	if config.RefreshTTL <= 0 {
		config.RefreshTTL = 30 * 24 * time.Hour
	}
	if config.BootstrapName == "" {
		config.BootstrapName = "Brain User"
	}
	if config.BootstrapRole == "" {
		config.BootstrapRole = RoleOwner
	}
	if !config.Required && !config.AllowAnonymous {
		config.AllowAnonymous = true
	}

	return &Manager{
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// SetStore attaches a persistent store to the manager.
func (m *Manager) SetStore(store SessionStore) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = store
}

// Config returns the active identity configuration.
func (m *Manager) Config() Config {
	if m == nil {
		return Config{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Login validates credentials and creates a new session.
func (m *Manager) Login(ctx context.Context, email, password string) (*Session, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if password == "" {
		return nil, ErrEmptyPassword
	}

	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()

	if config.Mode == ModeOIDC {
		return nil, ErrAuthenticationMode
	}

	if config.BootstrapEmail != "" && config.BootstrapPassword != "" {
		if !constantTimeEqual(email, config.BootstrapEmail) || !constantTimeEqual(password, config.BootstrapPassword) {
			return nil, ErrInvalidCredentials
		}
	} else if !config.AllowAnonymous && config.Required {
		return nil, ErrAuthDisabled
	}

	role := config.BootstrapRole
	if role == "" {
		role = RoleOwner
	}
	user := newUser(email, config.BootstrapName, role)
	session, err := newSession(user, config.SessionTTL, config.RefreshTTL)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sessions[session.Token] = session
	store := m.store
	m.mu.Unlock()

	if store != nil {
		if err := store.UpsertUser(ctx, session.User); err != nil {
			return nil, err
		}
		if err := store.SaveSession(ctx, *session); err != nil {
			return nil, err
		}
	}

	return session, nil
}

// IssueSession creates a new session for a verified user identity.
func (m *Manager) IssueSession(ctx context.Context, user User) (*Session, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(user.ID) == "" {
		return nil, fmt.Errorf("user id is empty")
	}
	if strings.TrimSpace(user.Email) == "" {
		return nil, ErrEmptyEmail
	}

	if user.Role == "" {
		user.Role = RoleMember
	}
	user.Capabilities = CapabilitiesForRole(user.Role)
	user.Sections = SectionsForRole(user.Role)
	if user.LastSeenAt.IsZero() {
		user.LastSeenAt = time.Now().UTC()
	}

	session, err := newSession(user, m.config.SessionTTL, m.config.RefreshTTL)
	if err != nil {
		return nil, err
	}
	if err := m.persistSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// Authenticate validates a bearer token and refreshes last-used metadata.
func (m *Manager) Authenticate(ctx context.Context, token string) (*Session, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrEmptyToken
	}

	m.mu.RLock()
	session, exists := m.sessions[token]
	store := m.store
	m.mu.RUnlock()
	if !exists {
		if store != nil {
			loaded, err := store.GetSessionByToken(ctx, token)
			if err != nil {
				return nil, err
			}
			session = &loaded
			m.mu.Lock()
			m.sessions[token] = session
			m.mu.Unlock()
		} else {
			return nil, ErrSessionNotFound
		}
	}

	now := time.Now().UTC()
	if now.After(session.ExpiresAt) {
		m.mu.Lock()
		delete(m.sessions, token)
		m.mu.Unlock()
		return nil, ErrSessionExpired
	}

	m.mu.Lock()
	session.LastUsed = now
	m.mu.Unlock()

	return session, nil
}

// Logout removes a session by token.
func (m *Manager) Logout(ctx context.Context, token string) error {
	if m == nil {
		return fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return ErrEmptyToken
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	store := m.store

	if _, exists := m.sessions[token]; !exists {
		return ErrSessionNotFound
	}
	delete(m.sessions, token)
	if store != nil {
		if err := store.DeleteSessionByToken(ctx, token); err != nil {
			return err
		}
	}
	return nil
}

// Status returns the current authentication posture and the current session if one exists.
func (m *Manager) Status(ctx context.Context, token string) Status {
	if m == nil {
		return Status{}
	}

	select {
	case <-ctx.Done():
		return Status{Message: ctx.Err().Error()}
	default:
	}

	m.mu.RLock()
	config := m.config
	session, exists := m.sessions[strings.TrimSpace(token)]
	store := m.store
	activeSessions := len(m.sessions)
	m.mu.RUnlock()
	if store != nil {
		if sessions, err := store.ListSessions(ctx); err == nil {
			activeSessions = len(sessions)
		}
	}

	if !exists && store != nil {
		if loaded, err := store.GetSessionByToken(ctx, token); err == nil {
			session = &loaded
			exists = true
		}
	}

	status := Status{
		Required:       config.Required,
		Mode:           config.Mode,
		ActiveSessions: activeSessions,
	}

	if exists && session != nil && !time.Now().UTC().After(session.ExpiresAt) {
		status.Authenticated = true
		user := session.User
		status.User = &user
		status.Session = &SessionView{
			ExpiresAt: session.ExpiresAt,
			LastUsed:  session.LastUsed,
		}
		status.AllowedSections = append([]string(nil), user.Sections...)
		return status
	}

	status.AllowedSections = PublicSections()
	if config.Required {
		status.Message = "authentication required"
	} else {
		status.Message = "login to unlock private sections"
	}
	return status
}

// Refresh rotates an existing refresh token into a brand-new session pair.
func (m *Manager) Refresh(ctx context.Context, refreshToken string) (*Session, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrEmptyToken
	}

	m.mu.RLock()
	store := m.store
	var matched *Session
	for _, session := range m.sessions {
		if session != nil && session.RefreshToken == refreshToken {
			copySession := *session
			matched = &copySession
			break
		}
	}
	m.mu.RUnlock()

	if matched == nil && store != nil {
		loaded, err := store.GetSessionByRefreshToken(ctx, refreshToken)
		if err != nil {
			return nil, err
		}
		matched = &loaded
	}
	if matched == nil {
		return nil, ErrSessionNotFound
	}
	if time.Now().UTC().After(matched.RefreshExpiresAt) {
		_ = m.revokeSession(ctx, matched.Token, refreshToken)
		return nil, ErrSessionExpired
	}

	rotated, err := newSession(matched.User, m.config.SessionTTL, m.config.RefreshTTL)
	if err != nil {
		return nil, err
	}
	rotated.User.LastSeenAt = time.Now().UTC()

	if err := m.persistSession(ctx, rotated); err != nil {
		return nil, err
	}
	if err := m.revokeSession(ctx, matched.Token, refreshToken); err != nil && !errors.Is(err, ErrSessionNotFound) {
		return nil, err
	}

	return rotated, nil
}

// UpsertUser stores or updates a user record.
func (m *Manager) UpsertUser(ctx context.Context, user User) error {
	if m == nil {
		return fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	user.Capabilities = CapabilitiesForRole(user.Role)
	user.Sections = SectionsForRole(user.Role)
	if user.LastSeenAt.IsZero() {
		user.LastSeenAt = time.Now().UTC()
	}

	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()
	if store != nil {
		return store.UpsertUser(ctx, user)
	}

	return nil
}

// ListUsers returns the persisted user directory.
func (m *Manager) ListUsers(ctx context.Context) ([]User, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.RLock()
	store := m.store
	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	m.mu.RUnlock()

	if store != nil {
		return store.ListUsers(ctx)
	}

	result := make([]User, 0, len(sessions))
	seen := make(map[string]struct{})
	for _, session := range sessions {
		if session == nil || session.User.ID == "" {
			continue
		}
		if _, exists := seen[session.User.ID]; exists {
			continue
		}
		seen[session.User.ID] = struct{}{}
		user := session.User
		result = append(result, user)
	}
	return result, nil
}

// UpdateUserRole updates a user's role and refreshes active sessions for that identity.
func (m *Manager) UpdateUserRole(ctx context.Context, userID string, role Role) (*User, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrEmptyEmail
	}

	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()

	var user User
	var err error
	if store != nil {
		user, err = store.GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}
	} else {
		m.mu.RLock()
		for _, session := range m.sessions {
			if session != nil && session.User.ID == userID {
				user = session.User
				break
			}
		}
		m.mu.RUnlock()
	}

	user.Role = role
	user.Capabilities = CapabilitiesForRole(role)
	user.Sections = SectionsForRole(role)
	user.LastSeenAt = time.Now().UTC()
	if err := m.UpsertUser(ctx, user); err != nil {
		return nil, err
	}

	m.mu.Lock()
	for token, session := range m.sessions {
		if session != nil && session.User.ID == userID {
			session.User = user
			m.sessions[token] = session
		}
	}
	m.mu.Unlock()

	return &user, nil
}

// CreateInvite stores an invitation for later onboarding and returns the persisted invite.
func (m *Manager) CreateInvite(ctx context.Context, invite Invite) (Invite, error) {
	if m == nil {
		return Invite{}, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return Invite{}, ctx.Err()
	default:
	}

	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = time.Now().UTC()
	}
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = invite.CreatedAt.Add(14 * 24 * time.Hour)
	}
	if invite.Token == "" {
		token, err := generateToken()
		if err != nil {
			return Invite{}, err
		}
		invite.Token = token
	}

	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()
	if store != nil {
		if err := store.UpsertInvite(ctx, invite); err != nil {
			return Invite{}, err
		}
	}
	return invite, nil
}

// ListInvites returns active invitations.
func (m *Manager) ListInvites(ctx context.Context) ([]Invite, error) {
	if m == nil {
		return nil, fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()
	if store != nil {
		return store.ListInvites(ctx)
	}
	return nil, nil
}

// ConsumeInvite marks an invite as used.
func (m *Manager) ConsumeInvite(ctx context.Context, token string) error {
	if m == nil {
		return fmt.Errorf("identity manager is nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()
	if store != nil {
		return store.ConsumeInvite(ctx, token, time.Now().UTC())
	}
	return nil
}

func (m *Manager) persistSession(ctx context.Context, session *Session) error {
	if m == nil || session == nil {
		return fmt.Errorf("identity manager is nil")
	}

	m.mu.Lock()
	m.sessions[session.Token] = session
	store := m.store
	m.mu.Unlock()

	if store == nil {
		return nil
	}
	if err := store.UpsertUser(ctx, session.User); err != nil {
		return err
	}
	if err := store.SaveSession(ctx, *session); err != nil {
		return err
	}
	return nil
}

func (m *Manager) revokeSession(ctx context.Context, token, refreshToken string) error {
	if m == nil {
		return fmt.Errorf("identity manager is nil")
	}

	token = strings.TrimSpace(token)
	refreshToken = strings.TrimSpace(refreshToken)

	m.mu.Lock()
	if token != "" {
		delete(m.sessions, token)
	}
	store := m.store
	m.mu.Unlock()

	if store == nil {
		return nil
	}
	if token != "" {
		if err := store.DeleteSessionByToken(ctx, token); err != nil && !errors.Is(err, ErrSessionNotFound) {
			return err
		}
	}
	if refreshToken != "" {
		if err := store.DeleteSessionByRefreshToken(ctx, refreshToken); err != nil && !errors.Is(err, ErrSessionNotFound) {
			return err
		}
	}
	return nil
}

// PublicSections returns the sections available before login.
func PublicSections() []string {
	return []string{"runtime", "samples", "reference"}
}

// DefaultSections returns the canonical desktop sections available to the widest audience.
func DefaultSections() []string {
	return []string{"runtime", "agents", "memory", "rules", "mcp-tools", "logs", "evals", "samples", "reference"}
}

// SectionsForRole returns the sections a role may access.
func SectionsForRole(role Role) []string {
	switch role {
	case RoleOwner, RoleAdmin:
		return DefaultSections()
	case RoleMember:
		return []string{"runtime", "agents", "memory", "mcp-tools", "logs", "samples", "reference"}
	case RoleViewer:
		return []string{"runtime", "memory", "logs", "reference"}
	default:
		return []string{"runtime", "reference"}
	}
}

// CapabilitiesForRole returns the fine-grained permissions for a role.
func CapabilitiesForRole(role Role) []Capability {
	switch role {
	case RoleOwner:
		return []Capability{
			CapabilityAuthManage,
			CapabilityInfraRead,
			CapabilityInfraWrite,
			CapabilityArtifactsRead,
			CapabilityArtifactsWrite,
			CapabilityArtifactsInstall,
			CapabilityLogsRead,
			CapabilitySyncRun,
			CapabilityMCPManage,
			CapabilityAgentManage,
			CapabilityPolicyManage,
		}
	case RoleAdmin:
		return []Capability{
			CapabilityInfraRead,
			CapabilityInfraWrite,
			CapabilityArtifactsRead,
			CapabilityArtifactsWrite,
			CapabilityArtifactsInstall,
			CapabilityLogsRead,
			CapabilitySyncRun,
			CapabilityMCPManage,
			CapabilityAgentManage,
			CapabilityPolicyManage,
		}
	case RoleMember:
		return []Capability{
			CapabilityInfraRead,
			CapabilityArtifactsRead,
			CapabilityArtifactsInstall,
			CapabilityLogsRead,
			CapabilitySyncRun,
			CapabilityMCPManage,
			CapabilityAgentManage,
		}
	case RoleViewer:
		return []Capability{
			CapabilityInfraRead,
			CapabilityArtifactsRead,
			CapabilityLogsRead,
		}
	default:
		return []Capability{CapabilityInfraRead}
	}
}

func newUser(email, name string, role Role) User {
	if name == "" {
		name = strings.TrimSpace(email)
	}
	if name == "" {
		name = "Brain User"
	}

	return User{
		ID:           normalizeUserID(email),
		Email:        email,
		Name:         name,
		Role:         role,
		Capabilities: CapabilitiesForRole(role),
		Sections:     SectionsForRole(role),
	}
}

func newSession(user User, ttl, refreshTTL time.Duration) (*Session, error) {
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}
	refreshToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now().UTC()
	return &Session{
		Token:             token,
		RefreshToken:      refreshToken,
		User:              user,
		CreatedAt:         now,
		ExpiresAt:         now.Add(ttl),
		RefreshExpiresAt:  now.Add(refreshTTL),
		LastUsed:          now,
	}, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func normalizeUserID(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "brain-user"
	}
	return strings.NewReplacer("@", "-", ".", "-").Replace(email)
}

func constantTimeEqual(left, right string) bool {
	leftBytes := []byte(left)
	rightBytes := []byte(right)
	if len(leftBytes) != len(rightBytes) {
		return false
	}
	return subtle.ConstantTimeCompare(leftBytes, rightBytes) == 1
}
