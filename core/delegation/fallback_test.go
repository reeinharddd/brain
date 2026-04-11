package delegation

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// mockExecutor is a mock implementation of AgentExecutor for testing
type mockExecutor struct {
	executeFunc func(ctx context.Context, node *AgentNode) ([]byte, error)
	callCount   int
}

func (m *mockExecutor) Execute(ctx context.Context, node *AgentNode) ([]byte, error) {
	m.callCount++
	if m.executeFunc != nil {
		return m.executeFunc(ctx, node)
	}
	return []byte("success"), nil
}

// failingExecutor always fails
type failingExecutor struct {
	callCount int
	err       error
}

func (f *failingExecutor) Execute(ctx context.Context, node *AgentNode) ([]byte, error) {
	f.callCount++
	if f.err != nil {
		return nil, f.err
	}
	return nil, fmt.Errorf("executor failed")
}

// successAfterNErrorsExecutor succeeds after N failures
type successAfterNErrorsExecutor struct {
	callCount   int
	failures    int
	maxFailures int
}

func (s *successAfterNErrorsExecutor) Execute(ctx context.Context, node *AgentNode) ([]byte, error) {
	s.callCount++
	if s.callCount <= s.maxFailures {
		s.failures++
		return nil, fmt.Errorf("attempt %d failed", s.callCount)
	}
	return []byte("success after retries"), nil
}

func TestExecuteWithFallback_Retry(t *testing.T) {
	t.Run("retry succeeds on first attempt", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    3,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if exec.callCount != 1 {
			t.Errorf("expected 1 call, got %d", exec.callCount)
		}
	})

	t.Run("retry succeeds after multiple attempts", func(t *testing.T) {
		exec := &successAfterNErrorsExecutor{maxFailures: 2}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    5,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if exec.callCount != 3 {
			t.Errorf("expected 3 calls (2 failures + 1 success), got %d", exec.callCount)
		}
		if result.Retries != 3 {
			t.Errorf("expected 3 retries, got %d", result.Retries)
		}
	})

	t.Run("retry exhausts all attempts", func(t *testing.T) {
		exec := &failingExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    2,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
		if exec.callCount != 2 {
			t.Errorf("expected 2 calls, got %d", exec.callCount)
		}
	})
}

func TestExecuteWithFallback_Escalate(t *testing.T) {
	t.Run("escalate succeeds", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition:   FailureTimeout,
					Action:      ActionEscalate,
					TargetAgent: "senior-agent",
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "junior-agent",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureTimeout)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if exec.callCount != 1 {
			t.Errorf("expected 1 call, got %d", exec.callCount)
		}
	})

	t.Run("escalate without target agent fails", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureTimeout,
					Action:    ActionEscalate,
					// No TargetAgent
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "junior-agent",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureTimeout)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})

	t.Run("escalate with failing executor", func(t *testing.T) {
		exec := &failingExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition:   FailureError,
					Action:      ActionEscalate,
					TargetAgent: "senior-agent",
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "junior-agent",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})
}

func TestExecuteWithFallback_Simplify(t *testing.T) {
	t.Run("simplify succeeds", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition:  FailureTokenLimit,
					Action:     ActionSimplify,
					Parameters: map[string]string{"max_tokens": "500"},
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			Input: TaskInput{
				Description: "complex task",
				Parameters:  map[string]string{"detail": "high"},
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureTokenLimit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if exec.callCount != 1 {
			t.Errorf("expected 1 call, got %d", exec.callCount)
		}
	})

	t.Run("simplify with failing executor", func(t *testing.T) {
		exec := &failingExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition:  FailureError,
					Action:     ActionSimplify,
					Parameters: map[string]string{"max_tokens": "500"},
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})
}

func TestExecuteWithFallback_Abort(t *testing.T) {
	t.Run("abort returns failure", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailurePolicyViolation,
					Action:    ActionAbort,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailurePolicyViolation)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
		if !errors.Is(err, ErrFallbackAborted) {
			t.Errorf("expected ErrFallbackAborted, got %v", err)
		}
		if exec.callCount != 0 {
			t.Errorf("expected 0 calls (aborted immediately), got %d", exec.callCount)
		}
	})
}

