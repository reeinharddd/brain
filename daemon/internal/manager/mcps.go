package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/reeinharrrd/brain/daemon/internal/environment"
	"gopkg.in/yaml.v3"
)

// MCP represents a single Model Context Protocol server
type MCP struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Type        string   `yaml:"type" json:"type"` // stdio, sse, etc
	Package     string   `yaml:"package" json:"package"`
	Command     string   `yaml:"command" json:"command"`
	Description string   `yaml:"description" json:"description"`
	Required    bool     `yaml:"required" json:"required"`
	Visibility   string   `yaml:"visibility" json:"visibility"`
	Profiles    []string `yaml:"profile" json:"profiles"` // which profiles include this MCP
	Features    []string `yaml:"features" json:"features"`
	EnvRequired []string `yaml:"env_required" json:"env_required"`
	SyncTo      []string `yaml:"sync-to" json:"sync_to"`
	Setup       string   `yaml:"setup" json:"setup"`
	Notes       string   `yaml:"notes" json:"notes"`
}

// MCPsManager manages all MCPs
type MCPsManager struct {
	mu        sync.RWMutex
	mcps      map[string]*MCP
	brainRoot string
	environment string
	logCh     chan string
}

// NewMCPsManager creates a new MCPs manager
func NewMCPsManager(brainRoot string, environment string, logCh chan string) *MCPsManager {
	return &MCPsManager{
		mcps:      make(map[string]*MCP),
		brainRoot: brainRoot,
		environment: environment,
		logCh:     logCh,
	}
}

// Load reads MCPs from registry.yml
func (r *MCPsManager) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	registryPath := filepath.Join(r.brainRoot, "mcp", "registry.yml")
	r.log(fmt.Sprintf("Loading MCPs from %s", registryPath))

	data, err := os.ReadFile(registryPath)
	if err != nil {
		r.log(fmt.Sprintf("Error reading registry: %v", err))
		return fmt.Errorf("cannot read MCPs registry: %w", err)
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		r.log(fmt.Sprintf("Error parsing YAML: %v", err))
		return fmt.Errorf("cannot parse MCPs registry: %w", err)
	}

	r.mcps = make(map[string]*MCP)

	mcpsData, ok := rawData["mcps"].(map[string]interface{})
	if !ok {
		r.log("No 'mcps' section in registry")
		return nil
	}

	for id, mcpData := range mcpsData {
		mcpMap, ok := mcpData.(map[string]interface{})
		if !ok {
			r.log(fmt.Sprintf("Invalid MCP data for %s", id))
			continue
		}

		mcp := &MCP{ID: id}

		if v, ok := mcpMap["id"].(string); ok {
			mcp.ID = v
		}
		if v, ok := mcpMap["name"].(string); ok {
			mcp.Name = v
		}
		if v, ok := mcpMap["version"].(string); ok {
			mcp.Version = v
		}
		if v, ok := mcpMap["type"].(string); ok {
			mcp.Type = v
		}
		if v, ok := mcpMap["package"].(string); ok {
			mcp.Package = v
		}
		if v, ok := mcpMap["command"].(string); ok {
			mcp.Command = v
		}
		if v, ok := mcpMap["description"].(string); ok {
			mcp.Description = v
		}
		if v, ok := mcpMap["setup"].(string); ok {
			mcp.Setup = v
		}
		if v, ok := mcpMap["notes"].(string); ok {
			mcp.Notes = v
		}
		if v, ok := mcpMap["required"].(bool); ok {
			mcp.Required = v
		}
		if v, ok := mcpMap["visibility"].(string); ok {
			mcp.Visibility = strings.ToLower(strings.TrimSpace(v))
		}
		if mcp.Visibility == "" {
			mcp.Visibility = environment.ProdSafe
		}

		// Parse profiles
		if profilesIface, ok := mcpMap["profile"].([]interface{}); ok {
			for _, p := range profilesIface {
				if prof, ok := p.(string); ok {
					mcp.Profiles = append(mcp.Profiles, prof)
				}
			}
		}

		// Parse features
		if featuresIface, ok := mcpMap["features"].([]interface{}); ok {
			for _, f := range featuresIface {
				if feat, ok := f.(string); ok {
					mcp.Features = append(mcp.Features, feat)
				}
			}
		}

		// Parse env_required
		if envIface, ok := mcpMap["env_required"].([]interface{}); ok {
			for _, e := range envIface {
				if env, ok := e.(string); ok {
					mcp.EnvRequired = append(mcp.EnvRequired, env)
				}
			}
		}

		// Parse sync-to
		if syncIface, ok := mcpMap["sync-to"].([]interface{}); ok {
			for _, s := range syncIface {
				if st, ok := s.(string); ok {
					mcp.SyncTo = append(mcp.SyncTo, st)
				}
			}
		}

		if !environment.AllowsVisibility(mcp.Visibility, r.environment) {
			r.log(fmt.Sprintf("Skipping %s in %s environment due to visibility=%s", mcp.ID, r.environment, mcp.Visibility))
			continue
		}

		r.mcps[mcp.ID] = mcp
		r.log(fmt.Sprintf("Loaded MCP: %s (v%s)", mcp.ID, mcp.Version))
	}

	r.log(fmt.Sprintf("Loaded %d MCPs", len(r.mcps)))
	return nil
}

// GetAll returns all MCPs
func (r *MCPsManager) GetAll(ctx context.Context) []*MCP {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mcps := make([]*MCP, 0, len(r.mcps))
	for _, mcp := range r.mcps {
		mcps = append(mcps, mcp)
	}
	return mcps
}

// GetByID returns a single MCP
func (r *MCPsManager) GetByID(ctx context.Context, id string) *MCP {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mcps[id]
}

// Search filters MCPs by keyword
func (r *MCPsManager) Search(ctx context.Context, query string) []*MCP {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*MCP
	for _, mcp := range r.mcps {
		if contains(mcp.Features, query) || contains([]string{mcp.Description, mcp.Name}, query) {
			results = append(results, mcp)
		}
	}
	return results
}

// GetByProfile returns MCPs that belong to a specific profile
func (r *MCPsManager) GetByProfile(ctx context.Context, profile string) []*MCP {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*MCP
	for _, mcp := range r.mcps {
		if contains(mcp.Profiles, profile) {
			results = append(results, mcp)
		}
	}
	return results
}

// GetRequired returns only required MCPs
func (r *MCPsManager) GetRequired(ctx context.Context) []*MCP {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*MCP
	for _, mcp := range r.mcps {
		if mcp.Required {
			results = append(results, mcp)
		}
	}
	return results
}

// Sync reloads the registry
func (r *MCPsManager) Sync(ctx context.Context) error {
	r.log("Syncing MCPs registry...")
	return r.Load(ctx)
}

// GetStatus returns registry status
func (r *MCPsManager) GetStatus(ctx context.Context) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"count":        len(r.mcps),
		"environment":  r.environment,
		"last_synced":  "",
	}
}

// Start initializes the registry
func (r *MCPsManager) Start(ctx context.Context) error {
	r.log("Starting MCPsManager")
	return r.Load(ctx)
}

// Stop cleans up resources
func (r *MCPsManager) Stop() error {
	r.log("Stopping MCPsManager")
	return nil
}

func (r *MCPsManager) log(msg string) {
	select {
	case r.logCh <- "[MCPsManager] " + msg:
	default:
	}
}
