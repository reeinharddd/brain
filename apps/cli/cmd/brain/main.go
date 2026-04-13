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
	coreruntime "github.com/reeinharrrd/brain/core/runtime"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "brain",
	Short: "Brain CLI",
}

const DAEMON_URL = "http://localhost:9090"
const DAEMON_WS = "ws://localhost:9090/ws"

func saveConfiguredRoot(root string) {
	_ = coreruntime.SaveConfiguredRoot(root)
}

func readConfiguredRoot() string {
	return coreruntime.ReadConfiguredRoot()
}

func resolveBrainRoot() string {
	return coreruntime.ResolveBrainRoot()
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

	cliDir := filepath.Join(brainRoot, "apps", "cli")
	daemonDir := filepath.Join(brainRoot, "apps", "daemon")

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
	cmd.Env = append(os.Environ(), coreruntime.BrainRootEnv+"="+resolveBrainRoot())
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
	desktopDir := filepath.Join(resolveBrainRoot(), "apps", "desktop")
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
		if err := runStatus(os.Args[2:]); err != nil {
			fmt.Println("Error:", err)
		}

	case "review":
		HandleReviewCommand(os.Args[2:])

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
		conn, _, err := websocket.DefaultDialer.Dial(authWebSocketURL(), nil)
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
	case "auth":
		HandleAuthCommand(os.Args)

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

	case "workflows":
		handleWorkflowsCommand(os.Args)

	case "delegation":
		handleDelegationCommand(os.Args)

	case "autoevolve":
		handleAutoevolveCommand(os.Args)

	case "memory":
		handleMemoryCommand(os.Args)

	case "context":
		handleContextCommand(os.Args)

	case "cost":
		handleCostCommand(os.Args)

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
	fmt.Println("  ps                             List managed processes")
	fmt.Println("  logs                           Stream real-time global logs")
	fmt.Println("  start <id> <command> [args..]  Start a specific process")
	fmt.Println("  stop <id>                      Stop a specific process")
	fmt.Println("  sync [--dry-run|status]        Trigger unified config sync")
	fmt.Println("  providers                      List available LLM providers")
	fmt.Println("  mcp serve                      Run the Brain MCP stdio server")
	fmt.Println("  skills                         Manage skills (list, create, search)")
	fmt.Println("  mcps                           Manage MCP servers")
	fmt.Println("  agents                         Manage agents")
	fmt.Println("  docs-rag <query>               Search documentation")
	fmt.Println("  workflows <subcommand>         Manage workflows (list, execute, status, dag)")
	fmt.Println("  delegation <subcommand>        Manage agent delegation (execute, status, cancel)")
	fmt.Println("  autoevolve <subcommand>        Self-improvement engine (run, approve, apply)")
	fmt.Println("  memory <subcommand>            Memory system (status, list)")
	fmt.Println("  context <subcommand>           Context management (bundle, curator)")
	fmt.Println("  cost <subcommand>              Cost tracking (budget, report)")
	fmt.Println("  hooks ...                      Run hook helpers without shell wrappers")
	fmt.Println("  auth <subcommand>             Manage daemon login sessions")
	fmt.Println("  test [suite] [flags]            Run test orchestrator (Phase 1)")
}

func runTestCommand(args []string) int {
	brainRoot := resolveBrainRoot()
	daemonDir := filepath.Join(brainRoot, "apps", "daemon")

	goArgs := []string{"run", "./cmd/testor", "test"}
	goArgs = append(goArgs, args...)

	cmd := exec.Command("go", goArgs...)
	cmd.Dir = daemonDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), coreruntime.BrainRootEnv+"="+brainRoot)

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

func handleWorkflowsCommand(args []string) {
	if len(args) < 3 {
		printWorkflowHelp()
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "list":
		resp, err := http.Get(DAEMON_URL + "/api/workflows/list")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if workflows, ok := res["workflows"].([]interface{}); ok {
			fmt.Println("Available Workflows:")
			for _, w := range workflows {
				fmt.Printf("  - %v\n", w)
			}
		}

	case "execute":
		if len(args) < 4 {
			fmt.Println("Usage: brain workflows execute <workflow-name> [--budget N]")
			return
		}
		name := args[3]
		payload := map[string]interface{}{"workflow": name}
		body, _ := json.Marshal(payload)
		resp, err := http.Post(DAEMON_URL+"/api/workflows/execute", "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if execID, ok := res["execution_id"].(string); ok {
			fmt.Printf("✅ Workflow '%s' started. Execution ID: %s\n", name, execID)
			fmt.Println("Use 'brain workflows status <execution_id>' to check progress.")
		} else {
			fmt.Printf("❌ Failed to start workflow: %v\n", res)
		}

	case "status":
		if len(args) < 4 {
			fmt.Println("Usage: brain workflows status <execution_id>")
			return
		}
		execID := args[3]
		resp, err := http.Get(DAEMON_URL + "/api/workflows/" + execID + "/status")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("Workflow Status [%s]:\n", execID)
		fmt.Printf("  Status: %v\n", res["status"])
		if nodes, ok := res["nodes"].([]interface{}); ok {
			fmt.Println("  Nodes:")
			for _, n := range nodes {
				fmt.Printf("    - %v\n", n)
			}
		}

	case "dag":
		if len(args) < 4 {
			fmt.Println("Usage: brain workflows dag <workflow-name>")
			return
		}
		name := args[3]
		resp, err := http.Get(DAEMON_URL + "/api/workflows/" + name + "/dag")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if nodes, ok := res["nodes"].([]interface{}); ok {
			fmt.Printf("Workflow DAG [%s]:\n", name)
			for _, n := range nodes {
				fmt.Printf("  - %v\n", n)
			}
		}

	default:
		printWorkflowHelp()
	}
}

