package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const DAEMON_URL = "http://localhost:9090"
const DAEMON_WS = "ws://localhost:9090/ws"

func configRootFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "brain", "root")
}

func saveConfiguredRoot(root string) {
	conf := configRootFilePath()
	_ = os.MkdirAll(filepath.Dir(conf), 0755)
	_ = os.WriteFile(conf, []byte(root+"\n"), 0644)
}

func readConfiguredRoot() string {
	conf := configRootFilePath()
	b, err := os.ReadFile(conf)
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

func isBrainRoot(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "manifest.yml")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "cli", "cmd", "brain", "main.go")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "daemon", "cmd", "braind", "main.go")); err != nil {
		return false
	}
	return true
}

func resolveBrainRoot() string {
	if envRoot := strings.TrimSpace(os.Getenv("BRAIN_ROOT")); isBrainRoot(envRoot) {
		return envRoot
	}

	if cwd, err := os.Getwd(); err == nil {
		search := cwd
		for {
			if isBrainRoot(search) {
				return search
			}
			parent := filepath.Dir(search)
			if parent == search {
				break
			}
			search = parent
		}
	}

	if configured := readConfiguredRoot(); configured != "" {
		return configured
	}

	home, _ := os.UserHomeDir()
	fallback := filepath.Join(home, ".brain")
	if isBrainRoot(fallback) {
		return fallback
	}

	return fallback
}

func daemonReachable() bool {
	resp, err := http.Get(DAEMON_URL + "/api/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func daemonBinaryPath() string {
	if path, err := exec.LookPath("braind"); err == nil {
		return path
	}
	home, _ := os.UserHomeDir()
	localFallback := filepath.Join(home, ".local", "bin", "braind")
	if _, err := os.Stat(localFallback); err == nil {
		return localFallback
	}
	return "braind"
}

func installGlobal() {
	home, _ := os.UserHomeDir()
	brainRoot := resolveBrainRoot()
	outDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Println("Failed to prepare ~/.local/bin:", err)
		return
	}

	cliDir := filepath.Join(brainRoot, "cli")
	daemonDir := filepath.Join(brainRoot, "daemon")

	brainOut := filepath.Join(outDir, "brain")
	braindOut := filepath.Join(outDir, "braind")

	_ = os.Remove(brainOut)
	_ = os.Remove(braindOut)

	cmd1 := exec.Command("go", "build", "-o", brainOut, "./cmd/brain")
	cmd1.Dir = cliDir
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		fmt.Println("Failed building global 'brain':", err)
		return
	}

	cmd2 := exec.Command("go", "build", "-o", braindOut, "./cmd/braind/main.go")
	cmd2.Dir = daemonDir
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		fmt.Println("Failed building global 'braind':", err)
		return
	}

	fmt.Println("Global install complete:")
	fmt.Println("  -", brainOut)
	fmt.Println("  -", braindOut)
	saveConfiguredRoot(brainRoot)
	fmt.Println("  - configured root:", brainRoot)
	fmt.Println("If 'brain' is still not found, add this to ~/.zshrc:")
	fmt.Println("  export PATH=\"$HOME/.local/bin:$PATH\"")
}

func daemonStart() {
	if daemonReachable() {
		fmt.Println("Daemon is already running.")
		return
	}

	cmd := exec.Command(daemonBinaryPath())
	cmd.Env = append(os.Environ(), "BRAIN_ROOT="+resolveBrainRoot())
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start daemon:", err)
		return
	}

	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		if daemonReachable() {
			fmt.Println("Daemon started.")
			return
		}
	}

	fmt.Println("Daemon process started but health endpoint is not reachable yet.")
}

func daemonStop() {
	cmd := exec.Command("pkill", "-x", "braind")
	_ = cmd.Run()
	if daemonReachable() {
		fmt.Println("Failed to stop daemon.")
		return
	}
	fmt.Println("Daemon stopped.")
}

