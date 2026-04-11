package delegation

import (
	"errors"
	"testing"
)

func TestNewDelegationGraph(t *testing.T) {
	t.Run("creates empty graph", func(t *testing.T) {
		g := NewDelegationGraph()
		if g == nil {
			t.Fatal("expected non-nil graph")
		}
		if g.Nodes == nil {
			t.Error("expected Nodes map to be initialized")
		}
		if g.Edges == nil {
			t.Error("expected Edges map to be initialized")
		}
		if g.ReverseEdges == nil {
			t.Error("expected ReverseEdges map to be initialized")
		}
	})
}

func TestAddNode(t *testing.T) {
	tests := []struct {
		name      string
		node      *AgentNode
		wantErr   bool
		errIs     error
	}{
		{
			name: "valid node",
			node: &AgentNode{
				ID:      "node1",
				AgentID: "agent1",
				Role:    "architect",
			},
			wantErr: false,
		},
		{
			name:    "nil node",
			node:    nil,
			wantErr: true,
			errIs:   ErrInvalidEdge,
		},
		{
			name: "empty node ID",
			node: &AgentNode{
				ID:      "",
				AgentID: "agent1",
			},
			wantErr: true,
			errIs:   ErrInvalidEdge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDelegationGraph()
			err := g.AddNode(tt.node)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if _, exists := g.Nodes[tt.node.ID]; !exists {
					t.Errorf("expected node %q to exist", tt.node.ID)
				}
			}
		})
	}

	t.Run("duplicate node", func(t *testing.T) {
		g := NewDelegationGraph()
		node := &AgentNode{ID: "node1", AgentID: "agent1"}
		if err := g.AddNode(node); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err := g.AddNode(node)
		if err == nil {
			t.Fatal("expected error for duplicate node")
		}
		if !errors.Is(err, ErrDuplicateNode) {
			t.Errorf("expected ErrDuplicateNode, got %v", err)
		}
	})
}

func TestAddEdge(t *testing.T) {
	t.Run("valid edge", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent2"})

		err := g.AddEdge("parent", "child")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		children := g.Edges["parent"]
		if len(children) != 1 || children[0] != "child" {
			t.Errorf("expected child to be added, got %v", children)
		}

		parents := g.ReverseEdges["child"]
		if len(parents) != 1 || parents[0] != "parent" {
			t.Errorf("expected parent to be in reverse edges, got %v", parents)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent2"})

		err := g.AddEdge("nonexistent", "child")
		if err == nil {
			t.Fatal("expected error for nonexistent parent")
		}
		if !errors.Is(err, ErrInvalidEdge) {
			t.Errorf("expected ErrInvalidEdge, got %v", err)
		}
	})

	t.Run("child not found", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})

		err := g.AddEdge("parent", "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent child")
		}
		if !errors.Is(err, ErrInvalidEdge) {
			t.Errorf("expected ErrInvalidEdge, got %v", err)
		}
	})

	t.Run("duplicate edge", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent2"})

		if err := g.AddEdge("parent", "child"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err := g.AddEdge("parent", "child")
		if err == nil {
			t.Fatal("expected error for duplicate edge")
		}
		if !errors.Is(err, ErrDuplicateEdge) {
			t.Errorf("expected ErrDuplicateEdge, got %v", err)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid DAG", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "c", AgentID: "agent3"})
		g.AddEdge("a", "b")
		g.AddEdge("a", "c")

		err := g.Validate()
		if err != nil {
			t.Fatalf("expected no error for valid DAG, got %v", err)
		}
	})

	t.Run("cycle detection", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "c", AgentID: "agent3"})
		g.AddEdge("a", "b")
		g.AddEdge("b", "c")
		g.AddEdge("c", "a") // cycle

		err := g.Validate()
		if err == nil {
			t.Fatal("expected error for cyclic graph")
		}
		if !errors.Is(err, ErrCycleDetected) {
			t.Errorf("expected ErrCycleDetected, got %v", err)
		}
	})

	t.Run("self-cycle", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		// Note: AddEdge would need a -> a which is a cycle, but we need both nodes to exist
		g.AddEdge("a", "a")

		err := g.Validate()
		if err == nil {
			t.Fatal("expected error for self-cycle")
		}
		if !errors.Is(err, ErrCycleDetected) {
			t.Errorf("expected ErrCycleDetected, got %v", err)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g := NewDelegationGraph()
		err := g.Validate()
		if err != nil {
			t.Fatalf("expected no error for empty graph, got %v", err)
		}
	})

	t.Run("disconnected nodes", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})

		err := g.Validate()
		if err != nil {
			t.Fatalf("expected no error for disconnected nodes, got %v", err)
		}
	})
}

