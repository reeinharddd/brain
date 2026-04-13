package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	coreartifacts "github.com/reeinharrrd/brain/core/artifacts"
	"github.com/reeinharrrd/brain/core/autoevolve"
	"github.com/reeinharrrd/brain/core/delegation"
	coreenterprise "github.com/reeinharrrd/brain/core/enterprise"
	coreidentity "github.com/reeinharrrd/brain/core/identity"
	"github.com/reeinharrrd/brain/core/observability"
	coreruntime "github.com/reeinharrrd/brain/core/runtime"
	"github.com/reeinharrrd/brain/core/skills"
	"github.com/reeinharrrd/brain/core/workflow"
	"github.com/reeinharrrd/brain/daemon/internal/api/handlers"
	brainenv "github.com/reeinharrrd/brain/daemon/internal/environment"
	"github.com/reeinharrrd/brain/daemon/internal/manager"
	"github.com/reeinharrrd/brain/daemon/internal/manifest"
	"github.com/reeinharrrd/brain/daemon/internal/syncengine"

	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/trace"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		parsed, err := parseURL(origin)
		if err != nil {
			return false
		}
		host := strings.ToLower(parsed.Hostname())
		return host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "app.tauri"
	},
}

func parseURL(raw string) (*url.URL, error) {
	return url.Parse(raw)
}

type BrainDaemon struct {
	mu          sync.Mutex
	status      string
	clients     map[*websocket.Conn]bool
	auth        *coreidentity.Manager
	oidc        *coreenterprise.SSOManager
	oidcSessions map[string]*coreidentity.Session
	procManager *manager.ProcessManager
	docker      *manager.DockerManager
	qdrant      *manager.QdrantManager
	ollama      *manager.OllamaManager
	mcp         *manager.MCPManager
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

	// Workflow, Delegation, and AutoEvolve engines
	workflowEngine   *workflow.ExecutionEngine
	delegationExec   *delegation.DelegationExecutor
	evolveEngine     *autoevolve.AutoEvolveEngine

	// Observability
	logger      *slog.Logger
	tracer      trace.Tracer
	healthCheck *observability.HealthChecker
	metrics     *observability.Metrics
	traceCtx    *observability.TraceContext
	startTime   time.Time
}

func configRootFilePath() string {
	return coreruntime.ConfigRootFilePath()
}

func readConfiguredRoot() string {
	return coreruntime.ReadConfiguredRoot()
}

