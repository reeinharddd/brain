package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zalando/go-keyring"
)

type storedAuthSession struct {
	Token            string    `json:"token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	Role             string    `json:"role"`
	Mode             string    `json:"mode"`
	Required         bool      `json:"required"`
	Capabilities     []string  `json:"capabilities"`
	AllowedSections  []string  `json:"allowed_sections"`
}

type authLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var authSessionMu sync.Mutex

func init() {
	installAuthTransport()
}

func installAuthTransport() {
	parsed, err := url.Parse(DAEMON_URL)
	if err != nil {
		return
	}
	base := http.DefaultTransport
	if base == nil {
		base = http.DefaultTransport
	}
	http.DefaultClient.Transport = &daemonAuthTransport{
		base:      base,
		daemonHost: parsed.Host,
	}
}

type daemonAuthTransport struct {
	base      http.RoundTripper
	daemonHost string
}

func (t *daemonAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if t == nil || t.base == nil {
		return http.DefaultTransport.RoundTrip(req)
	}

	clone := req.Clone(req.Context())
	if clone.URL != nil && clone.URL.Host == t.daemonHost {
		if token := loadStoredAuthToken(); token != "" && clone.Header.Get("Authorization") == "" {
			clone.Header = clone.Header.Clone()
			clone.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return t.base.RoundTrip(clone)
}

func HandleAuthCommand(args []string) {
	if len(args) < 3 {
		printAuthHelp()
		return
	}

	subcommand := args[2]
	switch subcommand {
	case "login":
		handleAuthLogin(args[3:])
	case "logout":
		handleAuthLogout()
	case "status":
		handleAuthStatus()
	default:
		fmt.Printf("Unknown auth subcommand: %s\n", subcommand)
		printAuthHelp()
	}
}

func handleAuthLogin(args []string) {
	flags := flag.NewFlagSet("auth login", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	email := flags.String("email", "", "Email address")
	password := flags.String("password", "", "Password")
	jsonOutput := flags.Bool("json", false, "Output JSON")
	if err := flags.Parse(args); err != nil {
		return
	}

	status := fetchDaemonAuthStatus()
	if strings.TrimSpace(*password) != "" {
		session, result, err := performBootstrapLogin(*email, *password)
		if err != nil {
			fmt.Fprintln(os.Stderr, "❌", err)
			return
		}
		printLoginResult(*jsonOutput, session, result)
		return
	}

	if strings.EqualFold(status.Mode, "oidc") {
		session, result, err := performOIDCLogin(*email)
		if err != nil {
			fmt.Fprintln(os.Stderr, "❌", err)
			return
		}
		printLoginResult(*jsonOutput, session, result)
		return
	}

	fmt.Fprintln(os.Stderr, "Usage: brain auth login --email <email> --password <password> or use OIDC mode on the daemon")
}

type daemonAuthStatus struct {
	Mode         string `json:"mode"`
	Required     bool   `json:"required"`
	Authenticated bool  `json:"authenticated"`
	Message      string `json:"message"`
}

type oidcStartResponse struct {
	Success          bool      `json:"success"`
	Provider         string    `json:"provider"`
	State            string    `json:"state"`
	AuthorizationURL string    `json:"authorization_url"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type oidcPollResponse struct {
	Ready   bool               `json:"ready"`
	State   string             `json:"state"`
	Session *oidcSessionResult `json:"session"`
	Message string             `json:"message"`
}

type oidcSessionResult struct {
	Success          bool                  `json:"success"`
	State            string                `json:"state"`
	Token            string                `json:"token"`
	RefreshToken     string                `json:"refresh_token"`
	ExpiresAt        time.Time             `json:"expires_at"`
	RefreshExpiresAt time.Time             `json:"refresh_expires_at"`
	User             *storedAuthUserResult `json:"user"`
	Message          string                `json:"message"`
}

type storedAuthUserResult struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities"`
	Sections     []string `json:"sections"`
	Provider     string   `json:"provider"`
	Subject      string   `json:"subject"`
}

func fetchDaemonAuthStatus() daemonAuthStatus {
	resp, err := http.Get(DAEMON_URL + "/api/auth/status")
	if err != nil {
		return daemonAuthStatus{}
	}
	defer resp.Body.Close()
	var result daemonAuthStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return daemonAuthStatus{}
	}
	return result
}

