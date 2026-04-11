package governance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ResolutionResult is the resolved policy for a request.
type ResolutionResult struct {
	ResolvedAt time.Time
	ScopeChain []string                // e.g., ["org:myorg", "user:alice", "workspace:brain"]
	Rules      map[string]*ResolvedRule
}

// ResolvedRule is a rule after resolution.
type ResolvedRule struct {
	ID              string
	Value           string
	Source          string // which scope defined it
	Class           PolicyClass
	OverrideAllowed bool
}

// OverrideRequest represents a policy override request.
type OverrideRequest struct {
	ID          string
	RuleID      string
	OldValue    string
	NewValue    string
	Requester   string
	Approved    *bool // nil = pending, true = approved, false = denied
	RequestedAt time.Time
	DecidedAt   time.Time
	DecidedBy   string
}

// PolicyResolver resolves policies through the hierarchy.
type PolicyResolver struct {
	mu         sync.RWMutex
	policySets map[PolicyScope]map[string]*PolicySet // scope -> scopeID -> set
	auditLog   *AuditLog

	// Track override requests
	overrideRequests map[string]*OverrideRequest // request ID -> request
	overrideMu       sync.RWMutex
}

// NewPolicyResolver creates a new PolicyResolver.
func NewPolicyResolver() *PolicyResolver {
	return &PolicyResolver{
		policySets:       make(map[PolicyScope]map[string]*PolicySet),
		auditLog:         NewAuditLog(10000),
		overrideRequests: make(map[string]*OverrideRequest),
	}
}

// SetPolicy sets a policy rule at a given scope.
func (r *PolicyResolver) SetPolicy(ctx context.Context, scope PolicyScope, scopeID string, rule *PolicyRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Parse scopeID to get the canonical key
	_, id := parseScopeID(scopeID)
	if id == "" {
		id = scopeID // fallback: use as-is if not parseable
	}

	if _, exists := r.policySets[scope]; !exists {
		r.policySets[scope] = make(map[string]*PolicySet)
	}

	ps, exists := r.policySets[scope][id]
	if !exists {
		var err error
		ps, err = NewPolicySet(scope, scopeID)
		if err != nil {
			return fmt.Errorf("failed to create policy set: %w", err)
		}
		r.policySets[scope][id] = ps
	}

	// Check for override of hard policy
	if existing, err := r.findRuleUpHierarchy(ctx, scope, scopeID, rule.ID); err == nil {
		if existing.Class == PolicyHard {
			return fmt.Errorf("cannot override hard policy %q defined at %s", rule.ID, existing.Source)
		}
	}

	if err := ps.AddRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to add rule to policy set: %w", err)
	}

	r.auditLog.Log(AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    "policy_set",
		Subject:   scopeID,
		Resource:  rule.ID,
		Details: map[string]string{
			"scope": string(scope),
			"class": string(rule.Class),
			"value": rule.Value,
		},
		Success: true,
	})

	return nil
}

// Resolve resolves policies for the given scope chain.
func (r *PolicyResolver) Resolve(ctx context.Context, scopeChain []string) (*ResolutionResult, error) {
	if len(scopeChain) == 0 {
		return nil, fmt.Errorf("scope chain cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := &ResolutionResult{
		ResolvedAt: time.Now().UTC(),
		ScopeChain: scopeChain,
		Rules:      make(map[string]*ResolvedRule),
	}

	// Walk scope chain from most general to most specific
	for _, scopeID := range scopeChain {
		scope, id := parseScopeID(scopeID)
		if scope == "" {
			continue
		}

		ps, exists := r.policySets[scope][id]
		if !exists {
			continue
		}

		for ruleID, rule := range ps.Rules {
			if existing, ok := result.Rules[ruleID]; ok {
				// Rule already defined by a more general scope
				// Check if this scope can override
				if canOverride(existing.Class, rule) {
					result.Rules[ruleID] = &ResolvedRule{
						ID:              ruleID,
						Value:           rule.Value,
						Source:          scopeID,
						Class:           rule.Class,
						OverrideAllowed: rule.OverrideAllowed,
					}
				}
				// If cannot override, keep the existing (more general) rule
			} else {
				result.Rules[ruleID] = &ResolvedRule{
					ID:              ruleID,
					Value:           rule.Value,
					Source:          scopeID,
					Class:           rule.Class,
					OverrideAllowed: rule.OverrideAllowed,
				}
			}
		}
	}

	r.auditLog.Log(AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    "policy_resolved",
		Subject:   scopeChain[len(scopeChain)-1],
		Resource:  fmt.Sprintf("scope_chain:%d_rules", len(result.Rules)),
		Details: map[string]string{
			"scope_chain": fmt.Sprintf("%v", scopeChain),
		},
		Success: true,
	})

	return result, nil
}

// Check checks a specific rule in the scope chain.
func (r *PolicyResolver) Check(ctx context.Context, scopeChain []string, ruleID string) (*ResolvedRule, error) {
	result, err := r.Resolve(ctx, scopeChain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve policies: %w", err)
	}

	rule, exists := result.Rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %q not found in scope chain", ruleID)
	}

	return rule, nil
}

// CanOverride checks if a rule can be overridden.
func (r *PolicyResolver) CanOverride(ctx context.Context, scopeChain []string, ruleID string) (bool, error) {
	rule, err := r.Check(ctx, scopeChain, ruleID)
	if err != nil {
		return false, err
	}

	switch rule.Class {
	case PolicyHard:
		return false, nil
	case PolicySoft:
		return true, nil
	case PolicyGuarded:
		return false, nil // requires approval
	default:
		return false, fmt.Errorf("unknown policy class: %q", rule.Class)
	}
}

