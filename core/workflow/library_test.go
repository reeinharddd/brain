package workflow

import (
	"context"
	"testing"
)

func TestWorkflowLibrary_ReturnsValidDAG(t *testing.T) {
	tests := []struct {
		name       string
		workflowFn func(*WorkflowLibrary) *WorkflowDAG
	}{
		{
			name:       "FeatureDev",
			workflowFn: (*WorkflowLibrary).FeatureDev,
		},
		{
			name:       "BugFix",
			workflowFn: (*WorkflowLibrary).BugFix,
		},
		{
			name:       "Refactor",
			workflowFn: (*WorkflowLibrary).Refactor,
		},
		{
			name:       "CodeReview",
			workflowFn: (*WorkflowLibrary).CodeReview,
		},
		{
			name:       "Migration",
			workflowFn: (*WorkflowLibrary).Migration,
		},
		{
			name:       "FullRelease",
			workflowFn: (*WorkflowLibrary).FullRelease,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &WorkflowLibrary{}
			dag := tt.workflowFn(lib)

			if dag == nil {
				t.Fatalf("expected non-nil DAG")
			}

			if dag.ID == "" {
				t.Error("expected non-empty ID")
			}

			if dag.Name == "" {
				t.Error("expected non-empty Name")
			}

			if dag.Nodes == nil {
				t.Error("expected non-nil Nodes map")
			}

			// Validate the DAG
			ctx := context.Background()
			if err := dag.Validate(ctx); err != nil {
				t.Errorf("DAG validation failed: %v", err)
			}
		})
	}
}

func TestWorkflowLibrary_ValidatePassesOnAllWorkflows(t *testing.T) {
	ctx := context.Background()
	lib := &WorkflowLibrary{}

	workflows := []struct {
		name string
		dag  *WorkflowDAG
	}{
		{"feature-dev", lib.FeatureDev()},
		{"bug-fix", lib.BugFix()},
		{"refactor", lib.Refactor()},
		{"code-review", lib.CodeReview()},
		{"migration", lib.Migration()},
		{"full-release", lib.FullRelease()},
	}

	for _, wf := range workflows {
		t.Run(wf.name, func(t *testing.T) {
			if err := wf.dag.Validate(ctx); err != nil {
				t.Errorf("workflow %q validation failed: %v", wf.name, err)
			}
		})
	}
}

func TestWorkflowLibrary_GetWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "known workflow: feature-dev",
			input:     "feature-dev",
			expectErr: false,
		},
		{
			name:      "known workflow: bug-fix",
			input:     "bug-fix",
			expectErr: false,
		},
		{
			name:      "known workflow: refactor",
			input:     "refactor",
			expectErr: false,
		},
		{
			name:      "known workflow: code-review",
			input:     "code-review",
			expectErr: false,
		},
		{
			name:      "known workflow: migration",
			input:     "migration",
			expectErr: false,
		},
		{
			name:      "known workflow: full-release",
			input:     "full-release",
			expectErr: false,
		},
		{
			name:      "unknown workflow",
			input:     "nonexistent",
			expectErr: true,
		},
		{
			name:      "empty workflow name",
			input:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &WorkflowLibrary{}
			dag, err := lib.GetWorkflow(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if dag != nil {
					t.Error("expected nil DAG on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if dag == nil {
				t.Error("expected non-nil DAG")
			}
		})
	}
}

func TestWorkflowLibrary_ListWorkflows(t *testing.T) {
	t.Run("returns 6 items", func(t *testing.T) {
		lib := &WorkflowLibrary{}
		workflows := lib.ListWorkflows()

		if len(workflows) != 6 {
			t.Errorf("expected 6 workflows, got %d", len(workflows))
		}
	})

	t.Run("contains all expected workflows", func(t *testing.T) {
		lib := &WorkflowLibrary{}
		workflows := lib.ListWorkflows()

		expected := map[string]bool{
			"feature-dev":  false,
			"bug-fix":      false,
			"refactor":     false,
			"code-review":  false,
			"migration":    false,
			"full-release": false,
		}

		for _, wf := range workflows {
			if _, ok := expected[wf]; ok {
				expected[wf] = true
			}
		}

		for wf, found := range expected {
			if !found {
				t.Errorf("expected workflow %q not found", wf)
			}
		}
	})
}

func TestWorkflowLibrary_NodeCount(t *testing.T) {
	tests := []struct {
		name       string
		workflowFn func(*WorkflowLibrary) *WorkflowDAG
		expectNodes int
	}{
		{
			name:        "FeatureDev has 5 nodes",
			workflowFn:  (*WorkflowLibrary).FeatureDev,
			expectNodes: 5,
		},
		{
			name:        "BugFix has 4 nodes",
			workflowFn:  (*WorkflowLibrary).BugFix,
			expectNodes: 4,
		},
		{
			name:        "Refactor has 4 nodes",
			workflowFn:  (*WorkflowLibrary).Refactor,
			expectNodes: 4,
		},
		{
			name:        "CodeReview has 2 nodes",
			workflowFn:  (*WorkflowLibrary).CodeReview,
			expectNodes: 2,
		},
		{
			name:        "Migration has 3 nodes",
			workflowFn:  (*WorkflowLibrary).Migration,
			expectNodes: 3,
		},
		{
			name:        "FullRelease has 6 nodes",
			workflowFn:  (*WorkflowLibrary).FullRelease,
			expectNodes: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &WorkflowLibrary{}
			dag := tt.workflowFn(lib)

			if len(dag.Nodes) != tt.expectNodes {
				t.Errorf("expected %d nodes, got %d", tt.expectNodes, len(dag.Nodes))
			}
		})
	}
}

