package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

// ---- Logger Tests ----

func TestNewLogger_Default(t *testing.T) {
	logger := NewLogger(nil)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := &LoggerConfig{
		Level:       "info",
		Format:      "json",
		Output:      &buf,
		ServiceName: "test-service",
		Version:     "1.0.0",
	}
	logger := NewLogger(cfg)
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("expected JSON output to contain level, got: %s", output)
	}
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("expected JSON output to contain message, got: %s", output)
	}
	if !strings.Contains(output, `"service":"test-service"`) {
		t.Errorf("expected JSON output to contain service name, got: %s", output)
	}
}

func TestNewLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := &LoggerConfig{
		Level:       "debug",
		Format:      "text",
		Output:      &buf,
		ServiceName: "test-service",
	}
	logger := NewLogger(cfg)
	logger.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "level=DEBUG") {
		t.Errorf("expected text output to contain level, got: %s", output)
	}
}

func TestNewLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	cfg := &LoggerConfig{
		Level:       "warn",
		Format:      "json",
		Output:      &buf,
		ServiceName: "test-service",
	}
	logger := NewLogger(cfg)
	logger.Info("info message")
	logger.Warn("warn message")

	output := buf.String()
	if strings.Contains(output, "info message") {
		t.Error("expected info message to be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("expected warn message to be present")
	}
}

func TestTraceIDFromContext(t *testing.T) {
	ctx := context.Background()
	
	// Without trace ID
	traceID := TraceIDFromContext(ctx)
	if traceID != "unknown" {
		t.Errorf("expected 'unknown', got %s", traceID)
	}

	// With trace ID
	ctx = WithTraceID(ctx, "test-trace-123")
	traceID = TraceIDFromContext(ctx)
	if traceID != "test-trace-123" {
		t.Errorf("expected 'test-trace-123', got %s", traceID)
	}
}

func TestExecutionIDFromContext(t *testing.T) {
	ctx := context.Background()
	
	// Without execution ID
	execID := ExecutionIDFromContext(ctx)
	if execID != "" {
		t.Errorf("expected empty string, got %s", execID)
	}

	// With execution ID
	ctx = WithExecutionID(ctx, "exec-456")
	execID = ExecutionIDFromContext(ctx)
	if execID != "exec-456" {
		t.Errorf("expected 'exec-456', got %s", execID)
	}
}

// ---- Tracer Tests ----

func TestDefaultTracerConfig(t *testing.T) {
	cfg := DefaultTracerConfig()
	if cfg.ServiceName != "brain-daemon" {
		t.Errorf("expected service name 'brain-daemon', got %s", cfg.ServiceName)
	}
	if cfg.Enabled {
		t.Error("expected tracer to be disabled by default")
	}
}

func TestInitTracer_Disabled(t *testing.T) {
	cfg := &TracerConfig{Enabled: false}
	tp, shutdown, err := InitTracer(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp != nil {
		t.Error("expected nil tracer when disabled")
	}
	if shutdown == nil {
		t.Error("expected non-nil shutdown function")
	}
	shutdown() // Should not panic
}

func TestInitTracer_Enabled(t *testing.T) {
	cfg := &TracerConfig{
		ServiceName:  "test-daemon",
		ServiceVersion: "0.1.0",
		OTLPEndpoint: "localhost:4318",
		OTLPInsecure: true,
		SampleRate:   1.0,
		Enabled:      true,
	}
	
	// This will fail to connect to OTLP endpoint, but should not panic
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	tp, shutdown, err := InitTracer(ctx, cfg)
	if err != nil {
		// Expected: connection refused
		t.Logf("expected error connecting to OTLP: %v", err)
	}
	
	if tp != nil {
		defer tp.Shutdown(context.Background())
	}
	if shutdown != nil {
		shutdown()
	}
}

func TestTracer_NamedTracer(t *testing.T) {
	tracer := Tracer("test-component")
	if tracer == nil {
		t.Error("expected non-nil tracer")
	}
}

func TestRecordError(t *testing.T) {
	// Create a tracer provider for testing
	tp := trace.NewTracerProvider()
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	
	// Record nil error (should not panic)
	RecordError(span, nil)
	
	// Record actual error
	testErr := &testError{msg: "test error"}
	RecordError(span, testErr)
	
	span.End()
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// ---- Metrics Tests ----

func TestNewMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	factory := NewMetricsFactory(reg)
	m := factory.NewMetrics()
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
	if m.ArtifactResolutionDuration == nil {
		t.Error("expected non-nil ArtifactResolutionDuration")
	}
	if m.DaemonUptime == nil {
		t.Error("expected non-nil DaemonUptime")
	}
}

func TestDefaultMetrics(t *testing.T) {
	if DefaultMetrics == nil {
		t.Fatal("expected non-nil DefaultMetrics")
	}
}

func TestMetrics_ObserveArtifactResolution(t *testing.T) {
	reg := prometheus.NewRegistry()
	factory := NewMetricsFactory(reg)
	m := factory.NewMetrics()

	m.ArtifactResolutionDuration.WithLabelValues("skill", "workspace").Observe(0.5)

	// Gather metrics to verify
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("expected at least one metric family")
	}
}

