package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the Brain daemon.
// Grouped by subsystem for clarity.
type Metrics struct {
	// Registry is the Prometheus registry used for these metrics.
	Registry prometheus.Gatherer

	// Artifact metrics
	ArtifactResolutionDuration *prometheus.HistogramVec
	ArtifactLoadErrors         *prometheus.CounterVec
	ArtifactRegistrySize       prometheus.Gauge

	// Context metrics
	ContextBundleSize         *prometheus.HistogramVec
	ContextCompressionRatio   *prometheus.GaugeVec
	ContextCuratorDuration    *prometheus.HistogramVec
	ContextCuratorSuggestions prometheus.Gauge

	// Policy metrics
	PolicyResolutionDuration *prometheus.HistogramVec
	PolicyOverrides          *prometheus.CounterVec
	PolicyViolations         *prometheus.CounterVec

	// Model routing metrics
	ModelRoutingDecisions *prometheus.CounterVec
	ModelCostUSD          *prometheus.CounterVec
	ModelLatency          *prometheus.HistogramVec

	// Daemon health metrics
	DaemonUptime     prometheus.Gauge
	ActiveSessions   prometheus.Gauge
	HTTPRequests     *prometheus.CounterVec
	HTTPDuration     *prometheus.HistogramVec
	HTTPErrors       *prometheus.CounterVec

	// Sync metrics
	SyncDuration     *prometheus.HistogramVec
	SyncSuccess      *prometheus.CounterVec
	SyncFailures     *prometheus.CounterVec
	SyncFilesChanged prometheus.Gauge
}

// MetricsFactory provides the promauto factory for creating metrics.
type MetricsFactory struct {
	Factory promauto.Factory
}

// NewMetricsFactory creates a MetricsFactory from a registry.
func NewMetricsFactory(reg prometheus.Registerer) MetricsFactory {
	return MetricsFactory{
		Factory: promauto.With(reg),
	}
}

// NewMetrics creates and registers all Prometheus metrics using the given factory.
func (f MetricsFactory) NewMetrics() *Metrics {
	m := &Metrics{}

	// Artifact metrics
	m.ArtifactResolutionDuration = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "artifact",
			Name:      "resolution_duration_seconds",
			Help:      "Time to resolve artifacts hierarchically",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"kind", "scope"},
	)
	m.ArtifactLoadErrors = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "artifact",
			Name:      "load_errors_total",
			Help:      "Total artifact load failures",
		},
		[]string{"kind", "reason"},
	)
	m.ArtifactRegistrySize = f.Factory.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "brain",
			Subsystem: "artifact",
			Name:      "registry_size",
			Help:      "Total number of registered artifacts",
		},
	)

	// Context metrics
	m.ContextBundleSize = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "context",
			Name:      "bundle_size_tokens",
			Help:      "Size of compiled context bundles in tokens",
			Buckets:   []float64{100, 500, 1000, 2000, 4000, 8000, 16000, 32000},
		},
		[]string{"scope_chain"},
	)
	m.ContextCompressionRatio = f.Factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "brain",
			Subsystem: "context",
			Name:      "compression_ratio",
			Help:      "Original/compressed token ratio for context bundles",
		},
		[]string{"bundle_id"},
	)
	m.ContextCuratorDuration = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "context",
			Name:      "curator_duration_seconds",
			Help:      "Duration of curator analysis runs",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"dry_run"},
	)
	m.ContextCuratorSuggestions = f.Factory.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "brain",
			Subsystem: "context",
			Name:      "curator_suggestions",
			Help:      "Number of curator suggestions generated",
		},
	)

	// Policy metrics
	m.PolicyResolutionDuration = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "policy",
			Name:      "resolution_duration_seconds",
			Help:      "Time to resolve policy hierarchy",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"scope_depth"},
	)
	m.PolicyOverrides = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "policy",
			Name:      "overrides_total",
			Help:      "Total policy overrides applied",
		},
		[]string{"class", "scope"},
	)
	m.PolicyViolations = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "policy",
			Name:      "violations_total",
			Help:      "Total policy violations detected",
		},
		[]string{"rule", "severity"},
	)

	// Model routing metrics
	m.ModelRoutingDecisions = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "model",
			Name:      "routing_decisions_total",
			Help:      "Total model routing decisions",
		},
		[]string{"model", "capability_tier", "reason"},
	)
	m.ModelCostUSD = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "model",
			Name:      "cost_usd_total",
			Help:      "Total USD spent on model API calls",
		},
		[]string{"model", "workspace"},
	)
	m.ModelLatency = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "model",
			Name:      "response_duration_seconds",
			Help:      "Model API response duration",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"model", "status"},
	)

	// Daemon health metrics
	m.DaemonUptime = f.Factory.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "brain",
			Subsystem: "daemon",
			Name:      "uptime_seconds",
			Help:      "Daemon uptime in seconds",
		},
	)
	m.ActiveSessions = f.Factory.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "brain",
			Subsystem: "daemon",
			Name:      "active_sessions",
			Help:      "Number of active client sessions",
		},
	)
	m.HTTPRequests = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	m.HTTPDuration = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	m.HTTPErrors = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "http",
			Name:      "errors_total",
			Help:      "Total HTTP errors",
		},
		[]string{"method", "path", "status"},
	)

	// Sync metrics
	m.SyncDuration = f.Factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "brain",
			Subsystem: "sync",
			Name:      "duration_seconds",
			Help:      "Sync operation duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"worker", "status"},
	)
	m.SyncSuccess = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "sync",
			Name:      "success_total",
			Help:      "Total successful sync operations",
		},
		[]string{"worker"},
	)
	m.SyncFailures = f.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "brain",
			Subsystem: "sync",
			Name:      "failures_total",
			Help:      "Total failed sync operations",
		},
		[]string{"worker", "reason"},
	)
	m.SyncFilesChanged = f.Factory.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "brain",
			Subsystem: "sync",
			Name:      "files_changed",
			Help:      "Number of files changed in last sync",
		},
	)

	return m
}

// DefaultMetrics is the global metrics instance using the default registry.
var DefaultMetrics = NewMetricsFactory(prometheus.DefaultRegisterer).NewMetrics()