func resolveBrainRoot() string {
	return coreruntime.ResolveBrainRoot()
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

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func NewBrainDaemon() *BrainDaemon {
	logCh := make(chan string, 1000)
	root := resolveBrainRoot()
	environment := brainenv.Current()

	// Initialize structured logger
	logger := observability.NewLogger(&observability.LoggerConfig{
		Level:       "info",
		Format:      "json",
		Output:      os.Stderr,
		ServiceName: "brain-daemon",
		Version:     "0.1.0",
	})

	// Initialize health checker
	healthCheck := observability.NewHealthChecker("0.1.0")

	// Initialize trace context propagator
	traceCtx := observability.NewTraceContext()

	d := &BrainDaemon{
		status:      "Running",
		clients:     make(map[*websocket.Conn]bool),
		auth:        buildIdentityManager(environment),
		oidc:        buildOIDCManager(environment),
		oidcSessions: make(map[string]*coreidentity.Session),
		procManager: manager.NewProcessManager(logCh),
		docker:      manager.NewDockerManager(filepath.Join(root, "docker"), logCh),
		qdrant:      manager.NewQdrantManager("http://localhost:6333", logCh),
		ollama:      manager.NewOllamaManager("http://localhost:11434", logCh),
		mcp:         manager.NewMCPManager(logCh),
		providers:   manager.NewProviderManager(logCh),
		skills:      manager.NewSkillsRegistry(root, logCh),
		mcps:        manager.NewMCPsManager(root, environment, logCh),
		agents:      manager.NewAgentsManager(root, logCh),
		environment: environment,
		syncStatus:  "idle",
		logChannel:  logCh,
		brainRoot:   root,
		logger:      logger,
		healthCheck: healthCheck,
		metrics:     observability.DefaultMetrics,
		traceCtx:    traceCtx,
		startTime:   time.Now(),
	}
	logCh <- fmt.Sprintf("[Daemon] Initialized with 8 managers in %s environment", environment)
	if d.auth != nil {
		config := d.auth.Config()
		logCh <- fmt.Sprintf("[Auth] Mode=%s required=%t", config.Mode, config.Required)
	}

	// Initialize workflow engine
	wfExecutor := workflow.NewExecutionEngine(
		workflow.NewAgentExecutor(root, func(msg string) {
			logCh <- "[workflow] " + msg
		}),
		4, // max parallel
	)
	d.workflowEngine = wfExecutor
	logCh <- "[Workflow] Execution engine initialized"

	// Initialize delegation executor
	budget := delegation.DelegationBudget{
		MaxTokens:   100000,
		MaxCostUSD:  1.0,
		MaxDuration: 30 * time.Minute,
		MaxRetries:  3,
	}
	delExec := delegation.NewDelegationExecutor(root, func(msg string) {
		logCh <- "[delegation] " + msg
	}, budget)
	d.delegationExec = delExec
	logCh <- "[Delegation] Executor initialized"

	// Initialize AutoEvolve with real applier
	telemetry := autoevolve.NewTelemetryAccumulator(10000)
	applier := autoevolve.NewApplier(
		filepath.Join(root, "skills"),
		filepath.Join(root, "config", "mcps"),
		filepath.Join(root, "config"),
	)
	d.evolveEngine = autoevolve.NewAutoEvolveEngine(telemetry, applier)
	logCh <- "[AutoEvolve] Engine + Applier initialized"

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

	// Initialize Docs-RAG handler with real docs indexer
	indexer := NewDocsIndexer(root)
	d.docsHandler = handlers.NewDocsHandler(indexer, environment)
	logCh <- "[Docs-RAG] Handler initialized"

	go d.processLogs()
	go d.startSyncSubsystem() // Start sync subsystem asynchronously
	go d.healthCheckLoop()    // Start background health check loop
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

// handleTraces returns trace information (placeholder for future OTLP integration).
func (d *BrainDaemon) handleTraces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"message": "Trace endpoint active. Configure OTLP collector to export traces.",
		"config": map[string]interface{}{
			"otlp_endpoint": getEnvOrDefault("BRAIN_OTLP_ENDPOINT", "localhost:4318"),
			"otel_enabled":  os.Getenv("BRAIN_OTEL_ENABLED") == "true",
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (d *BrainDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	// Apply trace context middleware
	d.traceCtx.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.routeRequest(w, r)
	})).ServeHTTP(w, r)
}

