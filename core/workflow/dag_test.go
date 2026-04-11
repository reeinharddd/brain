package workflow

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewWorkflowDAG(t *testing.T) {
	t.Run("creates DAG with correct properties", func(t *testing.T) {
		dag := NewWorkflowDAG("test-1", "Test Workflow", true)

		if dag.ID != "test-1" {
			t.Errorf("expected ID 'test-1', got %q", dag.ID)
		}
		if dag.Name != "Test Workflow" {
			t.Errorf("expected Name 'Test Workflow', got %q", dag.Name)
		}
		if !dag.Parallel {
			t.Error("expected Parallel to be true")
		}
		if dag.Nodes == nil {
			t.Error("expected Nodes map to be initialized")
		}
		if len(dag.Nodes) != 0 {
			t.Errorf("expected empty Nodes map, got %d nodes", len(dag.Nodes))
		}
	})

	t.Run("creates DAG with parallel disabled", func(t *testing.T) {
		dag := NewWorkflowDAG("test-2", "Linear Workflow", false)

		if dag.Parallel {
			t.Error("expected Parallel to be false")
		}
	})
}

func TestAddNode(t *testing.T) {
	tests := []struct {
		name      string
		dagID     string
		node      *WorkflowNode
		expectErr bool
	}{
		{
			name:  "adds valid node",
			dagID: "dag-1",
			node: &WorkflowNode{
				ID:      "node-1",
				Name:    "First Node",
				Agent:   "agent-a",
				Input:   map[string]string{"key": "value"},
				Timeout: 5 * time.Minute,
			},
			expectErr: false,
		},
		{
			name:      "rejects nil node",
			dagID:     "dag-2",
			node:      nil,
			expectErr: true,
		},
		{
			name:  "rejects node with empty ID",
			dagID: "dag-3",
			node: &WorkflowNode{
				ID:   "",
				Name: "Empty ID Node",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			dag := NewWorkflowDAG(tt.dagID, "Test", true)

			err := dag.AddNode(ctx, tt.node)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			node, err := dag.GetNode(ctx, tt.node.ID)
			if err != nil {
				t.Errorf("failed to get node: %v", err)
				return
			}

			if node.Status != StatusPending {
				t.Errorf("expected status %q, got %q", StatusPending, node.Status)
			}
			if node.Name != tt.node.Name {
				t.Errorf("expected name %q, got %q", tt.node.Name, node.Name)
			}
		})
	}

	t.Run("rejects duplicate node ID", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("dag-dup", "Test", true)

		node1 := &WorkflowNode{ID: "node-1", Name: "First"}
		node2 := &WorkflowNode{ID: "node-1", Name: "Duplicate"}

		if err := dag.AddNode(ctx, node1); err != nil {
			t.Fatalf("failed to add first node: %v", err)
		}

		err := dag.AddNode(ctx, node2)
		if err == nil {
			t.Error("expected error for duplicate node ID, got nil")
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid DAG with no nodes", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("empty", "Empty", true)

		if err := dag.Validate(ctx); err != nil {
			t.Errorf("unexpected error for empty DAG: %v", err)
		}
	})

	t.Run("valid linear DAG", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("linear", "Linear", false)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})

		if err := dag.Validate(ctx); err != nil {
			t.Errorf("unexpected error for valid linear DAG: %v", err)
		}
	})

	t.Run("valid diamond DAG", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("diamond", "Diamond", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "d", Name: "D", DependsOn: []string{"b", "c"}})

		if err := dag.Validate(ctx); err != nil {
			t.Errorf("unexpected error for valid diamond DAG: %v", err)
		}
	})

	t.Run("detects missing dependency", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("missing-dep", "Missing Dep", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"nonexistent"}})

		err := dag.Validate(ctx)
		if err == nil {
			t.Error("expected error for missing dependency, got nil")
		}
	})

	t.Run("detects simple cycle", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("cycle", "Cyclic", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A", DependsOn: []string{"b"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})

		err := dag.Validate(ctx)
		if err == nil {
			t.Error("expected error for cycle, got nil")
		}
	})

	t.Run("detects longer cycle", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("long-cycle", "Long Cycle", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A", DependsOn: []string{"c"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})

		err := dag.Validate(ctx)
		if err == nil {
			t.Error("expected error for cycle, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		dag := NewWorkflowDAG("cancelled", "Cancelled", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})

		err := dag.Validate(ctx)
		if err == nil {
			t.Error("expected error for cancelled context, got nil")
		}
	})
}

func TestGetReadyNodes(t *testing.T) {
	t.Run("linear workflow - first node ready", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("linear", "Linear", false)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})

		ready := dag.GetReadyNodes(ctx)
		if len(ready) != 1 || ready[0] != "a" {
			t.Errorf("expected ready nodes [a], got %v", ready)
		}
	})

	t.Run("linear workflow - after first completes", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("linear", "Linear", false)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})

		_ = dag.MarkCompleted(ctx, "a", &TaskOutput{Result: []byte("done")})

		ready := dag.GetReadyNodes(ctx)
		if len(ready) != 1 || ready[0] != "b" {
			t.Errorf("expected ready nodes [b], got %v", ready)
		}
	})

	t.Run("diamond workflow - multiple nodes ready", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("diamond", "Diamond", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "d", Name: "D", DependsOn: []string{"b", "c"}})

		_ = dag.MarkCompleted(ctx, "a", &TaskOutput{Result: []byte("done")})

		ready := dag.GetReadyNodes(ctx)
		if len(ready) != 2 {
			t.Fatalf("expected 2 ready nodes, got %d: %v", len(ready), ready)
		}

		readySet := make(map[string]bool)
		for _, id := range ready {
			readySet[id] = true
		}
		if !readySet["b"] || !readySet["c"] {
			t.Errorf("expected ready nodes [b, c], got %v", ready)
		}
	})

	t.Run("parallel workflow - all nodes ready", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("parallel", "Parallel", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "x", Name: "X"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "y", Name: "Y"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "z", Name: "Z"})

		ready := dag.GetReadyNodes(ctx)
		if len(ready) != 3 {
			t.Fatalf("expected 3 ready nodes, got %d: %v", len(ready), ready)
		}
	})

	t.Run("excludes completed and failed nodes", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("mixed", "Mixed", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B"})

		_ = dag.MarkCompleted(ctx, "a", &TaskOutput{Result: []byte("done")})
		_ = dag.MarkFailed(ctx, "b", fmt.Errorf("failed"))

		ready := dag.GetReadyNodes(ctx)
		if len(ready) != 0 {
			t.Errorf("expected 0 ready nodes, got %d: %v", len(ready), ready)
		}
	})

	t.Run("excludes nodes with incomplete dependencies", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("deps", "Deps", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"a"}})

		ready := dag.GetReadyNodes(ctx)
		if len(ready) != 1 || ready[0] != "a" {
			t.Errorf("expected ready nodes [a], got %v", ready)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		dag := NewWorkflowDAG("cancelled", "Cancelled", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		ready := dag.GetReadyNodes(ctx)
		if ready != nil {
			t.Errorf("expected nil for cancelled context, got %v", ready)
		}
	})
}

