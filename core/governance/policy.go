package governance

import (
	"context"
	"fmt"
	"time"
)

// PolicyClass defines enforcement level.
type PolicyClass string

const (
	PolicyHard   PolicyClass = "hard"    // Cannot be overridden
	PolicySoft   PolicyClass = "soft"    // Can be overridden with approval
	PolicyGuarded PolicyClass = "guarded" // Requires approval to override
)

// PolicyScope defines hierarchy level.
type PolicyScope string

const (
	ScopePlatform  PolicyScope = "platform"
	ScopeOrg       PolicyScope = "organization"
	ScopeTeam      PolicyScope = "team"
	ScopeUser      PolicyScope = "user"
	ScopeWorkspace PolicyScope = "workspace"
	ScopeProject   PolicyScope = "project"
	ScopeSession   PolicyScope = "session"
)

// PolicyRule represents a single policy.
type PolicyRule struct {
	ID                     string
	Name                   string
	Class                  PolicyClass
	Scope                  PolicyScope
	Value                  string
	Description            string
	OverrideAllowed        bool
	OverrideRequiresApproval bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// PolicySet is a collection of rules at a scope.
type PolicySet struct {
	Scope   PolicyScope
	ScopeID string                   // e.g., "org:myorg", "user:alice"
	Rules   map[string]*PolicyRule   // rule ID -> rule
}

// ValidatePolicyClass checks if the given string is a valid PolicyClass.
func ValidatePolicyClass(s string) (PolicyClass, error) {
	switch PolicyClass(s) {
	case PolicyHard, PolicySoft, PolicyGuarded:
		return PolicyClass(s), nil
	default:
		return "", fmt.Errorf("invalid policy class: %q", s)
	}
}

// ValidatePolicyScope checks if the given string is a valid PolicyScope.
func ValidatePolicyScope(s string) (PolicyScope, error) {
	switch PolicyScope(s) {
	case ScopePlatform, ScopeOrg, ScopeTeam, ScopeUser, ScopeWorkspace, ScopeProject, ScopeSession:
		return PolicyScope(s), nil
	default:
		return "", fmt.Errorf("invalid policy scope: %q", s)
	}
}

// NewPolicyRule creates a new PolicyRule with timestamps set.
func NewPolicyRule(id, name string, class PolicyClass, scope PolicyScope, value string) (*PolicyRule, error) {
	if id == "" {
		return nil, fmt.Errorf("policy rule ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("policy rule name cannot be empty")
	}
	if value == "" {
		return nil, fmt.Errorf("policy rule value cannot be empty")
	}

	now := time.Now().UTC()
	return &PolicyRule{
		ID:        id,
		Name:      name,
		Class:     class,
		Scope:     scope,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewPolicySet creates a new empty PolicySet.
func NewPolicySet(scope PolicyScope, scopeID string) (*PolicySet, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("scope ID cannot be empty")
	}
	return &PolicySet{
		Scope:   scope,
		ScopeID: scopeID,
		Rules:   make(map[string]*PolicyRule),
	}, nil
}

// AddRule adds a rule to the policy set.
func (ps *PolicySet) AddRule(ctx context.Context, rule *PolicyRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}
	if _, exists := ps.Rules[rule.ID]; exists {
		return fmt.Errorf("rule %q already exists in policy set %s:%s", rule.ID, ps.Scope, ps.ScopeID)
	}
	ps.Rules[rule.ID] = rule
	return nil
}

// RemoveRule removes a rule from the policy set.
func (ps *PolicySet) RemoveRule(ctx context.Context, ruleID string) error {
	if _, exists := ps.Rules[ruleID]; !exists {
		return fmt.Errorf("rule %q not found in policy set %s:%s", ruleID, ps.Scope, ps.ScopeID)
	}
	delete(ps.Rules, ruleID)
	return nil
}

// GetRule returns a rule by ID, or nil if not found.
func (ps *PolicySet) GetRule(ctx context.Context, ruleID string) (*PolicyRule, error) {
	rule, exists := ps.Rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %q not found in policy set %s:%s", ruleID, ps.Scope, ps.ScopeID)
	}
	return rule, nil
}

// ScopeHierarchy returns the ordered scope chain from most general to most specific.
func ScopeHierarchy() []PolicyScope {
	return []PolicyScope{
		ScopePlatform,
		ScopeOrg,
		ScopeTeam,
		ScopeUser,
		ScopeWorkspace,
		ScopeProject,
		ScopeSession,
	}
}

// ScopeIndex returns the position of a scope in the hierarchy (lower = more general).
func ScopeIndex(scope PolicyScope) int {
	for i, s := range ScopeHierarchy() {
		if s == scope {
			return i
		}
	}
	return -1
}

// IsMoreSpecific returns true if a is more specific than b in the hierarchy.
func IsMoreSpecific(a, b PolicyScope) bool {
	return ScopeIndex(a) > ScopeIndex(b)
}
