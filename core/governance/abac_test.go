package governance

import (
	"context"
	"testing"
)

func TestABACAddRemoveRule(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	t.Run("add rule", func(t *testing.T) {
		rule := ABACRule{
			ID:     "rule-1",
			Name:   "deny-external",
			Effect: "deny",
			Priority: 10,
			Conditions: []ABACCondition{
				{Attribute: "user.department", Operator: "eq", Value: "external"},
			},
		}

		if err := abac.AddRule(ctx, rule); err != nil {
			t.Errorf("AddRule() error = %v", err)
		}
	})

	t.Run("add duplicate rule", func(t *testing.T) {
		rule := ABACRule{
			ID:     "rule-1",
			Name:   "duplicate",
			Effect: "deny",
		}

		if err := abac.AddRule(ctx, rule); err == nil {
			t.Error("expected error for duplicate rule, got nil")
		}
	})

	t.Run("add rule with invalid effect", func(t *testing.T) {
		rule := ABACRule{
			ID:     "rule-invalid",
			Name:   "invalid",
			Effect: "invalid",
		}

		if err := abac.AddRule(ctx, rule); err == nil {
			t.Error("expected error for invalid effect, got nil")
		}
	})

	t.Run("add rule with empty ID", func(t *testing.T) {
		rule := ABACRule{
			ID:     "",
			Name:   "empty-id",
			Effect: "allow",
		}

		if err := abac.AddRule(ctx, rule); err == nil {
			t.Error("expected error for empty ID, got nil")
		}
	})

	t.Run("remove existing rule", func(t *testing.T) {
		if err := abac.RemoveRule(ctx, "rule-1"); err != nil {
			t.Errorf("RemoveRule() error = %v", err)
		}

		rules := abac.GetRules(ctx)
		for _, r := range rules {
			if r.ID == "rule-1" {
				t.Error("rule-1 should have been removed")
			}
		}
	})

	t.Run("remove non-existing rule", func(t *testing.T) {
		if err := abac.RemoveRule(ctx, "nonexistent"); err == nil {
			t.Error("expected error for non-existing rule, got nil")
		}
	})
}