func TestMetrics_HTTPRequests(t *testing.T) {
	reg := prometheus.NewRegistry()
	factory := NewMetricsFactory(reg)
	m := factory.NewMetrics()

	m.HTTPRequests.WithLabelValues("GET", "/api/v1/health", "200").Inc()

	// Gather and verify
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("expected at least one metric family")
	}
}

// ---- Health Check Tests ----

func TestHealthChecker_RegisterAndRun(t *testing.T) {
	hc := NewHealthChecker("0.1.0")
	
	hc.Register("test-component", 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	
	hc.RunCheck(context.Background(), "test-component")
	
	status := hc.GetOverallStatus()
	if status != HealthStatusHealthy {
		t.Errorf("expected healthy status, got %s", status)
	}
}

func TestHealthChecker_UnhealthyComponent(t *testing.T) {
	hc := NewHealthChecker("0.1.0")
	
	hc.Register("failing-component", 2*time.Second, func(ctx context.Context) error {
		return &testError{msg: "connection refused"}
	})
	
	hc.RunCheck(context.Background(), "failing-component")
	
	status := hc.GetOverallStatus()
	if status != HealthStatusUnhealthy {
		t.Errorf("expected unhealthy status, got %s", status)
	}
}

func TestHealthChecker_RunAll(t *testing.T) {
	hc := NewHealthChecker("0.1.0")
	
	hc.Register("component-1", 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	hc.Register("component-2", 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	
	hc.RunAll(context.Background())
	
	status := hc.GetOverallStatus()
	if status != HealthStatusHealthy {
		t.Errorf("expected healthy status, got %s", status)
	}
}

func TestHealthChecker_Response(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	
	hc.Register("db", 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	
	hc.RunAll(context.Background())
	
	response := hc.GetHealthResponse()
	if response.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", response.Version)
	}
	if response.Status != HealthStatusHealthy {
		t.Errorf("expected status healthy, got %s", response.Status)
	}
	if response.Uptime <= 0 {
		t.Error("expected positive uptime")
	}
	if len(response.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(response.Components))
	}
}

func TestHealthHandler_HTTP(t *testing.T) {
	hc := NewHealthChecker("0.1.0")
	
	hc.Register("test", 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	
	handler := hc.HealthHandler()
	
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	
	var response HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	
	if response.Version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %s", response.Version)
	}
}

func TestSimpleHealthHandler(t *testing.T) {
	handler := SimpleHealthHandler("0.2.0")
	
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	
	if response["version"] != "0.2.0" {
		t.Errorf("expected version '0.2.0', got %v", response["version"])
	}
	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}
}

// ---- Trace Context Tests ----

func TestTraceContext_Extract(t *testing.T) {
	tc := NewTraceContext()
	
	// Create a request with trace context headers
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	
	ctx := tc.Extract(req)
	if ctx == nil {
		t.Error("expected non-nil context")
	}
}

func TestTraceContext_Middleware(t *testing.T) {
	tc := NewTraceContext()
	
	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})
	
	handler := tc.Middleware(next)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if capturedCtx == nil {
		t.Error("expected context to be captured")
	}
}

