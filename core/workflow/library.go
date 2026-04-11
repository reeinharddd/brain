package workflow

import (
	"context"
	"fmt"
)

// WorkflowLibrary provides pre-built workflow templates
type WorkflowLibrary struct{}

// FeatureDev creates a feature development workflow
// architect -> (builder + tester) -> reviewer -> documenter
func (wl *WorkflowLibrary) FeatureDev() *WorkflowDAG {
	dag := NewWorkflowDAG("feature-dev", "Feature Development", true)

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "architect",
		Name:  "Architect",
		Agent: "architect",
		Input: map[string]string{
			"task": "design and plan",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "builder",
		Name:      "Builder",
		Agent:     "builder",
		DependsOn: []string{"architect"},
		Input: map[string]string{
			"task": "implement feature",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "tester",
		Name:      "Tester",
		Agent:     "tester",
		DependsOn: []string{"architect"},
		Input: map[string]string{
			"task": "write tests",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "reviewer",
		Name:      "Reviewer",
		Agent:     "reviewer",
		DependsOn: []string{"builder", "tester"},
		Input: map[string]string{
			"task": "code review",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "documenter",
		Name:      "Documenter",
		Agent:     "documenter",
		DependsOn: []string{"reviewer"},
		Input: map[string]string{
			"task": "update docs",
		},
	})

	return dag
}

// BugFix creates a bug fix workflow
// debugger -> builder -> tester -> reviewer
func (wl *WorkflowLibrary) BugFix() *WorkflowDAG {
	dag := NewWorkflowDAG("bug-fix", "Bug Fix", false)

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "debugger",
		Name:  "Debugger",
		Agent: "debugger",
		Input: map[string]string{
			"task": "investigate and identify root cause",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "builder",
		Name:      "Builder",
		Agent:     "builder",
		DependsOn: []string{"debugger"},
		Input: map[string]string{
			"task": "implement fix",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "tester",
		Name:      "Tester",
		Agent:     "tester",
		DependsOn: []string{"builder"},
		Input: map[string]string{
			"task": "verify fix with tests",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "reviewer",
		Name:      "Reviewer",
		Agent:     "reviewer",
		DependsOn: []string{"tester"},
		Input: map[string]string{
			"task": "review the fix",
		},
	})

	return dag
}

// Refactor creates a code refactoring workflow
// refactorer -> (builder + tester) -> reviewer
func (wl *WorkflowLibrary) Refactor() *WorkflowDAG {
	dag := NewWorkflowDAG("refactor", "Refactor", true)

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "refactorer",
		Name:  "Refactorer",
		Agent: "refactorer",
		Input: map[string]string{
			"task": "identify refactoring opportunities",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "builder",
		Name:      "Builder",
		Agent:     "builder",
		DependsOn: []string{"refactorer"},
		Input: map[string]string{
			"task": "implement refactoring",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "tester",
		Name:      "Tester",
		Agent:     "tester",
		DependsOn: []string{"refactorer"},
		Input: map[string]string{
			"task": "verify behavior unchanged",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "reviewer",
		Name:      "Reviewer",
		Agent:     "reviewer",
		DependsOn: []string{"builder", "tester"},
		Input: map[string]string{
			"task": "review refactored code",
		},
	})

	return dag
}

// CodeReview creates a code review workflow
// (reviewer + security-auditor) - parallel
func (wl *WorkflowLibrary) CodeReview() *WorkflowDAG {
	dag := NewWorkflowDAG("code-review", "Code Review", true)

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "reviewer",
		Name:  "Reviewer",
		Agent: "reviewer",
		Input: map[string]string{
			"task": "code quality review",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "security-auditor",
		Name:  "Security Auditor",
		Agent: "security-auditor",
		Input: map[string]string{
			"task": "security review",
		},
	})

	return dag
}

// Migration creates a migration workflow
// migrator -> tester -> documenter
func (wl *WorkflowLibrary) Migration() *WorkflowDAG {
	dag := NewWorkflowDAG("migration", "Migration", false)

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "migrator",
		Name:  "Migrator",
		Agent: "migrator",
		Input: map[string]string{
			"task": "perform migration",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "tester",
		Name:      "Tester",
		Agent:     "tester",
		DependsOn: []string{"migrator"},
		Input: map[string]string{
			"task": "verify migrated code",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "documenter",
		Name:      "Documenter",
		Agent:     "documenter",
		DependsOn: []string{"tester"},
		Input: map[string]string{
			"task": "update migration docs",
		},
	})

	return dag
}

// FullRelease creates a full release workflow
// Simplified: architect -> builder -> tester -> reviewer -> documenter -> migrator
func (wl *WorkflowLibrary) FullRelease() *WorkflowDAG {
	dag := NewWorkflowDAG("full-release", "Full Release", false)

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:    "architect",
		Name:  "Architect",
		Agent: "architect",
		Input: map[string]string{
			"task": "design and plan",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "builder",
		Name:      "Builder",
		Agent:     "builder",
		DependsOn: []string{"architect"},
		Input: map[string]string{
			"task": "implement features",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "tester",
		Name:      "Tester",
		Agent:     "tester",
		DependsOn: []string{"builder"},
		Input: map[string]string{
			"task": "test implementation",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "reviewer",
		Name:      "Reviewer",
		Agent:     "reviewer",
		DependsOn: []string{"tester"},
		Input: map[string]string{
			"task": "review code",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "documenter",
		Name:      "Documenter",
		Agent:     "documenter",
		DependsOn: []string{"reviewer"},
		Input: map[string]string{
			"task": "update documentation",
		},
	})

	_ = dag.AddNode(context.Background(), &WorkflowNode{
		ID:        "migrator",
		Name:      "Migrator",
		Agent:     "migrator",
		DependsOn: []string{"documenter"},
		Input: map[string]string{
			"task": "perform migration",
		},
	})

	return dag
}

// GetWorkflow returns a workflow by name
func (wl *WorkflowLibrary) GetWorkflow(name string) (*WorkflowDAG, error) {
	switch name {
	case "feature-dev":
		return wl.FeatureDev(), nil
	case "bug-fix":
		return wl.BugFix(), nil
	case "refactor":
		return wl.Refactor(), nil
	case "code-review":
		return wl.CodeReview(), nil
	case "migration":
		return wl.Migration(), nil
	case "full-release":
		return wl.FullRelease(), nil
	default:
		return nil, fmt.Errorf("unknown workflow: %q", name)
	}
}

// ListWorkflows returns the names of all available workflows
func (wl *WorkflowLibrary) ListWorkflows() []string {
	return []string{
		"feature-dev",
		"bug-fix",
		"refactor",
		"code-review",
		"migration",
		"full-release",
	}
}
