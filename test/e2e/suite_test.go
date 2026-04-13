// Package e2e provides end-to-end tests for the Brain daemon.
//
// These tests start a real daemon process on a random port and exercise
// the full auth flow -- no mocks, no Docker, no external services.
package e2e

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// testBootstrapEmail is the default bootstrap credential used by tests.
	testBootstrapEmail = "test@brain.local"
	// testBootstrapPassword is the default bootstrap password used by tests.
	testBootstrapPassword = "secret123"
	// testBootstrapName is the default bootstrap name.
	testBootstrapName = "Test User"
	// testBootstrapRole is the default bootstrap role.
	testBootstrapRole = "owner"
	// daemonPort is the fixed port the daemon listens on (hardcoded in main.go).
	daemonPort = 9090
	// daemonStartTimeout is how long we wait for the daemon to become ready.
	daemonStartTimeout = 15 * time.Second
	// daemonStopTimeout is how long we wait for the daemon to exit after SIGKILL.
	daemonStopTimeout = 5 * time.Second
	// healthPollInterval is how often we poll the health endpoint.
	healthPollInterval = 200 * time.Millisecond
	// requestTimeout is the default timeout for HTTP requests in tests.
	requestTimeout = 5 * time.Second
)

// daemonMu serializes all tests that start the daemon, since the daemon
// listens on a fixed port (9090) and cannot run multiple instances simultaneously.
var daemonMu sync.Mutex

// testEnv holds per-test daemon process state.
type testEnv struct {
	t          *testing.T
	port       int
	baseURL    string
	storePath  string
	tempDir    string
	brainRoot  string
	cmd        *exec.Cmd
	cmdMu      sync.Mutex
	daemonPath string
	// ownsDaemonLock tracks whether this env still holds the global daemonMu.
	ownsDaemonLock bool
}

// killDaemonOnPort kills any process listening on the given port using lsof/pkill.
func killDaemonOnPort(t *testing.T, port int) {
	t.Helper()
	// Use fuser to kill anything listening on the port.
	cmd := exec.Command("fuser", "-k", "-9", "tcp", fmt.Sprintf("%d", port))
	_ = cmd.Run()
	// Brief pause to ensure port is freed.
	time.Sleep(300 * time.Millisecond)
}

// buildDaemonBinary locates or builds the daemon binary for the current test run.
// It is resolved once per test package via sync.Once and reused.
var (
	daemonBinary   string
	daemonOnce     sync.Once
	daemonBuildErr error
)

// resolveDaemonBinary locates or builds the daemon binary.
func resolveDaemonBinary(t *testing.T) string {
	t.Helper()
	daemonOnce.Do(func() {
		// Check for pre-built binary at repo root first.
		root := findRepoRoot()
		prebuilt := filepath.Join(root, "braind")
		if runtime.GOOS == "windows" {
			prebuilt += ".exe"
		}
		if info, err := os.Stat(prebuilt); err == nil && !info.IsDir() {
			daemonBinary = prebuilt
			return
		}

		// Fall back to building from source.
		tmpDir, err := os.MkdirTemp("", "brain-e2e-build-*")
		if err != nil {
			daemonBuildErr = fmt.Errorf("failed to create temp dir for build: %w", err)
			return
		}
		daemonBinary = filepath.Join(tmpDir, "braind")
		if runtime.GOOS == "windows" {
			daemonBinary += ".exe"
		}

		daemonModDir := filepath.Join(root, "apps", "daemon", "cmd", "braind")
		cmd := exec.Command("go", "build", "-o", daemonBinary, ".")
		cmd.Dir = daemonModDir
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			daemonBuildErr = fmt.Errorf("failed to build daemon binary: %w\n%s", err, string(out))
			return
		}
	})
	if daemonBuildErr != nil {
		t.Fatalf("%v", daemonBuildErr)
	}
	return daemonBinary
}

// findRepoRoot walks up from cwd to find the repository root (containing go.work).
func findRepoRoot() string {
	base, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(base, "go.work")); err == nil {
			return base
		}
		parent := filepath.Dir(base)
		if parent == base {
			break
		}
		base = parent
	}
	// Fallback to known path.
	return "/mnt/main1tb/work/Personal/brain"
}

// newTestEnv creates an isolated test environment with a temp directory
// and the fixed daemon port (9090) but does NOT start the daemon yet.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "auth.sqlite")
	brainRoot := setupTestBrainRoot(t, tempDir)

	return &testEnv{
		t:          t,
		port:       daemonPort,
		baseURL:    fmt.Sprintf("http://127.0.0.1:%d", daemonPort),
		storePath:  storePath,
		tempDir:    tempDir,
		brainRoot:  brainRoot,
		daemonPath: resolveDaemonBinary(t),
	}
}

