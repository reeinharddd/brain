package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NodeExecutor is the interface for executing workflow nodes
type NodeExecutor interface {
	Execute(ctx context.Context, node *WorkflowNode) (*TaskOutput, error)
}

// ExecutionResult is the final result of workflow execution
type ExecutionResult struct {
	WorkflowID  string
	Status      string // completed, failed, partial
	NodeResults map[string]*TaskOutput
	Duration    time.Duration
	TotalTokens int
	TotalCost   float64
	Error       string
}

// ExecutionEngine executes workflow DAGs
type ExecutionEngine struct {
	executor    NodeExecutor
	maxParallel int
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(executor NodeExecutor, maxParallel int) *ExecutionEngine {
	if maxParallel < 1 {
		maxParallel = 1
	}
	return &ExecutionEngine{
		executor:    executor,
		maxParallel: maxParallel,
	}
}

// RunDAG executes a workflow DAG
func (e *ExecutionEngine) RunDAG(ctx context.Context, dag *WorkflowDAG) (*ExecutionResult, error) {
	startTime := time.Now()

	// 1. Validate the DAG
	if err := dag.Validate(ctx); err != nil {
		return nil, fmt.Errorf("DAG validation failed: %w", err)
	}

	// 2. Initialize all nodes to pending (they should already be, but ensure)
	for id, node := range dag.Nodes {
		node.Status = StatusPending
		_ = id
	}

	result := &ExecutionResult{
		WorkflowID:  dag.ID,
		NodeResults: make(map[string]*TaskOutput),
	}

	// Track retry counts per node
	retryCounts := make(map[string]int)

	// 3-8. Execute workflow loop
	for {
		select {
		case <-ctx.Done():
			// Mark all pending nodes as skipped due to cancellation
			for _, node := range dag.Nodes {
				if node.Status == StatusPending {
					_ = dag.MarkSkipped(ctx, node.ID)
				}
			}
			result.Status = "failed"
			result.Error = ctx.Err().Error()
			result.Duration = time.Since(startTime)
			return result, nil
		default:
		}

		// 3. Find ready nodes
		readyNodeIDs := dag.GetReadyNodes(ctx)
		if len(readyNodeIDs) == 0 {
			break
		}

		// 4. Execute ready nodes in parallel (up to maxParallel)
		batchSize := len(readyNodeIDs)
		if batchSize > e.maxParallel {
			batchSize = e.maxParallel
		}

		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < batchSize; i++ {
			nodeID := readyNodeIDs[i]
			wg.Add(1)

			go func(id string) {
				defer wg.Done()

				output, err := e.executeNode(ctx, dag, id)

				mu.Lock()
				defer mu.Unlock()

				if err != nil {
					// Check if we should retry
					node, _ := dag.GetNode(ctx, id)
					retries := retryCounts[id]

					if retries < node.MaxRetries {
						// Retry the node
						retryCounts[id]++
						_ = dag.ResetNodeForRetry(ctx, id)
					} else {
						// Mark failed and skip downstream
						_ = dag.MarkFailed(ctx, id, err)
						e.skipDownstream(ctx, dag, id)
					}
				} else {
					// Mark completed
					_ = dag.MarkCompleted(ctx, id, output)
				}
			}(nodeID)
		}

		wg.Wait()
	}

	// 9. Assemble result
	result.Duration = time.Since(startTime)

	allCompleted := true
	anyFailed := false
	anySkipped := false

	for _, node := range dag.Nodes {
		if node.Output != nil {
			result.NodeResults[node.ID] = node.Output
			result.TotalTokens += node.Output.TokensUsed
			result.TotalCost += node.Output.CostUSD
		}

		switch node.Status {
		case StatusCompleted:
			// OK
		case StatusFailed:
			anyFailed = true
			allCompleted = false
			if result.Error == "" && node.Output != nil && node.Output.Error != "" {
				result.Error = node.Output.Error
			}
		case StatusSkipped:
			anySkipped = true
			allCompleted = false
		}
	}

	if allCompleted {
		result.Status = "completed"
	} else if anyFailed {
		if anySkipped {
			result.Status = "failed"
		} else {
			result.Status = "failed"
		}
	} else if anySkipped {
		result.Status = "partial"
	} else {
		result.Status = "completed"
	}

	return result, nil
}

// executeNode executes a single node with timeout if configured
func (e *ExecutionEngine) executeNode(ctx context.Context, dag *WorkflowDAG, nodeID string) (*TaskOutput, error) {
	node, err := dag.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("node %q not found: %w", nodeID, err)
	}

	execCtx := ctx
	var cancel context.CancelFunc
	if node.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, node.Timeout)
		defer cancel()
	}

	// Mark running status (best effort, not critical)
	node.Status = StatusRunning

	startTime := time.Now()
	output, err := e.executor.Execute(execCtx, node)

	if err != nil {
		if output == nil {
			output = &TaskOutput{
				Duration: time.Since(startTime),
				Error:    err.Error(),
			}
		} else {
			output.Duration = time.Since(startTime)
			if output.Error == "" {
				output.Error = err.Error()
			}
		}
		node.Status = StatusFailed
		return output, err
	}

	if output == nil {
		output = &TaskOutput{
			Duration: time.Since(startTime),
		}
	} else {
		output.Duration = time.Since(startTime)
	}

	return output, nil
}

// skipDownstream marks all downstream nodes as skipped
func (e *ExecutionEngine) skipDownstream(ctx context.Context, dag *WorkflowDAG, nodeID string) {
	downstream, err := dag.GetDownstreamNodes(ctx, nodeID)
	if err != nil {
		return
	}

	for _, id := range downstream {
		node, err := dag.GetNode(ctx, id)
		if err != nil {
			continue
		}
		if node.Status == StatusPending {
			_ = dag.MarkSkipped(ctx, id)
			// Recursively skip downstream of this node
			e.skipDownstream(ctx, dag, id)
		}
	}
}