func TestStatusTracking(t *testing.T) {
	t.Run("mark node completed", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		output := &TaskOutput{
			Result:     []byte("result"),
			Duration:   1 * time.Second,
			TokensUsed: 100,
			CostUSD:    0.01,
		}

		if err := dag.MarkCompleted(ctx, "a", output); err != nil {
			t.Fatalf("failed to mark completed: %v", err)
		}

		node, _ := dag.GetNode(ctx, "a")
		if node.Status != StatusCompleted {
			t.Errorf("expected status %q, got %q", StatusCompleted, node.Status)
		}
		if node.Output != output {
			t.Error("expected output to be set")
		}
	})

	t.Run("mark node failed", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		testErr := fmt.Errorf("something went wrong")
		if err := dag.MarkFailed(ctx, "a", testErr); err != nil {
			t.Fatalf("failed to mark failed: %v", err)
		}

		node, _ := dag.GetNode(ctx, "a")
		if node.Status != StatusFailed {
			t.Errorf("expected status %q, got %q", StatusFailed, node.Status)
		}
		if node.Output.Error != "something went wrong" {
			t.Errorf("expected error message %q, got %q", "something went wrong", node.Output.Error)
		}
	})

	t.Run("mark node skipped", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		if err := dag.MarkSkipped(ctx, "a"); err != nil {
			t.Fatalf("failed to mark skipped: %v", err)
		}

		node, _ := dag.GetNode(ctx, "a")
		if node.Status != StatusSkipped {
			t.Errorf("expected status %q, got %q", StatusSkipped, node.Status)
		}
	})

	t.Run("get all statuses", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C"})

		_ = dag.MarkCompleted(ctx, "a", &TaskOutput{Result: []byte("done")})
		_ = dag.MarkFailed(ctx, "b", fmt.Errorf("error"))

		statuses := dag.GetAllStatuses(ctx)
		if len(statuses) != 3 {
			t.Fatalf("expected 3 statuses, got %d", len(statuses))
		}

		if statuses["a"] != StatusCompleted {
			t.Errorf("expected node a status %q, got %q", StatusCompleted, statuses["a"])
		}
		if statuses["b"] != StatusFailed {
			t.Errorf("expected node b status %q, got %q", StatusFailed, statuses["b"])
		}
		if statuses["c"] != StatusPending {
			t.Errorf("expected node c status %q, got %q", StatusPending, statuses["c"])
		}
	})

	t.Run("mark nonexistent node", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)

		if err := dag.MarkCompleted(ctx, "nonexistent", &TaskOutput{}); err == nil {
			t.Error("expected error for nonexistent node, got nil")
		}
		if err := dag.MarkFailed(ctx, "nonexistent", fmt.Errorf("err")); err == nil {
			t.Error("expected error for nonexistent node, got nil")
		}
		if err := dag.MarkSkipped(ctx, "nonexistent"); err == nil {
			t.Error("expected error for nonexistent node, got nil")
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("concurrent AddNode", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("concurrent", "Concurrent", true)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				nodeID := fmt.Sprintf("node-%d", id)
				_ = dag.AddNode(ctx, &WorkflowNode{ID: nodeID, Name: nodeID})
			}(i)
		}
		wg.Wait()

		statuses := dag.GetAllStatuses(ctx)
		if len(statuses) != 100 {
			t.Errorf("expected 100 nodes, got %d", len(statuses))
		}
	})

	t.Run("concurrent status updates", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("concurrent", "Concurrent", true)

		for i := 0; i < 100; i++ {
			nodeID := fmt.Sprintf("node-%d", i)
			_ = dag.AddNode(ctx, &WorkflowNode{ID: nodeID, Name: nodeID})
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				nodeID := fmt.Sprintf("node-%d", id)
				_ = dag.MarkCompleted(ctx, nodeID, &TaskOutput{Result: []byte("done")})
			}(i)
		}
		wg.Wait()

		statuses := dag.GetAllStatuses(ctx)
		for _, status := range statuses {
			if status != StatusCompleted {
				t.Errorf("expected all nodes completed, found %q", status)
				break
			}
		}
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("concurrent", "Concurrent", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				_, _ = dag.GetNode(ctx, "a")
			}()
			go func() {
				defer wg.Done()
				_ = dag.MarkCompleted(ctx, "a", &TaskOutput{Result: []byte("done")})
				_ = dag.AddNode(ctx, &WorkflowNode{ID: fmt.Sprintf("node-%d", i), Name: "Extra"})
			}()
		}
		wg.Wait()
	})
}

