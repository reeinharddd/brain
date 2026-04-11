package agentpool

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestManager_RegisterPool(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	t.Run("register a new pool", func(t *testing.T) {
		def := AgentDefinition{
			Role:         RoleBuilder,
			Name:         "builder",
			Capabilities: []AgentCapability{CapCodeGeneration},
		}
		config := PoolConfig{
			MinInstances:     1,
			MaxInstances:     5,
			IdleTimeout:      5 * time.Minute,
			QueueCapacity:    10,
			ScaleUpThreshold: 0.7,
			ScaleDownTimeout: 2 * time.Minute,
		}
		err := mgr.RegisterPool(ctx, RoleBuilder, def, config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mgr.PoolCount() != 1 {
			t.Errorf("expected 1 pool, got %d", mgr.PoolCount())
		}
	})

	t.Run("register duplicate pool", func(t *testing.T) {
		def := AgentDefinition{
			Role:         RoleBuilder,
			Name:         "builder",
			Capabilities: []AgentCapability{CapCodeGeneration},
		}
		config := PoolConfig{
			MinInstances:     1,
			MaxInstances:     5,
			IdleTimeout:      5 * time.Minute,
			QueueCapacity:    10,
			ScaleUpThreshold: 0.7,
			ScaleDownTimeout: 2 * time.Minute,
		}
		err := mgr.RegisterPool(ctx, RoleBuilder, def, config)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrDuplicatePool) {
			t.Errorf("expected ErrDuplicatePool, got %v", err)
		}
	})
}

