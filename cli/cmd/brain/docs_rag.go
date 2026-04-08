package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// HandleDocsRagCommand handles 'brain docs-rag' subcommands
func HandleDocsRagCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: brain docs-rag <command>")
		fmt.Println("  search <query>   - Search documentation")
		fmt.Println("  status          - Show docs search status")
		fmt.Println("  rebuild         - Rebuild documentation index")
		return
	}

	switch args[2] {
	case "search":
		cmdDocsRagSearch(args[3:])
	case "status":
		cmdDocsRagStatus()
	case "rebuild":
		fmt.Println("ℹ️  Rebuilding documentation index (via daemon)...")
		resp, err := http.Post(DAEMON_URL+"/api/docs-rag/rebuild", "application/json", nil)
		if err != nil {
			fmt.Println("❌ Failed to connect to daemon:", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			fmt.Println("✓ Rebuild triggered")
		} else {
			fmt.Println("❌ Rebuild failed:", resp.Status)
		}
	case "help", "-h", "--help":
		fmt.Println("Usage: brain docs-rag <command>")
		fmt.Println("  search <query>   - Search documentation")
		fmt.Println("  status          - Show docs search status")
		fmt.Println("  rebuild         - Rebuild documentation index")
	default:
		fmt.Printf("Unknown command: %s\n", args[2])
	}
}

func cmdDocsRagSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	limit := fs.Int("limit", 10, "Max results to return")
	domain := fs.String("domain", "", "Optional domain filter")
	jsonOutput := fs.Bool("json", false, "Output JSON format")
	fs.SetOutput(os.Stderr)

	err := fs.Parse(args)
	if err != nil {
		return
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintf(os.Stderr, "Error: query required\n")
		fmt.Fprintf(os.Stderr, "Usage: brain docs-rag search [flags] <query>\n")
		return
	}

	query := strings.Join(remaining, " ")

	// Build request to daemon
	searchParams := map[string]interface{}{
		"query": query,
		"limit": limit,
	}
	if *domain != "" {
		searchParams["domain"] = domain
	}
	body, _ := json.Marshal(searchParams)

	resp, err := http.Post(DAEMON_URL+"/api/docs-rag/search", "application/json", strings.NewReader(string(body)))
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if *jsonOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Search results for: %q\n", query)
		if m, ok := result.(map[string]interface{}); ok {
			if results, ok := m["results"].([]interface{}); ok {
				for i, r := range results {
					fmt.Printf("%d. %v\n", i+1, r)
				}
			}
		}
	}
}

func cmdDocsRagStatus() {
	resp, err := http.Get(DAEMON_URL + "/api/docs-rag/status")
	if err != nil {
		fmt.Println("❌ Failed to connect to daemon:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}
