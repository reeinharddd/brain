package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth represents the health of a single component.
type ComponentHealth struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Checked time.Time    `json:"checked_at"`
}

// HealthCheck represents a named health check function.
type HealthCheck struct {
	Name    string
	Check   func(ctx context.Context) error
	Timeout time.Duration
}

// HealthChecker manages component health checks.
type HealthChecker struct {
	mu       sync.RWMutex
	startTime time.Time
	version   string
	checks    map[string]*HealthCheck
	results   map[string]*ComponentHealth
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		version:   version,
		checks:    make(map[string]*HealthCheck),
		results:   make(map[string]*ComponentHealth),
	}
}

// Register adds a health check to the checker.
func (hc *HealthChecker) Register(name string, timeout time.Duration, check func(ctx context.Context) error) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = &HealthCheck{
		Name:    name,
		Check:   check,
		Timeout: timeout,
	}
}

// RunCheck executes a single health check and stores the result.
func (hc *HealthChecker) RunCheck(ctx context.Context, name string) {
	hc.mu.Lock()
	check, exists := hc.checks[name]
	hc.mu.Unlock()

	if !exists {
		return
	}

	timeout := check.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &ComponentHealth{
		Checked: time.Now(),
	}

	err := check.Check(checkCtx)
	if err == nil {
		result.Status = HealthStatusHealthy
		result.Message = "ok"
	} else {
		result.Status = HealthStatusUnhealthy
		result.Message = err.Error()
	}

	hc.mu.Lock()
	hc.results[name] = result
	hc.mu.Unlock()
}

// RunAll executes all registered health checks.
func (hc *HealthChecker) RunAll(ctx context.Context) {
	var wg sync.WaitGroup
	for name := range hc.checks {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			hc.RunCheck(ctx, n)
		}(name)
	}
	wg.Wait()
}

// GetOverallStatus returns the overall health status.
func (hc *HealthChecker) GetOverallStatus() HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if len(hc.results) == 0 {
		return HealthStatusUnknown
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, result := range hc.results {
		switch result.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// HealthResponse is the JSON response for health checks.
type HealthResponse struct {
	Status   HealthStatus             `json:"status"`
	Uptime   float64                  `json:"uptime_seconds"`
	Version  string                   `json:"version"`
	Checked  time.Time                `json:"checked_at"`
	Components map[string]*ComponentHealth `json:"components"`
}

// GetHealthResponse builds the full health response.
func (hc *HealthChecker) GetHealthResponse() *HealthResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	components := make(map[string]*ComponentHealth)
	for name, result := range hc.results {
		components[name] = result
	}

	return &HealthResponse{
		Status:     hc.GetOverallStatus(),
		Uptime:     time.Since(hc.startTime).Seconds(),
		Version:    hc.version,
		Checked:    time.Now(),
		Components: components,
	}
}

// HealthHandler returns an http.Handler for the health check endpoint.
func (hc *HealthChecker) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Run all checks before responding
		hc.RunAll(r.Context())

		response := hc.GetHealthResponse()

		w.Header().Set("Content-Type", "application/json")

		switch response.Status {
		case HealthStatusHealthy:
			w.WriteHeader(http.StatusOK)
		case HealthStatusDegraded:
			w.WriteHeader(http.StatusPartialContent)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		}
	})
}

// SimpleHealthHandler returns a simple health check endpoint without component checks.
func SimpleHealthHandler(version string) http.Handler {
	startTime := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":         "healthy",
			"uptime_seconds": time.Since(startTime).Seconds(),
			"version":        version,
			"checked_at":     time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(response)
	})
}
