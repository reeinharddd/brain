package governance

import (
	"context"
	"testing"
)

func TestRBACDefineRole(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBAC()

	t.Run("define admin role", func(t *testing.T) {
		role := Role{
			ID:   RoleAdmin,
			Name: "Administrator",
			Permissions: []Permission{
				{Resource: "artifact", Action: "read"},
				{Resource: "artifact", Action: "write"},
				{Resource: "artifact", Action: "delete"},
				{Resource: "artifact", Action: "admin"},
				{Resource: "skill", Action: "read"},
				{Resource: "skill", Action: "write"},
				{Resource: "skill", Action: "execute"},
				{Resource: "skill", Action: "admin"},
				{Resource: "policy", Action: "read"},
				{Resource: "policy", Action: "write"},
				{Resource: "policy", Action: "admin"},
			},
		}

		if err := rbac.DefineRole(ctx, role); err != nil {
			t.Errorf("DefineRole() error = %v", err)
		}
	})

	t.Run("define developer role", func(t *testing.T) {
		role := Role{
			ID:   RoleDeveloper,
			Name: "Developer",
			Permissions: []Permission{
				{Resource: "artifact", Action: "read"},
				{Resource: "artifact", Action: "write"},
				{Resource: "skill", Action: "read"},
				{Resource: "skill", Action: "execute"},
			},
		}

		if err := rbac.DefineRole(ctx, role); err != nil {
			t.Errorf("DefineRole() error = %v", err)
		}
	})

	t.Run("define reviewer role with inheritance", func(t *testing.T) {
		role := Role{
			ID:           RoleReviewer,
			Name:         "Reviewer",
			InheritsFrom: RoleDeveloper,
			Permissions: []Permission{
				{Resource: "artifact", Action: "read"},
			},
		}

		if err := rbac.DefineRole(ctx, role); err != nil {
			t.Errorf("DefineRole() error = %v", err)
		}
	})

	t.Run("define viewer role", func(t *testing.T) {
		role := Role{
			ID:   RoleViewer,
			Name: "Viewer",
			Permissions: []Permission{
				{Resource: "artifact", Action: "read"},
				{Resource: "skill", Action: "read"},
			},
		}

		if err := rbac.DefineRole(ctx, role); err != nil {
			t.Errorf("DefineRole() error = %v", err)
		}
	})

	t.Run("duplicate role error", func(t *testing.T) {
		role := Role{
			ID:   RoleAdmin,
			Name: "Administrator",
			Permissions: []Permission{
				{Resource: "artifact", Action: "read"},
			},
		}

		if err := rbac.DefineRole(ctx, role); err == nil {
			t.Error("expected error for duplicate role, got nil")
		}
	})

	t.Run("empty role ID", func(t *testing.T) {
		role := Role{
			ID:   "",
			Name: "Empty",
		}

		if err := rbac.DefineRole(ctx, role); err == nil {
			t.Error("expected error for empty role ID, got nil")
		}
	})
}

func TestRBACGetRole(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBAC()

	role := Role{
		ID:   RoleAdmin,
		Name: "Administrator",
		Permissions: []Permission{
			{Resource: "artifact", Action: "read"},
		},
	}
	rbac.DefineRole(ctx, role)

	t.Run("get existing role", func(t *testing.T) {
		got, err := rbac.GetRole(ctx, RoleAdmin)
		if err != nil {
			t.Fatalf("GetRole() error = %v", err)
		}
		if got.ID != RoleAdmin {
			t.Errorf("expected ID %q, got %q", RoleAdmin, got.ID)
		}
		if got.Name != "Administrator" {
			t.Errorf("expected name Administrator, got %q", got.Name)
		}
	})

	t.Run("get non-existing role", func(t *testing.T) {
		_, err := rbac.GetRole(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for non-existing role, got nil")
		}
	})
}

