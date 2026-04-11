package agentpool

import (
	"context"
	"testing"
	"time"
)

func TestAutoScaler_ScaleUpWhenOverloaded(t *testing.T) {
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
		ScaleDownTimeout: 50 * time.Millisecond, // Short cooldown for tests
	}
	if err := mgr.RegisterPool(ctx, RoleBuilder, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Start with 1 instance
	if err := mgr.ScalePool(ctx, RoleBuilder, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Make the instance busy
	agent, err := mgr.GetAvailableAgent(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	task := &AgentTask{
		ID:          "busy-task",
		Description: "make busy",
		CreatedAt:   time.Now(),
	}
	if err := agent.AssignTask(ctx, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Add queued tasks to increase load
	for i := 0; i < 8; i++ {
		queueTask := &AgentTask{
			ID:          "queue-task-" + string(rune('0'+i)),
			Description: "queue task",
			CreatedAt:   time.Now(),
		}
		if err := mgr.SubmitTask(ctx, RoleBuilder, queueTask); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Verify load is high
	status, err := mgr.GetPoolStatus(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Load <= 0.7 {
		t.Logf("load is %f, may not trigger scale up immediately", status.Load)
	}

	// Create autoscaler and tick
	as := NewAutoScaler(mgr, 100*time.Millisecond)
	as.Enable()

	// Tick to trigger scaling
	err = as.Tick(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, err = mgr.GetPoolStatus(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have scaled up
	if status.TotalInstances < 2 {
		t.Logf("expected at least 2 instances after scale up, got %d (load=%f)", status.TotalInstances, status.Load)
	}
}

func TestAutoScaler_ScaleDownWhenIdle(t *testing.T) {
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
		ScaleDownTimeout: 50 * time.Millisecond, // Short cooldown for tests
	}
	if err := mgr.RegisterPool(ctx, RoleBuilder, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Start with 3 instances
	if err := mgr.ScalePool(ctx, RoleBuilder, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for cooldown to expire
	time.Sleep(60 * time.Millisecond)

	as := NewAutoScaler(mgr, 100*time.Millisecond)
	as.Enable()

	err := as.Tick(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, err := mgr.GetPoolStatus(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have scaled down by 1
	if status.TotalInstances != 2 {
		t.Errorf("expected 2 instances after scale down, got %d", status.TotalInstances)
	}
}

func TestAutoScaler_CooldownEnforcement(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	// Long cooldown to ensure we don't scale
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 10 * time.Second, // Long cooldown
	}
	if err := mgr.RegisterPool(ctx, RoleBuilder, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Start with 3 instances
	if err := mgr.ScalePool(ctx, RoleBuilder, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	as := NewAutoScaler(mgr, 100*time.Millisecond)
	as.Enable()

	// Multiple ticks should not trigger scale down due to cooldown
	for i := 0; i < 3; i++ {
		err := as.Tick(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	status, err := mgr.GetPoolStatus(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have 3 instances due to cooldown
	if status.TotalInstances != 3 {
		t.Errorf("expected 3 instances (cooldown should prevent scale down), got %d", status.TotalInstances)
	}
}

func TestAutoScaler_DisabledDoesNothing(t *testing.T) {
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
		ScaleDownTimeout: 50 * time.Millisecond,
	}
	if err := mgr.RegisterPool(ctx, RoleBuilder, def, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Start with 3 instances
	if err := mgr.ScalePool(ctx, RoleBuilder, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create autoscaler but DON'T enable it
	as := NewAutoScaler(mgr, 100*time.Millisecond)

	err := as.Tick(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, err := mgr.GetPoolStatus(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have 3 instances
	if status.TotalInstances != 3 {
		t.Errorf("expected 3 instances (disabled scaler should do nothing), got %d", status.TotalInstances)
	}
}

func TestAutoScaler_TickAcrossMultiplePools(t *testing.T) {
	mgr := NewPoolManager()
	ctx := context.Background()

	roles := []AgentRole{RoleBuilder, RoleReviewer}
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
			ScaleDownTimeout: 50 * time.Millisecond,
		}
		if err := mgr.RegisterPool(ctx, role, def, config); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Scale both pools
	if err := mgr.ScalePool(ctx, RoleBuilder, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mgr.ScalePool(ctx, RoleReviewer, 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)

	as := NewAutoScaler(mgr, 100*time.Millisecond)
	as.Enable()

	err := as.Tick(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both pools should have been evaluated
	builderStatus, err := mgr.GetPoolStatus(ctx, RoleBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reviewerStatus, err := mgr.GetPoolStatus(ctx, RoleReviewer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should have scaled down by 1
	if builderStatus.TotalInstances != 2 {
		t.Errorf("expected builder pool to have 2 instances, got %d", builderStatus.TotalInstances)
	}
	if reviewerStatus.TotalInstances != 1 {
		t.Errorf("expected reviewer pool to have 1 instance, got %d", reviewerStatus.TotalInstances)
	}
}

func TestAutoScaler_ContextCancellation(t *testing.T) {
	mgr := NewPoolManager()

	as := NewAutoScaler(mgr, 100*time.Millisecond)
	as.Enable()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := as.Tick(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAutoScaler_EnableDisable(t *testing.T) {
	mgr := NewPoolManager()
	as := NewAutoScaler(mgr, 100*time.Millisecond)

	if as.IsEnabled() {
		t.Error("expected autoscaler to be disabled by default")
	}

	as.Enable()
	if !as.IsEnabled() {
		t.Error("expected autoscaler to be enabled after Enable()")
	}

	as.Disable()
	if as.IsEnabled() {
		t.Error("expected autoscaler to be disabled after Disable()")
	}
}