func TestGetNode(t *testing.T) {
	t.Run("get existing node", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A", Agent: "agent-1"})

		node, err := dag.GetNode(ctx, "a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if node.ID != "a" {
			t.Errorf("expected ID 'a', got %q", node.ID)
		}
		if node.Agent != "agent-1" {
			t.Errorf("expected Agent 'agent-1', got %q", node.Agent)
		}
	})

	t.Run("get nonexistent node", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)

		_, err := dag.GetNode(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent node, got nil")
		}
	})
}

func TestResetNodeForRetry(t *testing.T) {
	t.Run("resets failed node for retry", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A", MaxRetries: 3})

		_ = dag.MarkFailed(ctx, "a", fmt.Errorf("error"))

		if err := dag.ResetNodeForRetry(ctx, "a"); err != nil {
			t.Fatalf("failed to reset node: %v", err)
		}

		node, _ := dag.GetNode(ctx, "a")
		if node.Status != StatusPending {
			t.Errorf("expected status %q, got %q", StatusPending, node.Status)
		}
		if node.Output != nil {
			t.Error("expected output to be cleared")
		}
	})

	t.Run("fails for nonexistent node", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)

		if err := dag.ResetNodeForRetry(ctx, "nonexistent"); err == nil {
			t.Error("expected error for nonexistent node, got nil")
		}
	})
}

func TestGetDownstreamNodes(t *testing.T) {
	t.Run("gets downstream nodes", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "d", Name: "D", DependsOn: []string{"b"}})

		downstream, err := dag.GetDownstreamNodes(ctx, "a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(downstream) != 2 {
			t.Fatalf("expected 2 downstream nodes, got %d: %v", len(downstream), downstream)
		}

		downstreamSet := make(map[string]bool)
		for _, id := range downstream {
			downstreamSet[id] = true
		}
		if !downstreamSet["b"] || !downstreamSet["c"] {
			t.Errorf("expected downstream nodes [b, c], got %v", downstream)
		}
	})

	t.Run("no downstream nodes", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)

		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B"})

		downstream, err := dag.GetDownstreamNodes(ctx, "a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(downstream) != 0 {
			t.Errorf("expected 0 downstream nodes, got %d", len(downstream))
		}
	})

	t.Run("fails for nonexistent node", func(t *testing.T) {
		ctx := context.Background()
		dag := NewWorkflowDAG("test", "Test", true)

		_, err := dag.GetDownstreamNodes(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent node, got nil")
		}
	})
}
