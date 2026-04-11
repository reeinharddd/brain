package agentpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrQueueFull       = errors.New("task queue is full")
	ErrQueueEmpty      = errors.New("task queue is empty")
	ErrMaxInstances    = errors.New("maximum instances reached")
	ErrMinInstances    = errors.New("cannot scale below minimum instances")
	ErrNoTask          = errors.New("no task with matching ID in queue")
	ErrAgentNotFound   = errors.New("agent not found")
	ErrTaskNotAssigned = errors.New("task is not assigned to this agent")
)

// PoolConfig holds configuration for the agent pool
type PoolConfig struct {
	MinInstances     int
	MaxInstances     int
	IdleTimeout      time.Duration
	QueueCapacity    int
	ScaleUpThreshold float64 // load percentage to trigger scale up
	ScaleDownTimeout time.Duration
}

// PoolStatus represents the current status of the pool
type PoolStatus struct {
	Role               AgentRole
	TotalInstances     int
	AvailableInstances int
	BusyInstances      int
	QueuedTasks        int
	TotalSubmitted     int
	TotalCompleted     int
	TotalFailed        int
	Load               float64
}

// AgentPool manageses a pool of agent instances
type AgentPool struct {
	mu           sync.RWMutex
	definition   AgentDefinition
	config       PoolConfig
	instances    map[string]*Agent // instanceID -> agent
	taskQueue    []*AgentTask
	totalTasksSubmitted int
	totalTasksCompleted int
	totalTasksFailed    int
	lastScaleUp    time.Time
	lastScaleDown  time.Time
	nextInstanceID int
}

// NewAgentPool creates a new agent pool with the given definition and config.
func NewAgentPool(def AgentDefinition, config PoolConfig) *AgentPool {
	return &AgentPool{
		definition: def,
		config:     config,
		instances:  make(map[string]*Agent),
		taskQueue:  make([]*AgentTask, 0),
	}
}

// SubmitTask adds a task to the pool's queue.
func (p *AgentPool) SubmitTask(ctx context.Context, task *AgentTask) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("submit task cancelled: %w", ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.taskQueue) >= p.config.QueueCapacity {
		return fmt.Errorf("cannot submit task: %w", ErrQueueFull)
	}

	p.taskQueue = append(p.taskQueue, task)
	p.totalTasksSubmitted++
	return nil
}

// GetTask retrieves the next task from the queue.
func (p *AgentPool) GetTask(ctx context.Context) (*AgentTask, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("get task cancelled: %w", ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.taskQueue) == 0 {
		return nil, fmt.Errorf("cannot get task: %w", ErrQueueEmpty)
	}

	task := p.taskQueue[0]
	p.taskQueue = p.taskQueue[1:]
	return task, nil
}