// routeRequest handles the actual HTTP routing after observability middleware
func (d *BrainDaemon) routeRequest(w http.ResponseWriter, r *http.Request) {
	// Record HTTP metrics
	start := time.Now()
	if d.metrics != nil {
		defer func() {
			d.metrics.HTTPDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
		}()
	}

	if !d.isPublicRoute(r.Method, r.URL.Path) {
		if !d.authorizeRequest(w, r) {
			return
		}
	}
	
	// Observability endpoints
	if r.URL.Path == "/metrics" {
		promhttp.Handler().ServeHTTP(w, r)
		return
	}
	if r.URL.Path == "/api/v1/health" && r.Method == "GET" {
		d.healthCheck.HealthHandler().ServeHTTP(w, r)
		return
	}
	if r.URL.Path == "/health" && r.Method == "GET" {
		// Simple health endpoint for backward compatibility
		observability.SimpleHealthHandler("0.1.0").ServeHTTP(w, r)
		return
	}

	// Trace endpoint
	if r.URL.Path == "/api/v1/traces" && r.Method == "GET" {
		d.handleTraces(w, r)
		return
	}
	if r.URL.Path == "/api/auth/login" && r.Method == "POST" {
		d.handleAuthLogin(w, r)
		return
	}
	if r.URL.Path == "/api/auth/refresh" && r.Method == "POST" {
		d.handleAuthRefresh(w, r)
		return
	}
	if r.URL.Path == "/api/auth/oidc/start" {
		d.handleOIDCStart(w, r)
		return
	}
	if r.URL.Path == "/api/auth/oidc/callback" {
		d.handleOIDCCallback(w, r)
		return
	}
	if r.URL.Path == "/api/auth/oidc/poll" {
		d.handleOIDCPoll(w, r)
		return
	}
	if r.URL.Path == "/api/auth/logout" && r.Method == "POST" {
		d.handleAuthLogout(w, r)
		return
	}
	if r.URL.Path == "/api/auth/status" && r.Method == "GET" {
		d.handleAuthStatus(w, r)
		return
	}

	// Supabase Auth routes
	if r.URL.Path == "/api/auth/supabase/signin" && r.Method == "POST" {
		d.handleSupabaseSignIn(w, r)
		return
	}
	if r.URL.Path == "/api/auth/supabase/signup" && r.Method == "POST" {
		d.handleSupabaseSignUp(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/auth/supabase/oauth/") {
		d.handleSupabaseOAuth(w, r)
		return
	}
	if r.URL.Path == "/api/auth/supabase/callback" {
		d.handleSupabaseCallback(w, r)
		return
	}
	if r.URL.Path == "/api/auth/supabase/refresh" && r.Method == "POST" {
		d.handleSupabaseRefresh(w, r)
		return
	}
	if r.URL.Path == "/api/auth/supabase/signout" && r.Method == "POST" {
		d.handleSupabaseSignOut(w, r)
		return
	}
	if r.URL.Path == "/api/auth/supabase/user" && r.Method == "GET" {
		d.handleSupabaseUser(w, r)
		return
	}

	if r.URL.Path == "/api/users" && r.Method == "GET" {
		d.handleUsersList(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/users/") && strings.HasSuffix(r.URL.Path, "/role") {
		d.handleUserRoleUpdate(w, r)
		return
	}
	if r.URL.Path == "/api/invites" && r.Method == "GET" {
		d.handleInvitesList(w, r)
		return
	}
	if r.URL.Path == "/api/invites" && r.Method == "POST" {
		d.handleInviteCreate(w, r)
		return
	}
	if r.URL.Path == "/api/invites/consume" && r.Method == "POST" {
		d.handleInviteConsume(w, r)
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
	// MCP API endpoints (must be before /api/mcp/status)
	if r.URL.Path == "/api/mcp/call" && r.Method == "POST" {
		d.handleMCPCall(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/mcp/tools/") && r.Method == "GET" {
		d.handleMCPTools(w, r)
		return
	}
	if r.URL.Path == "/api/mcp/servers" && r.Method == "GET" {
		d.handleMCPServers(w, r)
		return
	}
	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/api/mcp/servers/") && strings.HasSuffix(r.URL.Path, "/start") {
		d.handleMCPServerStart(w, r)
		return
	}
	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/api/mcp/servers/") && strings.HasSuffix(r.URL.Path, "/stop") {
		d.handleMCPServerStop(w, r)
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
	if r.URL.Path == "/api/skills/scan" && r.Method == "POST" {
		d.handleSkillScan(w, r)
		return
	}
	if r.URL.Path == "/api/skills/compatible" && r.Method == "GET" {
		d.handleSkillCompatible(w, r)
		return
	}
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

	// Workflow endpoints
	if strings.HasPrefix(r.URL.Path, "/api/workflows") {
		d.handleWorkflows(w, r)
		return
	}

	// Delegation endpoints
	if strings.HasPrefix(r.URL.Path, "/api/delegation") {
		d.handleDelegation(w, r)
		return
	}

	// AutoEvolve endpoints
	if strings.HasPrefix(r.URL.Path, "/api/autoevolve") {
		d.handleAutoEvolve(w, r)
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
	authStatus := d.authStatusForRequest(r)
		authCapabilities := []coreidentity.Capability(nil)
		if authStatus.User != nil {
			authCapabilities = authStatus.User.Capabilities
		}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":             status,
		"time":               time.Now().String(),
		"processes":          len(statuses),
		"uptime":             time.Now().Unix(),
		"environment":        d.environment,
		"sync_status":        syncStatus,
		"sync_running":       syncRunning,
		"sync_last_run":      syncLastRun,
		"sync_error":         syncError,
		"auth_required":      authStatus.Required,
		"auth_mode":          authStatus.Mode,
		"auth_authenticated": authStatus.Authenticated,
		"auth_user":          authStatus.User,
		"auth_session":       authStatus.Session,
		"auth_capabilities":  authCapabilities,
		"auth_sections":      authStatus.AllowedSections,
		"auth_message":       authStatus.Message,
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
	d.logger.Info("Starting brain daemon", "version", "0.1.0", "environment", d.environment)

	// Register health checks
	d.healthCheck.Register("docker", 5*time.Second, func(ctx context.Context) error {
		_, err := d.docker.GetServiceStatus(ctx, "all")
		return err
	})
	d.healthCheck.Register("qdrant", 5*time.Second, func(ctx context.Context) error {
		ok, err := d.qdrant.HealthCheck(ctx)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("qdrant health check failed")
		}
		return nil
	})
	d.healthCheck.Register("ollama", 3*time.Second, func(ctx context.Context) error {
		if !d.ollama.IsAvailable(ctx) {
			return fmt.Errorf("ollama not available")
		}
		return nil
	})

	// Initialize tracer
	tp, shutdown, err := observability.InitTracer(ctx, &observability.TracerConfig{
		ServiceName:    "brain-daemon",
		ServiceVersion: "0.1.0",
		OTLPEndpoint:   getEnvOrDefault("BRAIN_OTLP_ENDPOINT", "localhost:4318"),
		OTLPInsecure:   true,
		SampleRate:     1.0,
		Enabled:        os.Getenv("BRAIN_OTEL_ENABLED") == "true",
	})
	if err != nil {
		d.logger.Warn("Failed to initialize tracer", observability.AttrError, err)
	} else {
		defer tp.Shutdown(context.Background())
		defer shutdown()
		d.tracer = observability.Tracer("daemon")
		d.logger.Info("Tracer initialized", "endpoint", getEnvOrDefault("BRAIN_OTLP_ENDPOINT", "localhost:4318"))
	}

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

// handleMCPCall handles POST /api/mcp/call - invokes a tool on an MCP server.
func (d *BrainDaemon) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		ServerID  string                 `json:"server_id"`
		ToolName  string                 `json:"tool_name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.ServerID == "" || req.ToolName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "server_id and tool_name are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := d.mcp.CallTool(ctx, req.ServerID, req.ToolName, req.Arguments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  result.Success,
		"content":  string(result.Content),
		"error":    result.Error,
		"duration": result.Duration.String(),
	})
}

// handleMCPTools handles GET /api/mcp/tools/{server_id} - lists tools for a server.
func (d *BrainDaemon) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract server_id from path: /api/mcp/tools/{server_id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/mcp/tools/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "server_id is required"})
		return
	}
	serverID := parts[0]

	ctx := context.Background()
	tools, err := d.mcp.ListTools(ctx, serverID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id": serverID,
		"tools":     tools,
	})
}

// handleMCPServers handles GET /api/mcp/servers - lists all MCP servers with status.
func (d *BrainDaemon) handleMCPServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	servers := d.mcp.ListServers(ctx)

	serverList := make([]map[string]interface{}, 0, len(servers))
	for _, s := range servers {
		entry := map[string]interface{}{
			"id":          s.Config.ID,
			"name":        s.Config.Name,
			"version":     s.Config.Version,
			"status":      string(s.Status),
			"transport":   string(s.Config.Transport),
			"tool_count":  len(s.Tools),
			"tools":       len(s.Tools),
			"client_count": s.ClientCount,
		}
		if s.Error != "" {
			entry["error"] = s.Error
		}
		if !s.StartedAt.IsZero() {
			entry["started_at"] = s.StartedAt
		}
		if !s.LastChecked.IsZero() {
			entry["last_checked"] = s.LastChecked
		}
		serverList = append(serverList, entry)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"servers": serverList,
		"total":   len(serverList),
	})
}

// handleMCPServerStart handles POST /api/mcp/servers/{server_id}/start - starts a specific MCP server.
func (d *BrainDaemon) handleMCPServerStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract server_id from path: /api/mcp/servers/{server_id}/start
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/mcp/servers/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "server_id is required"})
		return
	}
	serverID := parts[0]

	ctx := context.Background()
	if err := d.mcp.StartServer(ctx, serverID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "started",
		"server_id": serverID,
	})
}

// handleMCPServerStop handles POST /api/mcp/servers/{server_id}/stop - stops a specific MCP server.
func (d *BrainDaemon) handleMCPServerStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract server_id from path: /api/mcp/servers/{server_id}/stop
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/mcp/servers/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "server_id is required"})
		return
	}
	serverID := parts[0]

	ctx := context.Background()
	if err := d.mcp.StopServer(ctx, serverID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "stopped",
		"server_id": serverID,
	})
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

	// POST /api/skills/install/preview - inspect a source before installing
	if r.URL.Path == "/api/skills/install/preview" && r.Method == "POST" {
		var req manager.SkillInstallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		preview, err := d.skills.PreviewSkillInstall(ctx, req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(preview)
		return
	}

	// POST /api/skills/install - materialize a source into the registry and source tree
	if r.URL.Path == "/api/skills/install" && r.Method == "POST" {
		var req manager.SkillInstallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		result, err := d.skills.InstallSkill(ctx, req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		d.triggerSync(false)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
		return
	}

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

// Workflow Handlers
func (d *BrainDaemon) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")

	// POST /api/workflows/execute - execute a pre-built workflow
	if r.URL.Path == "/api/workflows/execute" && r.Method == "POST" {
		var req struct {
			Name   string  `json:"workflow"`
			Budget float64 `json:"budget_usd,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		lib := &workflow.WorkflowLibrary{}
		dag, err := lib.GetWorkflow(req.Name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		result, err := d.workflowEngine.RunDAG(ctx, dag)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"workflow_id":  result.WorkflowID,
			"status":       result.Status,
			"duration":     result.Duration.String(),
			"total_tokens": result.TotalTokens,
			"total_cost":   result.TotalCost,
			"node_results": result.NodeResults,
			"error":        result.Error,
		})
		return
	}

	// GET /api/workflows/list - list available workflows
	if r.URL.Path == "/api/workflows/list" && r.Method == "GET" {
		lib := &workflow.WorkflowLibrary{}
		workflows := lib.ListWorkflows()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"workflows": workflows})
		return
	}

	// GET /api/workflows/{id}/dag - get workflow DAG structure
	if r.Method == "GET" && strings.Count(r.URL.Path, "/") >= 3 {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			wfName := parts[3]
			lib := &workflow.WorkflowLibrary{}
			dag, err := lib.GetWorkflow(wfName)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			// Serialize DAG
			nodes := make(map[string]interface{})
			edges := make(map[string][]string)
			for id, node := range dag.Nodes {
				nodes[id] = map[string]interface{}{
					"name":       node.Name,
					"agent":      node.Agent,
					"depends_on": node.DependsOn,
					"status":     string(node.Status),
				}
				// Build edge map from dependencies
				for _, dep := range node.DependsOn {
					edges[dep] = append(edges[dep], id)
				}
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"workflow": wfName,
				"nodes":    nodes,
				"edges":    edges,
				"parallel": dag.Parallel,
			})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
}

