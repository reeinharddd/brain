package workflow

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockExecutor is a mock implementation of NodeExecutor for testing
type mockExecutor struct {
	mu            sync.Mutex
	executed      []string
	outputs       map[string]*TaskOutput
	errors        map[string]error
	executionTime map[string]time.Duration
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		executed:      make([]string, 0),
		outputs:       make(map[string]*TaskOutput),
		errors:        make(map[string]error),
		executionTime: make(map[string]time.Duration),
	}
}

func (m *mockExecutor) Execute(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.executed = append(m.executed, node.ID)

	if execTime, ok := m.executionTime[node.ID]; ok {
		select {
		case <-time.After(execTime):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if err, ok := m.errors[node.ID]; ok {
		return m.outputs[node.ID], err
	}

	output := m.outputs[node.ID]
	if output == nil {
		output = &TaskOutput{
			Result:     []byte(fmt.Sprintf("result-%s", node.ID)),
			TokensUsed: 100,
			CostUSD:    0.01,
		}
	}

	return output, nil
}

func (m *mockExecutor) SetOutput(nodeID string, output *TaskOutput) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputs[nodeID] = output
}

func (m *mockExecutor) SetError(nodeID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[nodeID] = err
}

func (m *mockExecutor) SetExecutionTime(nodeID string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executionTime[nodeID] = duration
}

func (m *mockExecutor) GetExecuted() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.executed))
	copy(result, m.executed)
	return result
}

func TestNewExecutionEngine(t *testing.T) {
	t.Run("creates engine with valid maxParallel", func(t *testing.T) {
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 4)

		if engine.maxParallel != 4 {
			t.Errorf("expected maxParallel 4, got %d", engine.maxParallel)
		}
	})

	t.Run("enforces minimum maxParallel of 1", func(t *testing.T) {
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 0)

		if engine.maxParallel != 1 {
			t.Errorf("expected maxParallel 1, got %d", engine.maxParallel)
		}
	})

	t.Run("handles negative maxParallel", func(t *testing.T) {
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, -5)

		if engine.maxParallel != 1 {
			t.Errorf("expected maxParallel 1, got %d", engine.maxParallel)
		}
	})
}

func TestRunDAG_Linear(t *testing.T) {
	t.Run("executes linear workflow in order", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 1)

		dag := NewWorkflowDAG("linear", "Linear", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		executed := executor.GetExecuted()
		if len(executed) != 3 {
			t.Fatalf("expected 3 executions, got %d", len(executed))
		}

		// Verify order: a before b before c
		if executed[0] != "a" || executed[1] != "b" || executed[2] != "c" {
			t.Errorf("expected execution order [a, b, c], got %v", executed)
		}
	})
}

func TestRunDAG_Diamond(t *testing.T) {
	t.Run("executes diamond workflow correctly", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 2)

		dag := NewWorkflowDAG("diamond", "Diamond", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "d", Name: "D", DependsOn: []string{"b", "c"}})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		executed := executor.GetExecuted()
		if len(executed) != 4 {
			t.Fatalf("expected 4 executions, got %d", len(executed))
		}

		// Verify: a must be first, d must be last
		if executed[0] != "a" {
			t.Errorf("expected first execution to be 'a', got %q", executed[0])
		}
		if executed[3] != "d" {
			t.Errorf("expected last execution to be 'd', got %q", executed[3])
		}
	})
}

func TestRunDAG_Parallel(t *testing.T) {
	t.Run("executes parallel nodes concurrently", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 3)

		dag := NewWorkflowDAG("parallel", "Parallel", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "x", Name: "X"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "y", Name: "Y"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "z", Name: "Z"})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		executed := executor.GetExecuted()
		if len(executed) != 3 {
			t.Fatalf("expected 3 executions, got %d", len(executed))
		}
	})
}

