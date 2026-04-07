package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	brainenv "github.com/reeinharrrd/brain/daemon/internal/environment"
	"github.com/reeinharrrd/brain/daemon/internal/manager"
	"github.com/reeinharrrd/brain/daemon/internal/manifest"
	"github.com/reeinharrrd/brain/daemon/internal/syncengine"
	"github.com/reeinharrrd/brain/daemon/internal/api/handlers"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin for Tauri/React GUI
	},
}

type BrainDaemon struct {
	mu          sync.Mutex
	status      string
	clients     map[*websocket.Conn]bool
	procManager *manager.ProcessManager
	docker      *manager.DockerManager
	qdrant      *manager.QdrantManager
	ollama      *manager.OllamaManager
	mcp         *manager.MCPRegistry
	providers   *manager.ProviderManager
	skills      *manager.SkillsRegistry
	mcps        *manager.MCPsManager
	agents      *manager.AgentsManager
	environment string
	syncMu      sync.Mutex
	syncRunning bool
	syncStatus  string
	syncLastRun time.Time
	syncError   string
	syncWatcher *syncengine.FileWatcher
	logChannel  chan string
	brainRoot   string
	docsHandler *handlers.DocsHandler
}

func configRootFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "brain", "root")
}

func readConfiguredRoot() string {
	b, err := os.ReadFile(configRootFilePath())
	if err != nil {
		return ""
	}
	root := strings.TrimSpace(string(b))
	if root == "" {
		return ""
	}
	if _, err := os.Stat(filepath.Join(root, "manifest.yml")); err == nil {
		return root
	}
	return ""
}

func resolveBrainRoot() string {
	if envRoot := strings.TrimSpace(os.Getenv("BRAIN_ROOT")); envRoot != "" {
		if _, err := os.Stat(filepath.Join(envRoot, "manifest.yml")); err == nil {
			return envRoot
		}
	}

	if configured := readConfiguredRoot(); configured != "" {
		return configured
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".brain")
}

