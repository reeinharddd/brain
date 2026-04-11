package governance

import (
	"context"
	"testing"
)

func TestNewPolicyRule(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		pclass  string
		pscope  string
		value   string
		wantErr bool
	}{
		{"hard policy", "rule-1", "hard", "platform", "deny-all", false},
		{"soft policy", "rule-2", "soft", "org", "allow-dev-tools", false},
		{"guarded policy", "rule-3", "guarded", "team", "require-review", false},
		{"empty ID", "", "hard", "platform", "deny-all", true},
		{"empty name", "rule-4", "hard", "platform", "", true},
		{"empty value", "rule-5", "hard", "platform", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc, err := ValidatePolicyClass(tt.pclass)
			if err != nil && !tt.wantErr {
				// For test cases that only test value errors, skip if class is invalid
			}
			ps, _ := ValidatePolicyScope(tt.pscope)

			rule, err := NewPolicyRule(tt.id, tt.name, pc, ps, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && rule == nil {
				t.Error("expected non-nil rule")
			}
			if !tt.wantErr && rule.ID != tt.id {
				t.Errorf("expected ID %q, got %q", tt.id, rule.ID)
			}
		})
	}
}

func TestNewPolicySet(t *testing.T) {
	tests := []struct {
		name    string
		scope   PolicyScope
		scopeID string
		wantErr bool
	}{
		{"valid org set", ScopeOrg, "org:acme", false},
		{"valid user set", ScopeUser, "user:alice", false},
		{"empty scope ID", ScopePlatform, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps, err := NewPolicySet(tt.scope, tt.scopeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPolicySet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ps.Scope != tt.scope {
					t.Errorf("expected scope %q, got %q", tt.scope, ps.Scope)
				}
				if ps.ScopeID != tt.scopeID {
					t.Errorf("expected scopeID %q, got %q", tt.scopeID, ps.ScopeID)
				}
				if ps.Rules == nil {
					t.Error("expected non-nil Rules map")
				}
			}
		})
	}
}

func TestPolicySetAddRemoveRule(t *testing.T) {
	ctx := context.Background()
	ps, err := NewPolicySet(ScopeOrg, "org:acme")
	if err != nil {
		t.Fatalf("NewPolicySet() error = %v", err)
	}

	t.Run("add rule", func(t *testing.T) {
		rule, err := NewPolicyRule("rule-1", "test-rule", PolicyHard, ScopeOrg, "deny-all")
		if err != nil {
			t.Fatalf("NewPolicyRule() error = %v", err)
		}

		if err := ps.AddRule(ctx, rule); err != nil {
			t.Errorf("AddRule() error = %v", err)
		}
		if len(ps.Rules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(ps.Rules))
		}
	})

	t.Run("add duplicate rule", func(t *testing.T) {
		rule, _ := NewPolicyRule("rule-1", "test-rule", PolicyHard, ScopeOrg, "deny-all")
		if err := ps.AddRule(ctx, rule); err == nil {
			t.Error("expected error for duplicate rule, got nil")
		}
	})

	t.Run("add nil rule", func(t *testing.T) {
		if err := ps.AddRule(ctx, nil); err == nil {
			t.Error("expected error for nil rule, got nil")
		}
	})

	t.Run("remove existing rule", func(t *testing.T) {
		if err := ps.RemoveRule(ctx, "rule-1"); err != nil {
			t.Errorf("RemoveRule() error = %v", err)
		}
		if len(ps.Rules) != 0 {
			t.Errorf("expected 0 rules, got %d", len(ps.Rules))
		}
	})

	t.Run("remove non-existing rule", func(t *testing.T) {
		if err := ps.RemoveRule(ctx, "nonexistent"); err == nil {
			t.Error("expected error for non-existing rule, got nil")
		}
	})
}

func TestPolicySetGetRule(t *testing.T) {
	ctx := context.Background()
	ps, _ := NewPolicySet(ScopeOrg, "org:acme")

	rule, _ := NewPolicyRule("rule-1", "test-rule", PolicySoft, ScopeOrg, "allow")
	ps.AddRule(ctx, rule)

	t.Run("get existing rule", func(t *testing.T) {
		got, err := ps.GetRule(ctx, "rule-1")
		if err != nil {
			t.Errorf("GetRule() error = %v", err)
		}
		if got.ID != "rule-1" {
			t.Errorf("expected ID rule-1, got %q", got.ID)
		}
	})

	t.Run("get non-existing rule", func(t *testing.T) {
		_, err := ps.GetRule(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for non-existing rule, got nil")
		}
	})
}

func TestValidatePolicyClass(t *testing.T) {
	tests := []struct {
		input   string
		want    PolicyClass
		wantErr bool
	}{
		{"hard", PolicyHard, false},
		{"soft", PolicySoft, false},
		{"guarded", PolicyGuarded, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ValidatePolicyClass(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePolicyClass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestValidatePolicyScope(t *testing.T) {
	tests := []struct {
		input   string
		want    PolicyScope
		wantErr bool
	}{
		{"platform", ScopePlatform, false},
		{"organization", ScopeOrg, false},
		{"team", ScopeTeam, false},
		{"user", ScopeUser, false},
		{"workspace", ScopeWorkspace, false},
		{"project", ScopeProject, false},
		{"session", ScopeSession, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ValidatePolicyScope(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePolicyScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestScopeHierarchy(t *testing.T) {
	hierarchy := ScopeHierarchy()
	if len(hierarchy) != 7 {
		t.Errorf("expected 7 scopes, got %d", len(hierarchy))
	}
	if hierarchy[0] != ScopePlatform {
		t.Errorf("expected first scope to be platform, got %q", hierarchy[0])
	}
	if hierarchy[6] != ScopeSession {
		t.Errorf("expected last scope to be session, got %q", hierarchy[6])
	}
}

func TestScopeIndex(t *testing.T) {
	tests := []struct {
		scope PolicyScope
		want  int
	}{
		{ScopePlatform, 0},
		{ScopeOrg, 1},
		{ScopeTeam, 2},
		{ScopeUser, 3},
		{ScopeWorkspace, 4},
		{ScopeProject, 5},
		{ScopeSession, 6},
		{PolicyScope("invalid"), -1},
	}

	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			got := ScopeIndex(tt.scope)
			if got != tt.want {
				t.Errorf("ScopeIndex(%q) = %d, want %d", tt.scope, got, tt.want)
			}
		})
	}
}

func TestIsMoreSpecific(t *testing.T) {
	tests := []struct {
		a    PolicyScope
		b    PolicyScope
		want bool
	}{
		{ScopeOrg, ScopePlatform, true},
		{ScopeSession, ScopeProject, true},
		{ScopePlatform, ScopeSession, false},
		{ScopeTeam, ScopeTeam, false},
		{ScopeUser, ScopeOrg, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.a)+"_vs_"+string(tt.b), func(t *testing.T) {
			got := IsMoreSpecific(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("IsMoreSpecific(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