func TestRunDAG_NodeFailure(t *testing.T) {
	t.Run("skips downstream nodes on failure", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		executor.SetError("b", fmt.Errorf("node b failed"))
		engine := NewExecutionEngine(executor, 2)

		dag := NewWorkflowDAG("failure", "Failure", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "d", Name: "D", DependsOn: []string{"b"}})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("expected status 'failed', got %q", result.Status)
		}

		executed := executor.GetExecuted()
		// Only a and b should be executed; c and d skipped
		if len(executed) != 2 {
			t.Errorf("expected 2 executions (c and d skipped), got %d: %v", len(executed), executed)
		}

		statuses := dag.GetAllStatuses(ctx)
		if statuses["c"] != StatusSkipped {
			t.Errorf("expected node c to be skipped, got %q", statuses["c"])
		}
		if statuses["d"] != StatusSkipped {
			t.Errorf("expected node d to be skipped, got %q", statuses["d"])
		}
	})

	t.Run("continues execution when non-dependent node fails", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		executor.SetError("b", fmt.Errorf("node b failed"))
		engine := NewExecutionEngine(executor, 2)

		dag := NewWorkflowDAG("partial-fail", "Partial Fail", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C"})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("expected status 'failed', got %q", result.Status)
		}

		// All three should be attempted (they're independent)
		executed := executor.GetExecuted()
		if len(executed) != 3 {
			t.Errorf("expected 3 executions, got %d: %v", len(executed), executed)
		}
	})
}

func TestRunDAG_Retry(t *testing.T) {
	t.Run("retries node on failure up to maxRetries", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		// First two calls fail, third succeeds
		callCount := 0
		executorFn := func(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
			executor.mu.Lock()
			defer executor.mu.Unlock()
			executor.executed = append(executor.executed, node.ID)

			callCount++
			if callCount <= 2 {
				return nil, fmt.Errorf("transient error")
			}
			return &TaskOutput{Result: []byte("success"), TokensUsed: 100, CostUSD: 0.01}, nil
		}

		dag := NewWorkflowDAG("retry", "Retry", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A", MaxRetries: 3})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})

		engine := &ExecutionEngine{
			executor:    &simpleExecutor{fn: executorFn},
			maxParallel: 1,
		}

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		executed := executor.GetExecuted()
		// "a" should be called 3 times (2 failures + 1 success), "b" once
		if len(executed) != 4 {
			t.Errorf("expected 4 executions (3 retries for a + 1 for b), got %d: %v", len(executed), executed)
		}
	})

	t.Run("fails after exhausting retries", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		executor.SetError("a", fmt.Errorf("persistent error"))
		engine := NewExecutionEngine(executor, 1)

		dag := NewWorkflowDAG("retry-fail", "Retry Fail", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A", MaxRetries: 2})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("expected status 'failed', got %q", result.Status)
		}

		executed := executor.GetExecuted()
		// "a" should be called 3 times (1 initial + 2 retries), "b" never
		if len(executed) != 3 {
			t.Errorf("expected 3 executions (1 + 2 retries), got %d: %v", len(executed), executed)
		}
	})
}

// simpleExecutor is a simple function-based executor for testing
type simpleExecutor struct {
	fn func(ctx context.Context, node *WorkflowNode) (*TaskOutput, error)
}

func (s *simpleExecutor) Execute(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
	return s.fn(ctx, node)
}

func TestRunDAG_ContextCancellation(t *testing.T) {
	t.Run("stops execution on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Use a slow executor so cancellation happens mid-execution
		var mu sync.Mutex
		var executed []string

		executorFn := func(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
			mu.Lock()
			executed = append(executed, node.ID)
			mu.Unlock()

			// Simulate slow execution
			select {
			case <-time.After(200 * time.Millisecond):
				return &TaskOutput{Result: []byte("done"), TokensUsed: 100, CostUSD: 0.01}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		engine := &ExecutionEngine{
			executor:    &simpleExecutor{fn: executorFn},
			maxParallel: 1,
		}

		dag := NewWorkflowDAG("cancel", "Cancel", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "c", Name: "C", DependsOn: []string{"b"}})

		// Cancel after first node starts executing
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("expected status 'failed', got %q", result.Status)
		}

		// Should have executed at least 1 node
		mu.Lock()
		count := len(executed)
		mu.Unlock()
		if count < 1 {
			t.Error("expected at least 1 execution before cancellation")
		}
	})

	t.Run("immediate cancellation returns error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 1)

		dag := NewWorkflowDAG("cancel-imm", "Cancel Immediate", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		_, err := engine.RunDAG(ctx, dag)
		if err == nil {
			t.Error("expected error for cancelled context, got nil")
		}
	})
}