func performBootstrapLogin(email, password string) (storedAuthSession, map[string]interface{}, error) {
	payload := authLoginRequest{Email: strings.TrimSpace(email), Password: password}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(DAEMON_URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return storedAuthSession{}, nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return storedAuthSession{}, nil, fmt.Errorf("failed to parse login response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok && msg != "" {
			return storedAuthSession{}, nil, errors.New(msg)
		}
		return storedAuthSession{}, nil, fmt.Errorf("login failed (%s)", resp.Status)
	}

	session := storedAuthSession{
		Token:           stringValue(result["token"]),
		RefreshToken:    stringValue(result["refresh_token"]),
		Email:           stringValueFromMap(result, "user", "email"),
		Name:            stringValueFromMap(result, "user", "name"),
		Role:            stringValueFromMap(result, "user", "role"),
		Mode:            stringValue(result["mode"]),
		Required:        boolValue(result["required"]),
		Capabilities:    stringSliceValue(result["capabilities"]),
		AllowedSections: stringSliceValue(result["allowed_sections"]),
	}
	if expiresAt, ok := result["expires_at"].(string); ok && expiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, expiresAt); err == nil {
			session.ExpiresAt = parsed
		}
	}
	if refreshExpiresAt, ok := result["refresh_expires_at"].(string); ok && refreshExpiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, refreshExpiresAt); err == nil {
			session.RefreshExpiresAt = parsed
		}
	}
	if err := saveStoredAuthSession(session); err != nil {
		fmt.Fprintln(os.Stderr, "⚠️  Login succeeded, but saving the session failed:", err)
	}
	return session, result, nil
}

func performOIDCLogin(emailHint string) (storedAuthSession, map[string]interface{}, error) {
	startURL := DAEMON_URL + "/api/auth/oidc/start"
	if strings.TrimSpace(emailHint) != "" {
		startURL += "?login_hint=" + url.QueryEscape(strings.TrimSpace(emailHint))
	}
	resp, err := http.Get(startURL)
	if err != nil {
		return storedAuthSession{}, nil, fmt.Errorf("failed to start oidc login: %w", err)
	}
	defer resp.Body.Close()

	var start oidcStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&start); err != nil {
		return storedAuthSession{}, nil, fmt.Errorf("failed to parse oidc start response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || !start.Success {
		return storedAuthSession{}, nil, fmt.Errorf("oidc login failed to start")
	}

	if err := openBrowserURL(start.AuthorizationURL); err != nil {
		fmt.Fprintln(os.Stderr, "⚠️  Could not open browser automatically:", err)
		fmt.Println("Open this URL to finish login:")
		fmt.Println(start.AuthorizationURL)
	}

	deadline := start.ExpiresAt
	if deadline.IsZero() {
		deadline = time.Now().UTC().Add(10 * time.Minute)
	}
	for time.Now().UTC().Before(deadline) {
		pollResp, err := http.Get(DAEMON_URL + "/api/auth/oidc/poll?state=" + url.QueryEscape(start.State))
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		var poll oidcPollResponse
		if err := json.NewDecoder(pollResp.Body).Decode(&poll); err != nil {
			pollResp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		pollResp.Body.Close()

		if poll.Ready && poll.Session != nil {
			session := storedAuthSession{
				Token:           poll.Session.Token,
				RefreshToken:    poll.Session.RefreshToken,
				Email:           poll.Session.User.Email,
				Name:            poll.Session.User.Name,
				Role:            poll.Session.User.Role,
				Mode:            "oidc",
				Required:        true,
				Capabilities:    poll.Session.User.Capabilities,
				AllowedSections: poll.Session.User.Sections,
				ExpiresAt:       poll.Session.ExpiresAt,
				RefreshExpiresAt: poll.Session.RefreshExpiresAt,
			}
			if err := saveStoredAuthSession(session); err != nil {
				fmt.Fprintln(os.Stderr, "⚠️  Login succeeded, but saving the session failed:", err)
			}
			return session, map[string]interface{}{
				"success":             true,
				"token":               session.Token,
				"refresh_token":       session.RefreshToken,
				"expires_at":          session.ExpiresAt.Format(time.RFC3339Nano),
				"refresh_expires_at":  session.RefreshExpiresAt.Format(time.RFC3339Nano),
				"mode":                session.Mode,
				"required":            session.Required,
				"user": map[string]interface{}{
					"email":        session.Email,
					"name":         session.Name,
					"role":         session.Role,
					"capabilities":  session.Capabilities,
					"sections":     session.AllowedSections,
				},
			}, nil
		}

		if poll.Message != "" {
			fmt.Fprintln(os.Stderr, poll.Message)
		}
		time.Sleep(2 * time.Second)
	}

	return storedAuthSession{}, nil, fmt.Errorf("oidc login timed out")
}

func printLoginResult(jsonOutput bool, session storedAuthSession, result map[string]interface{}) {
	if jsonOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return
	}

	role := session.Role
	if role == "" {
		role = "signed in"
	}
	fmt.Printf("✅ Logged in as %s (%s)\n", session.Email, role)
	if len(session.Capabilities) > 0 {
		fmt.Println("Capabilities:")
		for _, capability := range session.Capabilities {
			fmt.Printf("  - %s\n", capability)
		}
	}
	if len(session.AllowedSections) > 0 {
		fmt.Println("Sections:")
		for _, section := range session.AllowedSections {
			fmt.Printf("  - %s\n", section)
		}
	}
}

func openBrowserURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("url is empty")
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", raw).Run()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", raw).Run()
	default:
		return exec.Command("xdg-open", raw).Run()
	}
}

func handleAuthLogout() {
	token := loadStoredAuthToken()
	if token == "" {
		fmt.Println("No stored session found.")
		return
	}

	req, err := http.NewRequest(http.MethodPost, DAEMON_URL+"/api/auth/logout", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Failed to build request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		if msg, ok := result["error"].(string); ok && msg != "" {
			fmt.Fprintln(os.Stderr, "❌", msg)
			return
		}
		fmt.Fprintf(os.Stderr, "❌ logout failed (%s)\n", resp.Status)
		return
	}

	if err := clearStoredAuthSession(); err != nil {
		fmt.Fprintln(os.Stderr, "⚠️  Logged out on daemon, but could not clear local session:", err)
	} else {
		fmt.Println("✅ Logged out")
	}
}

func handleAuthStatus() {
	resp, err := http.Get(DAEMON_URL + "/api/auth/status")
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Failed to connect to daemon:", err)
		printLocalAuthStatus()
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintln(os.Stderr, "❌ Failed to parse auth status:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok && msg != "" {
			fmt.Fprintln(os.Stderr, "❌", msg)
			return
		}
		fmt.Fprintf(os.Stderr, "❌ auth status failed (%s)\n", resp.Status)
		return
	}

	prettyPrint(result)
}

func printLocalAuthStatus() {
	session, err := loadStoredAuthSession()
	if err != nil || session == nil {
		fmt.Println("Local auth session: none")
		return
	}

	fmt.Println("Local auth session:")
	fmt.Printf("  Email: %s\n", session.Email)
	fmt.Printf("  Role: %s\n", session.Role)
	fmt.Printf("  Mode: %s\n", session.Mode)
	fmt.Printf("  Expires: %s\n", session.ExpiresAt.Format(time.RFC3339))
	if len(session.Capabilities) > 0 {
		fmt.Printf("  Capabilities: %s\n", strings.Join(session.Capabilities, ", "))
	}
	if len(session.AllowedSections) > 0 {
		fmt.Printf("  Sections: %s\n", strings.Join(session.AllowedSections, ", "))
	}
}

func printAuthHelp() {
	fmt.Print(`
Usage: brain auth <subcommand>

Subcommands:
	login [--email <email>] [--password <password>] Log in against the daemon
  status                                        Show daemon and local auth status
  logout                                        Revoke the stored session
`)
}

func storedAuthSessionPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "brain", "auth.json")
	}
	return filepath.Join(home, ".config", "brain", "auth.json")
}

func saveStoredAuthSession(session storedAuthSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	if err := saveStoredAuthSessionSecure(string(data)); err == nil {
		return nil
	}
	if allowPlaintextFallback() {
		return saveStoredAuthSessionFile(session)
	}
	return fmt.Errorf("unable to save auth session securely")
}

