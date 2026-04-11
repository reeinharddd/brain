package artifacts

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ArtifactKey uniquely identifies an artifact by kind and ID.
type ArtifactKey struct {
	Kind string // skill, mcp, agent, rule
	ID   string
}

// ArtifactRecord holds the full record of a registered artifact.
type ArtifactRecord struct {
	Envelope       ArtifactEnvelope
	Dependencies   []ArtifactDependency
	VersionHistory []VersionEntry
	UsageMetrics   UsageMetrics
}

// ArtifactDependency describes a dependency on another artifact.
type ArtifactDependency struct {
	Kind        string
	ID          string
	VersionReq  string // semver constraint
	Optional    bool
	Description string
}

// ArtifactEnvelope contains the metadata for an artifact.
type ArtifactEnvelope struct {
	ID          string
	Kind        string
	Version     string
	Name        string
	Description string
	Scope       string // org, user, workspace, project
	Visibility  string // public, internal, private
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Checksum    string // sha256
	State       string // active, deprecated, archived
}

// VersionEntry records a historical version of an artifact.
type VersionEntry struct {
	Version   string
	CreatedAt time.Time
	Checksum  string
	Changelog string
}

// UsageMetrics tracks how an artifact is used at runtime.
type UsageMetrics struct {
	TotalActivations int
	LastActivated    time.Time
	AvgDuration      time.Duration
	SuccessRate      float64
	SurfacesUsed     map[string]int
	// internal tracking for computing averages
	totalDuration time.Duration
	successCount  int
}

// ValidationError reports a single dependency resolution problem.
type ValidationError struct {
	Key     ArtifactKey
	Dep     ArtifactDependency
	Message string
}

func (v ValidationError) String() string {
	return fmt.Sprintf("[%s:%s] depends on [%s:%s] (%s): %s",
		v.Key.Kind, v.Key.ID, v.Dep.Kind, v.Dep.ID, v.Dep.VersionReq, v.Message)
}

// ArtifactRegistry manages artifact registration, dependency tracking,
// version resolution, and usage analytics.
type ArtifactRegistry struct {
	mu           sync.RWMutex
	artifacts    map[ArtifactKey]*ArtifactRecord
	dependencies map[ArtifactKey][]ArtifactDependency
	reverseDeps  map[ArtifactKey][]ArtifactKey // artifact -> required-by (who depends on me)
	byKind       map[string][]ArtifactKey
}

// NewArtifactRegistry creates a ready-to-use registry.
func NewArtifactRegistry() *ArtifactRegistry {
	return &ArtifactRegistry{
		artifacts:    make(map[ArtifactKey]*ArtifactRecord),
		dependencies: make(map[ArtifactKey][]ArtifactDependency),
		reverseDeps:  make(map[ArtifactKey][]ArtifactKey),
		byKind:       make(map[string][]ArtifactKey),
	}
}

// Register adds an artifact to the registry. Returns an error if the
// artifact (same kind+id) is already registered.
func (r *ArtifactRegistry) Register(ctx context.Context, key ArtifactKey, envelope ArtifactEnvelope) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("register %s:%s: %w", key.Kind, key.ID, ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.artifacts[key]; exists {
		return fmt.Errorf("artifact %s:%s already registered: %w", key.Kind, key.ID, ErrDuplicate)
	}

	if envelope.Version == "" {
		envelope.Version = "0.0.0"
	}
	if envelope.CreatedAt.IsZero() {
		envelope.CreatedAt = time.Now()
	}
	if envelope.UpdatedAt.IsZero() {
		envelope.UpdatedAt = time.Now()
	}
	if envelope.State == "" {
		envelope.State = "active"
	}
	if envelope.Checksum == "" {
		h := sha256.Sum256([]byte(envelope.ID + envelope.Version + envelope.Name))
		envelope.Checksum = fmt.Sprintf("%x", h[:])
	}

	record := &ArtifactRecord{
		Envelope: envelope,
		VersionHistory: []VersionEntry{
			{
				Version:   envelope.Version,
				CreatedAt: envelope.CreatedAt,
				Checksum:  envelope.Checksum,
			},
		},
		UsageMetrics: UsageMetrics{
			SurfacesUsed: make(map[string]int),
		},
	}

	r.artifacts[key] = record
	r.byKind[key.Kind] = append(r.byKind[key.Kind], key)

	return nil
}

