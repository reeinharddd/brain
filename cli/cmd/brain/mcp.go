package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type mcpServer struct {
	brainDir string
	out      *bufio.Writer
}

func runMCPServer(args []string) {
	flags := flag.NewFlagSet("mcp", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	brainDir := flags.String("brain-dir", "", "Path to the Brain repository")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	root := resolveBrainRoot()
	if strings.TrimSpace(*brainDir) != "" {
		root = strings.TrimSpace(*brainDir)
	}

	server := &mcpServer{brainDir: root, out: bufio.NewWriter(os.Stdout)}
	if err := server.serve(os.Stdin); err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (s *mcpServer) serve(input io.Reader) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp, shouldRespond := s.handleRequest(req)
		if !shouldRespond {
			continue
		}

		if err := s.writeResponse(resp); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (s *mcpServer) handleRequest(req mcpRequest) (mcpResponse, bool) {
	if req.Method == "notifications/initialized" {
		return mcpResponse{}, false
	}

	if len(req.ID) == 0 {
		return mcpResponse{}, false
	}

	base := mcpResponse{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		base.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "brain-mcp-server",
				"version": "1.0.0",
			},
		}
		return base, true
	case "tools/list":
		base.Result = map[string]any{"tools": brainMCPTools()}
		return base, true
	case "tools/call":
		result, err := s.callTool(req.Params)
		if err != nil {
			base.Error = &mcpError{Code: -32602, Message: err.Error()}
			return base, true
		}
		base.Result = map[string]any{
			"content": []map[string]any{{
				"type": "text",
				"text": result,
			}},
		}
		return base, true
	default:
		base.Error = &mcpError{Code: -32601, Message: "Unknown method: " + req.Method}
		return base, true
	}
}

func (s *mcpServer) writeResponse(resp mcpResponse) error {
	encoded, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	if _, err := s.out.Write(encoded); err != nil {
		return err
	}
	if err := s.out.WriteByte('\n'); err != nil {
		return err
	}
	return s.out.Flush()
}

func (s *mcpServer) callTool(params json.RawMessage) (string, error) {
	var payload struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &payload); err != nil {
		return "", err
	}

	switch payload.Name {
	case "brain_get_rules":
		return brainGetRules(s.brainDir, stringArg(payload.Arguments, "topic"), intArg(payload.Arguments, "max_chars", 2000)), nil
	case "brain_get_agent":
		return brainGetAgent(s.brainDir, stringArg(payload.Arguments, "name")), nil
	case "brain_list_agents":
		return brainListAgents(s.brainDir), nil
	case "brain_get_command":
		return brainGetCommand(s.brainDir, stringArg(payload.Arguments, "name")), nil
	case "brain_route_task":
		return brainRouteTask(stringArg(payload.Arguments, "task_description")), nil
	case "brain_search_rules":
		return brainSearchRules(s.brainDir, stringArg(payload.Arguments, "query")), nil
	case "brain_get_provider":
		return brainGetProvider(s.brainDir, stringArg(payload.Arguments, "task_type")), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", payload.Name)
	}
}

func brainMCPTools() []mcpTool {
	return []mcpTool{
		{
			Name:        "brain_get_rules",
			Description: "Get relevant rules from canonical.md and modules by topic",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"topic": map[string]any{
						"type":        "string",
						"description": "Topic to search rules for",
					},
					"max_chars": map[string]any{
						"type":        "integer",
						"description": "Max characters to return",
						"default":     2000,
					},
				},
				"required": []string{"topic"},
			},
		},
		{
			Name:        "brain_get_agent",
			Description: "Get the full definition of a brain repo agent by name",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "brain_list_agents",
			Description: "List all available brain repo agents with their descriptions",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			Name:        "brain_get_command",
			Description: "Get the definition of a brain repo slash command",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "brain_route_task",
			Description: "Get the suggested agent and model tier for a task description",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_description": map[string]any{"type": "string"},
				},
				"required": []string{"task_description"},
			},
		},
		{
			Name:        "brain_search_rules",
			Description: "Full-text search across all brain repo rules and modules",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "brain_get_provider",
			Description: "Get recommended model for a specific task type based on providers.yml routing",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_type": map[string]any{"type": "string"},
				},
				"required": []string{"task_type"},
			},
		},
	}
}

