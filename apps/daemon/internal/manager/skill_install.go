package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	skillInstallWorkDirPrefix = "brain-skill-install-"
	skillInstallDefaultOwner  = "user:brain"
	skillInstallDefaultTrust  = "verified"
	skillInstallDefaultMode   = "local"
)

var remoteSkillLine = regexp.MustCompile(`^\s*[◻☑■✔✖●]\s*(.+?)(?:\s+\(|$)`)

var defaultSkillSyncTargets = []string{
	"cli", "vscode", "cursor", "claude-code", "codex-cli", "continue", "cline",
	"gemini-cli", "opencode", "zed", "jetbrains", "windsurf", "neovim", "aider",
}

// SkillInstallRequest describes a source install/import request.
type SkillInstallRequest struct {
	Source          string   `json:"source"`
	SourceType      string   `json:"source_type,omitempty"`
	Scope           string   `json:"scope,omitempty"`
	Skills          []string `json:"skills,omitempty"`
	InstallAll      bool     `json:"install_all,omitempty"`
	IncludeInternal bool     `json:"include_internal,omitempty"`
	Copy            bool     `json:"copy,omitempty"`
}

// SkillInstallVariant is a discovered installable variant.
type SkillInstallVariant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
	Internal    bool   `json:"internal,omitempty"`
}

// SkillInstallPreview summarizes what will be installed.
type SkillInstallPreview struct {
	Source            string               `json:"source"`
	SourceType        string               `json:"source_type"`
	Scope             string               `json:"scope"`
	RequiresSelection bool                 `json:"requires_selection"`
	Available         []SkillInstallVariant `json:"available"`
	Selected          []string             `json:"selected,omitempty"`
	Notes             []string             `json:"notes,omitempty"`
}

// SkillInstallResult reports the normalized items after install.
type SkillInstallResult struct {
	Preview   SkillInstallPreview `json:"preview"`
	Installed []*CatalogItem      `json:"installed"`
}

type skillInstallRecord struct {
	ID          string
	Name        string
	Description string
	Version     string
	Category    string
	Internal    bool
	Body        string
	FrontMatter map[string]interface{}
	Root        string
	Files       []string
}

// PreviewSkillInstall inspects a source and returns the available installable skills.
func (r *SkillsRegistry) PreviewSkillInstall(ctx context.Context, req SkillInstallRequest) (*SkillInstallPreview, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Source) == "" {
		return nil, fmt.Errorf("source is required")
	}

	preview := &SkillInstallPreview{
		Source:     req.Source,
		SourceType: normalizeSkillSourceType(req.Source, req.SourceType),
		Scope:      normalizeInstallScope(req.Scope),
	}

	records, notes, err := r.discoverSkillInstallCandidates(ctx, req, true)
	if err != nil {
		return nil, err
	}
	preview.Notes = append(preview.Notes, notes...)

	for _, record := range records {
		preview.Available = append(preview.Available, SkillInstallVariant{
			ID:          record.ID,
			Name:        record.Name,
			Description: record.Description,
			Path:        record.Root,
			Internal:    record.Internal,
		})
	}

	if len(req.Skills) > 0 {
		preview.Selected = append(preview.Selected, req.Skills...)
	}
	preview.RequiresSelection = len(preview.Available) > 1 && len(preview.Selected) == 0 && !req.InstallAll
	return preview, nil
}

// InstallSkill materializes the selected source into Brain's source tree and registry.
func (r *SkillsRegistry) InstallSkill(ctx context.Context, req SkillInstallRequest) (*SkillInstallResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Source) == "" {
		return nil, fmt.Errorf("source is required")
	}

	preview, err := r.PreviewSkillInstall(ctx, req)
	if err != nil {
		return nil, err
	}

	selectedNames := append([]string(nil), req.Skills...)
	if req.InstallAll {
		selectedNames = nil
	}
	if len(selectedNames) == 0 && len(preview.Available) > 1 && !req.InstallAll {
		return nil, fmt.Errorf("multiple skills were discovered; select one or more skills before installing")
	}

	records, cleanup, err := r.resolveInstallWorkspace(ctx, req, selectedNames)
	if err != nil {
		return nil, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	installed := make([]*CatalogItem, 0, len(records))
	createdIDs := make([]string, 0, len(records))
	for _, record := range records {
		item, err := r.persistImportedSkill(ctx, req, record)
		if err != nil {
			for i := len(createdIDs) - 1; i >= 0; i-- {
				_ = r.deleteImportedSkillCleanup(createdIDs[i])
			}
			return nil, err
		}
		installed = append(installed, item)
		createdIDs = append(createdIDs, item.ID)
	}

	return &SkillInstallResult{
		Preview:   *preview,
		Installed: installed,
	}, nil
}

