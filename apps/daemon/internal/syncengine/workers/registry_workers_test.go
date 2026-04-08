package workers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSkillsWorkerSyncWritesCanonicalJSON(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "out", "skills.json")

	worker := &SkillsWorker{}
	logCh := make(chan string, 10)
	catalog := []*CatalogItem{
		{
			ID:          "alpha",
			Name:        "Alpha",
			Kind:        "skill",
			Scope:       "global",
			Description: "Alpha skill",
			Type:        "internal",
			Maintained:  true,
			Source:      "registry.yml",
		},
	}

	if err := worker.Sync(tempDir, outputPath, logCh, func() []*CatalogItem { return catalog }); err != nil {
		t.Fatalf("sync skills: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	skills, ok := payload["skills"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected skills map in output, got %#v", payload["skills"])
	}

	alpha, ok := skills["alpha"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected alpha skill entry, got %#v", skills["alpha"])
	}

	if got := alpha["type"]; got != "internal" {
		t.Fatalf("expected type internal, got %#v", got)
	}
}

func TestMCPWorkerSyncWritesMCPServersJSON(t *testing.T) {
	tempDir := t.TempDir()
	registryPath := filepath.Join(tempDir, "registry.yml")
	outputPath := filepath.Join(tempDir, "out", "mcps.json")

	registryYAML := `mcps:
  memory:
    package: "@modelcontextprotocol/server-memory"
    profile: [standard, full]
    required: true
    purpose: Memory backend
    setup: "npx -y @modelcontextprotocol/server-memory"
`
	if err := os.WriteFile(registryPath, []byte(registryYAML), 0644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	worker := &MCPWorker{}
	logCh := make(chan string, 10)
	if err := worker.Sync(registryPath, outputPath, logCh); err != nil {
		t.Fatalf("sync mcps: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	servers, ok := payload["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected mcpServers map in output, got %#v", payload["mcpServers"])
	}

	memory, ok := servers["memory"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected memory server entry, got %#v", servers["memory"])
	}

	if got := memory["command"]; got != "bunx" {
		t.Fatalf("expected command bunx, got %#v", got)
	}
}

func TestAgentsWorkerSyncParsesFrontMatter(t *testing.T) {
	tempDir := t.TempDir()
	agentsDir := filepath.Join(tempDir, "agents")
	outputPath := filepath.Join(tempDir, "out", "agents.json")

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentMarkdown := `---
name: architect
description: Designs technical solutions and trade-offs
version: 2.1.0
model: claude-opus
temperature: 0.4
tags:
  - architecture
  - planning
maintained: true
---

# Architect Agent

Guidance text for the architect agent.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "architect.md"), []byte(agentMarkdown), 0644); err != nil {
		t.Fatalf("write agent markdown: %v", err)
	}

	worker := &AgentsWorker{}
	logCh := make(chan string, 10)
	if err := worker.Sync(agentsDir, outputPath, logCh); err != nil {
		t.Fatalf("sync agents: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	agents, ok := payload["agents"].([]interface{})
	if !ok {
		t.Fatalf("expected agents array in output, got %#v", payload["agents"])
	}
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	agent, ok := agents[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected agent object, got %#v", agents[0])
	}

	if got := agent["name"]; got != "architect" {
		t.Fatalf("expected name architect, got %#v", got)
	}
	if got := agent["content"]; got == "" {
		t.Fatalf("expected non-empty agent content")
	}
}
