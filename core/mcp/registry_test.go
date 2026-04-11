package mcp

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func testConfig(id, name, category string) MCPServerConfig {
	return MCPServerConfig{
		ID:          id,
		Name:        name,
		Version:     "1.0.0",
		Description: "Test server " + name,
		Category:    category,
		Transport:   TransportStdIO,
		Command:     "test-cmd",
		Args:        []string{"--arg1"},
		Env:         map[string]string{"ENV": "test"},
		Timeout:     30 * time.Second,
		RateLimit:   100,
	}
}

func TestMCPRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		config  MCPServerConfig
		wantErr bool
		errType error
	}{
		{
			name:    "valid config",
			config:  testConfig("test1", "Test Server 1", "official"),
			wantErr: false,
		},
		{
			name: "missing ID",
			config: MCPServerConfig{
				Name:    "No ID",
				Command: "test",
			},
			wantErr: true,
			errType: ErrInvalidConfig,
		},
		{
			name: "missing Name",
			config: MCPServerConfig{
				ID:      "no-name",
				Command: "test",
			},
			wantErr: true,
			errType: ErrInvalidConfig,
		},
		{
			name: "missing Command",
			config: MCPServerConfig{
				ID:   "no-cmd",
				Name: "No Command",
			},
			wantErr: true,
			errType: ErrInvalidConfig,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewMCPRegistry()
			err := reg.Register(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("Register() error = %v, want error wrapping %v", err, tt.errType)
				}
			}
		})
	}
}

