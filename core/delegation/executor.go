package delegation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DelegationExecutor runs delegation graphs with real subprocess execution
type DelegationExecutor struct {
	mu            sync.RWMutex
	workDir       string
	logFunc       func(string)
	active        map[string]*DelegationExecution
	budgetTracker *budgetTrackerWrapper
}

// DelegationExecution tracks a running delegation
type DelegationExecution struct {
	ID        string
	Graph     *DelegationGraph
	Status    string // running, completed, failed, cancelled
	Results   map[string]*NodeResult
	StartTime time.Time
	EndTime   time.Time
	Error     string
	mu        sync.RWMutex
}

// NodeResult is the result of executing a single delegation node
type NodeResult struct {
	NodeID   string
	AgentID  string
	Output   string
	Error    string
	Duration time.Duration
	Tokens   int
	Cost     float64
	Status   string // completed, failed, skipped
}

// BudgetTracker wraps the budget.go BudgetTracker with additional convenience methods
type budgetTrackerWrapper struct {
	*BudgetTracker
}

func newBudgetTrackerWrapper(budget DelegationBudget) *budgetTrackerWrapper {
	return &budgetTrackerWrapper{
		BudgetTracker: NewBudgetTracker(budget),
	}
}

// WithinBudget checks if current usage is within limits
func (bt *budgetTrackerWrapper) WithinBudget() bool {
	return !bt.BudgetTracker.IsExceeded()
}

// GetUsage returns current usage (tokens, cost, duration)
func (bt *budgetTrackerWrapper) GetUsage() (int, float64, time.Duration) {
	return bt.BudgetTracker.GetTokensUsed(), bt.BudgetTracker.GetCostUsed(), time.Since(bt.BudgetTracker.startTime)
}

// Add records resource consumption
func (bt *budgetTrackerWrapper) Add(tokens int, cost float64, dur time.Duration) {
	_ = bt.BudgetTracker.RecordTokenUsage(tokens)
	_ = bt.BudgetTracker.RecordCost(cost)
}

// NewDelegationExecutor creates a new delegation executor
func NewDelegationExecutor(workDir string, logFunc func(string), budget DelegationBudget) *DelegationExecutor {
	return &DelegationExecutor{
		workDir:       workDir,
		logFunc:       logFunc,
		active:        make(map[string]*DelegationExecution),
		budgetTracker: newBudgetTrackerWrapper(budget),
	}
}

// Execute starts a delegation graph asynchronously
func (de *DelegationExecutor) Execute(ctx context.Context, graph *DelegationGraph) (string, error) {
	execID := uuid.New().String()

	execution := &DelegationExecution{
		ID:        execID,
		Graph:     graph,
		Status:    "running",
		Results:   make(map[string]*NodeResult),
		StartTime: time.Now(),
	}

	de.mu.Lock()
	de.active[execID] = execution
	de.mu.Unlock()

	go de.runExecution(ctx, execution)

	return execID, nil
}

func (de *DelegationExecutor) runExecution(ctx context.Context, exec *DelegationExecution) {
	defer func() {
		exec.mu.Lock()
		if exec.Status == "running" {
			exec.Status = "completed"
		}
		exec.EndTime = time.Now()
		exec.mu.Unlock()
	}()

	graph := exec.Graph

	// Get execution order (topological sort)
	order, err := graph.GetExecutionOrder()
	if err != nil {
		exec.mu.Lock()
		exec.Status = "failed"
		exec.Error = err.Error()
		exec.mu.Unlock()
		de.logFunc(fmt.Sprintf("[delegation:%s] Failed: %v", exec.ID, err))
		return
	}

	// Execute nodes in order
	for _, nodeID := range order {
		select {
		case <-ctx.Done():
			exec.mu.Lock()
			exec.Status = "cancelled"
			exec.Error = ctx.Err().Error()
			exec.mu.Unlock()
			return
		default:
		}

		// Check budget
		if !de.budgetTracker.WithinBudget() {
			de.logFunc(fmt.Sprintf("[delegation:%s] Budget exceeded at node %s", exec.ID, nodeID))
			// Try fallback
			if !de.applyFallback(exec, "token_limit") {
				exec.mu.Lock()
				exec.Status = "failed"
				exec.Error = "budget exceeded"
				exec.mu.Unlock()
				return
			}
		}

		node, ok := graph.Nodes[nodeID]
		if !ok {
			continue
		}

		result := de.executeNode(ctx, node)

		exec.mu.Lock()
		exec.Results[nodeID] = result
		exec.mu.Unlock()

		if result.Status == "failed" {
			de.logFunc(fmt.Sprintf("[delegation:%s] Node %s failed: %s", exec.ID, nodeID, result.Error))
			if !de.applyFallbackForNode(exec, nodeID, result) {
				exec.mu.Lock()
				exec.Status = "failed"
				exec.Error = fmt.Sprintf("node %s failed: %s", nodeID, result.Error)
				exec.mu.Unlock()
				return
			}
		}
	}
}