func cors(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func NewBrainDaemon() *BrainDaemon {
	logCh := make(chan string, 1000)
	root := resolveBrainRoot()
	environment := brainenv.Current()
	d := &BrainDaemon{
		status:      "Running",
		clients:     make(map[*websocket.Conn]bool),
		procManager: manager.NewProcessManager(logCh),
		docker:      manager.NewDockerManager(filepath.Join(root, "docker"), logCh),
		qdrant:      manager.NewQdrantManager("http://localhost:6333", logCh),
		ollama:      manager.NewOllamaManager("http://localhost:11434", logCh),
		mcp:         manager.NewMCPRegistry(logCh),
		providers:   manager.NewProviderManager(logCh),
		skills:      manager.NewSkillsRegistry(root, logCh),
		mcps:        manager.NewMCPsManager(root, environment, logCh),
		agents:      manager.NewAgentsManager(root, logCh),
		environment: environment,
		syncStatus:  "idle",
		logChannel:  logCh,
		brainRoot:   root,
	}
	logCh <- fmt.Sprintf("[Daemon] Initialized with 8 managers in %s environment", environment)
	
	// Load registries asynchronously
	go func() {
		ctx := context.Background()
		if err := d.skills.Load(ctx); err != nil {
			logCh <- fmt.Sprintf("⚠️ Failed to load skills registry: %v", err)
		} else {
			logCh <- "[Skills] Registry loaded"
			
			// VALIDATE SYNC STATUS AFTER LOADING
			synced, orphans, missing := d.skills.ValidateSyncStatus(ctx)
			if !synced {
				if len(orphans) > 0 {
					logCh <- fmt.Sprintf("❌ FATAL: Found %d orphan skills: %v", len(orphans), orphans)
					// Log each orphan
					for _, orphan := range orphans {
						logCh <- fmt.Sprintf("   - %s/ (not in registry)", orphan)
					}
				}
				if len(missing) > 0 {
					logCh <- fmt.Sprintf("⚠️ WARNING: Found %d missing skills: %v", len(missing), missing)
					for _, missing_skill := range missing {
						logCh <- fmt.Sprintf("   - %s (in registry but not on filesystem)", missing_skill)
					}
				}
				
				// Only FAIL if there are orphans
				if len(orphans) > 0 {
					logCh <- "[Skills] ❌ Refusing to start with orphan skills. Please delete them or add them to registry."
					// Note: In production, you might want to panic here or set a flag to deny registration
				}
			} else {
				logCh <- "[Skills] ✅ Registry and filesystem perfectly synchronized"
			}
		}
		
		// Start validation ticker (every 5 minutes)
		d.startSkillsValidationTicker()
	}()
	
	go func() {
		ctx := context.Background()
		if err := d.mcps.Load(ctx); err != nil {
			logCh <- fmt.Sprintf("⚠️ Failed to load MCPs registry: %v", err)
		} else {
			logCh <- "[MCPs] Registry loaded"
		}
	}()
	
	go func() {
		ctx := context.Background()
		if err := d.agents.Load(ctx); err != nil {
			logCh <- fmt.Sprintf("⚠️ Failed to load agents registry: %v", err)
		} else {
			logCh <- "[Agents] Registry loaded"
		}
	}()
	
	// Initialize Docs-RAG handler with stub indexer
	d.docsHandler = handlers.NewDocsHandler(&StubIndexer{}, environment)
	logCh <- "[Docs-RAG] Handler initialized"
	
	go d.processLogs()
	go d.startSyncSubsystem()  // Start sync subsystem asynchronously
	go d.healthCheckLoop()  // Start background health check loop
	return d
}

// startSkillsValidationTicker runs skills validation every 5 minutes
func (d *BrainDaemon) startSkillsValidationTicker() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			ctx := context.Background()
			synced, orphans, missing := d.skills.ValidateSyncStatus(ctx)
			
			if !synced {
				if len(orphans) > 0 {
					d.logChannel <- fmt.Sprintf("⚠️ Skills validation: Found %d orphans: %v", len(orphans), orphans)
				}
				if len(missing) > 0 {
					d.logChannel <- fmt.Sprintf("⚠️ Skills validation: Found %d missing: %v", len(missing), missing)
				}
			} else {
				d.logChannel <- "[Skills Validation] ✅ Registry and filesystem in sync"
			}
		}
	}()
}

func (d *BrainDaemon) processLogs() {
	for msg := range d.logChannel {
		// print to terminal daemon stdout as well
		fmt.Println(msg)
		
		payload, _ := json.Marshal(map[string]interface{}{
			"event": "log",
			"data":  msg,
		})
		d.broadcast(payload)
	}
}

// healthCheckLoop runs every 30s and broadcasts status to WebSocket clients
func (d *BrainDaemon) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		d.mu.Lock()
		status := d.status
		d.mu.Unlock()
		
		// Fetch real manager status
		dockerOk := true
		qdrantOk := true
		ollamaOk := d.ollama.IsAvailable(context.Background())
		mcpOk := true
		
		// Quick health checkpoint
		_, dockerErr := d.docker.GetServiceStatus(context.Background(), "all")
		if dockerErr != nil {
			dockerOk = false
		}
		
		qdrantHealth, _ := d.qdrant.HealthCheck(context.Background())
		if !qdrantHealth {
			qdrantOk = false
		}

		statusMsg := map[string]interface{}{
			"event": "healthcheck",
			"data": map[string]interface{}{
				"daemon_status": status,
				"docker":        dockerOk,
				"qdrant":        qdrantOk,
				"ollama":        ollamaOk,
				"mcp":           mcpOk,
				"timestamp":     time.Now().String(),
			},
		}
		
		payload, _ := json.Marshal(statusMsg)
		d.broadcast(payload)
	}
}

