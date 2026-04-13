package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/reeinharrrd/brain/core/mcp"
)

// MCPManager manages MCP server lifecycle using the core connection manager.
type MCPManager struct {
	connMgr  *mcp.ConnectionManager
	registry *mcp.MCPRegistry
	proxy    *mcp.MCPProxy
	logCh    chan string
	mu       sync.RWMutex
	tools    map[string][]mcp.MCPTool // serverID -> known tools (fallback when server can't list)
}

// NewMCPManager creates a new MCP manager and starts the official servers that have binaries available.
func NewMCPManager(logCh chan string) *MCPManager {
	connMgr := mcp.NewConnectionManager()
	reg := mcp.NewMCPRegistry()

	m := &MCPManager{
		connMgr:  connMgr,
		registry: reg,
		proxy:    mcp.NewMCPProxy(connMgr, reg),
		logCh:    logCh,
		tools:    make(map[string][]mcp.MCPTool),
	}

	// Register all official server configs in the registry
	for _, cfg := range mcp.OfficialServers() {
		if err := reg.Register(context.Background(), cfg); err != nil {
			logCh <- fmt.Sprintf("[MCP] Failed to register %s: %v", cfg.ID, err)
		}
	}

	// Seed fallback tool definitions from the official package
	m.seedFallbackTools()

	// Start only the servers with available binaries (stdio transport, no external deps)
	startable := []string{"brain-mcp-filesystem", "brain-mcp-git", "brain-mcp-terminal"}
	for _, cfg := range mcp.OfficialServers() {
		if containsStr(startable, cfg.Command) {
			go func(c mcp.MCPServerConfig) {
				inst, err := connMgr.Start(context.Background(), c)
				if err != nil {
					logCh <- fmt.Sprintf("[MCP] Failed to start %s: %v", c.ID, err)
				} else {
					logCh <- fmt.Sprintf("[MCP] Started %s v%s (%d tools)", inst.Config.Name, inst.Config.Version, len(inst.Tools))
				}
			}(cfg)
		}
	}

	return m
}

// seedFallbackTools populates known tools from the official definitions in case the server
// fails to respond to tools/list (e.g. binary not on PATH).
func (m *MCPManager) seedFallbackTools() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tools["brain-filesystem"] = mcp.OfficialFilesystemTools()
	m.tools["brain-git"] = mcp.OfficialGitTools()
	m.tools["brain-github"] = mcp.OfficialGithubTools()
	m.tools["brain-terminal"] = mcp.OfficialTerminalTools()
	m.tools["brain-knowledge"] = mcp.OfficialKnowledgeTools()
	m.tools["brain-context"] = mcp.OfficialContextTools()
}

// CallTool proxies a tool call to the appropriate MCP server.
func (m *MCPManager) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*mcp.ToolResult, error) {
	return m.proxy.CallTool(ctx, mcp.ToolCall{
		ServerID:  serverID,
		ToolName:  toolName,
		Arguments: arguments,
	})
}

// ListServers returns all registered MCP servers with their current status.
func (m *MCPManager) ListServers(ctx context.Context) []*mcp.MCPServer {
	// Get running instances
	instances := m.connMgr.ListInstances(ctx)

	// Merge with all registered configs to show stopped servers too
	allConfigs := m.registry.List(ctx)

	// Build a map of running instances by ID
	instMap := make(map[string]*mcp.MCPServer)
	for _, inst := range instances {
		instMap[inst.Config.ID] = inst
	}

	result := make([]*mcp.MCPServer, 0, len(allConfigs))
	for _, cfg := range allConfigs {
		if inst, ok := instMap[cfg.ID]; ok {
			result = append(result, inst)
		} else {
			// Server is registered but not running
			result = append(result, &mcp.MCPServer{
				Config: cfg,
				Status: mcp.StatusStopped,
			})
		}
	}

	return result
}

// ListTools returns tools for a specific server. It first tries the running instance's
// tool list, falling back to the seeded definitions.
func (m *MCPManager) ListTools(ctx context.Context, serverID string) ([]mcp.MCPTool, error) {
	inst, err := m.connMgr.GetInstance(ctx, serverID)
	if err != nil {
		// Fallback to seeded tools
		m.mu.RLock()
		defer m.mu.RUnlock()
		if tools, ok := m.tools[serverID]; ok {
			return tools, nil
		}
		return nil, fmt.Errorf("server %q not found", serverID)
	}
	return inst.Tools, nil
}

// StartServer starts a specific MCP server from the official catalog.
func (m *MCPManager) StartServer(ctx context.Context, serverID string) error {
	for _, cfg := range mcp.OfficialServers() {
		if cfg.ID == serverID {
			_, err := m.connMgr.Start(ctx, cfg)
			if err != nil {
				return fmt.Errorf("failed to start server %q: %w", serverID, err)
			}
			m.log(fmt.Sprintf("started %s", serverID))
			return nil
		}
	}
	return fmt.Errorf("server %q not found in official catalog", serverID)
}

// StopServer stops a specific MCP server.
func (m *MCPManager) StopServer(ctx context.Context, serverID string) error {
	if err := m.connMgr.Stop(ctx, serverID); err != nil {
		return fmt.Errorf("failed to stop server %q: %w", serverID, err)
	}
	m.log(fmt.Sprintf("stopped %s", serverID))
	return nil
}

// GetStatus returns the status of all MCP servers.
func (m *MCPManager) GetStatus(ctx context.Context) map[string]interface{} {
	servers := m.ListServers(ctx)

	serverStatus := make([]map[string]interface{}, 0, len(servers))
	for _, s := range servers {
		entry := map[string]interface{}{
			"id":      s.Config.ID,
			"name":    s.Config.Name,
			"version": s.Config.Version,
			"status":  string(s.Status),
			"tools":   len(s.Tools),
		}
		if s.Error != "" {
			entry["error"] = s.Error
		}
		if s.StartedAt.IsZero() {
			entry["started_at"] = nil
		} else {
			entry["started_at"] = s.StartedAt
		}
		serverStatus = append(serverStatus, entry)
	}

	return map[string]interface{}{
		"servers": serverStatus,
		"total":   len(servers),
	}
}

// Sync is a no-op for the real manager (servers started in constructor).
func (m *MCPManager) Sync(ctx context.Context) error {
	m.log("syncing registry")
	return nil
}

// Start initializes the MCP manager (called by the daemon's Start).
func (m *MCPManager) Start(ctx context.Context) error {
	m.log("starting")
	return nil // servers already started in constructor
}

// Stop stops all running MCP servers.
func (m *MCPManager) Stop() error {
	m.log("stopping")

	instances := m.connMgr.ListInstances(context.Background())
	for _, inst := range instances {
		if inst.Status == mcp.StatusRunning {
			if err := m.connMgr.Stop(context.Background(), inst.Config.ID); err != nil {
				m.log(fmt.Sprintf("failed to stop %s: %v", inst.Config.ID, err))
			} else {
				m.log(fmt.Sprintf("stopped %s", inst.Config.ID))
			}
		}
	}
	return nil
}

// Proxy returns the underlying MCP proxy for direct access.
func (m *MCPManager) Proxy() *mcp.MCPProxy {
	return m.proxy
}

// ConnectionManager returns the underlying connection manager for direct access.
func (m *MCPManager) ConnectionManager() *mcp.ConnectionManager {
	return m.connMgr
}

func (m *MCPManager) log(msg string) {
	select {
	case m.logCh <- "[MCPManager] " + msg:
	default:
		fmt.Println("[MCPManager]", msg)
	}
}

// containsStr checks if a string slice contains a value.
func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