func startUI() {
	daemonStart()
	desktopDir := filepath.Join(resolveBrainRoot(), "desktop")
	cmd := exec.Command("bun", "run", "dev")
	cmd.Dir = desktopDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to start desktop dev UI:", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "env":
		brainEnv := strings.ToLower(strings.TrimSpace(os.Getenv("BRAIN_ENV")))
		if brainEnv == "" {
			brainEnv = "production"
		}
		fmt.Println("Brain environment:", brainEnv)
		fmt.Println("BRAIN_ROOT:", resolveBrainRoot())

	case "install-global":
		installGlobal()

	case "root":
		cwd, _ := os.Getwd()
		fmt.Println("cwd:", cwd)
		fmt.Println("configured-root:", readConfiguredRoot())
		fmt.Println("resolved-root:", resolveBrainRoot())

	case "daemon-start":
		daemonStart()

	case "daemon-stop":
		daemonStop()

	case "ui":
		startUI()

	case "status":
		resp, err := http.Get(DAEMON_URL + "/api/status")
		if err != nil {
			fmt.Println("❌ Daemon is not running or unreachable:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("🧠 Brain Daemon Status: %v (Time: %v)\n", res["status"], res["time"])
		if env, ok := res["environment"].(string); ok && env != "" {
			fmt.Printf("Environment: %v\n", env)
		}
		fmt.Printf("Active Managed Processes: %v\n", res["processes"])
		fmt.Printf("Sync Status: %v (running=%v)\n", res["sync_status"], res["sync_running"])
		if lastRun, ok := res["sync_last_run"]; ok {
			fmt.Printf("Last Sync: %v\n", lastRun)
		}
		if syncErr, ok := res["sync_error"].(string); ok && syncErr != "" {
			fmt.Printf("Sync Error: %v\n", syncErr)
		}

	case "ps":
		resp, err := http.Get(DAEMON_URL + "/api/processes")
		if err != nil {
			fmt.Println("❌ Daemon is not running or unreachable:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]string
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println("⚡ Managed Processes:")
		for id, state := range res {
			fmt.Printf("  - %s: %s\n", id, state)
		}

	case "logs":
		conn, _, err := websocket.DefaultDialer.Dial(DAEMON_WS, nil)
		if err != nil {
			fmt.Println("❌ WebSocket Failed to connect:", err)
			return
		}
		defer conn.Close()
		fmt.Println("📡 Listening for real-time daemon logs...")
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Disconnected.")
				break
			}
			var event map[string]interface{}
			if err := json.Unmarshal(msg, &event); err == nil {
				if event["event"] == "log" {
					fmt.Println(event["data"])
				} else {
					fmt.Printf("[%s] %v\n", event["event"], event["data"])
				}
			} else {
				fmt.Println(string(bytes.TrimSpace(msg)))
			}
		}

	case "mcp":
		if len(os.Args) >= 3 && os.Args[2] == "serve" {
			runMCPServer(os.Args[3:])
			return
		}
		fmt.Println("Usage: brain mcp serve [--brain-dir PATH]")
		fmt.Println("Runs the Brain MCP stdio server.")

	case "hooks":
		os.Exit(runHookCommand(os.Args[2:]))

	case "skills":
		HandleSkillsCommand(os.Args)

	case "mcps":
		HandleMCPsCommand(os.Args)

	case "agents":
		HandleAgentsCommand(os.Args)

	case "docs-rag":
		HandleDocsRagCommand(os.Args)

	case "start":
		if len(os.Args) > 3 {
			id := os.Args[2]
			cmd := os.Args[3]
			var args []string
			if len(os.Args) > 4 {
				args = os.Args[4:]
			}

			payload := map[string]interface{}{
				"id":      id,
				"command": cmd,
				"args":    args,
			}
			body, _ := json.Marshal(payload)

			resp, err := http.Post(DAEMON_URL+"/api/process/start", "application/json", bytes.NewBuffer(body))
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("❌ Failed to start process:", err)
				return
			}
			fmt.Printf("Started process: %s\n", id)
		} else {
			fmt.Println("Usage: brain start <id> <command> [args...]")
		}

	case "stop":
		if len(os.Args) > 2 {
			id := os.Args[2]
			payload := map[string]interface{}{"id": id}
			body, _ := json.Marshal(payload)

			resp, err := http.Post(DAEMON_URL+"/api/process/stop", "application/json", bytes.NewBuffer(body))
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("❌ Failed to stop process:", err)
				return
			}
			fmt.Printf("Stopped process: %s\n", id)
		} else {
			fmt.Println("Usage: brain stop <id>")
		}

	case "sync":
		handleSyncCommand(os.Args)

	case "daemon-orchestrate":
		resp, err := http.Post(DAEMON_URL+"/api/daemon/start", "application/json", nil)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		if resp.StatusCode == 200 {
			fmt.Println("✓ Daemon orchestration started")
		} else {
			fmt.Println("❌ Failed to start daemon")
		}
		resp.Body.Close()

	case "providers":
		resp, err := http.Get(DAEMON_URL + "/api/providers/available")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if providers, ok := res["available"].([]interface{}); ok {
			fmt.Println("Available Providers:")
			for _, p := range providers {
				fmt.Printf("  - %v\n", p)
			}
		}

	case "test":
		os.Exit(runTestCommand(os.Args[2:]))

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("🧠 Brain CLI (Go Daemon Client)")
	fmt.Println("Usage: brain <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  install-global                 Build/install brain and braind into ~/.local/bin")
	fmt.Println("  root                           Print current and resolved brain root")
	fmt.Println("  env                            Print the active Brain environment")
	fmt.Println("  daemon-start                   Start daemon in background")
	fmt.Println("  daemon-stop                    Stop daemon")
	fmt.Println("  daemon-orchestrate             Start daemon with full service orchestration")
	fmt.Println("  ui                             Start desktop web UI (and daemon)")
	fmt.Println("  status                         Check daemon status")
	fmt.Println("  mcp serve                      Run the Brain MCP stdio server")
	fmt.Println("  hooks ...                      Run hook helpers without shell wrappers")
	fmt.Println("  providers                      List available LLM providers")
	fmt.Println("  sync [--dry-run|status]        Trigger unified config sync")
	fmt.Println("  test [suite] [flags]            Run test orchestrator (Phase 1)")
	fmt.Println("  ps                             List managed processes")
	fmt.Println("  logs                           Stream real-time global logs")
	fmt.Println("  start <id> <command> [args..]  Start a specific process")
	fmt.Println("  stop <id>                      Stop a specific process")
}

