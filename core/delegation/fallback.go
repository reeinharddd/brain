package delegation

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Fallback error types
var (
	ErrFallbackChainExhausted = errors.New("delegation: fallback chain exhausted")
	ErrFallbackAborted        = errors.New("delegation: fallback aborted")
	ErrContextCanceled        = errors.New("delegation: context canceled during fallback")
)

// FallbackResult describes the result of fallback execution
type FallbackResult struct {
	Step    FallbackStep
	Success bool
	Output  []byte
	Error   string
	Retries int
}

// AgentExecutor is the interface for executing agent tasks
type AgentExecutor interface {
	Execute(ctx context.Context, node *AgentNode) ([]byte, error)
}

// FallbackExecutor executes fallback chains
type FallbackExecutor struct {
	chain       FallbackChain
	agentExec   AgentExecutor
	budget      *BudgetTracker
}

// NewFallbackExecutor creates a new fallback executor
func NewFallbackExecutor(chain FallbackChain, executor AgentExecutor) *FallbackExecutor {
	return &FallbackExecutor{
		chain:     chain,
		agentExec: executor,
	}
}

// NewFallbackExecutorWithBudget creates a new fallback executor with budget tracking
func NewFallbackExecutorWithBudget(chain FallbackChain, executor AgentExecutor, budget *BudgetTracker) *FallbackExecutor {
	return &FallbackExecutor{
		chain:     chain,
		agentExec: executor,
		budget:    budget,
	}
}

