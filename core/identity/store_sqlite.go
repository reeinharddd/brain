package identity

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/reeinharrrd/brain/core/governance"
	_ "modernc.org/sqlite"
)

type sqliteSessionStore struct {
	db *sql.DB
}

// NewStore opens a persistent SQLite-backed identity store.
func NewStore(path string) (SessionStore, error) {
	return newSessionStore(path)
}

func newSessionStore(path string) (SessionStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("store path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite store: %w", err)
	}

	store := &sqliteSessionStore{db: db}
	if err := store.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *sqliteSessionStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *sqliteSessionStore) init(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL,
			provider TEXT NOT NULL DEFAULT '',
			subject TEXT NOT NULL DEFAULT '',
			last_seen_at TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token_hash TEXT PRIMARY KEY,
			refresh_hash TEXT NOT NULL UNIQUE,
			user_id TEXT NOT NULL,
			token_expires_at TEXT NOT NULL,
			refresh_expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_used TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS invites (
			token TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			role TEXT NOT NULL,
			created_by TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			consumed_at TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS audits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			audit_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			action TEXT NOT NULL,
			subject TEXT NOT NULL,
			resource TEXT NOT NULL,
			success INTEGER NOT NULL,
			details_json TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_invites_email ON invites(email);`,
		`CREATE INDEX IF NOT EXISTS idx_audits_action ON audits(action);`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("init sqlite store: %w", err)
		}
	}

	return nil
}

func (s *sqliteSessionStore) UpsertUser(ctx context.Context, user User) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	if strings.TrimSpace(user.ID) == "" {
		return fmt.Errorf("user id is empty")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	lastSeen := ""
	if !user.LastSeenAt.IsZero() {
		lastSeen = user.LastSeenAt.UTC().Format(time.RFC3339Nano)
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, email, name, role, provider, subject, last_seen_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			email=excluded.email,
			name=excluded.name,
			role=excluded.role,
			provider=excluded.provider,
			subject=excluded.subject,
			last_seen_at=excluded.last_seen_at,
			updated_at=excluded.updated_at
	`, user.ID, user.Email, user.Name, string(user.Role), user.Provider, user.Subject, lastSeen, now, now)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}

func (s *sqliteSessionStore) GetUser(ctx context.Context, id string) (User, error) {
	if s == nil || s.db == nil {
		return User{}, fmt.Errorf("sqlite store is nil")
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, name, role, provider, subject, last_seen_at
		FROM users
		WHERE id = ?
	`, strings.TrimSpace(id))
	return scanUserRow(row)
}

func (s *sqliteSessionStore) ListUsers(ctx context.Context) ([]User, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("sqlite store is nil")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, email, name, role, provider, subject, last_seen_at
		FROM users
		ORDER BY updated_at DESC, email ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	result := make([]User, 0)
	for rows.Next() {
		user, err := scanUserRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users rows: %w", err)
	}
	return result, nil
}

func (s *sqliteSessionStore) SaveSession(ctx context.Context, session Session) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	if strings.TrimSpace(session.Token) == "" {
		return fmt.Errorf("session token is empty")
	}
	if strings.TrimSpace(session.RefreshToken) == "" {
		return fmt.Errorf("session refresh token is empty")
	}
	if strings.TrimSpace(session.User.ID) == "" {
		return fmt.Errorf("session user id is empty")
	}

	if err := s.UpsertUser(ctx, session.User); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (token_hash, refresh_hash, user_id, token_expires_at, refresh_expires_at, created_at, last_used)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(token_hash) DO UPDATE SET
			refresh_hash=excluded.refresh_hash,
			user_id=excluded.user_id,
			token_expires_at=excluded.token_expires_at,
			refresh_expires_at=excluded.refresh_expires_at,
			last_used=excluded.last_used
	`, hashToken(session.Token), hashToken(session.RefreshToken), session.User.ID, session.ExpiresAt.UTC().Format(time.RFC3339Nano), session.RefreshExpiresAt.UTC().Format(time.RFC3339Nano), session.CreatedAt.UTC().Format(time.RFC3339Nano), session.LastUsed.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `UPDATE users SET last_seen_at = ?, updated_at = ? WHERE id = ?`, session.LastUsed.UTC().Format(time.RFC3339Nano), now, session.User.ID)
	if err != nil {
		return fmt.Errorf("update user last seen: %w", err)
	}

	return nil
}

