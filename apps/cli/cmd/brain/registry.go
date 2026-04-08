package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
)

// HandleSkillsCommand handles 'brain skills' subcommands
func HandleSkillsCommand(args []string) {
	if len(args) < 3 {
		printSkillsHelp()
		return
	}

	subcommand := args[2]

	switch subcommand {
	case "list":
		skillsList()
	case "search":
		if len(args) < 4 {
			fmt.Println("Usage: brain skills search <query>")
			return
		}
		skillsSearch(args[3])
	case "info":
		if len(args) < 4 {
			fmt.Println("Usage: brain skills info <id>")
			return
		}
		skillsInfo(args[3])
	case "sync":
		skillsSync()
	case "validate":
		skillsValidate()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printSkillsHelp()
	}
}

func skillsList() {
	resp, err := http.Get(DAEMON_URL + "/api/skills")
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	skills, ok := result["skills"].([]interface{})
	if !ok {
		fmt.Println("❌ Invalid response format")
		return
	}

	if len(skills) == 0 {
		fmt.Println("ℹ️  No skills available")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tVERSION\tTYPE\tTAGS")
	fmt.Fprintln(w, "---\t---\t---\t---\t---")

	for _, skill := range skills {
		skillMap, ok := skill.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%v", skillMap["id"])
		name := fmt.Sprintf("%v", skillMap["name"])
		version := fmt.Sprintf("%v", skillMap["version"])
		typ := fmt.Sprintf("%v", skillMap["type"])
		tags := formatTags(skillMap["tags"])

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, name, version, typ, tags)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d skills\n", len(skills))
}

func skillsSearch(query string) {
	resp, err := http.Post(DAEMON_URL+"/api/skills/search?q="+query, "application/json", nil)
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	results, ok := result["results"].([]interface{})
	if !ok {
		fmt.Println("ℹ️  No results found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tVERSION\tDESCRIPTION")
	fmt.Fprintln(w, "---\t---\t---\t---")

	for _, skill := range results {
		skillMap, ok := skill.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%v", skillMap["id"])
		name := fmt.Sprintf("%v", skillMap["name"])
		version := fmt.Sprintf("%v", skillMap["version"])
		desc := fmt.Sprintf("%v", skillMap["description"])

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, version, desc)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d results\n", len(results))
}

func skillsInfo(id string) {
	resp, err := http.Get(DAEMON_URL + "/api/skills/" + id)
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Skill not found: %s\n", id)
		return
	}

	var skill map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&skill); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Printf("📚 Skill: %v\n", skill["name"])
	fmt.Printf("   ID:      %v\n", skill["id"])
	fmt.Printf("   Version: %v\n", skill["version"])
	fmt.Printf("   Type:    %v\n", skill["type"])
	fmt.Printf("   Desc:    %v\n", skill["description"])
	fmt.Printf("   Tags:    %s\n", formatTags(skill["tags"]))
	fmt.Printf("   File:    %v\n", skill["file"])
}

