package manager

import (
	"context"
	"fmt"
	"sync"
)

type MCPRegistry struct {
	mu     sync.RWMutex
	items  map[string]interface{}
	logCh  chan string
}

func NewMCPRegistry(logCh chan string) *MCPRegistry {
	return &MCPRegistry{
		items:  make(map[string]interface{}),
		logCh:  logCh,
	}
}

func (r *MCPRegistry) Sync(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log("syncing registry")
	return nil
}

func (r *MCPRegistry) GetStatus(ctx context.Context) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return map[string]interface{}{
		"count": len(r.items),
	}
}

func (r *MCPRegistry) Start(ctx context.Context) error {
	r.log("starting")
	return r.Sync(ctx)
}

func (r *MCPRegistry) Stop() error {
	r.log("stopping")
	return nil
}

func (r *MCPRegistry) log(msg string) {
	select {
	case r.logCh <- "[MCPRegistry] " + msg:
	default:
		fmt.Println("[MCPRegistry]", msg)
	}
}
