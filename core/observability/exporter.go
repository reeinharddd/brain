package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ExporterConfig holds configuration for telemetry exporters.
type ExporterConfig struct {
	// PrometheusEnabled enables the Prometheus metrics exporter.
	PrometheusEnabled bool
	// PrometheusListenAddr is the address to listen on for Prometheus scraping.
	PrometheusListenAddr string
	// PrometheusPath is the HTTP path for metrics scraping.
	PrometheusPath string
	// PrometheusRegistry is the Prometheus registry to use (defaults to DefaultRegisterer).
	PrometheusRegistry prometheus.Gatherer
}

// DefaultExporterConfig returns an ExporterConfig with sensible defaults.
func DefaultExporterConfig() *ExporterConfig {
	return &ExporterConfig{
		PrometheusEnabled:    true,
		PrometheusListenAddr: ":9090",
		PrometheusPath:       "/metrics",
	}
}

// Exporter manages telemetry exporters.
type Exporter struct {
	config   *ExporterConfig
	server   *http.Server
	shutdown func()
}

// NewExporter creates a new Exporter with the given configuration.
func NewExporter(cfg *ExporterConfig, shutdown func()) *Exporter {
	if cfg == nil {
		cfg = DefaultExporterConfig()
	}
	return &Exporter{
		config:   cfg,
		shutdown: shutdown,
	}
}

// Start begins serving metrics on the configured address.
func (e *Exporter) Start() error {
	if !e.config.PrometheusEnabled {
		return nil
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	registry := e.config.PrometheusRegistry
	if registry == nil {
		registry = prometheus.DefaultGatherer
	}

	mux.Handle(e.config.PrometheusPath, promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	e.server = &http.Server{
		Addr:              e.config.PrometheusListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't fail - metrics endpoint is not critical
			_ = fmt.Errorf("metrics server error: %w", err)
		}
	}()

	return nil
}

// Shutdown gracefully stops the exporter.
func (e *Exporter) Shutdown(ctx context.Context) error {
	if e.shutdown != nil {
		e.shutdown()
	}

	if e.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return e.server.Shutdown(shutdownCtx)
	}

	return nil
}