func TestGetRoot(t *testing.T) {
	t.Run("valid root", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "root", AgentID: "agent1", Role: "architect"})
		g.RootAgent = "root"

		node, err := g.GetRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if node.ID != "root" {
			t.Errorf("expected root ID to be 'root', got %q", node.ID)
		}
		if node.Role != "architect" {
			t.Errorf("expected role to be 'architect', got %q", node.Role)
		}
	})

	t.Run("no root set", func(t *testing.T) {
		g := NewDelegationGraph()
		_, err := g.GetRoot()
		if err == nil {
			t.Fatal("expected error when no root is set")
		}
		if !errors.Is(err, ErrNoRootAgent) {
			t.Errorf("expected ErrNoRootAgent, got %v", err)
		}
	})

	t.Run("root node not found", func(t *testing.T) {
		g := NewDelegationGraph()
		g.RootAgent = "nonexistent"
		_, err := g.GetRoot()
		if err == nil {
			t.Fatal("expected error when root node doesn't exist")
		}
		if !errors.Is(err, ErrNodeNotFound) {
			t.Errorf("expected ErrNodeNotFound, got %v", err)
		}
	})
}

func TestGetChildren(t *testing.T) {
	t.Run("multiple children", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "child1", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "child2", AgentID: "agent3"})
		g.AddEdge("parent", "child1")
		g.AddEdge("parent", "child2")

		children := g.GetChildren("parent")
		if len(children) != 2 {
			t.Fatalf("expected 2 children, got %d", len(children))
		}
	})

	t.Run("no children", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "leaf", AgentID: "agent1"})

		children := g.GetChildren("leaf")
		if len(children) != 0 {
			t.Errorf("expected 0 children, got %d", len(children))
		}
	})

	t.Run("nonexistent node", func(t *testing.T) {
		g := NewDelegationGraph()
		children := g.GetChildren("nonexistent")
		if len(children) != 0 {
			t.Errorf("expected 0 children for nonexistent node, got %d", len(children))
		}
	})

	t.Run("returned slice is a copy", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent2"})
		g.AddEdge("parent", "child")

		children := g.GetChildren("parent")
		children[0] = "modified"

		original := g.GetChildren("parent")
		if original[0] == "modified" {
			t.Error("expected returned slice to be a copy, not the original")
		}
	})
}

func TestGetParents(t *testing.T) {
	t.Run("single parent", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent2"})
		g.AddEdge("parent", "child")

		parents := g.GetParents("child")
		if len(parents) != 1 || parents[0] != "parent" {
			t.Errorf("expected 1 parent 'parent', got %v", parents)
		}
	})

	t.Run("multiple parents", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "p1", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "p2", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent3"})
		g.AddEdge("p1", "child")
		g.AddEdge("p2", "child")

		parents := g.GetParents("child")
		if len(parents) != 2 {
			t.Fatalf("expected 2 parents, got %d", len(parents))
		}
	})

	t.Run("no parents (root node)", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "root", AgentID: "agent1"})

		parents := g.GetParents("root")
		if len(parents) != 0 {
			t.Errorf("expected 0 parents for root, got %d", len(parents))
		}
	})

	t.Run("returned slice is a copy", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "parent", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "child", AgentID: "agent2"})
		g.AddEdge("parent", "child")

		parents := g.GetParents("child")
		parents[0] = "modified"

		original := g.GetParents("child")
		if original[0] == "modified" {
			t.Error("expected returned slice to be a copy")
		}
	})
}