func TestTraceIDFromSpan_InvalidSpan(t *testing.T) {
	ctx := context.Background()
	traceID := TraceIDFromSpan(ctx)
	if traceID != "" {
		t.Errorf("expected empty trace ID, got %s", traceID)
	}
}

func TestSpanIDFromSpan_InvalidSpan(t *testing.T) {
	ctx := context.Background()
	spanID := SpanIDFromSpan(ctx)
	if spanID != "" {
		t.Errorf("expected empty span ID, got %s", spanID)
	}
}

// ---- Exporter Tests ----

func TestDefaultExporterConfig(t *testing.T) {
	cfg := DefaultExporterConfig()
	if !cfg.PrometheusEnabled {
		t.Error("expected Prometheus enabled by default")
	}
	if cfg.PrometheusListenAddr != ":9090" {
		t.Errorf("expected listen addr '::9090', got %s", cfg.PrometheusListenAddr)
	}
	if cfg.PrometheusPath != "/metrics" {
		t.Errorf("expected metrics path '/metrics', got %s", cfg.PrometheusPath)
	}
}

func TestNewExporter(t *testing.T) {
	exporter := NewExporter(nil, func() {})
	if exporter == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestExporter_StartDisabled(t *testing.T) {
	cfg := &ExporterConfig{PrometheusEnabled: false}
	exporter := NewExporter(cfg, func() {})
	
	err := exporter.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExporter_MetricsEndpoint(t *testing.T) {
	// Create a custom registry with a test metric
	reg := prometheus.NewRegistry()
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "A test counter",
	})
	reg.MustRegister(counter)
	counter.Inc()
	
	// Create handler
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	
	body := rec.Body.String()
	if !strings.Contains(body, "test_counter") {
		t.Errorf("expected metrics body to contain 'test_counter', got: %s", body)
	}
}

func TestExporter_Shutdown(t *testing.T) {
	cfg := &ExporterConfig{PrometheusEnabled: false}
	exporter := NewExporter(cfg, func() {})
	
	err := exporter.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on shutdown: %v", err)
	}
}

// ---- Integration Test: Logger + Trace Context ----

func TestLoggerWithTraceContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := &LoggerConfig{
		Level:       "info",
		Format:      "json",
		Output:      &buf,
		ServiceName: "test-service",
	}
	logger := NewLogger(cfg)
	
	ctx := WithTraceID(context.Background(), "trace-abc123")
	traceID := TraceIDFromContext(ctx)
	
	logger.InfoContext(ctx, "processing request", "trace_id", traceID)
	
	output := buf.String()
	if !strings.Contains(output, "trace-abc123") {
		t.Errorf("expected trace ID in log output, got: %s", output)
	}
}

// ---- Benchmark Tests ----

func BenchmarkLogger_JSON(b *testing.B) {
	var buf bytes.Buffer
	cfg := &LoggerConfig{
		Level:       "info",
		Format:      "json",
		Output:      &buf,
		ServiceName: "benchmark",
	}
	logger := NewLogger(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("benchmark message", "key", "value")
	}
}

func BenchmarkHealthChecker_RunCheck(b *testing.B) {
	hc := NewHealthChecker("0.1.0")
	hc.Register("test", 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hc.RunCheck(context.Background(), "test")
	}
}

func BenchmarkMetrics_ArtifactResolution(b *testing.B) {
	reg := prometheus.NewRegistry()
	factory := NewMetricsFactory(reg)
	m := factory.NewMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ArtifactResolutionDuration.WithLabelValues("skill", "workspace").Observe(0.5)
	}
}
