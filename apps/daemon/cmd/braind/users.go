package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	coreidentity "github.com/reeinharrrd/brain/core/identity"
)

type userRoleUpdateRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type inviteCreateRequest struct {
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedBy string `json:"created_by"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type inviteCreateResponse struct {
	Success   bool              `json:"success"`
	Invite    coreidentity.Invite `json:"invite"`
	InviteURL string            `json:"invite_url,omitempty"`
}

type usersListResponse struct {
	Success bool                `json:"success"`
	Users   []coreidentity.User `json:"users"`
}

type invitesListResponse struct {
	Success bool                `json:"success"`
	Invites []coreidentity.Invite `json:"invites"`
}

func (d *BrainDaemon) handleUsersList(w http.ResponseWriter, r *http.Request) {
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
	if !d.authorizeCapability(w, r, coreidentity.CapabilityAuthManage) {
		return
	}

	users, err := d.auth.ListUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list users"})
		return
	}
	_ = json.NewEncoder(w).Encode(usersListResponse{Success: true, Users: users})
}

func (d *BrainDaemon) handleUserRoleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if d == nil || d.auth == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication service unavailable"})
		return
	}
	if !d.authorizeCapability(w, r, coreidentity.CapabilityAuthManage) {
		return
	}

	var req userRoleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		userID = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/users/"), "/role"))
	}
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing user id"})
		return
	}

	role := parseAuthRole(req.Role)
	if role == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid role, must be one of: owner, admin, member, viewer",
		})
		return
	}

	user, err := d.auth.UpdateUserRole(r.Context(), userID, role)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to update user role"})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"user":    user,
	})
}

func (d *BrainDaemon) handleInvitesList(w http.ResponseWriter, r *http.Request) {
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
	if !d.authorizeCapability(w, r, coreidentity.CapabilityAuthManage) {
		return
	}

	invites, err := d.auth.ListInvites(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list invites"})
		return
	}
	_ = json.NewEncoder(w).Encode(invitesListResponse{Success: true, Invites: invites})
}

func (d *BrainDaemon) handleInviteCreate(w http.ResponseWriter, r *http.Request) {
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
	if !d.authorizeCapability(w, r, coreidentity.CapabilityAuthManage) {
		return
	}

	var req inviteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "email is required"})
		return
	}
	if !isValidEmail(email) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid email format"})
		return
	}

	role := parseAuthRole(req.Role)
	if role == "" {
		role = coreidentity.RoleMember
	}

	invite := coreidentity.Invite{
		Email:     email,
		Role:      role,
		CreatedBy: strings.TrimSpace(req.CreatedBy),
		CreatedAt: time.Now().UTC(),
	}
	if invite.CreatedBy == "" {
		if status := d.authStatusForRequest(r); status.User != nil {
			invite.CreatedBy = status.User.Email
		}
	}
	if req.ExpiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.ExpiresAt); err == nil {
			invite.ExpiresAt = parsed
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid expires_at format, use RFC3339",
			})
			return
		}
	}
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = invite.CreatedAt.Add(defaultInviteTTL())
	}

	stored, err := d.auth.CreateInvite(r.Context(), invite)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create invite"})
		return
	}

	_ = json.NewEncoder(w).Encode(inviteCreateResponse{
		Success:   true,
		Invite:    stored,
		InviteURL: inviteURLForRequest(r, stored.Token),
	})
}

func inviteURLForRequest(r *http.Request, token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	baseURL := strings.TrimSpace(os.Getenv("BRAIN_PUBLIC_URL"))
	if baseURL == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		host := r.Host
		if host == "" {
			host = "127.0.0.1:9090"
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, host)
	}
	parsed := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf("%s/join?invite=%s", parsed, token)
}

func (d *BrainDaemon) handleInviteConsume(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invite token is required"})
		return
	}

	if err := d.auth.ConsumeInvite(r.Context(), req.Token); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired invite token"})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func defaultInviteTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("BRAIN_AUTH_INVITE_TTL_DAYS"))
	if raw == "" {
		return 14 * 24 * time.Hour
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 {
		return 14 * 24 * time.Hour
	}
	return time.Duration(days) * 24 * time.Hour
}

func isValidEmailAddr(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
