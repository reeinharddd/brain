package skills

import (
	"time"
)

// SkillStatus represents skill lifecycle state
type SkillStatus string

const (
	SkillActive     SkillStatus = "active"
	SkillDeprecated SkillStatus = "deprecated"
	SkillArchived   SkillStatus = "archived"
	SkillPending    SkillStatus = "pending_review"
)

// SkillCompatibility describes which surfaces support this skill
type SkillCompatibility struct {
	MinCapabilityTier int
	Surfaces          map[string]string // surface -> minimum version
}

// SecurityScanResult holds 8-point security scan results
type SecurityScanResult struct {
	ScannedAt   time.Time
	OverallPass bool
	Checks      map[string]SecurityCheck
}

// SecurityCheck represents one security check result
type SecurityCheck struct {
	Name        string
	Passed      bool
	Severity    string // "critical", "high", "medium", "low"
	Description string
	Findings    []string
}

// Skill represents a managed skill definition
type Skill struct {
	ID             string
	Name           string
	Version        string
	Description    string
	Category       string
	Tags           []string
	Status         SkillStatus
	Compatibility  SkillCompatibility
	SecurityResult *SecurityScanResult
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Content        map[string]string // filename -> content
	Prerequisites  []string
}
