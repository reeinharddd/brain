package skills

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SkillRegistry manages skill definitions
type SkillRegistry struct {
	mu      sync.RWMutex
	skills  map[string]map[string]*Skill // id -> version -> skill
	scanner *SecurityScanner
}

// NewSkillRegistry creates a new skill registry
func NewSkillRegistry(scanner *SecurityScanner) *SkillRegistry {
	if scanner == nil {
		scanner = NewSecurityScanner()
	}
	return &SkillRegistry{
		skills:  make(map[string]map[string]*Skill),
		scanner: scanner,
	}
}

// Register adds a skill to the registry after running a security scan
func (r *SkillRegistry) Register(ctx context.Context, skill *Skill) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Run security scan
	result, err := r.scanner.Scan(ctx, skill)
	if err != nil {
		return fmt.Errorf("security scan failed: %w", err)
	}
	skill.SecurityResult = result

	if !result.OverallPass {
		return fmt.Errorf("skill %s v%s failed security scan", skill.ID, skill.Version)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.skills[skill.ID] == nil {
		r.skills[skill.ID] = make(map[string]*Skill)
	}

	if _, exists := r.skills[skill.ID][skill.Version]; exists {
		return fmt.Errorf("skill %s version %s already exists", skill.ID, skill.Version)
	}

	now := time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = now
	}
	skill.UpdatedAt = now
	skill.Status = SkillActive

	r.skills[skill.ID][skill.Version] = skill
	return nil
}

// Get retrieves a specific skill version
func (r *SkillRegistry) Get(ctx context.Context, id, version string) (*Skill, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.skills[id]
	if !ok {
		return nil, fmt.Errorf("skill %s not found", id)
	}

	skill, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("skill %s version %s not found", id, version)
	}

	return skill, nil
}

// GetLatest retrieves the latest version of a skill
func (r *SkillRegistry) GetLatest(ctx context.Context, id string) (*Skill, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.skills[id]
	if !ok {
		return nil, fmt.Errorf("skill %s not found", id)
	}

	var latest *Skill
	for _, skill := range versions {
		if latest == nil || skill.UpdatedAt.After(latest.UpdatedAt) {
			latest = skill
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("skill %s has no versions", id)
	}

	return latest, nil
}

// List returns all registered skill IDs
func (r *SkillRegistry) List(ctx context.Context) []string {
	if err := ctx.Err(); err != nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.skills))
	for id := range r.skills {
		ids = append(ids, id)
	}
	return ids
}

// ListVersions returns all versions for a given skill ID
func (r *SkillRegistry) ListVersions(ctx context.Context, id string) []string {
	if err := ctx.Err(); err != nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.skills[id]
	if !ok {
		return nil
	}

	versionList := make([]string, 0, len(versions))
	for v := range versions {
		versionList = append(versionList, v)
	}
	return versionList
}

// Search finds skills matching query and/or category
func (r *SkillRegistry) Search(ctx context.Context, query, category string) []*Skill {
	if err := ctx.Err(); err != nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Skill
	for _, versions := range r.skills {
		for _, skill := range versions {
			if category != "" && skill.Category != category {
				continue
			}
			if query != "" {
				if !matchesQuery(skill, query) {
					continue
				}
			}
			results = append(results, skill)
		}
	}
	return results
}

func matchesQuery(skill *Skill, query string) bool {
	q := query
	if skill.ID == q || skill.Name == q || skill.Description == q {
		return true
	}
	for _, tag := range skill.Tags {
		if tag == q {
			return true
		}
	}
	return false
}

// Deprecate marks a skill version as deprecated
func (r *SkillRegistry) Deprecate(ctx context.Context, id, version string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	versions, ok := r.skills[id]
	if !ok {
		return fmt.Errorf("skill %s not found", id)
	}

	skill, ok := versions[version]
	if !ok {
		return fmt.Errorf("skill %s version %s not found", id, version)
	}

	skill.Status = SkillDeprecated
	skill.UpdatedAt = time.Now()
	return nil
}

// Delete removes a skill version from the registry
func (r *SkillRegistry) Delete(ctx context.Context, id, version string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	versions, ok := r.skills[id]
	if !ok {
		return fmt.Errorf("skill %s not found", id)
	}

	if _, ok := versions[version]; !ok {
		return fmt.Errorf("skill %s version %s not found", id, version)
	}

	delete(versions, version)
	if len(versions) == 0 {
		delete(r.skills, id)
	}
	return nil
}

// Count returns the total number of skill versions registered
func (r *SkillRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, versions := range r.skills {
		count += len(versions)
	}
	return count
}
