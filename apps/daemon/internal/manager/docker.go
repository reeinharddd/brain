package manager

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DockerManager orchestrates docker-compose lifecycle
type DockerManager struct {
	composeDir string
	logChannel chan string
}

// ServiceStatus represents status of a docker service
type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Healthy bool   `json:"healthy"`
	Port    string `json:"port,omitempty"`
}

// NewDockerManager creates a new docker manager
func NewDockerManager(brainRoot string, logCh chan string) *DockerManager {
	return &DockerManager{
		composeDir: brainRoot,
		logChannel: logCh,
	}
}

// ComposeUp starts docker-compose services
func (m *DockerManager) ComposeUp(ctx context.Context) error {
	m.log("Starting docker compose services...")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", filepath.Join(m.composeDir, "docker-compose.yml"), "up", "-d")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("docker compose up failed: %v - %s", err, out.String())
		m.log(msg)
		return fmt.Errorf("%s", msg)
	}
	
	m.log("Docker services started: " + out.String())
	return nil
}

// ComposeDown stops docker-compose services
func (m *DockerManager) ComposeDown(ctx context.Context) error {
	m.log("Stopping docker compose services...")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", filepath.Join(m.composeDir, "docker-compose.yml"), "down")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	if err := cmd.Run(); err != nil {
		m.log(fmt.Sprintf("docker compose down failed: %v", err))
		return fmt.Errorf("docker compose down failed: %w", err)
	}
	
	m.log("Docker services stopped")
	return nil
}

// ComposeRestart restarts docker-compose services
func (m *DockerManager) ComposeRestart(ctx context.Context) error {
	m.log("Restarting docker compose services...")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", filepath.Join(m.composeDir, "docker-compose.yml"), "restart")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	if err := cmd.Run(); err != nil {
		m.log(fmt.Sprintf("docker compose restart failed: %v", err))
		return fmt.Errorf("docker compose restart failed: %w", err)
	}
	
	m.log("Docker services restarted")
	return nil
}

// GetServiceStatus returns status of a specific service
func (m *DockerManager) GetServiceStatus(ctx context.Context, serviceName string) (ServiceStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", "name=brain-"+serviceName, "--format", "{{.Names}}|{{.Status}}")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return ServiceStatus{}, fmt.Errorf("failed to get service status: %w", err)
	}
	
	parts := strings.Split(strings.TrimSpace(out.String()), "|")
	if len(parts) < 2 {
		return ServiceStatus{Name: serviceName, Status: "stopped", Healthy: false}, nil
	}
	
	return ServiceStatus{
		Name:    parts[0],
		Status:  parts[1],
		Healthy: strings.Contains(parts[1], "healthy") || strings.Contains(parts[1], "Up"),
	}, nil
}

// HealthCheck verifies docker is running and services are healthy
func (m *DockerManager) HealthCheck(ctx context.Context) error {
	// Check if docker daemon is running
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		msg := "Docker daemon is not running"
		m.log(msg)
		return fmt.Errorf("%s", msg)
	}

	// Check qdrant health
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:6333/health")
	if err != nil {
		msg := fmt.Sprintf("Qdrant health check failed: %v", err)
		m.log(msg)
		return fmt.Errorf("%s", msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Qdrant returned status %d", resp.StatusCode)
		m.log(msg)
		return fmt.Errorf("%s", msg)
	}

	m.log("Docker health check passed")
	return nil
}

// WaitForService waits for a service to be healthy
func (m *DockerManager) WaitForService(ctx context.Context, serviceName string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		status, err := m.GetServiceStatus(ctx, serviceName)
		if err == nil && status.Healthy {
			m.log(fmt.Sprintf("Service %s is healthy", serviceName))
			return nil
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	
	return fmt.Errorf("service %s did not become healthy after %d retries", serviceName, maxRetries)
}

func (m *DockerManager) log(msg string) {
	select {
	case m.logChannel <- "[DockerManager] " + msg:
	default:
		fmt.Println("[DockerManager]", msg)
	}
}