func brainGetRules(brainDir, topic string, maxChars int) string {
	topic = strings.TrimSpace(strings.ToLower(topic))
	if topic == "" {
		return "No rules found for empty topic"
	}

	paths := firstExistingPaths(
		filepath.Join(brainDir, "artifacts", "rules", "canonical.md"),
		filepath.Join(brainDir, "rules", "canonical.md"),
	)
	paths = append(paths, firstExistingGlob(
		filepath.Join(brainDir, "artifacts", "rules", "modules", "*.md"),
		filepath.Join(brainDir, "rules", "modules", "*.md"),
	)...)

	var results []string
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, section := range splitMarkdownSections(string(content)) {
			if strings.Contains(strings.ToLower(section), topic) {
				results = append(results, fmt.Sprintf("[%s]\n%s", filepath.Base(path), strings.TrimSpace(section)))
			}
		}
	}

	if len(results) == 0 {
		return "No rules found for topic: " + topic
	}

	combined := strings.Join(results, "\n\n---\n\n")
	if maxChars > 0 && len(combined) > maxChars {
		return combined[:maxChars]
	}
	return combined
}

func brainGetAgent(brainDir, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Agent name is required"
	}

	path, content, err := readFirstExistingFile(
		filepath.Join(brainDir, "artifacts", "agents", name+".md"),
		filepath.Join(brainDir, "agents", name+".md"),
	)
	if err != nil {
		entries := firstExistingGlob(
			filepath.Join(brainDir, "artifacts", "agents", "*.md"),
			filepath.Join(brainDir, "agents", "*.md"),
		)
		available := make([]string, 0, len(entries))
		for _, entry := range entries {
			available = append(available, strings.TrimSuffix(filepath.Base(entry), ".md"))
		}
		sort.Strings(available)
		if len(available) == 0 {
			return fmt.Sprintf("Agent '%s' not found. No agents available", name)
		}
		return fmt.Sprintf("Agent '%s' not found. Available: %s", name, strings.Join(available, ", "))
	}

	_ = path
	return string(content)
}

func brainListAgents(brainDir string) string {
	entries := firstExistingGlob(
		filepath.Join(brainDir, "artifacts", "agents", "*.md"),
		filepath.Join(brainDir, "agents", "*.md"),
	)
	if len(entries) == 0 {
		return "artifacts/agents/ directory not found"
	}

	sort.Strings(entries)
	lines := make([]string, 0, len(entries))
	for _, path := range entries {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		desc := extractFrontMatterField(string(content), "description")
		if desc != "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", strings.TrimSuffix(filepath.Base(path), ".md"), desc))
			continue
		}
		lines = append(lines, "- "+strings.TrimSuffix(filepath.Base(path), ".md"))
	}
	return strings.Join(lines, "\n")
}

func brainGetCommand(brainDir, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Command name is required"
	}

	_, content, err := readFirstExistingFile(
		filepath.Join(brainDir, "artifacts", "commands", name+".md"),
		filepath.Join(brainDir, "commands", name+".md"),
	)
	if err != nil {
		entries := firstExistingGlob(
			filepath.Join(brainDir, "artifacts", "commands", "*.md"),
			filepath.Join(brainDir, "commands", "*.md"),
		)
		available := make([]string, 0, len(entries))
		for _, entry := range entries {
			available = append(available, strings.TrimSuffix(filepath.Base(entry), ".md"))
		}
		sort.Strings(available)
		return fmt.Sprintf("Command '%s' not found. Available: %s", name, strings.Join(available, ", "))
	}

	return string(content)
}

func firstExistingPaths(paths ...string) []string {
	var existing []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	return existing
}

func firstExistingGlob(patterns ...string) []string {
	for _, pattern := range patterns {
		entries, err := filepath.Glob(pattern)
		if err == nil && len(entries) > 0 {
			return entries
		}
	}
	return nil
}

func readFirstExistingFile(paths ...string) (string, []byte, error) {
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return path, content, nil
		}
	}
	return "", nil, os.ErrNotExist
}