func TestWorkflowLibrary_ParallelFlag(t *testing.T) {
	tests := []struct {
		name         string
		workflowFn   func(*WorkflowLibrary) *WorkflowDAG
		expectParallel bool
	}{
		{
			name:         "FeatureDev is parallel",
			workflowFn:   (*WorkflowLibrary).FeatureDev,
			expectParallel: true,
		},
		{
			name:         "BugFix is not parallel",
			workflowFn:   (*WorkflowLibrary).BugFix,
			expectParallel: false,
		},
		{
			name:         "Refactor is parallel",
			workflowFn:   (*WorkflowLibrary).Refactor,
			expectParallel: true,
		},
		{
			name:         "CodeReview is parallel",
			workflowFn:   (*WorkflowLibrary).CodeReview,
			expectParallel: true,
		},
		{
			name:         "Migration is not parallel",
			workflowFn:   (*WorkflowLibrary).Migration,
			expectParallel: false,
		},
		{
			name:         "FullRelease is not parallel",
			workflowFn:   (*WorkflowLibrary).FullRelease,
			expectParallel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &WorkflowLibrary{}
			dag := tt.workflowFn(lib)

			if dag.Parallel != tt.expectParallel {
				t.Errorf("expected Parallel=%v, got %v", tt.expectParallel, dag.Parallel)
			}
		})
	}
}

func TestWorkflowLibrary_WorkflowStructure(t *testing.T) {
	t.Run("FeatureDev has correct structure", func(t *testing.T) {
		lib := &WorkflowLibrary{}
		dag := lib.FeatureDev()
		ctx := context.Background()

		// Check architect is root node
		architect, err := dag.GetNode(ctx, "architect")
		if err != nil {
			t.Fatalf("expected architect node: %v", err)
		}
		if len(architect.DependsOn) != 0 {
			t.Errorf("expected architect to have no dependencies, got %v", architect.DependsOn)
		}

		// Check builder and tester depend on architect
		builder, err := dag.GetNode(ctx, "builder")
		if err != nil {
			t.Fatalf("expected builder node: %v", err)
		}
		if len(builder.DependsOn) != 1 || builder.DependsOn[0] != "architect" {
			t.Errorf("expected builder to depend on architect, got %v", builder.DependsOn)
		}

		tester, err := dag.GetNode(ctx, "tester")
		if err != nil {
			t.Fatalf("expected tester node: %v", err)
		}
		if len(tester.DependsOn) != 1 || tester.DependsOn[0] != "architect" {
			t.Errorf("expected tester to depend on architect, got %v", tester.DependsOn)
		}

		// Check reviewer depends on builder and tester
		reviewer, err := dag.GetNode(ctx, "reviewer")
		if err != nil {
			t.Fatalf("expected reviewer node: %v", err)
		}
		if len(reviewer.DependsOn) != 2 {
			t.Errorf("expected reviewer to have 2 dependencies, got %d", len(reviewer.DependsOn))
		}

		// Check documenter depends on reviewer
		documenter, err := dag.GetNode(ctx, "documenter")
		if err != nil {
			t.Fatalf("expected documenter node: %v", err)
		}
		if len(documenter.DependsOn) != 1 || documenter.DependsOn[0] != "reviewer" {
			t.Errorf("expected documenter to depend on reviewer, got %v", documenter.DependsOn)
		}
	})

	t.Run("CodeReview has parallel nodes", func(t *testing.T) {
		lib := &WorkflowLibrary{}
		dag := lib.CodeReview()
		ctx := context.Background()

		// Both nodes should have no dependencies
		reviewer, err := dag.GetNode(ctx, "reviewer")
		if err != nil {
			t.Fatalf("expected reviewer node: %v", err)
		}
		if len(reviewer.DependsOn) != 0 {
			t.Errorf("expected reviewer to have no dependencies, got %v", reviewer.DependsOn)
		}

		securityAuditor, err := dag.GetNode(ctx, "security-auditor")
		if err != nil {
			t.Fatalf("expected security-auditor node: %v", err)
		}
		if len(securityAuditor.DependsOn) != 0 {
			t.Errorf("expected security-auditor to have no dependencies, got %v", securityAuditor.DependsOn)
		}
	})

	t.Run("FullRelease has linear chain", func(t *testing.T) {
		lib := &WorkflowLibrary{}
		dag := lib.FullRelease()
		ctx := context.Background()

		expectedChain := []string{"architect", "builder", "tester", "reviewer", "documenter", "migrator"}

		for i, nodeID := range expectedChain {
			node, err := dag.GetNode(ctx, nodeID)
			if err != nil {
				t.Fatalf("expected %s node: %v", nodeID, err)
			}

			if i == 0 {
				if len(node.DependsOn) != 0 {
					t.Errorf("expected %s to have no dependencies, got %v", nodeID, node.DependsOn)
				}
			} else {
				expectedDep := expectedChain[i-1]
				if len(node.DependsOn) != 1 || node.DependsOn[0] != expectedDep {
					t.Errorf("expected %s to depend on %s, got %v", nodeID, expectedDep, node.DependsOn)
				}
			}
		}
	})
}
