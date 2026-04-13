package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"strings"
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
	case "add", "install":
		skillsInstall(args[3:])
	case "edit", "update":
		skillsUpdate(args[3:])
	case "remove", "delete":
		skillsRemove(args[3:])
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
	fmt.Printf("   Scope:   %v\n", skill["scope"])
	fmt.Printf("   Version: %v\n", skill["version"])
	fmt.Printf("   Type:    %v\n", skill["type"])
	fmt.Printf("   Desc:    %v\n", skill["description"])
	fmt.Printf("   Tags:    %s\n", formatTags(skill["tags"]))
	fmt.Printf("   File:    %v\n", skill["file"])
	if sourceURI := fmt.Sprintf("%v", skill["source_uri"]); sourceURI != "" && sourceURI != "<nil>" {
		fmt.Printf("   Source:  %v\n", sourceURI)
	}
	if sourceType := fmt.Sprintf("%v", skill["source_type"]); sourceType != "" && sourceType != "<nil>" {
		fmt.Printf("   Source Type: %v\n", sourceType)
	}
	if sourceVariant := fmt.Sprintf("%v", skill["source_variant"]); sourceVariant != "" && sourceVariant != "<nil>" {
		fmt.Printf("   Variant:  %v\n", sourceVariant)
	}
	if artifactPath := fmt.Sprintf("%v", skill["artifact_path"]); artifactPath != "" && artifactPath != "<nil>" {
		fmt.Printf("   Artifact: %v\n", artifactPath)
	}
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

type skillInstallCLIOptions struct {
	Source          string
	SourceType      string
	Scope           string
	Skills          []string
	InstallAll      bool
	PreviewOnly     bool
	IncludeInternal bool
	JSON            bool
}

type skillUpdateCLIOptions struct {
	ID            string
	Name          string
	Description   string
	Scope         string
	Version       string
	Type          string
	File          string
	SourceURI     string
	SourceType    string
	SourceVariant string
	ArtifactPath  string
	Category      string
	Tags          []string
	SyncTo        []string
	Maintained    *bool
}

func skillsInstall(args []string) {
	opts, err := parseSkillInstallArgs(args)
	if err != nil {
		fmt.Println("❌", err)
		return
	}
	if opts.Source == "" {
		fmt.Println("Usage: brain skills add <source> [--skill <name>] [--all] [--scope global|user|workspace|project] [--preview]")
		return
	}

	previewReq := map[string]interface{}{
		"source":            opts.Source,
		"source_type":       opts.SourceType,
		"scope":             opts.Scope,
		"skills":            opts.Skills,
		"install_all":       opts.InstallAll,
		"include_internal":  opts.IncludeInternal,
		"copy":              true,
	}
	if opts.PreviewOnly || len(opts.Skills) == 0 {
		preview, err := postSkillsInstallPreview(previewReq)
		if err != nil {
			fmt.Println("❌", err)
			return
		}
		printSkillsInstallPreview(preview)
		if opts.PreviewOnly {
			return
		}
		if len(opts.Skills) == 0 {
			if requiresSelection(preview) {
				fmt.Println("ℹ️  Re-run with --skill <name> (repeatable) or --all to continue.")
				return
			}
		}
	}

	result, err := postSkillsInstall(map[string]interface{}{
		"source":            opts.Source,
		"source_type":       opts.SourceType,
		"scope":             opts.Scope,
		"skills":            opts.Skills,
		"install_all":       opts.InstallAll,
		"include_internal":  opts.IncludeInternal,
		"copy":              true,
	})
	if err != nil {
		fmt.Println("❌", err)
		return
	}

	installed, _ := result["installed"].([]interface{})
	fmt.Printf("✅ Installed %d skill(s)\n", len(installed))
	for _, entry := range installed {
		if item, ok := entry.(map[string]interface{}); ok {
			fmt.Printf("  - %v (%v)\n", item["name"], item["id"])
		}
	}
	if opts.JSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	}
}

func skillsRemove(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: brain skills remove <id>")
		return
	}
	id := args[0]
	req, err := http.NewRequest(http.MethodDelete, DAEMON_URL+"/api/skills/"+id, nil)
	if err != nil {
		fmt.Println("❌ Failed to build request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
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

	if resp.StatusCode != http.StatusOK {
		fmt.Println("❌", result["error"])
		return
	}

	fmt.Printf("✅ Deleted skill %s\n", id)
}

func skillsUpdate(args []string) {
	opts, err := parseSkillUpdateArgs(args)
	if err != nil {
		fmt.Println("❌", err)
		return
	}
	if opts.ID == "" {
		fmt.Println("Usage: brain skills update <id> [--name <name>] [--description <text>] [--scope <scope>] [--tag <tag>] [--sync-to <target>]")
		return
	}

	current, err := fetchSkillItem(opts.ID)
	if err != nil {
		fmt.Println("❌", err)
		return
	}

	patchSkillItem(current, opts)
	body, _ := json.Marshal(current)
	req, err := http.NewRequest(http.MethodPut, DAEMON_URL+"/api/skills/"+opts.ID, bytes.NewReader(body))
	if err != nil {
		fmt.Println("❌ Failed to build request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
	if resp.StatusCode != http.StatusOK {
		fmt.Println("❌", result["error"])
		return
	}

	fmt.Printf("✅ Updated skill %s\n", opts.ID)
}

func parseSkillInstallArgs(args []string) (*skillInstallCLIOptions, error) {
	opts := &skillInstallCLIOptions{Scope: "global"}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--scope", "-s":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--scope requires a value")
			}
			opts.Scope = args[i+1]
			i++
		case "--source-type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--source-type requires a value")
			}
			opts.SourceType = args[i+1]
			i++
		case "--skill", "-k":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--skill requires a value")
			}
			opts.Skills = append(opts.Skills, args[i+1])
			i++
		case "--all":
			opts.InstallAll = true
		case "--preview":
			opts.PreviewOnly = true
		case "--include-internal":
			opts.IncludeInternal = true
		case "--json":
			opts.JSON = true
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			if opts.Source == "" {
				opts.Source = arg
			} else {
				opts.Skills = append(opts.Skills, arg)
			}
		}
	}
	return opts, nil
}

