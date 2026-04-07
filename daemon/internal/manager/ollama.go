package manager

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type OllamaManager struct {
	mu       sync.Mutex
	endpoint string
	enabled  bool
	logCh    chan string
}

func NewOllamaManager(endpoint string, logCh chan string) *OllamaManager {
	return &OllamaManager{
		endpoint: endpoint,
		logCh:    logCh,
		enabled:  false,
	}
}

func (m *OllamaManager) IsAvailable(ctx context.Context) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(m.endpoint + "/api/tags")
	if err != nil {
		m.log("not available")
		return false
	}
	defer resp.Body.Close()

	ok := resp.StatusCode == http.StatusOK
	m.mu.Lock()
	m.enabled = ok
	m.mu.Unlock()

	if ok {
		m.log("available")
	}
	return ok
}

func (m *OllamaManager) HealthCheck(ctx context.Context) bool {
	return m.IsAvailable(ctx)
}

func (m *OllamaManager) log(msg string) {
	select {
	case m.logCh <- "[Ollama] " + msg:
	default:
		fmt.Println("[Ollama]", msg)
	}
}
