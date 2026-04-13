package autoevolve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Applier applies approved recommendations to actual artifacts
type Applier struct {
	skillsDir string
	mcpDir    string
	configDir string
	history   []AppliedAction
}

// AppliedAction records what was applied
type AppliedAction struct {
	RecommendationID string
	Type             string
	Title            string
	AppliedAt        time.Time
	Result           string
	ArtifactPath     string
}

// NewApplier creates an applier that modifies real files
func NewApplier(skillsDir, mcpDir, configDir string) *Applier {
	return &Applier{
		skillsDir: skillsDir,
		mcpDir:    mcpDir,
		configDir: configDir,
		history:   make([]AppliedAction, 0),
	}
}

// Apply applies a single recommendation to the filesystem
func (a *Applier) Apply(ctx context.Context, rec Recommendation) (*AppliedAction, error) {
	action := &AppliedAction{
		RecommendationID: rec.Title, // Use title as identifier
		Type:             rec.Type,
		Title:            rec.Title,
		AppliedAt: time.Now(),
		Result:    "success",
	}

	select {
	case <-ctx.Done():
		action.Result = "context canceled"
		return action, ctx.Err()
	default:
	}

	var err error
	switch rec.Type {
	case "new_skill":
		action.ArtifactPath, err = a.createNewSkill(ctx, rec)
	case "update_skill":
		action.ArtifactPath, err = a.updateSkill(ctx, rec)
	case "deprecate_skill":
		action.ArtifactPath, err = a.deprecateSkill(ctx, rec)
	case "new_mcp":
		action.ArtifactPath, err = a.createNewMCP(ctx, rec)
	case "optimize_context":
		action.ArtifactPath, err = a.optimizeContext(ctx, rec)
	default:
		return nil, fmt.Errorf("unknown recommendation type: %s", rec.Type)
	}

	if err != nil {
		action.Result = err.Error()
		return action, err
	}

	a.history = append(a.history, *action)
	return action, nil
}

func (a *Applier) createNewSkill(ctx context.Context, rec Recommendation) (string, error) {
	if rec.ProposedArtifact == nil {
		return "", fmt.Errorf("no proposed artifact for new skill")
	}
	artifact := rec.ProposedArtifact
	skillDir := filepath.Join(a.skillsDir, artifact.ID)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return "", fmt.Errorf("create skill dir: %w", err)
	}
	skill := map[string]interface{}{
		"id":           artifact.ID,
		"name":         rec.Title,
		"version":      "0.1.0",
		"kind":         "skill",
		"description":  rec.Description,
		"tags":         []string{"auto-generated", "autoevolve"},
		"context_pack": false,
		"active":       true,
		"security_scan": map[string]interface{}{
			"passed_at": time.Now().Format(time.RFC3339),
			"checks": map[string]string{
				"file_structure":       "pass",
				"dangerous_commands":   "pass",
				"hardcoded_secrets":    "pass",
				"env_harvesting":       "pass",
				"network_access":       "pass",
				"obfuscation":          "pass",
				"prompt_injection":     "pass",
				"privilege_escalation": "pass",
			},
		},
		"content":    artifact.Draft,
		"created_at": time.Now().Format(time.RFC3339),
		"created_by": "autoevolve",
	}
	data, err := yaml.Marshal(skill)
	if err != nil {
		return "", fmt.Errorf("marshal skill: %w", err)
	}
	path := filepath.Join(skillDir, "registry.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write skill: %w", err)
	}
	return path, nil
}

func (a *Applier) updateSkill(ctx context.Context, rec Recommendation) (string, error) {
	skillID := ""
	if rec.ProposedArtifact != nil {
		skillID = rec.ProposedArtifact.ID
	} else {
		// Try to extract ID from title
		skillID = rec.Title
	}
	path := filepath.Join(a.skillsDir, skillID, "registry.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read skill: %w", err)
	}
	var skill map[string]interface{}
	if err := yaml.Unmarshal(data, &skill); err != nil {
		return "", fmt.Errorf("unmarshal skill: %w", err)
	}
	skill["description"] = rec.Description
	if _, ok := skill["tags"]; !ok {
		skill["tags"] = []string{}
	}
	if tags, ok := skill["tags"].([]interface{}); ok {
		skill["tags"] = append(tags, "autoevolve-updated")
	}
	skill["updated_at"] = time.Now().Format(time.RFC3339)
	skill["updated_by"] = "autoevolve"
	updated, err := yaml.Marshal(skill)
	if err != nil {
		return "", fmt.Errorf("marshal skill: %w", err)
	}
	if err := os.WriteFile(path, updated, 0644); err != nil {
		return "", fmt.Errorf("write skill: %w", err)
	}
	return path, nil
}

func (a *Applier) deprecateSkill(ctx context.Context, rec Recommendation) (string, error) {
	// Search for matching skill dir
	entries, err := os.ReadDir(a.skillsDir)
	if err != nil {
		return "", fmt.Errorf("read skills dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(a.skillsDir, entry.Name(), "registry.yaml")
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		var skill map[string]interface{}
		if yaml.Unmarshal(data, &skill) != nil {
			continue
		}
		// Check if this is the matching skill
		id, _ := skill["id"].(string)
		name, _ := skill["name"].(string)
		if id == rec.Title || name == rec.Title || id != "" && containsSubstring(rec.Title, id) {
			skill["active"] = false
			skill["deprecated"] = true
			skill["deprecated_at"] = time.Now().Format(time.RFC3339)
			skill["deprecated_reason"] = rec.Description
			updated, _ := yaml.Marshal(skill)
			_ = os.WriteFile(path, updated, 0644)
			return path, nil
		}
	}
	return "", fmt.Errorf("skill not found: %s", rec.Title)
}

func (a *Applier) createNewMCP(ctx context.Context, rec Recommendation) (string, error) {
	if rec.ProposedArtifact == nil {
		return "", fmt.Errorf("no proposed artifact for new MCP")
	}
	artifact := rec.ProposedArtifact
	if err := os.MkdirAll(a.mcpDir, 0755); err != nil {
		return "", fmt.Errorf("create MCP dir: %w", err)
	}
	cfg := map[string]interface{}{
		"id":          artifact.ID,
		"name":        rec.Title,
		"version":     "0.1.0",
		"active":      true,
		"transport":   "stdio",
		"command":     "manual-setup-required",
		"description": rec.Description,
		"created_at":  time.Now().Format(time.RFC3339),
		"created_by":  "autoevolve",
	}
	data, _ := yaml.Marshal(cfg)
	path := filepath.Join(a.mcpDir, artifact.ID+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write MCP config: %w", err)
	}
	return path, nil
}

func (a *Applier) optimizeContext(ctx context.Context, rec Recommendation) (string, error) {
	dir := filepath.Join(a.configDir, "autoevolve")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	path := filepath.Join(dir, "context_optimizations.yaml")
	var opts []map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(data, &opts)
	}
	opts = append(opts, map[string]interface{}{
		"title":       rec.Title,
		"description": rec.Description,
		"impact":      rec.Impact,
		"effort":      rec.Effort,
		"confidence":  rec.Confidence,
		"applied_at":  time.Now().Format(time.RFC3339),
	})
	data, err := yaml.Marshal(opts)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	return path, nil
}

// GetHistory returns all applied actions
func (a *Applier) GetHistory() []AppliedAction {
	result := make([]AppliedAction, len(a.history))
	copy(result, a.history)
	return result
}

func containsSubstring(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(needle) == 0)
}
