package delegation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DelegationMode defines how agents delegate
type DelegationMode string

const (
	DelegationFork       DelegationMode = "fork"        // Isolated context
	DelegationTeammate   DelegationMode = "teammate"    // Shared context
	DelegationWorktree   DelegationMode = "worktree"    // Parallel git worktrees
	DelegationConcurrent DelegationMode = "concurrent"  // Parallel threads
)

// FailureCondition defines when fallback activates
type FailureCondition string

const (
	FailureTimeout         FailureCondition = "timeout"
	FailureError           FailureCondition = "error"
	FailureTokenLimit      FailureCondition = "token_limit"
	FailurePolicyViolation FailureCondition = "policy_violation"
)

// FallbackAction defines what to do on failure
type FallbackAction string

const (
	ActionRetry     FallbackAction = "retry"
	ActionEscalate  FallbackAction = "escalate"
	ActionSimplify  FallbackAction = "simplify"
	ActionAbort     FallbackAction = "abort"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries    int
	BackoffBase   time.Duration
	BackoffFactor float64 // exponential backoff multiplier
}

// Constraint defines a constraint for an agent node
type Constraint struct {
	Type  string
	Value string
}

// TaskInput defines the input for a delegation node
type TaskInput struct {
	Description string
	Parameters  map[string]string
	Context     []byte // serialized context
}

// AgentNode represents a node in the delegation graph
type AgentNode struct {
	ID          string
	AgentID     string
	Role        string         // "architect", "builder", "reviewer", etc.
	Input       TaskInput
	Constraints []Constraint
	Timeout     time.Duration
	RetryPolicy RetryPolicy
	Metadata    map[string]string
}

// DelegationBudget defines budget limits for delegation execution
type DelegationBudget struct {
	MaxTokens   int
	MaxCostUSD  float64
	MaxDuration time.Duration
	MaxRetries  int
}

// FallbackStep defines a fallback action
type FallbackStep struct {
	Condition   FailureCondition
	Action      FallbackAction
	TargetAgent string
	Parameters  map[string]string
}

// FallbackChain defines ordered fallback steps
type FallbackChain struct {
	Steps []FallbackStep
}

// DelegationGraph represents the full delegation graph
type DelegationGraph struct {
	mu          sync.RWMutex
	ID          string
	RootAgent   string
	Nodes       map[string]*AgentNode
	Edges       map[string][]string  // parent -> children
	ReverseEdges map[string][]string // child -> parents (for efficient parent lookups)
	Mode        DelegationMode
	MaxDepth    int
	MaxParallel int
	Budget      DelegationBudget
	Fallback    FallbackChain
}

// Domain error types
var (
	ErrNodeNotFound        = errors.New("delegation: node not found")
	ErrCycleDetected       = errors.New("delegation: cycle detected in graph")
	ErrDuplicateNode       = errors.New("delegation: duplicate node ID")
	ErrDuplicateEdge       = errors.New("delegation: duplicate edge")
	ErrNoRootAgent         = errors.New("delegation: no root agent set")
	ErrInvalidEdge         = errors.New("delegation: invalid edge - node does not exist")
	ErrGraphNotInitialized = errors.New("delegation: graph not initialized")
)

