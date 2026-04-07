package manager

import (
	"context"
	"fmt"
	"os"
	"sync"
)

type ProviderManager struct {
	mu        sync.RWMutex
	providers map[string]string
	logCh     chan string
}

func NewProviderManager(logCh chan string) *ProviderManager {
	return &ProviderManager{
		providers: make(map[string]string),
		logCh:     logCh,
	}
}

func (p *ProviderManager) SelectProvider(ctx context.Context, taskType string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.log("selecting for task: " + taskType)

	switch taskType {
	case "exploration":
		return "openai", nil
	case "implementation":
		return "anthropic", nil
	case "planning":
		return "anthropic", nil
	default:
		return "anthropic", nil
	}
}

func (p *ProviderManager) CheckHealth(ctx context.Context, provider string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	switch provider {
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY") != ""
	case "openai":
		return os.Getenv("OPENAI_API_KEY") != ""
	case "gemini":
		return os.Getenv("GOOGLE_API_KEY") != ""
	case "ollama":
		return true
	default:
		return false
	}
}

func (p *ProviderManager) GetAvailable(ctx context.Context) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var available []string
	for _, prov := range []string{"anthropic", "openai", "gemini", "ollama"} {
		if p.CheckHealth(ctx, prov) {
			available = append(available, prov)
		}
	}
	p.log(fmt.Sprintf("available: %v", available))
	return available
}

func (p *ProviderManager) log(msg string) {
	select {
	case p.logCh <- "[ProviderManager] " + msg:
	default:
		fmt.Println("[ProviderManager]", msg)
	}
}
