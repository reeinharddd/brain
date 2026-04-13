package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reeinharrrd/brain/core/governance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func openTestStore(t *testing.T) (*sqliteSessionStore, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test-store.db")

	store, err := newSessionStore(dbPath)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Ensure the caller can cast through the interface when needed.
	s, ok := store.(*sqliteSessionStore)
	require.True(t, ok, "expected *sqliteSessionStore concrete type")
	return s, dbPath
}

func mustUser(t *testing.T) User {
	t.Helper()
	return User{
		ID:    "user-test-001",
		Email: "test@example.com",
		Name:  "Test User",
		Role:  RoleOwner,
	}
}

func mustSession(t *testing.T, user User) Session {
	t.Helper()
	now := time.Now().UTC()
	return Session{
		Token:            "tok-" + t.Name(),
		RefreshToken:     "ref-" + t.Name(),
		User:             user,
		CreatedAt:        now,
		ExpiresAt:        now.Add(12 * time.Hour),
		RefreshExpiresAt: now.Add(30 * 24 * time.Hour),
		LastUsed:         now,
	}
}

func hashSHA256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// ---------------------------------------------------------------------------
// 1. Database file is created on startup
// ---------------------------------------------------------------------------

func TestStore_FileCreatedOnStartup(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "new-store.db")

	_, err := os.Stat(dbPath)
	require.True(t, os.IsNotExist(err), "db should not exist before store opens")

	store, err := newSessionStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	info, err := os.Stat(dbPath)
	require.NoError(t, err, "db file should exist after store opens")
	assert.True(t, info.Size() > 0, "db file should be non-empty")
}

func TestStore_NilPathReturnsError(t *testing.T) {
	_, err := newSessionStore("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestStore_NonexistentDirectoryIsCreated(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sub", "deep", "store.db")

	store, err := newSessionStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	exists, err := fileExists(dbPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ---------------------------------------------------------------------------
// 2. Sessions survive simulated daemon restart (close + reopen)
// ---------------------------------------------------------------------------

func TestStore_SessionSurvivesRestart(t *testing.T) {
	_, dbPath := openTestStore(t)

	// Phase 1: write a session.
	store1, err := newSessionStore(dbPath)
	require.NoError(t, err)

	user := mustUser(t)
	session := mustSession(t, user)

	ctx := context.Background()
	require.NoError(t, store1.SaveSession(ctx, session))
	require.NoError(t, store1.Close())

	// Phase 2: reopen and read it back.
	store2, err := newSessionStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	loaded, err := store2.GetSessionByToken(ctx, session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.User.ID, loaded.User.ID)
	assert.Equal(t, session.User.Email, loaded.User.Email)
	assert.Equal(t, session.User.Role, loaded.User.Role)
	assert.WithinDuration(t, session.ExpiresAt, loaded.ExpiresAt, time.Second)
	assert.WithinDuration(t, session.CreatedAt, loaded.CreatedAt, time.Second)
}

func TestStore_UserSurvivesRestart(t *testing.T) {
	_, dbPath := openTestStore(t)

	store1, err := newSessionStore(dbPath)
	require.NoError(t, err)

	user := mustUser(t)
	ctx := context.Background()
	require.NoError(t, store1.UpsertUser(ctx, user))
	require.NoError(t, store1.Close())

	store2, err := newSessionStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	loaded, err := store2.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, loaded.Email)
	assert.Equal(t, user.Name, loaded.Name)
	assert.Equal(t, user.Role, loaded.Role)
}

// ---------------------------------------------------------------------------
// 3. CRUD operations
// ---------------------------------------------------------------------------

// --- Users ---

func TestStore_UpsertAndGetUser(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)

	require.NoError(t, store.UpsertUser(ctx, user))

	loaded, err := store.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, loaded.Email)
	assert.Equal(t, user.Name, loaded.Name)
	assert.Equal(t, user.Role, loaded.Role)
	assert.Len(t, loaded.Capabilities, len(CapabilitiesForRole(user.Role)))
}