// RequestOverride creates a policy override request.
func (r *PolicyResolver) RequestOverride(ctx context.Context, scopeChain []string, ruleID string, newValue string, requester string) (*OverrideRequest, error) {
	rule, err := r.Check(ctx, scopeChain, ruleID)
	if err != nil {
		return nil, fmt.Errorf("cannot request override: %w", err)
	}

	if rule.Class == PolicyHard {
		return nil, fmt.Errorf("cannot override hard policy %q", ruleID)
	}

	reqID := uuid.New().String()
	overrideReq := &OverrideRequest{
		ID:          reqID,
		RuleID:      ruleID,
		OldValue:    rule.Value,
		NewValue:    newValue,
		Requester:   requester,
		Approved:    nil, // pending
		RequestedAt: time.Now().UTC(),
	}

	r.overrideMu.Lock()
	r.overrideRequests[reqID] = overrideReq
	r.overrideMu.Unlock()

	r.auditLog.Log(AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    "override_requested",
		Subject:   requester,
		Resource:  ruleID,
		Details: map[string]string{
			"request_id": reqID,
			"old_value":  rule.Value,
			"new_value":  newValue,
		},
		Success: true,
	})

	return overrideReq, nil
}

// ApproveOverride approves or denies an override request.
func (r *PolicyResolver) ApproveOverride(ctx context.Context, requestID string, approved bool, decidedBy string) error {
	r.overrideMu.Lock()
	defer r.overrideMu.Unlock()

	req, exists := r.overrideRequests[requestID]
	if !exists {
		return fmt.Errorf("override request %q not found", requestID)
	}

	if req.Approved != nil {
		return fmt.Errorf("override request %q already decided", requestID)
	}

	req.Approved = &approved
	req.DecidedAt = time.Now().UTC()
	req.DecidedBy = decidedBy

	action := "override_denied"
	if approved {
		action = "override_approved"
	}

	r.auditLog.Log(AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    action,
		Subject:   decidedBy,
		Resource:  req.RuleID,
		Details: map[string]string{
			"request_id": requestID,
			"approved":   fmt.Sprintf("%v", approved),
		},
		Success: true,
	})

	return nil
}

// GetPolicySets returns the policy set for a given scope.
func (r *PolicyResolver) GetPolicySets(ctx context.Context, scope PolicyScope, scopeID string) (*PolicySet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Parse scopeID to get the canonical key
	_, id := parseScopeID(scopeID)
	if id == "" {
		id = scopeID
	}

	ps, exists := r.policySets[scope][id]
	if !exists {
		return nil, fmt.Errorf("policy set not found for %s:%s", scope, scopeID)
	}

	return ps, nil
}

// Count returns the total number of rules across all policy sets.
func (r *PolicyResolver) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, scopeMap := range r.policySets {
		for _, ps := range scopeMap {
			count += len(ps.Rules)
		}
	}
	return count
}

// GetAuditLog returns the audit log.
func (r *PolicyResolver) GetAuditLog() *AuditLog {
	return r.auditLog
}

// GetOverrideRequest returns an override request by ID.
func (r *PolicyResolver) GetOverrideRequest(ctx context.Context, requestID string) (*OverrideRequest, error) {
	r.overrideMu.RLock()
	defer r.overrideMu.RUnlock()

	req, exists := r.overrideRequests[requestID]
	if !exists {
		return nil, fmt.Errorf("override request %q not found", requestID)
	}

	return req, nil
}

// findRuleUpHierarchy searches for a rule in parent scopes.
func (r *PolicyResolver) findRuleUpHierarchy(ctx context.Context, currentScope PolicyScope, currentScopeID string, ruleID string) (*ResolvedRule, error) {
	hierarchy := ScopeHierarchy()
	currentIdx := ScopeIndex(currentScope)

	for i := 0; i < currentIdx; i++ {
		scope := hierarchy[i]
		for scopeID, ps := range r.policySets[scope] {
			if rule, exists := ps.Rules[ruleID]; exists {
				return &ResolvedRule{
					ID:    ruleID,
					Value: rule.Value,
					Source: fmt.Sprintf("%s:%s", scope, scopeID),
					Class: rule.Class,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("rule %q not found in parent scopes", ruleID)
}

// canOverride determines if a rule can be overridden by a more specific scope.
func canOverride(existingClass PolicyClass, newRule *PolicyRule) bool {
	switch existingClass {
	case PolicyHard:
		return false
	case PolicySoft:
		return true
	case PolicyGuarded:
		return false // requires approval flow
	default:
		return true
	}
}

// parseScopeID parses a scope ID like "org:acme" into scope and ID.
func parseScopeID(scopeID string) (PolicyScope, string) {
	// Map prefixes to scopes
	prefixes := map[string]PolicyScope{
		"platform:":  ScopePlatform,
		"org:":       ScopeOrg,
		"team:":      ScopeTeam,
		"user:":      ScopeUser,
		"workspace:": ScopeWorkspace,
		"project:":   ScopeProject,
		"session:":   ScopeSession,
	}

	for prefix, scope := range prefixes {
		if len(scopeID) > len(prefix) && scopeID[:len(prefix)] == prefix {
			return scope, scopeID[len(prefix):]
		}
	}

	return "", ""
}
