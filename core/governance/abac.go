package governance

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// ABACCondition defines a condition for access control.
type ABACCondition struct {
	Attribute string      // e.g., "artifact.kind", "user.department"
	Operator  string      // "eq", "neq", "in", "contains", "gt", "lt"
	Value     interface{}
}

// ABACRule defines an attribute-based access rule.
type ABACRule struct {
	ID          string
	Name        string
	Conditions  []ABACCondition
	Effect      string // "allow" or "deny"
	Priority    int    // Higher = evaluated first
	Description string
}

// ABAC manages attribute-based rules.
type ABAC struct {
	mu    sync.RWMutex
	rules []ABACRule
}

// NewABAC creates a new ABAC instance.
func NewABAC() *ABAC {
	return &ABAC{
		rules: make([]ABACRule, 0),
	}
}

// AddRule adds an ABAC rule.
func (a *ABAC) AddRule(ctx context.Context, rule ABACRule) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}

	// Check for duplicate
	for _, existing := range a.rules {
		if existing.ID == rule.ID {
			return fmt.Errorf("rule %q already exists", rule.ID)
		}
	}

	// Validate effect
	if rule.Effect != "allow" && rule.Effect != "deny" {
		return fmt.Errorf("invalid effect: %q (must be 'allow' or 'deny')", rule.Effect)
	}

	a.rules = append(a.rules, rule)
	return nil
}

// RemoveRule removes an ABAC rule by ID.
func (a *ABAC) RemoveRule(ctx context.Context, ruleID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, rule := range a.rules {
		if rule.ID == ruleID {
			a.rules = append(a.rules[:i], a.rules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("rule %q not found", ruleID)
}

// Evaluate returns all applicable deny rules for the given attributes.
func (a *ABAC) Evaluate(ctx context.Context, attributes map[string]interface{}) []ABACRule {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Sort rules by priority (descending)
	sorted := make([]ABACRule, len(a.rules))
	copy(sorted, a.rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	var denyRules []ABACRule
	for _, rule := range sorted {
		if rule.Effect == "deny" && a.matchesAllConditions(rule.Conditions, attributes) {
			denyRules = append(denyRules, rule)
		}
	}

	return denyRules
}

// IsAllowed checks if access is allowed for the given attributes.
func (a *ABAC) IsAllowed(ctx context.Context, attributes map[string]interface{}) (bool, string) {
	denyRules := a.Evaluate(ctx, attributes)
	if len(denyRules) > 0 {
		return false, denyRules[0].Name
	}
	return true, ""
}

// GetRules returns all ABAC rules.
func (a *ABAC) GetRules(ctx context.Context) []ABACRule {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]ABACRule, len(a.rules))
	copy(result, a.rules)
	return result
}

// matchesAllConditions checks if all conditions in a rule match the given attributes.
func (a *ABAC) matchesAllConditions(conditions []ABACCondition, attributes map[string]interface{}) bool {
	for _, cond := range conditions {
		attrValue, exists := attributes[cond.Attribute]
		if !exists {
			return false
		}

		if !a.evaluateCondition(cond, attrValue) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition against an attribute value.
func (a *ABAC) evaluateCondition(cond ABACCondition, attrValue interface{}) bool {
	switch cond.Operator {
	case "eq":
		return attrValue == cond.Value
	case "neq":
		return attrValue != cond.Value
	case "in":
		// cond.Value should be a slice or array
		return containsValue(cond.Value, attrValue)
	case "contains":
		// attrValue should be a slice, check if it contains cond.Value
		return containsValue(attrValue, cond.Value)
	case "gt":
		return compareNumbers(attrValue, cond.Value) > 0
	case "lt":
		return compareNumbers(attrValue, cond.Value) < 0
	default:
		return false
	}
}

// containsValue checks if a slice/array contains a specific value.
func containsValue(slice, value interface{}) bool {
	switch s := slice.(type) {
	case []interface{}:
		for _, v := range s {
			if v == value {
				return true
			}
		}
	case []string:
		v, ok := value.(string)
		if !ok {
			return false
		}
		for _, item := range s {
			if item == v {
				return true
			}
		}
	case []int:
		v, ok := value.(int)
		if !ok {
			return false
		}
		for _, item := range s {
			if item == v {
				return true
			}
		}
	case string:
		v, ok := value.(string)
		if !ok {
			return false
		}
		// Check if string contains substring
		return stringContains(s, v)
	}
	return false
}

// stringContains is a simple string contains check.
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if substr == "" {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// compareNumbers compares two numeric values. Returns -1, 0, or 1.
func compareNumbers(a, b interface{}) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)

	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}

// toFloat64 converts a numeric value to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	default:
		return 0
	}
}