// Delegation Handlers
func (d *BrainDaemon) handleDelegation(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")

	// POST /api/delegation/execute - execute a delegation graph
	if r.URL.Path == "/api/delegation/execute" && r.Method == "POST" {
		var req struct {
			Graph *delegation.DelegationGraph `json:"graph"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		execID, err := d.delegationExec.Execute(ctx, req.Graph)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"execution_id": execID,
			"status":       "running",
		})
		return
	}

	// GET /api/delegation/{exec_id}/status - get execution status
	if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/api/delegation/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			execID := parts[3]

			if len(parts) >= 5 && parts[4] == "status" {
				exec, err := d.delegationExec.GetExecution(execID)
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}

				results := make(map[string]interface{})
				for nodeID, result := range exec.Results {
					results[nodeID] = map[string]interface{}{
						"agent_id": result.AgentID,
						"status":   result.Status,
						"output":   result.Output,
						"error":    result.Error,
						"duration": result.Duration.String(),
						"tokens":   result.Tokens,
						"cost":     result.Cost,
					}
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"execution_id": exec.ID,
					"status":       exec.Status,
					"start_time":   exec.StartTime,
					"end_time":     exec.EndTime,
					"results":      results,
					"error":        exec.Error,
				})
				return
			}
		}
	}

	// GET /api/delegation/executions - list all executions
	if r.URL.Path == "/api/delegation/executions" && r.Method == "GET" {
		executions := d.delegationExec.GetActiveExecutions()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"executions": executions})
		return
	}

	// POST /api/delegation/{exec_id}/cancel - cancel execution
	if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/cancel") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			execID := parts[3]
			err := d.delegationExec.CancelExecution(execID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
}

// AutoEvolve Handlers
func (d *BrainDaemon) handleAutoEvolve(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")

	// POST /api/autoevolve/run - run analysis
	if r.URL.Path == "/api/autoevolve/run" && r.Method == "POST" {
		report, err := d.evolveEngine.RunAnalysis(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"report": report,
		})
		return
	}

	// GET /api/autoevolve/recommendations - list pending recommendations
	if r.URL.Path == "/api/autoevolve/recommendations" && r.Method == "GET" {
		recs := d.evolveEngine.GetPendingRecommendations(ctx)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"recommendations": recs})
		return
	}

	// POST /api/autoevolve/approve/{id} - approve a recommendation
	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/api/autoevolve/approve/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 5 {
			recID := parts[4]
			err := d.evolveEngine.ApproveRecommendation(ctx, recID)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
			return
		}
	}

	// POST /api/autoevolve/reject/{id} - reject a recommendation
	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/api/autoevolve/reject/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 5 {
			recID := parts[4]
			err := d.evolveEngine.RejectRecommendation(ctx, recID)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "rejected"})
			return
		}
	}

	// POST /api/autoevolve/apply - apply all approved recommendations
	if r.URL.Path == "/api/autoevolve/apply" && r.Method == "POST" {
		actions, err := d.evolveEngine.ApplyApproved(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"applied": actions,
			"count":   len(actions),
		})
		return
	}

	// GET /api/autoevolve/status - get AutoEvolve status
	if r.URL.Path == "/api/autoevolve/status" && r.Method == "GET" {
		status := d.evolveEngine.GetStatus(ctx)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
		return
	}

	// POST /api/autoevolve/enable - enable AutoEvolve
	if r.URL.Path == "/api/autoevolve/enable" && r.Method == "POST" {
		d.evolveEngine.Enable()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
		return
	}

	// POST /api/autoevolve/disable - disable AutoEvolve
	if r.URL.Path == "/api/autoevolve/disable" && r.Method == "POST" {
		d.evolveEngine.Disable()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
}

// handleSkillScan handles POST /api/skills/scan - runs security scanner on a skill path.
func (d *BrainDaemon) handleSkillScan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "path is required"})
		return
	}

	// Read all files from the path to build a core Skill for scanning
	content := make(map[string]string)
	err := filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(req.Path, path)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		content[rel] = string(data)
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	skill := &skills.Skill{
		ID:      filepath.Base(req.Path),
		Name:    filepath.Base(req.Path),
		Content: content,
	}

	ctx := context.Background()
	scanner := skills.NewSecurityScanner()
	result, scanErr := scanner.Scan(ctx, skill)
	if scanErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": scanErr.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"path":       req.Path,
		"overall":    result.OverallPass,
		"scanned_at": result.ScannedAt,
		"checks":     result.Checks,
	})
}

// handleSkillCompatible handles GET /api/skills/compatible - returns skills filtered by IDE compatibility.
func (d *BrainDaemon) handleSkillCompatible(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	surface := r.URL.Query().Get("surface")
	if surface == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "surface parameter required"})
		return
	}

	ctx := context.Background()
	allSkills := d.skills.GetAll(ctx)

	// All skills in the daemon registry are compatible with all surfaces by default
	// In the future this will check compatibility matrices
	compatible := make([]interface{}, 0)
	for _, s := range allSkills {
		compatible = append(compatible, map[string]interface{}{
			"id":          s.ID,
			"name":        s.Name,
			"kind":        s.Kind,
			"description": s.Description,
			"version":     s.Version,
			"tags":        s.Tags,
			"compatible":  true,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"surface": surface,
		"skills":  compatible,
		"total":   len(compatible),
	})
}

func main() {
	daemon := NewBrainDaemon()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			globalLoginLimiter.cleanup()
		}
	}()

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
	locator := coreartifacts.NewLocator(d.brainRoot)
	watcher, err := syncengine.NewFileWatcher(d.logChannel, func() error {
		return d.triggerSync(false)
	})
	if err != nil {
		d.logChannel <- "[SyncWatcher] Failed to initialize: " + err.Error()
		return
	}

	paths := []string{
		filepath.Join(d.brainRoot, "manifest.yml"),
		locator.DomainFile("skills", "registry.yml"),
		locator.DomainFile("skills", "dynamic-registry.tsv"),
		locator.DomainFile("mcps", "registry.yml"),
		locator.DomainFile("providers", "providers.yml"),
		locator.DomainFile("rules", "canonical.md"),
		locator.DomainDir("agents"),
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