func TestRunDAG_EmptyWorkflow(t *testing.T) {
	t.Run("handles empty DAG", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 1)

		dag := NewWorkflowDAG("empty", "Empty", true)

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		if len(result.NodeResults) != 0 {
			t.Errorf("expected 0 node results, got %d", len(result.NodeResults))
		}
	})
}

func TestRunDAG_MaxParallelEnforcement(t *testing.T) {
	t.Run("respects maxParallel limit", func(t *testing.T) {
		ctx := context.Background()
		var mu sync.Mutex
		var maxConcurrent int
		var currentConcurrent int

		executorFn := func(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond) // Simulate work

			mu.Lock()
			currentConcurrent--
			mu.Unlock()

			return &TaskOutput{Result: []byte("done"), TokensUsed: 50, CostUSD: 0.005}, nil
		}

		dag := NewWorkflowDAG("parallel-limit", "Parallel Limit", true)
		for i := 0; i < 10; i++ {
			_ = dag.AddNode(ctx, &WorkflowNode{ID: fmt.Sprintf("node-%d", i), Name: fmt.Sprintf("Node %d", i)})
		}

		engine := &ExecutionEngine{
			executor:    &simpleExecutor{fn: executorFn},
			maxParallel: 3,
		}

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		if maxConcurrent > 3 {
			t.Errorf("max concurrent executions was %d, expected at most 3", maxConcurrent)
		}
	})
}

func TestRunDAG_ExecutionResultAggregation(t *testing.T) {
	t.Run("aggregates duration, tokens, and cost", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		executor.SetOutput("a", &TaskOutput{Result: []byte("a"), TokensUsed: 100, CostUSD: 0.01, Duration: 100 * time.Millisecond})
		executor.SetOutput("b", &TaskOutput{Result: []byte("b"), TokensUsed: 200, CostUSD: 0.02, Duration: 200 * time.Millisecond})
		engine := NewExecutionEngine(executor, 1)

		dag := NewWorkflowDAG("agg", "Aggregation", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B", DependsOn: []string{"a"}})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.TotalTokens != 300 {
			t.Errorf("expected TotalTokens 300, got %d", result.TotalTokens)
		}

		expectedCost := 0.03
		if result.TotalCost < expectedCost-0.001 || result.TotalCost > expectedCost+0.001 {
			t.Errorf("expected TotalCost ~0.03, got %f", result.TotalCost)
		}

		if result.Duration <= 0 {
			t.Error("expected positive Duration")
		}
	})

	t.Run("includes error message in result", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		executor.SetError("a", fmt.Errorf("critical failure"))
		engine := NewExecutionEngine(executor, 1)

		dag := NewWorkflowDAG("error-result", "Error Result", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Error == "" {
			t.Error("expected error message in result")
		}
	})
}

func TestRunDAG_NodeResults(t *testing.T) {
	t.Run("collects all node results", func(t *testing.T) {
		ctx := context.Background()
		executor := newMockExecutor()
		engine := NewExecutionEngine(executor, 2)

		dag := NewWorkflowDAG("results", "Results", true)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "b", Name: "B"})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.NodeResults) != 2 {
			t.Errorf("expected 2 node results, got %d", len(result.NodeResults))
		}

		if _, ok := result.NodeResults["a"]; !ok {
			t.Error("missing result for node a")
		}
		if _, ok := result.NodeResults["b"]; !ok {
			t.Error("missing result for node b")
		}
	})
}

func TestRunDAG_ContextValidation(t *testing.T) {
	t.Run("propagates context to executor", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		executor := newMockExecutor()

		// Use a custom executor that checks context
		executorFn := func(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return executor.Execute(ctx, node)
		}

		engine := &ExecutionEngine{
			executor:    &simpleExecutor{fn: executorFn},
			maxParallel: 1,
		}

		dag := NewWorkflowDAG("ctx-prop", "Context Prop", false)
		_ = dag.AddNode(ctx, &WorkflowNode{ID: "a", Name: "A"})

		result, err := engine.RunDAG(ctx, dag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", result.Status)
		}

		cancel() // Cleanup
	})
}