func (s *sqliteSessionStore) GetSessionByToken(ctx context.Context, token string) (Session, error) {
	return s.getSession(ctx, `SELECT token_hash, refresh_hash, user_id, token_expires_at, refresh_expires_at, created_at, last_used FROM sessions WHERE token_hash = ?`, hashToken(token), token, false)
}

func (s *sqliteSessionStore) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (Session, error) {
	return s.getSession(ctx, `SELECT token_hash, refresh_hash, user_id, token_expires_at, refresh_expires_at, created_at, last_used FROM sessions WHERE refresh_hash = ?`, hashToken(refreshToken), refreshToken, true)
}

func (s *sqliteSessionStore) DeleteSessionByToken(ctx context.Context, token string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, hashToken(token))
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *sqliteSessionStore) DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE refresh_hash = ?`, hashToken(refreshToken))
	if err != nil {
		return fmt.Errorf("delete refresh session: %w", err)
	}
	return nil
}

func (s *sqliteSessionStore) ListSessions(ctx context.Context) ([]Session, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("sqlite store is nil")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT token_hash, refresh_hash, user_id, token_expires_at, refresh_expires_at, created_at, last_used
		FROM sessions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	result := make([]Session, 0)
	for rows.Next() {
		var tokenHash, refreshHash, userID, tokenExpiresAt, refreshExpiresAt, createdAt, lastUsed string
		if err := rows.Scan(&tokenHash, &refreshHash, &userID, &tokenExpiresAt, &refreshExpiresAt, &createdAt, &lastUsed); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		user, err := s.GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		tokenHashValue, _ := fromHash(tokenHash)
		refreshHashValue, _ := fromHash(refreshHash)
		result = append(result, Session{
			Token:             tokenHashValue,
			RefreshToken:       refreshHashValue,
			User:               user,
			CreatedAt:          mustParseTime(createdAt),
			ExpiresAt:          mustParseTime(tokenExpiresAt),
			RefreshExpiresAt:   mustParseTime(refreshExpiresAt),
			LastUsed:           mustParseTime(lastUsed),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list sessions rows: %w", err)
	}
	return result, nil
}

func (s *sqliteSessionStore) AppendAudit(ctx context.Context, entry governance.AuditEntry) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	details, err := json.Marshal(entry.Details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO audits (audit_id, timestamp, action, subject, resource, success, details_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.Timestamp.UTC().Format(time.RFC3339Nano), entry.Action, entry.Subject, entry.Resource, boolToInt(entry.Success), string(details))
	if err != nil {
		return fmt.Errorf("append audit entry: %w", err)
	}
	return nil
}

func (s *sqliteSessionStore) UpsertInvite(ctx context.Context, invite Invite) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	if strings.TrimSpace(invite.Token) == "" {
		return fmt.Errorf("invite token is empty")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO invites (token, email, role, created_by, created_at, expires_at, consumed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(token) DO UPDATE SET
			email=excluded.email,
			role=excluded.role,
			created_by=excluded.created_by,
			expires_at=excluded.expires_at,
			consumed_at=excluded.consumed_at
	`, invite.Token, invite.Email, string(invite.Role), invite.CreatedBy, invite.CreatedAt.UTC().Format(time.RFC3339Nano), invite.ExpiresAt.UTC().Format(time.RFC3339Nano), formatOptionalTime(invite.ConsumedAt))
	if err != nil {
		return fmt.Errorf("upsert invite: %w", err)
	}
	return nil
}