func TestRBACHasPermission(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBAC()

	// Define roles
	adminRole := Role{
		ID:   RoleAdmin,
		Name: "Admin",
		Permissions: []Permission{
			{Resource: "artifact", Action: "admin"},
			{Resource: "artifact", Action: "read"},
			{Resource: "artifact", Action: "write"},
		},
	}
	rbac.DefineRole(ctx, adminRole)

	devRole := Role{
		ID:   RoleDeveloper,
		Name: "Developer",
		Permissions: []Permission{
			{Resource: "artifact", Action: "read"},
			{Resource: "artifact", Action: "write"},
		},
	}
	rbac.DefineRole(ctx, devRole)

	reviewerRole := Role{
		ID:           RoleReviewer,
		Name:         "Reviewer",
		InheritsFrom: RoleDeveloper,
		Permissions: []Permission{
			{Resource: "artifact", Action: "read"},
		},
	}
	rbac.DefineRole(ctx, reviewerRole)

	tests := []struct {
		name     string
		roleID   string
		resource string
		action   string
		want     bool
	}{
		{"admin has artifact admin", RoleAdmin, "artifact", "admin", true},
		{"admin has artifact read", RoleAdmin, "artifact", "read", true},
		{"admin has artifact write", RoleAdmin, "artifact", "write", true},
		{"developer has artifact read", RoleDeveloper, "artifact", "read", true},
		{"developer has artifact write", RoleDeveloper, "artifact", "write", true},
		{"developer does not have artifact admin", RoleDeveloper, "artifact", "admin", false},
		{"reviewer has inherited read", RoleReviewer, "artifact", "read", true},
		{"reviewer has inherited write", RoleReviewer, "artifact", "write", true},
		{"reviewer does not have admin", RoleReviewer, "artifact", "admin", false},
		{"non-existent role has no permissions", "nonexistent", "artifact", "read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rbac.HasPermission(ctx, tt.roleID, tt.resource, tt.action)
			if got != tt.want {
				t.Errorf("HasPermission(%q, %q, %q) = %v, want %v", tt.roleID, tt.resource, tt.action, got, tt.want)
			}
		})
	}
}

func TestRBACGetEffectivePermissions(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBAC()

	// Define parent role
	parentRole := Role{
		ID:   "parent",
		Name: "Parent",
		Permissions: []Permission{
			{Resource: "artifact", Action: "read"},
			{Resource: "artifact", Action: "write"},
		},
	}
	rbac.DefineRole(ctx, parentRole)

	// Define child role that inherits from parent
	childRole := Role{
		ID:           "child",
		Name:         "Child",
		InheritsFrom: "parent",
		Permissions: []Permission{
			{Resource: "skill", Action: "read"},
			{Resource: "artifact", Action: "read"}, // duplicate with parent
		},
	}
	rbac.DefineRole(ctx, childRole)

	t.Run("effective permissions include inherited", func(t *testing.T) {
		perms := rbac.GetEffectivePermissions(ctx, "child")

		// Should have: artifact:read, artifact:write (inherited), skill:read
		// artifact:read appears in both but should only appear once
		foundArtifactRead := false
		foundArtifactWrite := false
		foundSkillRead := false

		for _, p := range perms {
			if p.Resource == "artifact" && p.Action == "read" {
				foundArtifactRead = true
			}
			if p.Resource == "artifact" && p.Action == "write" {
				foundArtifactWrite = true
			}
			if p.Resource == "skill" && p.Action == "read" {
				foundSkillRead = true
			}
		}

		if !foundArtifactRead {
			t.Error("expected artifact:read permission")
		}
		if !foundArtifactWrite {
			t.Error("expected artifact:write permission (inherited)")
		}
		if !foundSkillRead {
			t.Error("expected skill:read permission")
		}
	})

	t.Run("direct permissions only", func(t *testing.T) {
		perms := rbac.GetEffectivePermissions(ctx, "parent")
		if len(perms) != 2 {
			t.Errorf("expected 2 permissions, got %d", len(perms))
		}
	})
}

func TestRBACListRoles(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBAC()

	rbac.DefineRole(ctx, Role{
		ID:   RoleAdmin,
		Name: "Admin",
		Permissions: []Permission{
			{Resource: "artifact", Action: "admin"},
		},
	})
	rbac.DefineRole(ctx, Role{
		ID:   RoleViewer,
		Name: "Viewer",
		Permissions: []Permission{
			{Resource: "artifact", Action: "read"},
		},
	})

	roles := rbac.ListRoles(ctx)
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestRBACCircularInheritance(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBAC()

	// Create circular inheritance (should not cause infinite loop)
	rbac.DefineRole(ctx, Role{
		ID:           "role-a",
		Name:         "Role A",
		InheritsFrom: "role-b",
		Permissions: []Permission{
			{Resource: "artifact", Action: "read"},
		},
	})
	rbac.DefineRole(ctx, Role{
		ID:           "role-b",
		Name:         "Role B",
		InheritsFrom: "role-a",
		Permissions: []Permission{
			{Resource: "artifact", Action: "write"},
		},
	})

	// Should not hang
	hasPerm := rbac.HasPermission(ctx, "role-a", "artifact", "read")
	if !hasPerm {
		t.Error("expected role-a to have artifact:read")
	}

	perms := rbac.GetEffectivePermissions(ctx, "role-a")
	if len(perms) == 0 {
		t.Error("expected some permissions")
	}
}
