package workers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	coreartifacts "github.com/reeinharrrd/brain/core/artifacts"
	"gopkg.in/yaml.v3"
)

// CatalogItem represents a skill or context-pack
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

// SkillsWorker synchronizes skills to various targets
type SkillsWorker struct {
}

// Sync deploys skills and context-packs from daemon registry to target
func (w *SkillsWorker) Sync(brainRoot string, targetDir string, logger chan<- string, getCatalogFunc func() []*CatalogItem) error {
	logger <- fmt.Sprintf("[SkillsWorker] Starting sync from catalog to %s", targetDir)

	// Get current catalog from daemon
	catalog := getCatalogFunc()
	if len(catalog) == 0 {
		logger <- "[SkillsWorker] Empty catalog, nothing to sync"
		return nil
	}

	// Prepare output for CLI (skills.json)
	if strings.Contains(targetDir, "config") || strings.HasSuffix(targetDir, ".json") {
		jsonPath := targetDir
		if strings.HasSuffix(targetDir, "/") {
			jsonPath = filepath.Join(targetDir, "skills.json")
		}

		os.MkdirAll(filepath.Dir(jsonPath), 0755)

		// Prepare catalog as map for JSON output
		skillsMap := make(map[string]interface{})
		for _, item := range catalog {
			// For CLI compatibility, use legacy format but include new fields
			itemMap := map[string]interface{}{
				"id":          item.ID,
				"name":        item.Name,
				"kind":        item.Kind,
				"scope":       item.Scope,
				"description": item.Description,
				"source":      item.Source,
				"maintained":  item.Maintained,
			}
			if item.SourceType != "" {
				itemMap["source_type"] = item.SourceType
			}
			if item.SourceURI != "" {
				itemMap["source_uri"] = item.SourceURI
			}
			if item.SourceVariant != "" {
				itemMap["source_variant"] = item.SourceVariant
			}
			if item.ArtifactPath != "" {
				itemMap["artifact_path"] = item.ArtifactPath
			}

			// Add legacy fields if present
			if item.Type != "" {
				itemMap["type"] = item.Type
			}
			if item.Version != "" {
				itemMap["version"] = item.Version
			}
			if item.File != "" {
				itemMap["file"] = item.File
			}
			if len(item.SyncTo) > 0 {
				itemMap["sync_to"] = item.SyncTo
			}
			if len(item.Requires) > 0 {
				itemMap["requires"] = item.Requires
			}
			if item.Category != "" {
				itemMap["category"] = item.Category
			}
			if len(item.Tags) > 0 {
				itemMap["tags"] = item.Tags
			}

			skillsMap[item.ID] = itemMap
		}

		normalized := map[string]interface{}{"skills": skillsMap}
		jsonData, err := json.MarshalIndent(normalized, "", "  ")
		if err != nil {
			logger <- fmt.Sprintf("[SkillsWorker] Error marshaling JSON: %v", err)
			return err
		}

		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			logger <- fmt.Sprintf("[SkillsWorker] Error writing JSON: %v", err)
			return err
		}

		logger <- fmt.Sprintf("[SkillsWorker] Deployed %d items to %s", len(catalog), jsonPath)
	}

	logger <- "[SkillsWorker] Sync complete"
	return nil
}

// Skill represents a skill from registry (legacy, kept for backward compat)
type Skill struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Type        string   `json:"type" yaml:"type"`
	Description string   `json:"description" yaml:"description"`
	Tags        []string `json:"tags" yaml:"tags"`
	File        string   `json:"file" yaml:"file"`
	SyncTo      []string `json:"sync-to" yaml:"sync-to"`
	Requires    []string `json:"requires" yaml:"requires"`
	Maintained  bool     `json:"maintained" yaml:"maintained"`
	Category    string   `json:"category" yaml:"category"`
}

// MCPsWorker synchronizes MCPs to various targets
type MCPsWorker struct {
}

// MCP represents an MCP from registry
type MCP struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Type        string   `json:"type" yaml:"type"`
	Package     string   `json:"package" yaml:"package"`
	Command     string   `json:"command" yaml:"command"`
	Description string   `json:"description" yaml:"description"`
	Required    bool     `json:"required" yaml:"required"`
	Profiles    []string `json:"profiles" yaml:"profile"`
	Features    []string `json:"features" yaml:"features"`
	EnvRequired []string `json:"env_required" yaml:"env_required"`
	Setup       string   `json:"setup" yaml:"setup"`
	Notes       string   `json:"notes" yaml:"notes"`
}