func skillsSync() {
	resp, err := http.Post(DAEMON_URL+"/api/skills/sync", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Println("✅", result["status"])
}

func skillsValidate() {
	resp, err := http.Get(DAEMON_URL + "/api/skills/validate")
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	synced, ok := result["synced"].(bool)
	if !ok {
		fmt.Println("❌ Invalid response format")
		return
	}

	if synced {
		fmt.Println("✅ Skills perfectly synchronized")
		return
	}

	// Show orphans if any
	orphans, ok := result["orphans"].([]interface{})
	if ok && len(orphans) > 0 {
		fmt.Printf("❌ Found %d orphan skills (in filesystem but not in registry):\n", len(orphans))
		for _, orphan := range orphans {
			fmt.Printf("   - %v\n", orphan)
		}
	}

	// Show missing if any
	missing, ok := result["missing"].([]interface{})
	if ok && len(missing) > 0 {
		fmt.Printf("⚠️  Found %d missing skills (in registry but not in filesystem):\n", len(missing))
		for _, missing_skill := range missing {
			fmt.Printf("   - %v\n", missing_skill)
		}
	}
}

func printSkillsHelp() {
	fmt.Println(`
Usage: brain skills <subcommand> [options]

Subcommands:
  list              List all available skills
  search <query>    Search skills by keyword
  info <id>         Show detailed info about a skill
  sync              Sync skills registry
  validate          Validate skills synchronization status

Examples:
  brain skills list
  brain skills search refactoring
  brain skills info code-refactoring
  brain skills sync
  brain skills validate
	`)
}

// HandleMCPsCommand handles 'brain mcps' subcommands
func HandleMCPsCommand(args []string) {
	if len(args) < 3 {
		printMCPsHelp()
		return
	}

	subcommand := args[2]

	switch subcommand {
	case "list":
		mcpsList()
	case "search":
		if len(args) < 4 {
			fmt.Println("Usage: brain mcps search <query>")
			return
		}
		mcpsSearch(args[3])
	case "info":
		if len(args) < 4 {
			fmt.Println("Usage: brain mcps info <id>")
			return
		}
		mcpsInfo(args[3])
	case "sync":
		mcpsSync()
	case "status":
		mcpsStatus()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printMCPsHelp()
	}
}

func mcpsList() {
	resp, err := http.Get(DAEMON_URL + "/api/mcps")
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	mcps, ok := result["mcps"].([]interface{})
	if !ok {
		fmt.Println("❌ Invalid response format")
		return
	}

	if len(mcps) == 0 {
		fmt.Println("ℹ️  No MCPs available")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tVERSION\tTYPE\tFEATURES")
	fmt.Fprintln(w, "---\t---\t---\t---\t---")

	for _, mcp := range mcps {
		mcpMap, ok := mcp.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%v", mcpMap["id"])
		name := fmt.Sprintf("%v", mcpMap["name"])
		version := fmt.Sprintf("%v", mcpMap["version"])
		typ := fmt.Sprintf("%v", mcpMap["type"])
		features := formatTags(mcpMap["features"])

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, name, version, typ, features)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d MCPs\n", len(mcps))
}

func mcpsSearch(query string) {
	resp, err := http.Post(DAEMON_URL+"/api/mcps/search?q="+query, "application/json", nil)
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	results, ok := result["results"].([]interface{})
	if !ok {
		fmt.Println("ℹ️  No results found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tVERSION\tDESCRIPTION")
	fmt.Fprintln(w, "---\t---\t---\t---")

	for _, mcp := range results {
		mcpMap, ok := mcp.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%v", mcpMap["id"])
		name := fmt.Sprintf("%v", mcpMap["name"])
		version := fmt.Sprintf("%v", mcpMap["version"])
		desc := fmt.Sprintf("%v", mcpMap["description"])

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, version, desc)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d results\n", len(results))
}

func mcpsInfo(id string) {
	resp, err := http.Get(DAEMON_URL + "/api/mcps/" + id)
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ MCP not found: %s\n", id)
		return
	}

	var mcp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mcp); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Printf("📡 MCP: %v\n", mcp["name"])
	fmt.Printf("   ID:    %v\n", mcp["id"])
	fmt.Printf("   Vers:  %v\n", mcp["version"])
	fmt.Printf("   Type:  %v\n", mcp["type"])
	fmt.Printf("   Desc:  %v\n", mcp["description"])
	fmt.Printf("   Req:   %v\n", mcp["required"])
	fmt.Printf("   Feats: %s\n", formatTags(mcp["features"]))
}

func mcpsSync() {
	resp, err := http.Post(DAEMON_URL+"/api/mcps/sync", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Println("✅", result["status"])
}

func mcpsStatus() {
	resp, err := http.Get(DAEMON_URL + "/api/mcps")
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Println("🟢 MCPs Registry Status")
	fmt.Printf("   Total MCPs: %v\n", result)
}

func printMCPsHelp() {
	fmt.Println(`
Usage: brain mcps <subcommand> [options]

Subcommands:
  list              List all MCPs
  search <query>    Search MCPs by feature
  info <id>         Show detailed info about an MCP
  status            Show MCPs registry status
  sync              Sync MCPs registry

Examples:
  brain mcps list
  brain mcps search github
  brain mcps info github
  brain mcps sync
	`)
}

// HandleAgentsCommand handles 'brain agents' subcommands
func HandleAgentsCommand(args []string) {
	if len(args) < 3 {
		printAgentsHelp()
		return
	}

	subcommand := args[2]

	switch subcommand {
	case "list":
		agentsList()
	case "search":
		if len(args) < 4 {
			fmt.Println("Usage: brain agents search <query>")
			return
		}
		agentsSearch(args[3])
	case "info":
		if len(args) < 4 {
			fmt.Println("Usage: brain agents info <id>")
			return
		}
		agentsInfo(args[3])
	case "sync":
		agentsSync()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printAgentsHelp()
	}
}

func agentsList() {
	resp, err := http.Get(DAEMON_URL + "/api/agents")
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	agents, ok := result["agents"].([]interface{})
	if !ok {
		fmt.Println("❌ Invalid response format")
		return
	}

	if len(agents) == 0 {
		fmt.Println("ℹ️  No agents available")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tMODEL\tDESCRIPTION")
	fmt.Fprintln(w, "---\t---\t---\t---")

	for _, agent := range agents {
		agentMap, ok := agent.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%v", agentMap["id"])
		name := fmt.Sprintf("%v", agentMap["name"])
		model := fmt.Sprintf("%v", agentMap["model"])
		desc := fmt.Sprintf("%v", agentMap["description"])

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, model, desc)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d agents\n", len(agents))
}

func agentsSearch(query string) {
	resp, err := http.Post(DAEMON_URL+"/api/agents/search?q="+query, "application/json", nil)
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	results, ok := result["results"].([]interface{})
	if !ok {
		fmt.Println("ℹ️  No results found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tMODEL\tDESCRIPTION")
	fmt.Fprintln(w, "---\t---\t---\t---")

	for _, agent := range results {
		agentMap, ok := agent.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%v", agentMap["id"])
		name := fmt.Sprintf("%v", agentMap["name"])
		model := fmt.Sprintf("%v", agentMap["model"])
		desc := fmt.Sprintf("%v", agentMap["description"])

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, model, desc)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d results\n", len(results))
}

func agentsInfo(id string) {
	resp, err := http.Get(DAEMON_URL + "/api/agents/" + id)
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Agent not found: %s\n", id)
		return
	}

	var agent map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Printf("🤖 Agent: %v\n", agent["name"])
	fmt.Printf("   ID:    %v\n", agent["id"])
	fmt.Printf("   Model: %v\n", agent["model"])
	fmt.Printf("   Desc:  %v\n", agent["description"])
	fmt.Printf("   Temp:  %v\n", agent["temperature"])
}

func agentsSync() {
	resp, err := http.Post(DAEMON_URL+"/api/agents/sync", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("❌ Failed to parse response:", err)
		return
	}

	fmt.Println("✅", result["status"])
}

func printAgentsHelp() {
	fmt.Println(`
Usage: brain agents <subcommand> [options]

Subcommands:
  list              List all available agents
  search <query>    Search agents by keyword
  info <id>         Show detailed info about an agent
  sync              Sync agents registry

Examples:
  brain agents list
  brain agents search debugging
  brain agents info debugger
  brain agents sync
	`)
}

// Helper function to format tags
func formatTags(tagsIface interface{}) string {
	tags, ok := tagsIface.([]interface{})
	if !ok {
		return ""
	}

	var result []string
	for _, tag := range tags {
		if t, ok := tag.(string); ok {
			result = append(result, t)
		}
	}

	if len(result) == 0 {
		return ""
	}

	output := ""
	for i, tag := range result {
		if i > 0 {
			output += ", "
		}
		output += tag
	}
	return output
}