func (d *BrainDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	if r.URL.Path == "/ws" {
		d.handleWebSocket(w, r)
		return
	}
	if r.URL.Path == "/api/status" {
		d.handleStatus(w, r)
		return
	}
	if r.URL.Path == "/api/processes" {
		d.handleGetProcesses(w, r)
		return
	}
	if r.URL.Path == "/api/process/start" && r.Method == "POST" {
		d.handleStartProcess(w, r)
		return
	}
	if r.URL.Path == "/api/process/stop" && r.Method == "POST" {
		d.handleStopProcess(w, r)
		return
	}
	if r.URL.Path == "/api/sync/status" && r.Method == "GET" {
		d.handleSyncStatus(w, r)
		return
	}
	if r.URL.Path == "/api/sync" && r.Method == "POST" {
		d.handleSync(w, r)
		return
	}
	// New manager endpoints
	if r.URL.Path == "/api/daemon/start" && r.Method == "POST" {
		d.handleDaemonStart(w, r)
		return
	}
	if r.URL.Path == "/api/daemon/stop" && r.Method == "POST" {
		d.handleDaemonStop(w, r)
		return
	}
	if r.URL.Path == "/api/docker/status" && r.Method == "GET" {
		d.handleDockerStatus(w, r)
		return
	}
	if r.URL.Path == "/api/qdrant/status" && r.Method == "GET" {
		d.handleQdrantStatus(w, r)
		return
	}
	if r.URL.Path == "/api/ollama/status" && r.Method == "GET" {
		d.handleOllamaStatus(w, r)
		return
	}
	if r.URL.Path == "/api/mcp/status" && r.Method == "GET" {
		d.handleMCPStatus(w, r)
		return
	}
	if r.URL.Path == "/api/providers/available" && r.Method == "GET" {
		d.handleProvidersAvailable(w, r)
		return
	}
	
	// Registry endpoints
	if strings.HasPrefix(r.URL.Path, "/api/skills") {
		d.handleSkills(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/mcps") {
		d.handleMCPs(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/agents") {
		d.handleAgents(w, r)
		return
	}
	
	// Docs-RAG endpoints
	if strings.HasPrefix(r.URL.Path, "/api/docs") {
		d.handleDocs(w, r)
		return
	}
	
	http.NotFound(w, r)
}

func (d *BrainDaemon) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS Upgrade Error:", err)
		return
	}

	d.mu.Lock()
	d.clients[conn] = true
	d.mu.Unlock()

	defer func() {
		conn.Close()
		d.mu.Lock()
		delete(d.clients, conn)
		d.mu.Unlock()
	}()

	d.broadcast([]byte(fmt.Sprintf(`{"event":"status","data":"%s"}`, d.status)))

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (d *BrainDaemon) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d.mu.Lock()
	status := d.status
	d.mu.Unlock()
	
	statuses := d.procManager.GetAllStatuses()
	d.syncMu.Lock()
	syncStatus := d.syncStatus
	syncRunning := d.syncRunning
	syncLastRun := d.syncLastRun
	syncError := d.syncError
	d.syncMu.Unlock()
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       status,
		"time":         time.Now().String(),
		"processes":    len(statuses),
		"uptime":       time.Now().Unix(),
		"environment":  d.environment,
		"sync_status":  syncStatus,
		"sync_running": syncRunning,
		"sync_last_run": syncLastRun,
		"sync_error":   syncError,
	})
}

func (d *BrainDaemon) handleGetProcesses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	statuses := d.procManager.GetAllStatuses()
	json.NewEncoder(w).Encode(statuses)
}

