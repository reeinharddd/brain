package mcp

import (
	"context"
	"sort"
	"testing"
)

func TestOfficialServers_Count(t *testing.T) {
	servers := OfficialServers()
	if len(servers) != 5 {
		t.Errorf("OfficialServers() count = %v, want 5", len(servers))
	}
}

func TestOfficialServers_NonEmptyFields(t *testing.T) {
	servers := OfficialServers()

	for _, server := range servers {
		t.Run(server.ID, func(t *testing.T) {
			if server.ID == "" {
				t.Error("server ID is empty")
			}
			if server.Name == "" {
				t.Error("server Name is empty")
			}
			if server.Command == "" {
				t.Error("server Command is empty")
			}
			if server.Version == "" {
				t.Error("server Version is empty")
			}
			if server.Description == "" {
				t.Error("server Description is empty")
			}
			if server.Category != "official" {
				t.Errorf("server Category = %v, want 'official'", server.Category)
			}
			if server.Timeout == 0 {
				t.Error("server Timeout should be non-zero")
			}
			if server.RateLimit <= 0 {
				t.Error("server RateLimit should be positive")
			}
		})
	}
}

func TestOfficialServers_ToolsDefined(t *testing.T) {
	// Verify that tool definitions exist for each server type
	tests := []struct {
		name      string
		tools     []MCPTool
		minTools  int
	}{
		{"filesystem", OfficialFilesystemTools(), 3},
		{"git", OfficialGitTools(), 3},
		{"github", OfficialGithubTools(), 3},
		{"terminal", OfficialTerminalTools(), 1},
		{"knowledge", OfficialKnowledgeTools(), 2},
		{"context", OfficialContextTools(), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.tools) < tt.minTools {
				t.Errorf("tools count = %v, want at least %v", len(tt.tools), tt.minTools)
			}

			// Check that each tool has required fields
			for _, tool := range tt.tools {
				if tool.Name == "" {
					t.Errorf("tool Name is empty in %s", tt.name)
				}
				if tool.Description == "" {
					t.Errorf("tool Description is empty for %s/%s", tt.name, tool.Name)
				}
				if tool.InputSchema == nil {
					t.Errorf("tool InputSchema is nil for %s/%s", tt.name, tool.Name)
				}
			}
		})
	}
}

func TestOfficialServers_UniqueIDs(t *testing.T) {
	servers := OfficialServers()
	ids := make(map[string]bool)

	for _, server := range servers {
		if ids[server.ID] {
			t.Errorf("duplicate server ID: %s", server.ID)
		}
		ids[server.ID] = true
	}

	if len(ids) != len(servers) {
		t.Errorf("expected %d unique IDs, got %d", len(servers), len(ids))
	}
}

func TestOfficialServers_Validation(t *testing.T) {
	servers := OfficialServers()

	for _, server := range servers {
		t.Run(server.ID+"/validate", func(t *testing.T) {
			if err := server.Validate(); err != nil {
				t.Errorf("Validate() error = %v", err)
			}
		})
	}
}

func TestOfficialServers_SortedByCategory(t *testing.T) {
	servers := OfficialServers()

	// All should be in "official" category
	for _, server := range servers {
		if server.Category != "official" {
			t.Errorf("server %q Category = %q, want 'official'", server.ID, server.Category)
		}
	}
}

func TestOfficialServers_CanRegister(t *testing.T) {
	servers := OfficialServers()
	reg := NewMCPRegistry()

	for _, server := range servers {
		if err := reg.Register(testContext(), server); err != nil {
			t.Errorf("failed to register server %q: %v", server.ID, err)
		}
	}

	if reg.Count() != 5 {
		t.Errorf("registry Count() = %v, want 5", reg.Count())
	}

	// Verify all are in "official" category
	officialServers := reg.GetByCategory(testContext(), "official")
	if len(officialServers) != 5 {
		t.Errorf("GetByCategory(official) count = %v, want 5", len(officialServers))
	}
}

func TestOfficialServers_IDPrefixes(t *testing.T) {
	servers := OfficialServers()
	expectedPrefixes := []string{
		"brain-filesystem",
		"brain-git",
		"brain-github",
		"brain-terminal",
		"brain-knowledge",
	}

	actualIDs := make([]string, len(servers))
	for i, s := range servers {
		actualIDs[i] = s.ID
	}
	sort.Strings(actualIDs)
	sort.Strings(expectedPrefixes)

	for i, id := range actualIDs {
		if id != expectedPrefixes[i] {
			t.Errorf("server ID[%d] = %q, want %q", i, id, expectedPrefixes[i])
		}
	}
}

func testContext() context.Context {
	return context.Background()
}