// ExecuteWithFallback executes a node with fallback chain on failure
func (fe *FallbackExecutor) ExecuteWithFallback(ctx context.Context, node *AgentNode, failureType FailureCondition) (*FallbackResult, error) {
	// First, try to find a matching fallback step
	startIdx := -1
	for i, step := range fe.chain.Steps {
		if step.Condition == failureType {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		// No matching fallback step, return error
		return &FallbackResult{
			Success: false,
			Error:   fmt.Sprintf("no fallback step for condition: %s", failureType),
		}, fmt.Errorf("delegation: no fallback step for condition %q: %w", failureType, ErrFallbackChainExhausted)
	}

	// Execute fallback steps starting from the matching step
	var lastResult *FallbackResult
	var lastErr error
	for i := startIdx; i < len(fe.chain.Steps); i++ {
		step := fe.chain.Steps[i]

		// Check context cancellation
		select {
		case <-ctx.Done():
			return &FallbackResult{
				Step:    step,
				Success: false,
				Error:   ctx.Err().Error(),
			}, fmt.Errorf("delegation: context canceled: %w", ErrContextCanceled)
		default:
		}

		result, err := fe.executeStep(ctx, step, node)
		lastResult = result
		lastErr = err

		if err != nil {
			// Step failed, try next step if available
			continue
		}

		if result.Success {
			return result, nil
		}

		// Step didn't succeed, continue to next step
	}

	if lastResult == nil {
		return &FallbackResult{
			Success: false,
			Error:   "fallback chain exhausted without result",
		}, fmt.Errorf("delegation: fallback chain exhausted: %w", ErrFallbackChainExhausted)
	}

	if lastErr != nil {
		return lastResult, lastErr
	}

	return lastResult, fmt.Errorf("delegation: fallback chain exhausted: %w", ErrFallbackChainExhausted)
}

// executeStep executes a single fallback step
func (fe *FallbackExecutor) executeStep(ctx context.Context, step FallbackStep, node *AgentNode) (*FallbackResult, error) {
	retries := 0

	switch step.Action {
	case ActionRetry:
		return fe.executeRetry(ctx, step, node, &retries)

	case ActionEscalate:
		return fe.executeEscalate(ctx, step, node, &retries)

	case ActionSimplify:
		return fe.executeSimplify(ctx, step, node, &retries)

	case ActionAbort:
		return &FallbackResult{
			Step:    step,
			Success: false,
			Error:   "fallback action is abort",
		}, fmt.Errorf("delegation: fallback aborted: %w", ErrFallbackAborted)

	default:
		return &FallbackResult{
			Step:    step,
			Success: false,
			Error:   fmt.Sprintf("unknown fallback action: %s", step.Action),
		}, fmt.Errorf("delegation: unknown fallback action %q", step.Action)
	}
}

// executeRetry retries with the same agent
func (fe *FallbackExecutor) executeRetry(ctx context.Context, step FallbackStep, node *AgentNode, retries *int) (*FallbackResult, error) {
	maxRetries := node.RetryPolicy.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	backoffBase := node.RetryPolicy.BackoffBase
	if backoffBase <= 0 {
		backoffBase = 100 * time.Millisecond
	}
	backoffFactor := node.RetryPolicy.BackoffFactor
	if backoffFactor <= 0 {
		backoffFactor = 2.0
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		*retries = i + 1

		// Check budget if tracker is available
		if fe.budget != nil {
			if err := fe.budget.RecordRetry(); err != nil {
				return &FallbackResult{
					Step:    step,
					Success: false,
					Error:   fmt.Sprintf("retry budget exceeded: %v", err),
					Retries: *retries,
				}, err
			}
		}

		// Wait with backoff (skip on first retry)
		if i > 0 {
			backoff := time.Duration(float64(backoffBase) * pow(backoffFactor, float64(i-1)))
			select {
			case <-ctx.Done():
				return &FallbackResult{
					Step:    step,
					Success: false,
					Error:   ctx.Err().Error(),
					Retries: *retries,
				}, fmt.Errorf("delegation: context canceled during retry backoff: %w", ErrContextCanceled)
			case <-time.After(backoff):
			}
		}

		output, err := fe.agentExec.Execute(ctx, node)
		if err == nil {
			return &FallbackResult{
				Step:    step,
				Success: true,
				Output:  output,
				Retries: *retries,
			}, nil
		}
		lastErr = err
	}

	return &FallbackResult{
		Step:    step,
		Success: false,
		Error:   fmt.Sprintf("retry failed after %d attempts: %v", *retries, lastErr),
		Retries: *retries,
	}, fmt.Errorf("delegation: retry failed after %d attempts: %w", *retries, lastErr)
}

// executeEscalate escalates to a target agent
func (fe *FallbackExecutor) executeEscalate(ctx context.Context, step FallbackStep, node *AgentNode, retries *int) (*FallbackResult, error) {
	if step.TargetAgent == "" {
		return &FallbackResult{
			Step:    step,
			Success: false,
			Error:   "escalate action requires a target agent",
		}, fmt.Errorf("delegation: escalate requires target agent")
	}

	// Create a new node for the target agent
	escalatedNode := &AgentNode{
		ID:          "escalated_" + node.ID,
		AgentID:     step.TargetAgent,
		Role:        "escalated",
		Input:       node.Input,
		Constraints: node.Constraints,
		Timeout:     node.Timeout,
		RetryPolicy: node.RetryPolicy,
		Metadata:    mergeMetadata(node.Metadata, step.Parameters),
	}

	*retries = 1
	if fe.budget != nil {
		if err := fe.budget.RecordRetry(); err != nil {
			return &FallbackResult{
				Step:    step,
				Success: false,
				Error:   fmt.Sprintf("escalate budget exceeded: %v", err),
				Retries: *retries,
			}, err
		}
	}

	output, err := fe.agentExec.Execute(ctx, escalatedNode)
	if err != nil {
		return &FallbackResult{
			Step:    step,
			Success: false,
			Error:   fmt.Sprintf("escalate to %s failed: %v", step.TargetAgent, err),
			Retries: *retries,
		}, fmt.Errorf("delegation: escalate to %s failed: %w", step.TargetAgent, err)
	}

	return &FallbackResult{
		Step:    step,
		Success: true,
		Output:  output,
		Retries: *retries,
	}, nil
}

// executeSimplify simplifies the task and retries
func (fe *FallbackExecutor) executeSimplify(ctx context.Context, step FallbackStep, node *AgentNode, retries *int) (*FallbackResult, error) {
	// Create a simplified version of the node
	simplifiedNode := &AgentNode{
		ID:          "simplified_" + node.ID,
		AgentID:     node.AgentID,
		Role:        node.Role,
		Input:       simplifyInput(node.Input, step.Parameters),
		Constraints: node.Constraints,
		Timeout:     node.Timeout,
		RetryPolicy: node.RetryPolicy,
		Metadata:    mergeMetadata(node.Metadata, step.Parameters),
	}

	*retries = 1
	if fe.budget != nil {
		if err := fe.budget.RecordRetry(); err != nil {
			return &FallbackResult{
				Step:    step,
				Success: false,
				Error:   fmt.Sprintf("simplify budget exceeded: %v", err),
				Retries: *retries,
			}, err
		}
	}

	output, err := fe.agentExec.Execute(ctx, simplifiedNode)
	if err != nil {
		return &FallbackResult{
			Step:    step,
			Success: false,
			Error:   fmt.Sprintf("simplify failed: %v", err),
			Retries: *retries,
		}, fmt.Errorf("delegation: simplify failed: %w", err)
	}

	return &FallbackResult{
		Step:    step,
		Success: true,
		Output:  output,
		Retries: *retries,
	}, nil
}

// simplifyInput creates a simplified version of the task input
func simplifyInput(input TaskInput, params map[string]string) TaskInput {
	simplified := TaskInput{
		Description: "[Simplified] " + input.Description,
		Parameters:  make(map[string]string),
		Context:     input.Context,
	}

	// Copy original parameters
	for k, v := range input.Parameters {
		simplified.Parameters[k] = v
	}

	// Add simplification parameters
	for k, v := range params {
		simplified.Parameters[k] = v
	}

	return simplified
}

// mergeMetadata merges two metadata maps
func mergeMetadata(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

// pow computes base^exp for float64 values
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	// Handle fractional exponents approximately
	if exp != float64(int(exp)) {
		// Simple approximation for non-integer exponents
		// This is good enough for backoff calculations
		remaining := exp - float64(int(exp))
		result *= base * remaining
	}
	return result
}