func loadStoredAuthSession() (*storedAuthSession, error) {
	authSessionMu.Lock()
	defer authSessionMu.Unlock()

	var session *storedAuthSession
	var err error
	if session, err = loadStoredAuthSessionSecure(); err != nil {
		if allowPlaintextFallback() {
			session, err = loadStoredAuthSessionFile()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if session == nil {
		return nil, fmt.Errorf("no stored auth session")
	}

	if !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		if refreshed, refreshErr := refreshStoredAuthSession(session); refreshErr == nil {
			return refreshed, nil
		}
		_ = clearStoredAuthSession()
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

func clearStoredAuthSession() error {
	if err := clearStoredAuthSessionSecure(); err == nil {
		if allowPlaintextFallback() {
			_ = clearStoredAuthSessionFile()
		}
		return nil
	}
	if allowPlaintextFallback() {
		return clearStoredAuthSessionFile()
	}
	return nil
}

func loadStoredAuthToken() string {
	session, err := loadStoredAuthSession()
	if err != nil || session == nil {
		return ""
	}
	return session.Token
}

func saveStoredAuthSessionSecure(raw string) error {
	return keyring.Set(authKeyringService(), authKeyringAccount(), raw)
}

func loadStoredAuthSessionSecure() (*storedAuthSession, error) {
	raw, err := keyring.Get(authKeyringService(), authKeyringAccount())
	if err != nil {
		return nil, err
	}
	return unmarshalStoredAuthSession(raw)
}

func clearStoredAuthSessionSecure() error {
	if err := keyring.Delete(authKeyringService(), authKeyringAccount()); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}

func saveStoredAuthSessionFile(session storedAuthSession) error {
	path := storedAuthSessionPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadStoredAuthSessionFile() (*storedAuthSession, error) {
	path := storedAuthSessionPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return unmarshalStoredAuthSession(string(data))
}

func clearStoredAuthSessionFile() error {
	path := storedAuthSessionPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func unmarshalStoredAuthSession(raw string) (*storedAuthSession, error) {
	var session storedAuthSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, err
	}
	if !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		return &session, nil
	}
	return &session, nil
}

func allowPlaintextFallback() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("BRAIN_AUTH_ALLOW_PLAINTEXT_FALLBACK")))
	if value == "true" || value == "1" || value == "yes" {
		return true
	}
	return strings.ToLower(strings.TrimSpace(os.Getenv("BRAIN_ENV"))) == "development"
}

func authKeyringService() string {
	return "brain"
}

func authKeyringAccount() string {
	return "daemon-session"
}

func refreshStoredAuthSession(session *storedAuthSession) (*storedAuthSession, error) {
	if session == nil || strings.TrimSpace(session.RefreshToken) == "" {
		return nil, fmt.Errorf("refresh token unavailable")
	}
	if !session.RefreshExpiresAt.IsZero() && time.Now().UTC().After(session.RefreshExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	body, _ := json.Marshal(map[string]string{"refresh_token": session.RefreshToken})
	req, err := http.NewRequest(http.MethodPost, DAEMON_URL+"/api/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: http.DefaultTransport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result oidcSessionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || !result.Success {
		if result.Message != "" {
			return nil, errors.New(result.Message)
		}
		return nil, fmt.Errorf("refresh failed (%s)", resp.Status)
	}

	refreshed := &storedAuthSession{
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		Email:            result.User.Email,
		Name:             result.User.Name,
		Role:             result.User.Role,
		Mode:             session.Mode,
		Required:         session.Required,
		Capabilities:     result.User.Capabilities,
		AllowedSections:  result.User.Sections,
		ExpiresAt:        result.ExpiresAt,
		RefreshExpiresAt: result.RefreshExpiresAt,
	}
	if err := saveStoredAuthSession(*refreshed); err != nil {
		return nil, err
	}
	return refreshed, nil
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func stringValueFromMap(root map[string]interface{}, outerKey, innerKey string) string {
	if root == nil {
		return ""
	}
	inner, ok := root[outerKey].(map[string]interface{})
	if !ok {
		return ""
	}
	return stringValue(inner[innerKey])
}

func boolValue(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func stringSliceValue(value interface{}) []string {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func authWebSocketURL() string {
	parsed, err := url.Parse(DAEMON_WS)
	if err != nil {
		return DAEMON_WS
	}
	if token := loadStoredAuthToken(); token != "" {
		query := parsed.Query()
		query.Set("token", token)
		parsed.RawQuery = query.Encode()
	}
	return parsed.String()
}
