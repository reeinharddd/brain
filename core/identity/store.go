package identity

import (
	"context"
	"time"

	"github.com/reeinharrrd/brain/core/governance"
)

// Invite represents a pending account invitation or role assignment request.
type Invite struct {
	Token      string    `json:"token"`
	Email      string    `json:"email"`
	Role       Role      `json:"role"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	ConsumedAt time.Time `json:"consumed_at,omitempty"`
}

// SessionStore persists identity users, sessions, invites, and audit entries.
type SessionStore interface {
	Close() error
	UpsertUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, id string) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	SaveSession(ctx context.Context, session Session) error
	GetSessionByToken(ctx context.Context, token string) (Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (Session, error)
	DeleteSessionByToken(ctx context.Context, token string) error
	DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error
	ListSessions(ctx context.Context) ([]Session, error)
	AppendAudit(ctx context.Context, entry governance.AuditEntry) error
	UpsertInvite(ctx context.Context, invite Invite) error
	ListInvites(ctx context.Context) ([]Invite, error)
	GetInviteByToken(ctx context.Context, token string) (Invite, error)
	ConsumeInvite(ctx context.Context, token string, consumedAt time.Time) error
}