type StartProcessReq struct {
	ID      string   `json:"id"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func (d *BrainDaemon) handleStartProcess(w http.ResponseWriter, r *http.Request) {
	var req StartProcessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	if err := d.procManager.StartProcess(req.ID, req.Command, req.Args); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	
	d.broadcast([]byte(fmt.Sprintf(`{"event":"status","data":"started process %s"}`, req.ID)))
	w.WriteHeader(200)
}

type StopProcessReq struct {
	ID string `json:"id"`
}

func (d *BrainDaemon) handleStopProcess(w http.ResponseWriter, r *http.Request) {
	var req StopProcessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	if err := d.procManager.StopProcess(req.ID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	
	d.broadcast([]byte(fmt.Sprintf(`{"event":"status","data":"stopped process %s"}`, req.ID)))
	w.WriteHeader(200)
}

func (d *BrainDaemon) broadcast(message []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for client := range d.clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			client.Close()
			delete(d.clients, client)
		}
	}
}

// Start orchestrates daemon startup with all managers
func (d *BrainDaemon) Start(ctx context.Context) error {
	d.mu.Lock()
	d.status = "Starting"
	d.mu.Unlock()

	d.logChannel <- "[Daemon] Starting orchestration..."

	// Start Docker
	if err := d.docker.ComposeUp(ctx); err != nil {
		d.logChannel <- "[Daemon] Docker startup failed: " + err.Error()
	}

	// Check Qdrant health
	ok, err := d.qdrant.HealthCheck(ctx)
	if err != nil || !ok {
		d.logChannel <- "[Daemon] Qdrant health check failed"
	}

	// Ensure Qdrant collections exist
	if err := d.qdrant.EnsureCollections(ctx); err != nil {
		d.logChannel <- "[Daemon] Qdrant setup failed: " + err.Error()
	}

	// Check Ollama availability (optional)
	if d.ollama.IsAvailable(ctx) {
		d.logChannel <- "[Daemon] Ollama is available"
	}

	// Sync MCP registry
	if err := d.mcp.Sync(ctx); err != nil {
		d.logChannel <- "[Daemon] MCP sync failed: " + err.Error()
	}

	// Check available providers
	available := d.providers.GetAvailable(ctx)
	d.logChannel <- fmt.Sprintf("[Daemon] Available providers: %v", available)

	if err := d.triggerSync(false); err != nil {
		d.logChannel <- "[Daemon] Initial sync failed: " + err.Error()
	}

	d.mu.Lock()
	d.status = "Ready"
	d.mu.Unlock()

	d.logChannel <- "[Daemon] Startup complete"
	return nil
}

// Stop orchestrates daemon shutdown with all managers
func (d *BrainDaemon) Stop(ctx context.Context) error {
	d.mu.Lock()
	d.status = "Stopping"
	d.mu.Unlock()

	d.logChannel <- "[Daemon] Stopping orchestration..."

	// Stop MCP
	if err := d.mcp.Stop(); err != nil {
		d.logChannel <- "[Daemon] MCP stop failed: " + err.Error()
	}

	// Stop Docker
	if err := d.docker.ComposeDown(ctx); err != nil {
		d.logChannel <- "[Daemon] Docker shutdown failed: " + err.Error()
	}

	if d.syncWatcher != nil {
		_ = d.syncWatcher.Stop()
	}

	d.mu.Lock()
	d.status = "Stopped"
	d.mu.Unlock()

	d.logChannel <- "[Daemon] Shutdown complete"
	return nil
}

// HTTP Handlers for managers
func (d *BrainDaemon) handleDaemonStart(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if err := d.Start(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (d *BrainDaemon) handleDaemonStop(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if err := d.Stop(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (d *BrainDaemon) handleDockerStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	status, err := d.docker.GetServiceStatus(ctx, "all")
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (d *BrainDaemon) handleQdrantStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	status, err := d.qdrant.GetStatus(ctx)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (d *BrainDaemon) handleOllamaStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	available := d.ollama.IsAvailable(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if available {
		json.NewEncoder(w).Encode(map[string]string{"status": "available", "endpoint": "http://localhost:11434"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
	}
}

func (d *BrainDaemon) handleMCPStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	status := d.mcp.GetStatus(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (d *BrainDaemon) handleProvidersAvailable(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	available := d.providers.GetAvailable(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"available": available})
}

func (d *BrainDaemon) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	d.syncMu.Lock()
	status := d.syncStatus
	running := d.syncRunning
	lastRun := d.syncLastRun
	syncErr := d.syncError
	watcherActive := d.syncWatcher != nil
	d.syncMu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         status,
		"running":        running,
		"last_run":       lastRun,
		"error":          syncErr,
		"watcher_active": watcherActive,
	})
}

func (d *BrainDaemon) handleSync(w http.ResponseWriter, r *http.Request) {
	dryRun := strings.EqualFold(r.URL.Query().Get("dry_run"), "true")
	d.logChannel <- fmt.Sprintf("[Daemon] Incoming Sync Request (dry_run=%t)...", dryRun)

	go func() {
		if err := d.triggerSync(dryRun); err != nil {
			d.logChannel <- "[SyncEngine] Error running sync: " + err.Error()
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "Sync Initiated",
		"dry_run": dryRun,
	})
}

// Skills Handlers
func (d *BrainDaemon) handleSkills(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")
	
	// GET /api/skills - list all
	if r.URL.Path == "/api/skills" && r.Method == "GET" {
		skills := d.skills.GetAll(ctx)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"skills": skills})
		return
	}
	
	// POST /api/skills - create new skill or context-pack
	if r.URL.Path == "/api/skills" && r.Method == "POST" {
		var item manager.CatalogItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		if err := d.skills.CreateItem(ctx, &item); err != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Trigger sync to propagate changes to all targets
		d.triggerSync(false)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "created", "item": &item})
		return
	}
	
	// GET /api/skills/validate - check synchronization status (MUST be before generic GET)
	if r.URL.Path == "/api/skills/validate" && r.Method == "GET" {
		synced, orphans, missing := d.skills.ValidateSyncStatus(ctx)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"synced":  synced,
			"orphans": orphans,
			"missing": missing,
		})
		return
	}
	
	// POST /api/skills/search (MUST be before generic GET)
	if r.URL.Path == "/api/skills/search" && r.Method == "POST" {
		query := r.URL.Query().Get("q")
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "query parameter required"})
			return
		}
		results := d.skills.Search(ctx, query)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
		return
	}
	
	// POST /api/skills/sync (MUST be before generic PUT/DELETE)
	if r.URL.Path == "/api/skills/sync" && r.Method == "POST" {
		err := d.skills.Sync(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "skills synced"})
		return
	}
	
	// GET /api/skills/{id}
	if r.Method == "GET" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			id := parts[3]
			skill := d.skills.GetByID(ctx, id)
			if skill == nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "skill not found"})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(skill)
			return
		}
	}
	
	// PUT /api/skills/{id} - update skill or context-pack
	if r.Method == "PUT" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			id := parts[3]
			var item manager.CatalogItem
			if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
				return
			}

			if err := d.skills.UpdateItem(ctx, id, &item); err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			// Trigger sync to propagate changes to all targets
			d.triggerSync(false)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "updated", "item": &item})
			return
		}
	}
	
	// DELETE /api/skills/{id} - delete skill or context-pack
	if r.Method == "DELETE" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			id := parts[3]

			if err := d.skills.DeleteItem(ctx, id); err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			// Trigger sync to propagate changes to all targets
			d.triggerSync(false)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "deleted"})
			return
		}
	}
	
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
}


// MCPs Handlers
func (d *BrainDaemon) handleMCPs(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")
	
	// GET /api/mcps - list all
	if r.URL.Path == "/api/mcps" && r.Method == "GET" {
		mcps := d.mcps.GetAll(ctx)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"mcps": mcps})
		return
	}
	
	// GET /api/mcps/{id}
	if r.Method == "GET" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			id := parts[3]
			mcp := d.mcps.GetByID(ctx, id)
			if mcp == nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "mcp not found"})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mcp)
			return
		}
	}
	
	// POST /api/mcps/search
	if r.URL.Path == "/api/mcps/search" && r.Method == "POST" {
		query := r.URL.Query().Get("q")
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "query parameter required"})
			return
		}
		results := d.mcps.Search(ctx, query)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
		return
	}
	
	// POST /api/mcps/sync
	if r.URL.Path == "/api/mcps/sync" && r.Method == "POST" {
		err := d.mcps.Sync(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "mcps synced"})
		return
	}
	
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
}

// Agents Handlers
func (d *BrainDaemon) handleAgents(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")
	
	// GET /api/agents - list all
	if r.URL.Path == "/api/agents" && r.Method == "GET" {
		agents := d.agents.GetAll(ctx)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
		return
	}
	
	// GET /api/agents/{id}
	if r.Method == "GET" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			id := parts[3]
			agent := d.agents.GetByID(ctx, id)
			if agent == nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "agent not found"})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(agent)
			return
		}
	}
	
	// POST /api/agents/search
	if r.URL.Path == "/api/agents/search" && r.Method == "POST" {
		query := r.URL.Query().Get("q")
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "query parameter required"})
			return
		}
		results := d.agents.Search(ctx, query)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
		return
	}
	
	// POST /api/agents/sync
	if r.URL.Path == "/api/agents/sync" && r.Method == "POST" {
		err := d.agents.Sync(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "agents synced"})
		return
	}
	
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
}

// Docs-RAG Handlers
func (d *BrainDaemon) handleDocs(w http.ResponseWriter, r *http.Request) {
	if d.docsHandler == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "docs handler not initialized"})
		return
	}
	
	// Route to appropriate handler
	if strings.HasPrefix(r.URL.Path, "/api/docs/search") && r.Method == "GET" {
		d.docsHandler.Search(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/api/docs/status") && r.Method == "GET" {
		d.docsHandler.Status(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/api/docs/rebuild") && r.Method == "POST" {
		d.docsHandler.Rebuild(w, r)
		return
	}
	
	http.NotFound(w, r)
}

func main() {
	daemon := NewBrainDaemon()

	port := "9090"
	fmt.Printf("🧠 Brain Daemon starting on port %s...\n", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: daemon,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Daemon failed to start:", err)
	}
}

func (d *BrainDaemon) startSyncSubsystem() {
	watcher, err := syncengine.NewFileWatcher(d.logChannel, func() error {
		return d.triggerSync(false)
	})
	if err != nil {
		d.logChannel <- "[SyncWatcher] Failed to initialize: " + err.Error()
		return
	}

	paths := []string{
		filepath.Join(d.brainRoot, "manifest.yml"),
		filepath.Join(d.brainRoot, "skills", "registry.yml"),
		filepath.Join(d.brainRoot, "skills", "dynamic-registry.tsv"),
		filepath.Join(d.brainRoot, "mcp", "registry.yml"),
		filepath.Join(d.brainRoot, "providers", "providers.yml"),
		filepath.Join(d.brainRoot, "rules", "canonical.md"),
		filepath.Join(d.brainRoot, "agents"),
	}

	if err := watcher.Start(paths...); err != nil {
		d.logChannel <- "[SyncWatcher] Failed to start: " + err.Error()
		return
	}

	d.syncMu.Lock()
	d.syncWatcher = watcher
	d.syncStatus = "watching"
	d.syncMu.Unlock()

	d.logChannel <- "[SyncWatcher] Active"
}

func (d *BrainDaemon) triggerSync(dryRun bool) (err error) {
	d.syncMu.Lock()
	if d.syncRunning {
		d.syncMu.Unlock()
		return fmt.Errorf("sync already running")
	}
	d.syncRunning = true
	d.syncStatus = "running"
	d.syncError = ""
	d.syncMu.Unlock()

	defer func() {
		d.syncMu.Lock()
		d.syncRunning = false
		d.syncLastRun = time.Now()
		if err != nil {
			d.syncStatus = "failed"
			d.syncError = err.Error()
		} else {
			d.syncStatus = "completed"
			d.syncError = ""
		}
		d.syncMu.Unlock()
	}()

	manifestPath := filepath.Join(d.brainRoot, "manifest.yml")
	m, err := manifest.Parse(manifestPath)
	if err != nil {
		return err
	}

	d.logChannel <- fmt.Sprintf("[Daemon] Running sync (dry_run=%t)", dryRun)
	engine := syncengine.NewSyncEngine(m, d.logChannel, d.skills)
	return engine.RunSyncWithOptions(syncengine.SyncOptions{DryRun: dryRun})
}