// Get retrieves a registered artifact by key.
func (r *ArtifactRegistry) Get(ctx context.Context, key ArtifactKey) (*ArtifactRecord, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("get %s:%s: %w", key.Kind, key.ID, ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.artifacts[key]
	if !exists {
		return nil, fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	// Return a copy to prevent mutation
	recordCopy := *record
	recordCopy.UsageMetrics.SurfacesUsed = make(map[string]int)
	for k, v := range record.UsageMetrics.SurfacesUsed {
		recordCopy.UsageMetrics.SurfacesUsed[k] = v
	}
	depCopy := make([]ArtifactDependency, len(record.Dependencies))
	copy(depCopy, record.Dependencies)
	recordCopy.Dependencies = depCopy
	verCopy := make([]VersionEntry, len(record.VersionHistory))
	copy(verCopy, record.VersionHistory)
	recordCopy.VersionHistory = verCopy

	return &recordCopy, nil
}

// List returns all artifact keys of a given kind. If kind is empty,
// returns all artifacts.
func (r *ArtifactRegistry) List(ctx context.Context, kind string) ([]ArtifactKey, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("list: %w", ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if kind == "" {
		result := make([]ArtifactKey, 0, len(r.artifacts))
		for k := range r.artifacts {
			result = append(result, k)
		}
		sort.Slice(result, func(i, j int) bool {
			if result[i].Kind != result[j].Kind {
				return result[i].Kind < result[j].Kind
			}
			return result[i].ID < result[j].ID
		})
		return result, nil
	}

	keys, exists := r.byKind[kind]
	if !exists {
		return []ArtifactKey{}, nil
	}
	result := make([]ArtifactKey, len(keys))
	copy(result, keys)
	return result, nil
}

// AddDependency records that key depends on the artifact described by dep.
func (r *ArtifactRegistry) AddDependency(ctx context.Context, key ArtifactKey, dep ArtifactDependency) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("add dependency: %w", ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.artifacts[key]; !exists {
		return fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	record := r.artifacts[key]

	// Check for duplicate dependency
	for _, existing := range record.Dependencies {
		if existing.Kind == dep.Kind && existing.ID == dep.ID {
			return fmt.Errorf("duplicate dependency %s:%s on %s:%s: %w",
				key.Kind, key.ID, dep.Kind, dep.ID, ErrDuplicate)
		}
	}

	record.Dependencies = append(record.Dependencies, dep)
	r.dependencies[key] = record.Dependencies

	// Update reverse dependency map
	depKey := ArtifactKey{Kind: dep.Kind, ID: dep.ID}
	var found bool
	for _, k := range r.reverseDeps[depKey] {
		if k == key {
			found = true
			break
		}
	}
	if !found {
		r.reverseDeps[depKey] = append(r.reverseDeps[depKey], key)
	}

	return nil
}

// ResolveDependencies returns the full transitive dependency chain for key,
// in depth-first order, with deduplication.
func (r *ArtifactRegistry) ResolveDependencies(ctx context.Context, key ArtifactKey) ([]ArtifactKey, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("resolve dependencies: %w", ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.artifacts[key]; !exists {
		return nil, fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	seen := make(map[ArtifactKey]bool)
	var result []ArtifactKey
	var resolve func(ArtifactKey) error
	resolve = func(k ArtifactKey) error {
		if seen[k] {
			return nil
		}
		seen[k] = true

		record, ok := r.artifacts[k]
		if !ok {
			return nil
		}

		for _, dep := range record.Dependencies {
			depKey := ArtifactKey{Kind: dep.Kind, ID: dep.ID}
			if _, depExists := r.artifacts[depKey]; !depExists && !dep.Optional {
				// Dependency not found and not optional -- still include in results
				// but note it as a resolution issue; we don't fail hard here
				// so callers can decide what to do. Validation is separate.
			}
			if err := resolve(depKey); err != nil {
				return err
			}
		}

		result = append(result, k)
		return nil
	}

	if err := resolve(key); err != nil {
		return nil, err
	}

	// Remove the root key itself from the result (we only return dependencies)
	if len(result) > 0 && result[len(result)-1] == key {
		result = result[:len(result)-1]
	}

	return result, nil
}

// RecordUsage logs a single activation of an artifact on a given surface.
func (r *ArtifactRegistry) RecordUsage(ctx context.Context, key ArtifactKey, surface string, duration time.Duration, success bool) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("record usage: %w", ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.artifacts[key]
	if !exists {
		return fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	m := &record.UsageMetrics
	m.TotalActivations++
	m.LastActivated = time.Now()
	m.totalDuration += duration

	if success {
		m.successCount++
	}

	if m.TotalActivations > 0 {
		m.AvgDuration = m.totalDuration / time.Duration(m.TotalActivations)
		m.SuccessRate = float64(m.successCount) / float64(m.TotalActivations)
	}

	if surface != "" {
		m.SurfacesUsed[surface]++
	}

	return nil
}

// GetUsageMetrics returns a copy of the current usage metrics for an artifact.
func (r *ArtifactRegistry) GetUsageMetrics(ctx context.Context, key ArtifactKey) (*UsageMetrics, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("get usage metrics %s:%s: %w", key.Kind, key.ID, ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.artifacts[key]
	if !exists {
		return nil, fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	// Return a copy
	m := record.UsageMetrics
	// Copy the map
	surfaces := make(map[string]int)
	for k, v := range m.SurfacesUsed {
		surfaces[k] = v
	}
	m.SurfacesUsed = surfaces
	return &m, nil
}

// Deprecate marks an artifact as deprecated without removing it.
func (r *ArtifactRegistry) Deprecate(ctx context.Context, key ArtifactKey) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("deprecate %s:%s: %w", key.Kind, key.ID, ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.artifacts[key]
	if !exists {
		return fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	record.Envelope.State = "deprecated"
	record.Envelope.UpdatedAt = time.Now()
	return nil
}

// Rollback reverts an artifact to a previously recorded version.
func (r *ArtifactRegistry) Rollback(ctx context.Context, key ArtifactKey, version string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("rollback %s:%s: %w", key.Kind, key.ID, ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.artifacts[key]
	if !exists {
		return fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	// Find the version in history
	var target *VersionEntry
	for i := range record.VersionHistory {
		if record.VersionHistory[i].Version == version {
			target = &record.VersionHistory[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("version %q not found in history for %s:%s: %w", version, key.Kind, key.ID, ErrNotFound)
	}

	// Restore version state
	record.Envelope.Version = target.Version
	record.Envelope.Checksum = target.Checksum
	record.Envelope.UpdatedAt = time.Now()

	return nil
}

// FindVersion returns the best matching version entry for a key based on a semver constraint.
func (r *ArtifactRegistry) FindVersion(ctx context.Context, key ArtifactKey, versionReq string) (*ArtifactRecord, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("find version: %w", ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.artifacts[key]
	if !exists {
		return nil, fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	if versionReq == "" || versionReq == "latest" {
		// Find highest version
		return highestVersionRecord(record)
	}

	// Find best matching version from history
	var best *VersionEntry
	for i := range record.VersionHistory {
		candidate := &record.VersionHistory[i]
		if MatchVersion(candidate.Version, versionReq) {
			if best == nil || CompareVersions(candidate.Version, best.Version) > 0 {
				best = candidate
			}
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no version matching %q for %s:%s: %w", versionReq, key.Kind, key.ID, ErrNotFound)
	}

	// Return a copy with the resolved version
	recordCopy := *record
	recordCopy.Envelope.Version = best.Version
	recordCopy.Envelope.Checksum = best.Checksum
	recordCopy.UsageMetrics.SurfacesUsed = make(map[string]int)
	for k, v := range record.UsageMetrics.SurfacesUsed {
		recordCopy.UsageMetrics.SurfacesUsed[k] = v
	}
	depCopy := make([]ArtifactDependency, len(record.Dependencies))
	copy(depCopy, record.Dependencies)
	recordCopy.Dependencies = depCopy
	verCopy := make([]VersionEntry, len(record.VersionHistory))
	copy(verCopy, record.VersionHistory)
	recordCopy.VersionHistory = verCopy

	return &recordCopy, nil
}

// ValidateDependencies checks all registered artifacts and returns any
// dependency resolution errors found.
func (r *ArtifactRegistry) ValidateDependencies(ctx context.Context) []ValidationError {
	select {
	case <-ctx.Done():
		return []ValidationError{{Message: ctx.Err().Error()}}
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var errors []ValidationError

	for key, record := range r.artifacts {
		for _, dep := range record.Dependencies {
			depKey := ArtifactKey{Kind: dep.Kind, ID: dep.ID}

			depRecord, exists := r.artifacts[depKey]
			if !exists {
				if dep.Optional {
					continue // optional deps that don't exist are fine
				}
				errors = append(errors, ValidationError{
					Key:     key,
					Dep:     dep,
					Message: "dependency not found",
				})
				continue
			}

			// Check version constraint
			if dep.VersionReq != "" && !MatchVersion(depRecord.Envelope.Version, dep.VersionReq) {
				errors = append(errors, ValidationError{
					Key:     key,
					Dep:     dep,
					Message: fmt.Sprintf("version %q does not satisfy constraint %q", depRecord.Envelope.Version, dep.VersionReq),
				})
			}
		}
	}

	return errors
}

// Size returns the total number of registered artifacts.
func (r *ArtifactRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.artifacts)
}

// AddVersion records a new version in the artifact's history.
func (r *ArtifactRegistry) AddVersion(ctx context.Context, key ArtifactKey, version, changelog string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("add version: %w", ctx.Err())
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.artifacts[key]
	if !exists {
		return fmt.Errorf("artifact %s:%s not found: %w", key.Kind, key.ID, ErrNotFound)
	}

	h := sha256.Sum256([]byte(key.Kind + key.ID + version))
	checksum := fmt.Sprintf("%x", h[:])

	entry := VersionEntry{
		Version:   version,
		CreatedAt: time.Now(),
		Checksum:  checksum,
		Changelog: changelog,
	}
	record.VersionHistory = append(record.VersionHistory, entry)
	record.Envelope.Version = version
	record.Envelope.Checksum = checksum
	record.Envelope.UpdatedAt = time.Now()

	return nil
}

// ReverseDeps returns all artifact keys that depend on the given key.
func (r *ArtifactRegistry) ReverseDeps(ctx context.Context, key ArtifactKey) ([]ArtifactKey, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("reverse deps: %w", ctx.Err())
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	deps := r.reverseDeps[key]
	result := make([]ArtifactKey, len(deps))
	copy(result, deps)
	return result, nil
}

// Sentinel errors for the artifacts package.
var (
	ErrNotFound  = fmt.Errorf("not found")
	ErrDuplicate = fmt.Errorf("duplicate")
)

// highestVersionRecord returns a copy of the record with the highest version from history.
func highestVersionRecord(record *ArtifactRecord) (*ArtifactRecord, error) {
	if len(record.VersionHistory) == 0 {
		return nil, fmt.Errorf("no versions in history")
	}

	best := &record.VersionHistory[0]
	for i := range record.VersionHistory {
		if CompareVersions(record.VersionHistory[i].Version, best.Version) > 0 {
			best = &record.VersionHistory[i]
		}
	}

	recordCopy := *record
	recordCopy.Envelope.Version = best.Version
	recordCopy.Envelope.Checksum = best.Checksum
	recordCopy.UsageMetrics.SurfacesUsed = make(map[string]int)
	for k, v := range record.UsageMetrics.SurfacesUsed {
		recordCopy.UsageMetrics.SurfacesUsed[k] = v
	}
	depCopy := make([]ArtifactDependency, len(record.Dependencies))
	copy(depCopy, record.Dependencies)
	recordCopy.Dependencies = depCopy
	verCopy := make([]VersionEntry, len(record.VersionHistory))
	copy(verCopy, record.VersionHistory)
	recordCopy.VersionHistory = verCopy

	return &recordCopy, nil
}
