package agentpool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentRole defines the agent's specialization
type AgentRole string

const (
	RoleArchitect       AgentRole = "architect"
	RoleBuilder         AgentRole = "builder"
	RoleReviewer        AgentRole = "reviewer"
	RoleDebugger        AgentRole = "debugger"
	RoleRefactorer      AgentRole = "refactorer"
	RoleTester          AgentRole = "tester"
	RoleDocumenter      AgentRole = "documenter"
	RoleMigrator        AgentRole = "migrator"
	RoleSecurityAuditor AgentRole = "security-auditor"
)

// AgentCapability defines what an agent can do
type AgentCapability string

const (
	CapSystemDesign      AgentCapability = "system_design"
	CapArchitectureReview AgentCapability = "architecture_review"
	CapCodeGeneration    AgentCapability = "code_generation"
	CapCodeReview        AgentCapability = "code_review"
	CapDebugging         AgentCapability = "debugging"
	CapRefactoring       AgentCapability = "refactoring"
	CapTestGeneration    AgentCapability = "test_generation"
	CapDocumentation     AgentCapability = "documentation"
	CapMigration         AgentCapability = "migration"
	CapSecurityAudit     AgentCapability = "security_audit"
)

// AgentStatus defines agent operational state
type AgentStatus string

const (
	StatusIdle    AgentStatus = "idle"
	StatusBusy    AgentStatus = "busy"
	StatusError   AgentStatus = "error"
	StatusScaling AgentStatus = "scaling"
)

// AgentTask represents a task assigned to an agent
type AgentTask struct {
	ID          string
	Description string
	Input       map[string]string
	Priority    int
	CreatedAt   time.Time
	Deadline    time.Time
}

// Agent represents a single agent instance
type Agent struct {
	mu               sync.RWMutex
	ID               string
	Role             AgentRole
	Name             string
	Status           AgentStatus
	Capabilities     []AgentCapability
	CurrentTask      *AgentTask
	TasksCompleted   int
	TasksFailed      int
	CreatedAt        time.Time
	LastActive       time.Time
	Metadata         map[string]string
}

// AssignTask assigns a task to the agent.
func (a *Agent) AssignTask(ctx context.Context, task *AgentTask) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	select {
	case <-ctx.Done():
		return fmt.Errorf("assign task cancelled: %w", ctx.Err())
	default:
	}

	if a.Status != StatusIdle {
		return fmt.Errorf("agent %s is not idle, current status: %s", a.ID, a.Status)
	}

	a.CurrentTask = task
	a.Status = StatusBusy
	a.LastActive = time.Now()
	return nil
}

// ReleaseTask marks the current task as completed.
func (a *Agent) ReleaseTask(ctx context.Context, success bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	select {
	case <-ctx.Done():
		return fmt.Errorf("release task cancelled: %w", ctx.Err())
	default:
	}

	if a.CurrentTask == nil {
		return fmt.Errorf("agent %s has no current task", a.ID)
	}

	if success {
		a.TasksCompleted++
	} else {
		a.TasksFailed++
	}

	a.CurrentTask = nil
	a.Status = StatusIdle
	a.LastActive = time.Now()
	return nil
}

// IsAvailable returns true if the agent is idle.
func (a *Agent) IsAvailable() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status == StatusIdle
}

// GetStatus returns the agent's current status.
func (a *Agent) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}

// AgentDefinition defines an agent type that can be instantiated
type AgentDefinition struct {
	Role               AgentRole
	Name               string
	Capabilities       []AgentCapability
	MaxConcurrentTasks int
	Constraints        map[string]string
	CostBudgetPerTask  float64
	CostBudgetDaily    float64
}
