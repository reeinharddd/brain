package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/reeinharrrd/brain/mcp/docs-rag-mcp/internal/indexer"
	"github.com/reeinharrrd/brain/mcp/docs-rag-mcp/internal/tools"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "search":
		cmdSearch(os.Args[2:])
	case "status":
		cmdStatus()
	case "rebuild":
		cmdRebuild()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	limit := fs.Int("limit", 10, "Max results to return")
	domain := fs.String("domain", "", "Optional domain filter")
	json_ := fs.Bool("json", false, "Output JSON format")

	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintf(os.Stderr, "Error: query required\n")
		fmt.Fprintf(os.Stderr, "Usage: brain docs-rag search [flags] <query>\n")
		os.Exit(1)
	}

	query := strings.Join(remaining, " ")

	// Get brain root and create indexer
	brainRoot := os.Getenv("BRAIN_ROOT")
	if brainRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		brainRoot = homeDir + "/.brain"
	}

	idx, err := indexer.NewIndexer(brainRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating indexer: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Perform search
	req := tools.SearchRequest{
		Query:  query,
		Limit:  *limit,
		Domain: *domain,
	}

	resp := tools.DocsSearch(ctx, idx, req)

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "Search error: %s\n", resp.Error)
		os.Exit(1)
	}

	if *json_ {
		printJSON(resp)
	} else {
		printSearchResults(resp)
	}
}

func cmdStatus() {
	brainRoot := os.Getenv("BRAIN_ROOT")
	if brainRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		brainRoot = homeDir + "/.brain"
	}

	idx, err := indexer.NewIndexer(brainRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating indexer: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp := tools.DocsStatus(ctx, idx)

	fmt.Printf("Documentation Index Status\n")
	fmt.Printf("--------------------------\n")
	fmt.Printf("State:              %s\n", resp.IndexStatus.State)
	fmt.Printf("Document Count:     %d\n", resp.IndexStatus.DocumentCount)
	fmt.Printf("Last Rebuild:       %s\n", resp.IndexStatus.LastRebuildTime)
	fmt.Printf("Qdrant Health:      %s\n", resp.IndexStatus.QdrantHealth)

	if len(resp.IndexStatus.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range resp.IndexStatus.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
}

func cmdRebuild() {
	brainRoot := os.Getenv("BRAIN_ROOT")
	if brainRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		brainRoot = homeDir + "/.brain"
	}

	brainEnv := os.Getenv("BRAIN_ENV")
	if brainEnv == "" {
		brainEnv = "development"
	}

	idx, err := indexer.NewIndexer(brainRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating indexer: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("Rebuilding documentation index...\n")

	resp := tools.DocsRebuild(ctx, idx, brainEnv)

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "Rebuild error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Rebuild complete!\n")
	fmt.Printf("  Documents indexed: %d\n", resp.DocumentCount)
	fmt.Printf("  Duration: %s\n", resp.Duration)
}

func printSearchResults(resp tools.SearchResponse) {
	fmt.Printf("Search Results (%d found)\n", resp.Metadata.ResultsCount)
	fmt.Printf("===========================\n\n")

	for i, result := range resp.Results {
		fmt.Printf("%d. %s\n", i+1, result.Title)
		fmt.Printf("   Path:     %s\n", result.Path)
		fmt.Printf("   Category: %s\n", result.Category)
		fmt.Printf("   Priority: %s\n", result.RAGPriority)
		fmt.Printf("   Score:    %.2f\n", result.Score)
		if result.Snippet != "" {
			fmt.Printf("   Snippet:  %s...\n", result.Snippet[:min(len(result.Snippet), 100)])
		}
		fmt.Printf("\n")
	}

	fmt.Printf("Metadata:\n")
	fmt.Printf("  Total Indexed: %d\n", resp.Metadata.TotalIndexed)
	fmt.Printf("  Index Status:  %s\n", resp.Metadata.IndexStatus)
}

func printJSON(resp tools.SearchResponse) {
	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Brain Docs-RAG CLI

Commands:
  search [flags] <query>    Search documentation
    -limit N                Max results (default: 10)
    -domain DOMAIN          Filter by domain
    -json                   Output JSON format

  status                    Show index status

  rebuild                   Rebuild the index (dev-only)

  help                      Show this help message

Examples:
  brain docs-rag search authentication
  brain docs-rag search -limit 5 "daemon architecture"
  brain docs-rag search -domain architecture "MCP server"
  brain docs-rag status
  brain docs-rag rebuild

`)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
