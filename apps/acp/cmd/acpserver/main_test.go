package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	url := "http://localhost:9090"
	s := NewServer(url)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.daemonURL != url {
		t.Errorf("expected daemonURL %q, got %q", url, s.daemonURL)
	}
	if s.scanner == nil {
		t.Error("scanner should not be nil")
	}
}

func TestHandleMethodInitialize(t *testing.T) {
	s := NewServer("http://localhost:8080")
	result, err := s.handleMethod("initialize", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	serverInfo, ok := m["serverInfo"].(map[string]string)
	if !ok {
		t.Fatal("expected serverInfo map")
	}
	if name := serverInfo["name"]; name != "brain-acp-server" {
		t.Errorf("expected name %q, got %q", "brain-acp-server", name)
	}

	capabilities, ok := m["capabilities"].(map[string]bool)
	if !ok {
		t.Fatal("expected capabilities map")
	}
	for _, capName := range []string{"tools", "context", "policy", "memory"} {
		if val, exists := capabilities[capName]; !exists || val != true {
			t.Errorf("capability %q should be true, got %v", capName, val)
		}
	}
}

func TestHandleMethodToolsList(t *testing.T) {
	s := NewServer("http://localhost:8080")
	result, err := s.handleMethod("tools/list", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	tools, ok := m["tools"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected tools slice")
	}
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	expectedNames := []string{"brain_resolve_artifact", "brain_get_context", "brain_check_policy"}
	for i, name := range expectedNames {
		if tools[i]["name"] != name {
			t.Errorf("expected tool %d to be %q, got %v", i, name, tools[i]["name"])
		}
	}
}

func TestHandleMethodToolsCall(t *testing.T) {
	s := NewServer("http://localhost:8080")
	params := map[string]interface{}{
		"name": "brain_resolve_artifact",
	}
	result, err := s.handleMethod("tools/call", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	content, ok := m["content"].([]map[string]string)
	if !ok {
		t.Fatal("expected content slice")
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
	if content[0]["type"] != "text" {
		t.Errorf("expected content type %q, got %q", "text", content[0]["type"])
	}
}

func TestHandleMethodUnknown(t *testing.T) {
	s := NewServer("http://localhost:8080")
	_, err := s.handleMethod("unknown/method", nil)
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
}

func TestSendResultNoPanic(t *testing.T) {
	s := NewServer("http://localhost:8080")
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = old
	}()

	result := map[string]string{"status": "ok"}
	s.sendResult(1, result)

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)

	var resp ACPResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc %q, got %q", "2.0", resp.JSONRPC)
	}
	if resp.ID != 1 {
		t.Errorf("expected id 1, got %d", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}
}

func TestSendErrorNoPanic(t *testing.T) {
	s := NewServer("http://localhost:8080")
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = old
	}()

	s.sendError(42, -32603, "internal error")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)

	var resp ACPResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc %q, got %q", "2.0", resp.JSONRPC)
	}
	if resp.ID != 42 {
		t.Errorf("expected id 42, got %d", resp.ID)
	}
	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != -32603 {
		t.Errorf("expected error code -32603, got %d", resp.Error.Code)
	}
	if resp.Error.Message != "internal error" {
		t.Errorf("expected error message %q, got %q", "internal error", resp.Error.Message)
	}
}
