package workers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type MCPWorker struct{}

type Registry struct {
	MCPs map[string]struct {
		Package string   `yaml:"package"`
		Command string   `yaml:"command"`
		Args    []string `yaml:"args"`
		Profile []string `yaml:"profile"`
	} `yaml:"mcps"`
}

func (w *MCPWorker) Sync(registryYamlPath string, outPath string, logger chan<- string) error {
	logger <- fmt.Sprintf("[MCPWorker] Compiling MCP JSON from %s", registryYamlPath)

	data, err := ioutil.ReadFile(registryYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read registry.yml: %v", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return fmt.Errorf("failed to parse registry.yml: %v", err)
	}

	mcpServers := make(map[string]interface{})

	for name, details := range reg.MCPs {
		inValidProfile := false
		for _, p := range details.Profile {
			if p == "standard" || p == "full" {
				inValidProfile = true
				break
			}
		}
		if !inValidProfile {
			continue
		}

		var cmd string
		var args []string

		if details.Command != "" {
			cmd = details.Command
			home, _ := os.UserHomeDir()
			for _, arg := range details.Args {
				args = append(args, strings.ReplaceAll(arg, "${HOME}", home))
			}
		} else if details.Package != "" {
			cmd = "bash"
			home, _ := os.UserHomeDir()
			argStr := fmt.Sprintf("-lc \"if command -v npx-nvm >/dev/null 2>&1; then NPX_BIN=$(command -v npx-nvm); else NPX_BIN=$(command -v npx); fi; exec \\\"$NPX_BIN\\\" -y %s", details.Package)

			if name == "memory" {
				argStr += fmt.Sprintf(" \\\"%s/.brain/memory\\\"", home)
			} else if name == "filesystem" {
				argStr += fmt.Sprintf(" \\\"%s\\\"", home)
			}
			argStr += "\""
			args = []string{"-c", argStr}
		} else {
			continue
		}

		mcpServers[name] = map[string]interface{}{
			"command": cmd,
			"args":    args,
		}
	}

	outputJSON := map[string]interface{}{}

	if _, err := os.Stat(outPath); err == nil {
		existingData, _ := ioutil.ReadFile(outPath)
		var existing map[string]interface{}
		if json.Unmarshal(existingData, &existing) == nil {
			if existingMcpServers, ok := existing["mcpServers"].(map[string]interface{}); ok {
				for k, v := range mcpServers {
					existingMcpServers[k] = v
				}
			} else {
				existing["mcpServers"] = mcpServers
			}
			outputJSON = existing
		}
	} else {
		outputJSON["mcpServers"] = mcpServers
	}

	b, _ := json.MarshalIndent(outputJSON, "", "  ")
	os.MkdirAll(filepath.Dir(outPath), 0755)

	if err := ioutil.WriteFile(outPath, b, 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %v", err)
	}

	logger <- fmt.Sprintf("[MCPWorker] Wrote MCP config successfully to %s", outPath)
	return nil
}
