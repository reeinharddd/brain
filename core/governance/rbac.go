package governance

import (
	"context"
	"fmt"
	"sync"
)

// Role defines an access role.
type Role struct {
	ID           string
	Name         string
	Permissions  []Permission
	InheritsFrom string // parent role ID (optional)
}

// Permission defines an allowed action on a resource.
type Permission struct {
	Resource string // "artifact", "skill", "mcp", "agent", "policy"
	Action   string // "read", "write", "execute", "delete", "admin"
	Scope    string // optional scope restriction
}

// RBAC manages role definitions.
type RBAC struct {
	mu    sync.RWMutex
	roles map[string]*Role // role ID -> role
}

// Common roles.
const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleReviewer  = "reviewer"
	RoleViewer    = "viewer"
)

// NewRBAC creates a new RBAC instance.
func NewRBAC() *RBAC {
	return &RBAC{
		roles: make(map[string]*Role),
	}
}

// DefineRole defines or updates a role.
func (r *RBAC) DefineRole(ctx context.Context, role Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if role.ID == "" {
		return fmt.Errorf("role ID cannot be empty")
	}

	if _, exists := r.roles[role.ID]; exists {
		return fmt.Errorf("role %q already exists", role.ID)
	}

	// Store a copy
	roleCopy := Role{
		ID:           role.ID,
		Name:         role.Name,
		Permissions:  make([]Permission, len(role.Permissions)),
		InheritsFrom: role.InheritsFrom,
	}
	copy(roleCopy.Permissions, role.Permissions)

	r.roles[role.ID] = &roleCopy
	return nil
}

// GetRole returns a role by ID.
func (r *RBAC) GetRole(ctx context.Context, roleID string) (*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, exists := r.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role %q not found", roleID)
	}

	// Return a copy
	roleCopy := *role
	roleCopy.Permissions = make([]Permission, len(role.Permissions))
	copy(roleCopy.Permissions, role.Permissions)
	return &roleCopy, nil
}

// HasPermission checks if a role has a specific permission (direct or inherited).
func (r *RBAC) HasPermission(ctx context.Context, roleID, resource, action string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visited := make(map[string]bool)
	return r.hasPermissionRecursive(roleID, resource, action, visited)
}

func (r *RBAC) hasPermissionRecursive(roleID, resource, action string, visited map[string]bool) bool {
	if visited[roleID] {
		return false // prevent infinite loops
	}
	visited[roleID] = true

	role, exists := r.roles[roleID]
	if !exists {
		return false
	}

	// Check direct permissions
	for _, perm := range role.Permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}

	// Check inherited permissions
	if role.InheritsFrom != "" {
		return r.hasPermissionRecursive(role.InheritsFrom, resource, action, visited)
	}

	return false
}

// GetEffectivePermissions returns all permissions a role has, including inherited ones.
func (r *RBAC) GetEffectivePermissions(ctx context.Context, roleID string) []Permission {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visited := make(map[string]bool)
	permSet := make(map[string]Permission) // key: resource:action:scope

	r.collectPermissions(roleID, permSet, visited)

	result := make([]Permission, 0, len(permSet))
	for _, p := range permSet {
		result = append(result, p)
	}
	return result
}

func (r *RBAC) collectPermissions(roleID string, permSet map[string]Permission, visited map[string]bool) {
	if visited[roleID] {
		return
	}
	visited[roleID] = true

	role, exists := r.roles[roleID]
	if !exists {
		return
	}

	// Add direct permissions
	for _, perm := range role.Permissions {
		key := perm.Resource + ":" + perm.Action + ":" + perm.Scope
		if _, exists := permSet[key]; !exists {
			permSet[key] = perm
		}
	}

	// Collect inherited permissions
	if role.InheritsFrom != "" {
		r.collectPermissions(role.InheritsFrom, permSet, visited)
	}
}

// ListRoles returns all defined roles.
func (r *RBAC) ListRoles(ctx context.Context) []Role {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Role, 0, len(r.roles))
	for _, role := range r.roles {
		roleCopy := *role
		roleCopy.Permissions = make([]Permission, len(role.Permissions))
		copy(roleCopy.Permissions, role.Permissions)
		result = append(result, roleCopy)
	}
	return result
}