func normalizeSkillSourceType(source, hint string) string {
	hint = strings.ToLower(strings.TrimSpace(hint))
	if hint != "" && hint != "auto" {
		return hint
	}

	source = strings.TrimSpace(source)
	if isLocalSkillSource(source) {
		return "local"
	}
	if strings.HasPrefix(source, "npx:") || strings.HasPrefix(source, "npm:") {
		return "npm"
	}
	if isGitishSource(source) {
		return "git"
	}
	return "git"
}

func normalizeInstallScope(scope string) string {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		return "global"
	}
	switch scope {
	case "global", "user", "workspace", "project", "organization", "team", "local":
		return scope
	default:
		return "global"
	}
}

func isLocalSkillSource(source string) bool {
	source = strings.TrimSpace(source)
	if source == "" {
		return false
	}
	if strings.HasPrefix(source, "file://") || strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~") {
		return true
	}
	if info, err := os.Stat(expandSkillPath(source)); err == nil && info.IsDir() {
		return true
	}
	return false
}

func isGitishSource(source string) bool {
	source = strings.TrimSpace(source)
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "ssh://")
}

func expandSkillPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func discoverSkillRoots(root string) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("source root is required")
	}

	var roots []string
	seen := make(map[string]bool)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry == nil {
			return nil
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(entry.Name(), "SKILL.md") {
			parent := filepath.Dir(path)
			if !seen[parent] {
				seen[parent] = true
				roots = append(roots, parent)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(roots)
	return roots, nil
}

func (r *SkillsRegistry) discoverSkillInstallCandidates(ctx context.Context, req SkillInstallRequest, previewOnly bool) ([]skillInstallRecord, []string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		return nil, nil, fmt.Errorf("source is required")
	}

	if isLocalSkillSource(source) {
		root := source
		if strings.HasPrefix(root, "file://") {
			root = strings.TrimPrefix(root, "file://")
		}
		root = expandSkillPath(root)

		roots, err := discoverSkillRoots(root)
		if err != nil {
			return nil, nil, err
		}

		records := make([]skillInstallRecord, 0, len(roots))
		for _, skillRoot := range roots {
			record, err := parseSkillInstallRecord(skillRoot)
			if err != nil {
				return nil, nil, err
			}
			records = append(records, record)
		}

		if len(req.Skills) > 0 {
			records = filterSkillInstallRecords(records, req.Skills, req.InstallAll)
		}

		if previewOnly && len(records) == 0 {
			return nil, []string{"No SKILL.md files found in the provided local source"}, nil
		}
		return records, []string{"Local source detected"}, nil
	}

	if previewOnly {
		variants, notes, err := r.previewRemoteSkillNames(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		records := make([]skillInstallRecord, 0, len(variants))
		for _, variant := range variants {
			records = append(records, skillInstallRecord{
				ID:          variant.ID,
				Name:        variant.Name,
				Description: variant.Description,
				Internal:    variant.Internal,
			})
		}
		return records, notes, nil
	}

	workDir, err := os.MkdirTemp("", skillInstallWorkDirPrefix)
	if err != nil {
		return nil, nil, err
	}

	if _, err := r.runSkillsAddCommand(ctx, workDir, req, req.Skills); err != nil {
		_ = os.RemoveAll(workDir)
		return nil, nil, err
	}

	roots, err := discoverSkillRoots(workDir)
	if err != nil {
		_ = os.RemoveAll(workDir)
		return nil, nil, err
	}

	records := make([]skillInstallRecord, 0, len(roots))
	for _, skillRoot := range roots {
		record, err := parseSkillInstallRecord(skillRoot)
		if err != nil {
			_ = os.RemoveAll(workDir)
			return nil, nil, err
		}
		records = append(records, record)
	}

	if len(req.Skills) > 0 {
		records = filterSkillInstallRecords(records, req.Skills, req.InstallAll)
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("skills installer completed but no SKILL.md files were discovered in %s", workDir)
	}

	return records, []string{fmt.Sprintf("Installed to temporary workspace %s", workDir)}, nil
}

func (r *SkillsRegistry) previewRemoteSkillNames(ctx context.Context, req SkillInstallRequest) ([]skillInstallRecord, []string, error) {
	output, err := r.runSkillsListCommand(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	records := parseSkillsListOutput(output)

	return records, []string{"Remote source previewed via upstream skills CLI"}, nil
}

func parseSkillsListOutput(output string) []skillInstallRecord {
	lines := strings.Split(output, "\n")
	records := make([]skillInstallRecord, 0)
	seen := make(map[string]bool)

	inSkillsSection := false
	var current *skillInstallRecord
	var description []string

	flush := func() {
		if current == nil {
			return
		}
		if len(description) > 0 {
			current.Description = strings.TrimSpace(strings.Join(description, " "))
		}
		if current.ID == "" {
			current.ID = slugifySkillName(current.Name)
		}
		if current.Name == "" && current.ID != "" {
			current.Name = titleFromID(current.ID)
		}
		if current.ID != "" && !seen[current.ID] {
			seen[current.ID] = true
			records = append(records, *current)
		}
		current = nil
		description = nil
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current != nil && len(description) > 0 {
				// Blank line after description; keep collecting only until next skill name.
				continue
			}
			continue
		}

		if strings.Contains(trimmed, "Available Skills") {
			inSkillsSection = true
			continue
		}
		if !inSkillsSection {
			continue
		}
		if strings.Contains(trimmed, "Use --skill ") && strings.Contains(trimmed, "install specific skills") {
			break
		}
		if strings.HasPrefix(trimmed, "└") {
			break
		}

		if strings.HasPrefix(line, "│    ") && !strings.HasPrefix(line, "│      ") {
			// Starting a new skill entry.
			flush()
			name := strings.TrimSpace(strings.TrimPrefix(line, "│    "))
			if name == "" {
				continue
			}
			current = &skillInstallRecord{
				ID:   slugifySkillName(name),
				Name: name,
			}
			continue
		}

		if current != nil && strings.HasPrefix(line, "│      ") {
			description = append(description, strings.TrimSpace(strings.TrimPrefix(line, "│      ")))
			continue
		}

		// Ignore everything else in the CLI chrome.
	}

	flush()
	return records
}

func (r *SkillsRegistry) runSkillsListCommand(ctx context.Context, req SkillInstallRequest) (string, error) {
	args := []string{"--yes", "skills", "add", req.Source, "--list"}
	cmd := exec.CommandContext(ctx, "npx", args...)
	if req.IncludeInternal {
		cmd.Env = append(os.Environ(), "INSTALL_INTERNAL_SKILLS=1")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("skills preview failed: %w\n%s", err, string(output))
	}
	return string(output), nil
}

func (r *SkillsRegistry) runSkillsAddCommand(ctx context.Context, workDir string, req SkillInstallRequest, selected []string) (string, error) {
	args := []string{"--yes", "skills", "add", req.Source, "-a", "universal", "--copy", "-y"}
	if req.InstallAll {
		args = append(args, "--all")
	}
	if len(selected) > 0 && !req.InstallAll {
		for _, skill := range selected {
			skill = strings.TrimSpace(skill)
			if skill != "" {
				args = append(args, "--skill", skill)
			}
		}
	}

	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = workDir
	if req.IncludeInternal {
		cmd.Env = append(os.Environ(), "INSTALL_INTERNAL_SKILLS=1")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("skills install failed: %w\n%s", err, string(output))
	}
	return string(output), nil
}

func parseSkillInstallRecord(root string) (skillInstallRecord, error) {
	markdownPath := filepath.Join(root, "SKILL.md")
	data, err := os.ReadFile(markdownPath)
	if err != nil {
		return skillInstallRecord{}, fmt.Errorf("failed to read skill markdown %s: %w", markdownPath, err)
	}

	content := string(data)
	record := skillInstallRecord{
		Root:  root,
		Files: collectSkillFiles(root),
		ID:    filepath.Base(root),
	}

	if frontMatter, body, ok := splitMarkdownFrontMatter(content); ok {
		record.Body = body
		record.FrontMatter = frontMatter
		if v, ok := frontMatter["name"].(string); ok && strings.TrimSpace(v) != "" {
			record.Name = strings.TrimSpace(v)
		}
		if v, ok := frontMatter["description"].(string); ok && strings.TrimSpace(v) != "" {
			record.Description = strings.TrimSpace(v)
		}
		if v, ok := frontMatter["version"].(string); ok && strings.TrimSpace(v) != "" {
			record.Version = strings.TrimSpace(v)
		}
		if meta, ok := frontMatter["metadata"].(map[string]interface{}); ok {
			if internal, ok := meta["internal"].(bool); ok {
				record.Internal = internal
			}
		}
	}

	if record.Name == "" {
		record.Name = titleFromID(record.ID)
	}
	if record.Description == "" {
		record.Description = record.Name
	}
	if record.Version == "" {
		record.Version = "1.0.0"
	}
	if strings.TrimSpace(record.Category) == "" {
		record.Category = "imported"
	}

	if frontMatter, body, ok := splitMarkdownFrontMatter(content); ok {
		if v, ok := frontMatter["category"].(string); ok && strings.TrimSpace(v) != "" {
			record.Category = strings.TrimSpace(v)
		}
		record.Body = body
	}

	return record, nil
}

func collectSkillFiles(root string) []string {
	var files []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(files)
	return files
}

func filterSkillInstallRecords(records []skillInstallRecord, selected []string, installAll bool) []skillInstallRecord {
	if installAll || len(selected) == 0 {
		return records
	}
	selectedSet := make(map[string]bool, len(selected))
	for _, value := range selected {
		value = strings.TrimSpace(value)
		if value != "" {
			selectedSet[value] = true
		}
	}
	if len(selectedSet) == 0 {
		return records
	}
	filtered := make([]skillInstallRecord, 0, len(records))
	for _, record := range records {
		if selectedSet[record.ID] || selectedSet[record.Name] {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func (r *SkillsRegistry) resolveInstallWorkspace(ctx context.Context, req SkillInstallRequest, selected []string) ([]skillInstallRecord, func(), error) {
	_ = ctx
	if isLocalSkillSource(req.Source) {
		root := req.Source
		if strings.HasPrefix(root, "file://") {
			root = strings.TrimPrefix(root, "file://")
		}
		root = expandSkillPath(root)
		roots, err := discoverSkillRoots(root)
		if err != nil {
			return nil, nil, err
		}
		records := make([]skillInstallRecord, 0, len(roots))
		for _, skillRoot := range roots {
			record, err := parseSkillInstallRecord(skillRoot)
			if err != nil {
				return nil, nil, err
			}
			records = append(records, record)
		}
		records = filterSkillInstallRecords(records, selected, req.InstallAll)
		return records, nil, nil
	}

	workDir, err := os.MkdirTemp("", skillInstallWorkDirPrefix)
	if err != nil {
		return nil, nil, err
	}

	if _, err := r.runSkillsAddCommand(ctx, workDir, req, selected); err != nil {
		_ = os.RemoveAll(workDir)
		return nil, nil, err
	}

	roots, err := discoverSkillRoots(workDir)
	if err != nil {
		_ = os.RemoveAll(workDir)
		return nil, nil, err
	}

	records := make([]skillInstallRecord, 0, len(roots))
	for _, skillRoot := range roots {
		record, err := parseSkillInstallRecord(skillRoot)
		if err != nil {
			_ = os.RemoveAll(workDir)
			return nil, nil, err
		}
		records = append(records, record)
	}
	records = filterSkillInstallRecords(records, selected, req.InstallAll)
	cleanup := func() { _ = os.RemoveAll(workDir) }
	return records, cleanup, nil
}

func (r *SkillsRegistry) persistImportedSkill(ctx context.Context, req SkillInstallRequest, record skillInstallRecord) (*CatalogItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sourceRoot := r.skillSourcesDir()
	if err := os.MkdirAll(sourceRoot, 0755); err != nil {
		return nil, err
	}

	id := record.ID
	if id == "" {
		id = slugifySkillName(record.Name)
	}
	if id == "" {
		return nil, fmt.Errorf("could not determine skill ID for %q", record.Name)
	}

	destRoot := filepath.Join(sourceRoot, id)
	if err := os.RemoveAll(destRoot); err != nil {
		return nil, err
	}
	if err := copySkillDirectory(record.Root, destRoot); err != nil {
		return nil, err
	}

	artifactPath := filepath.Join(destRoot, "artifact.yml")
	if err := writeSkillArtifactManifest(artifactPath, req, record, destRoot); err != nil {
		return nil, err
	}

	item := &CatalogItem{
		ID:            id,
		Name:          record.Name,
		Kind:          "skill",
		Scope:         normalizeInstallScope(req.Scope),
		Description:   record.Description,
		Tags:          mergeSkillTags(record, req),
		Path:          filepath.ToSlash(filepath.Join(destRoot, "SKILL.md")),
		Version:       record.Version,
		Maintained:    true,
		Source:        "registry.yml",
		SourceType:    normalizeSkillSourceType(req.Source, req.SourceType),
		SourceURI:     req.Source,
		SourceVariant:  record.Name,
		ArtifactPath:   filepath.ToSlash(artifactPath),
		Type:          "external",
		File:          filepath.ToSlash(filepath.Join(destRoot, "SKILL.md")),
		SyncTo:        append([]string(nil), defaultSkillSyncTargets...),
		Category:      record.Category,
	}
	if item.SourceType == "local" && strings.TrimSpace(req.SourceType) != "" {
		item.SourceType = strings.ToLower(strings.TrimSpace(req.SourceType))
	}

	if record.Internal {
		item.Type = "internal"
	}

	if len(record.Files) > 0 {
		item.Requires = nil
	}

	if err := r.CreateItem(ctx, item); err != nil {
		_ = os.RemoveAll(destRoot)
		_ = os.Remove(artifactPath)
		return nil, err
	}

	return item, nil
}

func mergeSkillTags(record skillInstallRecord, req SkillInstallRequest) []string {
	tags := make([]string, 0, 4)
	if record.Category != "" {
		tags = append(tags, record.Category)
	}
	if record.Internal {
		tags = append(tags, "internal")
	}
	if sourceType := normalizeSkillSourceType(req.Source, req.SourceType); sourceType != "" {
		tags = append(tags, sourceType)
	}
	return uniqueStrings(tags)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func copySkillDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

func writeSkillArtifactManifest(path string, req SkillInstallRequest, record skillInstallRecord, destRoot string) error {
	files := collectSkillFiles(destRoot)
	manifest := map[string]interface{}{
		"apiVersion": "brain/v1",
		"kind":       "skill",
		"id":         record.ID,
		"displayName": record.Name,
		"description": record.Description,
		"scope":      normalizeInstallScope(req.Scope),
		"owner":      skillInstallDefaultOwner,
		"visibility": scopeToVisibility(req.Scope),
		"lifecycle": map[string]interface{}{
			"state": "active",
		},
		"source": map[string]interface{}{
			"type": normalizeSkillSourceType(req.Source, req.SourceType),
			"uri":  req.Source,
		},
		"sync": map[string]interface{}{
			"mode":         skillInstallDefaultMode,
			"cloudEnabled": true,
		},
		"security": map[string]interface{}{
			"trust": skillInstallDefaultTrust,
			"permissions": map[string]interface{}{
				"fs":   "read",
				"net":  "restricted",
				"exec": "none",
			},
		},
		"content": map[string]interface{}{
			"entrypoint": "SKILL.md",
			"files":      files,
		},
		"skill": map[string]interface{}{
			"category": normalizeSkillCategory(record.Category),
			"compatibility": map[string]interface{}{
				"minCapabilityTier":     "standard",
				"preferredCapabilityTier": "powerful",
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func normalizeSkillCategory(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "imported"
	}
	return value
}

func scopeToVisibility(scope string) string {
	scope = normalizeInstallScope(scope)
	switch scope {
	case "project", "workspace":
		return "workspace"
	case "team", "organization":
		return "organization"
	default:
		return "private"
	}
}

func splitMarkdownFrontMatter(content string) (map[string]interface{}, string, bool) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, content, false
	}
	trimmed := strings.TrimPrefix(content, "---\n")
	trimmed = strings.TrimPrefix(trimmed, "---\r\n")
	end := strings.Index(trimmed, "\n---\n")
	if end < 0 {
		return nil, content, false
	}
	frontRaw := trimmed[:end]
	body := trimmed[end+5:]
	var frontMatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontRaw), &frontMatter); err != nil {
		return nil, content, false
	}
	return frontMatter, body, true
}

func slugifySkillName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "'", "")
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`-+`).ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func titleFromID(id string) string {
	parts := strings.FieldsFunc(id, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func (r *SkillsRegistry) deleteImportedSkillCleanup(id string) error {
	sourceRoot := filepath.Join(r.skillSourcesDir(), id)
	if err := os.RemoveAll(sourceRoot); err != nil {
		return err
	}
	manifestPath := filepath.Join(r.brainRoot, "artifacts", "skills", "manifests", id+".artifact.json")
	_ = os.Remove(manifestPath)
	return nil
}

// Helper for tests and CLI previews.
func (r *SkillsRegistry) listSkillInstallPreviewJSON(ctx context.Context, req SkillInstallRequest) (string, error) {
	preview, err := r.PreviewSkillInstall(ctx, req)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(preview, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyReaderToFile(dst string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}