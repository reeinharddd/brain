package manager

import (
	"bufio"
	"fmt"

	"net"
	"os/exec"
	"sync"
	"time"
)

type ProcessState string

const (
	StateStarting ProcessState = "Starting"
	StateRunning  ProcessState = "Running"
	StateStopped  ProcessState = "Stopped"
	StateFailed   ProcessState = "Failed"
)

type Process struct {
	ID        string
	Command   string
	Args      []string
	State     ProcessState
	Port      int
	cmd       *exec.Cmd
	StartTime time.Time
}

type ProcessManager struct {
	mu        sync.Mutex
	processes map[string]*Process
	logger    chan<- string
}

func NewProcessManager(logger chan<- string) *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
		logger:    logger,
	}
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (m *ProcessManager) StartProcess(id, command string, args []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, exists := m.processes[id]; exists && (p.State == StateRunning || p.State == StateStarting) {
		return fmt.Errorf("process %s is already running", id)
	}

	p := &Process{
		ID:        id,
		Command:   command,
		Args:      args,
		State:     StateStarting,
		StartTime: time.Now(),
	}
	m.processes[id] = p

	go m.runProcess(p)
	return nil
}

func (m *ProcessManager) StopProcess(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.processes[id]
	if !exists {
		return fmt.Errorf("process %s not found", id)
	}

	if p.cmd != nil && p.cmd.Process != nil {
		err := p.cmd.Process.Kill()
		p.State = StateStopped
		m.logger <- fmt.Sprintf("[ProcessManager] Stopped %s", id)
		return err
	}
	return fmt.Errorf("process %s is not running", id)
}

func (m *ProcessManager) GetStatus(id string) (ProcessState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.processes[id]
	if !exists {
		return StateStopped, fmt.Errorf("process %s not found", id)
	}
	return p.State, nil
}

func (m *ProcessManager) GetAllStatuses() map[string]ProcessState {
	m.mu.Lock()
	defer m.mu.Unlock()

	statuses := make(map[string]ProcessState)
	for id, p := range m.processes {
		statuses[id] = p.State
	}
	return statuses
}

func (m *ProcessManager) runProcess(p *Process) {
	cmd := exec.Command(p.Command, p.Args...)
	p.cmd = cmd

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		p.State = StateFailed
		m.logger <- fmt.Sprintf("[ProcessManager] Failed to start %s: %v", p.ID, err)
		return
	}

	p.State = StateRunning
	m.logger <- fmt.Sprintf("[ProcessManager] Started %s (PID: %d)", p.ID, cmd.Process.Pid)

	// Stream stdout & stderr
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m.logger <- fmt.Sprintf("[%s] %s", p.ID, scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m.logger <- fmt.Sprintf("[%s|ERR] %s", p.ID, scanner.Text())
		}
	}()

	err := cmd.Wait()
	if err != nil {
		p.State = StateFailed
		m.logger <- fmt.Sprintf("[ProcessManager] %s exited with error: %v", p.ID, err)
	} else {
		p.State = StateStopped
		m.logger <- fmt.Sprintf("[ProcessManager] %s exited cleanly", p.ID)
	}
}
