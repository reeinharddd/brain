package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// AgentExecutor implements NodeExecutor using real subprocesses
type AgentExecutor struct {
	agentPool map[string]*AgentConfig
	workDir   string
	logFunc   func(string)
}

// AgentConfig defines how to invoke an agent
type AgentConfig struct {
	Role        string
	Command     string
	Args        []string
	Env         map[string]string
	Description string
}

// NewAgentExecutor creates a new executor with default agent registrations
func NewAgentExecutor(workDir string, logFunc func(string)) *AgentExecutor {
	e := &AgentExecutor{
		workDir:   workDir,
		logFunc:   logFunc,
		agentPool: make(map[string]*AgentConfig),
	}
	e.registerDefaultAgents()
	return e
}

func (e *AgentExecutor) registerDefaultAgents() {
	roles := []string{
		"architect", "builder", "reviewer", "debugger",
		"refactorer", "tester", "documenter", "migrator", "security-auditor",
	}
	for _, role := range roles {
		e.agentPool[role] = &AgentConfig{
			Role:        role,
			Command:     "echo",
			Args:        []string{fmt.Sprintf(`[agent:%s] Task executed successfully`, role)},
			Description: fmt.Sprintf("%s agent", role),
			Env:         map[string]string{"BRAIN_AGENT_ROLE": role},
		}
	}
}

// RegisterAgent registers a custom agent
func (e *AgentExecutor) RegisterAgent(role string, cfg *AgentConfig) {
	e.agentPool[role] = cfg
}

// Execute runs a single workflow node
func (e *AgentExecutor) Execute(ctx context.Context, node *WorkflowNode) (*TaskOutput, error) {
	agent, ok := e.agentPool[node.Agent]
	if !ok {
		return nil, fmt.Errorf("agent %q not found in pool", node.Agent)
	}

	e.logFunc(fmt.Sprintf("[executor] Running %s agent for task: %s", node.Agent, node.Name))

	start := time.Now()

	// Build command
	cmd := exec.CommandContext(ctx, agent.Command, agent.Args...)

	// Set environment
	cmd.Env = make([]string, 0)
	for k, v := range agent.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Pass task input as stdin JSON
	if len(node.Input) > 0 {
		inputData, _ := json.Marshal(node.Input)
		cmd.Stdin = bytes.NewReader(inputData)
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	output := &TaskOutput{
		Duration: duration,
	}

	if stdout.Len() > 0 {
		output.Result = []byte(stdout.String())
		// Rough token estimation (~4 chars per token)
		output.TokensUsed = len(output.Result) / 4
		output.CostUSD = float64(output.TokensUsed) / 1000000.0 * 10.0
	}

	if err != nil {
		errMsg := fmt.Sprintf("agent failed: %v", err)
		if stderr.Len() > 0 {
			errMsg += "\nstderr: " + stderr.String()
		}
		output.Error = errMsg
		return output, err
	}

	return output, nil
}