func parseSkillUpdateArgs(args []string) (*skillUpdateCLIOptions, error) {
	opts := &skillUpdateCLIOptions{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--name":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--name requires a value")
			}
			opts.Name = args[i+1]
			i++
		case "--description", "--desc":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--description requires a value")
			}
			opts.Description = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--scope requires a value")
			}
			opts.Scope = args[i+1]
			i++
		case "--version":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--version requires a value")
			}
			opts.Version = args[i+1]
			i++
		case "--type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--type requires a value")
			}
			opts.Type = args[i+1]
			i++
		case "--file":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--file requires a value")
			}
			opts.File = args[i+1]
			i++
		case "--source-uri":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--source-uri requires a value")
			}
			opts.SourceURI = args[i+1]
			i++
		case "--source-type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--source-type requires a value")
			}
			opts.SourceType = args[i+1]
			i++
		case "--source-variant":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--source-variant requires a value")
			}
			opts.SourceVariant = args[i+1]
			i++
		case "--artifact-path":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--artifact-path requires a value")
			}
			opts.ArtifactPath = args[i+1]
			i++
		case "--category":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--category requires a value")
			}
			opts.Category = args[i+1]
			i++
		case "--tag":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--tag requires a value")
			}
			opts.Tags = append(opts.Tags, args[i+1])
			i++
		case "--sync-to":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--sync-to requires a value")
			}
			opts.SyncTo = append(opts.SyncTo, args[i+1])
			i++
		case "--maintained":
			value := true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				lower := strings.ToLower(args[i+1])
				value = lower != "false" && lower != "0" && lower != "no"
				i++
			}
			opts.Maintained = &value
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			if opts.ID == "" {
				opts.ID = arg
			} else {
				return nil, fmt.Errorf("unexpected argument: %s", arg)
			}
		}
	}
	return opts, nil
}

func fetchSkillItem(id string) (map[string]interface{}, error) {
	resp, err := http.Get(DAEMON_URL + "/api/skills/" + id)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer resp.Body.Close()

	var item map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to parse skill response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skill not found: %s", id)
	}
	return item, nil
}

func patchSkillItem(item map[string]interface{}, opts *skillUpdateCLIOptions) {
	if opts.Name != "" {
		item["name"] = opts.Name
	}
	if opts.Description != "" {
		item["description"] = opts.Description
	}
	if opts.Scope != "" {
		item["scope"] = opts.Scope
	}
	if opts.Version != "" {
		item["version"] = opts.Version
	}
	if opts.Type != "" {
		item["type"] = opts.Type
	}
	if opts.File != "" {
		item["file"] = opts.File
		item["path"] = opts.File
	}
	if opts.SourceURI != "" {
		item["source_uri"] = opts.SourceURI
	}
	if opts.SourceType != "" {
		item["source_type"] = opts.SourceType
	}
	if opts.SourceVariant != "" {
		item["source_variant"] = opts.SourceVariant
	}
	if opts.ArtifactPath != "" {
		item["artifact_path"] = opts.ArtifactPath
	}
	if opts.Category != "" {
		item["category"] = opts.Category
	}
	if len(opts.Tags) > 0 {
		item["tags"] = opts.Tags
	}
	if len(opts.SyncTo) > 0 {
		item["sync_to"] = opts.SyncTo
	}
	if opts.Maintained != nil {
		item["maintained"] = *opts.Maintained
	}
}

func postSkillsInstallPreview(payload map[string]interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(payload)
	resp, err := http.Post(DAEMON_URL+"/api/skills/install/preview", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse preview response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("preview failed: %v", result["error"])
	}
	return result, nil
}

func postSkillsInstall(payload map[string]interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(payload)
	resp, err := http.Post(DAEMON_URL+"/api/skills/install", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse install response: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("install failed: %v", result["error"])
	}
	return result, nil
}

func printSkillsInstallPreview(preview map[string]interface{}) {
	fmt.Println("🧪 Install preview:")
	fmt.Printf("  Source: %v\n", preview["source"])
	fmt.Printf("  Type:   %v\n", preview["source_type"])
	fmt.Printf("  Scope:  %v\n", preview["scope"])

	available, _ := preview["available"].([]interface{})
	if len(available) == 0 {
		fmt.Println("  No installable skills discovered")
		return
	}

	fmt.Println("  Available skills:")
	for _, entry := range available {
		item, _ := entry.(map[string]interface{})
		fmt.Printf("    - %v: %v\n", item["name"], item["description"])
	}
}

func requiresSelection(preview map[string]interface{}) bool {
	value, _ := preview["requires_selection"].(bool)
	return value
}

func printSkillsHelp() {
	fmt.Println(`
Usage: brain skills <subcommand> [options]

Subcommands:
  list              List all available skills
  search <query>    Search skills by keyword
  info <id>         Show detailed info about a skill
	add <source>      Install a skill source into Brain
	update <id>       Modify a skill in the registry
	remove <id>       Delete a skill from the registry
  sync              Sync skills registry
  validate          Validate skills synchronization status

Examples:
  brain skills list
  brain skills search refactoring
  brain skills info code-refactoring
	brain skills add https://github.com/org/repo --skill my-skill
	brain skills update my-skill --description "Better docs"
	brain skills remove code-refactoring
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