func runTestCommand(args []string) int {
	brainRoot := resolveBrainRoot()
	daemonDir := filepath.Join(brainRoot, "daemon")

	goArgs := []string{"run", "./cmd/testor", "test"}
	goArgs = append(goArgs, args...)

	cmd := exec.Command("go", goArgs...)
	cmd.Dir = daemonDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "BRAIN_ROOT="+brainRoot)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Println("❌ Failed to run test orchestrator:", err)
		return 2
	}

	return 0
}

func handleSyncCommand(args []string) {
	if len(args) < 3 {
		triggerSync(false)
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "status":
		syncStatus()
	case "--dry-run", "-n", "dry-run":
		triggerSync(true)
	default:
		fmt.Printf("Unknown sync subcommand: %s\n", subcommand)
		printSyncHelp()
	}
}

func triggerSync(dryRun bool) {
	url := DAEMON_URL + "/api/sync"
	if dryRun {
		url += "?dry_run=true"
	}

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Println("❌ Failed to trigger sync:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Failed to trigger sync (status %d)\n", resp.StatusCode)
		return
	}

	if dryRun {
		fmt.Println("🔎 Dry-run sync triggered. Use 'brain sync status' to inspect progress.")
		return
	}

	fmt.Println("🔄 Synchronization triggered. Use 'brain sync status' or 'brain logs' to view progress.")
}

func syncStatus() {
	resp, err := http.Get(DAEMON_URL + "/api/sync/status")
	if err != nil {
		fmt.Println("❌ Failed to fetch sync status:", err)
		return
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		fmt.Println("❌ Failed to parse sync status:", err)
		return
	}

	fmt.Println("🧭 Sync Status:")
	fmt.Printf("  Status: %v\n", res["status"])
	fmt.Printf("  Running: %v\n", res["running"])
	fmt.Printf("  Last Run: %v\n", res["last_run"])
	fmt.Printf("  Watcher: %v\n", res["watcher_active"])
	if errVal, ok := res["error"].(string); ok && errVal != "" {
		fmt.Printf("  Error: %v\n", errVal)
	}
}

func printSyncHelp() {
	fmt.Println(`
Usage: brain sync [subcommand]

Subcommands:
  status           Show sync status
  --dry-run        Preview changes without writing files

Examples:
  brain sync
  brain sync --dry-run
  brain sync status
	`)
}
