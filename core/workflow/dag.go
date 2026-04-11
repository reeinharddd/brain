package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// NodeStatus represents the execution state of a workflow node
type NodeStatus string

const (
	StatusPending   NodeStatus = "pending"
	StatusRunning   NodeStatus = "running"
	StatusCompleted NodeStatus = "completed"
	StatusFailed    NodeStatus = "failed"
	StatusSkipped   NodeStatus = "skipped"
)

// TaskOutput represents the output of an executed node
type TaskOutput struct {
	Result     []byte
	Error      string
	Duration   time.Duration
	TokensUsed int
	CostUSD    float64
}

// WorkflowNode represents a single step in a workflow
type WorkflowNode struct {
	ID         string
	Name       string
	Agent      string // Agent pool member to execute
	Input      map[string]string
	Output     *TaskOutput
	DependsOn  []string // Node IDs this depends on
	Timeout    time.Duration
	MaxRetries int
	Status     NodeStatus
}

// WorkflowDAG represents a complete workflow as a DAG
type WorkflowDAG struct {
	ID         string
	Name       string
	Nodes      map[string]*WorkflowNode
	Parallel   bool // Allow parallel execution
	mu         sync.RWMutex
}

// NewWorkflowDAG creates a new workflow DAG
func NewWorkflowDAG(id, name string, parallel bool) *WorkflowDAG {
	return &WorkflowDAG{
		ID:       id,
		Name:     name,
		Nodes:    make(map[string]*WorkflowNode),
		Parallel: parallel,
	}
}

// AddNode adds a node to the workflow DAG
func (d *WorkflowDAG) AddNode(ctx context.Context, node *WorkflowNode) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}
	if node.ID == "" {
		return errors.New("node ID cannot be empty")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Nodes[node.ID]; exists {
		return fmt.Errorf("node with ID %q already exists", node.ID)
	}

	// Initialize node status to pending
	node.Status = StatusPending
	if node.Input == nil {
		node.Input = make(map[string]string)
	}

	d.Nodes[node.ID] = node
	return nil
}

// Validate checks the DAG for cycles and missing dependencies
func (d *WorkflowDAG) Validate(ctx context.Context) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.Nodes) == 0 {
		return nil
	}

	// Check for missing dependencies
	for id, node := range d.Nodes {
		for _, depID := range node.DependsOn {
			if _, exists := d.Nodes[depID]; !exists {
				return fmt.Errorf("node %q has missing dependency %q", id, depID)
			}
		}
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(nodeID string) error
	hasCycle = func(nodeID string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		visited[nodeID] = true
		recStack[nodeID] = true

		node, exists := d.Nodes[nodeID]
		if !exists {
			return fmt.Errorf("node %q not found", nodeID)
		}

		for _, depID := range node.DependsOn {
			if !visited[depID] {
				if err := hasCycle(depID); err != nil {
					return err
				}
			} else if recStack[depID] {
				return fmt.Errorf("cycle detected involving node %q", nodeID)
			}
		}

		recStack[nodeID] = false
		return nil
	}

	for nodeID := range d.Nodes {
		if !visited[nodeID] {
			if err := hasCycle(nodeID); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetReadyNodes returns node IDs whose dependencies are all met (completed)
func (d *WorkflowDAG) GetReadyNodes(ctx context.Context) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var ready []string
	for id, node := range d.Nodes {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if node.Status != StatusPending {
			continue
		}

		allDepsMet := true
		for _, depID := range node.DependsOn {
			dep, exists := d.Nodes[depID]
			if !exists {
				allDepsMet = false
				break
			}
			if dep.Status != StatusCompleted {
				allDepsMet = false
				break
			}
		}

		if allDepsMet {
			ready = append(ready, id)
		}
	}

	return ready
}

// GetNode returns a node by ID
func (d *WorkflowDAG) GetNode(ctx context.Context, id string) (*WorkflowNode, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	node, exists := d.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %q not found", id)
	}

	return node, nil
}

// GetAllStatuses returns a map of all node IDs to their statuses
func (d *WorkflowDAG) GetAllStatuses(ctx context.Context) map[string]NodeStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()

	statuses := make(map[string]NodeStatus, len(d.Nodes))
	for id, node := range d.Nodes {
		statuses[id] = node.Status
	}

	return statuses
}

// MarkCompleted marks a node as completed with the given output
func (d *WorkflowDAG) MarkCompleted(ctx context.Context, id string, output *TaskOutput) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	node, exists := d.Nodes[id]
	if !exists {
		return fmt.Errorf("node %q not found", id)
	}

	node.Status = StatusCompleted
	node.Output = output
	return nil
}

// MarkFailed marks a node as failed
func (d *WorkflowDAG) MarkFailed(ctx context.Context, id string, err error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	node, exists := d.Nodes[id]
	if !exists {
		return fmt.Errorf("node %q not found", id)
	}

	node.Status = StatusFailed
	if node.Output == nil {
		node.Output = &TaskOutput{}
	}
	if err != nil {
		node.Output.Error = err.Error()
	}
	return nil
}

// MarkSkipped marks a node as skipped
func (d *WorkflowDAG) MarkSkipped(ctx context.Context, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	node, exists := d.Nodes[id]
	if !exists {
		return fmt.Errorf("node %q not found", id)
	}

	node.Status = StatusSkipped
	return nil
}

// GetRetries returns the max retries count for a node
func (d *WorkflowDAG) GetRetries(ctx context.Context, id string) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	node, exists := d.Nodes[id]
	if !exists {
		return 0, fmt.Errorf("node %q not found", id)
	}

	return node.MaxRetries, nil
}

// ResetNodeForRetry resets a node's status to pending for retry
func (d *WorkflowDAG) ResetNodeForRetry(ctx context.Context, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	node, exists := d.Nodes[id]
	if !exists {
		return fmt.Errorf("node %q not found", id)
	}

	node.Status = StatusPending
	node.Output = nil
	return nil
}

// GetDownstreamNodes returns all nodes that depend on the given node
func (d *WorkflowDAG) GetDownstreamNodes(ctx context.Context, id string) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if _, exists := d.Nodes[id]; !exists {
		return nil, fmt.Errorf("node %q not found", id)
	}

	var downstream []string
	for nodeID, node := range d.Nodes {
		if nodeID == id {
			continue
		}
		for _, depID := range node.DependsOn {
			if depID == id {
				downstream = append(downstream, nodeID)
				break
			}
		}
	}

	return downstream, nil
}