func (de *DelegationExecutor) executeNode(ctx context.Context, node *AgentNode) *NodeResult {
	start := time.Now()
	result := &NodeResult{
		NodeID:  node.ID,
		AgentID: node.AgentID,
		Status:  "completed",
	}

	// Build command - in production this invokes a real agent binary
	// For now, use echo as placeholder with agent role info
	cmd := exec.CommandContext(ctx, "echo", fmt.Sprintf(`[agent:%s][role:%s] Task: %s`, node.AgentID, node.Role, node.Input.Description))

	// Set environment variables
	cmd.Env = []string{
		fmt.Sprintf("BRAIN_AGENT_ID=%s", node.AgentID),
		fmt.Sprintf("BRAIN_AGENT_ROLE=%s", node.Role),
		fmt.Sprintf("BRAIN_DELEGATION_MODE=%s", "fork"), // Default mode
	}

	// Prepare input
	if node.Input.Description != "" || len(node.Input.Parameters) > 0 {
		inputData, _ := json.Marshal(node.Input)
		cmd.Stdin = bytes.NewReader(inputData)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)

	if stdout.Len() > 0 {
		result.Output = stdout.String()
		result.Tokens = len(result.Output) / 4
		result.Cost = float64(result.Tokens) / 1000000.0 * 10.0
	}

	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("%v: %s", err, stderr.String())
		return result
	}

	// Update budget tracker
	de.budgetTracker.Add(result.Tokens, result.Cost, result.Duration)

	return result
}

func (de *DelegationExecutor) applyFallback(exec *DelegationExecution, condition string) bool {
	for _, step := range exec.Graph.Fallback.Steps {
		// Map string condition to FailureCondition
		var fc FailureCondition
		switch condition {
		case "timeout":
			fc = FailureTimeout
		case "error":
			fc = FailureError
		case "token_limit":
			fc = FailureTokenLimit
		case "policy_violation":
			fc = FailurePolicyViolation
		}

		if step.Condition == fc {
			de.logFunc(fmt.Sprintf("[delegation:%s] Applying fallback: %s -> %s",
				exec.ID, step.Condition, step.Action))
			switch step.Action {
			case ActionAbort:
				return false
			case ActionRetry:
				return true
			case ActionSimplify:
				return true
			case ActionEscalate:
				return true
			}
		}
	}
	return false
}

func (de *DelegationExecutor) applyFallbackForNode(exec *DelegationExecution, nodeID string, result *NodeResult) bool {
	return de.applyFallback(exec, "error")
}

// GetExecution returns the current status of an execution
func (de *DelegationExecutor) GetExecution(execID string) (*DelegationExecution, error) {
	de.mu.RLock()
	defer de.mu.RUnlock()

	exec, ok := de.active[execID]
	if !ok {
		return nil, fmt.Errorf("execution not found: %s", execID)
	}
	return exec, nil
}

// CancelExecution cancels a running execution
func (de *DelegationExecutor) CancelExecution(execID string) error {
	de.mu.Lock()
	defer de.mu.Unlock()

	exec, ok := de.active[execID]
	if !ok {
		return fmt.Errorf("execution not found: %s", execID)
	}

	exec.mu.Lock()
	exec.Status = "cancelled"
	exec.EndTime = time.Now()
	exec.mu.Unlock()

	return nil
}

// GetActiveExecutions returns all active executions
func (de *DelegationExecutor) GetActiveExecutions() []*DelegationExecution {
	de.mu.RLock()
	defer de.mu.RUnlock()

	result := make([]*DelegationExecution, 0, len(de.active))
	for _, exec := range de.active {
		result = append(result, exec)
	}
	return result
}

// GetBudgetUsage returns current budget usage
func (de *DelegationExecutor) GetBudgetUsage() (tokens int, cost float64, duration time.Duration) {
	return de.budgetTracker.GetUsage()
}