func printWorkflowHelp() {
	fmt.Println("Usage: brain workflows <subcommand>")
	fmt.Println("\nSubcommands:")
	fmt.Println("  list                          List available workflows")
	fmt.Println("  execute <name>                Execute a workflow")
	fmt.Println("  status <execution_id>         Check workflow execution status")
	fmt.Println("  dag <name>                    Show workflow DAG structure")
}

func handleDelegationCommand(args []string) {
	if len(args) < 3 {
		printDelegationHelp()
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "execute":
		if len(args) < 4 {
			fmt.Println("Usage: brain delegation execute <graph_json_file>")
			return
		}
		data, err := os.ReadFile(args[3])
		if err != nil {
			fmt.Println("❌ Failed to read graph file:", err)
			return
		}
		var graph map[string]interface{}
		if err := json.Unmarshal(data, &graph); err != nil {
			fmt.Println("❌ Invalid JSON:", err)
			return
		}
		body, _ := json.Marshal(graph)
		resp, err := http.Post(DAEMON_URL+"/api/delegation/execute", "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if execID, ok := res["execution_id"].(string); ok {
			fmt.Printf("✅ Delegation started. Execution ID: %s\n", execID)
		} else {
			fmt.Printf("❌ Failed: %v\n", res)
		}

	case "status":
		if len(args) < 4 {
			fmt.Println("Usage: brain delegation status <execution_id>")
			return
		}
		execID := args[3]
		resp, err := http.Get(DAEMON_URL + "/api/delegation/" + execID + "/status")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("Delegation [%s]:\n", execID)
		fmt.Printf("  Status: %v\n", res["status"])
		if results, ok := res["results"].([]interface{}); ok {
			fmt.Println("  Results:")
			for _, r := range results {
				fmt.Printf("    - %v\n", r)
			}
		}

	case "cancel":
		if len(args) < 4 {
			fmt.Println("Usage: brain delegation cancel <execution_id>")
			return
		}
		execID := args[3]
		req, _ := http.NewRequest("POST", DAEMON_URL+"/api/delegation/"+execID+"/cancel", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("✅ Delegation %s cancelled\n", execID)

	case "executions":
		resp, err := http.Get(DAEMON_URL + "/api/delegation/executions")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if execs, ok := res["executions"].([]interface{}); ok {
			fmt.Printf("Active Executions (%d):\n", len(execs))
			for _, e := range execs {
				fmt.Printf("  - %v\n", e)
			}
		}

	default:
		printDelegationHelp()
	}
}

func printDelegationHelp() {
	fmt.Println("Usage: brain delegation <subcommand>")
	fmt.Println("\nSubcommands:")
	fmt.Println("  execute <graph.json>          Execute a delegation graph")
	fmt.Println("  status <execution_id>         Check execution status")
	fmt.Println("  cancel <execution_id>         Cancel running execution")
	fmt.Println("  executions                    List all active executions")
}

func handleAutoevolveCommand(args []string) {
	if len(args) < 3 {
		printAutoevolveHelp()
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "status":
		resp, err := http.Get(DAEMON_URL + "/api/autoevolve/status")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println("AutoEvolve Status:")
		fmt.Printf("  Enabled: %v\n", res["enabled"])
		fmt.Printf("  Telemetry Events: %v\n", res["telemetry_count"])
		fmt.Printf("  Pending Approvals: %v\n", res["pending_approvals"])
		fmt.Printf("  Applied: %v\n", res["history_count"])

	case "run":
		resp, err := http.Post(DAEMON_URL+"/api/autoevolve/run", "application/json", nil)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if report, ok := res["report"]; ok {
			fmt.Println("✅ Analysis complete:")
			prettyPrint(report)
		} else {
			fmt.Printf("❌ Failed: %v\n", res)
		}

	case "recommendations":
		resp, err := http.Get(DAEMON_URL + "/api/autoevolve/recommendations")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if recs, ok := res["recommendations"].([]interface{}); ok {
			fmt.Printf("Recommendations (%d):\n", len(recs))
			for i, r := range recs {
				fmt.Printf("  %d. %v\n", i+1, r)
			}
		}

	case "approve":
		if len(args) < 4 {
			fmt.Println("Usage: brain autoevolve approve <id>")
			return
		}
		id := args[3]
		req, _ := http.NewRequest("POST", DAEMON_URL+"/api/autoevolve/approve/"+id, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("✅ Recommendation %s approved\n", id)

	case "reject":
		if len(args) < 4 {
			fmt.Println("Usage: brain autoevolve reject <id>")
			return
		}
		id := args[3]
		req, _ := http.NewRequest("POST", DAEMON_URL+"/api/autoevolve/reject/"+id, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("✅ Recommendation %s rejected\n", id)

	case "apply":
		resp, err := http.Post(DAEMON_URL+"/api/autoevolve/apply", "application/json", nil)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if applied, ok := res["applied"].([]interface{}); ok {
			fmt.Printf("✅ Applied %d recommendations:\n", len(applied))
			for _, a := range applied {
				fmt.Printf("  - %v\n", a)
			}
		}

	case "enable":
		req, _ := http.NewRequest("POST", DAEMON_URL+"/api/autoevolve/enable", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Println("✅ AutoEvolve enabled")

	case "disable":
		req, _ := http.NewRequest("POST", DAEMON_URL+"/api/autoevolve/disable", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Println("✅ AutoEvolve disabled")

	default:
		printAutoevolveHelp()
	}
}

func printAutoevolveHelp() {
	fmt.Println("Usage: brain autoevolve <subcommand>")
	fmt.Println("\nSubcommands:")
	fmt.Println("  status                          Show AutoEvolve status")
	fmt.Println("  run                             Run analysis")
	fmt.Println("  recommendations                 List pending recommendations")
	fmt.Println("  approve <id>                    Approve a recommendation")
	fmt.Println("  reject <id>                     Reject a recommendation")
	fmt.Println("  apply                           Apply all approved recommendations")
	fmt.Println("  enable                          Enable AutoEvolve engine")
	fmt.Println("  disable                         Disable AutoEvolve engine")
}

func handleMemoryCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: brain memory <subcommand>")
		fmt.Println("\nSubcommands:")
		fmt.Println("  status         Show memory status")
		fmt.Println("  list [scope]   List memory entries")
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "status":
		resp, err := http.Get(DAEMON_URL + "/api/qdrant/status")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println("Memory Status:")
		prettyPrint(res)

	case "list":
		scope := "global"
		if len(args) >= 4 {
			scope = args[3]
		}
		resp, err := http.Get(DAEMON_URL + "/api/mcp/call?server_id=brain-knowledge&tool=list_collections")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("Memory Collections [%s]:\n", scope)
		prettyPrint(res)

	default:
		fmt.Println("Unknown memory subcommand:", subcommand)
	}
}

func handleContextCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: brain context <subcommand>")
		fmt.Println("\nSubcommands:")
		fmt.Println("  bundle         Show context bundle info")
		fmt.Println("  curator run    Run context curator")
		fmt.Println("  curator report Show curator report")
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "bundle":
		resp, err := http.Get(DAEMON_URL + "/api/status")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println("Context Status:")
		fmt.Printf("  Environment: %v\n", res["environment"])
		fmt.Printf("  Processes: %v\n", res["processes"])

	case "curator":
		if len(args) < 4 {
			fmt.Println("Usage: brain context curator <run|report>")
			return
		}
		action := args[3]
		switch action {
		case "run":
			resp, err := http.Post(DAEMON_URL+"/api/context/curator/run?dry_run=true", "application/json", nil)
			if err != nil {
				fmt.Println("❌ Error:", err)
				return
			}
			defer resp.Body.Close()
			fmt.Println("✅ Curator analysis started (dry-run)")
		case "report":
			resp, err := http.Get(DAEMON_URL + "/api/context/curator/report")
			if err != nil {
				fmt.Println("❌ Error:", err)
				return
			}
			defer resp.Body.Close()
			var res map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&res)
			fmt.Println("Curator Report:")
			prettyPrint(res)
		}

	default:
		fmt.Println("Unknown context subcommand:", subcommand)
	}
}

func handleCostCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: brain cost <subcommand>")
		fmt.Println("\nSubcommands:")
		fmt.Println("  budget [id]          Show budget status")
		fmt.Println("  report               Show cost report")
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "budget":
		id := "default"
		if len(args) >= 4 {
			id = args[3]
		}
		resp, err := http.Get(DAEMON_URL + "/api/runtime/budget/" + id)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Printf("Budget [%s]:\n", id)
		prettyPrint(res)

	case "report":
		resp, err := http.Get(DAEMON_URL + "/api/cost/report")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println("Cost Report:")
		prettyPrint(res)

	default:
		fmt.Println("Unknown cost subcommand:", subcommand)
	}
}

func prettyPrint(v interface{}) {
	data, _ := json.MarshalIndent(v, "  ", "  ")
	fmt.Println(string(data))
}
