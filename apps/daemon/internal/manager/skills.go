package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	coreartifacts "github.com/reeinharrrd/brain/core/artifacts"
	"gopkg.in/yaml.v3"
)

// Skill represents a single skill from registry
type Skill struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Type        string   `yaml:"type" json:"type"` // internal or external
	Description string   `yaml:"description" json:"description"`
	Tags        []string `yaml:"tags" json:"tags"`
	File        string   `yaml:"file" json:"file"`
	SyncTo      []string `yaml:"sync-to" json:"sync_to"` // targets: cli, daemon, vscode, cursor, etc
	Requires    []string `yaml:"requires" json:"requires"`
	Maintained  bool     `yaml:"maintained" json:"maintained"`
	Category    string   `yaml:"category" json:"category"`
}

// CatalogItem represents a skill or context-pack with canonical fields and legacy aliases
type CatalogItem struct {
	// Canonical fields
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`  // "skill" or "context-pack"
	Scope       string   `json:"scope"` // "global" or "workspace"
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Path        string   `json:"path"` // file or context path

	// Metadata
	Version       string `json:"version,omitempty"`
	Maintained    bool   `json:"maintained"`
	Source        string `json:"source"` // "registry.yml" or "dynamic-registry.tsv"
	SourceType    string `json:"source_type,omitempty"`
	SourceURI     string `json:"source_uri,omitempty"`
	SourceVariant string `json:"source_variant,omitempty"`
	ArtifactPath  string `json:"artifact_path,omitempty"`

	// Legacy aliases for backward compatibility with CLI and existing consumers
	Type     string   `json:"type,omitempty"`    // alias for Kind in legacy YAML (internal/external or context-pack)
	File     string   `json:"file,omitempty"`    // alias for Path
	SyncTo   []string `json:"sync_to,omitempty"` // targets
	Requires []string `json:"requires,omitempty"`
	Category string   `json:"category,omitempty"`
}

// SkillsRegistry manages all skills and context-packs
type SkillsRegistry struct {
	mu        sync.RWMutex
	catalog   map[string]*CatalogItem // unified catalog: skills + context-packs
	brainRoot string
	logCh     chan string
}

// NewSkillsRegistry creates a new skills registry
func NewSkillsRegistry(brainRoot string, logCh chan string) *SkillsRegistry {
	return &SkillsRegistry{
		catalog:   make(map[string]*CatalogItem),
		brainRoot: brainRoot,
		logCh:     logCh,
	}
}

func (r *SkillsRegistry) registryPath() string {
	return coreartifacts.NewLocator(r.brainRoot).DomainFile("skills", "registry.yml")
}

func (r *SkillsRegistry) contextPacksPath() string {
	return coreartifacts.NewLocator(r.brainRoot).DomainFile("skills", "dynamic-registry.tsv")
}

func (r *SkillsRegistry) skillsDir() string {
	return coreartifacts.NewLocator(r.brainRoot).DomainDir("skills")
}

func (r *SkillsRegistry) skillSourcesDir() string {
	candidates := []string{
		filepath.Join(r.brainRoot, ".github", "skills"),
		filepath.Join(r.brainRoot, "artifacts", "skills"),
		filepath.Join(r.brainRoot, "skills"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return candidates[0]
}

// Load reads skills from registry.yml and context-packs from dynamic-registry.tsv
func (r *SkillsRegistry) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.catalog = make(map[string]*CatalogItem)

	// Load skills from YAML registry
	if err := r.loadSkillsFromYAML(); err != nil {
		r.log(fmt.Sprintf("Warning: Error loading YAML registry: %v", err))
		// Continue even if YAML fails; TSV may still work
	}

	// Load context-packs from TSV registry
	if err := r.loadContextPacksFromTSV(); err != nil {
		r.log(fmt.Sprintf("Warning: Error loading TSV registry: %v", err))
		// Continue even if TSV fails; YAML may still work
	}

	r.log(fmt.Sprintf("Loaded %d items (skills + context-packs)", len(r.catalog)))
	return nil
}

// loadSkillsFromYAML loads skills from the YAML registry
func (r *SkillsRegistry) loadSkillsFromYAML() error {
	registryPath := r.registryPath()
	r.log(fmt.Sprintf("Loading skills from %s", registryPath))

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("cannot read skills registry: %w", err)
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("cannot parse skills registry: %w", err)
	}

	skillsData, ok := rawData["skills"].(map[string]interface{})
	if !ok {
		r.log("No 'skills' section in registry")
		return nil
	}

	count := 0
	for id, skillData := range skillsData {
		skillMap, ok := skillData.(map[string]interface{})
		if !ok {
			r.log(fmt.Sprintf("Invalid skill data for %s", id))
			continue
		}

		// Check for duplicate ID
		if _, exists := r.catalog[id]; exists {
			r.log(fmt.Sprintf("Warning: Duplicate ID '%s' found in both YAML and TSV. Skipping YAML entry.", id))
			continue
		}

		item := &CatalogItem{
			ID:     id,
			Kind:   "skill",
			Scope:  "global",
			Source: "registry.yml",
		}

		if v, ok := skillMap["id"].(string); ok {
			item.ID = v
		}
		if v, ok := skillMap["name"].(string); ok {
			item.Name = v
		}
		if v, ok := skillMap["version"].(string); ok {
			item.Version = v
		}
		if v, ok := skillMap["type"].(string); ok {
			item.Type = v
		}
		if v, ok := skillMap["description"].(string); ok {
			item.Description = v
		}
		if v, ok := skillMap["file"].(string); ok {
			item.File = v
			item.Path = v
		}
		if v, ok := skillMap["category"].(string); ok {
			item.Category = v
		}
		if v, ok := skillMap["source_type"].(string); ok {
			item.SourceType = v
		}
		if v, ok := skillMap["source_uri"].(string); ok {
			item.SourceURI = v
		}
		if v, ok := skillMap["source_variant"].(string); ok {
			item.SourceVariant = v
		}
		if v, ok := skillMap["artifact_path"].(string); ok {
			item.ArtifactPath = v
		}
		if v, ok := skillMap["maintained"].(bool); ok {
			item.Maintained = v
		} else {
			item.Maintained = true // default true
		}

		// Parse tags
		if tagsIface, ok := skillMap["tags"].([]interface{}); ok {
			for _, tag := range tagsIface {
				if t, ok := tag.(string); ok {
					item.Tags = append(item.Tags, t)
				}
			}
		}

		// Parse sync-to
		if syncIface, ok := skillMap["sync-to"].([]interface{}); ok {
			for _, s := range syncIface {
				if st, ok := s.(string); ok {
					item.SyncTo = append(item.SyncTo, st)
				}
			}
		}

		// Parse requires
		if reqIface, ok := skillMap["requires"].([]interface{}); ok {
			for _, req := range reqIface {
				if r, ok := req.(string); ok {
					item.Requires = append(item.Requires, r)
				}
			}
		}

		r.catalog[item.ID] = item
		count++
		r.log(fmt.Sprintf("Loaded skill: %s (v%s)", item.ID, item.Version))
	}

	r.log(fmt.Sprintf("Loaded %d skills from YAML", count))
	return nil
}

// loadContextPacksFromTSV loads context-packs from the TSV registry
func (r *SkillsRegistry) loadContextPacksFromTSV() error {
	tsvPath := r.contextPacksPath()
	r.log(fmt.Sprintf("Loading context-packs from %s", tsvPath))

	data, err := os.ReadFile(tsvPath)
	if err != nil {
		return fmt.Errorf("cannot read TSV registry: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for i, line := range lines {
		// Skip comment lines and empty lines
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip header (line 0 if it has 'skill_id')
		if i == 0 && strings.Contains(line, "skill_id") {
			continue
		}

		// Parse TSV line: skill_id \t title \t detect_tags \t context_path \t summary
		parts := strings.Split(line, "\t")
		if len(parts) < 5 {
			r.log(fmt.Sprintf("Invalid TSV line %d: expected 5 columns, got %d", i, len(parts)))
			continue
		}

		id := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		tagsStr := strings.TrimSpace(parts[2])
		contextPath := strings.TrimSpace(parts[3])
		summary := strings.TrimSpace(parts[4])

		// Check for duplicate ID
		if _, exists := r.catalog[id]; exists {
			r.log(fmt.Sprintf("Warning: Duplicate ID '%s' found. Keeping existing entry from YAML.", id))
			continue
		}

		item := &CatalogItem{
			ID:          id,
			Name:        title,
			Kind:        "context-pack",
			Scope:       "global",
			Description: summary,
			Path:        contextPath,
			Source:      "dynamic-registry.tsv",
			Maintained:  true,
		}

		// Parse comma-separated tags
		if tagsStr != "" {
			item.Tags = strings.Split(tagsStr, ",")
			// Trim spaces from each tag
			for i, tag := range item.Tags {
				item.Tags[i] = strings.TrimSpace(tag)
			}
		}

		r.catalog[item.ID] = item
		count++
		r.log(fmt.Sprintf("Loaded context-pack: %s", id))
	}

	r.log(fmt.Sprintf("Loaded %d context-packs from TSV", count))
	return nil
}

// GetAll returns all catalog items (skills and context-packs)
func (r *SkillsRegistry) GetAll(ctx context.Context) []*CatalogItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*CatalogItem, 0, len(r.catalog))
	for _, item := range r.catalog {
		items = append(items, item)
	}
	return items
}

// GetByID returns a single catalog item by ID
func (r *SkillsRegistry) GetByID(ctx context.Context, id string) *CatalogItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.catalog[id]
}

// Search filters catalog items by keyword in tags or description
func (r *SkillsRegistry) Search(ctx context.Context, query string) []*CatalogItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*CatalogItem
	for _, item := range r.catalog {
		// Search in tags (exact match)
		if contains(item.Tags, query) {
			results = append(results, item)
			continue
		}
		// Search in description and name (substring match, case-insensitive)
		if strings.Contains(strings.ToLower(item.Description), query) ||
			strings.Contains(strings.ToLower(item.Name), query) {
			results = append(results, item)
		}
	}
	return results
}

// GetByTarget returns catalog items that sync to a specific target
func (r *SkillsRegistry) GetByTarget(ctx context.Context, target string) []*CatalogItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*CatalogItem
	for _, item := range r.catalog {
		if contains(item.SyncTo, target) {
			results = append(results, item)
		}
	}
	return results
}

// Sync reloads the registry from disk
func (r *SkillsRegistry) Sync(ctx context.Context) error {
	r.log("Syncing skills registry...")
	return r.Load(ctx)
}

// GetStatus returns registry status
func (r *SkillsRegistry) GetStatus(ctx context.Context) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"count":       len(r.catalog),
		"last_synced": "",
	}
}

// Start initializes the registry
func (r *SkillsRegistry) Start(ctx context.Context) error {
	r.log("Starting SkillsRegistry")
	return r.Load(ctx)
}

// Stop cleans up resources
func (r *SkillsRegistry) Stop() error {
	r.log("Stopping SkillsRegistry")
	return nil
}

// CreateItem adds a new skill or context-pack
func (r *SkillsRegistry) CreateItem(ctx context.Context, item *CatalogItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicates
	if _, exists := r.catalog[item.ID]; exists {
		return fmt.Errorf("item with ID '%s' already exists", item.ID)
	}

	// Route to appropriate writer based on kind
	if item.Kind == "context-pack" {
		return r.createContextPackLocked(item)
	}
	return r.createSkillLocked(item)
}

// UpdateItem modifies an existing skill or context-pack
func (r *SkillsRegistry) UpdateItem(ctx context.Context, id string, item *CatalogItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.catalog[id]
	if !exists {
		return fmt.Errorf("item with ID '%s' not found", id)
	}

	// Preserve kind and source from original
	item.Kind = existing.Kind
	item.Source = existing.Source

	// Route to appropriate updater based on kind
	if existing.Kind == "context-pack" {
		return r.updateContextPackLocked(id, item)
	}
	return r.updateSkillLocked(id, item)
}

// DeleteItem removes a skill or context-pack
func (r *SkillsRegistry) DeleteItem(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.catalog[id]
	if !exists {
		return fmt.Errorf("item with ID '%s' not found", id)
	}

	// Route to appropriate deleter based on kind
	if existing.Kind == "context-pack" {
		return r.deleteContextPackLocked(id)
	}
	return r.deleteSkillLocked(id)
}

// createSkillLocked adds a new skill to registry.yml (must hold lock)
func (r *SkillsRegistry) createSkillLocked(item *CatalogItem) error {
	registryPath := r.registryPath()

	// Read current YAML
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("cannot read registry: %w", err)
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("cannot parse registry: %w", err)
	}

	skillsMap, ok := rawData["skills"].(map[string]interface{})
	if !ok {
		skillsMap = make(map[string]interface{})
		rawData["skills"] = skillsMap
	}

	// Add new skill
	skillEntry := r.catalogItemToYAMLSkill(item)
	skillsMap[item.ID] = skillEntry

	// Write to temp file
	tempPath := registryPath + ".tmp"
	newData, err := yaml.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("cannot marshal YAML: %w", err)
	}

	if err := os.WriteFile(tempPath, newData, 0644); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, registryPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("cannot move temp file: %w", err)
	}

	// Update in-memory catalog
	item.Source = "registry.yml"
	r.catalog[item.ID] = item
	r.log(fmt.Sprintf("Created skill: %s", item.ID))

	return nil
}

// updateSkillLocked modifies an existing skill in registry.yml (must hold lock)
func (r *SkillsRegistry) updateSkillLocked(id string, item *CatalogItem) error {
	registryPath := r.registryPath()

	// Read current YAML
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("cannot read registry: %w", err)
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("cannot parse registry: %w", err)
	}

	skillsMap, ok := rawData["skills"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("skills section not found")
	}

	// Update skill
	skillEntry := r.catalogItemToYAMLSkill(item)
	skillsMap[id] = skillEntry

	// Write to temp file
	tempPath := registryPath + ".tmp"
	newData, err := yaml.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("cannot marshal YAML: %w", err)
	}

	if err := os.WriteFile(tempPath, newData, 0644); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, registryPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("cannot move temp file: %w", err)
	}

	// Update in-memory catalog
	r.catalog[id] = item
	r.log(fmt.Sprintf("Updated skill: %s", id))

	return nil
}

// deleteSkillLocked removes a skill from registry.yml (must hold lock)
func (r *SkillsRegistry) deleteSkillLocked(id string) error {
	registryPath := r.registryPath()

	// Read current YAML
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("cannot read registry: %w", err)
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("cannot parse registry: %w", err)
	}

	skillsMap, ok := rawData["skills"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("skills section not found")
	}

	// Delete skill
	delete(skillsMap, id)

	// Write to temp file
	tempPath := registryPath + ".tmp"
	newData, err := yaml.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("cannot marshal YAML: %w", err)
	}

	if err := os.WriteFile(tempPath, newData, 0644); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, registryPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("cannot move temp file: %w", err)
	}

	// Update in-memory catalog
	delete(r.catalog, id)
	r.log(fmt.Sprintf("Deleted skill: %s", id))

	return nil
}

// createContextPackLocked adds a new context-pack to dynamic-registry.tsv (must hold lock)
func (r *SkillsRegistry) createContextPackLocked(item *CatalogItem) error {
	tsvPath := r.contextPacksPath()

	// Read current TSV
	data, err := os.ReadFile(tsvPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot read TSV: %w", err)
	}

	// Parse existing lines
	var lines []string
	if len(data) > 0 {
		lines = strings.Split(string(data), "\n")
	}

	// Create new line
	tagsStr := strings.Join(item.Tags, ",")
	newLine := fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
		item.ID,
		item.Name,
		tagsStr,
		item.Path,
		item.Description,
	)

	// Append to lines
	lines = append(lines, newLine)

	// Write to temp file
	tempPath := tsvPath + ".tmp"
	content := strings.Join(lines, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, tsvPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("cannot move temp file: %w", err)
	}

	// Update in-memory catalog
	item.Source = "dynamic-registry.tsv"
	r.catalog[item.ID] = item
	r.log(fmt.Sprintf("Created context-pack: %s", item.ID))

	return nil
}

// updateContextPackLocked modifies an existing context-pack in dynamic-registry.tsv (must hold lock)
func (r *SkillsRegistry) updateContextPackLocked(id string, item *CatalogItem) error {
	tsvPath := r.contextPacksPath()

	// Read current TSV
	data, err := os.ReadFile(tsvPath)
	if err != nil {
		return fmt.Errorf("cannot read TSV: %w", err)
	}

	// Parse existing lines
	lines := strings.Split(string(data), "\n")
	var updatedLines []string
	found := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), id+"\t") {
			// Replace this line
			tagsStr := strings.Join(item.Tags, ",")
			newLine := fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
				item.ID,
				item.Name,
				tagsStr,
				item.Path,
				item.Description,
			)
			updatedLines = append(updatedLines, newLine)
			found = true
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !found {
		return fmt.Errorf("context-pack '%s' not found in TSV", id)
	}

	// Write to temp file
	tempPath := tsvPath + ".tmp"
	content := strings.Join(updatedLines, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, tsvPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("cannot move temp file: %w", err)
	}

	// Update in-memory catalog
	r.catalog[id] = item
	r.log(fmt.Sprintf("Updated context-pack: %s", id))

	return nil
}

// deleteContextPackLocked removes a context-pack from dynamic-registry.tsv (must hold lock)
func (r *SkillsRegistry) deleteContextPackLocked(id string) error {
	tsvPath := r.contextPacksPath()

	// Read current TSV
	data, err := os.ReadFile(tsvPath)
	if err != nil {
		return fmt.Errorf("cannot read TSV: %w", err)
	}

	// Parse existing lines
	lines := strings.Split(string(data), "\n")
	var updatedLines []string
	found := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), id+"\t") {
			// Skip this line (delete it)
			found = true
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !found {
		return fmt.Errorf("context-pack '%s' not found in TSV", id)
	}

	// Write to temp file
	tempPath := tsvPath + ".tmp"
	content := strings.Join(updatedLines, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, tsvPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("cannot move temp file: %w", err)
	}

	// Update in-memory catalog
	delete(r.catalog, id)
	r.log(fmt.Sprintf("Deleted context-pack: %s", id))

	return nil
}

// catalogItemToYAMLSkill converts a CatalogItem to YAML skill map format
func (r *SkillsRegistry) catalogItemToYAMLSkill(item *CatalogItem) map[string]interface{} {
	m := make(map[string]interface{})

	if item.Name != "" {
		m["name"] = item.Name
	}
	if item.Version != "" {
		m["version"] = item.Version
	}
	if item.Type != "" {
		m["type"] = item.Type
	}
	if item.Description != "" {
		m["description"] = item.Description
	}
	if item.File != "" {
		m["file"] = item.File
	}
	if len(item.SyncTo) > 0 {
		m["sync-to"] = item.SyncTo
	}
	if item.Category != "" {
		m["category"] = item.Category
	}
	if item.SourceType != "" {
		m["source_type"] = item.SourceType
	}
	if item.SourceURI != "" {
		m["source_uri"] = item.SourceURI
	}
	if item.SourceVariant != "" {
		m["source_variant"] = item.SourceVariant
	}
	if item.ArtifactPath != "" {
		m["artifact_path"] = item.ArtifactPath
	}
	if len(item.Tags) > 0 {
		m["tags"] = item.Tags
	}
	if len(item.Requires) > 0 {
		m["requires"] = item.Requires
	}
	m["maintained"] = item.Maintained

	return m
}

// ValidateSyncStatus checks if the filesystem is synchronized with the registry
// Returns (synced bool, orphans []string, missing []string)
func (r *SkillsRegistry) ValidateSyncStatus(ctx context.Context) (bool, []string, []string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orphans := r.getOrphansLocked()
	missing := r.getMissingLocked()

	// Ensure nil slices are returned as empty arrays for JSON encoding
	if orphans == nil {
		orphans = []string{}
	}
	if missing == nil {
		missing = []string{}
	}

	synced := len(orphans) == 0 && len(missing) == 0
	return synced, orphans, missing
}

// getOrphansLocked finds filesystem directories not in registry (must hold read lock)
func (r *SkillsRegistry) getOrphansLocked() []string {
	skillsPath := r.skillSourcesDir()

	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		return []string{}
	}

	var orphans []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		// Skip special directories
		if dirName == "." || dirName == ".." || dirName == ".git" ||
			dirName == ".gitignore" || dirName == "__pycache__" ||
			dirName == "manifests" ||
			strings.HasPrefix(dirName, ".") {
			continue
		}

		// Check if this directory is in the catalog
		if _, exists := r.catalog[dirName]; !exists {
			orphans = append(orphans, dirName)
		}
	}

	return orphans
}

// getMissingLocked finds registry items whose directories don't exist (must hold read lock)
func (r *SkillsRegistry) getMissingLocked() []string {
	var missing []string

	for id := range r.catalog {
		skillPath := filepath.Join(r.skillSourcesDir(), id)
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			missing = append(missing, id)
		}
	}

	return missing
}

func (r *SkillsRegistry) log(msg string) {
	select {
	case r.logCh <- "[SkillsRegistry] " + msg:
	default:
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