func TestGetExecutionOrder(t *testing.T) {
	t.Run("linear graph", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "c", AgentID: "agent3"})
		g.AddEdge("a", "b")
		g.AddEdge("b", "c")

		order, err := g.GetExecutionOrder()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(order) != 3 {
			t.Fatalf("expected 3 nodes in order, got %d", len(order))
		}
		// 'a' must come before 'b', 'b' before 'c'
		posA, posB, posC := -1, -1, -1
		for i, id := range order {
			switch id {
			case "a":
				posA = i
			case "b":
				posB = i
			case "c":
				posC = i
			}
		}
		if posA >= posB || posB >= posC {
			t.Errorf("expected order a -> b -> c, got %v", order)
		}
	})

	t.Run("diamond graph", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "top", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "left", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "right", AgentID: "agent3"})
		g.AddNode(&AgentNode{ID: "bottom", AgentID: "agent4"})
		g.AddEdge("top", "left")
		g.AddEdge("top", "right")
		g.AddEdge("left", "bottom")
		g.AddEdge("right", "bottom")

		order, err := g.GetExecutionOrder()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(order) != 4 {
			t.Fatalf("expected 4 nodes, got %d", len(order))
		}

		// Verify topological constraints
		pos := make(map[string]int)
		for i, id := range order {
			pos[id] = i
		}

		if pos["top"] >= pos["left"] || pos["top"] >= pos["right"] {
			t.Errorf("top should come before left and right, got %v", order)
		}
		if pos["left"] >= pos["bottom"] || pos["right"] >= pos["bottom"] {
			t.Errorf("left and right should come before bottom, got %v", order)
		}
	})

	t.Run("cyclic graph returns error", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})
		g.AddEdge("a", "b")
		g.AddEdge("b", "a")

		_, err := g.GetExecutionOrder()
		if err == nil {
			t.Fatal("expected error for cyclic graph")
		}
		if !errors.Is(err, ErrCycleDetected) {
			t.Errorf("expected ErrCycleDetected, got %v", err)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g := NewDelegationGraph()
		order, err := g.GetExecutionOrder()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(order) != 0 {
			t.Errorf("expected empty order, got %v", order)
		}
	})

	t.Run("disconnected nodes", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})

		order, err := g.GetExecutionOrder()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(order) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(order))
		}
	})
}

func TestDepth(t *testing.T) {
	t.Run("single node", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		if depth := g.Depth(); depth != 1 {
			t.Errorf("expected depth 1, got %d", depth)
		}
	})

	t.Run("linear chain", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "b", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "c", AgentID: "agent3"})
		g.AddEdge("a", "b")
		g.AddEdge("b", "c")

		if depth := g.Depth(); depth != 3 {
			t.Errorf("expected depth 3, got %d", depth)
		}
	})

	t.Run("diamond graph depth", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "top", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "left", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "right", AgentID: "agent3"})
		g.AddNode(&AgentNode{ID: "bottom", AgentID: "agent4"})
		g.AddEdge("top", "left")
		g.AddEdge("top", "right")
		g.AddEdge("left", "bottom")
		g.AddEdge("right", "bottom")

		if depth := g.Depth(); depth != 3 {
			t.Errorf("expected depth 3, got %d", depth)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g := NewDelegationGraph()
		if depth := g.Depth(); depth != 0 {
			t.Errorf("expected depth 0, got %d", depth)
		}
	})

	t.Run("tree structure", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "root", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "l1", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "r1", AgentID: "agent3"})
		g.AddNode(&AgentNode{ID: "l2", AgentID: "agent4"})
		g.AddNode(&AgentNode{ID: "r2", AgentID: "agent5"})
		g.AddEdge("root", "l1")
		g.AddEdge("root", "r1")
		g.AddEdge("l1", "l2")
		g.AddEdge("r1", "r2")

		if depth := g.Depth(); depth != 3 {
			t.Errorf("expected depth 3, got %d", depth)
		}
	})

	t.Run("disconnected components", func(t *testing.T) {
		g := NewDelegationGraph()
		g.AddNode(&AgentNode{ID: "a1", AgentID: "agent1"})
		g.AddNode(&AgentNode{ID: "a2", AgentID: "agent2"})
		g.AddNode(&AgentNode{ID: "b1", AgentID: "agent3"})
		g.AddNode(&AgentNode{ID: "b2", AgentID: "agent4"})
		g.AddNode(&AgentNode{ID: "b3", AgentID: "agent5"})
		g.AddEdge("a1", "a2")
		g.AddEdge("b1", "b2")
		g.AddEdge("b2", "b3")

		if depth := g.Depth(); depth != 3 {
			t.Errorf("expected depth 3 (from b component), got %d", depth)
		}
	})
}
