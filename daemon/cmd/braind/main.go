package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/reeinharrrd/brain/daemon/internal/manager"
	"github.com/reeinharrrd/brain/daemon/internal/manifest"
	"github.com/reeinharrrd/brain/daemon/internal/syncengine"

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
	logChannel  chan string
	brainRoot   string
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
	d := &BrainDaemon{
		status:      "Running",
		clients:     make(map[*websocket.Conn]bool),
		procManager: manager.NewProcessManager(logCh),
		logChannel:  logCh,
		brainRoot:   root,
	}
	logCh <- "[Daemon] Using brain root: " + root
	go d.processLogs()
	return d
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
	if r.URL.Path == "/api/sync" && r.Method == "POST" {
		d.handleSync(w, r)
		return
	}
	http.NotFound(w, r)
}

func (d *BrainDaemon) handleSync(w http.ResponseWriter, r *http.Request) {
	d.logChannel <- "[Daemon] Incoming Sync Request..."
	manifestPath := filepath.Join(d.brainRoot, "manifest.yml")
	m, err := manifest.Parse(manifestPath)
	if err != nil {
		d.logChannel <- "[Daemon] Failed to parse manifest: " + err.Error()
		http.Error(w, "Parse failed", 500)
		return
	}

	go func() {
		engine := syncengine.NewSyncEngine(m, d.logChannel)
		if err := engine.RunSync(); err != nil {
			d.logChannel <- "[SyncEngine] Error running sync: " + err.Error()
		}
	}()

	w.WriteHeader(200)
	w.Write([]byte(`{"status": "Sync Initiated"}`))
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
	statuses := d.procManager.GetAllStatuses()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    d.status,
		"time":      time.Now().String(),
		"processes": len(statuses),
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