// Sync deploys MCPs to a target
func (w *MCPsWorker) Sync(registryPath string, targetDir string, logger chan<- string) error {
	logger <- fmt.Sprintf("[MCPsWorker] Starting sync from %s to %s", registryPath, targetDir)

	// Read registry
	data, err := os.ReadFile(registryPath)
	if err != nil {
		logger <- fmt.Sprintf("[MCPsWorker] Error reading registry: %v", err)
		return err
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		logger <- fmt.Sprintf("[MCPsWorker] Error parsing YAML: %v", err)
		return err
	}

	// Extract MCPs
	mcpsMap := make(map[string]interface{})
	if mcpsData, ok := rawData["mcps"].(map[string]interface{}); ok {
		mcpsMap = mcpsData
	}

	// Convert to JSON for CLI config
	if strings.Contains(targetDir, "config") || strings.HasSuffix(targetDir, ".json") {
		jsonPath := targetDir
		if strings.HasSuffix(targetDir, "/") {
			jsonPath = filepath.Join(targetDir, "mcps.json")
		}

		os.MkdirAll(filepath.Dir(jsonPath), 0755)

		jsonData, err := json.MarshalIndent(mcpsMap, "", "  ")
		if err != nil {
			logger <- fmt.Sprintf("[MCPsWorker] Error marshaling JSON: %v", err)
			return err
		}

		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			logger <- fmt.Sprintf("[MCPsWorker] Error writing JSON: %v", err)
			return err
		}

		logger <- fmt.Sprintf("[MCPsWorker] Deployed to %s", jsonPath)
	}

	logger <- "[MCPsWorker] Sync complete"
	return nil
}

// AgentsWorker synchronizes agents to various targets
type AgentsWorker struct {
}

func canonicalAgentPromptPath(agentPath string) string {
	return coreartifacts.CanonicalizePath(agentPath, "agents")
}

// Agent represents an agent
type Agent struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Version     string   `json:"version" yaml:"version"`
	Model       string   `json:"model" yaml:"model"`
	Temperature float64  `json:"temperature" yaml:"temperature"`
	PromptFile  string   `json:"prompt_file" yaml:"prompt_file"`
	Content     string   `json:"content" yaml:"content"`
	Tags        []string `json:"tags" yaml:"tags"`
	Maintained  bool     `json:"maintained" yaml:"maintained"`
}

// Sync deploys agents to a target
func (w *AgentsWorker) Sync(agentsDir string, targetDir string, logger chan<- string) error {
	logger <- fmt.Sprintf("[AgentsWorker] Starting sync from %s to %s", agentsDir, targetDir)

	agents, err := loadAgentsFromDirectory(agentsDir)
	if err != nil {
		return err
	}

	// Convert to JSON for CLI config
	if strings.Contains(targetDir, "config") || strings.HasSuffix(targetDir, ".json") {
		jsonPath := targetDir
		if strings.HasSuffix(targetDir, "/") {
			jsonPath = filepath.Join(targetDir, "agents.json")
		}

		os.MkdirAll(filepath.Dir(jsonPath), 0755)

		normalized := map[string]interface{}{"agents": agents}
		jsonData, err := json.MarshalIndent(normalized, "", "  ")
		if err != nil {
			logger <- fmt.Sprintf("[AgentsWorker] Error marshaling JSON: %v", err)
			return err
		}

		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			logger <- fmt.Sprintf("[AgentsWorker] Error writing JSON: %v", err)
			return err
		}

		logger <- fmt.Sprintf("[AgentsWorker] Deployed to %s", jsonPath)
	}

	logger <- "[AgentsWorker] Sync complete"
	return nil
}

func loadAgentsFromDirectory(agentsDir string) ([]Agent, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	agents := make([]Agent, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		agentPath := filepath.Join(agentsDir, entry.Name())
		agent, err := parseAgentMarkdown(agentPath)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func parseAgentMarkdown(agentPath string) (Agent, error) {
	data, err := os.ReadFile(agentPath)
	if err != nil {
		return Agent{}, fmt.Errorf("failed to read agent markdown %s: %w", agentPath, err)
	}

	content := string(data)
	agent := Agent{
		PromptFile:  canonicalAgentPromptPath(agentPath),
		Maintained:  true,
		Version:     "1.0.0",
		Model:       "claude-opus",
		Temperature: 0.5,
	}
	agent.ID = strings.TrimSuffix(filepath.Base(agentPath), filepath.Ext(agentPath))
	agent.Name = strings.Title(agent.ID)

	if frontMatter, body, ok := splitFrontMatter(content); ok {
		if v, exists := frontMatter["name"].(string); exists && v != "" {
			agent.Name = v
		}
		if v, exists := frontMatter["description"].(string); exists && v != "" {
			agent.Description = v
		}
		if v, exists := frontMatter["version"].(string); exists && v != "" {
			agent.Version = v
		}
		if v, exists := frontMatter["model"].(string); exists && v != "" {
			agent.Model = v
		}
		if v, exists := frontMatter["temperature"].(float64); exists {
			agent.Temperature = v
		}
		if v, exists := frontMatter["maintained"].(bool); exists {
			agent.Maintained = v
		}
		if tags, exists := frontMatter["tags"].([]interface{}); exists {
			for _, tag := range tags {
				if t, ok := tag.(string); ok {
					agent.Tags = append(agent.Tags, t)
				}
			}
		}
		agent.Content = strings.TrimSpace(body)
	} else {
		agent.Content = strings.TrimSpace(content)
	}

	return agent, nil
}

func splitFrontMatter(content string) (map[string]interface{}, string, bool) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return nil, "", false
	}

	parts := strings.SplitN(trimmed, "\n---\n", 2)
	if len(parts) != 2 {
		return nil, "", false
	}

	frontMatterRaw := strings.TrimPrefix(parts[0], "---")
	frontMatterRaw = strings.TrimSpace(frontMatterRaw)

	var frontMatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontMatterRaw), &frontMatter); err != nil {
		return nil, "", false
	}

	return frontMatter, parts[1], true
}