func TestABACEvaluateConditions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		conditions []ABACCondition
		attributes map[string]interface{}
		wantMatch  bool
	}{
		{
			name: "eq match",
			conditions: []ABACCondition{
				{Attribute: "user.role", Operator: "eq", Value: "admin"},
			},
			attributes: map[string]interface{}{
				"user.role": "admin",
			},
			wantMatch: true,
		},
		{
			name: "eq no match",
			conditions: []ABACCondition{
				{Attribute: "user.role", Operator: "eq", Value: "admin"},
			},
			attributes: map[string]interface{}{
				"user.role": "viewer",
			},
			wantMatch: false,
		},
		{
			name: "neq match",
			conditions: []ABACCondition{
				{Attribute: "user.status", Operator: "neq", Value: "banned"},
			},
			attributes: map[string]interface{}{
				"user.status": "active",
			},
			wantMatch: true,
		},
		{
			name: "neq no match",
			conditions: []ABACCondition{
				{Attribute: "user.status", Operator: "neq", Value: "banned"},
			},
			attributes: map[string]interface{}{
				"user.status": "banned",
			},
			wantMatch: false,
		},
		{
			name: "in match",
			conditions: []ABACCondition{
				{Attribute: "artifact.kind", Operator: "in", Value: []string{"doc", "image", "code"}},
			},
			attributes: map[string]interface{}{
				"artifact.kind": "code",
			},
			wantMatch: true,
		},
		{
			name: "in no match",
			conditions: []ABACCondition{
				{Attribute: "artifact.kind", Operator: "in", Value: []string{"doc", "image"}},
			},
			attributes: map[string]interface{}{
				"artifact.kind": "video",
			},
			wantMatch: false,
		},
		{
			name: "contains match",
			conditions: []ABACCondition{
				{Attribute: "user.tags", Operator: "contains", Value: "trusted"},
			},
			attributes: map[string]interface{}{
				"user.tags": []string{"trusted", "active"},
			},
			wantMatch: true,
		},
		{
			name: "contains no match",
			conditions: []ABACCondition{
				{Attribute: "user.tags", Operator: "contains", Value: "trusted"},
			},
			attributes: map[string]interface{}{
				"user.tags": []string{"new", "unverified"},
			},
			wantMatch: false,
		},
		{
			name: "gt match",
			conditions: []ABACCondition{
				{Attribute: "artifact.size", Operator: "gt", Value: 100},
			},
			attributes: map[string]interface{}{
				"artifact.size": 200,
			},
			wantMatch: true,
		},
		{
			name: "gt no match",
			conditions: []ABACCondition{
				{Attribute: "artifact.size", Operator: "gt", Value: 100},
			},
			attributes: map[string]interface{}{
				"artifact.size": 50,
			},
			wantMatch: false,
		},
		{
			name: "lt match",
			conditions: []ABACCondition{
				{Attribute: "artifact.size", Operator: "lt", Value: 100},
			},
			attributes: map[string]interface{}{
				"artifact.size": 50,
			},
			wantMatch: true,
		},
		{
			name: "lt no match",
			conditions: []ABACCondition{
				{Attribute: "artifact.size", Operator: "lt", Value: 100},
			},
			attributes: map[string]interface{}{
				"artifact.size": 200,
			},
			wantMatch: false,
		},
		{
			name: "missing attribute",
			conditions: []ABACCondition{
				{Attribute: "user.department", Operator: "eq", Value: "engineering"},
			},
			attributes: map[string]interface{}{
				"user.role": "admin",
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abac := NewABAC()
			rule := ABACRule{
				ID:         "test-rule",
				Name:       "Test Rule",
				Effect:     "deny",
				Priority:   10,
				Conditions: tt.conditions,
			}
			abac.AddRule(ctx, rule)

			denyRules := abac.Evaluate(ctx, tt.attributes)
			matched := len(denyRules) > 0

			if matched != tt.wantMatch {
				t.Errorf("evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestABACDenyOnMatchingRule(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	denyRule := ABACRule{
		ID:     "deny-external",
		Name:   "Deny External Users",
		Effect: "deny",
		Priority: 10,
		Conditions: []ABACCondition{
			{Attribute: "user.type", Operator: "eq", Value: "external"},
		},
	}
	abac.AddRule(ctx, denyRule)

	t.Run("deny rule matches", func(t *testing.T) {
		attrs := map[string]interface{}{
			"user.type": "external",
		}
		allowed, reason := abac.IsAllowed(ctx, attrs)
		if allowed {
			t.Error("expected access to be denied")
		}
		if reason != "Deny External Users" {
			t.Errorf("expected reason 'Deny External Users', got %q", reason)
		}
	})

	t.Run("deny rule does not match", func(t *testing.T) {
		attrs := map[string]interface{}{
			"user.type": "internal",
		}
		allowed, reason := abac.IsAllowed(ctx, attrs)
		if !allowed {
			t.Errorf("expected access to be allowed, got reason: %q", reason)
		}
	})
}

func TestABACAllowWhenNoDenyMatches(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	// Only allow rules, no deny rules
	allowRule := ABACRule{
		ID:     "allow-internal",
		Name:   "Allow Internal",
		Effect: "allow",
		Priority: 10,
		Conditions: []ABACCondition{
			{Attribute: "user.type", Operator: "eq", Value: "internal"},
		},
	}
	abac.AddRule(ctx, allowRule)

	attrs := map[string]interface{}{
		"user.type": "internal",
	}
	allowed, reason := abac.IsAllowed(ctx, attrs)
	if !allowed {
		t.Errorf("expected access to be allowed, got reason: %q", reason)
	}
}

func TestABACPriorityOrdering(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	// Add rules with different priorities
	rule1 := ABACRule{
		ID:       "low-priority",
		Name:     "Low Priority Deny",
		Effect:   "deny",
		Priority: 5,
		Conditions: []ABACCondition{
			{Attribute: "user.type", Operator: "eq", Value: "external"},
		},
	}
	rule2 := ABACRule{
		ID:       "high-priority",
		Name:     "High Priority Deny",
		Effect:   "deny",
		Priority: 100,
		Conditions: []ABACCondition{
			{Attribute: "user.type", Operator: "eq", Value: "external"},
		},
	}

	abac.AddRule(ctx, rule1)
	abac.AddRule(ctx, rule2)

	attrs := map[string]interface{}{
		"user.type": "external",
	}

	denyRules := abac.Evaluate(ctx, attrs)
	if len(denyRules) != 2 {
		t.Fatalf("expected 2 deny rules, got %d", len(denyRules))
	}

	// First rule should be the highest priority
	if denyRules[0].Priority != 100 {
		t.Errorf("expected first rule priority 100, got %d", denyRules[0].Priority)
	}
	if denyRules[1].Priority != 5 {
		t.Errorf("expected second rule priority 5, got %d", denyRules[1].Priority)
	}
}

func TestABACMultipleConditionsAND(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	// Rule with multiple conditions (AND logic)
	rule := ABACRule{
		ID:     "deny-external-non-trusted",
		Name:   "Deny External Non-Trusted",
		Effect: "deny",
		Priority: 10,
		Conditions: []ABACCondition{
			{Attribute: "user.type", Operator: "eq", Value: "external"},
			{Attribute: "user.trusted", Operator: "eq", Value: false},
		},
	}
	abac.AddRule(ctx, rule)

	tests := []struct {
		name       string
		attributes map[string]interface{}
		wantAllow bool
	}{
		{
			name: "both conditions match - deny",
			attributes: map[string]interface{}{
				"user.type":    "external",
				"user.trusted": false,
			},
			wantAllow: false,
		},
		{
			name: "only first condition matches - allow",
			attributes: map[string]interface{}{
				"user.type":    "external",
				"user.trusted": true,
			},
			wantAllow: true,
		},
		{
			name: "only second condition matches - allow",
			attributes: map[string]interface{}{
				"user.type":    "internal",
				"user.trusted": false,
			},
			wantAllow: true,
		},
		{
			name: "neither condition matches - allow",
			attributes: map[string]interface{}{
				"user.type":    "internal",
				"user.trusted": true,
			},
			wantAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _ := abac.IsAllowed(ctx, tt.attributes)
			if allowed != tt.wantAllow {
				t.Errorf("IsAllowed() = %v, want %v", allowed, tt.wantAllow)
			}
		})
	}
}

func TestABACEmptyAttributes(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	rule := ABACRule{
		ID:     "deny-all",
		Name:   "Deny All",
		Effect: "deny",
		Priority: 10,
		Conditions: []ABACCondition{
			{Attribute: "user.type", Operator: "eq", Value: "external"},
		},
	}
	abac.AddRule(ctx, rule)

	// Empty attributes should not match any condition
	allowed, reason := abac.IsAllowed(ctx, map[string]interface{}{})
	if !allowed {
		t.Errorf("expected access to be allowed with empty attributes, got reason: %q", reason)
	}
}

func TestABACGetRules(t *testing.T) {
	ctx := context.Background()
	abac := NewABAC()

	abac.AddRule(ctx, ABACRule{ID: "rule-1", Name: "Rule 1", Effect: "deny", Priority: 10})
	abac.AddRule(ctx, ABACRule{ID: "rule-2", Name: "Rule 2", Effect: "allow", Priority: 5})

	rules := abac.GetRules(ctx)
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}