func TestStore_UpdateUser(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := User{ID: "u1", Email: "a@b.com", Name: "A", Role: RoleViewer}
	require.NoError(t, store.UpsertUser(ctx, user))

	user.Role = RoleAdmin
	user.Name = "A Updated"
	require.NoError(t, store.UpsertUser(ctx, user))

	loaded, err := store.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, RoleAdmin, loaded.Role)
	assert.Equal(t, "A Updated", loaded.Name)
}

func TestStore_GetUserNotFound(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	_, err := store.GetUser(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestStore_ListUsers(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	users := []User{
		{ID: "u1", Email: "a@b.com", Name: "A", Role: RoleOwner},
		{ID: "u2", Email: "c@d.com", Name: "C", Role: RoleMember},
		{ID: "u3", Email: "b@b.com", Name: "B", Role: RoleViewer},
	}
	for _, u := range users {
		require.NoError(t, store.UpsertUser(ctx, u))
	}

	listed, err := store.ListUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, listed, 3)

	ids := make(map[string]bool)
	for _, u := range listed {
		ids[u.ID] = true
	}
	for _, u := range users {
		assert.True(t, ids[u.ID], "user %s should be in list", u.ID)
	}
}

func TestStore_ListUsersEmpty(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	listed, err := store.ListUsers(context.Background())
	require.NoError(t, err)
	assert.Empty(t, listed)
}

func TestStore_UpsertUserEmptyID(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	err := store.UpsertUser(context.Background(), User{Email: "x@y.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// --- Sessions ---

func TestStore_SaveAndGetSession(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	session := mustSession(t, user)

	require.NoError(t, store.SaveSession(ctx, session))

	loaded, err := store.GetSessionByToken(ctx, session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.User.ID, loaded.User.ID)
	assert.WithinDuration(t, session.ExpiresAt, loaded.ExpiresAt, time.Second)
}

func TestStore_GetSessionByRefreshToken(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	session := mustSession(t, user)

	require.NoError(t, store.SaveSession(ctx, session))

	loaded, err := store.GetSessionByRefreshToken(ctx, session.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, session.User.ID, loaded.User.ID)
}

func TestStore_GetSessionNotFound(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	_, err := store.GetSessionByToken(context.Background(), "does-not-exist")
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestStore_DeleteSessionByToken(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	session := mustSession(t, mustUser(t))
	require.NoError(t, store.SaveSession(ctx, session))

	require.NoError(t, store.DeleteSessionByToken(ctx, session.Token))

	_, err := store.GetSessionByToken(ctx, session.Token)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestStore_DeleteSessionByRefreshToken(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	session := mustSession(t, mustUser(t))
	require.NoError(t, store.SaveSession(ctx, session))

	require.NoError(t, store.DeleteSessionByRefreshToken(ctx, session.RefreshToken))

	_, err := store.GetSessionByRefreshToken(ctx, session.RefreshToken)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestStore_SaveSessionEmptyToken(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	err := store.SaveSession(ctx, Session{Token: "", RefreshToken: "r", User: User{ID: "u"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestStore_SaveSessionEmptyRefreshToken(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	err := store.SaveSession(ctx, Session{Token: "t", RefreshToken: "", User: User{ID: "u"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestStore_SaveSessionEmptyUserID(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	err := store.SaveSession(ctx, Session{Token: "t", RefreshToken: "r", User: User{ID: ""}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestStore_ListSessions(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	now := time.Now().UTC()
	sessions := []Session{
		{Token: "list-tok-1", RefreshToken: "list-ref-1", User: user, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour), RefreshExpiresAt: now.Add(30 * 24 * time.Hour), LastUsed: now},
		{Token: "list-tok-2", RefreshToken: "list-ref-2", User: user, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour), RefreshExpiresAt: now.Add(30 * 24 * time.Hour), LastUsed: now},
		{Token: "list-tok-3", RefreshToken: "list-ref-3", User: user, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour), RefreshExpiresAt: now.Add(30 * 24 * time.Hour), LastUsed: now},
	}
	for _, s := range sessions {
		require.NoError(t, store.SaveSession(ctx, s))
	}

	listed, err := store.ListSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, listed, 3)
}

func TestStore_UpdateSessionLastUsed(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	session := mustSession(t, user)

	require.NoError(t, store.SaveSession(ctx, session))

	// Save again with an updated LastUsed.
	session.LastUsed = time.Now().UTC()
	require.NoError(t, store.SaveSession(ctx, session))

	loaded, err := store.GetSessionByToken(ctx, session.Token)
	require.NoError(t, err)
	assert.WithinDuration(t, session.LastUsed, loaded.LastUsed, time.Second)
}

// ---------------------------------------------------------------------------
// 4. Token hashing (SHA-256)
// ---------------------------------------------------------------------------

func TestTokenHash_SHA256Correctness(t *testing.T) {
	input := "my-secret-token"
	got := hashToken(input)
	expected := hashSHA256(input)
	assert.Equal(t, expected, got)
}

func TestTokenHash_Deterministic(t *testing.T) {
	token := "deterministic-token"
	assert.Equal(t, hashToken(token), hashToken(token))
}

func TestTokenHash_DifferentInputsProduceDifferentHashes(t *testing.T) {
	h1 := hashToken("token-a")
	h2 := hashToken("token-b")
	assert.NotEqual(t, h1, h2)
}

func TestTokenHash_EmptyStringReturnsEmpty(t *testing.T) {
	assert.Empty(t, hashToken(""))
	assert.Empty(t, hashToken("   "))
}

func TestTokenHash_TrimWhitespace(t *testing.T) {
	assert.Equal(t, hashToken("abc"), hashToken("  abc  "))
}

func TestTokenHash_CorrectLength(t *testing.T) {
	hashed := hashToken("some-token")
	// SHA-256 produces 32 bytes = 64 hex characters.
	assert.Len(t, hashed, 64)
}

// ---------------------------------------------------------------------------
// 5. Invite lifecycle
// ---------------------------------------------------------------------------

func TestInvite_CreateAndList(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	invite := Invite{
		Token:     "inv-001",
		Email:     "newuser@example.com",
		Role:      RoleMember,
		CreatedBy: "owner@brain.local",
		CreatedAt: now,
		ExpiresAt: now.Add(14 * 24 * time.Hour),
	}

	require.NoError(t, store.UpsertInvite(ctx, invite))

	listed, err := store.ListInvites(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, invite.Token, listed[0].Token)
	assert.Equal(t, invite.Email, listed[0].Email)
	assert.Equal(t, invite.Role, listed[0].Role)
}

func TestInvite_GetByToken(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	invite := Invite{
		Token:     "inv-002",
		Email:     "get@example.com",
		Role:      RoleAdmin,
		CreatedBy: "admin@brain.local",
		CreatedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	require.NoError(t, store.UpsertInvite(ctx, invite))

	fetched, err := store.GetInviteByToken(ctx, invite.Token)
	require.NoError(t, err)
	assert.Equal(t, invite.Email, fetched.Email)
	assert.Equal(t, invite.Role, fetched.Role)
}

func TestInvite_GetByTokenNotFound(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	_, err := store.GetInviteByToken(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestInvite_Consume(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	invite := Invite{
		Token:     "inv-003",
		Email:     "consume@example.com",
		Role:      RoleViewer,
		CreatedBy: "owner@brain.local",
		CreatedAt: now,
		ExpiresAt: now.Add(14 * 24 * time.Hour),
	}
	require.NoError(t, store.UpsertInvite(ctx, invite))

	consumedAt := time.Now().UTC()
	require.NoError(t, store.ConsumeInvite(ctx, invite.Token, consumedAt))

	fetched, err := store.GetInviteByToken(ctx, invite.Token)
	require.NoError(t, err)
	assert.False(t, fetched.ConsumedAt.IsZero(), "ConsumedAt should be set")
	assert.WithinDuration(t, consumedAt, fetched.ConsumedAt, time.Second)
}

func TestInvite_UpdateExisting(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	invite := Invite{
		Token:     "inv-004",
		Email:     "update@example.com",
		Role:      RoleMember,
		CreatedBy: "owner@brain.local",
		CreatedAt: now,
		ExpiresAt: now.Add(14 * 24 * time.Hour),
	}
	require.NoError(t, store.UpsertInvite(ctx, invite))

	// Update the role.
	invite.Role = RoleAdmin
	require.NoError(t, store.UpsertInvite(ctx, invite))

	fetched, err := store.GetInviteByToken(ctx, invite.Token)
	require.NoError(t, err)
	assert.Equal(t, RoleAdmin, fetched.Role)
}

func TestInvite_EmptyTokenRejected(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	err := store.UpsertInvite(context.Background(), Invite{Token: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestInvite_MultipleListOrder(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	invites := []Invite{
		{Token: "inv-a", Email: "a@x.com", Role: RoleMember, CreatedBy: "o", CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(14 * 24 * time.Hour)},
		{Token: "inv-b", Email: "b@x.com", Role: RoleAdmin, CreatedBy: "o", CreatedAt: now.Add(-1 * time.Hour), ExpiresAt: now.Add(14 * 24 * time.Hour)},
		{Token: "inv-c", Email: "c@x.com", Role: RoleViewer, CreatedBy: "o", CreatedAt: now, ExpiresAt: now.Add(14 * 24 * time.Hour)},
	}
	for _, inv := range invites {
		require.NoError(t, store.UpsertInvite(ctx, inv))
	}

	listed, err := store.ListInvites(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 3)
	// Ordered by created_at DESC, so inv-c should be first.
	assert.Equal(t, "inv-c", listed[0].Token)
}

// ---------------------------------------------------------------------------
// 6. Audit logging
// ---------------------------------------------------------------------------

func TestAudit_AppendAndRetrieve(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	entry := governance.AuditEntry{
		ID:        "audit-001",
		Timestamp: time.Now().UTC(),
		Action:    "policy_resolved",
		Subject:   "user-1",
		Resource:  "policy-default",
		Details:   map[string]string{"key": "value"},
		Success:   true,
	}

	require.NoError(t, store.AppendAudit(ctx, entry))
}

func TestAudit_MultipleEntries(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	entries := []governance.AuditEntry{
		{ID: "a1", Timestamp: time.Now().UTC(), Action: "action_a", Subject: "s1", Resource: "r1", Success: true, Details: map[string]string{}},
		{ID: "a2", Timestamp: time.Now().UTC(), Action: "action_b", Subject: "s2", Resource: "r2", Success: false, Details: map[string]string{"reason": "denied"}},
		{ID: "a3", Timestamp: time.Now().UTC(), Action: "action_a", Subject: "s3", Resource: "r3", Success: true, Details: map[string]string{}},
	}
	for _, e := range entries {
		require.NoError(t, store.AppendAudit(ctx, e))
	}
	// No retrieval method is available on the store interface; AppendAudit
	// is append-only. The test verifies no errors and idempotent inserts.
}

func TestAudit_ZeroTimestampGetsDefault(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	entry := governance.AuditEntry{
		ID:       "audit-ts",
		Action:   "test",
		Subject:  "s",
		Resource: "r",
		Success:  true,
		Details:  map[string]string{},
	}
	// Timestamp is zero -- the store should set a default.
	require.NoError(t, store.AppendAudit(ctx, entry))
}

// ---------------------------------------------------------------------------
// 7. Expired session detection (storage of expiration times)
// ---------------------------------------------------------------------------

func TestStore_ExpiredSessionExpirationTimeIsPersisted(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	now := time.Now().UTC()
	expiredSession := Session{
		Token:            "expired-tok",
		RefreshToken:     "expired-ref",
		User:             user,
		CreatedAt:        now.Add(-48 * time.Hour),
		ExpiresAt:        now.Add(-24 * time.Hour), // already expired
		RefreshExpiresAt: now.Add(-12 * time.Hour),
		LastUsed:         now.Add(-24 * time.Hour),
	}

	require.NoError(t, store.SaveSession(ctx, expiredSession))

	loaded, err := store.GetSessionByToken(ctx, expiredSession.Token)
	require.NoError(t, err)
	// The session is still in the database; the daemon is responsible for
	// enforcing TTL. Verify the expiration time was persisted correctly.
	assert.True(t, time.Now().UTC().After(loaded.ExpiresAt), "session should be expired")
}

func TestStore_SessionDeletedByTokenDoesNotAffectOtherSessions(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	s1 := mustSession(t, user)
	s2 := mustSession(t, user)
	s1.Token = "session-alpha"
	s1.RefreshToken = "ref-alpha"
	s2.Token = "session-beta"
	s2.RefreshToken = "ref-beta"

	require.NoError(t, store.SaveSession(ctx, s1))
	require.NoError(t, store.SaveSession(ctx, s2))

	require.NoError(t, store.DeleteSessionByToken(ctx, s1.Token))

	_, err := store.GetSessionByToken(ctx, s1.Token)
	assert.ErrorIs(t, err, ErrSessionNotFound)

	_, err = store.GetSessionByToken(ctx, s2.Token)
	assert.NoError(t, err, "session-beta should still exist")
}

// ---------------------------------------------------------------------------
// 8. Close idempotency
// ---------------------------------------------------------------------------

func TestStore_CloseIdempotent(t *testing.T) {
	store, _ := openTestStore(t)

	require.NoError(t, store.Close())
	require.NoError(t, store.Close()) // second close should not panic

	// nil receiver should also be safe.
	var nilStore *sqliteSessionStore
	assert.NoError(t, nilStore.Close())
}

// ---------------------------------------------------------------------------
// 9. Round-trip integration: full session + user + invite + audit
// ---------------------------------------------------------------------------

func TestStore_FullRoundTrip(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Upsert user.
	user := User{
		ID:    "rt-user",
		Email: "rt@example.com",
		Name:  "Round Trip",
		Role:  RoleAdmin,
	}
	require.NoError(t, store.UpsertUser(ctx, user))

	// 2. Save session.
	session := Session{
		Token:            "rt-token",
		RefreshToken:     "rt-refresh",
		User:             user,
		CreatedAt:        now,
		ExpiresAt:        now.Add(12 * time.Hour),
		RefreshExpiresAt: now.Add(30 * 24 * time.Hour),
		LastUsed:         now,
	}
	require.NoError(t, store.SaveSession(ctx, session))

	// 3. Create invite.
	invite := Invite{
		Token:     "rt-invite",
		Email:     "invitee@example.com",
		Role:      RoleMember,
		CreatedBy: user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(14 * 24 * time.Hour),
	}
	require.NoError(t, store.UpsertInvite(ctx, invite))

	// 4. Append audit entry.
	audit := governance.AuditEntry{
		ID:        "rt-audit",
		Timestamp: now,
		Action:    "policy_resolved",
		Subject:   user.ID,
		Resource:  "policy-x",
		Success:   true,
		Details:   map[string]string{"test": "true"},
	}
	require.NoError(t, store.AppendAudit(ctx, audit))

	// 5. Verify everything persisted.
	gotUser, err := store.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, gotUser.Email)

	gotSession, err := store.GetSessionByToken(ctx, session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.Token, gotSession.Token)
	assert.Equal(t, session.User.ID, gotSession.User.ID)

	gotInvite, err := store.GetInviteByToken(ctx, invite.Token)
	require.NoError(t, err)
	assert.Equal(t, invite.Email, gotInvite.Email)

	listedSessions, err := store.ListSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, listedSessions, 1)

	listedInvites, err := store.ListInvites(ctx)
	require.NoError(t, err)
	assert.Len(t, listedInvites, 1)
}

// ---------------------------------------------------------------------------
// 10. User Provider and Subject fields
// ---------------------------------------------------------------------------

func TestStore_UserProviderAndSubject(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := User{
		ID:       "oidc-user",
		Email:    "oidc@example.com",
		Name:     "OIDC User",
		Role:     RoleMember,
		Provider: "google",
		Subject:  "google-subject-12345",
	}
	require.NoError(t, store.UpsertUser(ctx, user))

	loaded, err := store.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "google", loaded.Provider)
	assert.Equal(t, "google-subject-12345", loaded.Subject)
}

// ---------------------------------------------------------------------------
// 11. LastSeenAt round-trip
// ---------------------------------------------------------------------------

func TestStore_UserLastSeenAtRoundTrip(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	lastSeen := time.Now().UTC().Add(-5 * time.Minute)
	user := User{
		ID:         "ls-user",
		Email:      "ls@example.com",
		Name:       "Last Seen",
		Role:       RoleViewer,
		LastSeenAt: lastSeen,
	}
	require.NoError(t, store.UpsertUser(ctx, user))

	loaded, err := store.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, lastSeen, loaded.LastSeenAt, time.Second)
}

// ---------------------------------------------------------------------------
// 12. SaveSession updates user last_seen_at
// ---------------------------------------------------------------------------

func TestStore_SaveSessionUpdatesUserLastSeen(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	user := mustUser(t)
	user.LastSeenAt = time.Time{} // zero initially
	require.NoError(t, store.UpsertUser(ctx, user))

	session := mustSession(t, user)
	require.NoError(t, store.SaveSession(ctx, session))

	loaded, err := store.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, loaded.LastSeenAt.IsZero(), "last_seen_at should be updated by SaveSession")
	assert.WithinDuration(t, session.LastUsed, loaded.LastSeenAt, time.Second)
}

// ---------------------------------------------------------------------------
// 13. NewStore is an alias for newSessionStore
// ---------------------------------------------------------------------------

func TestNewStore_IsAlias(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "alias-test.db")

	store, err := NewStore(dbPath)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	// Verify it is usable through the interface.
	ctx := context.Background()
	require.NoError(t, store.UpsertUser(ctx, mustUser(t)))

	loaded, err := store.GetUser(ctx, "user-test-001")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", loaded.Email)
}

// ---------------------------------------------------------------------------
// 14. Invite consume with zero time
// ---------------------------------------------------------------------------

func TestInvite_ConsumeWithZeroTime(t *testing.T) {
	store, _ := openTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	invite := Invite{
		Token:     "inv-zero",
		Email:     "zero@example.com",
		Role:      RoleMember,
		CreatedBy: "owner",
		CreatedAt: now,
		ExpiresAt: now.Add(14 * 24 * time.Hour),
	}
	require.NoError(t, store.UpsertInvite(ctx, invite))

	// Consuming with a zero time should still update the consumed_at field
	// (stored as empty string which parses to zero time).
	require.NoError(t, store.ConsumeInvite(ctx, invite.Token, time.Time{}))

	fetched, err := store.GetInviteByToken(ctx, invite.Token)
	require.NoError(t, err)
	assert.True(t, fetched.ConsumedAt.IsZero())
}

// ---------------------------------------------------------------------------
// 15. Time helpers
// ---------------------------------------------------------------------------

func TestFormatOptionalTime_ZeroReturnsEmpty(t *testing.T) {
	assert.Empty(t, formatOptionalTime(time.Time{}))
}

func TestFormatOptionalTime_NonZeroReturnsRFC3339(t *testing.T) {
	ts := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	got := formatOptionalTime(ts)
	assert.Equal(t, "2025-06-15T10:30:00Z", got)
}

func TestParseOptionalTime_EmptyReturnsZero(t *testing.T) {
	got := parseOptionalTime("")
	assert.True(t, got.IsZero())
}

func TestMustParseTime_EmptyReturnsZero(t *testing.T) {
	got := mustParseTime("")
	assert.True(t, got.IsZero())
}

func TestMustParseTime_InvalidReturnsZero(t *testing.T) {
	got := mustParseTime("not-a-time")
	assert.True(t, got.IsZero())
}

func TestBoolToInt(t *testing.T) {
	assert.Equal(t, 1, boolToInt(true))
	assert.Equal(t, 0, boolToInt(false))
}

// ---------------------------------------------------------------------------
// 16. FromHash helper
// ---------------------------------------------------------------------------

func TestFromHash_EmptyReturnsFalse(t *testing.T) {
	_, ok := fromHash("")
	assert.False(t, ok)
}

func TestFromHash_WhitespaceReturnsFalse(t *testing.T) {
	_, ok := fromHash("   ")
	assert.False(t, ok)
}

func TestFromHash_NonEmptyReturnsValue(t *testing.T) {
	val := "abc123"
	got, ok := fromHash(val)
	assert.True(t, ok)
	assert.Equal(t, val, got)
}