func (s *sqliteSessionStore) ListInvites(ctx context.Context) ([]Invite, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("sqlite store is nil")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT token, email, role, created_by, created_at, expires_at, consumed_at
		FROM invites
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list invites: %w", err)
	}
	defer rows.Close()

	result := make([]Invite, 0)
	for rows.Next() {
		var token, email, role, createdBy, createdAt, expiresAt, consumedAt string
		if err := rows.Scan(&token, &email, &role, &createdBy, &createdAt, &expiresAt, &consumedAt); err != nil {
			return nil, fmt.Errorf("scan invite row: %w", err)
		}
		result = append(result, Invite{
			Token:      token,
			Email:      email,
			Role:       Role(role),
			CreatedBy:  createdBy,
			CreatedAt:  mustParseTime(createdAt),
			ExpiresAt:  mustParseTime(expiresAt),
			ConsumedAt: parseOptionalTime(consumedAt),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list invites rows: %w", err)
	}
	return result, nil
}

func (s *sqliteSessionStore) GetInviteByToken(ctx context.Context, token string) (Invite, error) {
	if s == nil || s.db == nil {
		return Invite{}, fmt.Errorf("sqlite store is nil")
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT token, email, role, created_by, created_at, expires_at, consumed_at
		FROM invites
		WHERE token = ?
	`, strings.TrimSpace(token))
	var invite Invite
	var role, createdAt, expiresAt, consumedAt string
	if err := row.Scan(&invite.Token, &invite.Email, &role, &invite.CreatedBy, &createdAt, &expiresAt, &consumedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invite{}, ErrSessionNotFound
		}
		return Invite{}, fmt.Errorf("get invite: %w", err)
	}
	invite.Role = Role(role)
	invite.CreatedAt = mustParseTime(createdAt)
	invite.ExpiresAt = mustParseTime(expiresAt)
	invite.ConsumedAt = parseOptionalTime(consumedAt)
	return invite, nil
}

func (s *sqliteSessionStore) ConsumeInvite(ctx context.Context, token string, consumedAt time.Time) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store is nil")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE invites SET consumed_at = ? WHERE token = ?`, formatOptionalTime(consumedAt), strings.TrimSpace(token))
	if err != nil {
		return fmt.Errorf("consume invite: %w", err)
	}
	return nil
}

func (s *sqliteSessionStore) getSession(ctx context.Context, query string, hash string, rawToken string, isRefresh bool) (Session, error) {
	if s == nil || s.db == nil {
		return Session{}, fmt.Errorf("sqlite store is nil")
	}
	row := s.db.QueryRowContext(ctx, query, hash)
	var tokenHash, refreshHash, userID, tokenExpiresAt, refreshExpiresAt, createdAt, lastUsed string
	if err := row.Scan(&tokenHash, &refreshHash, &userID, &tokenExpiresAt, &refreshExpiresAt, &createdAt, &lastUsed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	tokenValue := rawToken
	if !isRefresh {
		tokenValue = rawToken
	}
	return Session{
		Token:           tokenValue,
		RefreshToken:     rawToken,
		User:             user,
		CreatedAt:        mustParseTime(createdAt),
		ExpiresAt:        mustParseTime(tokenExpiresAt),
		RefreshExpiresAt: mustParseTime(refreshExpiresAt),
		LastUsed:         mustParseTime(lastUsed),
	}, nil
}

func scanUserRow(row *sql.Row) (User, error) {
	var user User
	var role, provider, subject, lastSeenAt string
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &role, &provider, &subject, &lastSeenAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrSessionNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	user.Role = Role(role)
	user.Provider = provider
	user.Subject = subject
	user.LastSeenAt = parseOptionalTime(lastSeenAt)
	user.Capabilities = CapabilitiesForRole(user.Role)
	user.Sections = SectionsForRole(user.Role)
	return user, nil
}

func scanUserRows(rows *sql.Rows) (User, error) {
	var user User
	var role, provider, subject, lastSeenAt string
	if err := rows.Scan(&user.ID, &user.Email, &user.Name, &role, &provider, &subject, &lastSeenAt); err != nil {
		return User{}, fmt.Errorf("scan user row: %w", err)
	}
	user.Role = Role(role)
	user.Provider = provider
	user.Subject = subject
	user.LastSeenAt = parseOptionalTime(lastSeenAt)
	user.Capabilities = CapabilitiesForRole(user.Role)
	user.Sections = SectionsForRole(user.Role)
	return user, nil
}

func hashToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func mustParseTime(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func parseOptionalTime(value string) time.Time {
	return mustParseTime(value)
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func fromHash(value string) (string, bool) {
	if strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}