func TestMCPRegistry_DuplicateRegistration(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	cfg := testConfig("dup1", "Duplicate", "official")
	if err := reg.Register(ctx, cfg); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	err := reg.Register(ctx, cfg)
	if err == nil {
		t.Fatal("second Register() expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicateServer) {
		t.Errorf("Register() error = %v, want error wrapping %v", err, ErrDuplicateServer)
	}
}

func TestMCPRegistry_Get(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	cfg := testConfig("get1", "Get Test", "official")
	if err := reg.Register(ctx, cfg); err != nil {
		t.Fatalf("failed to register server: %v", err)
	}

	t.Run("existing server", func(t *testing.T) {
		got, err := reg.Get(ctx, "get1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.ID != "get1" {
			t.Errorf("Get() ID = %v, want %v", got.ID, "get1")
		}
		if got.Name != "Get Test" {
			t.Errorf("Get() Name = %v, want %v", got.Name, "Get Test")
		}
	})

	t.Run("non-existent server", func(t *testing.T) {
		_, err := reg.Get(ctx, "nonexistent")
		if err == nil {
			t.Fatal("Get() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("Get() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerNotFound {
			t.Errorf("Get() error = %v, want %v", serverErr.Err, ErrServerNotFound)
		}
	})
}

func isServerError(err error, out **ServerError) bool {
	if se, ok := err.(*ServerError); ok {
		*out = se
		return true
	}
	return false
}

func TestMCPRegistry_List(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	configs := []MCPServerConfig{
		testConfig("list1", "List Test 1", "official"),
		testConfig("list2", "List Test 2", "community"),
		testConfig("list3", "List Test 3", "official"),
	}

	for _, cfg := range configs {
		if err := reg.Register(ctx, cfg); err != nil {
			t.Fatalf("failed to register server: %v", err)
		}
	}

	t.Run("list all servers", func(t *testing.T) {
		got := reg.List(ctx)
		if len(got) != 3 {
			t.Fatalf("List() count = %v, want 3", len(got))
		}
	})

	t.Run("empty registry", func(t *testing.T) {
		emptyReg := NewMCPRegistry()
		got := emptyReg.List(ctx)
		if len(got) != 0 {
			t.Errorf("List() count = %v, want 0", len(got))
		}
	})
}

func TestMCPRegistry_GetByCategory(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	configs := []MCPServerConfig{
		testConfig("cat1", "Cat Test 1", "official"),
		testConfig("cat2", "Cat Test 2", "community"),
		testConfig("cat3", "Cat Test 3", "official"),
		testConfig("cat4", "Cat Test 4", "enterprise"),
	}

	for _, cfg := range configs {
		if err := reg.Register(ctx, cfg); err != nil {
			t.Fatalf("failed to register server: %v", err)
		}
	}

	tests := []struct {
		name     string
		category string
		want     int
	}{
		{"official category", "official", 2},
		{"community category", "community", 1},
		{"enterprise category", "enterprise", 1},
		{"non-existent category", "private", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reg.GetByCategory(ctx, tt.category)
			if len(got) != tt.want {
				t.Errorf("GetByCategory() count = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestMCPRegistry_Unregister(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	cfg := testConfig("unreg1", "Unregister Test", "official")
	if err := reg.Register(ctx, cfg); err != nil {
		t.Fatalf("failed to register server: %v", err)
	}

	t.Run("existing server", func(t *testing.T) {
		if err := reg.Unregister(ctx, "unreg1"); err != nil {
			t.Fatalf("Unregister() error = %v", err)
		}
		if reg.Count() != 0 {
			t.Errorf("Count() = %v, want 0 after unregister", reg.Count())
		}
	})

	t.Run("non-existent server", func(t *testing.T) {
		err := reg.Unregister(ctx, "nonexistent")
		if err == nil {
			t.Fatal("Unregister() expected error, got nil")
		}
		var serverErr *ServerError
		if !isServerError(err, &serverErr) {
			t.Fatalf("Unregister() error type = %T, want *ServerError", err)
		}
		if serverErr.Err != ErrServerNotFound {
			t.Errorf("Unregister() error = %v, want %v", serverErr.Err, ErrServerNotFound)
		}
	})

	t.Run("category cleaned up", func(t *testing.T) {
		reg2 := NewMCPRegistry()
		_ = reg2.Register(ctx, testConfig("cat-clean1", "Cat Clean 1", "testcat"))
		_ = reg2.Register(ctx, testConfig("cat-clean2", "Cat Clean 2", "testcat"))

		_ = reg2.Unregister(ctx, "cat-clean1")
		_ = reg2.Unregister(ctx, "cat-clean2")

		// Category should be cleaned up when empty
		remaining := reg2.GetByCategory(ctx, "testcat")
		if len(remaining) != 0 {
			t.Errorf("GetByCategory() after all unregistered = %v, want 0", len(remaining))
		}
	})
}

func TestMCPRegistry_Count(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	if reg.Count() != 0 {
		t.Errorf("Count() = %v, want 0 for empty registry", reg.Count())
	}

	_ = reg.Register(ctx, testConfig("count1", "Count 1", "official"))
	if reg.Count() != 1 {
		t.Errorf("Count() = %v, want 1", reg.Count())
	}

	_ = reg.Register(ctx, testConfig("count2", "Count 2", "community"))
	if reg.Count() != 2 {
		t.Errorf("Count() = %v, want 2", reg.Count())
	}
}

func TestMCPRegistry_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	reg := NewMCPRegistry()

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = reg.Register(ctx, testConfig(
				"conc"+string(rune('0'+id)),
				"Concurrent "+string(rune('0'+id)),
				"test",
			))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = reg.List(ctx)
			_ = reg.GetByCategory(ctx, "test")
			_ = reg.Count()
		}()
	}

	wg.Wait()
	// Just ensure no race condition panic
}

func TestMCPRegistry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reg := NewMCPRegistry()

	t.Run("Register with canceled context", func(t *testing.T) {
		err := reg.Register(ctx, testConfig("cancel1", "Cancel 1", "test"))
		if err == nil {
			t.Fatal("Register() expected error with canceled context")
		}
	})

	t.Run("Get with canceled context", func(t *testing.T) {
		_, err := reg.Get(ctx, "any")
		if err == nil {
			t.Fatal("Get() expected error with canceled context")
		}
	})

	t.Run("Unregister with canceled context", func(t *testing.T) {
		err := reg.Unregister(ctx, "any")
		if err == nil {
			t.Fatal("Unregister() expected error with canceled context")
		}
	})
}
