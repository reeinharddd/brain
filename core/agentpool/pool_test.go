package agentpool

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPool_SubmitGetCompleteTask(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     3,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)

	t.Run("submit task to queue", func(t *testing.T) {
		ctx := context.Background()
		task := &AgentTask{
			ID:          "task-1",
			Description: "build something",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		err := pool.SubmitTask(ctx, task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("get task from queue", func(t *testing.T) {
		ctx := context.Background()
		task, err := pool.GetTask(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if task.ID != "task-1" {
			t.Errorf("expected task-1, got %s", task.ID)
		}
	})

	t.Run("get task from empty queue", func(t *testing.T) {
		ctx := context.Background()
		_, err := pool.GetTask(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrQueueEmpty) {
			t.Errorf("expected ErrQueueEmpty, got %v", err)
		}
	})
}

func TestPool_TaskLifecycle(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     3,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)

	ctx := context.Background()

	// Scale up to create agents
	agents, err := pool.ScaleUp(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error scaling up: %v", err)
	}

	agent := agents[0]

	t.Run("submit and assign task", func(t *testing.T) {
		task := &AgentTask{
			ID:          "lifecycle-task",
			Description: "lifecycle test",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		if err := pool.SubmitTask(ctx, task); err != nil {
			t.Fatalf("unexpected error submitting task: %v", err)
		}

		// Get task from queue
		gotTask, err := pool.GetTask(ctx)
		if err != nil {
			t.Fatalf("unexpected error getting task: %v", err)
		}

		// Assign task to agent
		if err := agent.AssignTask(ctx, gotTask); err != nil {
			t.Fatalf("unexpected error assigning task: %v", err)
		}

		if agent.GetStatus() != StatusBusy {
			t.Errorf("expected status busy, got %s", agent.GetStatus())
		}
	})

	t.Run("complete task", func(t *testing.T) {
		err := pool.CompleteTask(ctx, "lifecycle-task", agent.ID)
		if err != nil {
			t.Fatalf("unexpected error completing task: %v", err)
		}

		if agent.GetStatus() != StatusIdle {
			t.Errorf("expected status idle, got %s", agent.GetStatus())
		}
		if agent.TasksCompleted != 1 {
			t.Errorf("expected 1 task completed, got %d", agent.TasksCompleted)
		}
	})

	t.Run("fail task", func(t *testing.T) {
		task := &AgentTask{
			ID:          "fail-task",
			Description: "fail test",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		if err := pool.SubmitTask(ctx, task); err != nil {
			t.Fatalf("unexpected error submitting task: %v", err)
		}

		gotTask, err := pool.GetTask(ctx)
		if err != nil {
			t.Fatalf("unexpected error getting task: %v", err)
		}

		if err := agent.AssignTask(ctx, gotTask); err != nil {
			t.Fatalf("unexpected error assigning task: %v", err)
		}

		err = pool.FailTask(ctx, "fail-task", agent.ID, "some error")
		if err != nil {
			t.Fatalf("unexpected error failing task: %v", err)
		}

		if agent.GetStatus() != StatusIdle {
			t.Errorf("expected status idle, got %s", agent.GetStatus())
		}
		if agent.TasksFailed != 1 {
			t.Errorf("expected 1 task failed, got %d", agent.TasksFailed)
		}
	})
}

func TestPool_ScaleUp(t *testing.T) {
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

	t.Run("scale up from zero", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()

		agents, err := pool.ScaleUp(ctx, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(agents) != 3 {
			t.Errorf("expected 3 agents, got %d", len(agents))
		}

		for _, a := range agents {
			if a.Status != StatusIdle {
				t.Errorf("expected status idle, got %s", a.Status)
			}
			if a.Role != RoleReviewer {
				t.Errorf("expected role reviewer, got %s", a.Role)
			}
		}
	})

	t.Run("cannot exceed max instances", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()

		_, err := pool.ScaleUp(ctx, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = pool.ScaleUp(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrMaxInstances) {
			t.Errorf("expected ErrMaxInstances, got %v", err)
		}
	})

	t.Run("scale up partial when count exceeds max", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()

		_, err := pool.ScaleUp(ctx, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		agents, err := pool.ScaleUp(ctx, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should only scale up by 2 to reach max of 5
		if len(agents) != 2 {
			t.Errorf("expected 2 agents (to reach max 5), got %d", len(agents))
		}

		status := pool.GetStatus(ctx)
		if status.TotalInstances != 5 {
			t.Errorf("expected 5 total instances, got %d", status.TotalInstances)
		}
	})
}

func TestPool_ScaleDown(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleTester,
		Name:         "tester",
		Capabilities: []AgentCapability{CapTestGeneration},
	}
	config := PoolConfig{
		MinInstances:     2,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}

	t.Run("scale down idle agents", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()

		_, err := pool.ScaleUp(ctx, 4)
		if err != nil {
			t.Fatalf("unexpected error scaling up: %v", err)
		}

		err = pool.ScaleDown(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error scaling down: %v", err)
		}

		status := pool.GetStatus(ctx)
		if status.TotalInstances != 2 {
			t.Errorf("expected 2 instances, got %d", status.TotalInstances)
		}
	})

	t.Run("cannot scale below minimum", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()

		_, err := pool.ScaleUp(ctx, 3)
		if err != nil {
			t.Fatalf("unexpected error scaling up: %v", err)
		}

		err = pool.ScaleDown(ctx, 5)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrMinInstances) {
			t.Errorf("expected ErrMinInstances, got %v", err)
		}
	})
}

func TestPool_GetStatus(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleArchitect,
		Name:         "architect",
		Capabilities: []AgentCapability{CapSystemDesign},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     5,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)
	ctx := context.Background()

	t.Run("initial status", func(t *testing.T) {
		status := pool.GetStatus(ctx)
		if status.TotalInstances != 0 {
			t.Errorf("expected 0 total instances, got %d", status.TotalInstances)
		}
		if status.Load != 0.0 {
			t.Errorf("expected 0.0 load, got %f", status.Load)
		}
		if status.Role != RoleArchitect {
			t.Errorf("expected role architect, got %s", status.Role)
		}
	})

	t.Run("status after scaling and submitting", func(t *testing.T) {
		_, err := pool.ScaleUp(ctx, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		task := &AgentTask{
			ID:          "status-task",
			Description: "status test",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		if err := pool.SubmitTask(ctx, task); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		status := pool.GetStatus(ctx)
		if status.TotalInstances != 3 {
			t.Errorf("expected 3 total instances, got %d", status.TotalInstances)
		}
		if status.AvailableInstances != 3 {
			t.Errorf("expected 3 available instances, got %d", status.AvailableInstances)
		}
		if status.QueuedTasks != 1 {
			t.Errorf("expected 1 queued task, got %d", status.QueuedTasks)
		}
		if status.TotalSubmitted != 1 {
			t.Errorf("expected 1 total submitted, got %d", status.TotalSubmitted)
		}
	})
}

func TestPool_GetAvailableInstance(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     3,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)
	ctx := context.Background()

	t.Run("no instances available", func(t *testing.T) {
		_, err := pool.GetAvailableInstance(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("get available instance", func(t *testing.T) {
		_, err := pool.ScaleUp(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		agent, err := pool.GetAvailableInstance(ctx)
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
}

func TestPool_GetLoad(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleReviewer,
		Name:         "reviewer",
		Capabilities: []AgentCapability{CapCodeReview},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     4,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}

	t.Run("zero load with no instances", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		load := pool.GetLoad()
		if load != 0.0 {
			t.Errorf("expected 0.0 load, got %f", load)
		}
	})

	t.Run("zero load with idle instances", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()
		_, err := pool.ScaleUp(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		load := pool.GetLoad()
		if load != 0.0 {
			t.Errorf("expected 0.0 load, got %f", load)
		}
	})

	t.Run("load increases with busy agents", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()
		agents, err := pool.ScaleUp(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		task := &AgentTask{
			ID:          "load-task-1",
			Description: "load test",
			CreatedAt:   time.Now(),
		}
		if err := agents[0].AssignTask(ctx, task); err != nil {
			t.Fatalf("unexpected error assigning task: %v", err)
		}

		load := pool.GetLoad()
		// 1 busy out of 2 = 0.5 busy ratio * 0.7 = 0.35
		expected := 0.5 * 0.7
		if load < expected-0.01 || load > expected+0.01 {
			t.Errorf("expected load ~%f, got %f", expected, load)
		}
	})

	t.Run("full load with all busy", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()
		agents, err := pool.ScaleUp(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, agent := range agents {
			task := &AgentTask{
				ID:          "load-task-" + agent.ID,
				Description: "load test",
				CreatedAt:   time.Now(),
			}
			if err := agent.AssignTask(ctx, task); err != nil {
				t.Fatalf("unexpected error assigning task: %v", err)
			}
		}

		load := pool.GetLoad()
		expected := 1.0 * 0.7 // 100% busy
		if load < expected-0.01 || load > expected+0.01 {
			t.Errorf("expected load ~%f, got %f", expected, load)
		}
	})

	t.Run("load is 1.0 when tasks exist but no agents", func(t *testing.T) {
		pool := NewAgentPool(def, config)
		ctx := context.Background()
		task := &AgentTask{
			ID:          "orphan-task",
			Description: "orphan test",
			CreatedAt:   time.Now(),
		}
		if err := pool.SubmitTask(ctx, task); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		load := pool.GetLoad()
		if load != 1.0 {
			t.Errorf("expected 1.0 load, got %f", load)
		}
	})
}

func TestPool_QueueCapacityEnforcement(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     3,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    2,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)
	ctx := context.Background()

	// Fill the queue
	for i := 0; i < 2; i++ {
		task := &AgentTask{
			ID:          "queue-task-" + string(rune('0'+i)),
			Description: "queue test",
			Priority:    1,
			CreatedAt:   time.Now(),
		}
		if err := pool.SubmitTask(ctx, task); err != nil {
			t.Fatalf("unexpected error submitting task %d: %v", i, err)
		}
	}

	// Try to exceed capacity
	task := &AgentTask{
		ID:          "overflow-task",
		Description: "overflow test",
		Priority:    1,
		CreatedAt:   time.Now(),
	}
	err := pool.SubmitTask(ctx, task)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}

func TestPool_ConcurrentOperations(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     2,
		MaxInstances:     10,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    50,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)
	ctx := context.Background()

	_, err := pool.ScaleUp(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done := make(chan bool, 3)

	// Concurrent submitters
	go func() {
		for i := 0; i < 20; i++ {
			task := &AgentTask{
				ID:          "concurrent-task-" + string(rune('0'+i%10)) + string(rune('0'+i/10)),
				Description: "concurrent test",
				Priority:    1,
				CreatedAt:   time.Now(),
			}
			pool.SubmitTask(ctx, task) // ignore errors for overflow
		}
		done <- true
	}()

	// Concurrent getters
	go func() {
		for i := 0; i < 20; i++ {
			pool.GetTask(ctx) // ignore errors for empty queue
		}
		done <- true
	}()

	// Concurrent status checks
	go func() {
		for i := 0; i < 20; i++ {
			pool.GetStatus(ctx)
			pool.GetLoad()
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     3,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)

	t.Run("submit task with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		task := &AgentTask{
			ID:          "cancelled-task",
			Description: "cancelled test",
			CreatedAt:   time.Now(),
		}
		err := pool.SubmitTask(ctx, task)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("get task with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := pool.GetTask(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("scale up with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := pool.ScaleUp(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("scale down with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := pool.ScaleDown(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPool_CompleteTaskErrors(t *testing.T) {
	def := AgentDefinition{
		Role:         RoleBuilder,
		Name:         "builder",
		Capabilities: []AgentCapability{CapCodeGeneration},
	}
	config := PoolConfig{
		MinInstances:     1,
		MaxInstances:     3,
		IdleTimeout:      5 * time.Minute,
		QueueCapacity:    10,
		ScaleUpThreshold: 0.7,
		ScaleDownTimeout: 2 * time.Minute,
	}
	pool := NewAgentPool(def, config)
	ctx := context.Background()

	t.Run("complete task with unknown agent", func(t *testing.T) {
		err := pool.CompleteTask(ctx, "task-1", "unknown-agent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrAgentNotFound) {
			t.Errorf("expected ErrAgentNotFound, got %v", err)
		}
	})

	t.Run("complete unassigned task", func(t *testing.T) {
		agents, err := pool.ScaleUp(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		agent := agents[0]

		err = pool.CompleteTask(ctx, "unassigned-task", agent.ID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrTaskNotAssigned) {
			t.Errorf("expected ErrTaskNotAssigned, got %v", err)
		}
	})
}