// CompleteTask marks a task as completed.
func (p *AgentPool) CompleteTask(ctx context.Context, taskID string, agentID string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("complete task cancelled: %w", ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	agent, ok := p.instances[agentID]
	if !ok {
		return fmt.Errorf("cannot complete task: %w for %s", ErrAgentNotFound, agentID)
	}

	agent.mu.Lock()
	defer agent.mu.Unlock()

	if agent.CurrentTask == nil || agent.CurrentTask.ID != taskID {
		return fmt.Errorf("cannot complete task: %w", ErrTaskNotAssigned)
	}

	agent.CurrentTask = nil
	agent.Status = StatusIdle
	agent.TasksCompleted++
	agent.LastActive = time.Now()
	p.totalTasksCompleted++
	return nil
}

// FailTask marks a task as failed.
func (p *AgentPool) FailTask(ctx context.Context, taskID string, agentID string, reason string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("fail task cancelled: %w", ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	agent, ok := p.instances[agentID]
	if !ok {
		return fmt.Errorf("cannot fail task: %w for %s", ErrAgentNotFound, agentID)
	}

	agent.mu.Lock()
	defer agent.mu.Unlock()

	if agent.CurrentTask == nil || agent.CurrentTask.ID != taskID {
		return fmt.Errorf("cannot fail task: %w", ErrTaskNotAssigned)
	}

	agent.CurrentTask = nil
	agent.Status = StatusIdle
	agent.TasksFailed++
	agent.LastActive = time.Now()
	p.totalTasksFailed++
	return nil
}

// ScaleUp creates new agent instances.
func (p *AgentPool) ScaleUp(ctx context.Context, count int) ([]*Agent, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("scale up cancelled: %w", ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	currentCount := len(p.instances)
	available := p.config.MaxInstances - currentCount
	if available <= 0 {
		return nil, fmt.Errorf("cannot scale up: %w", ErrMaxInstances)
	}

	if count > available {
		count = available
	}

	agents := make([]*Agent, 0, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		p.nextInstanceID++
		id := fmt.Sprintf("%s-%d", p.definition.Role, p.nextInstanceID)
		agent := &Agent{
			ID:           id,
			Role:         p.definition.Role,
			Name:         fmt.Sprintf("%s-%d", p.definition.Name, p.nextInstanceID),
			Status:       StatusIdle,
			Capabilities: p.definition.Capabilities,
			CreatedAt:    now,
			LastActive:   now,
			Metadata:     make(map[string]string),
		}
		p.instances[id] = agent
		agents = append(agents, agent)
	}

	p.lastScaleUp = now
	return agents, nil
}

// ScaleDown removes idle agent instances.
func (p *AgentPool) ScaleDown(ctx context.Context, count int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("scale down cancelled: %w", ctx.Err())
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	currentCount := len(p.instances)
	if currentCount-count < p.config.MinInstances {
		return fmt.Errorf("cannot scale down: would go below %w (min=%d)", ErrMinInstances, p.config.MinInstances)
	}

	// Find idle agents to remove
	removed := 0
	for id, agent := range p.instances {
		if removed >= count {
			break
		}
		agent.mu.RLock()
		isIdle := agent.Status == StatusIdle
		agent.mu.RUnlock()

		if isIdle {
			delete(p.instances, id)
			removed++
		}
	}

	if removed < count {
		return fmt.Errorf("could only scale down by %d of %d requested: not enough idle instances", removed, count)
	}

	p.lastScaleDown = time.Now()
	return nil
}

// GetStatus returns the current status of the pool.
func (p *AgentPool) GetStatus(ctx context.Context) PoolStatus {
	select {
	case <-ctx.Done():
		return PoolStatus{}
	default:
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	total := len(p.instances)
	busy := 0
	for _, agent := range p.instances {
		agent.mu.RLock()
		if agent.Status == StatusBusy {
			busy++
		}
		agent.mu.RUnlock()
	}

	return PoolStatus{
		Role:               p.definition.Role,
		TotalInstances:     total,
		AvailableInstances: total - busy,
		BusyInstances:      busy,
		QueuedTasks:        len(p.taskQueue),
		TotalSubmitted:     p.totalTasksSubmitted,
		TotalCompleted:     p.totalTasksCompleted,
		TotalFailed:        p.totalTasksFailed,
		Load:               p.getLoadLocked(),
	}
}

// GetInstances returns a copy of the instance list.
func (p *AgentPool) GetInstances(ctx context.Context) []*Agent {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Agent, 0, len(p.instances))
	for _, agent := range p.instances {
		result = append(result, agent)
	}
	return result
}

// GetAvailableInstance returns an idle agent instance.
func (p *AgentPool) GetAvailableInstance(ctx context.Context) (*Agent, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("get available instance cancelled: %w", ctx.Err())
	default:
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, agent := range p.instances {
		if agent.IsAvailable() {
			return agent, nil
		}
	}

	return nil, errors.New("no available instance")
}

// GetLoad returns the current load as a percentage (0.0 to 1.0+).
func (p *AgentPool) GetLoad() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getLoadLocked()
}

// getLoadLocked calculates load without acquiring the lock (caller must hold lock).
func (p *AgentPool) getLoadLocked() float64 {
	total := len(p.instances)
	if total == 0 {
		if len(p.taskQueue) > 0 {
			return 1.0 // overloaded if there are tasks but no agents
		}
		return 0.0
	}

	busy := 0
	for _, agent := range p.instances {
		agent.mu.RLock()
		if agent.Status == StatusBusy {
			busy++
		}
		agent.mu.RUnlock()
	}

	busyRatio := float64(busy) / float64(total)
	queueRatio := 0.0
	if p.config.QueueCapacity > 0 {
		queueRatio = float64(len(p.taskQueue)) / float64(p.config.QueueCapacity)
	}

	// Load is a weighted combination of busy agents and queued tasks
	load := busyRatio*0.7 + queueRatio*0.3
	return load
}