func TestManager_SubmitTask(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	def := AgentDefinition{
		Role:         RoleReviewer,
		Name:         "reviewer",
		Capabilities: []AgentCapability{CapCodeReview},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	if err := mgr.RegisterPool(ctx, RoleReviewer, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("submit task to known pool", func(t *testing.T) {
		task := &AgentTask{
			ID:          "mgr-task-1",
			Description: "manager test",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		err := mgr.SubmitTask(ctx, RoleReviewer, task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("submit task to unknown pool", func(t *testing.T) {
		task := &AgentTask{
			ID:          "mgr-task-2",
			Description: "unknown pool test",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		err := mgr.SubmitTask(ctx, RoleArchitect, task)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUnknownPool) {
			t.Errorf("expected ErrUnknownPool, got %v", err)
		}
	})
}

func TestManager_GetPoolStatus(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	def := AgentDefinition{
		Role:         RoleTester,
		Name:         "tester",
		Capabilities: []AgentCapability{CapTestGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	if err := mgr.RegisterPool(ctx, RoleTester, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("get pool status", func(t *testing.T) {
		status, err := mgr.GetPoolStatus(ctx, RoleTester)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.Role != RoleTester {
			t.Errorf("expected role tester, got %s", status.Role)
		}
	})

	t.Run("get unknown pool status", func(t *testing.T) {
		_, err := mgr.GetPoolStatus(ctx, RoleArchitect)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUnknownPool) {
			t.Errorf("expected ErrUnknownPool, got %v", err)
		}
	})
}

func TestManager_GetAllStatuses(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	roles := []AgentRole{RoleBuilder, RoleReviewer, RoleTester}
	for _, role := range roles {
		def := AgentDefinition{
			Role:         role,
			Name:         string(role),
			Capabilities: []AgentCapability{CapCodeGeneration},
		}
		config := PoolConfig{
			MinInstances:     1,
			MaxInstances:     5,
			IdleTimeout:      5 * time.Minute,
			QueueCapacity:    10,
			ScaleUpThreshold: 0.7,
			ScaleDownTimeout: 2 * time.Minute,
		}
		if err := mgr.RegisterPool(ctx, role, def, config); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	statuses := mgr.GetAllStatuses(ctx)
	if len(statuses) != 3 {
		t.Errorf("expected 3 statuses, got %d", len(statuses))
	}
}

func TestManager_GetAvailableAgent(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	def := AgentDefinition{
		Role:         RoleDebugger,
		Name:         "debugger",
		Capabilities: []AgentCapability{CapDebugging},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	if err := mgr.RegisterPool(ctx, RoleDebugger, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("get available agent from empty pool", func(t *testing.T) {
		_, err := mgr.GetAvailableAgent(ctx, RoleDebugger)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("get available agent after scaling", func(t *testing.T) {
		if err := mgr.ScalePool(ctx, RoleDebugger, 2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		agent, err := mgr.GetAvailableAgent(ctx, RoleDebugger)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if agent == nil {
			t.Fatal("expected agent, got nil")
		}
		if agent.Status != StatusIdle {
			t.Errorf("expected status idle, got %s", agent.Status)
		}
	})

	t.Run("get available agent from unknown pool", func(t *testing.T) {
		_, err := mgr.GetAvailableAgent(ctx, RoleArchitect)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUnknownPool) {
			t.Errorf("expected ErrUnknownPool, got %v", err)
		}
	})
}

func TestManager_ScalePool(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	if err := mgr.RegisterPool(ctx, RoleBuilder, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("scale up pool", func(t *testing.T) {
		err := mgr.ScalePool(ctx, RoleBuilder, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		status, err := mgr.GetPoolStatus(ctx, RoleBuilder)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.TotalInstances != 3 {
			t.Errorf("expected 3 instances, got %d", status.TotalInstances)
		}
	})

	t.Run("scale down pool", func(t *testing.T) {
		err := mgr.ScalePool(ctx, RoleBuilder, -1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		status, err := mgr.GetPoolStatus(ctx, RoleBuilder)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.TotalInstances != 2 {
			t.Errorf("expected 2 instances, got %d", status.TotalInstances)
		}
	})

	t.Run("scale unknown pool", func(t *testing.T) {
		err := mgr.ScalePool(ctx, RoleArchitect, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUnknownPool) {
			t.Errorf("expected ErrUnknownPool, got %v", err)
		}
	})
}

func TestManager_ListPools(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	t.Run("empty pool list", func(t *testing.T) {
		roles := mgr.ListPools(ctx)
		if len(roles) != 0 {
			t.Errorf("expected 0 roles, got %d", len(roles))
		}
	})

	t.Run("list pools after registration", func(t *testing.T) {
		roles := []AgentRole{RoleArchitect, RoleBuilder, RoleReviewer}
		for _, role := range roles {
			def := AgentDefinition{
				Role:         role,
				Name:         string(role),
				Capabilities: []AgentCapability{CapCodeGeneration},
			}
			config := PoolConfig{
				MinInstances:     1,
				MaxInstances:     5,
				IdleTimeout:      5 * time.Minute,
				QueueCapacity:    10,
				ScaleUpThreshold: 0.7,
				ScaleDownTimeout: 2 * time.Minute,
			}
			if err := mgr.RegisterPool(ctx, role, def, config); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}

		listedRoles := mgr.ListPools(ctx)
		if len(listedRoles) != 3 {
			t.Errorf("expected 3 roles, got %d", len(listedRoles))
		}
	})
}

func TestManager_PoolCount(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	if mgr.PoolCount() != 0 {
		t.Errorf("expected 0 pools, got %d", mgr.PoolCount())
	}

	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	if err := mgr.RegisterPool(ctx, RoleBuilder, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mgr.PoolCount() != 1 {
		t.Errorf("expected 1 pool, got %d", mgr.PoolCount())
	}
}

func TestManager_ContextCancellation(t *testing.T) {
	mgr := NewPoolManager()

	t.Run("register pool with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		def := AgentDefinition{
			Role:         RoleBuilder,
			Name:         "builder",
			Capabilities: []AgentCapability{CapCodeGeneration},
		}
		config := PoolConfig{
			MinInstances:     1,
			MaxInstances:     5,
			IdleTimeout:      5 * time.Minute,
			QueueCapacity:    10,
			ScaleUpThreshold: 0.7,
			ScaleDownTimeout: 2 * time.Minute,
		}
		err := mgr.RegisterPool(ctx, RoleBuilder, def, config)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
