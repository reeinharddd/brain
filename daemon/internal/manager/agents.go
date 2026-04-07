package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Agent represents a single AI agent
type Agent struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Version     string   `yaml:"version" json:"version"`
	Model       string   `yaml:"model" json:"model"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	PromptFile  string   `yaml:"prompt_file" json:"prompt_file"`
	Tags        []string `yaml:"tags" json:"tags"`
	Maintained  bool     `yaml:"maintained" json:"maintained"`
}

// AgentsManager manages all agents
type AgentsManager struct {
	mu        sync.RWMutex
	agents    map[string]*Agent
	brainRoot string
	logCh     chan string
}

// NewAgentsManager creates a new agents manager
func NewAgentsManager(brainRoot string, logCh chan string) *AgentsManager {
	return &AgentsManager{
		agents:    make(map[string]*Agent),
		brainRoot: brainRoot,
		logCh:     logCh,
	}
}

// Load reads agents from the agents/ folder and manifest
func (r *AgentsManager) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.log("Loading agents from agents/ folder")

	agentsDir := filepath.Join(r.brainRoot, "agents")
	r.agents = make(map[string]*Agent)

	// List all .md files in agents/
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		r.log(fmt.Sprintf("Error reading agents directory: %v", err))
		return fmt.Errorf("cannot read agents directory: %w", err)
	}

	// Try to load agents manifest if it exists
	manifestPath := filepath.Join(r.brainRoot, "agents", "manifest.yml")
	var manifestData map[string]interface{}
	if data, err := os.ReadFile(manifestPath); err == nil {
		if err := yaml.Unmarshal(data, &manifestData); err == nil {
			if agentsIface, ok := manifestData["agents"].([]interface{}); ok {
				for _, agentIface := range agentsIface {
					if agentMap, ok := agentIface.(map[string]interface{}); ok {
						agent := r.parseAgent(agentMap)
						if agent != nil {
							r.agents[agent.ID] = agent
						}
					}
				}
			}
		}
	}

	// If no manifest, scan .md files
	if len(r.agents) == 0 {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			id := strings.TrimSuffix(entry.Name(), ".md")
			agent := &Agent{
				ID:          id,
				Name:        strings.Title(id),
				PromptFile:  filepath.Join("agents", entry.Name()),
				Version:     "1.0.0",
				Model:       "claude-opus",
				Temperature: 0.5,
				Maintained:  true,
			}

			r.agents[id] = agent
			r.log(fmt.Sprintf("Loaded agent: %s", id))
		}
	}

	r.log(fmt.Sprintf("Loaded %d agents", len(r.agents)))
	return nil
}

// parseAgent parses an agent from YAML data
func (r *AgentsManager) parseAgent(data map[string]interface{}) *Agent {
	agent := &Agent{}

	if v, ok := data["id"].(string); ok {
		agent.ID = v
	} else if v, ok := data["name"].(string); ok {
		agent.ID = strings.ToLower(strings.ReplaceAll(v, " ", "-"))
	}

	if agent.ID == "" {
		return nil
	}

	if v, ok := data["name"].(string); ok {
		agent.Name = v
	} else {
		agent.Name = strings.Title(agent.ID)
	}

	if v, ok := data["description"].(string); ok {
		agent.Description = v
	}

	if v, ok := data["version"].(string); ok {
		agent.Version = v
	} else {
		agent.Version = "1.0.0"
	}

	if v, ok := data["model"].(string); ok {
		agent.Model = v
	} else {
		agent.Model = "claude-opus"
	}

	if v, ok := data["temperature"].(float64); ok {
		agent.Temperature = v
	} else {
		agent.Temperature = 0.5
	}

	if v, ok := data["prompt_file"].(string); ok {
		agent.PromptFile = v
	} else if v, ok := data["prompt"].(string); ok {
		agent.PromptFile = v
	}

	if v, ok := data["maintained"].(bool); ok {
		agent.Maintained = v
	} else {
		agent.Maintained = true
	}

	// Parse tags
	if tagsIface, ok := data["tags"].([]interface{}); ok {
		for _, tag := range tagsIface {
			if t, ok := tag.(string); ok {
				agent.Tags = append(agent.Tags, t)
			}
		}
	}

	return agent
}

// GetAll returns all agents
func (r *AgentsManager) GetAll(ctx context.Context) []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetByID returns a single agent
func (r *AgentsManager) GetByID(ctx context.Context, id string) *Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agents[id]
}

// Search filters agents by keyword
func (r *AgentsManager) Search(ctx context.Context, query string) []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Agent
	for _, agent := range r.agents {
		if contains(agent.Tags, query) || contains([]string{agent.Description, agent.Name}, query) {
			results = append(results, agent)
		}
	}
	return results
}

// Sync reloads the agents list
func (r *AgentsManager) Sync(ctx context.Context) error {
	r.log("Syncing agents...")
	return r.Load(ctx)
}

// GetStatus returns manager status
func (r *AgentsManager) GetStatus(ctx context.Context) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"count":       len(r.agents),
		"last_synced": "",
	}
}

// Start initializes the manager
func (r *AgentsManager) Start(ctx context.Context) error {
	r.log("Starting AgentsManager")
	return r.Load(ctx)
}

// Stop cleans up resources
func (r *AgentsManager) Stop() error {
	r.log("Stopping AgentsManager")
	return nil
}

func (r *AgentsManager) log(msg string) {
	select {
	case r.logCh <- "[AgentsManager] " + msg:
	default:
	}
}