// NewDelegationGraph creates an empty delegation graph
func NewDelegationGraph() *DelegationGraph {
	return &DelegationGraph{
		Nodes:        make(map[string]*AgentNode),
		Edges:        make(map[string][]string),
		ReverseEdges: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (g *DelegationGraph) AddNode(node *AgentNode) error {
	if node == nil {
		return fmt.Errorf("delegation: cannot add nil node: %w", ErrInvalidEdge)
	}
	if node.ID == "" {
		return fmt.Errorf("delegation: node ID cannot be empty: %w", ErrInvalidEdge)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[node.ID]; exists {
		return fmt.Errorf("delegation: node %q already exists: %w", node.ID, ErrDuplicateNode)
	}

	g.Nodes[node.ID] = node
	return nil
}

// AddEdge adds an edge from parent to child
func (g *DelegationGraph) AddEdge(parentID, childID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[parentID]; !exists {
		return fmt.Errorf("delegation: parent node %q does not exist: %w", parentID, ErrInvalidEdge)
	}
	if _, exists := g.Nodes[childID]; !exists {
		return fmt.Errorf("delegation: child node %q does not exist: %w", childID, ErrInvalidEdge)
	}

	// Check for duplicate edge
	for _, child := range g.Edges[parentID] {
		if child == childID {
			return fmt.Errorf("delegation: edge from %q to %q already exists: %w", parentID, childID, ErrDuplicateEdge)
		}
	}

	g.Edges[parentID] = append(g.Edges[parentID], childID)
	g.ReverseEdges[childID] = append(g.ReverseEdges[childID], parentID)

	return nil
}

// Validate checks the graph is a valid DAG (no cycles)
func (g *DelegationGraph) Validate() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Use DFS-based cycle detection
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(nodeID string) error
	dfs = func(nodeID string) error {
		visited[nodeID] = true
		recStack[nodeID] = true

		for _, child := range g.Edges[nodeID] {
			if !visited[child] {
				if err := dfs(child); err != nil {
					return err
				}
			} else if recStack[child] {
				return fmt.Errorf("delegation: cycle detected involving node %q: %w", child, ErrCycleDetected)
			}
		}

		recStack[nodeID] = false
		return nil
	}

	for nodeID := range g.Nodes {
		if !visited[nodeID] {
			if err := dfs(nodeID); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetRoot returns the root node
func (g *DelegationGraph) GetRoot() (*AgentNode, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.RootAgent == "" {
		return nil, fmt.Errorf("delegation: root agent not set: %w", ErrNoRootAgent)
	}

	node, exists := g.Nodes[g.RootAgent]
	if !exists {
		return nil, fmt.Errorf("delegation: root node %q not found: %w", g.RootAgent, ErrNodeNotFound)
	}

	return node, nil
}

// GetChildren returns children of a node
func (g *DelegationGraph) GetChildren(nodeID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	children := g.Edges[nodeID]
	if children == nil {
		return []string{}
	}

	// Return a copy to prevent mutation
	result := make([]string, len(children))
	copy(result, children)
	return result
}

// GetParents returns parents of a node
func (g *DelegationGraph) GetParents(nodeID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	parents := g.ReverseEdges[nodeID]
	if parents == nil {
		return []string{}
	}

	// Return a copy to prevent mutation
	result := make([]string, len(parents))
	copy(result, parents)
	return result
}

// GetExecutionOrder returns topological sort of nodes using Kahn's algorithm
func (g *DelegationGraph) GetExecutionOrder() ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.Nodes) == 0 {
		return []string{}, nil
	}

	// Calculate in-degrees
	inDegree := make(map[string]int)
	for nodeID := range g.Nodes {
		inDegree[nodeID] = 0
	}
	for _, children := range g.Edges {
		for _, child := range children {
			inDegree[child]++
		}
	}

	// Start with nodes that have no incoming edges
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	result := make([]string, 0, len(g.Nodes))

	for len(queue) > 0 {
		// Take first element from queue
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		// Reduce in-degree for all children
		for _, child := range g.Edges[nodeID] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("delegation: graph contains a cycle, cannot compute topological order: %w", ErrCycleDetected)
	}

	return result, nil
}

// Depth returns the maximum depth of the graph
func (g *DelegationGraph) Depth() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.Nodes) == 0 {
		return 0
	}

	maxDepth := 0
	visited := make(map[string]bool)

	var dfs func(nodeID string, currentDepth int)
	dfs = func(nodeID string, currentDepth int) {
		visited[nodeID] = true

		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}

		for _, child := range g.Edges[nodeID] {
			if !visited[child] {
				dfs(child, currentDepth+1)
			}
		}
		visited[nodeID] = false // Allow revisiting for different paths
	}

	// Start DFS from all root nodes (nodes with no parents)
	for nodeID := range g.Nodes {
		if len(g.ReverseEdges[nodeID]) == 0 {
			dfs(nodeID, 1)
		}
	}

	// If no root nodes found (all nodes have parents or single node), start from any node
	if maxDepth == 0 && len(g.Nodes) > 0 {
		for nodeID := range g.Nodes {
			dfs(nodeID, 1)
			break
		}
	}

	return maxDepth
}

// Ensure context.Context is used (imported but not directly used in this file)
var _ = context.Background