func TestExecuteWithFallback_ChainExhaustion(t *testing.T) {
	t.Run("no matching step", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureTimeout,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrFallbackChainExhausted) {
			t.Errorf("expected ErrFallbackChainExhausted, got %v", err)
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})

	t.Run("empty chain", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrFallbackChainExhausted) {
			t.Errorf("expected ErrFallbackChainExhausted, got %v", err)
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})

	t.Run("all steps fail", func(t *testing.T) {
		exec := &failingExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition:   FailureError,
					Action:      ActionRetry,
				},
				{
					Condition:   FailureError,
					Action:      ActionEscalate,
					TargetAgent: "senior-agent",
				},
				{
					Condition: FailureError,
					Action:    ActionAbort,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    1,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Should have tried retry (failed), then escalate (failed), then abort
		if result.Success {
			t.Error("expected failure, got success")
		}
	})
}

func TestExecuteWithFallback_ContextCancellation(t *testing.T) {
	t.Run("context canceled before execution", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := fe.ExecuteWithFallback(ctx, node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrContextCanceled) {
			t.Errorf("expected ErrContextCanceled, got %v", err)
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})

	t.Run("context canceled during execution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		execFunc := func(ctx context.Context, node *AgentNode) ([]byte, error) {
			cancel() // Cancel context during execution
			// Wait a bit to ensure context is canceled
			time.Sleep(10 * time.Millisecond)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return []byte("success"), nil
			}
		}
		exec := &mockExecutor{executeFunc: execFunc}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    3,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(ctx, node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrContextCanceled) {
			t.Errorf("expected ErrContextCanceled, got %v", err)
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})
}

func TestExecuteWithFallback_CustomExecutor(t *testing.T) {
	t.Run("custom executor with node inspection", func(t *testing.T) {
		execFunc := func(ctx context.Context, node *AgentNode) ([]byte, error) {
			// Custom logic based on node properties
			if node.AgentID == "special-agent" {
				return []byte("special result"), nil
			}
			return []byte("default result"), nil
		}
		exec := &mockExecutor{executeFunc: execFunc}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition:   FailureError,
					Action:      ActionEscalate,
					TargetAgent: "special-agent",
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "regular-agent",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if string(result.Output) != "special result" {
			t.Errorf("expected 'special result', got %q", string(result.Output))
		}
	})
}

func TestExecuteWithFallback_MultiStepChain(t *testing.T) {
	t.Run("first step succeeds", func(t *testing.T) {
		exec := &mockExecutor{}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
				{
					Condition:   FailureError,
					Action:      ActionEscalate,
					TargetAgent: "senior-agent",
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if exec.callCount != 1 {
			t.Errorf("expected 1 call, got %d", exec.callCount)
		}
	})

	t.Run("first step fails, second succeeds", func(t *testing.T) {
		// First call fails, subsequent calls succeed
		callCount := 0
		execFunc := func(ctx context.Context, node *AgentNode) ([]byte, error) {
			callCount++
			if node.AgentID == "agent1" {
				return nil, fmt.Errorf("agent1 failed")
			}
			return []byte("escalated success"), nil
		}
		exec := &mockExecutor{executeFunc: execFunc}
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
				{
					Condition:   FailureError,
					Action:      ActionEscalate,
					TargetAgent: "senior-agent",
				},
			},
		}
		fe := NewFallbackExecutor(chain, exec)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    1,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if string(result.Output) != "escalated success" {
			t.Errorf("expected 'escalated success', got %q", string(result.Output))
		}
	})
}

func TestFallbackExecutor_Budget(t *testing.T) {
	t.Run("budget tracking with retry", func(t *testing.T) {
		exec := &successAfterNErrorsExecutor{maxFailures: 1}
		budget := DelegationBudget{
			MaxRetries: 5,
		}
		bt := NewBudgetTracker(budget)
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutorWithBudget(chain, exec, bt)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    3,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success, got failure: %s", result.Error)
		}
		if bt.GetRetriesUsed() < 1 {
			t.Errorf("expected at least 1 retry used, got %d", bt.GetRetriesUsed())
		}
	})

	t.Run("budget exceeded during retry", func(t *testing.T) {
		exec := &failingExecutor{}
		budget := DelegationBudget{
			MaxRetries: 1,
		}
		bt := NewBudgetTracker(budget)
		chain := FallbackChain{
			Steps: []FallbackStep{
				{
					Condition: FailureError,
					Action:    ActionRetry,
				},
			},
		}
		fe := NewFallbackExecutorWithBudget(chain, exec, bt)

		node := &AgentNode{
			ID:      "test",
			AgentID: "agent1",
			RetryPolicy: RetryPolicy{
				MaxRetries:    5,
				BackoffBase:   1 * time.Millisecond,
				BackoffFactor: 1.0,
			},
		}

		result, err := fe.ExecuteWithFallback(context.Background(), node, FailureError)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Success {
			t.Error("expected failure, got success")
		}
	})
}