// setupTestBrainRoot creates a minimal brain root directory structure
// so the daemon can resolve paths correctly.
func setupTestBrainRoot(t *testing.T, baseDir string) string {
	t.Helper()

	root := filepath.Join(baseDir, "brain-root")
	requiredDirs := []string{
		filepath.Join(root, "apps", "cli", "cmd", "brain"),
		filepath.Join(root, "apps", "daemon", "cmd", "braind"),
		filepath.Join(root, "config", "mcps"),
		filepath.Join(root, "skills"),
		filepath.Join(root, "config"),
	}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Write a minimal manifest.yml.
	manifest := []byte("version: 1\nname: test-brain\n")
	if err := os.WriteFile(filepath.Join(root, "manifest.yml"), manifest, 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Write placeholder main.go files so IsBrainRoot passes.
	placeholder := []byte("package main\n")
	for _, path := range []string{
		filepath.Join(root, "apps", "cli", "cmd", "brain", "main.go"),
		filepath.Join(root, "apps", "daemon", "cmd", "braind", "main.go"),
	} {
		if err := os.WriteFile(path, placeholder, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return root
}

// envVars returns the environment variables for starting the daemon.
func (e *testEnv) getEnvVars() []string {
	env := os.Environ()
	// Override only the auth and path-related variables.
	env = append(env,
		"BRAIN_AUTH_REQUIRED=true",
		"BRAIN_AUTH_MODE=bootstrap",
		"BRAIN_AUTH_BOOTSTRAP_EMAIL="+testBootstrapEmail,
		"BRAIN_AUTH_BOOTSTRAP_PASSWORD="+testBootstrapPassword,
		"BRAIN_AUTH_BOOTSTRAP_NAME="+testBootstrapName,
		"BRAIN_AUTH_BOOTSTRAP_ROLE="+testBootstrapRole,
		"BRAIN_AUTH_STORE_PATH="+e.storePath,
		"BRAIN_ROOT="+e.brainRoot,
		"BRAIN_PROFILE=local",
		"HOME="+e.tempDir,
	)
	return env
}

// startDaemon starts the daemon process and waits for it to become ready.
// It acquires a global lock because the daemon listens on a fixed port.
func (e *testEnv) startDaemon() {
	e.t.Helper()

	daemonMu.Lock()
	e.ownsDaemonLock = true

	// Kill any existing daemon process on port 9090.
	killDaemonOnPort(e.t, daemonPort)

	e.cmd = exec.Command(e.daemonPath)
	e.cmd.Env = e.getEnvVars()
	e.cmd.Dir = e.brainRoot
	e.cmd.Stdout = io.Discard
	e.cmd.Stderr = io.Discard

	if err := e.cmd.Start(); err != nil {
		daemonMu.Unlock()
		e.ownsDaemonLock = false
		e.t.Fatalf("failed to start daemon: %v", err)
	}

	if err := waitForDaemonReady(e.baseURL, daemonStartTimeout); err != nil {
		// Try to get daemon output for debugging.
		daemonMu.Unlock()
		e.ownsDaemonLock = false
		e.t.Fatalf("daemon did not become ready: %v", err)
	}
}

// stopDaemon stops the running daemon process and releases the global lock.
// Safe to call multiple times; only releases the lock if this env owns it.
func (e *testEnv) stopDaemon() {
	e.t.Helper()
	e.stopDaemonProcess()
	if e.ownsDaemonLock {
		e.ownsDaemonLock = false
		daemonMu.Unlock()
	}
}

// stopDaemonProcess stops the daemon process without touching the global lock.
func (e *testEnv) stopDaemonProcess() {
	e.cmdMu.Lock()
	defer e.cmdMu.Unlock()

	if e.cmd == nil || e.cmd.Process == nil {
		return
	}

	// Try graceful shutdown first.
	_ = e.cmd.Process.Signal(os.Interrupt)

	done := make(chan error, 1)
	go func() {
		done <- e.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited cleanly.
	case <-time.After(2 * time.Second):
		// Force kill.
		_ = e.cmd.Process.Kill()
		<-done
	}

	e.cmd = nil
}

// restartDaemon stops and restarts the daemon, preserving the same store.
// The global lock is held throughout the restart.
func (e *testEnv) restartDaemon() {
	e.t.Helper()

	if !e.ownsDaemonLock {
		e.t.Fatal("restartDaemon called but this env does not own the daemon lock")
	}

	// Stop the current daemon process (does NOT release daemonMu).
	e.stopDaemonProcess()

	// Brief pause to ensure port is freed.
	time.Sleep(500 * time.Millisecond)

	// Kill anything else on the port just in case.
	killDaemonOnPort(e.t, daemonPort)

	e.cmd = exec.Command(e.daemonPath)
	e.cmd.Env = e.getEnvVars()
	e.cmd.Dir = e.brainRoot
	e.cmd.Stdout = io.Discard
	e.cmd.Stderr = io.Discard

	if err := e.cmd.Start(); err != nil {
		e.t.Fatalf("failed to restart daemon: %v", err)
	}

	if err := waitForDaemonReady(e.baseURL, daemonStartTimeout); err != nil {
		e.t.Fatalf("restarted daemon did not become ready: %v", err)
	}
}

// waitForDaemonReady polls the health endpoint until the daemon responds
// or the timeout is reached.
func waitForDaemonReady(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{
		Timeout:   1 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(healthPollInterval)
	}
	return fmt.Errorf("health endpoint did not respond within %s", timeout)
}

// httpClient returns a test HTTP client with timeout.
func (e *testEnv) httpClient() *http.Client {
	return &http.Client{Timeout: requestTimeout}
}

// doJSON performs an HTTP request and decodes the JSON response into respBody.
func (e *testEnv) doJSON(method, path string, body interface{}, headers map[string]string, respBody interface{}) (*http.Response, error) {
	e.t.Helper()

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			e.t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, e.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := e.httpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return resp, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return resp, nil
}

// login performs a bootstrap login and returns the token and refresh token.
func (e *testEnv) login(email, password string) (token, refreshToken string) {
	e.t.Helper()

	type loginResp struct {
		Success      bool   `json:"success"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	var result loginResp
	_, err := e.doJSON(http.MethodPost, "/api/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil, &result)
	if err != nil {
		e.t.Fatalf("login request failed: %v", err)
	}
	if !result.Success {
		e.t.Fatalf("login unsuccessful")
	}
	if result.Token == "" {
		e.t.Fatalf("login response missing token")
	}
	return result.Token, result.RefreshToken
}

// authHeader returns the Authorization header value for the given token.
func authHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// skipIfSlow skips the test if the -short flag is set.
func skipIfSlow(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping slow E2E test in -short mode")
	}
}

// TestMain can be used for package-level setup/teardown.
// We locate or build the daemon binary once here.
func TestMain(m *testing.M) {
	// Locate or build the daemon binary.
	repoRoot := findRepoRoot()
	prebuilt := filepath.Join(repoRoot, "braind")
	if runtime.GOOS == "windows" {
		prebuilt += ".exe"
	}

	if info, err := os.Stat(prebuilt); err == nil && !info.IsDir() {
		daemonBinary = prebuilt
	} else {
		// Fall back to building from source.
		tmpDir, err := os.MkdirTemp("", "brain-e2e-build-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create temp dir for build: %v\n", err)
			os.Exit(1)
		}
		bin := filepath.Join(tmpDir, "braind")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}

		daemonModDir := filepath.Join(repoRoot, "apps", "daemon", "cmd", "braind")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = daemonModDir
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to build daemon binary: %v\n%s\n", err, string(out))
			os.Exit(1)
		}
		daemonBinary = bin

		defer func() {
			_ = os.Remove(bin)
			_ = os.Remove(tmpDir)
		}()
	}

	code := m.Run()
	os.Exit(code)
}

// --- Convenience types for response parsing ---

// StatusResponse represents the /api/auth/status response.
type StatusResponse struct {
	Required      bool   `json:"required"`
	Mode          string `json:"mode"`
	Authenticated bool   `json:"authenticated"`
	User          *struct {
		ID           string   `json:"id"`
		Email        string   `json:"email"`
		Name         string   `json:"name"`
		Role         string   `json:"role"`
		Capabilities []string `json:"capabilities"`
		Sections     []string `json:"sections"`
	} `json:"user,omitempty"`
	Session        *struct {
		ExpiresAt string `json:"expires_at"`
		LastUsed  string `json:"last_used"`
	} `json:"session,omitempty"`
	AllowedSections []string `json:"allowed_sections"`
	ActiveSessions  int      `json:"active_sessions"`
	Message         string   `json:"message"`
}

// ErrorResponse represents a JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// InviteResponse represents an invite creation response.
type InviteResponse struct {
	Success bool `json:"success"`
	Invite  struct {
		Token string `json:"token"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"invite"`
}

// ConsumeInviteResponse represents the invite consume response.
type ConsumeInviteResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// RefreshResponse represents the /api/auth/refresh response.
type RefreshResponse struct {
	Success      bool   `json:"success"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Message      string `json:"message"`
}

// AgentsResponse represents the /api/agents response.
type AgentsResponse struct {
	Agents []interface{} `json:"agents"`
}

// Helper: assert that a string slice contains a value.
func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}

// Helper: assert response status code matches expected.
func requireStatusCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}

// Helper: check if daemon process is still running.
func (e *testEnv) isRunning() bool {
	e.cmdMu.Lock()
	defer e.cmdMu.Unlock()
	if e.cmd == nil || e.cmd.Process == nil {
		return false
	}
	// Signal 0 checks if process is alive.
	return e.cmd.Process.Signal(os.Interrupt) == nil
}

// Helper: get port as string.
func (e *testEnv) portStr() string {
	return strconv.Itoa(e.port)
}

// Helper: parse an integer from a string field in a map.
func parseIntField(t *testing.T, m map[string]interface{}, key string) int {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Fatalf("response missing field %q", key)
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	default:
		t.Fatalf("field %q has unexpected type %T", key, v)
	}
	return 0
}