func brainRouteTask(taskDescription string) string {
	task := strings.ToLower(strings.TrimSpace(taskDescription))
	routing := []struct {
		keywords []string
		agent    string
		tier     string
	}{
		{[]string{"plan", "architect", "design", "spec", "roadmap"}, "planner", "powerful"},
		{[]string{"research", "investigate", "find", "look up"}, "researcher", "standard"},
		{[]string{"implement", "build", "code", "write", "create"}, "implementer", "standard"},
		{[]string{"debug", "fix", "error", "bug", "broken"}, "debugger", "standard"},
		{[]string{"review", "check", "audit", "approve"}, "reviewer", "standard"},
		{[]string{"refactor", "improve", "clean", "restructure"}, "refactor", "standard"},
		{[]string{"document", "readme", "comment", "adr"}, "documenter", "fast"},
		{[]string{"security", "vulnerability", "secret", "unsafe"}, "guardian", "standard"},
		{[]string{"ui", "design", "component", "style", "visual"}, "designer", "powerful"},
		{[]string{"orchestrate", "coordinate", "delegate", "complex"}, "orchestrator", "powerful"},
	}

	for _, rule := range routing {
		for _, keyword := range rule.keywords {
			if strings.Contains(task, keyword) {
				return mustJSON(map[string]any{
					"suggested_agent": rule.agent,
					"model_tier":      rule.tier,
					"reasoning":       fmt.Sprintf("Task contains keywords matching %s profile", rule.agent),
				})
			}
		}
	}

	return mustJSON(map[string]any{
		"suggested_agent": "orchestrator",
		"model_tier":      "standard",
		"reasoning":       "No specific pattern matched; defaulting to orchestrator",
	})
}

func brainSearchRules(brainDir, query string) string {
	queryTerms := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if len(queryTerms) == 0 {
		return "No matches for: " + query
	}

	paths := []string{filepath.Join(brainDir, "rules", "canonical.md")}
	modulePaths, _ := filepath.Glob(filepath.Join(brainDir, "rules", "modules", "*.md"))
	paths = append(paths, modulePaths...)

	var results []string
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			lineLower := strings.ToLower(line)
			matched := true
			for _, term := range queryTerms {
				if !strings.Contains(lineLower, term) {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}

			start := maxInt(0, i-1)
			end := minInt(len(lines), i+3)
			snippet := strings.Join(lines[start:end], "\n")
			results = append(results, fmt.Sprintf("[%s:%d] %s", filepath.Base(path), i+1, snippet))
		}
	}

	if len(results) == 0 {
		return "No matches for: " + query
	}

	if len(results) > 10 {
		results = results[:10]
	}
	return strings.Join(results, "\n\n")
}

func brainGetProvider(brainDir, taskType string) string {
	content, err := os.ReadFile(filepath.Join(brainDir, "providers", "providers.yml"))
	if err != nil {
		return "providers.yml not found"
	}

	routingTier := findTaskRoutingTier(string(content), taskType)
	if routingTier == "" {
		routingTier = "standard"
	}

	model := findFirstTierModel(string(content), routingTier)
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	return mustJSON(map[string]any{
		"task_type": taskType,
		"tier":      routingTier,
		"model":     model,
	})
}

func findTaskRoutingTier(content, taskType string) string {
	taskType = strings.ToLower(strings.TrimSpace(taskType))
	lines := strings.Split(content, "\n")
	inRouting := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inRouting {
			if trimmed == "task_routing:" {
				inRouting = true
			}
			continue
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.ToLower(strings.TrimSpace(parts[0])) == taskType {
			return strings.TrimSpace(strings.SplitN(parts[1], "#", 2)[0])
		}
	}

	return ""
}

func findFirstTierModel(content, tier string) string {
	tier = strings.TrimSpace(tier)
	if tier == "" {
		return ""
	}

	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(tier) + `:\s*(\S+)`)
	match := pattern.FindStringSubmatch(content)
	if len(match) >= 2 {
		return strings.TrimSpace(strings.SplitN(match[1], "#", 2)[0])
	}
	return ""
}

func splitMarkdownSections(content string) []string {
	lines := strings.Split(content, "\n")
	var sections []string
	var current strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "#") && current.Len() > 0 {
			sections = append(sections, current.String())
			current.Reset()
		}
		current.WriteString(line)
		current.WriteByte('\n')
	}

	if current.Len() > 0 {
		sections = append(sections, current.String())
	}

	return sections
}

func extractFrontMatterField(content, field string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return ""
	}

	rest := strings.TrimPrefix(content, "---")
	rest = strings.TrimLeft(rest, "\r\n")
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return ""
	}

	frontMatter := rest[:end]
	prefix := field + ":"
	for _, line := range strings.Split(frontMatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}

	return ""
}

func stringArg(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if value, ok := values[key]; ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(values map[string]any, key string, fallback int) int {
	if values == nil {
		return fallback
	}
	value, ok := values[key]
	if !ok {
		return fallback
	}

	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return fallback
	}
}

func mustJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `{"error":"failed to encode response"}`
	}
	return string(encoded)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
